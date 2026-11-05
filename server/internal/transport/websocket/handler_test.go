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

	welcome := readServerMessage(t, ctx, connection)
	if welcome.Type != protocol.MessageWelcome || welcome.PlayerID == "" {
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

func TestSnapshotMessageIncludesCombatEvents(t *testing.T) {
	message := snapshotMessage(simulation.Snapshot{
		Targets: []simulation.TargetSnapshot{
			{
				ID:              "dummy",
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
			ID:                    "player-one",
			Health:                40,
			MaximumHealth:         100,
			Defeated:              true,
			RespawnAtPulse:        20,
			EquippedWeapon:        simulation.WeaponDagger,
			LastEquipmentSequence: 3,
			LastEquipmentPulse:    11,
			DodgeRate:             0.15,
			Inventory: simulation.InventorySnapshot{
				Capacity: 8,
				Items: []simulation.InventoryItem{{
					Slot: 0, ItemID: "sword", Quantity: 1,
				}},
			},
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
	if !found || dummy.TargetID != "player-one" || dummy.NextAttackPulse != 16 ||
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
	if player.Health != 40 || player.MaximumHealth != 100 || !player.Defeated || player.RespawnAtPulse != 20 ||
		player.EquippedWeapon != "dagger" || player.LastEquipmentSequence != 3 || player.LastEquipmentPulse != 11 ||
		player.DodgeRate != 0.15 {
		t.Fatalf("player vital state = %+v", player)
	}
	if player.Inventory.Capacity != 8 || len(player.Inventory.Items) != 1 ||
		player.Inventory.Items[0].Slot != 0 || player.Inventory.Items[0].ItemID != "sword" ||
		player.Inventory.Items[0].Quantity != 1 {
		t.Fatalf("player inventory = %+v", player.Inventory)
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
