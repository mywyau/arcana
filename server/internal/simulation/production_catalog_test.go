package simulation

import "testing"

func TestProductionRecipeCatalogIsOrderedAndDetached(t *testing.T) {
	first := ProductionRecipeCatalog()
	wantCount := len(cookingRecipes) + len(smeltingRecipes) + len(forgingRecipes) + len(fletchingRecipes)
	if len(first) != wantCount {
		t.Fatalf("catalog recipe count = %d, want %d", len(first), wantCount)
	}
	for _, recipe := range first {
		if recipe.Station == string(craftingStationForging) && recipe.Group == "" {
			t.Fatalf("forging recipe %q has no catalogue group", recipe.ID)
		}
	}
	wantGroups := map[string]string{
		string(ForgingBronzeSword):    "bronze",
		string(ForgingIronSword):      "iron",
		string(ForgingMithrilSword):   "mithril",
		string(ForgingSapphireAmulet): "jewellery",
	}
	for _, recipe := range first {
		if want, exists := wantGroups[recipe.ID]; exists {
			if recipe.Group != want {
				t.Errorf("recipe %q group = %q, want %q", recipe.ID, recipe.Group, want)
			}
			delete(wantGroups, recipe.ID)
		}
	}
	for _, recipe := range first {
		if recipe.Station == "fletching" && recipe.RewardXP < 1 {
			t.Fatalf("fletching recipe %q has no XP reward", recipe.ID)
		}
		if recipe.ID == string(ForgingSapphireRing) {
			if recipe.Station != string(craftingStationSmelting) || recipe.Group != "jewellery" || recipe.RequiredTool != ItemRingMould {
				t.Fatalf("ring catalogue recipe = %+v", recipe)
			}
		}
		if recipe.ID == string(ForgingSapphireAmulet) && recipe.RequiredTool != ItemAmuletMould {
			t.Fatalf("amulet catalogue recipe = %+v", recipe)
		}
	}
	if len(wantGroups) != 0 {
		t.Fatalf("catalogue is missing grouped forging recipes: %+v", wantGroups)
	}
	for index, recipe := range first {
		if recipe.ID == "" || recipe.Name == "" || recipe.RequiredLevel < 1 || len(recipe.Costs) == 0 ||
			recipe.RewardItem == "" || recipe.RewardQuantity < 1 || recipe.Interval == 0 {
			t.Fatalf("catalog recipe %d is incomplete: %+v", index, recipe)
		}
		if index > 0 {
			previous := first[index-1]
			if previous.Station > recipe.Station || previous.Station == recipe.Station &&
				(previous.RequiredLevel > recipe.RequiredLevel ||
					previous.RequiredLevel == recipe.RequiredLevel && previous.ID > recipe.ID) {
				t.Fatalf("catalog is not ordered at %q then %q", previous.ID, recipe.ID)
			}
		}
	}

	first[0].Costs[0].Quantity = 999
	second := ProductionRecipeCatalog()
	if second[0].Costs[0].Quantity == 999 {
		t.Fatal("ProductionRecipeCatalog() returned shared cost data")
	}
}
