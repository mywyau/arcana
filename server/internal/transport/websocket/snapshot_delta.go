package websockettransport

import (
	"maps"
	"sort"
	"time"

	"github.com/mywyau/arcana/server/internal/protocol"
)

const fullReconciliationInterval = 10 * time.Second

type snapshotDeltaStats struct {
	Full           bool
	PlayerUpdates  int
	TargetUpdates  int
	PlayerRemovals int
	TargetRemovals int
}

// snapshotDeltaTracker owns the public baseline for one connection. It never
// stores the receiving player's private state.
type snapshotDeltaTracker struct {
	players  map[string]protocol.PlayerSnapshot
	targets  map[string]protocol.TargetSnapshot
	lastFull time.Time
}

func newSnapshotDeltaTracker() *snapshotDeltaTracker {
	return &snapshotDeltaTracker{
		players: make(map[string]protocol.PlayerSnapshot),
		targets: make(map[string]protocol.TargetSnapshot),
	}
}

func (t *snapshotDeltaTracker) apply(
	message *protocol.WorldSnapshot,
	localPlayerID string,
	now time.Time,
) snapshotDeltaStats {
	full := t.lastFull.IsZero() || now.Sub(t.lastFull) >= fullReconciliationInterval
	stats := snapshotDeltaStats{Full: full}
	message.FullReconciliation = full

	currentPlayers := make(map[string]protocol.PlayerSnapshot, len(message.Players))
	localPlayers := make([]protocol.PlayerSnapshot, 0, 1)
	for _, player := range message.Players {
		if player.ID == localPlayerID {
			localPlayers = append(localPlayers, player)
			continue
		}
		currentPlayers[player.ID] = player
		previous, found := t.players[player.ID]
		if update, changed := playerUpdate(previous, player, full || !found); changed {
			message.PlayerUpdates = append(message.PlayerUpdates, update)
			stats.PlayerUpdates++
		}
	}
	for playerID := range t.players {
		if _, found := currentPlayers[playerID]; found {
			continue
		}
		message.RemovedPlayerIDs = append(message.RemovedPlayerIDs, playerID)
		stats.PlayerRemovals++
	}
	sort.Strings(message.RemovedPlayerIDs)
	message.Players = localPlayers
	t.players = currentPlayers

	currentTargets := make(map[string]protocol.TargetSnapshot, len(message.Targets))
	for _, target := range message.Targets {
		currentTargets[target.ID] = target
		previous, found := t.targets[target.ID]
		if update, changed := targetUpdate(previous, target, full || !found); changed {
			message.TargetUpdates = append(message.TargetUpdates, update)
			stats.TargetUpdates++
		}
	}
	for targetID := range t.targets {
		if _, found := currentTargets[targetID]; found {
			continue
		}
		message.RemovedTargetIDs = append(message.RemovedTargetIDs, targetID)
		stats.TargetRemovals++
	}
	sort.Strings(message.RemovedTargetIDs)
	message.Targets = nil
	t.targets = currentTargets

	if full {
		t.lastFull = now
	}
	return stats
}

func playerUpdate(
	previous, current protocol.PlayerSnapshot,
	full bool,
) (protocol.PlayerUpdate, bool) {
	update := protocol.PlayerUpdate{ID: current.ID}
	changed := false
	if full || previous.Position != current.Position {
		update.Position = pointer(current.Position)
		changed = true
	}
	if full || previous.Speed != current.Speed {
		update.Speed = pointer(current.Speed)
		changed = true
	}
	if full || previous.CombatTarget != current.CombatTarget {
		update.CombatTarget = pointer(current.CombatTarget)
		changed = true
	}
	if full || previous.Health != current.Health {
		update.Health = pointer(current.Health)
		changed = true
	}
	if full || previous.MaximumHealth != current.MaximumHealth {
		update.MaximumHealth = pointer(current.MaximumHealth)
		changed = true
	}
	if full || previous.Defeated != current.Defeated {
		update.Defeated = pointer(current.Defeated)
		changed = true
	}
	if full || previous.RespawnAtPulse != current.RespawnAtPulse {
		update.RespawnAtPulse = pointer(current.RespawnAtPulse)
		changed = true
	}
	if full || previous.EquippedWeapon != current.EquippedWeapon {
		update.EquippedWeapon = pointer(current.EquippedWeapon)
		changed = true
	}
	if full || !maps.Equal(previous.EquippedArmour, current.EquippedArmour) {
		armour := maps.Clone(current.EquippedArmour)
		update.EquippedArmour = &armour
		changed = true
	}
	if full || previous.CombatLevel != current.CombatLevel {
		update.CombatLevel = pointer(current.CombatLevel)
		changed = true
	}
	return update, changed
}

func targetUpdate(
	previous, current protocol.TargetSnapshot,
	full bool,
) (protocol.TargetUpdate, bool) {
	update := protocol.TargetUpdate{ID: current.ID}
	changed := false
	if full || previous.Name != current.Name {
		update.Name = pointer(current.Name)
		changed = true
	}
	if full || previous.Kind != current.Kind {
		update.Kind = pointer(current.Kind)
		changed = true
	}
	if full || previous.Position != current.Position {
		update.Position = pointer(current.Position)
		changed = true
	}
	if full || previous.Health != current.Health {
		update.Health = pointer(current.Health)
		changed = true
	}
	if full || previous.MaximumHealth != current.MaximumHealth {
		update.MaximumHealth = pointer(current.MaximumHealth)
		changed = true
	}
	if full || previous.TargetID != current.TargetID {
		update.TargetID = pointer(current.TargetID)
		changed = true
	}
	if full || previous.NextAttackPulse != current.NextAttackPulse {
		update.NextAttackPulse = pointer(current.NextAttackPulse)
		changed = true
	}
	if full || previous.Defeated != current.Defeated {
		update.Defeated = pointer(current.Defeated)
		changed = true
	}
	if full || previous.RespawnAtPulse != current.RespawnAtPulse {
		update.RespawnAtPulse = pointer(current.RespawnAtPulse)
		changed = true
	}
	if full || previous.DodgeRate != current.DodgeRate {
		update.DodgeRate = pointer(current.DodgeRate)
		changed = true
	}
	if full || previous.Accuracy != current.Accuracy {
		update.Accuracy = pointer(current.Accuracy)
		changed = true
	}
	if full || previous.Evasion != current.Evasion {
		update.Evasion = pointer(current.Evasion)
		changed = true
	}
	if full || previous.AttackType != current.AttackType {
		update.AttackType = pointer(current.AttackType)
		changed = true
	}
	if full || previous.CombatLevel != current.CombatLevel {
		update.CombatLevel = pointer(current.CombatLevel)
		changed = true
	}
	return update, changed
}

func pointer[T any](value T) *T {
	return &value
}
