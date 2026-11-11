package simulation

import (
	"fmt"
	"sort"
)

const (
	AmmunitionPouchCapacity     = 10
	BaseAmmunitionPouchUnlocked = 2
)

type AmmunitionPouchSnapshot struct {
	Capacity      int
	UnlockedSlots int
	Items         []InventoryItem
}

type AmmunitionTransferCommand struct {
	PlayerID      PlayerID
	Sequence      uint64
	Deposit       bool
	InventorySlot int
	PouchSlot     int
}

func newAmmunitionPouch() inventoryState {
	return inventoryState{capacity: BaseAmmunitionPouchUnlocked}
}

func ammunitionForWeapon(player *playerState, weapon WeaponDefinition) (ItemCost, int, bool) {
	if weapon.Ammunition.Quantity <= 0 {
		return ItemCost{}, 0, true
	}
	if player == nil {
		return ItemCost{}, 0, false
	}
	pouch := player.ammunitionPouch
	for slot := 0; slot < pouch.capacity; slot++ {
		item, exists := inventoryItemAt(pouch, slot)
		if !exists {
			continue
		}
		definition, defined := itemDefinition(item.ItemID)
		if !defined || definition.AmmunitionFamily != weapon.AmmunitionFamily || item.Quantity < weapon.Ammunition.Quantity {
			continue
		}
		if _, unmet := unmetEquipRequirement(player, definition); unmet {
			continue
		}
		return ItemCost{ItemID: item.ItemID, Quantity: weapon.Ammunition.Quantity}, definition.AmmunitionStrength, true
	}
	return ItemCost{}, 0, false
}

func ammunitionStrengthForPlayer(player *playerState, weapon WeaponDefinition) int {
	if player == nil {
		return 0
	}
	_, strength, available := ammunitionForWeapon(player, weapon)
	if !available {
		return 0
	}
	return strength
}

func snapshotAmmunitionPouch(pouch inventoryState) AmmunitionPouchSnapshot {
	return AmmunitionPouchSnapshot{
		Capacity: AmmunitionPouchCapacity, UnlockedSlots: pouch.capacity,
		Items: append([]InventoryItem(nil), pouch.items...),
	}
}

func ammunitionPouchFromSnapshot(snapshot AmmunitionPouchSnapshot) (inventoryState, error) {
	if snapshot.Capacity == 0 && snapshot.UnlockedSlots == 0 && len(snapshot.Items) == 0 {
		return newAmmunitionPouch(), nil
	}
	if snapshot.Capacity != AmmunitionPouchCapacity || snapshot.UnlockedSlots < 1 ||
		snapshot.UnlockedSlots > snapshot.Capacity {
		return inventoryState{}, fmt.Errorf("ammunition pouch capacity is invalid")
	}
	pouch := inventoryState{capacity: snapshot.UnlockedSlots, items: append([]InventoryItem(nil), snapshot.Items...)}
	seenItems := make(map[string]struct{}, len(pouch.items))
	seenSlots := make(map[int]struct{}, len(pouch.items))
	for _, item := range pouch.items {
		definition, exists := itemDefinition(item.ItemID)
		if !exists || !definition.Ammunition || item.Slot < 0 || item.Slot >= pouch.capacity ||
			item.Quantity <= 0 || item.InstanceID != "" || item.Reinforcement != 0 {
			return inventoryState{}, fmt.Errorf("ammunition pouch slot %d is invalid", item.Slot)
		}
		if _, duplicate := seenItems[item.ItemID]; duplicate {
			return inventoryState{}, fmt.Errorf("ammunition %q is split across pouch slots", item.ItemID)
		}
		if _, duplicate := seenSlots[item.Slot]; duplicate {
			return inventoryState{}, fmt.Errorf("ammunition pouch slot %d is duplicated", item.Slot)
		}
		seenItems[item.ItemID] = struct{}{}
		seenSlots[item.Slot] = struct{}{}
	}
	sort.Slice(pouch.items, func(i, j int) bool { return pouch.items[i].Slot < pouch.items[j].Slot })
	return pouch, nil
}

func (w *World) QueueAmmunitionTransfer(command AmmunitionTransferCommand) error {
	if command.Sequence == 0 {
		return fmt.Errorf("ammunition transfer sequence must be positive")
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	player, exists := w.players[command.PlayerID]
	if !exists || player.defeated {
		return fmt.Errorf("player %q cannot use the ammunition pouch", command.PlayerID)
	}
	if command.Sequence <= player.lastAmmunitionSequence {
		return fmt.Errorf("ammunition sequence %d is not newer than %d", command.Sequence, player.lastAmmunitionSequence)
	}
	if pending, exists := w.pendingAmmunitionTransfers[command.PlayerID]; exists && command.Sequence <= pending.Sequence {
		return fmt.Errorf("ammunition sequence %d is not newer than pending sequence %d", command.Sequence, pending.Sequence)
	}
	if _, _, err := ammunitionTransferAfter(player, command); err != nil {
		return err
	}
	w.pendingAmmunitionTransfers[command.PlayerID] = command
	return nil
}

func ammunitionTransferAfter(player *playerState, command AmmunitionTransferCommand) (inventoryState, inventoryState, error) {
	inventory := cloneInventory(player.inventory)
	pouch := cloneInventory(player.ammunitionPouch)
	if command.Deposit {
		item, exists := inventoryItemAt(inventory, command.InventorySlot)
		definition, defined := itemDefinition(item.ItemID)
		if !exists || !defined || !definition.Ammunition {
			return inventoryState{}, inventoryState{}, fmt.Errorf("inventory slot %d does not contain ammunition", command.InventorySlot)
		}
		if requirement, unmet := unmetEquipRequirement(player, definition); unmet {
			return inventoryState{}, inventoryState{}, fmt.Errorf("item %q requires %s level %d", item.ItemID, requirement.Skill, requirement.Level)
		}
		if _, removed := removeInventorySlot(&inventory, command.InventorySlot); !removed || !pouch.add(item.ItemID, item.Quantity) {
			return inventoryState{}, inventoryState{}, fmt.Errorf("ammunition pouch has no unlocked slot for %q", item.ItemID)
		}
		return inventory, pouch, nil
	}
	item, exists := inventoryItemAt(pouch, command.PouchSlot)
	if !exists {
		return inventoryState{}, inventoryState{}, fmt.Errorf("ammunition pouch slot %d is empty", command.PouchSlot)
	}
	if !inventory.add(item.ItemID, item.Quantity) {
		return inventoryState{}, inventoryState{}, fmt.Errorf("inventory is full")
	}
	if _, removed := removeInventorySlot(&pouch, command.PouchSlot); !removed {
		return inventoryState{}, inventoryState{}, fmt.Errorf("ammunition pouch slot %d is empty", command.PouchSlot)
	}
	return inventory, pouch, nil
}

func (w *World) applyPendingAmmunitionTransfers() {
	for id, command := range w.pendingAmmunitionTransfers {
		delete(w.pendingAmmunitionTransfers, id)
		player, exists := w.players[id]
		if !exists || player.defeated {
			continue
		}
		inventory, pouch, err := ammunitionTransferAfter(player, command)
		if err != nil {
			continue
		}
		player.inventory = inventory
		player.ammunitionPouch = pouch
		player.lastAmmunitionSequence = command.Sequence
	}
}
