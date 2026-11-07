package simulation

import (
	"errors"
	"fmt"
)

var (
	errMissingSmeltingInputs = errors.New("required smelting inputs are missing")
	errSmeltingInventoryFull = errors.New("inventory is full")
)

type SmeltingRecipeID string

const (
	SmeltingBronzeBar  SmeltingRecipeID = "bronze_bar"
	SmeltingIronBar    SmeltingRecipeID = "iron_bar"
	SmeltingMithrilBar SmeltingRecipeID = "mithril_bar"
)

type smeltingRecipeDefinition struct {
	ID             SmeltingRecipeID
	Name           string
	RequiredLevel  int
	Costs          []ItemCost
	RewardItem     string
	RewardQuantity int
	RewardXP       int64
	Interval       uint64
}

var smeltingRecipes = mustBuildSmeltingRecipes([]smeltingRecipeDefinition{
	{
		ID: SmeltingBronzeBar, Name: "Bronze Bar", RequiredLevel: 1,
		Costs:      []ItemCost{{ItemID: ItemCopperOre, Quantity: 1}, {ItemID: ItemTinOre, Quantity: 1}},
		RewardItem: ItemBronzeBar, RewardQuantity: 1, RewardXP: 14, Interval: 4,
	},
	{
		ID: SmeltingIronBar, Name: "Iron Bar", RequiredLevel: 2,
		Costs:      []ItemCost{{ItemID: ItemIronOre, Quantity: 1}, {ItemID: ItemCoal, Quantity: 1}},
		RewardItem: ItemIronBar, RewardQuantity: 1, RewardXP: 18, Interval: 5,
	},
	{
		ID: SmeltingMithrilBar, Name: "Mithril Bar", RequiredLevel: 5,
		Costs:      []ItemCost{{ItemID: ItemMithrilOre, Quantity: 1}, {ItemID: ItemCoal, Quantity: 2}},
		RewardItem: ItemMithrilBar, RewardQuantity: 1, RewardXP: 35, Interval: 6,
	},
})

func buildSmeltingRecipes(definitions []smeltingRecipeDefinition) (map[SmeltingRecipeID]smeltingRecipeDefinition, error) {
	recipes := make(map[SmeltingRecipeID]smeltingRecipeDefinition, len(definitions))
	for index, definition := range definitions {
		if definition.ID == "" || definition.Name == "" || definition.RequiredLevel < 1 ||
			definition.RewardQuantity <= 0 || definition.RewardXP <= 0 || definition.Interval == 0 || len(definition.Costs) == 0 {
			return nil, fmt.Errorf("smelting recipe %d is incomplete", index)
		}
		if _, exists := recipes[definition.ID]; exists {
			return nil, fmt.Errorf("smelting recipe %q is duplicated", definition.ID)
		}
		output, exists := itemDefinition(definition.RewardItem)
		if !exists || output.Kind != ItemKindResource {
			return nil, fmt.Errorf("smelting recipe %q has invalid output %q", definition.ID, definition.RewardItem)
		}
		for _, cost := range definition.Costs {
			input, inputExists := itemDefinition(cost.ItemID)
			if !inputExists || input.Kind != ItemKindResource || cost.Quantity <= 0 {
				return nil, fmt.Errorf("smelting recipe %q has invalid input %+v", definition.ID, cost)
			}
		}
		definition.Costs = append([]ItemCost(nil), definition.Costs...)
		recipes[definition.ID] = definition
	}
	return recipes, nil
}

func mustBuildSmeltingRecipes(definitions []smeltingRecipeDefinition) map[SmeltingRecipeID]smeltingRecipeDefinition {
	recipes, err := buildSmeltingRecipes(definitions)
	if err != nil {
		panic(err)
	}
	return recipes
}

func smeltingRecipe(id SmeltingRecipeID) (smeltingRecipeDefinition, bool) {
	recipe, exists := smeltingRecipes[id]
	return recipe, exists
}

func applySmeltingRecipe(player *playerState, recipe smeltingRecipeDefinition) error {
	if player == nil {
		return fmt.Errorf("player is required")
	}
	if SkillLevelForExperience(player.skills[SkillSmithing]) < recipe.RequiredLevel {
		return fmt.Errorf("smithing level %d is required", recipe.RequiredLevel)
	}
	candidate, err := smeltingInventoryAfter(player.inventory, recipe)
	if err != nil {
		return err
	}
	player.inventory = candidate
	grantSkillExperience(player, SkillSmithing, recipe.RewardXP)
	return nil
}

func smeltingInventoryAfter(inventory inventoryState, recipe smeltingRecipeDefinition) (inventoryState, error) {
	candidate := cloneInventory(inventory)
	if !payItemCosts(&candidate, recipe.Costs) {
		return inventoryState{}, errMissingSmeltingInputs
	}
	if !candidate.add(recipe.RewardItem, recipe.RewardQuantity) {
		return inventoryState{}, errSmeltingInventoryFull
	}
	return candidate, nil
}
