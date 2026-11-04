package simulation

import (
	"fmt"
	"math"
	"sort"
	"time"
)

const (
	actionPulseInterval       = 600 * time.Millisecond
	swordAttackIntervalPulses = 4
	swordRange                = 2.0
	swordApproachRange        = 1.5
	swordDamage               = 10
	dummyMaximumHealth        = 100
	dummyTargetID             = "dummy"
)

type TargetCommand struct {
	PlayerID PlayerID
	Sequence uint64
	TargetID string
}

type DummySnapshot struct {
	Position      Position
	Health        int
	MaximumHealth int
}

type combatState struct {
	elapsed        time.Duration
	pulse          uint64
	dummy          DummySnapshot
	pendingTargets map[PlayerID]TargetCommand
}

func newCombatState() *combatState {
	return &combatState{
		dummy: DummySnapshot{
			Position:      Position{X: -4, Z: -2},
			Health:        dummyMaximumHealth,
			MaximumHealth: dummyMaximumHealth,
		},
		pendingTargets: make(map[PlayerID]TargetCommand),
	}
}

// QueueTarget records an intention to engage TargetID, or to disengage when it
// is empty. Target changes are applied on the next action pulse.
func (w *World) QueueTarget(command TargetCommand) error {
	if command.Sequence == 0 {
		return fmt.Errorf("action sequence must be positive")
	}
	if command.TargetID != "" && command.TargetID != dummyTargetID {
		return fmt.Errorf("target %q does not exist", command.TargetID)
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	if _, leaving := w.pendingLeaves[command.PlayerID]; leaving {
		return fmt.Errorf("player %q is leaving", command.PlayerID)
	}
	player, exists := w.players[command.PlayerID]
	if !exists {
		player, exists = w.pendingJoins[command.PlayerID]
		if !exists {
			return fmt.Errorf("player %q does not exist", command.PlayerID)
		}
	}
	if command.Sequence <= player.lastActionSequence {
		return fmt.Errorf("action sequence %d is not newer than %d", command.Sequence, player.lastActionSequence)
	}
	if pending, exists := w.combat.pendingTargets[command.PlayerID]; exists && command.Sequence <= pending.Sequence {
		return fmt.Errorf("action sequence %d is not newer than pending sequence %d", command.Sequence, pending.Sequence)
	}

	w.combat.pendingTargets[command.PlayerID] = command
	return nil
}

func (w *World) advanceCombat(elapsed time.Duration) {
	w.combat.elapsed += elapsed
	for w.combat.elapsed >= actionPulseInterval {
		w.combat.elapsed -= actionPulseInterval
		w.combat.pulse++
		w.applyTargetCommands()
		w.resolveAutoAttacks()
	}
}

func (w *World) applyTargetCommands() {
	playerIDs := make([]PlayerID, 0, len(w.combat.pendingTargets))
	for id := range w.combat.pendingTargets {
		playerIDs = append(playerIDs, id)
	}
	sort.Slice(playerIDs, func(i, j int) bool { return playerIDs[i] < playerIDs[j] })

	for _, id := range playerIDs {
		command := w.combat.pendingTargets[id]
		player, exists := w.players[id]
		if !exists {
			delete(w.combat.pendingTargets, id)
			continue
		}

		player.lastActionSequence = command.Sequence
		player.lastActionPulse = w.combat.pulse
		player.combatTarget = command.TargetID
		if command.TargetID == "" {
			player.lastActionResult = "disengaged"
			player.nextAttackPulse = 0
			if player.autoApproaching {
				stopPlayer(player)
			}
		} else {
			player.lastActionResult = "engaged"
			player.nextAttackPulse = w.combat.pulse
		}
		player.autoApproaching = false
		delete(w.combat.pendingTargets, id)
	}
}

func (w *World) resolveAutoAttacks() {
	playerIDs := make([]PlayerID, 0, len(w.players))
	for id := range w.players {
		playerIDs = append(playerIDs, id)
	}
	sort.Slice(playerIDs, func(i, j int) bool { return playerIDs[i] < playerIDs[j] })

	for _, id := range playerIDs {
		player := w.players[id]
		if player.combatTarget != dummyTargetID {
			continue
		}
		if w.combat.dummy.Health <= 0 {
			player.lastActionResult = "target_defeated"
			if player.autoApproaching {
				stopPlayer(player)
				player.autoApproaching = false
			}
			continue
		}
		if distance(player.position, w.combat.dummy.Position) > swordRange {
			if err := w.approachDummy(player); err != nil {
				player.lastActionResult = "target_unreachable"
				player.autoApproaching = false
			} else {
				player.lastActionResult = "approaching"
			}
			continue
		}
		if player.autoApproaching {
			stopPlayer(player)
			player.autoApproaching = false
		}
		if w.combat.pulse < player.nextAttackPulse {
			continue
		}

		w.combat.dummy.Health = max(0, w.combat.dummy.Health-swordDamage)
		player.lastActionPulse = w.combat.pulse
		player.lastActionResult = "hit"
		player.nextAttackPulse = w.combat.pulse + swordAttackIntervalPulses
	}
}

func (w *World) approachDummy(player *playerState) error {
	dx := player.position.X - w.combat.dummy.Position.X
	dz := player.position.Z - w.combat.dummy.Position.Z
	length := math.Hypot(dx, dz)
	if length == 0 {
		dx = 1
		length = 1
	}
	destination := Position{
		X: w.combat.dummy.Position.X + dx/length*swordApproachRange,
		Y: w.combat.dummy.Position.Y,
		Z: w.combat.dummy.Position.Z + dz/length*swordApproachRange,
	}
	path := []Position{destination}
	if w.pathfinder != nil {
		var err error
		path, err = w.pathfinder.FindPath(player.position, destination)
		if err != nil {
			return fmt.Errorf("find combat path: %w", err)
		}
	}
	if len(path) == 0 {
		return fmt.Errorf("find combat path: empty path")
	}

	player.destination = positionPointer(destination)
	player.path = clonePositions(path)
	player.autoApproaching = true
	return nil
}

func stopPlayer(player *playerState) {
	player.destination = nil
	player.path = nil
}

func distance(a, b Position) float64 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	dz := a.Z - b.Z
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}
