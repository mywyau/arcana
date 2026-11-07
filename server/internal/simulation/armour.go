package simulation

import "fmt"

type ArmourSlot string
type ArmourSetID string

const (
	ArmourSlotHead ArmourSlot = "head"
	ArmourSlotBody ArmourSlot = "body"
	ArmourSlotFeet ArmourSlot = "feet"
	ArmourSlotNeck ArmourSlot = "neck"

	ItemTrainingHelm   = "training_helm"
	ItemBronzeHelm     = "bronze_helm"
	ItemBronzePlate    = "bronze_platebody"
	ItemIronHelm       = "iron_helm"
	ItemIronPlate      = "iron_platebody"
	ItemMithrilHelm    = "mithril_helm"
	ItemMithrilPlate   = "mithril_platebody"
	ItemTrainingPlate  = "training_plate"
	ItemSwiftHelm      = "swift_helm"
	ItemSwiftBody      = "swift_body"
	ItemSwiftBoots     = "swift_boots"
	ItemAmuletFerocity = "amulet_ferocity"
	ItemAmuletHawk     = "amulet_hawk"
	ItemAmuletEmbers   = "amulet_embers"
	ItemAmuletFury     = "amulet_fury"

	armourSetSwift ArmourSetID = "swift"

	minimumAttackIntervalPulses = 2
)

type DefenceBonuses struct {
	Stab   int
	Slash  int
	Crush  int
	Ranged int
	Magic  int
}

type OffenceBonuses struct {
	MeleeAccuracy  int
	RangedAccuracy int
	MagicAccuracy  int
	MeleeStrength  int
	RangedStrength int
	MagicStrength  int
}

type ArmourDefinition struct {
	Slot        ArmourSlot
	Set         ArmourSetID
	Defence     DefenceBonuses
	Offence     OffenceBonuses
	HealthBonus int
}

type ArmourSetDefinition struct {
	RequiredPieces         int
	AttackIntervalModifier int
}

var armourSetDefinitions = map[ArmourSetID]ArmourSetDefinition{
	armourSetSwift: {RequiredPieces: 3, AttackIntervalModifier: -1},
}

func armourSlot(itemID string) (ArmourSlot, bool) {
	definition, exists := armourDefinition(itemID)
	return definition.Slot, exists
}

func armourDefinition(itemID string) (ArmourDefinition, bool) {
	item, exists := itemDefinition(itemID)
	if !exists || item.Kind != ItemKindArmour {
		return ArmourDefinition{}, false
	}
	return item.Armour, true
}

func equippedDefenceBonuses(equipped map[ArmourSlot]string) DefenceBonuses {
	var total DefenceBonuses
	for _, itemID := range equipped {
		definition, exists := armourDefinition(itemID)
		if !exists {
			continue
		}
		total.Stab += definition.Defence.Stab
		total.Slash += definition.Defence.Slash
		total.Crush += definition.Defence.Crush
		total.Ranged += definition.Defence.Ranged
		total.Magic += definition.Defence.Magic
	}
	return total
}

func equippedDefenceBonusesForPlayer(player *playerState) DefenceBonuses {
	if player == nil {
		return DefenceBonuses{}
	}
	total := equippedDefenceBonuses(player.equippedArmour)
	for slot, instanceID := range player.equippedArmourInstances {
		if equippedReinforcement(player, instanceID) <= 0 || player.equippedArmour[slot] != ItemBronzeHelm {
			continue
		}
		total.Stab++
		total.Slash++
		total.Crush++
	}
	return total
}

func equippedOffenceBonuses(equipped map[ArmourSlot]string) OffenceBonuses {
	var total OffenceBonuses
	for _, itemID := range equipped {
		definition, exists := armourDefinition(itemID)
		if !exists {
			continue
		}
		total.MeleeAccuracy += definition.Offence.MeleeAccuracy
		total.RangedAccuracy += definition.Offence.RangedAccuracy
		total.MagicAccuracy += definition.Offence.MagicAccuracy
		total.MeleeStrength += definition.Offence.MeleeStrength
		total.RangedStrength += definition.Offence.RangedStrength
		total.MagicStrength += definition.Offence.MagicStrength
	}
	return total
}

func equipmentOffenceForWeapon(player *playerState, weapon WeaponDefinition) (accuracy, strength int) {
	if player == nil {
		return 0, 0
	}
	bonuses := equippedOffenceBonuses(player.equippedArmour)
	switch {
	case weapon.RangedStrength > 0:
		return bonuses.RangedAccuracy, bonuses.RangedStrength
	case weapon.MagicStrength > 0:
		return bonuses.MagicAccuracy, bonuses.MagicStrength
	default:
		return bonuses.MeleeAccuracy, bonuses.MeleeStrength
	}
}

type DamageType string

const (
	DamageTypeStab   DamageType = "stab"
	DamageTypeSlash  DamageType = "slash"
	DamageTypeCrush  DamageType = "crush"
	DamageTypeRanged DamageType = "ranged"
	DamageTypeMagic  DamageType = "magic"
)

func defenceBonusForDamageType(bonuses DefenceBonuses, damageType DamageType) int {
	switch damageType {
	case DamageTypeStab:
		return bonuses.Stab
	case DamageTypeSlash:
		return bonuses.Slash
	case DamageTypeRanged:
		return bonuses.Ranged
	case DamageTypeMagic:
		return bonuses.Magic
	default:
		return bonuses.Crush
	}
}

func armourDamageReduction(player *playerState, damageType DamageType) float64 {
	bonus := max(0, defenceBonusForDamageType(equippedDefenceBonusesForPlayer(player), damageType))
	return float64(bonus) / float64(bonus+100)
}

func equippedHealthBonus(equipped map[ArmourSlot]string) int {
	bonus := 0
	for _, itemID := range equipped {
		if definition, exists := armourDefinition(itemID); exists {
			bonus += definition.HealthBonus
		}
	}
	return bonus
}

func refreshPlayerMaximumHealth(player *playerState) {
	if player == nil {
		return
	}
	player.maximumHealth = playerMaximumHealth + equippedHealthBonus(player.equippedArmour)
	player.health = min(player.health, player.maximumHealth)
}

func equippedAttackIntervalModifier(equipped map[ArmourSlot]string) int {
	setPieces := make(map[ArmourSetID]int)
	for _, itemID := range equipped {
		definition, exists := armourDefinition(itemID)
		if exists && definition.Set != "" {
			setPieces[definition.Set]++
		}
	}
	modifier := 0
	for setID, pieces := range setPieces {
		definition, exists := armourSetDefinitions[setID]
		if exists && pieces >= definition.RequiredPieces {
			modifier += definition.AttackIntervalModifier
		}
	}
	return modifier
}

func effectiveWeaponAttackInterval(player *playerState, weapon WeaponDefinition) uint64 {
	modifier := 0
	if player != nil {
		modifier = equippedAttackIntervalModifier(player.equippedArmour)
	}
	return uint64(max(minimumAttackIntervalPulses, int(weapon.AttackInterval)+modifier))
}

func cloneEquippedArmour(equipped map[ArmourSlot]string) map[ArmourSlot]string {
	clone := make(map[ArmourSlot]string, len(equipped))
	for slot, itemID := range equipped {
		clone[slot] = itemID
	}
	return clone
}

func cloneEquippedArmourInstances(equipped map[ArmourSlot]string) map[ArmourSlot]string {
	clone := make(map[ArmourSlot]string, len(equipped))
	for slot, instanceID := range equipped {
		clone[slot] = instanceID
	}
	return clone
}

func cloneEquippedArmourReinforcements(equipped map[ArmourSlot]int) map[ArmourSlot]int {
	clone := make(map[ArmourSlot]int, len(equipped))
	for slot, reinforcement := range equipped {
		clone[slot] = reinforcement
	}
	return clone
}

func validateEquippedArmour(equipped map[ArmourSlot]string, inventory inventoryState) error {
	for slot, itemID := range equipped {
		definedSlot, exists := armourSlot(itemID)
		if !exists || definedSlot != slot {
			return fmt.Errorf("armour item %q is invalid for slot %q", itemID, slot)
		}
		if !inventory.contains(itemID) {
			return fmt.Errorf("equipped armour %q is not in inventory", itemID)
		}
	}
	return nil
}
