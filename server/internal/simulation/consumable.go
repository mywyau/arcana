package simulation

import (
	"fmt"
	"sort"
)

type ConsumeItemCommand struct {
	PlayerID PlayerID
	Sequence uint64
	ItemID   string
}

func (w *World) QueueConsumeItem(command ConsumeItemCommand) error {
	if command.Sequence == 0 {
		return fmt.Errorf("consume sequence must be positive")
	}
	item, exists := itemDefinition(command.ItemID)
	if !exists || item.Kind != ItemKindConsumable {
		return fmt.Errorf("item %q is not consumable", command.ItemID)
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
	if command.Sequence <= player.lastConsumeSequence {
		return fmt.Errorf("consume sequence %d is not newer than %d", command.Sequence, player.lastConsumeSequence)
	}
	if pending, exists := w.pendingConsumables[command.PlayerID]; exists && command.Sequence <= pending.Sequence {
		return fmt.Errorf("consume sequence %d is not newer than pending sequence %d", command.Sequence, pending.Sequence)
	}
	if !player.inventory.contains(command.ItemID) {
		return fmt.Errorf("player %q does not own item %q", command.PlayerID, command.ItemID)
	}
	if player.health >= player.maximumHealth {
		return fmt.Errorf("player %q is already at full health", command.PlayerID)
	}
	if w.combat.pulse < player.nextConsumePulse {
		return fmt.Errorf("consumable cooldown lasts until pulse %d", player.nextConsumePulse)
	}
	w.pendingConsumables[command.PlayerID] = command
	return nil
}

func (w *World) applyPendingConsumables() {
	playerIDs := make([]PlayerID, 0, len(w.pendingConsumables))
	for id := range w.pendingConsumables {
		playerIDs = append(playerIDs, id)
	}
	sort.Slice(playerIDs, func(i, j int) bool { return playerIDs[i] < playerIDs[j] })
	for _, id := range playerIDs {
		command := w.pendingConsumables[id]
		delete(w.pendingConsumables, id)
		player, exists := w.players[id]
		if !exists || player.defeated || player.health >= player.maximumHealth ||
			w.combat.pulse < player.nextConsumePulse {
			continue
		}
		item, exists := itemDefinition(command.ItemID)
		if !exists || item.Kind != ItemKindConsumable || !player.inventory.remove(command.ItemID, 1) {
			continue
		}
		player.health = min(player.maximumHealth, player.health+item.Consumable.Healing)
		player.lastConsumeSequence = command.Sequence
		player.nextConsumePulse = w.combat.pulse + item.Consumable.CooldownPulses
	}
}
