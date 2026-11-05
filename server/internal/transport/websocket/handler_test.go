package websockettransport

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"github.com/mywyau/arcana/server/internal/character"
	"github.com/mywyau/arcana/server/internal/identity"
	"github.com/mywyau/arcana/server/internal/protocol"
	"github.com/mywyau/arcana/server/internal/simulation"
)

func TestHandlerMovesConnectionPlayerAndSendsSnapshots(t *testing.T) {
	world := simulation.NewWorld()
	const interval = 5 * time.Millisecond
	runner, err := simulation.NewRunner(world, interval)
	if err != nil {
		t.Fatalf("NewRunner() returned an unexpected error: %v", err)
	}
	runnerCtx, stopRunner := context.WithCancel(context.Background())
	defer stopRunner()
	go runner.Run(runnerCtx)

	handler, err := NewHandler(world, interval)
	if err != nil {
		t.Fatalf("NewHandler() returned an unexpected error: %v", err)
	}
	server := startTestServer(t, handler)
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	connection, _, err := websocket.Dial(ctx, websocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("websocket.Dial() returned an unexpected error: %v", err)
	}
	defer connection.CloseNow()

	authenticateGuest(t, ctx, connection, "")
	welcome := readServerMessage(t, ctx, connection)
	if welcome.Type != protocol.MessageWelcome || welcome.PlayerID == "" || welcome.GuestToken == "" {
		t.Fatalf("welcome message = %+v", welcome)
	}

	command := protocol.ClientMessage{
		Type:        protocol.MessageMoveTo,
		Sequence:    1,
		Destination: &protocol.Position{Y: 10},
	}
	if err := wsjson.Write(ctx, connection, command); err != nil {
		t.Fatalf("wsjson.Write() returned an unexpected error: %v", err)
	}

	for {
		message := readServerMessage(t, ctx, connection)
		if message.Type != protocol.MessageSnapshot || message.Snapshot == nil {
			continue
		}
		player, exists := findPlayer(message.Snapshot, welcome.PlayerID)
		if exists && player.LastCommandSequence == 1 {
			if player.Position.Y <= 0 {
				t.Fatalf("player position = %+v, want positive Y movement", player.Position)
			}
			break
		}
	}
}

func TestHandlerReportsInvalidMessages(t *testing.T) {
	world := simulation.NewWorld()
	handler, err := NewHandler(world, 5*time.Millisecond)
	if err != nil {
		t.Fatalf("NewHandler() returned an unexpected error: %v", err)
	}
	server := startTestServer(t, handler)
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	connection, _, err := websocket.Dial(ctx, websocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("websocket.Dial() returned an unexpected error: %v", err)
	}
	defer connection.CloseNow()
	authenticateGuest(t, ctx, connection, "")
	_ = readServerMessage(t, ctx, connection)

	if err := wsjson.Write(ctx, connection, map[string]any{"type": "dance"}); err != nil {
		t.Fatalf("wsjson.Write() returned an unexpected error: %v", err)
	}

	for {
		message := readServerMessage(t, ctx, connection)
		if message.Type == protocol.MessageError {
			if message.Error == nil || message.Error.Code != "invalid_message" {
				t.Fatalf("error message = %+v", message)
			}
			break
		}
	}
}

func TestHandlerSavesCharacterOnDisconnect(t *testing.T) {
	world := simulation.NewWorld()
	const interval = 5 * time.Millisecond
	runner, err := simulation.NewRunner(world, interval)
	if err != nil {
		t.Fatalf("NewRunner() returned an unexpected error: %v", err)
	}
	runnerCtx, stopRunner := context.WithCancel(context.Background())
	defer stopRunner()
	go runner.Run(runnerCtx)

	guestIdentities, err := identity.NewGuestManager(identity.NewMemoryGuestRepository())
	if err != nil {
		t.Fatalf("NewGuestManager() returned an unexpected error: %v", err)
	}
	characters := character.NewMemoryRepository()
	handler, err := NewHandlerWithRepositories(world, interval, guestIdentities, characters)
	if err != nil {
		t.Fatalf("NewHandlerWithRepositories() returned an unexpected error: %v", err)
	}
	server := startTestServer(t, handler)
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	connection, _, err := websocket.Dial(ctx, websocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("websocket.Dial() returned an unexpected error: %v", err)
	}
	authenticateGuest(t, ctx, connection, "")
	welcome := readServerMessage(t, ctx, connection)
	if err := wsjson.Write(ctx, connection, protocol.ClientMessage{
		Type: protocol.MessageMoveTo, Sequence: 1,
		Destination: &protocol.Position{Y: 10},
	}); err != nil {
		t.Fatalf("write movement: %v", err)
	}

	for {
		message := readServerMessage(t, ctx, connection)
		if message.Type != protocol.MessageSnapshot || message.Snapshot == nil {
			continue
		}
		player, found := findPlayer(message.Snapshot, welcome.PlayerID)
		if found && player.Position.Y > 0 {
			break
		}
	}
	if err := connection.Close(websocket.StatusNormalClosure, "test complete"); err != nil {
		t.Fatalf("close connection: %v", err)
	}

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		state, found, loadErr := characters.Load(ctx, welcome.PlayerID)
		if loadErr != nil {
			t.Fatalf("load saved character: %v", loadErr)
		}
		if found && state.Position.Y > 0 {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("character movement was not saved after disconnect")
}

func TestSaveAllPersistsEveryActiveCharacter(t *testing.T) {
	world := simulation.NewWorld()
	if err := world.AddPlayer("player-one", simulation.Position{X: 2}, 4); err != nil {
		t.Fatalf("AddPlayer(player-one): %v", err)
	}
	if err := world.AddPlayer("player-two", simulation.Position{Z: -3}, 4); err != nil {
		t.Fatalf("AddPlayer(player-two): %v", err)
	}
	characters := character.NewMemoryRepository()
	guestIdentities, err := identity.NewGuestManager(identity.NewMemoryGuestRepository())
	if err != nil {
		t.Fatalf("NewGuestManager() returned an unexpected error: %v", err)
	}
	handler, err := NewHandlerWithRepositories(
		world,
		5*time.Millisecond,
		guestIdentities,
		characters,
	)
	if err != nil {
		t.Fatalf("NewHandlerWithRepositories() returned an unexpected error: %v", err)
	}

	if err := handler.SaveAll(context.Background()); err != nil {
		t.Fatalf("SaveAll() returned an unexpected error: %v", err)
	}
	for _, playerID := range []string{"player-one", "player-two"} {
		if _, found, loadErr := characters.Load(context.Background(), playerID); loadErr != nil || !found {
			t.Fatalf("Load(%s) = found %t, error %v", playerID, found, loadErr)
		}
	}
}

func authenticateGuest(t *testing.T, ctx context.Context, connection *websocket.Conn, token string) {
	t.Helper()
	if err := wsjson.Write(ctx, connection, protocol.ClientMessage{
		Type:       protocol.MessageAuthenticate,
		GuestToken: token,
	}); err != nil {
		t.Fatalf("write authentication message: %v", err)
	}
}

func TestSnapshotMessageIncludesCombatEvents(t *testing.T) {
	message := snapshotMessage(simulation.Snapshot{
		Targets: []simulation.TargetSnapshot{
			{
				ID:              "dummy",
				Kind:            "creature",
				TargetID:        "player-one",
				NextAttackPulse: 16,
				Defeated:        true,
				RespawnAtPulse:  20,
				DodgeRate:       0.2,
			},
			{
				ID:            "upper_dummy",
				Position:      simulation.Position{X: -4, Y: 3, Z: 2},
				Health:        75,
				MaximumHealth: 100,
				DodgeRate:     0.1,
			},
		},
		Drops: []simulation.DropSnapshot{{
			ID:       "drop-1",
			Position: simulation.Position{X: 2, Y: 3, Z: 4},
			ItemID:   "training_shard",
			Quantity: 2,
		}},
		CombatEvents: []simulation.CombatEvent{{
			ID:       7,
			Pulse:    12,
			Type:     simulation.CombatEventDamage,
			SourceID: "player-one",
			TargetID: "dummy",
			Amount:   10,
		}},
		Players: []simulation.PlayerSnapshot{{
			ID:                      "player-one",
			Health:                  40,
			MaximumHealth:           100,
			Mana:                    80,
			MaximumMana:             100,
			Prayer:                  70,
			MaximumPrayer:           100,
			Defeated:                true,
			RespawnAtPulse:          20,
			EquippedWeapon:          simulation.WeaponDagger,
			LastEquipmentSequence:   3,
			LastEquipmentPulse:      11,
			CombatStyle:             simulation.CombatStyleStab,
			LastCombatStyleSequence: 4,
			DodgeRate:               0.15,
			Inventory: simulation.InventorySnapshot{
				Capacity: 8,
				Items: []simulation.InventoryItem{{
					Slot: 0, ItemID: "sword", Quantity: 1,
				}},
			},
			Skills: []simulation.SkillSnapshot{{
				ID: simulation.SkillAttack, Experience: 125, Level: 2,
				CurrentLevelExperience: 100, NextLevelExperience: 400,
			}},
		}},
	})

	if len(message.CombatEvents) != 1 {
		t.Fatalf("combat event count = %d, want 1", len(message.CombatEvents))
	}
	event := message.CombatEvents[0]
	if event.ID != 7 || event.Pulse != 12 || event.Type != "damage" ||
		event.SourceID != "player-one" || event.TargetID != "dummy" || event.Amount != 10 {
		t.Fatalf("combat event = %+v", event)
	}
	dummy, found := findTarget(message, "dummy")
	if !found || dummy.Kind != "creature" || dummy.TargetID != "player-one" || dummy.NextAttackPulse != 16 ||
		!dummy.Defeated || dummy.RespawnAtPulse != 20 || dummy.DodgeRate != 0.2 {
		t.Fatalf("dummy combat state = %+v, found %t", dummy, found)
	}
	upperDummy, found := findTarget(message, "upper_dummy")
	if !found || upperDummy.Position.X != -4 || upperDummy.Position.Y != 3 ||
		upperDummy.Position.Z != 2 || upperDummy.Health != 75 ||
		upperDummy.MaximumHealth != 100 || upperDummy.DodgeRate != 0.1 {
		t.Fatalf("upper dummy combat state = %+v, found %t", upperDummy, found)
	}
	if len(message.Drops) != 1 || message.Drops[0].ID != "drop-1" ||
		message.Drops[0].Position.X != 2 || message.Drops[0].Position.Y != 3 ||
		message.Drops[0].Position.Z != 4 || message.Drops[0].ItemID != "training_shard" ||
		message.Drops[0].Quantity != 2 {
		t.Fatalf("drop snapshot = %+v", message.Drops)
	}
	player := message.Players[0]
	if player.Health != 40 || player.MaximumHealth != 100 || player.Mana != 80 || player.MaximumMana != 100 ||
		player.Prayer != 70 || player.MaximumPrayer != 100 || !player.Defeated || player.RespawnAtPulse != 20 ||
		player.EquippedWeapon != "dagger" || player.LastEquipmentSequence != 3 || player.LastEquipmentPulse != 11 ||
		player.CombatStyle != "stab" || player.LastCombatStyleSequence != 4 || player.DodgeRate != 0.15 {
		t.Fatalf("player vital state = %+v", player)
	}
	if player.Inventory.Capacity != 8 || len(player.Inventory.Items) != 1 ||
		player.Inventory.Items[0].Slot != 0 || player.Inventory.Items[0].ItemID != "sword" ||
		player.Inventory.Items[0].Quantity != 1 {
		t.Fatalf("player inventory = %+v", player.Inventory)
	}
	if len(player.Skills) != 1 || player.Skills[0].ID != "attack" ||
		player.Skills[0].Experience != 125 || player.Skills[0].Level != 2 ||
		player.Skills[0].CurrentLevelExperience != 100 || player.Skills[0].NextLevelExperience != 400 {
		t.Fatalf("player skills = %+v", player.Skills)
	}
}

func readServerMessage(t *testing.T, ctx context.Context, connection *websocket.Conn) protocol.ServerMessage {
	t.Helper()
	var message protocol.ServerMessage
	if err := wsjson.Read(ctx, connection, &message); err != nil {
		t.Fatalf("wsjson.Read() returned an unexpected error: %v", err)
	}
	return message
}

func startTestServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Skipf("local sockets are unavailable: %v", err)
	}
	server := httptest.NewUnstartedServer(handler)
	server.Listener = listener
	server.Start()
	return server
}

func websocketURL(serverURL string) string {
	return "ws" + strings.TrimPrefix(serverURL, "http")
}

func findPlayer(snapshot *protocol.WorldSnapshot, id string) (protocol.PlayerSnapshot, bool) {
	for _, player := range snapshot.Players {
		if player.ID == id {
			return player, true
		}
	}
	return protocol.PlayerSnapshot{}, false
}

func findTarget(snapshot *protocol.WorldSnapshot, id string) (protocol.TargetSnapshot, bool) {
	for _, target := range snapshot.Targets {
		if target.ID == id {
			return target, true
		}
	}
	return protocol.TargetSnapshot{}, false
}
