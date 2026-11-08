package simulation

import "fmt"

// validateTimedAction keeps action-specific rules separate from the shared
// authoritative approach, interruption, and pulse lifecycle.
func (w *World) validateTimedAction(player *playerState, spot gatheringSpotDefinition, command *GatherCommand) error {
	switch spot.ActionKind {
	case timedActionGathering:
		return validateRenewableGathering(player, spot, command)
	case timedActionProduction:
		return validateProductionAction(player, spot, command)
	case timedActionWorldInteraction:
		return validateWorldInteraction(player, spot, command)
	default:
		return fmt.Errorf("gathering spot %q has an invalid action kind", spot.ID)
	}
}

func timedActionInterval(spot gatheringSpotDefinition, recipeID string) uint64 {
	if recipe, exists := foodRecipe(spot.Station, recipeID); exists {
		return recipe.Interval
	}
	if recipe, exists := smeltingRecipe(SmeltingRecipeID(recipeID)); spot.Station == craftingStationSmelting && exists {
		return recipe.Interval
	}
	if recipe, exists := forgingRecipe(ForgingRecipeID(recipeID)); spot.Station == craftingStationForging && exists {
		return recipe.Interval
	}
	if recipe, exists := reinforcementRecipe(recipeID); spot.Station == craftingStationForging && exists {
		return recipe.Interval
	}
	return spot.Interval
}
