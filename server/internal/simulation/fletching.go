package simulation

import (
	"fmt"
	"sort"
)

const (
	FletchArrowShafts            = "arrow_shafts"
	FletchBronzeArrows           = "bronze_arrows"
	FletchIronArrows             = "iron_arrows"
	FletchSteelArrows            = "steel_arrows"
	FletchMithrilArrows          = "mithril_arrows"
	FletchAdamantiteArrows       = "adamantite_arrows"
	FletchUnstrungShortbow       = "unstrung_shortbow"
	FletchShortbow               = "shortbow"
	FletchUnstrungOakShortbow    = "unstrung_oak_shortbow"
	FletchOakShortbow            = "oak_shortbow"
	FletchUnstrungWillowShortbow = "unstrung_willow_shortbow"
	FletchWillowShortbow         = "willow_shortbow"
	FletchUnstrungMapleShortbow  = "unstrung_maple_shortbow"
	FletchMapleShortbow          = "maple_shortbow"
	FletchUnstrungYewShortbow    = "unstrung_yew_shortbow"
	FletchYewShortbow            = "yew_shortbow"
)

type FletchCommand struct {
	PlayerID PlayerID
	Sequence uint64
	RecipeID string
	Quantity int
}

type fletchingRecipe struct {
	ID            string
	Name          string
	RequiredLevel int
	Costs         []ItemCost
	RequiredTool  string
	RewardItem    string
	RewardAmount  int
	Experience    int64
	Interval      uint64
}

type pendingFletch struct {
	command   FletchCommand
	remaining int
	nextPulse uint64
}

var fletchingRecipes = map[string]fletchingRecipe{
	FletchArrowShafts: {
		ID: FletchArrowShafts, Name: "20 Arrow Shafts", RequiredLevel: 1,
		Costs: []ItemCost{{ItemID: ItemNormalLogs, Quantity: 1}}, RequiredTool: ItemKnife,
		RewardItem: ItemArrowShafts, RewardAmount: 20, Experience: 10, Interval: 1,
	},
	FletchBronzeArrows: {
		ID: FletchBronzeArrows, Name: "20 Bronze Arrows", RequiredLevel: 1,
		Costs:      []ItemCost{{ItemID: ItemArrowShafts, Quantity: 20}, {ItemID: ItemFeathers, Quantity: 20}, {ItemID: ItemBronzeArrowTips, Quantity: 20}},
		RewardItem: ItemBronzeArrow, RewardAmount: 20, Experience: 15, Interval: 1,
	},
	FletchIronArrows: {
		ID: FletchIronArrows, Name: "20 Iron Arrows", RequiredLevel: 2,
		Costs: []ItemCost{
			{ItemID: ItemArrowShafts, Quantity: 20},
			{ItemID: ItemFeathers, Quantity: 20},
			{ItemID: ItemIronArrowTips, Quantity: 20},
		},
		RewardItem: ItemIronArrow, RewardAmount: 20, Experience: 22, Interval: 1,
	},
	FletchSteelArrows: {
		ID: FletchSteelArrows, Name: "20 Steel Arrows", RequiredLevel: 4,
		Costs:      []ItemCost{{ItemID: ItemArrowShafts, Quantity: 20}, {ItemID: ItemFeathers, Quantity: 20}, {ItemID: ItemSteelArrowTips, Quantity: 20}},
		RewardItem: ItemSteelArrow, RewardAmount: 20, Experience: 34, Interval: 1,
	},
	FletchMithrilArrows: {
		ID: FletchMithrilArrows, Name: "20 Mithril Arrows", RequiredLevel: 6,
		Costs:      []ItemCost{{ItemID: ItemArrowShafts, Quantity: 20}, {ItemID: ItemFeathers, Quantity: 20}, {ItemID: ItemMithrilArrowTips, Quantity: 20}},
		RewardItem: ItemMithrilArrow, RewardAmount: 20, Experience: 50, Interval: 1,
	},
	FletchAdamantiteArrows: {
		ID: FletchAdamantiteArrows, Name: "20 Adamantite Arrows", RequiredLevel: 8,
		Costs:      []ItemCost{{ItemID: ItemArrowShafts, Quantity: 20}, {ItemID: ItemFeathers, Quantity: 20}, {ItemID: ItemAdamantiteArrowTips, Quantity: 20}},
		RewardItem: ItemAdamantiteArrow, RewardAmount: 20, Experience: 70, Interval: 1,
	},
	FletchUnstrungShortbow: {
		ID: FletchUnstrungShortbow, Name: "Unstrung Shortbow", RequiredLevel: 1,
		Costs: []ItemCost{{ItemID: ItemNormalLogs, Quantity: 1}}, RequiredTool: ItemKnife,
		RewardItem: ItemUnstrungShortbow, RewardAmount: 1, Experience: 6, Interval: 2,
	},
	FletchShortbow: {
		ID: FletchShortbow, Name: "Shortbow", RequiredLevel: 1,
		Costs:      []ItemCost{{ItemID: ItemUnstrungShortbow, Quantity: 1}, {ItemID: ItemBowstring, Quantity: 1}},
		RewardItem: string(WeaponBow), RewardAmount: 1, Experience: 6, Interval: 1,
	},
	FletchUnstrungOakShortbow: {
		ID: FletchUnstrungOakShortbow, Name: "Unstrung Oak Shortbow", RequiredLevel: 2,
		Costs: []ItemCost{{ItemID: ItemOakLogs, Quantity: 1}}, RequiredTool: ItemKnife,
		RewardItem: ItemUnstrungOakShortbow, RewardAmount: 1, Experience: 10, Interval: 2,
	},
	FletchOakShortbow: {
		ID: FletchOakShortbow, Name: "Oak Shortbow", RequiredLevel: 2,
		Costs:      []ItemCost{{ItemID: ItemUnstrungOakShortbow, Quantity: 1}, {ItemID: ItemBowstring, Quantity: 1}},
		RewardItem: string(WeaponOakShortbow), RewardAmount: 1, Experience: 10, Interval: 1,
	},
	FletchUnstrungWillowShortbow: {
		ID: FletchUnstrungWillowShortbow, Name: "Unstrung Willow Shortbow", RequiredLevel: 3,
		Costs: []ItemCost{{ItemID: ItemWillowLogs, Quantity: 1}}, RequiredTool: ItemKnife,
		RewardItem: ItemUnstrungWillowShortbow, RewardAmount: 1, Experience: 15, Interval: 2,
	},
	FletchWillowShortbow: {
		ID: FletchWillowShortbow, Name: "Willow Shortbow", RequiredLevel: 3,
		Costs:      []ItemCost{{ItemID: ItemUnstrungWillowShortbow, Quantity: 1}, {ItemID: ItemBowstring, Quantity: 1}},
		RewardItem: string(WeaponWillowShortbow), RewardAmount: 1, Experience: 15, Interval: 1,
	},
	FletchUnstrungMapleShortbow: {
		ID: FletchUnstrungMapleShortbow, Name: "Unstrung Maple Shortbow", RequiredLevel: 5,
		Costs: []ItemCost{{ItemID: ItemMapleLogs, Quantity: 1}}, RequiredTool: ItemKnife,
		RewardItem: ItemUnstrungMapleShortbow, RewardAmount: 1, Experience: 24, Interval: 2,
	},
	FletchMapleShortbow: {
		ID: FletchMapleShortbow, Name: "Maple Shortbow", RequiredLevel: 5,
		Costs:      []ItemCost{{ItemID: ItemUnstrungMapleShortbow, Quantity: 1}, {ItemID: ItemBowstring, Quantity: 1}},
		RewardItem: string(WeaponMapleShortbow), RewardAmount: 1, Experience: 24, Interval: 1,
	},
	FletchUnstrungYewShortbow: {
		ID: FletchUnstrungYewShortbow, Name: "Unstrung Yew Shortbow", RequiredLevel: 8,
		Costs: []ItemCost{{ItemID: ItemYewLogs, Quantity: 1}}, RequiredTool: ItemKnife,
		RewardItem: ItemUnstrungYewShortbow, RewardAmount: 1, Experience: 38, Interval: 2,
	},
	FletchYewShortbow: {
		ID: FletchYewShortbow, Name: "Yew Shortbow", RequiredLevel: 8,
		Costs:      []ItemCost{{ItemID: ItemUnstrungYewShortbow, Quantity: 1}, {ItemID: ItemBowstring, Quantity: 1}},
		RewardItem: string(WeaponYewShortbow), RewardAmount: 1, Experience: 38, Interval: 1,
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
	if err := validateFletchingLevel(player, recipe); err != nil {
		return err
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
		command: command, remaining: remaining, nextPulse: w.combat.pulse + recipe.Interval,
	}
	player.lastFletchSequence = command.Sequence
	player.fletchRemaining = remaining
	player.nextFletchPulse = w.combat.pulse + recipe.Interval
	player.fletchingResult = "fletching"
	return nil
}

func validateFletchingLevel(player *playerState, recipe fletchingRecipe) error {
	if player == nil || SkillLevelForExperience(player.skills[SkillFletching]) < recipe.RequiredLevel {
		return fmt.Errorf("fletching level %d is required", recipe.RequiredLevel)
	}
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
		pending.nextPulse = w.combat.pulse + recipe.Interval
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
