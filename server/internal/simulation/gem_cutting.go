package simulation

import "fmt"

var cutGemOutputs = map[string]string{
	ItemUncutSapphire: ItemSapphire,
	ItemUncutEmerald:  ItemEmerald,
	ItemUncutRuby:     ItemRuby,
	ItemUncutDiamond:  ItemDiamond,
}

// CutGem applies a small, server-authoritative crafting action. The chisel is
// required but is a reusable tool; one uncut gem becomes one cut gem.
func (w *World) CutGem(playerID PlayerID, sequence uint64, itemID string) error {
	if sequence == 0 {
		return fmt.Errorf("gem cutting sequence must be positive")
	}
	output, exists := cutGemOutputs[itemID]
	if !exists {
		return fmt.Errorf("item %q is not an uncut gem", itemID)
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	player, exists := w.players[playerID]
	if !exists {
		player, exists = w.pendingJoins[playerID]
	}
	if !exists || player.defeated {
		return fmt.Errorf("player %q cannot cut gems", playerID)
	}
	if sequence <= player.lastGemCutSequence {
		return fmt.Errorf("gem cutting sequence %d is not newer than %d", sequence, player.lastGemCutSequence)
	}
	if !player.inventory.contains(ItemChisel) || !player.inventory.contains(itemID) {
		return fmt.Errorf("a chisel and %s are required", itemID)
	}
	candidate := cloneInventory(player.inventory)
	if !candidate.remove(itemID, 1) || !candidate.add(output, 1) {
		return fmt.Errorf("inventory cannot hold the cut gem")
	}
	player.inventory = candidate
	player.lastGemCutSequence = sequence
	return nil
}
