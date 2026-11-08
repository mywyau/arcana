package simulation

import (
	"errors"
	"fmt"
)

func validateRenewableGathering(player *playerState, spot gatheringSpotDefinition, command *GatherCommand) error {
	if spot.Station != craftingStationFishing {
		return nil
	}
	if command.RecipeID != "" && command.RecipeID != spot.RecipeID {
		return fmt.Errorf("fishing spot %q does not contain recipe %q", spot.ID, command.RecipeID)
	}
	command.RecipeID = spot.RecipeID
	recipe, exists := foodRecipe(spot.Station, command.RecipeID)
	if !exists {
		return fmt.Errorf("%s recipe %q does not exist", spot.Station, command.RecipeID)
	}
	if SkillLevelForExperience(player.skills[spot.Skill]) < recipe.RequiredLevel {
		return fmt.Errorf("%s level %d is required", spot.Skill, recipe.RequiredLevel)
	}
	if _, err := foodInventoryAfter(player.inventory, recipe); err != nil {
		return err
	}
	command.RecipeID = recipe.ID
	return nil
}

func (w *World) resolveRenewableGathering(player *playerState, spot gatheringSpotDefinition) {
	if spot.Station == craftingStationFishing {
		recipe, exists := foodRecipe(spot.Station, player.gatheringRecipe)
		if !exists {
			clearGathering(player, "recipe_missing")
			return
		}
		if err := applyFoodRecipe(player, spot.Skill, recipe); err != nil {
			if errors.Is(err, errFoodInventoryFull) {
				clearGathering(player, "inventory_full")
			} else {
				clearGathering(player, "fishing_failed")
			}
			return
		}
		player.gatheringResult = "caught_" + recipe.RewardItem
		player.nextGatherPulse = w.combat.pulse + recipe.Interval
		return
	}

	if spot.MaximumAmount > 0 && spot.CurrentAmount <= 0 {
		result := spot.DepletedResult
		if result == "" {
			result = "vein_depleted"
		}
		clearGathering(player, result)
		return
	}
	candidate := cloneInventory(player.inventory)
	if !payItemCosts(&candidate, spot.Costs) {
		clearGathering(player, spot.MissingResult)
		return
	}
	if !candidate.add(spot.RewardItem, spot.RewardQuantity) {
		clearGathering(player, "inventory_full")
		return
	}
	player.inventory = candidate
	if spot.MaximumAmount > 0 {
		spot.CurrentAmount--
		if spot.NextRecharge == 0 {
			spot.NextRecharge = w.combat.pulse + spot.RechargePulses
		}
		w.gatheringSpots[spot.ID] = spot
	}
	grantSkillExperience(player, spot.Skill, spot.RewardXP)
	player.gatheringResult = spot.RewardResult
	player.nextGatherPulse = w.combat.pulse + spot.Interval
}
