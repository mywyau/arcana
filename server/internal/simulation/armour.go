package simulation

import "fmt"

type ArmourSlot string
type ArmourSetID string

const (
	ArmourSlotHead ArmourSlot = "head"
	ArmourSlotBody ArmourSlot = "body"
	ArmourSlotFeet ArmourSlot = "feet"

	ItemTrainingHelm  = "training_helm"
	ItemBronzeHelm    = "bronze_helm"
	ItemBronzePlate   = "bronze_platebody"
	ItemIronHelm      = "iron_helm"
	ItemIronPlate     = "iron_platebody"
	ItemMithrilHelm   = "mithril_helm"
	ItemMithrilPlate  = "mithril_platebody"
	ItemTrainingPlate = "training_plate"
	ItemSwiftHelm     = "swift_helm"
	ItemSwiftBody     = "swift_body"
	ItemSwiftBoots    = "swift_boots"

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

type ArmourDefinition struct {
	Slot    ArmourSlot
	Set     ArmourSetID
	Defence DefenceBonuses
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
