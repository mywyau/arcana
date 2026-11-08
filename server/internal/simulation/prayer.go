package simulation

import (
	"fmt"
	"sort"
)

type PrayerID string
type PrayerSetID string

const (
	PrayerEvasion        PrayerID    = "evasion"
	PrayerStrength       PrayerID    = "strength"
	PrayerAccuracy       PrayerID    = "accuracy"
	PrayerDefence        PrayerID    = "defence"
	PrayerRangedStrength PrayerID    = "ranged_strength"
	PrayerMagicStrength  PrayerID    = "magic_strength"
	PrayerRegeneration   PrayerID    = "regeneration"
	PrayerSetCommon      PrayerSetID = "common_litany"
	PrayerSetWar         PrayerSetID = "war_covenant"

	evasionDodgeBonus                  = 0.05
	evasionDrainIntervalPulses         = 10 // Ten 600 ms action pulses is six seconds.
	strengthPrayerMeleeDamageBonus     = 0.05
	strengthPrayerDrainIntervalPulses  = 10 // Ten 600 ms action pulses is six seconds.
	accuracyPrayerMeleeScoreBonus      = 5
	accuracyPrayerDrainIntervalPulses  = 10 // Ten 600 ms action pulses is six seconds.
	defencePrayerDamageReduction       = 0.05
	defencePrayerDrainIntervalPulses   = 10 // Ten 600 ms action pulses is six seconds.
	rangedPrayerDamageBonus            = 0.05
	magicPrayerDamageBonus             = 0.05
	offensivePrayerDrainIntervalPulses = 10
	regenerationHealthFraction         = 0.05
	regenerationDrainIntervalPulses    = 10
)

type prayerSetDefinition struct {
	Name      string
	PrayerIDs map[PrayerID]bool
}

var prayerSets = map[PrayerSetID]prayerSetDefinition{
	PrayerSetCommon: {
		Name: "Common Litany",
		PrayerIDs: map[PrayerID]bool{
			PrayerEvasion: true, PrayerStrength: true, PrayerAccuracy: true, PrayerDefence: true,
			PrayerRangedStrength: true, PrayerMagicStrength: true, PrayerRegeneration: true,
		},
	},
	// The first alternate covenant intentionally keeps the starter prayers. It
	// establishes safe switching now; future god-specific prayers can diverge
	// without changing the network or altar interaction model.
	PrayerSetWar: {
		Name: "War Covenant",
		PrayerIDs: map[PrayerID]bool{
			PrayerEvasion: true, PrayerStrength: true, PrayerAccuracy: true, PrayerDefence: true,
			PrayerRangedStrength: true, PrayerMagicStrength: true, PrayerRegeneration: true,
		},
	},
}

func prayerAllowedInSet(setID PrayerSetID, prayerID PrayerID) bool {
	if setID == "" {
		setID = PrayerSetCommon
	}
	set, exists := prayerSets[setID]
	return exists && set.PrayerIDs[prayerID]
}

type PrayerCommand struct {
	PlayerID PlayerID
	Sequence uint64
	PrayerID PrayerID
	Active   bool
}

type prayerDefinition struct {
	ID                  PrayerID
	Name                string
	RequiredLevel       int
	MeleeDamageBonus    float64
	RangedDamageBonus   float64
	MagicDamageBonus    float64
	HealthRegeneration  float64
	MeleeAccuracyBonus  int
	DamageReduction     float64
	DodgeBonus          float64
	DrainAmount         int
	DrainIntervalPulses uint64
	ExclusiveGroup      string
}

type prayerBonuses struct {
	MeleeDamage     float64
	RangedDamage    float64
	MagicDamage     float64
	MeleeAccuracy   int
	DamageReduction float64
	Dodge           float64
}

var prayerDefinitions = map[PrayerID]prayerDefinition{
	PrayerEvasion: {
		ID: PrayerEvasion, Name: "Evasion", RequiredLevel: 1,
		DodgeBonus:          evasionDodgeBonus,
		DrainAmount:         1,
		DrainIntervalPulses: evasionDrainIntervalPulses,
	},
	PrayerStrength: {
		ID: PrayerStrength, Name: "Strength", RequiredLevel: 1,
		MeleeDamageBonus:    strengthPrayerMeleeDamageBonus,
		DrainAmount:         1,
		DrainIntervalPulses: strengthPrayerDrainIntervalPulses,
	},
	PrayerAccuracy: {
		ID: PrayerAccuracy, Name: "Accuracy", RequiredLevel: 1,
		MeleeAccuracyBonus:  accuracyPrayerMeleeScoreBonus,
		DrainAmount:         1,
		DrainIntervalPulses: accuracyPrayerDrainIntervalPulses,
	},
	PrayerDefence: {
		ID: PrayerDefence, Name: "Defence", RequiredLevel: 1,
		DamageReduction:     defencePrayerDamageReduction,
		DrainAmount:         1,
		DrainIntervalPulses: defencePrayerDrainIntervalPulses,
	},
	PrayerRangedStrength: {
		ID: PrayerRangedStrength, Name: "Ranged Strength", RequiredLevel: 1,
		RangedDamageBonus: rangedPrayerDamageBonus, DrainAmount: 1,
		DrainIntervalPulses: offensivePrayerDrainIntervalPulses,
	},
	PrayerMagicStrength: {
		ID: PrayerMagicStrength, Name: "Magic Strength", RequiredLevel: 1,
		MagicDamageBonus: magicPrayerDamageBonus, DrainAmount: 1,
		DrainIntervalPulses: offensivePrayerDrainIntervalPulses,
	},
	PrayerRegeneration: {
		ID: PrayerRegeneration, Name: "Regeneration", RequiredLevel: 1,
		HealthRegeneration: regenerationHealthFraction, DrainAmount: 1,
		DrainIntervalPulses: regenerationDrainIntervalPulses,
	},
}

var prayerCatalogOrder = []PrayerID{
	PrayerAccuracy,
	PrayerStrength,
	PrayerDefence,
	PrayerEvasion,
	PrayerRangedStrength,
	PrayerMagicStrength,
	PrayerRegeneration,
}

var prayerSetCatalogOrder = []PrayerSetID{
	PrayerSetCommon,
	PrayerSetWar,
}

func prayerDefinitionFor(id PrayerID) (prayerDefinition, bool) {
	definition, exists := prayerDefinitions[id]
	return definition, exists
}

// QueuePrayer records an authoritative prayer toggle for the next action pulse.
func (w *World) QueuePrayer(command PrayerCommand) error {
	if command.Sequence == 0 {
		return fmt.Errorf("prayer sequence must be positive")
	}
	definition, exists := prayerDefinitionFor(command.PrayerID)
	if !exists {
		return fmt.Errorf("prayer %q does not exist", command.PrayerID)
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
	if SkillLevelForExperience(player.skills[SkillPrayer]) < definition.RequiredLevel {
		return fmt.Errorf("prayer %q requires Prayer level %d", command.PrayerID, definition.RequiredLevel)
	}
	if !prayerAllowedInSet(player.activePrayerSet, command.PrayerID) {
		return fmt.Errorf("prayer %q is not part of prayer set %q", command.PrayerID, player.activePrayerSet)
	}
	if command.Active && player.prayer <= 0 {
		return fmt.Errorf("player %q has no prayer points", command.PlayerID)
	}
	if command.Sequence <= player.lastPrayerSequence {
		return fmt.Errorf("prayer sequence %d is not newer than %d", command.Sequence, player.lastPrayerSequence)
	}
	if pending, found := w.combat.pendingPrayers[command.PlayerID]; found && command.Sequence <= pending.Sequence {
		return fmt.Errorf("prayer sequence %d is not newer than pending sequence %d", command.Sequence, pending.Sequence)
	}
	w.combat.pendingPrayers[command.PlayerID] = command
	return nil
}

func (w *World) applyPrayerCommands() {
	playerIDs := make([]PlayerID, 0, len(w.combat.pendingPrayers))
	for id := range w.combat.pendingPrayers {
		playerIDs = append(playerIDs, id)
	}
	sort.Slice(playerIDs, func(i, j int) bool { return playerIDs[i] < playerIDs[j] })
	for _, id := range playerIDs {
		command := w.combat.pendingPrayers[id]
		if player, exists := w.players[id]; exists && !player.defeated {
			if command.Active && player.prayer > 0 {
				player.activePrayers[command.PrayerID] = true
			} else {
				delete(player.activePrayers, command.PrayerID)
				delete(player.prayerDrainProgress, command.PrayerID)
			}
			player.lastPrayerSequence = command.Sequence
		}
		delete(w.combat.pendingPrayers, id)
	}
}

func (w *World) drainActivePrayers() {
	for _, player := range w.players {
		if player.defeated || len(player.activePrayers) == 0 {
			continue
		}
		prayerIDs := activePrayerIDs(player)
		for _, id := range prayerIDs {
			definition, exists := prayerDefinitionFor(id)
			if !exists {
				delete(player.activePrayers, id)
				delete(player.prayerDrainProgress, id)
				continue
			}
			player.prayerDrainProgress[id]++
			if player.prayerDrainProgress[id] < definition.DrainIntervalPulses {
				continue
			}
			player.prayerDrainProgress[id] = 0
			if definition.HealthRegeneration > 0 && player.health < player.maximumHealth {
				healing := max(1, int(float64(player.maximumHealth)*definition.HealthRegeneration+0.5))
				player.health = min(player.maximumHealth, player.health+healing)
			}
			player.prayer = max(0, player.prayer-definition.DrainAmount)
			if player.prayer == 0 {
				clearActivePrayers(player)
				break
			}
		}
	}
}

func effectivePlayerDodgeRate(player *playerState) float64 {
	if player == nil {
		return 0
	}
	return min(0.95, player.dodgeRate+activePrayerBonuses(player).Dodge)
}

func activePrayerBonuses(player *playerState) prayerBonuses {
	bonuses := prayerBonuses{}
	if player == nil {
		return bonuses
	}
	for id, active := range player.activePrayers {
		if !active {
			continue
		}
		if definition, exists := prayerDefinitionFor(id); exists {
			bonuses.MeleeDamage += definition.MeleeDamageBonus
			bonuses.RangedDamage += definition.RangedDamageBonus
			bonuses.MagicDamage += definition.MagicDamageBonus
			bonuses.MeleeAccuracy += definition.MeleeAccuracyBonus
			bonuses.DamageReduction += definition.DamageReduction
			bonuses.Dodge += definition.DodgeBonus
		}
	}
	return bonuses
}

func prayerAccuracyBonus(player *playerState, weapon WeaponDefinition) int {
	if weapon.MeleeStrength <= 0 {
		return 0
	}
	return activePrayerBonuses(player).MeleeAccuracy
}

func reduceIncomingDamage(player *playerState, damage int, damageType DamageType) int {
	if damage <= 0 {
		return 0
	}
	prayerReduction := min(0.90, activePrayerBonuses(player).DamageReduction)
	armourReduction := armourDamageReduction(player, damageType)
	return max(0, int(float64(damage)*(1-prayerReduction)*(1-armourReduction)+0.5))
}

func activePrayerIDs(player *playerState) []PrayerID {
	ids := make([]PrayerID, 0, len(player.activePrayers))
	for id, active := range player.activePrayers {
		if active {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func clearActivePrayers(player *playerState) {
	clear(player.activePrayers)
	clear(player.prayerDrainProgress)
}
