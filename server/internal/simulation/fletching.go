package simulation

import (
	"fmt"
	"sort"
)

const (
	FletchArrowShafts = "arrow_shafts"
	FletchIronArrows  = "iron_arrows"
)

type FletchCommand struct {
	PlayerID PlayerID
	Sequence uint64
	RecipeID string
	Quantity int
}

type fletchingRecipe struct {
	ID           string
	Costs        []ItemCost
	RequiredTool string
	RewardItem   string
	RewardAmount int
	Experience   int64
}

type pendingFletch struct {
	command   FletchCommand
	remaining int
	nextPulse uint64
}

var fletchingRecipes = map[string]fletchingRecipe{
	FletchArrowShafts: {
		ID: FletchArrowShafts, Costs: []ItemCost{{ItemID: ItemNormalLogs, Quantity: 1}},
		RequiredTool: ItemKnife, RewardItem: ItemArrowShafts, RewardAmount: 20, Experience: 10,
	},
	FletchIronArrows: {
		ID: FletchIronArrows,
		Costs: []ItemCost{
			{ItemID: ItemArrowShafts, Quantity: 20},
			{ItemID: ItemFeathers, Quantity: 20},
			{ItemID: ItemIronArrowTips, Quantity: 20},
		},
		RewardItem: ItemIronArrow, RewardAmount: 20, Experience: 20,
	},
}

func (w *World) QueueFletch(command FletchCommand) error {
	if command.Sequence == 0 {
		return fmt.Errorf("fletching sequence must be positive")
	}
	recipe, exists := fletchingRecipes[command.RecipeID]
	if !exists {
		return fmt.Errorf("fletching recipe %q does not exist", command.RecipeID)
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	player, exists := w.players[command.PlayerID]
	if !exists || player.defeated {
		return fmt.Errorf("player %q cannot fletch", command.PlayerID)
	}
	if command.Sequence <= player.lastFletchSequence || player.fletchRemaining > 0 {
		return fmt.Errorf("fletching command is not available")
	}
	if _, err := fletchedInventoryAfter(player.inventory, recipe); err != nil {
		return err
	}
	maximum := maximumRecipeCount(player.inventory, recipe.Costs)
	remaining := command.Quantity
	if remaining == 0 {
		remaining = maximum
	}
	if remaining < 1 || remaining > maximum {
		return fmt.Errorf("cannot fletch %d batch(es); ingredients allow %d", remaining, maximum)
	}
	w.pendingFletches[command.PlayerID] = pendingFletch{
		command: command, remaining: remaining, nextPulse: w.combat.pulse + 1,
	}
	player.lastFletchSequence = command.Sequence
	player.fletchRemaining = remaining
	player.nextFletchPulse = w.combat.pulse + 1
	player.fletchingResult = "fletching"
	return nil
}

func (w *World) applyPendingFletches() {
	playerIDs := make([]PlayerID, 0, len(w.pendingFletches))
	for id := range w.pendingFletches {
		playerIDs = append(playerIDs, id)
	}
	sort.Slice(playerIDs, func(i, j int) bool { return playerIDs[i] < playerIDs[j] })
	for _, id := range playerIDs {
		pending := w.pendingFletches[id]
		player, exists := w.players[id]
		if !exists || player.defeated {
			delete(w.pendingFletches, id)
			continue
		}
		if w.combat.pulse < pending.nextPulse {
			continue
		}
		recipe := fletchingRecipes[pending.command.RecipeID]
		inventory, err := fletchedInventoryAfter(player.inventory, recipe)
		if err != nil {
			player.fletchRemaining = 0
			player.nextFletchPulse = 0
			player.fletchingResult = "fletching_stopped"
			delete(w.pendingFletches, id)
			continue
		}
		player.inventory = inventory
		grantSkillExperience(player, SkillFletching, recipe.Experience)
		player.fletchingResult = recipe.RewardItem
		pending.remaining--
		player.fletchRemaining = pending.remaining
		if pending.remaining <= 0 {
			player.nextFletchPulse = 0
			delete(w.pendingFletches, id)
			continue
		}
		pending.nextPulse = w.combat.pulse + 1
		player.nextFletchPulse = pending.nextPulse
		w.pendingFletches[id] = pending
	}
}

func fletchedInventoryAfter(inventory inventoryState, recipe fletchingRecipe) (inventoryState, error) {
	candidate := cloneInventory(inventory)
	if recipe.RequiredTool != "" && !candidate.contains(recipe.RequiredTool) {
		return inventoryState{}, fmt.Errorf("required tool %q is missing", recipe.RequiredTool)
	}
	if !payItemCosts(&candidate, recipe.Costs) {
		return inventoryState{}, fmt.Errorf("required fletching materials are missing")
	}
	if !candidate.add(recipe.RewardItem, recipe.RewardAmount) {
		return inventoryState{}, fmt.Errorf("inventory is full")
	}
	return candidate, nil
}

func maximumRecipeCount(inventory inventoryState, costs []ItemCost) int {
	maximum := int(^uint(0) >> 1)
	for _, cost := range costs {
		maximum = min(maximum, inventory.quantity(cost.ItemID)/cost.Quantity)
	}
	if maximum == int(^uint(0)>>1) {
		return 0
	}
	return maximum
}
