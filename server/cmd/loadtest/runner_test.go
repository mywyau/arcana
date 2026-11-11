package main

import (
	"context"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mywyau/arcana/server/internal/simulation"
	websockettransport "github.com/mywyau/arcana/server/internal/transport/websocket"
)

func TestValidateEndpointRequiresExplicitRemotePermission(t *testing.T) {
	for _, endpoint := range []string{
		"ws://localhost:8080/ws",
		"ws://127.0.0.1:8080/ws",
		"ws://[::1]:8080/ws",
	} {
		if err := validateEndpoint(endpoint, false); err != nil {
			t.Errorf("validateEndpoint(%q) error = %v", endpoint, err)
		}
	}
	if err := validateEndpoint("wss://game.example.com/ws", false); err == nil {
		t.Fatal("remote endpoint was accepted without allow-remote")
	}
	if err := validateEndpoint("wss://game.example.com/ws", true); err != nil {
		t.Fatalf("explicitly allowed remote endpoint error = %v", err)
	}
}

func TestMovementDestinationsStayWithinPrototypeMap(t *testing.T) {
	for _, scenario := range []string{scenarioClustered, scenarioSpread} {
		for clientID := 0; clientID < 100; clientID++ {
			random := rand.New(rand.NewSource(int64(clientID + 1)))
			destination := movementDestination(scenario, clientID, 100, random)
			if destination.X < -50 || destination.X > 50 || destination.Z < -50 || destination.Z > 50 {
				t.Fatalf("%s destination for client %d = %+v", scenario, clientID, destination)
			}
		}
	}
}

func TestRunLoadTestConnectsAndReadsSnapshots(t *testing.T) {
	world := simulation.NewWorld()
	handler, err := websockettransport.NewHandler(world, 5*time.Millisecond)
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Skipf("local sockets are unavailable: %v", err)
	}
	server := httptest.NewUnstartedServer(handler)
	server.Listener = listener
	server.Start()
	defer server.Close()

	report := runLoadTest(context.Background(), loadTestConfig{
		URL:     strings.Replace(server.URL, "http://", "ws://", 1),
		Clients: 2, RampDuration: 10 * time.Millisecond, HoldDuration: 50 * time.Millisecond,
		Scenario: scenarioClustered,
	}, log.New(io.Discard, "", 0))
	if report.Connected != 2 || report.ConnectFailures != 0 || report.UnexpectedDisconnects != 0 {
		t.Fatalf("load test report = %+v", report)
	}
	if report.Snapshots == 0 || report.Bytes == 0 {
		t.Fatalf("load test did not read snapshots: %+v", report)
	}
}

func TestLoadTestReportIncludesRates(t *testing.T) {
	report := loadTestReport{
		Elapsed: 2 * time.Second, Started: 2, Connected: 2,
		Messages: 20, Snapshots: 18, Bytes: 2048,
	}
	text := report.String()
	for _, expected := range []string{"connected=2", "messages_per_second=10.0", "received=2.0KiB", "received_per_second=1.0KiB"} {
		if !strings.Contains(text, expected) {
			t.Fatalf("report %q does not include %q", text, expected)
		}
	}
}
