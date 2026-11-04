package simulation

import (
	"fmt"
	"math"
	"sync"
	"time"
)

type PlayerID string

// Position is a precise point on the horizontal world plane.
type Position struct {
	X float64
	Z float64
}

// PlayerState is a read-only snapshot of a player in the simulation.
type PlayerState struct {
	ID          PlayerID
	Position    Position
	Destination *Position
	Speed       float64
}

// World owns the active gameplay state. Gameplay state changes happen in Step.
type World struct {
	mu           sync.RWMutex
	tick         uint64
	players      map[PlayerID]*PlayerState
	destinations map[PlayerID]Position
}

func NewWorld() *World {
	return &World{
		players:      make(map[PlayerID]*PlayerState),
		destinations: make(map[PlayerID]Position),
	}
}

// AddPlayer registers a player before it enters the active simulation.
func (w *World) AddPlayer(id PlayerID, position Position, speed float64) error {
	if id == "" {
		return fmt.Errorf("player ID is required")
	}
	if !position.valid() {
		return fmt.Errorf("player position must be finite")
	}
	if math.IsNaN(speed) || math.IsInf(speed, 0) || speed <= 0 {
		return fmt.Errorf("player speed must be finite and positive")
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	if _, exists := w.players[id]; exists {
		return fmt.Errorf("player %q already exists", id)
	}

	w.players[id] = &PlayerState{ID: id, Position: position, Speed: speed}
	return nil
}

// SetDestination records the latest click-to-move intention for a player.
// The intention is applied to gameplay state on the next Step.
func (w *World) SetDestination(id PlayerID, destination Position) error {
	if !destination.valid() {
		return fmt.Errorf("destination must be finite")
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	if _, exists := w.players[id]; !exists {
		return fmt.Errorf("player %q does not exist", id)
	}

	w.destinations[id] = destination
	return nil
}

// Step advances all active gameplay state by one fixed simulation step.
func (w *World) Step(elapsed time.Duration) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.tick++
	if elapsed <= 0 {
		return
	}

	for id, destination := range w.destinations {
		player := w.players[id]
		player.Destination = positionPointer(destination)
		delete(w.destinations, id)
	}

	seconds := elapsed.Seconds()
	for _, player := range w.players {
		movePlayer(player, seconds)
	}
}

// Player returns a copy of the current player state.
func (w *World) Player(id PlayerID) (PlayerState, bool) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	player, exists := w.players[id]
	if !exists {
		return PlayerState{}, false
	}

	snapshot := *player
	if player.Destination != nil {
		snapshot.Destination = positionPointer(*player.Destination)
	}
	return snapshot, true
}

// Tick returns the number of completed simulation steps.
func (w *World) Tick() uint64 {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.tick
}

func movePlayer(player *PlayerState, seconds float64) {
	if player.Destination == nil {
		return
	}

	dx := player.Destination.X - player.Position.X
	dz := player.Destination.Z - player.Position.Z
	distance := math.Hypot(dx, dz)
	maximumMovement := player.Speed * seconds

	if distance <= maximumMovement {
		player.Position = *player.Destination
		player.Destination = nil
		return
	}

	player.Position.X += dx / distance * maximumMovement
	player.Position.Z += dz / distance * maximumMovement
}

func (p Position) valid() bool {
	return !math.IsNaN(p.X) && !math.IsInf(p.X, 0) &&
		!math.IsNaN(p.Z) && !math.IsInf(p.Z, 0)
}

func positionPointer(position Position) *Position {
	return &position
}
