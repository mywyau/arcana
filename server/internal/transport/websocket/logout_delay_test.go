package websockettransport

import (
	"context"
	"testing"
	"time"

	"github.com/mywyau/arcana/server/internal/simulation"
)

func TestDelayedLogoutKeepsCharacterThenRemovesIt(t *testing.T) {
	world, handler, stopRunner := newLogoutTestHandler(t)
	defer stopRunner()
	const playerID simulation.PlayerID = "player-one"
	if err := world.AddPlayer(playerID, simulation.Position{}, simulation.DefaultPlayerMovementSpeed); err != nil {
		t.Fatalf("AddPlayer() error = %v", err)
	}
	handler.disconnectedLogoutDelay = 40 * time.Millisecond

	handler.scheduleDelayedLogout(playerID)
	if _, found := world.Player(playerID); !found {
		t.Fatal("player left immediately, want character to remain during logout delay")
	}

	waitForPlayerToLeave(t, world, playerID)
}

func TestReconnectCancelsDelayedLogout(t *testing.T) {
	world, handler, stopRunner := newLogoutTestHandler(t)
	defer stopRunner()
	const playerID simulation.PlayerID = "player-one"
	if err := world.AddPlayer(playerID, simulation.Position{}, simulation.DefaultPlayerMovementSpeed); err != nil {
		t.Fatalf("AddPlayer() error = %v", err)
	}
	handler.disconnectedLogoutDelay = 25 * time.Millisecond

	handler.scheduleDelayedLogout(playerID)
	if cancelled := handler.cancelPendingLogout(playerID); !cancelled {
		t.Fatal("cancelPendingLogout() = false, want true")
	}
	if resumed := world.ResumePlayer(playerID); !resumed {
		t.Fatal("ResumePlayer() = false, want true")
	}
	time.Sleep(2 * handler.disconnectedLogoutDelay)

	if _, found := world.Player(playerID); !found {
		t.Fatal("player left after reconnect cancelled delayed logout")
	}
}

func TestExplicitLogoutRemovesCharacterImmediately(t *testing.T) {
	world, handler, stopRunner := newLogoutTestHandler(t)
	defer stopRunner()
	const playerID simulation.PlayerID = "player-one"
	if err := world.AddPlayer(playerID, simulation.Position{}, simulation.DefaultPlayerMovementSpeed); err != nil {
		t.Fatalf("AddPlayer() error = %v", err)
	}

	handler.removeCharacterNow(playerID)

	if _, found := world.Player(playerID); found {
		t.Fatal("player remains after explicit logout")
	}
}

func newLogoutTestHandler(t *testing.T) (*simulation.World, *Handler, context.CancelFunc) {
	t.Helper()
	world := simulation.NewWorld()
	const interval = time.Millisecond
	runner, err := simulation.NewRunner(world, interval)
	if err != nil {
		t.Fatalf("NewRunner() error = %v", err)
	}
	runnerCtx, stopRunner := context.WithCancel(context.Background())
	go runner.Run(runnerCtx)
	handler, err := NewHandler(world, interval)
	if err != nil {
		stopRunner()
		t.Fatalf("NewHandler() error = %v", err)
	}
	return world, handler, stopRunner
}
