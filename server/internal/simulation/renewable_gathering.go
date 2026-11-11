package simulation

import (
	"errors"
	"fmt"
	"math"
)

func gatheringSuccessChance(spot gatheringSpotDefinition, skillLevel int) float64 {
	if skillLevel < spot.RequiredLevel {
		return 0
	}
	chance := spot.BaseSuccessChance + float64(skillLevel-spot.RequiredLevel)*spot.SuccessChancePerLevel
	return math.Max(0, math.Min(chance, spot.MaximumSuccessChance))
}

func (w *World) gatheringAttemptSucceeded(player *playerState, spot gatheringSpotDefinition) bool {
	level := SkillLevelForExperience(player.skills[spot.Skill])
	chance := gatheringSuccessChance(spot, level)
	if bonus, found := gatheringToolBonus(player, spot.RequiredTool); found {
		chance = math.Min(chance+bonus, spot.MaximumSuccessChance)
	}
	return w.rollSource.Float64() < chance
}

func gatheringToolBonus(player *playerState, requiredTool string) (float64, bool) {
	required, exists := itemDefinition(requiredTool)
	if !exists || required.GatheringTool.Family == "" {
		return 0, player.inventory.contains(requiredTool)
	}
	bestBonus := 0.0
	found := false
	for _, item := range player.inventory.items {
		definition, itemExists := itemDefinition(item.ItemID)
		if !itemExists || item.Quantity <= 0 ||
			definition.GatheringTool.Family != required.GatheringTool.Family {
			continue
		}
		if _, unmet := unmetEquipRequirement(player, definition); unmet {
			continue
		}
		found = true
		bestBonus = math.Max(bestBonus, definition.GatheringTool.SuccessBonus)
	}
	return bestBonus, found
}

func unmetGatheringToolRequirement(player *playerState, requiredTool string) (SkillRequirement, bool) {
	required, exists := itemDefinition(requiredTool)
	if !exists || required.GatheringTool.Family == "" {
		return SkillRequirement{}, false
	}
	var lowest SkillRequirement
	found := false
	for _, item := range player.inventory.items {
		definition, itemExists := itemDefinition(item.ItemID)
		if !itemExists || item.Quantity <= 0 ||
			definition.GatheringTool.Family != required.GatheringTool.Family {
			continue
		}
		requirement, unmet := unmetEquipRequirement(player, definition)
		if unmet && (!found || requirement.Level < lowest.Level) {
			lowest = requirement
			found = true
		}
	}
	return lowest, found
}

func recordUnsuccessfulGatheringAttempt(
	player *playerState,
	spot gatheringSpotDefinition,
	pulse, interval uint64,
) {
	result := spot.UnsuccessfulResult
	if result == "" {
		result = "gathering_unsuccessful"
	}
	player.gatheringResult = result
	player.nextGatherPulse = pulse + interval
}

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
	// Level, tool and inventory requirements are resolved after the player
	// reaches the fishing spot. Rejecting them here makes the client begin its
	// predicted approach and then get corrected backwards by the server.
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
		candidate := cloneInventory(player.inventory)
		if !candidate.add(recipe.RewardItem, recipe.RewardQuantity) {
			clearGathering(player, "inventory_full")
			return
		}
		if !w.gatheringAttemptSucceeded(player, spot) {
			recordUnsuccessfulGatheringAttempt(player, spot, w.combat.pulse, recipe.Interval)
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
	if !w.gatheringAttemptSucceeded(player, spot) {
		recordUnsuccessfulGatheringAttempt(player, spot, w.combat.pulse, spot.Interval)
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
