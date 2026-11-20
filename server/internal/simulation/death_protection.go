package simulation

import (
	"fmt"
	"sort"
	"strings"
)

const BaseKeptItemsOnDeath = 4

type SetDeathProtectionCommand struct {
	PlayerID PlayerID
	Sequence uint64
	Slots    []int
	Equipped []string
	entries  map[int]protectedInventoryEntry
}

type protectedInventoryEntry struct {
	Slot       int
	ItemID     string
	InstanceID string
}

func (w *World) QueueSetDeathProtection(command SetDeathProtectionCommand) error {
	if command.Sequence == 0 {
		return fmt.Errorf("death protection sequence must be positive")
	}
	if len(command.Slots)+len(command.Equipped) > BaseKeptItemsOnDeath {
		return fmt.Errorf("at most %d items can be protected", BaseKeptItemsOnDeath)
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
	if command.Sequence <= player.lastDeathProtectionSequence {
		return fmt.Errorf("death protection sequence %d is not newer than %d", command.Sequence, player.lastDeathProtectionSequence)
	}
	if pending, exists := w.pendingDeathProtections[command.PlayerID]; exists && command.Sequence <= pending.Sequence {
		return fmt.Errorf("death protection sequence %d is not newer than pending sequence %d", command.Sequence, pending.Sequence)
	}

	entries, err := restoreProtectedInventory(player.inventory, command.Slots)
	if err != nil {
		return err
	}
	if _, err := validateEquippedProtection(player, command.Equipped); err != nil {
		return err
	}
	command.entries = entries
	command.Equipped = append([]string(nil), command.Equipped...)
	w.pendingDeathProtections[command.PlayerID] = command
	return nil
}

func (w *World) applyPendingDeathProtections() {
	playerIDs := make([]PlayerID, 0, len(w.pendingDeathProtections))
	for id := range w.pendingDeathProtections {
		playerIDs = append(playerIDs, id)
	}
	sort.Slice(playerIDs, func(i, j int) bool { return playerIDs[i] < playerIDs[j] })
	for _, id := range playerIDs {
		command := w.pendingDeathProtections[id]
		delete(w.pendingDeathProtections, id)
		player, exists := w.players[id]
		if !exists || player.defeated {
			continue
		}
		// Always acknowledge a queued selection, even when its items moved before the tick.
		player.lastDeathProtectionSequence = command.Sequence
		player.inventory.revision++
		equipped, err := validateEquippedProtection(player, command.Equipped)
		if err != nil {
			continue
		}
		valid := true
		for slot, entry := range command.entries {
			item, ok := player.inventory.itemAtSlot(slot)
			if !ok || !protectedEntryMatchesItem(entry, item) {
				valid = false
				break
			}
		}
		if !valid {
			continue
		}
		player.protectedInventory = command.entries
		player.protectedEquipment = equipped
	}
}

func validateEquippedProtection(player *playerState, keys []string) (map[string]bool, error) {
	var selected map[string]bool
	for _, key := range keys {
		if !hasEquippedKey(player, key) || selected[key] {
			return nil, fmt.Errorf("invalid or duplicate equipped protection selection")
		}
		if selected == nil {
			selected = make(map[string]bool)
		}
		selected[key] = true
	}
	return selected, nil
}

func protectedEquipmentKeys(player *playerState) []string {
	var keys []string
	for key := range player.protectedEquipment {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func protectedInventorySlots(player *playerState) []int {
	slots := make([]int, 0, len(player.protectedInventory))
	for slot, entry := range player.protectedInventory {
		item, exists := player.inventory.itemAtSlot(slot)
		if exists && protectedEntryMatchesItem(entry, item) {
			slots = append(slots, slot)
		}
	}
	sort.Ints(slots)
	return slots
}

func protectedEntryMatchesItem(entry protectedInventoryEntry, item InventoryItem) bool {
	return entry.Slot == item.Slot && entry.ItemID == item.ItemID && entry.InstanceID == item.InstanceID
}

func moveProtectedInventorySlots(player *playerState, fromSlot, toSlot int) {
	from, fromProtected := player.protectedInventory[fromSlot]
	to, toProtected := player.protectedInventory[toSlot]
	delete(player.protectedInventory, fromSlot)
	delete(player.protectedInventory, toSlot)
	if fromProtected {
		from.Slot = toSlot
		player.protectedInventory[toSlot] = from
	}
	if toProtected {
		to.Slot = fromSlot
		player.protectedInventory[fromSlot] = to
	}
}

func reconcileDeathProtection(player *playerState) bool {
	changed := false
	for slot, entry := range player.protectedInventory {
		item, exists := player.inventory.itemAtSlot(slot)
		if exists && protectedEntryMatchesItem(entry, item) {
			continue
		}
		if entry.InstanceID != "" && hasEquippedKey(player, "gear:"+entry.InstanceID) {
			if player.protectedEquipment == nil {
				player.protectedEquipment = make(map[string]bool)
			}
			player.protectedEquipment["gear:"+entry.InstanceID] = true
		}
		delete(player.protectedInventory, slot)
		changed = true
	}
	for key := range player.protectedEquipment {
		if hasEquippedKey(player, key) {
			continue
		}
		if instance, ok := strings.CutPrefix(key, "gear:"); ok {
			if item, found := player.inventory.itemByInstance(instance); found {
				if player.protectedInventory == nil {
					player.protectedInventory = make(map[int]protectedInventoryEntry)
				}
				player.protectedInventory[item.Slot] = protectedInventoryEntry{Slot: item.Slot, ItemID: item.ItemID, InstanceID: instance}
			}
		}
		delete(player.protectedEquipment, key)
		changed = true
	}
	return changed
}

func restoreProtectedInventory(inventory inventoryState, slots []int) (map[int]protectedInventoryEntry, error) {
	if len(slots) > BaseKeptItemsOnDeath {
		return nil, fmt.Errorf("at most %d items can be protected", BaseKeptItemsOnDeath)
	}
	entries := make(map[int]protectedInventoryEntry, len(slots))
	for _, slot := range slots {
		if _, duplicate := entries[slot]; duplicate {
			return nil, fmt.Errorf("protected inventory slot %d is duplicated", slot)
		}
		item, exists := inventory.itemAtSlot(slot)
		if !exists {
			return nil, fmt.Errorf("protected inventory slot %d is empty", slot)
		}
		entries[slot] = protectedInventoryEntry{Slot: slot, ItemID: item.ItemID, InstanceID: item.InstanceID}
	}
	return entries, nil
}

func (w *World) dropUnprotectedItemsOnDeath(player *playerState) int {
	reconcileDeathProtection(player)
	dropped := 0
	kept := make([]InventoryItem, 0, BaseKeptItemsOnDeath)
	for _, item := range player.inventory.items {
		entry, protected := player.protectedInventory[item.Slot]
		if protected && protectedEntryMatchesItem(entry, item) {
			kept = append(kept, item)
			continue
		}
		w.spawnInventoryItemOrStackDrop(player.position, item)
		dropped++
	}
	player.inventory.items = kept
	player.inventory.revision++
	for _, item := range equippedItems(player) {
		if player.protectedEquipment[item.key] {
			continue
		}
		w.spawnInventoryItemOrStackDrop(player.position, item.InventoryItem)
		item.remove(player)
		dropped++
	}
	player.equipmentRevision++
	refreshPlayerMaximumHealth(player)
	return dropped
}

func (w *World) spawnInventoryItemOrStackDrop(position Position, item InventoryItem) {
	if item.InstanceID != "" {
		w.spawnInventoryItemDrop(position, item)
		return
	}
	w.spawnDrop(position, item.ItemID, item.Quantity)
}
