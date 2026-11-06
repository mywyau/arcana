package simulation

import "fmt"

type ArmourSlot string

const (
	ArmourSlotHead ArmourSlot = "head"
	ArmourSlotBody ArmourSlot = "body"

	ItemTrainingHelm  = "training_helm"
	ItemTrainingPlate = "training_plate"
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
	Defence DefenceBonuses
}

var armourDefinitions = map[string]ArmourDefinition{
	ItemTrainingHelm: {
		Slot:    ArmourSlotHead,
		Defence: DefenceBonuses{Stab: 2, Slash: 3, Crush: 2, Ranged: 1},
	},
	ItemTrainingPlate: {
		Slot:    ArmourSlotBody,
		Defence: DefenceBonuses{Stab: 8, Slash: 10, Crush: 7, Ranged: 6, Magic: -2},
	},
}

func armourSlot(itemID string) (ArmourSlot, bool) {
	definition, exists := armourDefinitions[itemID]
	return definition.Slot, exists
}

func equippedDefenceBonuses(equipped map[ArmourSlot]string) DefenceBonuses {
	var total DefenceBonuses
	for _, itemID := range equipped {
		definition, exists := armourDefinitions[itemID]
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

func cloneEquippedArmour(equipped map[ArmourSlot]string) map[ArmourSlot]string {
	clone := make(map[ArmourSlot]string, len(equipped))
	for slot, itemID := range equipped {
		clone[slot] = itemID
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
