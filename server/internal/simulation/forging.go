package simulation

import (
	"errors"
	"fmt"
)

var (
	errMissingForgingInputs = errors.New("required forging inputs are missing")
	errForgingInventoryFull = errors.New("inventory is full")
)

type ForgingRecipeID string

const (
	ForgingBronzeSword   ForgingRecipeID = "bronze_sword"
	ForgingBronzeHelm    ForgingRecipeID = "bronze_helm"
	ForgingBronzePickaxe ForgingRecipeID = "bronze_pickaxe"
)

type forgingRecipeDefinition struct {
	ID             ForgingRecipeID
	Name           string
	RequiredLevel  int
	Costs          []ItemCost
	RewardItem     string
	RewardQuantity int
	RewardXP       int64
	Interval       uint64
}

var forgingRecipes = mustBuildForgingRecipes([]forgingRecipeDefinition{
	{
		ID: ForgingBronzeSword, Name: "Bronze Sword", RequiredLevel: 1,
		Costs:      []ItemCost{{ItemID: ItemBronzeBar, Quantity: 2}},
		RewardItem: string(WeaponBronzeSword), RewardQuantity: 1, RewardXP: 20, Interval: 4,
	},
	{
		ID: ForgingBronzeHelm, Name: "Bronze Helm", RequiredLevel: 1,
		Costs:      []ItemCost{{ItemID: ItemBronzeBar, Quantity: 2}},
		RewardItem: ItemBronzeHelm, RewardQuantity: 1, RewardXP: 20, Interval: 4,
	},
	{
		ID: ForgingBronzePickaxe, Name: "Bronze Pickaxe", RequiredLevel: 1,
		Costs:      []ItemCost{{ItemID: ItemBronzeBar, Quantity: 3}},
		RewardItem: ItemBronzePickaxe, RewardQuantity: 1, RewardXP: 30, Interval: 5,
	},
})

func buildForgingRecipes(definitions []forgingRecipeDefinition) (map[ForgingRecipeID]forgingRecipeDefinition, error) {
	recipes := make(map[ForgingRecipeID]forgingRecipeDefinition, len(definitions))
	for index, definition := range definitions {
		if definition.ID == "" || definition.Name == "" || definition.RequiredLevel < 1 ||
			definition.RewardQuantity <= 0 || definition.RewardXP <= 0 || definition.Interval == 0 || len(definition.Costs) == 0 {
			return nil, fmt.Errorf("forging recipe %d is incomplete", index)
		}
		if _, exists := recipes[definition.ID]; exists {
			return nil, fmt.Errorf("forging recipe %q is duplicated", definition.ID)
		}
		if output, exists := itemDefinition(definition.RewardItem); !exists || output.Virtual {
			return nil, fmt.Errorf("forging recipe %q has invalid output %q", definition.ID, definition.RewardItem)
		}
		for _, cost := range definition.Costs {
			input, exists := itemDefinition(cost.ItemID)
			if !exists || input.Kind != ItemKindResource || cost.Quantity <= 0 {
				return nil, fmt.Errorf("forging recipe %q has invalid input %+v", definition.ID, cost)
			}
		}
		definition.Costs = append([]ItemCost(nil), definition.Costs...)
		recipes[definition.ID] = definition
	}
	return recipes, nil
}

func mustBuildForgingRecipes(definitions []forgingRecipeDefinition) map[ForgingRecipeID]forgingRecipeDefinition {
	recipes, err := buildForgingRecipes(definitions)
	if err != nil {
		panic(err)
	}
	return recipes
}

func forgingRecipe(id ForgingRecipeID) (forgingRecipeDefinition, bool) {
	recipe, exists := forgingRecipes[id]
	return recipe, exists
}

func applyForgingRecipe(player *playerState, recipe forgingRecipeDefinition) error {
	if player == nil {
		return fmt.Errorf("player is required")
	}
	if SkillLevelForExperience(player.skills[SkillSmithing]) < recipe.RequiredLevel {
		return fmt.Errorf("smithing level %d is required", recipe.RequiredLevel)
	}
	candidate, err := forgingInventoryAfter(player.inventory, recipe)
	if err != nil {
		return err
	}
	player.inventory = candidate
	grantSkillExperience(player, SkillSmithing, recipe.RewardXP)
	return nil
}

func forgingInventoryAfter(inventory inventoryState, recipe forgingRecipeDefinition) (inventoryState, error) {
	candidate := cloneInventory(inventory)
	if !payItemCosts(&candidate, recipe.Costs) {
		return inventoryState{}, errMissingForgingInputs
	}
	if !candidate.add(recipe.RewardItem, recipe.RewardQuantity) {
		return inventoryState{}, errForgingInventoryFull
	}
	return candidate, nil
}
