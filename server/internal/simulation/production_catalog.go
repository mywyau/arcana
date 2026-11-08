package simulation

import "sort"

// ProductionRecipeSnapshot is the client-facing, server-authoritative recipe
// metadata needed to render a production interface.
type ProductionRecipeSnapshot struct {
	Station        string
	Group          string
	ID             string
	Name           string
	RequiredLevel  int
	Costs          []ItemCost
	RewardItem     string
	RewardQuantity int
	Interval       uint64
}

// ProductionRecipeCatalog returns a detached, deterministically ordered copy
// of the production recipes currently exposed to the client.
func ProductionRecipeCatalog() []ProductionRecipeSnapshot {
	recipes := make([]ProductionRecipeSnapshot, 0, len(cookingRecipes)+len(smeltingRecipes)+len(forgingRecipes))
	for _, recipe := range cookingRecipes {
		recipes = append(recipes, ProductionRecipeSnapshot{
			Station: string(craftingStationCooking), ID: recipe.ID, Name: recipe.Name,
			RequiredLevel: recipe.RequiredLevel, Costs: append([]ItemCost(nil), recipe.Costs...),
			RewardItem: recipe.RewardItem, RewardQuantity: recipe.RewardQuantity, Interval: recipe.Interval,
		})
	}
	for _, recipe := range smeltingRecipes {
		recipes = append(recipes, ProductionRecipeSnapshot{
			Station: string(craftingStationSmelting), ID: string(recipe.ID), Name: recipe.Name,
			RequiredLevel: recipe.RequiredLevel, Costs: append([]ItemCost(nil), recipe.Costs...),
			RewardItem: recipe.RewardItem, RewardQuantity: recipe.RewardQuantity, Interval: recipe.Interval,
		})
	}
	for _, recipe := range forgingRecipes {
		recipes = append(recipes, ProductionRecipeSnapshot{
			Station: string(craftingStationForging), Group: forgingRecipeGroup(recipe),
			ID: string(recipe.ID), Name: recipe.Name, RequiredLevel: recipe.RequiredLevel,
			Costs: append([]ItemCost(nil), recipe.Costs...), RewardItem: recipe.RewardItem,
			RewardQuantity: recipe.RewardQuantity, Interval: recipe.Interval,
		})
	}
	sort.Slice(recipes, func(i, j int) bool {
		if recipes[i].Station != recipes[j].Station {
			return recipes[i].Station < recipes[j].Station
		}
		if recipes[i].RequiredLevel != recipes[j].RequiredLevel {
			return recipes[i].RequiredLevel < recipes[j].RequiredLevel
		}
		return recipes[i].ID < recipes[j].ID
	})
	return recipes
}

func forgingRecipeGroup(recipe forgingRecipeDefinition) string {
	for _, cost := range recipe.Costs {
		switch cost.ItemID {
		case ItemBronzeBar:
			return "bronze"
		case ItemIronBar:
			return "iron"
		case ItemMithrilBar:
			return "mithril"
		case ItemGoldBar:
			return "jewellery"
		}
	}
	return "other"
}
