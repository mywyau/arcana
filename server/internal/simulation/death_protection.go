package simulation

import (
	"fmt"
	"sort"
)

const BaseKeptItemsOnDeath = 4

type SetDeathProtectionCommand struct {
	PlayerID PlayerID
	Sequence uint64
	Slots    []int
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
	if len(command.Slots) > BaseKeptItemsOnDeath {
		return fmt.Errorf("at most %d inventory items can be protected", BaseKeptItemsOnDeath)
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

	seen := make(map[int]struct{}, len(command.Slots))
	for _, slot := range command.Slots {
		if slot < 0 || slot >= player.inventory.capacity {
			return fmt.Errorf("protected inventory slot %d is invalid", slot)
		}
		if _, duplicate := seen[slot]; duplicate {
			return fmt.Errorf("protected inventory slot %d is duplicated", slot)
		}
		seen[slot] = struct{}{}
		if _, occupied := player.inventory.itemAtSlot(slot); !occupied {
			return fmt.Errorf("protected inventory slot %d is empty", slot)
		}
	}
	command.Slots = append([]int(nil), command.Slots...)
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
		entries := make(map[int]protectedInventoryEntry, len(command.Slots))
		valid := true
		for _, slot := range command.Slots {
			item, exists := player.inventory.itemAtSlot(slot)
			if !exists {
				valid = false
				break
			}
			entries[slot] = protectedInventoryEntry{Slot: slot, ItemID: item.ItemID, InstanceID: item.InstanceID}
		}
		if !valid {
			continue
		}
		player.protectedInventory = entries
		player.lastDeathProtectionSequence = command.Sequence
		player.inventory.revision++
	}
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

func pruneProtectedInventory(player *playerState) bool {
	changed := false
	for slot, entry := range player.protectedInventory {
		item, exists := player.inventory.itemAtSlot(slot)
		if exists && protectedEntryMatchesItem(entry, item) {
			continue
		}
		delete(player.protectedInventory, slot)
		changed = true
	}
	return changed
}

func restoreProtectedInventory(inventory inventoryState, slots []int) (map[int]protectedInventoryEntry, error) {
	if len(slots) > BaseKeptItemsOnDeath {
		return nil, fmt.Errorf("at most %d inventory items can be protected", BaseKeptItemsOnDeath)
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
	pruneProtectedInventory(player)

	if player.equippedWeapon != WeaponUnarmed {
		w.spawnInventoryItemDrop(player.position, InventoryItem{
			ItemID: string(player.equippedWeapon), Quantity: 1,
			InstanceID: player.equippedWeaponInstanceID, Reinforcement: player.equippedWeaponReinforcement,
		})
		dropped++
		player.equippedWeapon = WeaponUnarmed
		player.equippedWeaponInstanceID = ""
		player.equippedWeaponReinforcement = 0
		player.combatStyle = defaultCombatStyle(WeaponUnarmed)
	}
	for slot, itemID := range player.equippedArmour {
		w.spawnInventoryItemDrop(player.position, InventoryItem{
			ItemID: itemID, Quantity: 1,
			InstanceID: player.equippedArmourInstances[slot], Reinforcement: player.equippedArmourReinforcements[slot],
		})
		dropped++
	}
	clear(player.equippedArmour)
	clear(player.equippedArmourInstances)
	clear(player.equippedArmourReinforcements)
	if player.equippedSpellbook != "" {
		w.spawnDrop(player.position, player.equippedSpellbook, 1)
		dropped++
		player.equippedSpellbook = ""
		player.selectedSpell = ""
	}
	player.equipmentRevision++

	dropped += w.dropContainerItems(player.position, &player.ammunitionPouch)
	dropped += w.dropContainerItems(player.position, &player.ringRack)
	dropped += w.dropContainerItems(player.position, &player.runePouch)
	player.maximumHealth = maximumHealthForSkills(player.skills)
	return dropped
}

func (w *World) dropContainerItems(position Position, container *inventoryState) int {
	for _, item := range container.items {
		w.spawnInventoryItemOrStackDrop(position, item)
	}
	dropped := len(container.items)
	if dropped > 0 {
		container.items = nil
		container.revision++
	}
	return dropped
}

func (w *World) spawnInventoryItemOrStackDrop(position Position, item InventoryItem) {
	if item.InstanceID != "" {
		w.spawnInventoryItemDrop(position, item)
		return
	}
	w.spawnDrop(position, item.ItemID, item.Quantity)
}
