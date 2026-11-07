package simulation

import "fmt"

type reinforcementRecipeDefinition struct {
	ID            string
	Name          string
	RequiredLevel int
	TargetItem    string
	Costs         []ItemCost
	RewardXP      int64
	Interval      uint64
}

var reinforcementRecipes = map[string]reinforcementRecipeDefinition{
	"reinforce_bronze_sword": {
		ID: "reinforce_bronze_sword", Name: "Reinforce Bronze Sword", RequiredLevel: 2,
		TargetItem: string(WeaponBronzeSword), Costs: []ItemCost{{ItemID: ItemBronzeBar, Quantity: 3}}, RewardXP: 30, Interval: 5,
	},
	"reinforce_bronze_helm": {
		ID: "reinforce_bronze_helm", Name: "Reinforce Bronze Helm", RequiredLevel: 2,
		TargetItem: ItemBronzeHelm, Costs: []ItemCost{{ItemID: ItemBronzeBar, Quantity: 2}}, RewardXP: 24, Interval: 5,
	},
}

func reinforcementRecipe(id string) (reinforcementRecipeDefinition, bool) {
	recipe, exists := reinforcementRecipes[id]
	return recipe, exists
}

func validateReinforcement(player *playerState, recipe reinforcementRecipeDefinition, instanceID string) error {
	if player == nil {
		return fmt.Errorf("player is required")
	}
	if SkillLevelForExperience(player.skills[SkillSmithing]) < recipe.RequiredLevel {
		return fmt.Errorf("smithing level %d is required", recipe.RequiredLevel)
	}
	item, exists := player.inventory.itemByInstance(instanceID)
	if !exists || item.ItemID != recipe.TargetItem {
		return fmt.Errorf("target equipment instance is missing")
	}
	if item.Reinforcement >= 1 {
		return fmt.Errorf("equipment is already reinforced")
	}
	candidate := cloneInventory(player.inventory)
	if !payItemCosts(&candidate, recipe.Costs) {
		return fmt.Errorf("required reinforcement bars are missing")
	}
	return nil
}

func applyReinforcement(player *playerState, recipe reinforcementRecipeDefinition, instanceID string) error {
	if err := validateReinforcement(player, recipe, instanceID); err != nil {
		return err
	}
	if !payItemCosts(&player.inventory, recipe.Costs) {
		return fmt.Errorf("required reinforcement bars are missing")
	}
	for index := range player.inventory.items {
		if player.inventory.items[index].InstanceID == instanceID {
			player.inventory.items[index].Reinforcement = 1
			grantSkillExperience(player, SkillSmithing, recipe.RewardXP)
			return nil
		}
	}
	return fmt.Errorf("target equipment instance is missing")
}

func equippedReinforcement(player *playerState, instanceID string) int {
	if player == nil || instanceID == "" {
		return 0
	}
	if instanceID == player.equippedWeaponInstanceID {
		return player.equippedWeaponReinforcement
	}
	for slot, equippedInstanceID := range player.equippedArmourInstances {
		if instanceID == equippedInstanceID {
			return player.equippedArmourReinforcements[slot]
		}
	}
	return 0
}
