package simulation

import (
	"errors"
	"fmt"
)

var (
	errMissingFoodInputs = errors.New("required cooking inputs are missing")
	errFoodInventoryFull = errors.New("inventory is full")
)

type foodRecipeDefinition struct {
	ID             string
	Name           string
	RequiredLevel  int
	Costs          []ItemCost
	RewardItem     string
	RewardQuantity int
	RewardXP       int64
	Interval       uint64
}

var fishingRecipes = mustBuildFoodRecipes(SkillFishing, []foodRecipeDefinition{
	{ID: ItemRawMinnow, Name: "Minnow", RequiredLevel: 1, RewardItem: ItemRawMinnow, RewardQuantity: 1, RewardXP: 10, Interval: 4},
	{ID: ItemRawTrout, Name: "Trout", RequiredLevel: 2, RewardItem: ItemRawTrout, RewardQuantity: 1, RewardXP: 16, Interval: 4},
	{ID: ItemRawSalmon, Name: "Salmon", RequiredLevel: 3, RewardItem: ItemRawSalmon, RewardQuantity: 1, RewardXP: 24, Interval: 5},
	{ID: ItemRawTuna, Name: "Tuna", RequiredLevel: 4, RewardItem: ItemRawTuna, RewardQuantity: 1, RewardXP: 34, Interval: 6},
	{ID: ItemRawLobster, Name: "Lobster", RequiredLevel: 5, RewardItem: ItemRawLobster, RewardQuantity: 1, RewardXP: 46, Interval: 7},
})

var cookingRecipes = mustBuildFoodRecipes(SkillCooking, []foodRecipeDefinition{
	{ID: ItemCookedMinnow, Name: "Minnow", RequiredLevel: 1, Costs: []ItemCost{{ItemID: ItemRawMinnow, Quantity: 1}}, RewardItem: ItemCookedMinnow, RewardQuantity: 1, RewardXP: 10, Interval: 4},
	{ID: ItemCookedTrout, Name: "Trout", RequiredLevel: 2, Costs: []ItemCost{{ItemID: ItemRawTrout, Quantity: 1}}, RewardItem: ItemCookedTrout, RewardQuantity: 1, RewardXP: 16, Interval: 4},
	{ID: ItemCookedSalmon, Name: "Salmon", RequiredLevel: 3, Costs: []ItemCost{{ItemID: ItemRawSalmon, Quantity: 1}}, RewardItem: ItemCookedSalmon, RewardQuantity: 1, RewardXP: 24, Interval: 5},
	{ID: ItemCookedTuna, Name: "Tuna", RequiredLevel: 4, Costs: []ItemCost{{ItemID: ItemRawTuna, Quantity: 1}}, RewardItem: ItemCookedTuna, RewardQuantity: 1, RewardXP: 34, Interval: 6},
	{ID: ItemCookedLobster, Name: "Lobster", RequiredLevel: 5, Costs: []ItemCost{{ItemID: ItemRawLobster, Quantity: 1}}, RewardItem: ItemCookedLobster, RewardQuantity: 1, RewardXP: 46, Interval: 7},
})

func buildFoodRecipes(skill SkillID, definitions []foodRecipeDefinition) (map[string]foodRecipeDefinition, error) {
	recipes := make(map[string]foodRecipeDefinition, len(definitions))
	for index, definition := range definitions {
		if definition.ID == "" || definition.Name == "" || definition.RequiredLevel < 1 ||
			definition.RewardQuantity <= 0 || definition.RewardXP <= 0 || definition.Interval == 0 {
			return nil, fmt.Errorf("%s recipe %d is incomplete", skill, index)
		}
		if _, exists := recipes[definition.ID]; exists {
			return nil, fmt.Errorf("%s recipe %q is duplicated", skill, definition.ID)
		}
		output, exists := itemDefinition(definition.RewardItem)
		if !exists || (skill == SkillFishing && output.Kind != ItemKindResource) ||
			(skill == SkillCooking && output.Kind != ItemKindConsumable) {
			return nil, fmt.Errorf("%s recipe %q has invalid output %q", skill, definition.ID, definition.RewardItem)
		}
		if skill == SkillFishing && len(definition.Costs) != 0 {
			return nil, fmt.Errorf("fishing recipe %q cannot require items", definition.ID)
		}
		if skill == SkillCooking && len(definition.Costs) == 0 {
			return nil, fmt.Errorf("cooking recipe %q requires an input", definition.ID)
		}
		for _, cost := range definition.Costs {
			input, inputExists := itemDefinition(cost.ItemID)
			if !inputExists || input.Kind != ItemKindResource || cost.Quantity <= 0 {
				return nil, fmt.Errorf("%s recipe %q has invalid input %+v", skill, definition.ID, cost)
			}
		}
		definition.Costs = append([]ItemCost(nil), definition.Costs...)
		recipes[definition.ID] = definition
	}
	return recipes, nil
}

func mustBuildFoodRecipes(skill SkillID, definitions []foodRecipeDefinition) map[string]foodRecipeDefinition {
	recipes, err := buildFoodRecipes(skill, definitions)
	if err != nil {
		panic(err)
	}
	return recipes
}

func foodRecipe(station craftingStationID, id string) (foodRecipeDefinition, bool) {
	if id == "" {
		if station == craftingStationFishing {
			id = ItemRawMinnow
		} else if station == craftingStationCooking {
			id = ItemCookedMinnow
		}
	}
	if station == craftingStationFishing {
		recipe, exists := fishingRecipes[id]
		return recipe, exists
	}
	if station == craftingStationCooking {
		recipe, exists := cookingRecipes[id]
		return recipe, exists
	}
	return foodRecipeDefinition{}, false
}

func foodInventoryAfter(inventory inventoryState, recipe foodRecipeDefinition) (inventoryState, error) {
	candidate := cloneInventory(inventory)
	if !payItemCosts(&candidate, recipe.Costs) {
		return inventoryState{}, errMissingFoodInputs
	}
	if !candidate.add(recipe.RewardItem, recipe.RewardQuantity) {
		return inventoryState{}, errFoodInventoryFull
	}
	return candidate, nil
}

func applyFoodRecipe(player *playerState, skill SkillID, recipe foodRecipeDefinition) error {
	if player == nil {
		return fmt.Errorf("player is required")
	}
	if SkillLevelForExperience(player.skills[skill]) < recipe.RequiredLevel {
		return fmt.Errorf("%s level %d is required", skill, recipe.RequiredLevel)
	}
	candidate, err := foodInventoryAfter(player.inventory, recipe)
	if err != nil {
		return err
	}
	player.inventory = candidate
	grantSkillExperience(player, skill, recipe.RewardXP)
	return nil
}
