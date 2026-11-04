package simulation

import (
	"math"
	"testing"
	"time"
)

func TestWorldMovesPlayerContinuouslyTowardDestination(t *testing.T) {
	world := NewWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	if err := world.SetDestination("player-one", Position{X: 3, Z: 4}); err != nil {
		t.Fatalf("SetDestination() returned an unexpected error: %v", err)
	}

	world.Step(500 * time.Millisecond)

	player := mustPlayer(t, world, "player-one")
	assertPosition(t, player.Position, Position{X: 1.2, Z: 1.6})
	if player.Destination == nil {
		t.Fatal("Destination is nil before the player arrives")
	}
}

func TestWorldStopsPlayerAtDestinationWithoutOvershooting(t *testing.T) {
	world := NewWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	if err := world.SetDestination("player-one", Position{X: 1}); err != nil {
		t.Fatalf("SetDestination() returned an unexpected error: %v", err)
	}

	world.Step(time.Second)

	player := mustPlayer(t, world, "player-one")
	assertPosition(t, player.Position, Position{X: 1})
	if player.Destination != nil {
		t.Fatalf("Destination = %+v after arrival, want nil", player.Destination)
	}
}

func TestWorldUsesLatestDestinationIntention(t *testing.T) {
	world := NewWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 2)
	if err := world.SetDestination("player-one", Position{X: 10}); err != nil {
		t.Fatalf("first SetDestination() returned an unexpected error: %v", err)
	}
	if err := world.SetDestination("player-one", Position{Z: 10}); err != nil {
		t.Fatalf("second SetDestination() returned an unexpected error: %v", err)
	}

	world.Step(500 * time.Millisecond)

	player := mustPlayer(t, world, "player-one")
	assertPosition(t, player.Position, Position{Z: 1})
}

func TestWorldRejectsInvalidMovementInput(t *testing.T) {
	world := NewWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 2)

	if err := world.SetDestination("missing", Position{}); err == nil {
		t.Fatal("SetDestination() error = nil for missing player")
	}
	if err := world.SetDestination("player-one", Position{X: math.NaN()}); err == nil {
		t.Fatal("SetDestination() error = nil for non-finite destination")
	}
}

func TestWorldStepAdvancesTick(t *testing.T) {
	world := NewWorld()

	world.Step(time.Second)
	world.Step(time.Second)

	if got, want := world.Tick(), uint64(2); got != want {
		t.Fatalf("Tick() = %d, want %d", got, want)
	}
}

func mustAddPlayer(t *testing.T, world *World, id PlayerID, position Position, speed float64) {
	t.Helper()
	if err := world.AddPlayer(id, position, speed); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
}

func mustPlayer(t *testing.T, world *World, id PlayerID) PlayerState {
	t.Helper()
	player, exists := world.Player(id)
	if !exists {
		t.Fatalf("Player(%q) does not exist", id)
	}
	return player
}

func assertPosition(t *testing.T, got, want Position) {
	t.Helper()
	const tolerance = 0.000_001
	if math.Abs(got.X-want.X) > tolerance || math.Abs(got.Z-want.Z) > tolerance {
		t.Fatalf("position = %+v, want %+v", got, want)
	}
}
