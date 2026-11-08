package simulation

import (
	"errors"
	"fmt"
)

func validateProductionAction(player *playerState, spot gatheringSpotDefinition, command *GatherCommand) error {
	switch spot.Station {
	case craftingStationCooking:
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
	case craftingStationSmelting:
		recipe, exists := smeltingRecipe(SmeltingRecipeID(command.RecipeID))
		if exists {
			if SkillLevelForExperience(player.skills[SkillSmithing]) < recipe.RequiredLevel {
				return fmt.Errorf("smithing level %d is required", recipe.RequiredLevel)
			}
			if _, err := smeltingInventoryAfter(player.inventory, recipe); err != nil {
				return err
			}
			break
		}
		jewellery, jewelleryExists := forgingRecipe(ForgingRecipeID(command.RecipeID))
		if !jewelleryExists || forgingRecipeStation(jewellery) != craftingStationSmelting {
			return fmt.Errorf("furnace recipe %q does not exist", command.RecipeID)
		}
		if SkillLevelForExperience(player.skills[SkillSmithing]) < jewellery.RequiredLevel {
			return fmt.Errorf("smithing level %d is required", jewellery.RequiredLevel)
		}
		if _, err := forgingInventoryAfter(player.inventory, jewellery); err != nil {
			return err
		}
	case craftingStationForging:
		recipe, exists := forgingRecipe(ForgingRecipeID(command.RecipeID))
		if !exists {
			reinforcement, reinforcementExists := reinforcementRecipe(command.RecipeID)
			if !reinforcementExists {
				return fmt.Errorf("forging recipe %q does not exist", command.RecipeID)
			}
			return validateReinforcement(player, reinforcement, command.InstanceID)
		}
		if SkillLevelForExperience(player.skills[SkillSmithing]) < recipe.RequiredLevel {
			return fmt.Errorf("smithing level %d is required", recipe.RequiredLevel)
		}
		if _, err := forgingInventoryAfter(player.inventory, recipe); err != nil {
			return err
		}
	case craftingStationRunecrafting:
		return nil
	default:
		return fmt.Errorf("production station %q is not supported", spot.Station)
	}
	return nil
}

func (w *World) resolveProductionAction(player *playerState, spot gatheringSpotDefinition) {
	switch spot.Station {
	case craftingStationCooking:
		w.resolveCookingAction(player, spot)
	case craftingStationSmelting:
		if _, exists := smeltingRecipe(SmeltingRecipeID(player.gatheringRecipe)); exists {
			w.resolveSmeltingAction(player)
		} else {
			w.resolveForgingAction(player)
		}
	case craftingStationForging:
		w.resolveForgingAction(player)
	case craftingStationRunecrafting:
		w.resolveIngredientProduction(player, spot)
	default:
		clearGathering(player, "station_missing")
	}
}

func (w *World) resolveCookingAction(player *playerState, spot gatheringSpotDefinition) {
	recipe, exists := foodRecipe(spot.Station, player.gatheringRecipe)
	if !exists {
		clearGathering(player, "recipe_missing")
		return
	}
	if err := applyFoodRecipe(player, spot.Skill, recipe); err != nil {
		if errors.Is(err, errFoodInventoryFull) {
			clearGathering(player, "inventory_full")
		} else {
			clearGathering(player, "missing_"+recipe.Costs[0].ItemID)
		}
		return
	}
	w.finishProductionStep(player, recipe.RewardItem, recipe.Interval)
}

func (w *World) resolveSmeltingAction(player *playerState) {
	recipe, exists := smeltingRecipe(SmeltingRecipeID(player.gatheringRecipe))
	if !exists {
		clearGathering(player, "recipe_missing")
		return
	}
	if err := applySmeltingRecipe(player, recipe); err != nil {
		if errors.Is(err, errSmeltingInventoryFull) {
			clearGathering(player, "inventory_full")
		} else {
			clearGathering(player, "missing_smelting_inputs")
		}
		return
	}
	w.finishProductionStep(player, "smelted_"+recipe.RewardItem, recipe.Interval)
}

func (w *World) resolveForgingAction(player *playerState) {
	recipe, exists := forgingRecipe(ForgingRecipeID(player.gatheringRecipe))
	if !exists {
		reinforcement, reinforcementExists := reinforcementRecipe(player.gatheringRecipe)
		if !reinforcementExists {
			clearGathering(player, "recipe_missing")
			return
		}
		if err := applyReinforcement(player, reinforcement, player.gatheringInstanceID); err != nil {
			clearGathering(player, "reinforcement_failed")
			return
		}
		level := 0
		if item, itemExists := player.inventory.itemByInstance(player.gatheringInstanceID); itemExists {
			level = item.Reinforcement
		}
		clearGathering(player, fmt.Sprintf("reinforced_%s_to_%d", reinforcement.TargetItem, level))
		return
	}
	if err := applyForgingRecipe(player, recipe); err != nil {
		if errors.Is(err, errForgingInventoryFull) {
			clearGathering(player, "inventory_full")
		} else {
			clearGathering(player, "missing_forging_inputs")
		}
		return
	}
	player.gatheringResult = "forged_" + recipe.RewardItem
	player.nextGatherPulse = w.combat.pulse + recipe.Interval
}

func (w *World) resolveIngredientProduction(player *playerState, spot gatheringSpotDefinition) {
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
	grantSkillExperience(player, spot.Skill, spot.RewardXP)
	w.finishProductionStep(player, spot.RewardResult, spot.Interval)
}

func (w *World) finishProductionStep(player *playerState, result string, interval uint64) {
	if player.productionBatch.completeStep() {
		clearGathering(player, result)
		return
	}
	player.gatheringResult = result
	player.nextGatherPulse = w.combat.pulse + interval
}
