package simulation

import (
	"fmt"
	"sort"
	"strings"
)

type BrewCommand struct {
	PlayerID PlayerID
	Sequence uint64
	ItemIDs  []string
}

type brewingRecipe struct {
	OutputItem string
	Experience int64
}

var brewingIngredientRoles = map[string]string{
	ItemVialOfWater: "base",
	ItemSunleaf:     "herb",
	ItemMoonleaf:    "herb",
	ItemSacredDust:  "secondary",
	ItemFishOil:     "secondary",
}

var brewingRecipes = map[string]brewingRecipe{
	brewKey([]string{ItemVialOfWater, ItemSunleaf}):                 {OutputItem: ItemMinorHealingPotion, Experience: 10},
	brewKey([]string{ItemVialOfWater, ItemSunleaf, ItemSacredDust}): {OutputItem: ItemPrayerPotion, Experience: 20},
}

func (w *World) QueueBrew(command BrewCommand) error {
	if command.Sequence == 0 || len(command.ItemIDs) < 2 || len(command.ItemIDs) > 3 {
		return fmt.Errorf("brewing requires a positive sequence and two or three items")
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	player, exists := w.players[command.PlayerID]
	if !exists || player.defeated {
		return fmt.Errorf("player %q cannot brew", command.PlayerID)
	}
	if command.Sequence <= player.lastBrewSequence {
		return fmt.Errorf("brew sequence %d is not newer than %d", command.Sequence, player.lastBrewSequence)
	}
	if pending, exists := w.pendingBrews[command.PlayerID]; exists && command.Sequence <= pending.Sequence {
		return fmt.Errorf("brew sequence %d is not newer than pending sequence %d", command.Sequence, pending.Sequence)
	}
	if _, _, err := brewedInventoryAfter(player.inventory, command.ItemIDs); err != nil {
		return err
	}
	command.ItemIDs = append([]string(nil), command.ItemIDs...)
	w.pendingBrews[command.PlayerID] = command
	return nil
}

func (w *World) applyPendingBrews() {
	playerIDs := make([]PlayerID, 0, len(w.pendingBrews))
	for id := range w.pendingBrews {
		playerIDs = append(playerIDs, id)
	}
	sort.Slice(playerIDs, func(i, j int) bool { return playerIDs[i] < playerIDs[j] })
	for _, id := range playerIDs {
		command := w.pendingBrews[id]
		delete(w.pendingBrews, id)
		player, exists := w.players[id]
		if !exists || player.defeated {
			continue
		}
		inventory, recipe, err := brewedInventoryAfter(player.inventory, command.ItemIDs)
		if err != nil {
			continue
		}
		player.inventory = inventory
		player.lastBrewSequence = command.Sequence
		grantSkillExperience(player, SkillHerblore, recipe.Experience)
	}
}

func brewedInventoryAfter(inventory inventoryState, itemIDs []string) (inventoryState, brewingRecipe, error) {
	if len(itemIDs) < 2 || len(itemIDs) > 3 {
		return inventoryState{}, brewingRecipe{}, fmt.Errorf("brewing requires two or three items")
	}
	roleCounts := map[string]int{}
	itemCounts := map[string]int{}
	for _, itemID := range itemIDs {
		role, allowed := brewingIngredientRoles[itemID]
		if !allowed {
			return inventoryState{}, brewingRecipe{}, fmt.Errorf("item %q is not a brewing ingredient", itemID)
		}
		roleCounts[role]++
		itemCounts[itemID]++
	}
	if roleCounts["base"] != 1 || roleCounts["herb"] < 1 {
		return inventoryState{}, brewingRecipe{}, fmt.Errorf("a brew requires one water vial and at least one herb")
	}
	candidate := cloneInventory(inventory)
	for itemID, quantity := range itemCounts {
		if !candidate.remove(itemID, quantity) {
			return inventoryState{}, brewingRecipe{}, fmt.Errorf("required brewing ingredients are missing")
		}
	}
	recipe, known := brewingRecipes[brewKey(itemIDs)]
	if !known {
		recipe = brewingRecipe{OutputItem: ItemMurkyBrew, Experience: 5}
	}
	if !candidate.add(recipe.OutputItem, 1) {
		return inventoryState{}, brewingRecipe{}, fmt.Errorf("inventory is full")
	}
	return candidate, recipe, nil
}

func brewKey(itemIDs []string) string {
	keyItems := append([]string(nil), itemIDs...)
	sort.Strings(keyItems)
	return strings.Join(keyItems, "+")
}
