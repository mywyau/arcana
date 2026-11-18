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

// snapshotDeltaTracker owns the public baseline for one connection. For
// private state it stores revision numbers only, never container contents.
type snapshotDeltaTracker struct {
	players                 map[string]protocol.PlayerSnapshot
	targets                 map[string]protocol.TargetSnapshot
	lastFull                time.Time
	privateStateInitialized bool
	lastInventoryRevision   uint64
	lastEquipmentRevision   uint64
	lastBankRevision        uint64
	lastAmmunitionRevision  uint64
	lastRingRackRevision    uint64
	lastRunePouchRevision   uint64
	lastSkillsRevision      uint64
	lastCombatBoostRevision uint64
	lastBrewingRevision     uint64
	lastFletchingRevision   uint64
	lastGatheringRevision   uint64
}

// applyPrivateState removes unchanged private containers from the local
// player's otherwise complete snapshot. Public equipment remains present so
// the local and nearby player models can still be rendered from every frame.
func (t *snapshotDeltaTracker) applyPrivateState(
	message *protocol.WorldSnapshot,
	localPlayerID string,
	full bool,
) {
	for index := range message.Players {
		player := &message.Players[index]
		if player.ID != localPlayerID {
			continue
		}

		inventoryChanged := full || !t.privateStateInitialized ||
			player.InventoryRevision != t.lastInventoryRevision
		if !inventoryChanged {
			player.Inventory = nil
		}

		equipmentChanged := full || !t.privateStateInitialized ||
			player.EquipmentRevision != t.lastEquipmentRevision
		if !equipmentChanged {
			player.EquippedWeaponInstanceID = ""
			player.EquippedWeaponReinforcement = 0
			player.EquippedArmourInstances = nil
			player.EquippedArmourReinforcements = nil
			player.EquippedSpellbook = ""
			player.ActiveSpellbook = ""
		}

		bankChanged := full || !t.privateStateInitialized ||
			player.BankRevision != t.lastBankRevision
		if !bankChanged {
			player.Bank = nil
		}

		ammunitionChanged := full || !t.privateStateInitialized ||
			player.AmmunitionPouchRevision != t.lastAmmunitionRevision
		if !ammunitionChanged {
			player.AmmunitionPouch = nil
		}

		ringRackChanged := full || !t.privateStateInitialized ||
			player.RingRackRevision != t.lastRingRackRevision
		if !ringRackChanged {
			player.RingRack = nil
		}

		runePouchChanged := full || !t.privateStateInitialized ||
			player.RunePouchRevision != t.lastRunePouchRevision
		if !runePouchChanged {
			player.RunePouch = nil
		}

		skillsChanged := full || !t.privateStateInitialized ||
			player.SkillsRevision != t.lastSkillsRevision
		if !skillsChanged {
			player.Skills = nil
		}

		combatBoostsChanged := full || !t.privateStateInitialized ||
			player.CombatBoostsRevision != t.lastCombatBoostRevision
		if !combatBoostsChanged {
			player.ActiveCombatBoosts = nil
		}

		brewingChanged := full || !t.privateStateInitialized ||
			player.BrewingRevision != t.lastBrewingRevision
		if !brewingChanged {
			player.LastBrewSequence = 0
			player.BrewRemaining = 0
			player.NextBrewPulse = 0
			player.BrewingResult = ""
		}

		fletchingChanged := full || !t.privateStateInitialized ||
			player.FletchingRevision != t.lastFletchingRevision
		if !fletchingChanged {
			player.LastFletchSequence = 0
			player.FletchRemaining = 0
			player.NextFletchPulse = 0
			player.FletchingResult = ""
		}

		gatheringChanged := full || !t.privateStateInitialized ||
			player.GatheringRevision != t.lastGatheringRevision
		if !gatheringChanged {
			player.LastGatherSequence = 0
			player.GatheringSpot = ""
			player.GatheringResult = ""
			player.NextGatherPulse = 0
		}

		t.lastInventoryRevision = player.InventoryRevision
		t.lastEquipmentRevision = player.EquipmentRevision
		t.lastBankRevision = player.BankRevision
		t.lastAmmunitionRevision = player.AmmunitionPouchRevision
		t.lastRingRackRevision = player.RingRackRevision
		t.lastRunePouchRevision = player.RunePouchRevision
		t.lastSkillsRevision = player.SkillsRevision
		t.lastCombatBoostRevision = player.CombatBoostsRevision
		t.lastBrewingRevision = player.BrewingRevision
		t.lastFletchingRevision = player.FletchingRevision
		t.lastGatheringRevision = player.GatheringRevision
		t.privateStateInitialized = true
		return
	}
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
	if full || previous.TelegraphType != current.TelegraphType {
		update.TelegraphType = pointer(current.TelegraphType)
		changed = true
	}
	if full || previous.TelegraphPosition != current.TelegraphPosition {
		update.TelegraphPosition = pointer(current.TelegraphPosition)
		changed = true
	}
	if full || previous.TelegraphRadius != current.TelegraphRadius {
		update.TelegraphRadius = pointer(current.TelegraphRadius)
		changed = true
	}
	if full || previous.TelegraphImpactPulse != current.TelegraphImpactPulse {
		update.TelegraphImpactPulse = pointer(current.TelegraphImpactPulse)
		changed = true
	}
	return update, changed
}

func pointer[T any](value T) *T {
	return &value
}
