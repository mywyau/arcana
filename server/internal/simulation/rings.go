package simulation

import (
	"fmt"
	"sort"
)

const RingRackCapacity = 5

type RingRackSnapshot struct {
	Capacity int
	Items    []InventoryItem
}

type RingTransferCommand struct {
	PlayerID   PlayerID
	Sequence   uint64
	Deposit    bool
	ItemID     string
	InstanceID string
	RackSlot   int
}

func newRingRack() inventoryState { return inventoryState{capacity: RingRackCapacity} }

func snapshotRingRack(rack inventoryState) RingRackSnapshot {
	return RingRackSnapshot{Capacity: RingRackCapacity, Items: append([]InventoryItem(nil), rack.items...)}
}

func ringRackFromSnapshot(snapshot RingRackSnapshot) (inventoryState, error) {
	if snapshot.Capacity == 0 && len(snapshot.Items) == 0 {
		return newRingRack(), nil
	}
	if snapshot.Capacity != RingRackCapacity {
		return inventoryState{}, fmt.Errorf("ring rack capacity must be %d", RingRackCapacity)
	}
	rack := inventoryState{capacity: RingRackCapacity, items: append([]InventoryItem(nil), snapshot.Items...)}
	seenSlots := make(map[int]bool, len(rack.items))
	seenInstances := make(map[string]bool, len(rack.items))
	for _, item := range rack.items {
		definition, exists := itemDefinition(item.ItemID)
		if !exists || definition.Kind != ItemKindArmour || definition.Armour.Slot != ArmourSlotRing ||
			item.Slot < 0 || item.Slot >= RingRackCapacity || item.Quantity != 1 || item.InstanceID == "" ||
			seenSlots[item.Slot] || seenInstances[item.InstanceID] {
			return inventoryState{}, fmt.Errorf("ring rack slot %d is invalid", item.Slot)
		}
		seenSlots[item.Slot] = true
		seenInstances[item.InstanceID] = true
	}
	sort.Slice(rack.items, func(i, j int) bool { return rack.items[i].Slot < rack.items[j].Slot })
	return rack, nil
}

func (w *World) QueueRingTransfer(command RingTransferCommand) error {
	if command.Sequence == 0 {
		return fmt.Errorf("ring sequence must be positive")
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	player, exists := w.players[command.PlayerID]
	if !exists || player.defeated {
		return fmt.Errorf("player %q cannot change rings", command.PlayerID)
	}
	if command.Sequence <= player.lastRingSequence {
		return fmt.Errorf("ring sequence %d is not newer than %d", command.Sequence, player.lastRingSequence)
	}
	if pending, exists := w.pendingRingTransfers[command.PlayerID]; exists && command.Sequence <= pending.Sequence {
		return fmt.Errorf("ring sequence %d is not newer than pending sequence %d", command.Sequence, pending.Sequence)
	}
	if _, _, err := ringTransferAfter(player, command); err != nil {
		return err
	}
	w.pendingRingTransfers[command.PlayerID] = command
	return nil
}

func ringTransferAfter(player *playerState, command RingTransferCommand) (inventoryState, inventoryState, error) {
	inventory := cloneInventory(player.inventory)
	rack := cloneInventory(player.ringRack)
	if command.Deposit {
		item, exists := inventory.itemByInstance(command.InstanceID)
		definition, defined := itemDefinition(item.ItemID)
		if !exists || !defined || item.ItemID != command.ItemID || definition.Kind != ItemKindArmour || definition.Armour.Slot != ArmourSlotRing {
			return inventoryState{}, inventoryState{}, fmt.Errorf("player does not own ring %q", command.ItemID)
		}
		if requirement, unmet := unmetEquipRequirement(player, definition); unmet {
			return inventoryState{}, inventoryState{}, fmt.Errorf("ring %q requires %s level %d", command.ItemID, requirement.Skill, requirement.Level)
		}
		if _, removed := inventory.removeInstance(command.InstanceID); !removed || !rack.addInventoryItem(item) {
			return inventoryState{}, inventoryState{}, fmt.Errorf("all five ring slots are occupied")
		}
		return inventory, rack, nil
	}
	item, exists := inventoryItemAt(rack, command.RackSlot)
	if !exists {
		return inventoryState{}, inventoryState{}, fmt.Errorf("ring slot %d is empty", command.RackSlot)
	}
	if !inventory.addInventoryItem(item) {
		return inventoryState{}, inventoryState{}, fmt.Errorf("inventory is full")
	}
	if _, removed := removeInventorySlot(&rack, command.RackSlot); !removed {
		return inventoryState{}, inventoryState{}, fmt.Errorf("ring slot %d is empty", command.RackSlot)
	}
	return inventory, rack, nil
}

func (w *World) applyPendingRingTransfers() {
	for id, command := range w.pendingRingTransfers {
		delete(w.pendingRingTransfers, id)
		player, exists := w.players[id]
		if !exists || player.defeated {
			continue
		}
		inventory, rack, err := ringTransferAfter(player, command)
		if err != nil {
			continue
		}
		player.inventory = inventory
		player.ringRack = rack
		player.lastRingSequence = command.Sequence
		refreshPlayerMaximumHealth(player)
	}
}
