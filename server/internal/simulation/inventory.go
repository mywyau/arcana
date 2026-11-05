package simulation

import (
	"fmt"
	"sort"
)

const BaseInventoryCapacity = 8

// InventoryItem is one occupied slot in a player's authoritative inventory.
type InventoryItem struct {
	Slot     int
	ItemID   string
	Quantity int
}

// InventorySnapshot is a detached view of a player's inventory.
type InventorySnapshot struct {
	Capacity int
	Items    []InventoryItem
}

type inventoryState struct {
	capacity int
	items    []InventoryItem
}

func newStarterInventory() inventoryState {
	return inventoryState{
		capacity: BaseInventoryCapacity,
		items: []InventoryItem{
			{Slot: 0, ItemID: string(WeaponSword), Quantity: 1},
			{Slot: 1, ItemID: string(WeaponDagger), Quantity: 1},
			{Slot: 2, ItemID: string(WeaponBow), Quantity: 1},
			{Slot: 3, ItemID: string(WeaponMace), Quantity: 1},
			{Slot: 4, ItemID: string(WeaponBattleAxe), Quantity: 1},
			{Slot: 5, ItemID: string(WeaponStaff), Quantity: 1},
		},
	}
}

func (inventory inventoryState) contains(itemID string) bool {
	for _, item := range inventory.items {
		if item.ItemID == itemID && item.Quantity > 0 {
			return true
		}
	}
	return false
}

func (inventory inventoryState) canAdd(itemID string) bool {
	return inventory.contains(itemID) || len(inventory.items) < inventory.capacity
}

func (inventory *inventoryState) add(itemID string, quantity int) bool {
	if itemID == "" || quantity <= 0 || !inventory.canAdd(itemID) {
		return false
	}
	for index := range inventory.items {
		if inventory.items[index].ItemID == itemID {
			inventory.items[index].Quantity += quantity
			return true
		}
	}
	occupied := make(map[int]struct{}, len(inventory.items))
	for _, item := range inventory.items {
		occupied[item.Slot] = struct{}{}
	}
	for slot := 0; slot < inventory.capacity; slot++ {
		if _, exists := occupied[slot]; exists {
			continue
		}
		inventory.items = append(inventory.items, InventoryItem{Slot: slot, ItemID: itemID, Quantity: quantity})
		sort.Slice(inventory.items, func(i, j int) bool { return inventory.items[i].Slot < inventory.items[j].Slot })
		return true
	}
	return false
}

func (inventory *inventoryState) remove(itemID string, quantity int) bool {
	if itemID == "" || quantity <= 0 {
		return false
	}
	for index := range inventory.items {
		if inventory.items[index].ItemID != itemID || inventory.items[index].Quantity < quantity {
			continue
		}
		inventory.items[index].Quantity -= quantity
		if inventory.items[index].Quantity == 0 {
			inventory.items = append(inventory.items[:index], inventory.items[index+1:]...)
		}
		return true
	}
	return false
}

func snapshotInventory(inventory inventoryState) InventorySnapshot {
	return InventorySnapshot{
		Capacity: inventory.capacity,
		Items:    append([]InventoryItem(nil), inventory.items...),
	}
}

func inventoryFromSnapshot(snapshot InventorySnapshot) (inventoryState, error) {
	if snapshot.Capacity <= 0 || snapshot.Capacity > 100 {
		return inventoryState{}, fmt.Errorf("inventory capacity must be between 1 and 100")
	}
	if len(snapshot.Items) > snapshot.Capacity {
		return inventoryState{}, fmt.Errorf("inventory has more items than its capacity")
	}

	items := append([]InventoryItem(nil), snapshot.Items...)
	occupiedSlots := make(map[int]struct{}, len(items))
	itemIDs := make(map[string]struct{}, len(items))
	for _, item := range items {
		if item.Slot < 0 || item.Slot >= snapshot.Capacity {
			return inventoryState{}, fmt.Errorf("inventory slot %d is outside its capacity", item.Slot)
		}
		if item.ItemID == "" || item.Quantity <= 0 {
			return inventoryState{}, fmt.Errorf("inventory slot %d has an invalid item", item.Slot)
		}
		if _, exists := occupiedSlots[item.Slot]; exists {
			return inventoryState{}, fmt.Errorf("inventory slot %d is duplicated", item.Slot)
		}
		if _, exists := itemIDs[item.ItemID]; exists {
			return inventoryState{}, fmt.Errorf("inventory item %q is duplicated", item.ItemID)
		}
		occupiedSlots[item.Slot] = struct{}{}
		itemIDs[item.ItemID] = struct{}{}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Slot < items[j].Slot })
	return inventoryState{capacity: snapshot.Capacity, items: items}, nil
}
