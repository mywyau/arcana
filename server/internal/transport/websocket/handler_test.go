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
		Dummy: simulation.DummySnapshot{
			TargetID:        "player-one",
			NextAttackPulse: 16,
			Defeated:        true,
			RespawnAtPulse:  20,
		},
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
	if message.Dummy.TargetID != "player-one" || message.Dummy.NextAttackPulse != 16 ||
		!message.Dummy.Defeated || message.Dummy.RespawnAtPulse != 20 {
		t.Fatalf("dummy combat state = %+v", message.Dummy)
	}
	player := message.Players[0]
	if player.Health != 40 || player.MaximumHealth != 100 || !player.Defeated || player.RespawnAtPulse != 20 ||
		player.EquippedWeapon != "dagger" || player.LastEquipmentSequence != 3 || player.LastEquipmentPulse != 11 {
		t.Fatalf("player vital state = %+v", player)
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
