package simulation

import (
	"fmt"
	"sort"
)

const (
	pickupRange       = 3.0
	trainingShardItem = "training_shard"
)

type DropSnapshot struct {
	ID       string
	Position Position
	ItemID   string
	Quantity int
}

type PickupCommand struct {
	PlayerID PlayerID
	Sequence uint64
	DropID   string
}

type DropItemCommand struct {
	PlayerID PlayerID
	Sequence uint64
	ItemID   string
}

func (w *World) QueueDropItem(command DropItemCommand) error {
	if command.Sequence == 0 {
		return fmt.Errorf("drop sequence must be positive")
	}
	if command.ItemID == "" {
		return fmt.Errorf("item ID is required")
	}
	_, weaponExists := weaponDefinition(WeaponID(command.ItemID))
	_, armourExists := armourSlot(command.ItemID)
	if !weaponExists && !armourExists {
		return fmt.Errorf("item %q is not droppable equipment", command.ItemID)
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
	if command.Sequence <= player.lastDropSequence {
		return fmt.Errorf("drop sequence %d is not newer than %d", command.Sequence, player.lastDropSequence)
	}
	if pending, exists := w.pendingItemDrops[command.PlayerID]; exists && command.Sequence <= pending.Sequence {
		return fmt.Errorf("drop sequence %d is not newer than pending sequence %d", command.Sequence, pending.Sequence)
	}
	if !player.inventory.contains(command.ItemID) {
		return fmt.Errorf("player %q does not own item %q", command.PlayerID, command.ItemID)
	}
	if command.ItemID == string(player.equippedWeapon) {
		return fmt.Errorf("item %q is currently equipped", command.ItemID)
	}
	for _, equippedItemID := range player.equippedArmour {
		if command.ItemID == equippedItemID {
			return fmt.Errorf("item %q is currently equipped", command.ItemID)
		}
	}
	w.pendingItemDrops[command.PlayerID] = command
	return nil
}

func (w *World) QueuePickup(command PickupCommand) error {
	if command.Sequence == 0 {
		return fmt.Errorf("pickup sequence must be positive")
	}
	if command.DropID == "" {
		return fmt.Errorf("drop ID is required")
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
	if command.Sequence <= player.lastPickupSequence {
		return fmt.Errorf("pickup sequence %d is not newer than %d", command.Sequence, player.lastPickupSequence)
	}
	if pending, exists := w.pendingPickups[command.PlayerID]; exists && command.Sequence <= pending.Sequence {
		return fmt.Errorf("pickup sequence %d is not newer than pending sequence %d", command.Sequence, pending.Sequence)
	}
	drop, exists := w.drops[command.DropID]
	if !exists {
		return fmt.Errorf("drop %q does not exist", command.DropID)
	}
	if distance(player.position, drop.Position) > pickupRange {
		return fmt.Errorf("drop %q is outside the %.1f unit pickup range", command.DropID, pickupRange)
	}
	if !player.inventory.canAdd(drop.ItemID) {
		return fmt.Errorf("inventory is full")
	}
	w.pendingPickups[command.PlayerID] = command
	return nil
}

func (w *World) applyPendingPickups() {
	playerIDs := make([]PlayerID, 0, len(w.pendingPickups))
	for id := range w.pendingPickups {
		playerIDs = append(playerIDs, id)
	}
	sort.Slice(playerIDs, func(i, j int) bool { return playerIDs[i] < playerIDs[j] })
	for _, id := range playerIDs {
		command := w.pendingPickups[id]
		delete(w.pendingPickups, id)
		player, exists := w.players[id]
		if !exists || player.defeated {
			continue
		}
		player.lastPickupSequence = command.Sequence
		drop, exists := w.drops[command.DropID]
		if !exists || distance(player.position, drop.Position) > pickupRange {
			continue
		}
		if player.inventory.add(drop.ItemID, drop.Quantity) {
			delete(w.drops, command.DropID)
		}
	}
}

func (w *World) applyPendingItemDrops() {
	playerIDs := make([]PlayerID, 0, len(w.pendingItemDrops))
	for id := range w.pendingItemDrops {
		playerIDs = append(playerIDs, id)
	}
	sort.Slice(playerIDs, func(i, j int) bool { return playerIDs[i] < playerIDs[j] })
	for _, id := range playerIDs {
		command := w.pendingItemDrops[id]
		delete(w.pendingItemDrops, id)
		player, exists := w.players[id]
		if !exists || player.defeated {
			continue
		}
		player.lastDropSequence = command.Sequence
		if command.ItemID == string(player.equippedWeapon) {
			continue
		}
		equipped := false
		for _, equippedItemID := range player.equippedArmour {
			if command.ItemID == equippedItemID {
				equipped = true
				break
			}
		}
		if !equipped && player.inventory.remove(command.ItemID, 1) {
			w.spawnDrop(player.position, command.ItemID, 1)
		}
	}
}

func (w *World) spawnDrop(position Position, itemID string, quantity int) {
	if itemID == "" || quantity <= 0 {
		return
	}
	w.nextDropID++
	id := fmt.Sprintf("drop-%d", w.nextDropID)
	w.drops[id] = DropSnapshot{ID: id, Position: position, ItemID: itemID, Quantity: quantity}
}
