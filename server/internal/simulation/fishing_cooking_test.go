package simulation

import (
	"strings"
	"testing"
)

func TestFishRecipesProvideLevelledProgression(t *testing.T) {
	tests := []struct {
		level  int
		raw    string
		cooked string
		heal   int
	}{
		{1, ItemRawMinnow, ItemCookedMinnow, 8},
		{2, ItemRawTrout, ItemCookedTrout, 12},
		{3, ItemRawSalmon, ItemCookedSalmon, 16},
		{4, ItemRawTuna, ItemCookedTuna, 20},
		{5, ItemRawLobster, ItemCookedLobster, 25},
	}
	for _, test := range tests {
		fishing, fishingExists := foodRecipe(craftingStationFishing, test.raw)
		cooking, cookingExists := foodRecipe(craftingStationCooking, test.cooked)
		food, foodExists := itemDefinition(test.cooked)
		if !fishingExists || fishing.RequiredLevel != test.level || fishing.RewardItem != test.raw {
			t.Errorf("fishing recipe %q = %+v, %t", test.raw, fishing, fishingExists)
		}
		if !cookingExists || cooking.RequiredLevel != test.level || cooking.RewardItem != test.cooked ||
			len(cooking.Costs) != 1 || cooking.Costs[0].ItemID != test.raw {
			t.Errorf("cooking recipe %q = %+v, %t", test.cooked, cooking, cookingExists)
		}
		if !foodExists || food.Consumable.Healing != test.heal || food.StackLimit != 10 {
			t.Errorf("food definition %q = %+v, %t", test.cooked, food, foodExists)
		}
	}
}

func TestApplySelectedFishRecipesCatchesAndCooksAtomically(t *testing.T) {
	player := &playerState{inventory: newEmptyInventory(), skills: newDefaultSkills()}
	player.skills[SkillFishing] = SkillExperienceForLevel(3)
	player.skills[SkillCooking] = SkillExperienceForLevel(3)

	fishing, _ := foodRecipe(craftingStationFishing, ItemRawSalmon)
	if err := applyFoodRecipe(player, SkillFishing, fishing); err != nil {
		t.Fatalf("applyFoodRecipe(fishing) returned an unexpected error: %v", err)
	}
	cooking, _ := foodRecipe(craftingStationCooking, ItemCookedSalmon)
	if err := applyFoodRecipe(player, SkillCooking, cooking); err != nil {
		t.Fatalf("applyFoodRecipe(cooking) returned an unexpected error: %v", err)
	}
	if player.inventory.quantity(ItemRawSalmon) != 0 || player.inventory.quantity(ItemCookedSalmon) != 1 {
		t.Fatalf("inventory after catch and cook = %+v", player.inventory)
	}
}

func TestFoodRecipeRejectsInsufficientLevelWithoutConsumingFish(t *testing.T) {
	player := &playerState{inventory: newEmptyInventory(), skills: newDefaultSkills()}
	player.inventory.add(ItemRawLobster, 1)
	recipe, _ := foodRecipe(craftingStationCooking, ItemCookedLobster)
	if err := applyFoodRecipe(player, SkillCooking, recipe); err == nil || !strings.Contains(err.Error(), "level 5") {
		t.Fatalf("applyFoodRecipe() error = %v, want level requirement", err)
	}
	if player.inventory.quantity(ItemRawLobster) != 1 {
		t.Fatalf("failed cooking attempt consumed raw lobster: %+v", player.inventory)
	}
}
