package simulation

import (
	"fmt"
	"sort"
)

const (
	RunePouchCapacity          = 9
	BaseRunePouchUnlockedSlots = 2
)

type RunePouchSnapshot struct {
	Capacity      int
	UnlockedSlots int
	Items         []InventoryItem
}

type RuneTransferCommand struct {
	PlayerID      PlayerID
	Sequence      uint64
	Deposit       bool
	InventorySlot int
	PouchSlot     int
}

func isRune(itemID string) bool {
	switch itemID {
	case ItemAirRune, ItemWaterRune, ItemEarthRune, ItemFireRune, ItemHolyRune, ItemDarkRune, ItemFocusRune:
		return true
	default:
		return false
	}
}

func newRunePouch() inventoryState { return inventoryState{capacity: BaseRunePouchUnlockedSlots} }

func snapshotRunePouch(pouch inventoryState) RunePouchSnapshot {
	return RunePouchSnapshot{
		Capacity: RunePouchCapacity, UnlockedSlots: pouch.capacity,
		Items: append([]InventoryItem(nil), pouch.items...),
	}
}

func runePouchFromSnapshot(snapshot RunePouchSnapshot) (inventoryState, error) {
	if snapshot.Capacity == 0 && snapshot.UnlockedSlots == 0 && len(snapshot.Items) == 0 {
		return newRunePouch(), nil
	}
	if snapshot.Capacity != RunePouchCapacity || snapshot.UnlockedSlots < BaseRunePouchUnlockedSlots ||
		snapshot.UnlockedSlots > snapshot.Capacity {
		return inventoryState{}, fmt.Errorf(
			"rune pouch must have %d total slots and between %d and %d unlocked slots",
			RunePouchCapacity, BaseRunePouchUnlockedSlots, RunePouchCapacity,
		)
	}
	pouch := inventoryState{capacity: snapshot.UnlockedSlots, items: append([]InventoryItem(nil), snapshot.Items...)}
	seenItems := make(map[string]bool, len(pouch.items))
	seenSlots := make(map[int]bool, len(pouch.items))
	for _, item := range pouch.items {
		if !isRune(item.ItemID) || item.Slot < 0 || item.Slot >= pouch.capacity || item.Quantity <= 0 ||
			item.InstanceID != "" || item.Reinforcement != 0 || seenItems[item.ItemID] || seenSlots[item.Slot] {
			return inventoryState{}, fmt.Errorf("rune pouch slot %d is invalid", item.Slot)
		}
		seenItems[item.ItemID] = true
		seenSlots[item.Slot] = true
	}
	sort.Slice(pouch.items, func(i, j int) bool { return pouch.items[i].Slot < pouch.items[j].Slot })
	return pouch, nil
}

func (w *World) QueueRuneTransfer(command RuneTransferCommand) error {
	if command.Sequence == 0 {
		return fmt.Errorf("rune transfer sequence must be positive")
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	player, exists := w.players[command.PlayerID]
	if !exists || player.defeated {
		return fmt.Errorf("player %q cannot use the rune pouch", command.PlayerID)
	}
	if command.Sequence <= player.lastRuneSequence {
		return fmt.Errorf("rune sequence %d is not newer than %d", command.Sequence, player.lastRuneSequence)
	}
	if pending, exists := w.pendingRuneTransfers[command.PlayerID]; exists && command.Sequence <= pending.Sequence {
		return fmt.Errorf("rune sequence %d is not newer than pending sequence %d", command.Sequence, pending.Sequence)
	}
	if _, _, err := runeTransferAfter(player, command); err != nil {
		return err
	}
	w.pendingRuneTransfers[command.PlayerID] = command
	return nil
}

func runeTransferAfter(player *playerState, command RuneTransferCommand) (inventoryState, inventoryState, error) {
	inventory := cloneInventory(player.inventory)
	pouch := cloneInventory(player.runePouch)
	if command.Deposit {
		item, exists := inventoryItemAt(inventory, command.InventorySlot)
		if !exists || !isRune(item.ItemID) {
			return inventoryState{}, inventoryState{}, fmt.Errorf("inventory slot %d does not contain runes", command.InventorySlot)
		}
		if _, removed := removeInventorySlot(&inventory, command.InventorySlot); !removed || !pouch.add(item.ItemID, item.Quantity) {
			return inventoryState{}, inventoryState{}, fmt.Errorf("rune pouch has no slot for %q", item.ItemID)
		}
		return inventory, pouch, nil
	}
	item, exists := inventoryItemAt(pouch, command.PouchSlot)
	if !exists {
		return inventoryState{}, inventoryState{}, fmt.Errorf("rune pouch slot %d is empty", command.PouchSlot)
	}
	if !inventory.add(item.ItemID, item.Quantity) {
		return inventoryState{}, inventoryState{}, fmt.Errorf("inventory is full")
	}
	if _, removed := removeInventorySlot(&pouch, command.PouchSlot); !removed {
		return inventoryState{}, inventoryState{}, fmt.Errorf("rune pouch slot %d is empty", command.PouchSlot)
	}
	return inventory, pouch, nil
}

func (w *World) applyPendingRuneTransfers() {
	for id, command := range w.pendingRuneTransfers {
		delete(w.pendingRuneTransfers, id)
		player, exists := w.players[id]
		if !exists || player.defeated {
			continue
		}
		inventory, pouch, err := runeTransferAfter(player, command)
		if err != nil {
			continue
		}
		player.inventory = inventory
		player.runePouch = pouch
		player.lastRuneSequence = command.Sequence
	}
}

func paySpellCosts(player *playerState, costs []ItemCost) bool {
	if player == nil {
		return false
	}
	inventory := cloneInventory(player.inventory)
	pouch := cloneInventory(player.runePouch)
	for _, cost := range costs {
		remaining := cost.Quantity
		fromPouch := min(remaining, pouch.quantity(cost.ItemID))
		if fromPouch > 0 {
			pouch.remove(cost.ItemID, fromPouch)
			remaining -= fromPouch
		}
		if remaining > 0 && !inventory.remove(cost.ItemID, remaining) {
			return false
		}
	}
	player.inventory = inventory
	player.runePouch = pouch
	return true
}
