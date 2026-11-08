package simulation

import (
	"crypto/rand"
	"encoding/hex"
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
	Slot          int
	ItemID        string
	Quantity      int
	InstanceID    string
	Reinforcement int
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
	total := 0
	for _, item := range inventory.items {
		if item.ItemID == itemID && item.Quantity > 0 {
			total += item.Quantity
		}
	}
	return total
}

func (inventory inventoryState) canAdd(itemID string, quantity int) bool {
	if itemID == "" || quantity <= 0 {
		return false
	}
	if itemIsInstanced(itemID) {
		return len(inventory.items)+quantity <= inventory.capacity
	}
	stackLimit := itemStackLimit(itemID)
	if stackLimit == 0 {
		return inventory.contains(itemID) || len(inventory.items) < inventory.capacity
	}
	available := (inventory.capacity - len(inventory.items)) * stackLimit
	for _, item := range inventory.items {
		if item.ItemID == itemID && item.Quantity < stackLimit {
			available += stackLimit - item.Quantity
		}
	}
	return available >= quantity
}

func (inventory *inventoryState) add(itemID string, quantity int) bool {
	if !inventory.canAdd(itemID, quantity) {
		return false
	}
	stackLimit := itemStackLimit(itemID)
	if itemIsInstanced(itemID) {
		for count := 0; count < quantity; count++ {
			if !inventory.addInventoryItem(InventoryItem{ItemID: itemID, Quantity: 1, InstanceID: newItemInstanceID()}) {
				return false
			}
		}
		return true
	}
	for index := range inventory.items {
		if inventory.items[index].ItemID != itemID {
			continue
		}
		if stackLimit == 0 {
			inventory.items[index].Quantity += quantity
			return true
		}
		space := stackLimit - inventory.items[index].Quantity
		added := min(space, quantity)
		inventory.items[index].Quantity += added
		quantity -= added
		if quantity == 0 {
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
		stackQuantity := quantity
		if stackLimit > 0 {
			stackQuantity = min(stackLimit, quantity)
		}
		inventory.items = append(inventory.items, InventoryItem{Slot: slot, ItemID: itemID, Quantity: stackQuantity})
		quantity -= stackQuantity
		if quantity == 0 {
			sort.Slice(inventory.items, func(i, j int) bool { return inventory.items[i].Slot < inventory.items[j].Slot })
			return true
		}
	}
	return false
}

func (inventory *inventoryState) addInventoryItem(item InventoryItem) bool {
	if item.ItemID == "" || item.Quantity <= 0 || len(inventory.items) >= inventory.capacity {
		return false
	}
	occupied := make(map[int]struct{}, len(inventory.items))
	for _, existing := range inventory.items {
		occupied[existing.Slot] = struct{}{}
	}
	for slot := 0; slot < inventory.capacity; slot++ {
		if _, exists := occupied[slot]; exists {
			continue
		}
		item.Slot = slot
		inventory.items = append(inventory.items, item)
		sort.Slice(inventory.items, func(i, j int) bool { return inventory.items[i].Slot < inventory.items[j].Slot })
		return true
	}
	return false
}

func (inventory inventoryState) itemByInstance(instanceID string) (InventoryItem, bool) {
	for _, item := range inventory.items {
		if item.InstanceID == instanceID && instanceID != "" {
			return item, true
		}
	}
	return InventoryItem{}, false
}

func (inventory *inventoryState) removeInstance(instanceID string) (InventoryItem, bool) {
	for index, item := range inventory.items {
		if item.InstanceID != instanceID || instanceID == "" {
			continue
		}
		inventory.items = append(inventory.items[:index], inventory.items[index+1:]...)
		return item, true
	}
	return InventoryItem{}, false
}

func itemIsInstanced(itemID string) bool {
	definition, exists := itemDefinition(itemID)
	return exists && (definition.Kind == ItemKindWeapon || definition.Kind == ItemKindArmour)
}

func newItemInstanceID() string {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		panic(fmt.Sprintf("generate item instance ID: %v", err))
	}
	return hex.EncodeToString(value[:])
}

func (inventory *inventoryState) remove(itemID string, quantity int) bool {
	if itemID == "" || quantity <= 0 {
		return false
	}
	if inventory.quantity(itemID) < quantity {
		return false
	}
	for index := 0; index < len(inventory.items) && quantity > 0; {
		if inventory.items[index].ItemID != itemID {
			index++
			continue
		}
		removed := min(inventory.items[index].Quantity, quantity)
		inventory.items[index].Quantity -= removed
		quantity -= removed
		if inventory.items[index].Quantity == 0 {
			inventory.items = append(inventory.items[:index], inventory.items[index+1:]...)
			continue
		}
		index++
	}
	return true
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
	expandedItems, err := expandLegacyEquipmentStacks(snapshot.Items, snapshot.Capacity)
	if err != nil {
		return inventoryState{}, err
	}
	snapshot.Items = expandedItems
	if len(snapshot.Items) > snapshot.Capacity {
		return inventoryState{}, fmt.Errorf("inventory has more items than its capacity")
	}

	items := append([]InventoryItem(nil), snapshot.Items...)
	occupiedSlots := make(map[int]struct{}, len(items))
	instanceIDs := make(map[string]struct{}, len(items))
	for index := range items {
		item := &items[index]
		if item.Slot < 0 || item.Slot >= snapshot.Capacity {
			return inventoryState{}, fmt.Errorf("inventory slot %d is outside its capacity", item.Slot)
		}
		if item.ItemID == "" || item.Quantity <= 0 {
			return inventoryState{}, fmt.Errorf("inventory slot %d has an invalid item", item.Slot)
		}
		definition, exists := itemDefinition(item.ItemID)
		if !exists || definition.Virtual {
			return inventoryState{}, fmt.Errorf("inventory slot %d contains unknown item %q", item.Slot, item.ItemID)
		}
		if itemIsInstanced(item.ItemID) {
			if item.Quantity != 1 {
				return inventoryState{}, fmt.Errorf("instanced inventory item %q must have quantity 1", item.ItemID)
			}
			if item.InstanceID == "" {
				item.InstanceID = newItemInstanceID()
			}
			if item.Reinforcement < 0 || item.Reinforcement > maximumReinforcementLevel {
				return inventoryState{}, fmt.Errorf("instanced inventory item %q has invalid reinforcement +%d", item.ItemID, item.Reinforcement)
			}
			if _, duplicate := instanceIDs[item.InstanceID]; duplicate {
				return inventoryState{}, fmt.Errorf("inventory instance %q is duplicated", item.InstanceID)
			}
			instanceIDs[item.InstanceID] = struct{}{}
		} else if item.InstanceID != "" || item.Reinforcement != 0 {
			return inventoryState{}, fmt.Errorf("stacked item %q cannot have equipment metadata", item.ItemID)
		}
		if definition.StackLimit > 0 && item.Quantity > definition.StackLimit {
			return inventoryState{}, fmt.Errorf("inventory slot %d exceeds the stack limit for %q", item.Slot, item.ItemID)
		}
		if _, exists := occupiedSlots[item.Slot]; exists {
			return inventoryState{}, fmt.Errorf("inventory slot %d is duplicated", item.Slot)
		}
		occupiedSlots[item.Slot] = struct{}{}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Slot < items[j].Slot })
	return inventoryState{capacity: snapshot.Capacity, items: items}, nil
}

func expandLegacyEquipmentStacks(items []InventoryItem, capacity int) ([]InventoryItem, error) {
	expanded := append([]InventoryItem(nil), items...)
	occupied := make(map[int]struct{}, len(expanded))
	for _, item := range expanded {
		occupied[item.Slot] = struct{}{}
	}
	for index := 0; index < len(expanded); index++ {
		if !itemIsInstanced(expanded[index].ItemID) || expanded[index].Quantity <= 1 {
			continue
		}
		extra := expanded[index].Quantity - 1
		expanded[index].Quantity = 1
		for count := 0; count < extra; count++ {
			freeSlot := -1
			for slot := 0; slot < capacity; slot++ {
				if _, used := occupied[slot]; !used {
					freeSlot = slot
					break
				}
			}
			if freeSlot < 0 {
				return nil, fmt.Errorf("legacy equipment stack %q exceeds inventory capacity", expanded[index].ItemID)
			}
			occupied[freeSlot] = struct{}{}
			expanded = append(expanded, InventoryItem{Slot: freeSlot, ItemID: expanded[index].ItemID, Quantity: 1})
		}
	}
	return expanded, nil
}
