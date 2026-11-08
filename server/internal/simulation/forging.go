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
	ForgingBronzePlate   ForgingRecipeID = "bronze_platebody"
	ForgingBronzePickaxe ForgingRecipeID = "bronze_pickaxe"
	ForgingBronzeShield  ForgingRecipeID = "bronze_shield"
	ForgingBronzeLegs    ForgingRecipeID = "bronze_legs"
	ForgingBronzeGloves  ForgingRecipeID = "bronze_gloves"
	ForgingBronzeBoots   ForgingRecipeID = "bronze_boots"
	ForgingIronSword     ForgingRecipeID = "iron_sword"
	ForgingIronArrowTips ForgingRecipeID = "iron_arrow_tips"
	ForgingIronHelm      ForgingRecipeID = "iron_helm"
	ForgingIronPlate     ForgingRecipeID = "iron_platebody"
	ForgingIronShield    ForgingRecipeID = "iron_shield"
	ForgingIronLegs      ForgingRecipeID = "iron_legs"
	ForgingIronGloves    ForgingRecipeID = "iron_gloves"
	ForgingIronBoots     ForgingRecipeID = "iron_boots"
	ForgingMithrilSword  ForgingRecipeID = "mithril_sword"
	ForgingMithrilHelm   ForgingRecipeID = "mithril_helm"
	ForgingMithrilPlate  ForgingRecipeID = "mithril_platebody"
	ForgingMithrilShield ForgingRecipeID = "mithril_shield"
	ForgingMithrilLegs   ForgingRecipeID = "mithril_legs"
	ForgingMithrilGloves ForgingRecipeID = "mithril_gloves"
	ForgingMithrilBoots  ForgingRecipeID = "mithril_boots"
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
		ID: ForgingBronzePlate, Name: "Bronze Platebody", RequiredLevel: 2,
		Costs:      []ItemCost{{ItemID: ItemBronzeBar, Quantity: 5}},
		RewardItem: ItemBronzePlate, RewardQuantity: 1, RewardXP: 50, Interval: 6,
	},
	{
		ID: ForgingBronzePickaxe, Name: "Bronze Pickaxe", RequiredLevel: 1,
		Costs:      []ItemCost{{ItemID: ItemBronzeBar, Quantity: 3}},
		RewardItem: ItemBronzePickaxe, RewardQuantity: 1, RewardXP: 30, Interval: 5,
	},
	{ID: ForgingBronzeShield, Name: "Bronze Shield", RequiredLevel: 2, Costs: []ItemCost{{ItemID: ItemBronzeBar, Quantity: 3}}, RewardItem: ItemBronzeShield, RewardQuantity: 1, RewardXP: 30, Interval: 5},
	{ID: ForgingBronzeLegs, Name: "Bronze Legs", RequiredLevel: 2, Costs: []ItemCost{{ItemID: ItemBronzeBar, Quantity: 4}}, RewardItem: ItemBronzeLegs, RewardQuantity: 1, RewardXP: 40, Interval: 5},
	{ID: ForgingBronzeGloves, Name: "Bronze Gloves", RequiredLevel: 1, Costs: []ItemCost{{ItemID: ItemBronzeBar, Quantity: 1}}, RewardItem: ItemBronzeGloves, RewardQuantity: 1, RewardXP: 10, Interval: 3},
	{ID: ForgingBronzeBoots, Name: "Bronze Boots", RequiredLevel: 1, Costs: []ItemCost{{ItemID: ItemBronzeBar, Quantity: 2}}, RewardItem: ItemBronzeBoots, RewardQuantity: 1, RewardXP: 20, Interval: 4},
	{
		ID: ForgingIronArrowTips, Name: "20 Iron Arrow Tips", RequiredLevel: 2,
		Costs:      []ItemCost{{ItemID: ItemIronBar, Quantity: 1}},
		RewardItem: ItemIronArrowTips, RewardQuantity: 20, RewardXP: 18, Interval: 3,
	},
	{
		ID: ForgingIronSword, Name: "Iron Sword", RequiredLevel: 3,
		Costs:      []ItemCost{{ItemID: ItemIronBar, Quantity: 2}},
		RewardItem: string(WeaponSword), RewardQuantity: 1, RewardXP: 36, Interval: 5,
	},
	{
		ID: ForgingIronHelm, Name: "Iron Helm", RequiredLevel: 3,
		Costs:      []ItemCost{{ItemID: ItemIronBar, Quantity: 2}},
		RewardItem: ItemIronHelm, RewardQuantity: 1, RewardXP: 36, Interval: 5,
	},
	{
		ID: ForgingIronPlate, Name: "Iron Platebody", RequiredLevel: 4,
		Costs:      []ItemCost{{ItemID: ItemIronBar, Quantity: 5}},
		RewardItem: ItemIronPlate, RewardQuantity: 1, RewardXP: 90, Interval: 7,
	},
	{ID: ForgingIronShield, Name: "Iron Shield", RequiredLevel: 4, Costs: []ItemCost{{ItemID: ItemIronBar, Quantity: 3}}, RewardItem: ItemIronShield, RewardQuantity: 1, RewardXP: 54, Interval: 5},
	{ID: ForgingIronLegs, Name: "Iron Legs", RequiredLevel: 4, Costs: []ItemCost{{ItemID: ItemIronBar, Quantity: 4}}, RewardItem: ItemIronLegs, RewardQuantity: 1, RewardXP: 72, Interval: 6},
	{ID: ForgingIronGloves, Name: "Iron Gloves", RequiredLevel: 3, Costs: []ItemCost{{ItemID: ItemIronBar, Quantity: 1}}, RewardItem: ItemIronGloves, RewardQuantity: 1, RewardXP: 18, Interval: 3},
	{ID: ForgingIronBoots, Name: "Iron Boots", RequiredLevel: 3, Costs: []ItemCost{{ItemID: ItemIronBar, Quantity: 2}}, RewardItem: ItemIronBoots, RewardQuantity: 1, RewardXP: 36, Interval: 4},
	{
		ID: ForgingMithrilSword, Name: "Mithril Sword", RequiredLevel: 6,
		Costs:      []ItemCost{{ItemID: ItemMithrilBar, Quantity: 3}},
		RewardItem: string(WeaponMithrilSword), RewardQuantity: 1, RewardXP: 90, Interval: 6,
	},
	{
		ID: ForgingMithrilHelm, Name: "Mithril Helm", RequiredLevel: 6,
		Costs:      []ItemCost{{ItemID: ItemMithrilBar, Quantity: 3}},
		RewardItem: ItemMithrilHelm, RewardQuantity: 1, RewardXP: 90, Interval: 6,
	},
	{
		ID: ForgingMithrilPlate, Name: "Mithril Platebody", RequiredLevel: 7,
		Costs:      []ItemCost{{ItemID: ItemMithrilBar, Quantity: 6}},
		RewardItem: ItemMithrilPlate, RewardQuantity: 1, RewardXP: 180, Interval: 8,
	},
	{ID: ForgingMithrilShield, Name: "Mithril Shield", RequiredLevel: 7, Costs: []ItemCost{{ItemID: ItemMithrilBar, Quantity: 4}}, RewardItem: ItemMithrilShield, RewardQuantity: 1, RewardXP: 120, Interval: 6},
	{ID: ForgingMithrilLegs, Name: "Mithril Legs", RequiredLevel: 7, Costs: []ItemCost{{ItemID: ItemMithrilBar, Quantity: 5}}, RewardItem: ItemMithrilLegs, RewardQuantity: 1, RewardXP: 150, Interval: 7},
	{ID: ForgingMithrilGloves, Name: "Mithril Gloves", RequiredLevel: 6, Costs: []ItemCost{{ItemID: ItemMithrilBar, Quantity: 2}}, RewardItem: ItemMithrilGloves, RewardQuantity: 1, RewardXP: 60, Interval: 4},
	{ID: ForgingMithrilBoots, Name: "Mithril Boots", RequiredLevel: 6, Costs: []ItemCost{{ItemID: ItemMithrilBar, Quantity: 3}}, RewardItem: ItemMithrilBoots, RewardQuantity: 1, RewardXP: 90, Interval: 5},
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
