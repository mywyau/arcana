package simulation

import (
	"fmt"
	"sort"
)

const (
	BaseInventoryCapacity = 8
	MaxInventoryCapacity  = 50
	ItemAirRune           = "air_rune"
)

type ItemCost struct {
	ItemID   string
	Quantity int
}

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

type MoveInventoryItemCommand struct {
	PlayerID PlayerID
	Sequence uint64
	FromSlot int
	ToSlot   int
}

func newEmptyInventory() inventoryState {
	return inventoryState{capacity: BaseInventoryCapacity}
}

func (inventory inventoryState) contains(itemID string) bool {
	return inventory.quantity(itemID) > 0
}

func (inventory inventoryState) quantity(itemID string) int {
	for _, item := range inventory.items {
		if item.ItemID == itemID && item.Quantity > 0 {
			return item.Quantity
		}
	}
	return 0
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

func (inventory *inventoryState) swapSlots(fromSlot, toSlot int) bool {
	if fromSlot < 0 || fromSlot >= inventory.capacity || toSlot < 0 || toSlot >= inventory.capacity || fromSlot == toSlot {
		return false
	}
	fromIndex := -1
	toIndex := -1
	for index := range inventory.items {
		if inventory.items[index].Slot == fromSlot {
			fromIndex = index
		}
		if inventory.items[index].Slot == toSlot {
			toIndex = index
		}
	}
	if fromIndex < 0 {
		return false
	}
	inventory.items[fromIndex].Slot = toSlot
	if toIndex >= 0 {
		inventory.items[toIndex].Slot = fromSlot
	}
	sort.Slice(inventory.items, func(i, j int) bool { return inventory.items[i].Slot < inventory.items[j].Slot })
	return true
}

func (w *World) QueueMoveInventoryItem(command MoveInventoryItemCommand) error {
	if command.Sequence == 0 {
		return fmt.Errorf("inventory move sequence must be positive")
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	player, exists := w.players[command.PlayerID]
	if !exists {
		return fmt.Errorf("player %q does not exist", command.PlayerID)
	}
	if player.defeated {
		return fmt.Errorf("player %q is defeated", command.PlayerID)
	}
	if command.FromSlot < 0 || command.FromSlot >= player.inventory.capacity ||
		command.ToSlot < 0 || command.ToSlot >= player.inventory.capacity || command.FromSlot == command.ToSlot {
		return fmt.Errorf("inventory move from %d to %d is invalid", command.FromSlot, command.ToSlot)
	}
	if command.Sequence <= player.lastInventoryMoveSequence {
		return fmt.Errorf("inventory move sequence %d is not newer than %d", command.Sequence, player.lastInventoryMoveSequence)
	}
	if pending, found := w.pendingInventoryMoves[command.PlayerID]; found && command.Sequence <= pending.Sequence {
		return fmt.Errorf("inventory move sequence %d is not newer than pending sequence %d", command.Sequence, pending.Sequence)
	}
	itemFound := false
	for _, item := range player.inventory.items {
		if item.Slot == command.FromSlot {
			itemFound = true
			break
		}
	}
	if !itemFound {
		return fmt.Errorf("inventory slot %d is empty", command.FromSlot)
	}
	w.pendingInventoryMoves[command.PlayerID] = command
	return nil
}

func (w *World) applyPendingInventoryMoves() {
	playerIDs := make([]PlayerID, 0, len(w.pendingInventoryMoves))
	for id := range w.pendingInventoryMoves {
		playerIDs = append(playerIDs, id)
	}
	sort.Slice(playerIDs, func(i, j int) bool { return playerIDs[i] < playerIDs[j] })
	for _, id := range playerIDs {
		command := w.pendingInventoryMoves[id]
		delete(w.pendingInventoryMoves, id)
		player, exists := w.players[id]
		if !exists || player.defeated {
			continue
		}
		if player.inventory.swapSlots(command.FromSlot, command.ToSlot) {
			player.lastInventoryMoveSequence = command.Sequence
		}
	}
}

func snapshotInventory(inventory inventoryState) InventorySnapshot {
	return InventorySnapshot{
		Capacity: inventory.capacity,
		Items:    append([]InventoryItem(nil), inventory.items...),
	}
}

func inventoryFromSnapshot(snapshot InventorySnapshot) (inventoryState, error) {
	if snapshot.Capacity <= 0 || snapshot.Capacity > MaxInventoryCapacity {
		return inventoryState{}, fmt.Errorf("inventory capacity must be between 1 and %d", MaxInventoryCapacity)
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
		if definition, exists := itemDefinition(item.ItemID); !exists || definition.Virtual {
			return inventoryState{}, fmt.Errorf("inventory slot %d contains unknown item %q", item.Slot, item.ItemID)
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
