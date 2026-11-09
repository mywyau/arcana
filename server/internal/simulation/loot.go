package simulation

import (
	"fmt"
	"sort"
	"time"
)

const (
	pickupRange       = 3.0
	worldDropLifetime = 2 * time.Minute
)

type DropSnapshot struct {
	ID            string
	Position      Position
	ItemID        string
	Quantity      int
	InstanceID    string
	Reinforcement int
	Infinite      bool
}

type PickupCommand struct {
	PlayerID PlayerID
	Sequence uint64
	DropID   string
}

type DropItemCommand struct {
	PlayerID   PlayerID
	Sequence   uint64
	ItemID     string
	InstanceID string
}

type respawningDrop struct {
	Position     Position
	ItemID       string
	Quantity     int
	RespawnAfter time.Duration
	Remaining    time.Duration
}

// AddWorldDrop places initial world loot before the simulation starts.
func (w *World) AddWorldDrop(position Position, itemID string, quantity int) error {
	return w.addInitialDrop(position, itemID, quantity, false)
}

// AddRespawningWorldDrop places ordinary shared world loot that reappears a
// fixed time after a successful pickup.
func (w *World) AddRespawningWorldDrop(
	position Position,
	itemID string,
	quantity int,
	respawnAfter time.Duration,
) error {
	if err := validateInitialDrop(position, itemID, quantity); err != nil {
		return err
	}
	if respawnAfter <= 0 {
		return fmt.Errorf("drop respawn time must be positive")
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	id := w.addDrop(position, itemID, quantity, false)
	w.respawningDrops[id] = respawningDrop{
		Position: position, ItemID: itemID, Quantity: quantity, RespawnAfter: respawnAfter,
	}
	return nil
}

// AddTestSupply places a non-depleting item source used by development chests.
func (w *World) AddTestSupply(position Position, itemID string, quantity int) error {
	return w.addInitialDrop(position, itemID, quantity, true)
}

func (w *World) addInitialDrop(position Position, itemID string, quantity int, infinite bool) error {
	if err := validateInitialDrop(position, itemID, quantity); err != nil {
		return err
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.addDrop(position, itemID, quantity, infinite)
	return nil
}

func validateInitialDrop(position Position, itemID string, quantity int) error {
	if !position.valid() {
		return fmt.Errorf("drop position must be finite")
	}
	if quantity <= 0 {
		return fmt.Errorf("drop quantity must be positive")
	}
	if definition, exists := itemDefinition(itemID); !exists || definition.Virtual {
		return fmt.Errorf("unknown drop item %q", itemID)
	}
	return nil
}

func (w *World) addDrop(position Position, itemID string, quantity int, infinite bool) string {
	w.nextDropID++
	id := fmt.Sprintf("drop-%d", w.nextDropID)
	w.drops[id] = DropSnapshot{
		ID: id, Position: position, ItemID: itemID, Quantity: quantity, Infinite: infinite,
	}
	return id
}

func (w *World) QueueDropItem(command DropItemCommand) error {
	if command.Sequence == 0 {
		return fmt.Errorf("drop sequence must be positive")
	}
	if command.ItemID == "" {
		return fmt.Errorf("item ID is required")
	}
	if item, exists := itemDefinition(command.ItemID); !exists || item.Virtual {
		return fmt.Errorf("item %q is not droppable", command.ItemID)
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
	if pending, exists := w.combat.pendingEquipment[command.PlayerID]; exists && string(pending.ItemID) == command.ItemID {
		return fmt.Errorf("item %q has a pending equipment change", command.ItemID)
	}
	if !player.inventory.contains(command.ItemID) {
		return fmt.Errorf("player %q does not own item %q", command.PlayerID, command.ItemID)
	}
	if command.InstanceID != "" {
		item, exists := player.inventory.itemByInstance(command.InstanceID)
		if !exists || item.ItemID != command.ItemID {
			return fmt.Errorf("player %q does not own item instance %q", command.PlayerID, command.InstanceID)
		}
	}
	if (command.InstanceID != "" && command.InstanceID == player.equippedWeaponInstanceID) ||
		(command.InstanceID == "" && command.ItemID == string(player.equippedWeapon)) {
		return fmt.Errorf("item %q is currently equipped", command.ItemID)
	}
	for _, instanceID := range player.equippedArmourInstances {
		if command.InstanceID != "" && command.InstanceID == instanceID {
			return fmt.Errorf("item %q is currently equipped", command.ItemID)
		}
	}
	if command.InstanceID == "" {
		for _, equippedItemID := range player.equippedArmour {
			if command.ItemID == equippedItemID {
				return fmt.Errorf("item %q is currently equipped", command.ItemID)
			}
		}
	}
	if command.ItemID == player.equippedSpellbook {
		return fmt.Errorf("item %q is currently equipped", command.ItemID)
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
	if pending, exists := w.pendingPickups[command.PlayerID]; exists && command.Sequence <= pending.command.Sequence {
		return fmt.Errorf("pickup sequence %d is not newer than pending sequence %d", command.Sequence, pending.command.Sequence)
	}
	drop, exists := w.drops[command.DropID]
	if !exists {
		return fmt.Errorf("drop %q does not exist", command.DropID)
	}
	if !player.inventory.canAdd(drop.ItemID, drop.Quantity) {
		return fmt.Errorf("inventory is full")
	}
	path := []Position{drop.Position}
	if distance(player.position, drop.Position) > pickupRange && w.pathfinder != nil {
		var err error
		path, err = w.pathfinder.FindPath(player.position, drop.Position)
		if err != nil {
			return fmt.Errorf("find pickup path: %w", err)
		}
	}
	if len(path) == 0 {
		return fmt.Errorf("find pickup path: empty path")
	}
	delete(w.pendingMoves, command.PlayerID)
	delete(w.pendingGathering, command.PlayerID)
	clearGathering(player, "")
	w.pendingPickups[command.PlayerID] = pendingPickup{command: command, path: clonePositions(path)}
	return nil
}

func (w *World) startPendingPickupApproaches() {
	for id, pickup := range w.pendingPickups {
		if pickup.started {
			continue
		}
		player, exists := w.players[id]
		if !exists || player.defeated {
			delete(w.pendingPickups, id)
			continue
		}
		player.lastPickupSequence = pickup.command.Sequence
		player.autoApproaching = false
		player.autoPickingUp = true
		if drop, exists := w.drops[pickup.command.DropID]; exists && distance(player.position, drop.Position) > pickupRange {
			player.destination = positionPointer(drop.Position)
			player.path = clonePositions(pickup.path)
		} else {
			stopPlayer(player)
		}
		pickup.started = true
		w.pendingPickups[id] = pickup
	}
}

func (w *World) applyPendingPickups() {
	playerIDs := make([]PlayerID, 0, len(w.pendingPickups))
	for id := range w.pendingPickups {
		playerIDs = append(playerIDs, id)
	}
	sort.Slice(playerIDs, func(i, j int) bool { return playerIDs[i] < playerIDs[j] })
	for _, id := range playerIDs {
		pickup := w.pendingPickups[id]
		command := pickup.command
		player, exists := w.players[id]
		if !exists || player.defeated {
			delete(w.pendingPickups, id)
			continue
		}
		drop, exists := w.drops[command.DropID]
		if !exists || !player.inventory.canAdd(drop.ItemID, drop.Quantity) {
			delete(w.pendingPickups, id)
			if player.autoPickingUp {
				stopPlayer(player)
			}
			player.autoPickingUp = false
			continue
		}
		if distance(player.position, drop.Position) > pickupRange {
			continue
		}
		added := false
		if drop.InstanceID != "" {
			added = player.inventory.addInventoryItem(InventoryItem{ItemID: drop.ItemID, Quantity: 1, InstanceID: drop.InstanceID, Reinforcement: drop.Reinforcement})
		} else {
			added = player.inventory.add(drop.ItemID, drop.Quantity)
		}
		if added && !drop.Infinite {
			if respawn, exists := w.respawningDrops[command.DropID]; exists {
				delete(w.respawningDrops, command.DropID)
				respawn.Remaining = respawn.RespawnAfter
				w.pendingDropRespawns = append(w.pendingDropRespawns, respawn)
			}
			delete(w.drops, command.DropID)
			delete(w.dropLifetimes, command.DropID)
		}
		delete(w.pendingPickups, id)
		if player.autoPickingUp {
			stopPlayer(player)
		}
		player.autoPickingUp = false
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
		if (command.InstanceID != "" && command.InstanceID == player.equippedWeaponInstanceID) ||
			(command.InstanceID == "" && command.ItemID == string(player.equippedWeapon)) {
			continue
		}
		equipped := false
		for _, instanceID := range player.equippedArmourInstances {
			if command.InstanceID != "" && command.InstanceID == instanceID {
				equipped = true
				break
			}
		}
		if command.InstanceID == "" {
			for _, equippedItemID := range player.equippedArmour {
				if command.ItemID == equippedItemID {
					equipped = true
					break
				}
			}
		}
		if command.ItemID == player.equippedSpellbook {
			equipped = true
		}
		if !equipped && command.InstanceID != "" {
			if item, removed := player.inventory.removeInstance(command.InstanceID); removed {
				w.spawnInventoryItemDrop(player.position, item)
			}
			continue
		}
		quantity := player.inventory.quantity(command.ItemID)
		if !equipped && quantity > 0 && player.inventory.remove(command.ItemID, quantity) {
			w.spawnDrop(player.position, command.ItemID, quantity)
		}
	}
}

func (w *World) spawnInventoryItemDrop(position Position, item InventoryItem) {
	w.nextDropID++
	id := fmt.Sprintf("drop-%d", w.nextDropID)
	w.drops[id] = DropSnapshot{ID: id, Position: position, ItemID: item.ItemID, Quantity: 1, InstanceID: item.InstanceID, Reinforcement: item.Reinforcement}
	w.dropLifetimes[id] = worldDropLifetime
}

func (w *World) spawnDrop(position Position, itemID string, quantity int) {
	w.spawnDropWithLifetime(position, itemID, quantity, worldDropLifetime)
}

func (w *World) spawnPersistentDrop(position Position, itemID string, quantity int) {
	w.spawnDropWithLifetime(position, itemID, quantity, 0)
}

func (w *World) spawnDropWithLifetime(position Position, itemID string, quantity int, lifetime time.Duration) {
	if itemID == "" || quantity <= 0 {
		return
	}
	w.nextDropID++
	id := fmt.Sprintf("drop-%d", w.nextDropID)
	w.drops[id] = DropSnapshot{ID: id, Position: position, ItemID: itemID, Quantity: quantity}
	if lifetime > 0 {
		w.dropLifetimes[id] = lifetime
	}
}

func (w *World) advanceDropLifetimes(elapsed time.Duration) {
	for id, remaining := range w.dropLifetimes {
		remaining -= elapsed
		if remaining > 0 {
			w.dropLifetimes[id] = remaining
			continue
		}
		delete(w.dropLifetimes, id)
		delete(w.drops, id)
	}
}

func (w *World) advanceDropRespawns(elapsed time.Duration) {
	pending := w.pendingDropRespawns[:0]
	for _, respawn := range w.pendingDropRespawns {
		respawn.Remaining -= elapsed
		if respawn.Remaining > 0 {
			pending = append(pending, respawn)
			continue
		}
		id := w.addDrop(respawn.Position, respawn.ItemID, respawn.Quantity, false)
		respawn.Remaining = 0
		w.respawningDrops[id] = respawn
	}
	w.pendingDropRespawns = pending
}
