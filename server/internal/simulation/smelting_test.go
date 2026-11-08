package simulation

import (
	"strings"
	"testing"
)

func TestSmeltingRecipesUseExpectedInputsAndProgression(t *testing.T) {
	tests := []struct {
		id     SmeltingRecipeID
		level  int
		output string
		costs  []ItemCost
	}{
		{SmeltingBronzeBar, 1, ItemBronzeBar, []ItemCost{{ItemID: ItemCopperOre, Quantity: 1}, {ItemID: ItemTinOre, Quantity: 1}}},
		{SmeltingIronBar, 2, ItemIronBar, []ItemCost{{ItemID: ItemIronOre, Quantity: 1}, {ItemID: ItemCoal, Quantity: 1}}},
		{SmeltingMithrilBar, 5, ItemMithrilBar, []ItemCost{{ItemID: ItemMithrilOre, Quantity: 1}, {ItemID: ItemCoal, Quantity: 2}}},
		{SmeltingGoldBar, 4, ItemGoldBar, []ItemCost{{ItemID: ItemGoldOre, Quantity: 1}}},
	}
	for _, test := range tests {
		recipe, exists := smeltingRecipe(test.id)
		if !exists || recipe.RequiredLevel != test.level || recipe.RewardItem != test.output || len(recipe.Costs) != len(test.costs) {
			t.Errorf("smeltingRecipe(%q) = %+v, %t", test.id, recipe, exists)
			continue
		}
		for index := range test.costs {
			if recipe.Costs[index] != test.costs[index] {
				t.Errorf("smeltingRecipe(%q) cost %d = %+v, want %+v", test.id, index, recipe.Costs[index], test.costs[index])
			}
		}
	}
}

func TestApplySmeltingRecipeAtomicallyConsumesInputsAndGrantsXP(t *testing.T) {
	player := &playerState{inventory: newEmptyInventory(), skills: newDefaultSkills()}
	player.inventory.add(ItemIronOre, 1)
	player.inventory.add(ItemCoal, 1)
	player.skills[SkillSmithing] = SkillExperienceForLevel(2)
	recipe, _ := smeltingRecipe(SmeltingIronBar)
	if err := applySmeltingRecipe(player, recipe); err != nil {
		t.Fatalf("applySmeltingRecipe() returned an unexpected error: %v", err)
	}
	if player.inventory.quantity(ItemIronOre) != 0 || player.inventory.quantity(ItemCoal) != 0 || player.inventory.quantity(ItemIronBar) != 1 {
		t.Fatalf("inventory after smelting = %+v", player.inventory)
	}
	if got := player.skills[SkillSmithing]; got != SkillExperienceForLevel(2)+recipe.RewardXP {
		t.Fatalf("Smithing XP = %d", got)
	}
}

func TestApplySmeltingRecipeDoesNotConsumePartialInputs(t *testing.T) {
	player := &playerState{inventory: newEmptyInventory(), skills: newDefaultSkills()}
	player.inventory.add(ItemMithrilOre, 1)
	player.inventory.add(ItemCoal, 1)
	player.skills[SkillSmithing] = SkillExperienceForLevel(5)
	recipe, _ := smeltingRecipe(SmeltingMithrilBar)
	if err := applySmeltingRecipe(player, recipe); err == nil || !strings.Contains(err.Error(), "inputs") {
		t.Fatalf("applySmeltingRecipe() error = %v, want missing inputs", err)
	}
	if player.inventory.quantity(ItemMithrilOre) != 1 || player.inventory.quantity(ItemCoal) != 1 {
		t.Fatalf("failed smelt consumed inputs: %+v", player.inventory)
	}
}

func TestBuildSmeltingRecipesRejectsInvalidCatalogReferences(t *testing.T) {
	_, err := buildSmeltingRecipes([]smeltingRecipeDefinition{{
		ID: "invalid", Name: "Invalid", RequiredLevel: 1,
		Costs:      []ItemCost{{ItemID: "missing_ore", Quantity: 1}},
		RewardItem: ItemBronzeBar, RewardQuantity: 1, RewardXP: 1, Interval: 1,
	}})
	if err == nil {
		t.Fatal("buildSmeltingRecipes() error = nil for an unknown input")
	}
}
