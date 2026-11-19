package simulation

import (
	"fmt"
	"math"
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
	PrayerProtectMelee   PrayerID    = "protect_melee"
	PrayerProtectRanged  PrayerID    = "protect_ranged"
	PrayerProtectMagic   PrayerID    = "protect_magic"
	PrayerMeleeMastery   PrayerID    = "melee_mastery"
	PrayerRangedMastery  PrayerID    = "ranged_mastery"
	PrayerMagicMastery   PrayerID    = "magic_mastery"
	PrayerMercy          PrayerID    = "mercy"
	PrayerWrath          PrayerID    = "wrath"
	PrayerSetCommon      PrayerSetID = "common_litany"
	PrayerSetWar         PrayerSetID = "war_covenant"

	evasionDodgeBonus                  = 0.05
	evasionDrainIntervalPulses         = 10 // Ten 600 ms action pulses is six seconds.
	strengthPrayerMeleeDamageBonus     = 0.05
	strengthPrayerDrainIntervalPulses  = 10 // Ten 600 ms action pulses is six seconds.
	accuracyPrayerMeleeHitChanceBonus  = 0.05
	accuracyPrayerDrainIntervalPulses  = 10 // Ten 600 ms action pulses is six seconds.
	defencePrayerDamageReduction       = 0.05
	defencePrayerDrainIntervalPulses   = 10 // Ten 600 ms action pulses is six seconds.
	rangedPrayerDamageBonus            = 0.05
	magicPrayerDamageBonus             = 0.05
	offensivePrayerDrainIntervalPulses = 10
	regenerationHealthFraction         = 0.05
	regenerationDrainIntervalPulses    = 10
	protectionDamageReduction          = 0.30
	protectionDrainIntervalPulses      = 6 // Six 600 ms action pulses is 3.6 seconds.
	offensiveExclusiveGroup            = "offensive_boost"
	protectionExclusiveGroup           = "protection"
	masteryRequiredLevel               = 40
	masteryBonus                       = 0.05
	masteryDrainIntervalPulses         = 6
	mercyRequiredLevel                 = 30
	mercyHealthThreshold               = 0.20
	mercyHealFraction                  = 0.20
	mercyPrayerCost                    = 25
	emergencyExclusiveGroup            = "emergency"
	wrathRequiredLevel                 = 45
	wrathDamageFraction                = 0.20
	wrathRadius                        = 4.0
	wrathDrainIntervalPulses           = 6
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
			PrayerProtectMelee: true, PrayerProtectRanged: true, PrayerProtectMagic: true,
			PrayerMeleeMastery: true, PrayerRangedMastery: true, PrayerMagicMastery: true,
			PrayerMercy: true,
			PrayerWrath: true,
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
			PrayerProtectMelee: true, PrayerProtectRanged: true, PrayerProtectMagic: true,
			PrayerMeleeMastery: true, PrayerRangedMastery: true, PrayerMagicMastery: true,
			PrayerMercy: true,
			PrayerWrath: true,
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
	ID                       PrayerID
	Name                     string
	RequiredLevel            int
	MeleeDamageBonus         float64
	RangedDamageBonus        float64
	MagicDamageBonus         float64
	HealthRegeneration       float64
	MeleeHitChanceBonus      float64
	RangedHitChanceBonus     float64
	MagicHitChanceBonus      float64
	DamageReduction          float64
	MeleeProtection          float64
	RangedProtection         float64
	MagicProtection          float64
	EmergencyHealthThreshold float64
	EmergencyHealFraction    float64
	EmergencyPrayerCost      int
	DeathBurstDamageFraction float64
	DeathBurstRadius         float64
	DodgeBonus               float64
	DrainAmount              int
	DrainIntervalPulses      uint64
	ExclusiveGroup           string
}

type prayerBonuses struct {
	MeleeDamage      float64
	RangedDamage     float64
	MagicDamage      float64
	MeleeHitChance   float64
	RangedHitChance  float64
	MagicHitChance   float64
	DamageReduction  float64
	MeleeProtection  float64
	RangedProtection float64
	MagicProtection  float64
	Dodge            float64
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
		ExclusiveGroup:      offensiveExclusiveGroup,
	},
	PrayerAccuracy: {
		ID: PrayerAccuracy, Name: "Accuracy", RequiredLevel: 1,
		MeleeHitChanceBonus: accuracyPrayerMeleeHitChanceBonus,
		DrainAmount:         1,
		DrainIntervalPulses: accuracyPrayerDrainIntervalPulses,
		ExclusiveGroup:      offensiveExclusiveGroup,
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
		DrainIntervalPulses: offensivePrayerDrainIntervalPulses, ExclusiveGroup: offensiveExclusiveGroup,
	},
	PrayerMagicStrength: {
		ID: PrayerMagicStrength, Name: "Magic Strength", RequiredLevel: 1,
		MagicDamageBonus: magicPrayerDamageBonus, DrainAmount: 1,
		DrainIntervalPulses: offensivePrayerDrainIntervalPulses, ExclusiveGroup: offensiveExclusiveGroup,
	},
	PrayerRegeneration: {
		ID: PrayerRegeneration, Name: "Regeneration", RequiredLevel: 1,
		HealthRegeneration: regenerationHealthFraction, DrainAmount: 1,
		DrainIntervalPulses: regenerationDrainIntervalPulses,
	},
	PrayerProtectMelee: {
		ID: PrayerProtectMelee, Name: "Protect from Melee", RequiredLevel: 1,
		MeleeProtection: protectionDamageReduction, DrainAmount: 1,
		DrainIntervalPulses: protectionDrainIntervalPulses, ExclusiveGroup: protectionExclusiveGroup,
	},
	PrayerProtectRanged: {
		ID: PrayerProtectRanged, Name: "Protect from Ranged", RequiredLevel: 1,
		RangedProtection: protectionDamageReduction, DrainAmount: 1,
		DrainIntervalPulses: protectionDrainIntervalPulses, ExclusiveGroup: protectionExclusiveGroup,
	},
	PrayerProtectMagic: {
		ID: PrayerProtectMagic, Name: "Protect from Magic", RequiredLevel: 1,
		MagicProtection: protectionDamageReduction, DrainAmount: 1,
		DrainIntervalPulses: protectionDrainIntervalPulses, ExclusiveGroup: protectionExclusiveGroup,
	},
	PrayerMeleeMastery: {
		ID: PrayerMeleeMastery, Name: "Melee Mastery", RequiredLevel: masteryRequiredLevel,
		MeleeHitChanceBonus: masteryBonus, MeleeDamageBonus: masteryBonus, DamageReduction: masteryBonus,
		DrainAmount: 1, DrainIntervalPulses: masteryDrainIntervalPulses, ExclusiveGroup: offensiveExclusiveGroup,
	},
	PrayerRangedMastery: {
		ID: PrayerRangedMastery, Name: "Ranged Mastery", RequiredLevel: masteryRequiredLevel,
		RangedHitChanceBonus: masteryBonus, RangedDamageBonus: masteryBonus, DamageReduction: masteryBonus,
		DrainAmount: 1, DrainIntervalPulses: masteryDrainIntervalPulses, ExclusiveGroup: offensiveExclusiveGroup,
	},
	PrayerMagicMastery: {
		ID: PrayerMagicMastery, Name: "Magic Mastery", RequiredLevel: masteryRequiredLevel,
		MagicHitChanceBonus: masteryBonus, MagicDamageBonus: masteryBonus, DamageReduction: masteryBonus,
		DrainAmount: 1, DrainIntervalPulses: masteryDrainIntervalPulses, ExclusiveGroup: offensiveExclusiveGroup,
	},
	PrayerMercy: {
		ID: PrayerMercy, Name: "Mercy", RequiredLevel: mercyRequiredLevel,
		EmergencyHealthThreshold: mercyHealthThreshold,
		EmergencyHealFraction:    mercyHealFraction, EmergencyPrayerCost: mercyPrayerCost,
		ExclusiveGroup: emergencyExclusiveGroup,
	},
	PrayerWrath: {
		ID: PrayerWrath, Name: "Wrath", RequiredLevel: wrathRequiredLevel,
		DeathBurstDamageFraction: wrathDamageFraction, DeathBurstRadius: wrathRadius,
		DrainAmount: 1, DrainIntervalPulses: wrathDrainIntervalPulses,
		ExclusiveGroup: emergencyExclusiveGroup,
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
	PrayerProtectMelee,
	PrayerProtectRanged,
	PrayerProtectMagic,
	PrayerMeleeMastery,
	PrayerRangedMastery,
	PrayerMagicMastery,
	PrayerMercy,
	PrayerWrath,
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
	if command.Active && definition.EmergencyPrayerCost > 0 && player.prayer < definition.EmergencyPrayerCost {
		return fmt.Errorf("prayer %q requires %d prayer points", command.PrayerID, definition.EmergencyPrayerCost)
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
				definition, _ := prayerDefinitionFor(command.PrayerID)
				deactivateExclusivePrayers(player, command.PrayerID, definition.ExclusiveGroup)
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

func deactivateExclusivePrayers(player *playerState, activated PrayerID, exclusiveGroup string) {
	if player == nil || exclusiveGroup == "" {
		return
	}
	for prayerID := range player.activePrayers {
		if prayerID == activated {
			continue
		}
		definition, exists := prayerDefinitionFor(prayerID)
		if !exists || definition.ExclusiveGroup != exclusiveGroup {
			continue
		}
		delete(player.activePrayers, prayerID)
		delete(player.prayerDrainProgress, prayerID)
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
			if definition.EmergencyPrayerCost > 0 {
				if player.prayer < definition.EmergencyPrayerCost {
					delete(player.activePrayers, id)
					delete(player.prayerDrainProgress, id)
				}
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

func triggerEmergencyPrayer(player *playerState) int {
	if player == nil || player.health <= 0 || player.maximumHealth <= 0 ||
		!player.activePrayers[PrayerMercy] {
		return 0
	}
	definition, exists := prayerDefinitionFor(PrayerMercy)
	if !exists || definition.EmergencyPrayerCost <= 0 || player.prayer < definition.EmergencyPrayerCost ||
		float64(player.health)/float64(player.maximumHealth) >= definition.EmergencyHealthThreshold {
		return 0
	}
	healing := max(1, int(math.Round(float64(player.maximumHealth)*definition.EmergencyHealFraction)))
	previousHealth := player.health
	player.health = min(player.maximumHealth, player.health+healing)
	player.prayer = max(0, player.prayer-definition.EmergencyPrayerCost)
	delete(player.activePrayers, PrayerMercy)
	delete(player.prayerDrainProgress, PrayerMercy)
	if player.prayer == 0 {
		clearActivePrayers(player)
	}
	return player.health - previousHealth
}

func (w *World) triggerWrath(source *playerState) {
	if w == nil || source == nil || !source.activePrayers[PrayerWrath] {
		return
	}
	definition, exists := prayerDefinitionFor(PrayerWrath)
	if !exists || definition.DeathBurstDamageFraction <= 0 || definition.DeathBurstRadius <= 0 {
		return
	}
	// Remove Wrath before applying damage so a chained death cannot trigger the
	// same player's burst recursively.
	delete(source.activePrayers, PrayerWrath)
	delete(source.prayerDrainProgress, PrayerWrath)
	damage := max(1, int(math.Round(float64(source.maximumHealth)*definition.DeathBurstDamageFraction)))
	w.combat.recordEvent(CombatEvent{
		Pulse: w.combat.pulse, Type: CombatEventWrath,
		SourceID: string(source.id), TargetID: string(source.id), Amount: damage,
	})

	playerIDs := make([]PlayerID, 0, len(w.players))
	for playerID := range w.players {
		playerIDs = append(playerIDs, playerID)
	}
	sort.Slice(playerIDs, func(first, second int) bool { return playerIDs[first] < playerIDs[second] })
	for _, playerID := range playerIDs {
		target := w.players[playerID]
		if target == nil || target == source || target.defeated ||
			!playerInPvPField(source) || !playerInPvPField(target) ||
			distance(source.position, target.position) > definition.DeathBurstRadius {
			continue
		}
		damageDealt := min(damage, target.health)
		target.health = max(0, target.health-damageDealt)
		triggerEmergencyPrayer(target)
		w.combat.recordEvent(CombatEvent{
			Pulse: w.combat.pulse, Type: CombatEventDamage,
			SourceID: string(source.id), TargetID: string(target.id), Amount: damageDealt,
		})
		if target.health == 0 {
			w.defeatPlayer(target)
		}
	}

	for _, entry := range w.combat.dummyTargets() {
		target := entry.dummy
		if target == nil || target.Defeated ||
			distance(source.position, target.Position) > definition.DeathBurstRadius {
			continue
		}
		damageDealt := damage
		if !target.InfiniteHealth {
			damageDealt = min(damage, target.Health)
			target.Health = max(0, target.Health-damageDealt)
		}
		w.combat.recordEvent(CombatEvent{
			Pulse: w.combat.pulse, Type: CombatEventDamage,
			SourceID: string(source.id), TargetID: entry.id, Amount: damageDealt,
		})
		if target.Health == 0 {
			w.defeatDummy(entry.id, target)
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
			bonuses.MeleeHitChance += definition.MeleeHitChanceBonus
			bonuses.RangedHitChance += definition.RangedHitChanceBonus
			bonuses.MagicHitChance += definition.MagicHitChanceBonus
			bonuses.DamageReduction += definition.DamageReduction
			bonuses.MeleeProtection += definition.MeleeProtection
			bonuses.RangedProtection += definition.RangedProtection
			bonuses.MagicProtection += definition.MagicProtection
			bonuses.Dodge += definition.DodgeBonus
		}
	}
	return bonuses
}

func prayerHitChanceBonus(player *playerState, weapon WeaponDefinition) float64 {
	bonuses := activePrayerBonuses(player)
	switch {
	case weapon.RangedStrength > 0:
		return bonuses.RangedHitChance
	case weapon.MagicStrength > 0:
		return bonuses.MagicHitChance
	case weapon.MeleeStrength > 0:
		return bonuses.MeleeHitChance
	default:
		return 0
	}
}

func prayerSpellHitChanceBonus(player *playerState) float64 {
	return activePrayerBonuses(player).MagicHitChance
}

func reduceIncomingDamage(player *playerState, damage int, damageType DamageType, element MagicElement) int {
	if damage <= 0 {
		return 0
	}
	bonuses := activePrayerBonuses(player)
	generalReduction := min(0.90, bonuses.DamageReduction)
	styleReduction := protectionReductionForDamageType(bonuses, damageType)
	prayerReduction := 1 - (1-generalReduction)*(1-styleReduction)
	armourReduction := armourDamageReduction(player, damageType)
	mitigated := float64(damage) * (1 - prayerReduction) * (1 - armourReduction)
	if damageType == DamageTypeMagic && element != "" {
		mitigated *= resistanceMultiplier(playerElementalResistances(player), element)
	}
	return max(0, int(math.Round(mitigated+1e-9)))
}

func protectionReductionForDamageType(bonuses prayerBonuses, damageType DamageType) float64 {
	switch damageType {
	case DamageTypeStab, DamageTypeSlash, DamageTypeCrush:
		return min(0.90, bonuses.MeleeProtection)
	case DamageTypeRanged:
		return min(0.90, bonuses.RangedProtection)
	case DamageTypeMagic:
		return min(0.90, bonuses.MagicProtection)
	default:
		return 0
	}
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
