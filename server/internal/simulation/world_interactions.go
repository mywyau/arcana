package simulation

import "fmt"

func validateWorldInteraction(player *playerState, spot gatheringSpotDefinition, command *GatherCommand) error {
	if spot.Station != craftingStationPrayer {
		return nil
	}
	switch command.RecipeID {
	case "", prayerRestoreRecipeID:
		command.RecipeID = prayerRestoreRecipeID
	case prayerSacrificeBonesID:
		if !player.inventory.contains(ItemBones) {
			return fmt.Errorf("bones are required")
		}
	case prayerAdoptCommonID, prayerAdoptWarID:
		if command.RecipeID != spot.RecipeID {
			return fmt.Errorf("prayer set %q is not available at altar %q", command.RecipeID, spot.ID)
		}
	default:
		return fmt.Errorf("prayer altar recipe %q does not exist", command.RecipeID)
	}
	return nil
}

func (w *World) resolveWorldInteraction(player *playerState, spot gatheringSpotDefinition) {
	if spot.Station == craftingStationPrayer {
		w.resolvePrayerAltarAction(player, spot)
		return
	}
	if player.gatheringTraversal {
		player.position = player.gatheringCompletionPosition
		stopPlayer(player)
	}
	grantSkillExperience(player, spot.Skill, spot.RewardXP)
	clearGathering(player, spot.RewardResult)
}

func (w *World) resolvePrayerAltarAction(player *playerState, spot gatheringSpotDefinition) {
	if player.gatheringRecipe == prayerAdoptCommonID || player.gatheringRecipe == prayerAdoptWarID {
		setID := PrayerSetCommon
		if player.gatheringRecipe == prayerAdoptWarID {
			setID = PrayerSetWar
		}
		player.activePrayerSet = setID
		clearActivePrayers(player)
		clearGathering(player, "prayer_set_"+string(setID))
		return
	}
	if player.gatheringRecipe == prayerSacrificeBonesID {
		if !player.inventory.remove(ItemBones, 1) {
			clearGathering(player, "missing_bones")
			return
		}
		grantSkillExperience(player, SkillPrayer, prayerBoneExperience)
		if !player.inventory.contains(ItemBones) {
			clearGathering(player, "sacrificed_all_bones")
			return
		}
		player.gatheringResult = "sacrificed_bone"
		player.nextGatherPulse = w.combat.pulse + 1
		return
	}
	player.prayer = player.maximumPrayer
	clearGathering(player, spot.RewardResult)
}
