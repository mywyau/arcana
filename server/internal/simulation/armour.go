package simulation

import "fmt"

type ArmourSlot string

const (
	ArmourSlotHead ArmourSlot = "head"
	ArmourSlotBody ArmourSlot = "body"

	ItemTrainingHelm  = "training_helm"
	ItemTrainingPlate = "training_plate"
)

var armourSlots = map[string]ArmourSlot{
	ItemTrainingHelm:  ArmourSlotHead,
	ItemTrainingPlate: ArmourSlotBody,
}

func armourSlot(itemID string) (ArmourSlot, bool) {
	slot, exists := armourSlots[itemID]
	return slot, exists
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
