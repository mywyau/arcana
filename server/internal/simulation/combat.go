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
	playerMaximumHealth       = 100
	dummyMaximumHealth        = 100
	dummyAttackDamage         = 15
	dummyAttackIntervalPulses = 4
	dummyAttackRange          = 2.0
	dummyLeashRange           = 6.0
	respawnDelayPulses        = 5
	dummyTargetID             = "dummy"
	combatEventHistoryLimit   = 32
)

type CombatEventType string

const CombatEventDamage CombatEventType = "damage"

// CombatEvent describes an authoritative outcome for client presentation.
type CombatEvent struct {
	ID       uint64
	Pulse    uint64
	Type     CombatEventType
	SourceID string
	TargetID string
	Amount   int
}

type TargetCommand struct {
	PlayerID PlayerID
	Sequence uint64
	TargetID string
}

type DummySnapshot struct {
	Position        Position
	Health          int
	MaximumHealth   int
	TargetID        PlayerID
	NextAttackPulse uint64
	Defeated        bool
	RespawnAtPulse  uint64
}

type combatState struct {
	elapsed        time.Duration
	pulse          uint64
	dummy          DummySnapshot
	pendingTargets map[PlayerID]TargetCommand
	nextEventID    uint64
	events         []CombatEvent
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
	if player.defeated {
		return fmt.Errorf("player %q is defeated", command.PlayerID)
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
		w.respawnDefeatedEntities()
		w.applyTargetCommands()
		w.resolveAutoAttacks()
		w.resolveDummyAttack()
	}
}

func (w *World) respawnDefeatedEntities() {
	if w.combat.dummy.Defeated && w.combat.pulse >= w.combat.dummy.RespawnAtPulse {
		w.combat.dummy.Health = w.combat.dummy.MaximumHealth
		w.combat.dummy.Defeated = false
		w.combat.dummy.RespawnAtPulse = 0
	}
	for _, player := range w.players {
		if !player.defeated || w.combat.pulse < player.respawnAtPulse {
			continue
		}
		player.position = player.spawnPosition
		player.health = player.maximumHealth
		player.defeated = false
		player.respawnAtPulse = 0
		player.lastActionPulse = w.combat.pulse
		player.lastActionResult = "respawned"
		if location, err := w.locate(player.position); err == nil {
			player.navigationLocation = location
		}
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
		if !exists || player.defeated {
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
			if w.combat.dummy.TargetID == id {
				w.clearDummyTarget()
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
		if player.defeated || player.combatTarget != dummyTargetID {
			continue
		}
		if w.combat.dummy.Defeated {
			w.clearPlayerTarget(player, "target_defeated")
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
		w.combat.recordEvent(CombatEvent{
			Pulse:    w.combat.pulse,
			Type:     CombatEventDamage,
			SourceID: string(player.id),
			TargetID: dummyTargetID,
			Amount:   swordDamage,
		})
		if w.combat.dummy.TargetID == "" {
			w.combat.dummy.TargetID = player.id
			w.combat.dummy.NextAttackPulse = w.combat.pulse + dummyAttackIntervalPulses
		}
		if w.combat.dummy.Health == 0 {
			w.defeatDummy()
		}
	}
}

func (w *World) resolveDummyAttack() {
	dummy := &w.combat.dummy
	if dummy.Defeated || dummy.TargetID == "" {
		return
	}
	player, exists := w.players[dummy.TargetID]
	if !exists || player.defeated || player.combatTarget != dummyTargetID {
		w.clearDummyTarget()
		return
	}
	distanceToPlayer := distance(dummy.Position, player.position)
	if distanceToPlayer > dummyLeashRange {
		w.clearDummyTarget()
		return
	}
	if distanceToPlayer > dummyAttackRange || w.combat.pulse < dummy.NextAttackPulse {
		return
	}

	player.health = max(0, player.health-dummyAttackDamage)
	dummy.NextAttackPulse = w.combat.pulse + dummyAttackIntervalPulses
	w.combat.recordEvent(CombatEvent{
		Pulse:    w.combat.pulse,
		Type:     CombatEventDamage,
		SourceID: dummyTargetID,
		TargetID: string(player.id),
		Amount:   dummyAttackDamage,
	})
	if player.health == 0 {
		w.defeatPlayer(player)
	}
}

func (w *World) defeatPlayer(player *playerState) {
	player.defeated = true
	player.respawnAtPulse = w.combat.pulse + respawnDelayPulses
	player.lastActionPulse = w.combat.pulse
	player.lastActionResult = "defeated"
	player.combatTarget = ""
	player.nextAttackPulse = 0
	player.autoApproaching = false
	stopPlayer(player)
	delete(w.pendingMoves, player.id)
	delete(w.combat.pendingTargets, player.id)
	if w.combat.dummy.TargetID == player.id {
		w.clearDummyTarget()
	}
}

func (w *World) defeatDummy() {
	w.combat.dummy.Defeated = true
	w.combat.dummy.RespawnAtPulse = w.combat.pulse + respawnDelayPulses
	w.clearDummyTarget()
	for _, player := range w.players {
		if player.combatTarget == dummyTargetID {
			w.clearPlayerTarget(player, "target_defeated")
		}
	}
}

func (w *World) clearPlayerTarget(player *playerState, result string) {
	player.combatTarget = ""
	player.nextAttackPulse = 0
	player.lastActionPulse = w.combat.pulse
	player.lastActionResult = result
	if player.autoApproaching {
		stopPlayer(player)
		player.autoApproaching = false
	}
}

func (w *World) clearDummyTarget() {
	w.combat.dummy.TargetID = ""
	w.combat.dummy.NextAttackPulse = 0
}

func (combat *combatState) recordEvent(event CombatEvent) {
	combat.nextEventID++
	event.ID = combat.nextEventID
	combat.events = append(combat.events, event)
	if len(combat.events) > combatEventHistoryLimit {
		combat.events = append([]CombatEvent(nil), combat.events[len(combat.events)-combatEventHistoryLimit:]...)
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
