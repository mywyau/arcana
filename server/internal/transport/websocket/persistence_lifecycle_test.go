package websockettransport

import (
	"context"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"github.com/mywyau/arcana/server/internal/account"
	"github.com/mywyau/arcana/server/internal/character"
	"github.com/mywyau/arcana/server/internal/identity"
	"github.com/mywyau/arcana/server/internal/protocol"
	"github.com/mywyau/arcana/server/internal/simulation"
)

func TestAccountCharacterPersistsAcrossPeriodicSaveDisconnectAndHandlerRestart(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	accountRepository := account.NewMemoryRepository()
	accountManager, err := account.NewManager(accountRepository, accountRepository)
	if err != nil {
		t.Fatalf("account.NewManager() error = %v", err)
	}
	guestManager, err := identity.NewGuestManager(identity.NewMemoryGuestRepository())
	if err != nil {
		t.Fatalf("identity.NewGuestManager() error = %v", err)
	}
	characters := character.NewMemoryRepository()
	authentication, err := accountManager.RegisterNewCharacter(
		ctx, "persistence@example.com", "strong-password", characters,
	)
	if err != nil {
		t.Fatalf("RegisterNewCharacter() error = %v", err)
	}

	initial := simulation.DefaultPersistentPlayerState()
	initial.Position = simulation.Position{X: 3, Y: 2, Z: -1}
	initial.Health = 76
	initial.Prayer = 37
	initial.ActivePrayerSet = simulation.PrayerSetWar
	initial.ExperienceLocked = true
	initial.Inventory.Items = []simulation.InventoryItem{{Slot: 0, ItemID: simulation.ItemCoins, Quantity: 321}}
	if err := characters.Save(ctx, authentication.PlayerID, initial); err != nil {
		t.Fatalf("seed character: %v", err)
	}

	firstWorld, firstStop := startPersistenceTestWorld(t)
	firstHandler, err := NewHandlerWithRepositoriesAndOptions(
		firstWorld, 5*time.Millisecond, guestManager, characters,
		HandlerOptions{AccountManager: accountManager, RequireAccounts: true},
	)
	if err != nil {
		firstStop()
		t.Fatalf("create first handler: %v", err)
	}
	firstHandler.characterSaveInterval = 20 * time.Millisecond
	firstHandler.disconnectedLogoutDelay = 20 * time.Millisecond
	firstServer := startTestServer(t, firstHandler)

	firstConnection := connectWithSession(
		t, ctx, websocketURL(firstServer.URL), authentication.SessionToken,
	)
	firstWelcome := readServerMessage(t, ctx, firstConnection)
	if firstWelcome.Type != protocol.MessageWelcome || firstWelcome.PlayerID != authentication.PlayerID {
		t.Fatalf("first welcome = %+v", firstWelcome)
	}
	firstSnapshot := waitForLifecyclePlayer(t, ctx, firstConnection, authentication.PlayerID, func(player protocol.PlayerSnapshot) bool {
		return player.Position == (protocol.Position{X: 3, Y: 2, Z: -1})
	})
	if firstSnapshot.Health != 76 || firstSnapshot.Prayer != 37 ||
		firstSnapshot.ActivePrayerSet != string(simulation.PrayerSetWar) || !firstSnapshot.ExperienceLocked ||
		len(firstSnapshot.Inventory.Items) != 1 {
		t.Fatalf("initial restored snapshot = %+v", firstSnapshot)
	}

	if err := wsjson.Write(ctx, firstConnection, protocol.ClientMessage{
		Type: protocol.MessageMoveTo, Sequence: 1,
		Destination: &protocol.Position{X: 4, Y: 2, Z: -1},
	}); err != nil {
		t.Fatalf("write movement: %v", err)
	}
	moved := waitForLifecyclePlayer(t, ctx, firstConnection, authentication.PlayerID, func(player protocol.PlayerSnapshot) bool {
		return player.Position.X > 3
	})

	periodic := waitForPersistedPosition(t, ctx, characters, authentication.PlayerID, func(position simulation.Position) bool {
		return position.X > 3
	})
	if periodic.Health != 76 || periodic.Prayer != 37 || periodic.ActivePrayerSet != simulation.PrayerSetWar {
		t.Fatalf("periodically saved state = %+v", periodic)
	}

	if err := firstHandler.SaveAll(ctx); err != nil {
		t.Fatalf("SaveAll() error = %v", err)
	}
	if err := firstConnection.Close(websocket.StatusNormalClosure, "persistence lifecycle test"); err != nil {
		t.Fatalf("close first connection: %v", err)
	}
	waitForPlayerToLeave(t, firstWorld, simulation.PlayerID(authentication.PlayerID))
	firstServer.Close()
	firstStop()

	saved, found, err := characters.Load(ctx, authentication.PlayerID)
	if err != nil || !found {
		t.Fatalf("load disconnected character = found %t, error %v", found, err)
	}
	if saved.Position.X < moved.Position.X-0.1 || saved.Prayer != 37 || saved.ActivePrayerSet != simulation.PrayerSetWar {
		t.Fatalf("disconnect save = %+v, moved snapshot = %+v", saved, moved)
	}

	secondWorld, secondStop := startPersistenceTestWorld(t)
	defer secondStop()
	secondHandler, err := NewHandlerWithRepositoriesAndOptions(
		secondWorld, 5*time.Millisecond, guestManager, characters,
		HandlerOptions{AccountManager: accountManager, RequireAccounts: true},
	)
	if err != nil {
		t.Fatalf("create restarted handler: %v", err)
	}
	secondServer := startTestServer(t, secondHandler)
	defer secondServer.Close()
	secondConnection := connectWithSession(
		t, ctx, websocketURL(secondServer.URL), authentication.SessionToken,
	)
	defer secondConnection.CloseNow()
	secondWelcome := readServerMessage(t, ctx, secondConnection)
	if secondWelcome.Type != protocol.MessageWelcome || secondWelcome.PlayerID != authentication.PlayerID {
		t.Fatalf("restarted welcome = %+v", secondWelcome)
	}
	restored := waitForLifecyclePlayer(t, ctx, secondConnection, authentication.PlayerID, func(player protocol.PlayerSnapshot) bool {
		return player.Position.X >= saved.Position.X-0.1
	})
	if restored.Health != saved.Health || restored.Prayer != saved.Prayer ||
		restored.ActivePrayerSet != string(saved.ActivePrayerSet) || !restored.ExperienceLocked ||
		len(restored.Inventory.Items) != 1 || restored.Inventory.Items[0].Quantity != 321 {
		t.Fatalf("state after handler restart = %+v, saved = %+v", restored, saved)
	}
}

func startPersistenceTestWorld(t *testing.T) (*simulation.World, context.CancelFunc) {
	t.Helper()
	world := simulation.NewWorld()
	runner, err := simulation.NewRunner(world, 5*time.Millisecond)
	if err != nil {
		t.Fatalf("simulation.NewRunner() error = %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	go runner.Run(ctx)
	return world, cancel
}

func waitForLifecyclePlayer(
	t *testing.T,
	ctx context.Context,
	connection *websocket.Conn,
	playerID string,
	ready func(protocol.PlayerSnapshot) bool,
) protocol.PlayerSnapshot {
	t.Helper()
	for {
		message := readServerMessage(t, ctx, connection)
		if message.Type != protocol.MessageSnapshot || message.Snapshot == nil {
			continue
		}
		player, found := findPlayer(message.Snapshot, playerID)
		if found && ready(player) {
			return player
		}
	}
}

func waitForPersistedPosition(
	t *testing.T,
	ctx context.Context,
	repository *character.MemoryRepository,
	playerID string,
	ready func(simulation.Position) bool,
) simulation.PersistentPlayerState {
	t.Helper()
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()
	for {
		state, found, err := repository.Load(ctx, playerID)
		if err != nil {
			t.Fatalf("load periodically saved character: %v", err)
		}
		if found && ready(state.Position) {
			return state
		}
		select {
		case <-ctx.Done():
			t.Fatalf("wait for periodic character save: %v", ctx.Err())
		case <-ticker.C:
		}
	}
}
