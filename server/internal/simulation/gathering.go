package simulation

import (
	"fmt"
	"sort"
)

const (
	starterFishingSpotID      = "starter_pond"
	starterTroutSpotID        = "trout_pond"
	starterSalmonSpotID       = "salmon_pond"
	starterTunaSpotID         = "tuna_pond"
	starterLobsterSpotID      = "lobster_pond"
	starterCookingRangeID     = "starter_cooking_range"
	starterFurnaceID          = "starter_furnace"
	starterAnvilID            = "starter_anvil"
	gatheringRange            = 2.5
	fishingInterval           = 4
	fishingExperience         = 10
	cookingInterval           = 4
	cookingExperience         = 10
	starterCopperVeinID       = "starter_copper_vein"
	starterTinVeinID          = "starter_tin_vein"
	starterIronVeinID         = "starter_iron_vein"
	starterCoalVeinID         = "starter_coal_vein"
	starterMithrilVeinID      = "starter_mithril_vein"
	starterGoldVeinID         = "starter_gold_vein"
	starterNormalTreeID       = "starter_normal_tree"
	starterOakTreeID          = "starter_oak_tree"
	starterWillowTreeID       = "starter_willow_tree"
	starterMapleTreeID        = "starter_maple_tree"
	starterYewTreeID          = "starter_yew_tree"
	starterAgilityObstacleID  = "starter_balance_beam"
	starterPrayerAltarID      = "starter_prayer_altar"
	warPrayerAltarID          = "war_prayer_altar"
	starterAirRuneAltarID     = "starter_air_rune_altar"
	miningInterval            = 4
	miningExperience          = 10
	copperVeinCapacity        = 20
	tinVeinCapacity           = 20
	copperRechargePulses      = 10
	ironVeinCapacity          = 16
	coalVeinCapacity          = 12
	mithrilVeinCapacity       = 8
	goldVeinCapacity          = 10
	normalTreeCapacity        = 10
	oakTreeCapacity           = 8
	willowTreeCapacity        = 8
	mapleTreeCapacity         = 6
	yewTreeCapacity           = 4
	agilityObstacleExperience = 50
	agilityObstacleInterval   = 5
	agilityBeamHalfLength     = 3.25
	agilityEndpointTolerance  = 0.15
	prayerBoneExperience      = 10
	prayerRestoreRecipeID     = "restore_prayer"
	prayerSacrificeBonesID    = "sacrifice_bones"
	prayerAdoptCommonID       = "adopt_common_litany"
	prayerAdoptWarID          = "adopt_war_covenant"
	runecraftingInterval      = 3
	runecraftingExperience    = 12
)

type craftingStationID string

type timedActionKind uint8

const (
	timedActionUnknown timedActionKind = iota
	timedActionGathering
	timedActionProduction
	timedActionWorldInteraction
)

const (
	craftingStationFishing      craftingStationID = "fishing"
	craftingStationCooking      craftingStationID = "cooking"
	craftingStationSmelting     craftingStationID = "smelting"
	craftingStationForging      craftingStationID = "forging"
	craftingStationPrayer       craftingStationID = "prayer_altar"
	craftingStationRunecrafting craftingStationID = "runecrafting"
)

type GatheringSpotSnapshot struct {
	ID            string
	Name          string
	Skill         SkillID
	Position      Position
	Range         float64
	RequiredLevel int
	CurrentAmount int
	MaximumAmount int
	Station       craftingStationID
	RecipeID      string
}

type gatheringSpotDefinition struct {
	GatheringSpotSnapshot
	ActionKind     timedActionKind
	RewardItem     string
	RewardQuantity int
	RewardXP       int64
	Interval       uint64
	Costs          []ItemCost
	ActiveResult   string
	RewardResult   string
	MissingResult  string
	DepletedResult string
	RequiredTool   string
	RechargePulses uint64
	NextRecharge   uint64
}

type GatherCommand struct {
	PlayerID   PlayerID
	Sequence   uint64
	SpotID     string
	RecipeID   string
	InstanceID string
	Quantity   int
}

type pendingGather struct {
	command             GatherCommand
	path                []Position
	interactionPosition Position
	completionPosition  Position
	traversal           bool
}

func agilityBeamEndpoints(center Position) (Position, Position) {
	left := center
	left.X -= agilityBeamHalfLength
	right := center
	right.X += agilityBeamHalfLength
	return left, right
}

func newGatheringSpots() map[string]gatheringSpotDefinition {
	return map[string]gatheringSpotDefinition{
		starterFishingSpotID: fishingSpot(starterFishingSpotID, "Minnow Pond", Position{X: -14, Z: 14}, ItemRawMinnow),
		starterTroutSpotID:   fishingSpot(starterTroutSpotID, "Trout Pond", Position{X: -23, Z: 18}, ItemRawTrout),
		starterSalmonSpotID:  fishingSpot(starterSalmonSpotID, "Salmon Pond", Position{X: -25, Z: 8}, ItemRawSalmon),
		starterTunaSpotID:    fishingSpot(starterTunaSpotID, "Tuna Fishing Spot", Position{X: -24, Z: -3}, ItemRawTuna),
		starterLobsterSpotID: fishingSpot(starterLobsterSpotID, "Lobster Fishing Spot", Position{X: -24, Z: -14}, ItemRawLobster),
		starterCookingRangeID: {
			GatheringSpotSnapshot: GatheringSpotSnapshot{
				ID: starterCookingRangeID, Name: "Cooking Range", Skill: SkillCooking,
				Position: Position{X: -10, Z: -21}, Range: gatheringRange, RequiredLevel: 1,
				Station: craftingStationCooking,
			},
			ActionKind:   timedActionProduction,
			ActiveResult: "cooking",
		},
		starterFurnaceID: {
			GatheringSpotSnapshot: GatheringSpotSnapshot{
				ID: starterFurnaceID, Name: "Furnace", Skill: SkillSmithing,
				Position: Position{X: 10, Z: 12}, Range: gatheringRange, RequiredLevel: 1,
				Station: craftingStationSmelting,
			},
			ActionKind:   timedActionProduction,
			ActiveResult: "smelting",
		},
		starterAnvilID: {
			GatheringSpotSnapshot: GatheringSpotSnapshot{
				ID: starterAnvilID, Name: "Anvil", Skill: SkillSmithing,
				Position: Position{X: 7, Z: 11}, Range: gatheringRange, RequiredLevel: 1,
				Station: craftingStationForging,
			},
			ActionKind:   timedActionProduction,
			ActiveResult: "forging",
		},
		starterCopperVeinID: {
			GatheringSpotSnapshot: GatheringSpotSnapshot{
				ID: starterCopperVeinID, Name: "Copper Vein", Skill: SkillMining,
				Position: Position{X: 19, Z: 12}, Range: gatheringRange, RequiredLevel: 1,
				CurrentAmount: copperVeinCapacity, MaximumAmount: copperVeinCapacity,
			},
			ActionKind: timedActionGathering,
			RewardItem: ItemCopperOre, RewardQuantity: 1, RewardXP: miningExperience,
			Interval: miningInterval, ActiveResult: "mining", RewardResult: "mined_copper_ore",
			MissingResult: "missing_bronze_pickaxe", RequiredTool: ItemBronzePickaxe,
			RechargePulses: copperRechargePulses,
		},
		starterTinVeinID:     miningVein(starterTinVeinID, "Tin Vein", Position{X: 22, Z: 15}, 1, ItemTinOre, 10, tinVeinCapacity, 10, 4),
		starterIronVeinID:    miningVein(starterIronVeinID, "Iron Vein", Position{X: 25, Z: 11}, 2, ItemIronOre, 16, ironVeinCapacity, 12, 4),
		starterCoalVeinID:    miningVein(starterCoalVeinID, "Coal Vein", Position{X: 20, Z: 18}, 3, ItemCoal, 24, coalVeinCapacity, 16, 5),
		starterMithrilVeinID: miningVein(starterMithrilVeinID, "Mithril Vein", Position{X: 25, Z: 18}, 5, ItemMithrilOre, 40, mithrilVeinCapacity, 24, 6),
		starterGoldVeinID:    miningVein(starterGoldVeinID, "Gold Vein", Position{X: 22, Z: 22}, 4, ItemGoldOre, 32, goldVeinCapacity, 20, 5),
		starterNormalTreeID:  woodcuttingTree(starterNormalTreeID, "Tree", Position{X: -16, Z: -2}, 1, ItemNormalLogs, 10, normalTreeCapacity, 8, 4),
		starterOakTreeID:     woodcuttingTree(starterOakTreeID, "Oak Tree", Position{X: -7, Z: 24}, 2, ItemOakLogs, 16, oakTreeCapacity, 12, 5),
		starterWillowTreeID:  woodcuttingTree(starterWillowTreeID, "Willow Tree", Position{X: -24, Z: 7}, 3, ItemWillowLogs, 22, willowTreeCapacity, 14, 5),
		starterMapleTreeID:   woodcuttingTree(starterMapleTreeID, "Maple Tree", Position{X: 25, Z: 8}, 5, ItemMapleLogs, 34, mapleTreeCapacity, 18, 6),
		starterYewTreeID:     woodcuttingTree(starterYewTreeID, "Yew Tree", Position{X: 18, Z: -24}, 8, ItemYewLogs, 52, yewTreeCapacity, 24, 7),
		starterAgilityObstacleID: {
			GatheringSpotSnapshot: GatheringSpotSnapshot{
				ID: starterAgilityObstacleID, Name: "Balance Beam", Skill: SkillAgility,
				Position: Position{X: -36, Z: 32}, Range: gatheringRange, RequiredLevel: 1,
			},
			ActionKind: timedActionWorldInteraction,
			RewardXP:   agilityObstacleExperience, Interval: agilityObstacleInterval,
			ActiveResult: "balancing", RewardResult: "completed_agility_obstacle",
		},
		starterPrayerAltarID: {
			GatheringSpotSnapshot: GatheringSpotSnapshot{
				ID: starterPrayerAltarID, Name: "Prayer Altar", Skill: SkillPrayer,
				Position: Position{X: -8, Z: -14}, Range: gatheringRange, RequiredLevel: 1,
				Station: craftingStationPrayer, RecipeID: prayerAdoptCommonID,
			},
			ActionKind: timedActionWorldInteraction,
			Interval:   1, ActiveResult: "praying", RewardResult: "prayer_restored",
		},
		warPrayerAltarID: {
			GatheringSpotSnapshot: GatheringSpotSnapshot{
				ID: warPrayerAltarID, Name: "War Altar", Skill: SkillPrayer,
				Position: Position{X: 13, Z: -20}, Range: gatheringRange, RequiredLevel: 1,
				Station: craftingStationPrayer, RecipeID: prayerAdoptWarID,
			},
			ActionKind: timedActionWorldInteraction,
			Interval:   1, ActiveResult: "praying", RewardResult: "prayer_restored",
		},
		starterAirRuneAltarID: {
			GatheringSpotSnapshot: GatheringSpotSnapshot{
				ID: starterAirRuneAltarID, Name: "Air Rune Altar", Skill: SkillRunecrafting,
				Position: Position{X: 32, Z: -12}, Range: gatheringRange, RequiredLevel: 1,
				Station: craftingStationRunecrafting,
			},
			ActionKind: timedActionProduction,
			RewardItem: ItemAirRune, RewardQuantity: 5, RewardXP: runecraftingExperience,
			Interval: runecraftingInterval, Costs: []ItemCost{{ItemID: ItemBlankRune, Quantity: 1}},
			ActiveResult: "runecrafting", RewardResult: "crafted_air_runes",
			MissingResult: "missing_blank_runes",
		},
	}
}

func fishingSpot(id, name string, position Position, recipeID string) gatheringSpotDefinition {
	recipe, exists := fishingRecipes[recipeID]
	if !exists {
		panic(fmt.Sprintf("fishing spot %q uses unknown recipe %q", id, recipeID))
	}
	requiredTool := ItemFishingRod
	missingResult := "missing_fishing_rod"
	if recipeID == ItemRawLobster {
		requiredTool = ItemLobsterCage
		missingResult = "missing_lobster_cage"
	}
	return gatheringSpotDefinition{
		GatheringSpotSnapshot: GatheringSpotSnapshot{
			ID: id, Name: name, Skill: SkillFishing, Position: position, Range: gatheringRange,
			RequiredLevel: recipe.RequiredLevel, Station: craftingStationFishing, RecipeID: recipeID,
		},
		ActionKind:   timedActionGathering,
		ActiveResult: "fishing", RequiredTool: requiredTool, MissingResult: missingResult,
	}
}

func miningVein(id, name string, position Position, level int, reward string, xp int64, capacity int, recharge, interval uint64) gatheringSpotDefinition {
	return gatheringSpotDefinition{
		GatheringSpotSnapshot: GatheringSpotSnapshot{
			ID: id, Name: name, Skill: SkillMining, Position: position, Range: gatheringRange,
			RequiredLevel: level, CurrentAmount: capacity, MaximumAmount: capacity,
		},
		ActionKind: timedActionGathering,
		RewardItem: reward, RewardQuantity: 1, RewardXP: xp, Interval: interval,
		ActiveResult: "mining", RewardResult: "mined_" + reward,
		MissingResult: "missing_bronze_pickaxe", RequiredTool: ItemBronzePickaxe, RechargePulses: recharge,
	}
}

func woodcuttingTree(id, name string, position Position, level int, reward string, xp int64, capacity int, recharge, interval uint64) gatheringSpotDefinition {
	return gatheringSpotDefinition{
		GatheringSpotSnapshot: GatheringSpotSnapshot{
			ID: id, Name: name, Skill: SkillWoodcutting, Position: position, Range: gatheringRange,
			RequiredLevel: level, CurrentAmount: capacity, MaximumAmount: capacity,
		},
		ActionKind: timedActionGathering,
		RewardItem: reward, RewardQuantity: 1, RewardXP: xp, Interval: interval,
		ActiveResult: "woodcutting", RewardResult: "chopped_" + reward,
		MissingResult: "missing_bronze_axe", DepletedResult: "tree_depleted",
		RequiredTool: ItemBronzeAxe, RechargePulses: recharge,
	}
}

func (w *World) QueueGather(command GatherCommand) error {
	if command.Sequence == 0 {
		return fmt.Errorf("gather sequence must be positive")
	}
	if command.Quantity < 0 || command.Quantity > 10_000 {
		return fmt.Errorf("gather quantity must be between 0 and 10000")
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	spot, exists := w.gatheringSpots[command.SpotID]
	if !exists {
		return fmt.Errorf("gathering spot %q does not exist", command.SpotID)
	}
	player, exists := w.players[command.PlayerID]
	if !exists {
		return fmt.Errorf("player %q does not exist", command.PlayerID)
	}
	if player.defeated {
		return fmt.Errorf("player %q is defeated", command.PlayerID)
	}
	if command.Sequence <= player.lastGatherSequence {
		return fmt.Errorf("gather sequence %d is not newer than %d", command.Sequence, player.lastGatherSequence)
	}
	if pending, exists := w.pendingGathering[command.PlayerID]; exists && command.Sequence <= pending.command.Sequence {
		return fmt.Errorf("gather sequence %d is not newer than pending sequence %d", command.Sequence, pending.command.Sequence)
	}
	// Natural resources validate requirements after the authoritative approach.
	// Stations keep their early validation because selecting a recipe is a
	// separate UI action rather than a direct world interaction.
	if spot.Station != "" && SkillLevelForExperience(player.skills[spot.Skill]) < spot.RequiredLevel {
		return fmt.Errorf("%s level %d is required", spot.Skill, spot.RequiredLevel)
	}
	if err := w.validateTimedAction(player, spot, &command); err != nil {
		return err
	}
	if command.Quantity > 0 && !usesProductionBatch(spot.Station) {
		return fmt.Errorf("quantity is not supported for gathering spot %q", spot.ID)
	}
	if spot.Station != "" && spot.MaximumAmount > 0 && spot.CurrentAmount <= 0 {
		return fmt.Errorf("gathering spot %q is depleted", spot.ID)
	}

	interactionPosition := spot.Position
	completionPosition := spot.Position
	traversal := spot.ID == starterAgilityObstacleID
	approachDistance := spot.Range
	if traversal {
		left, right := agilityBeamEndpoints(spot.Position)
		interactionPosition, completionPosition = left, right
		if distance(player.position, right) < distance(player.position, left) {
			interactionPosition, completionPosition = right, left
		}
		approachDistance = agilityEndpointTolerance
	}
	path := []Position{interactionPosition}
	if distance(player.position, interactionPosition) > approachDistance && w.pathfinder != nil {
		var err error
		if traversal {
			path, err = w.pathfinder.FindPath(player.position, interactionPosition)
		} else {
			path, err = w.pathfinder.FindPathWithinRange(player.position, interactionPosition, spot.Range-0.25)
		}
		if err != nil {
			return fmt.Errorf("find gathering path: %w", err)
		}
	}
	if len(path) == 0 {
		return fmt.Errorf("find gathering path: empty path")
	}
	delete(w.pendingMoves, command.PlayerID)
	delete(w.pendingPickups, command.PlayerID)
	delete(w.combat.pendingTargets, command.PlayerID)
	player.autoPickingUp = false
	player.combatTarget = ""
	player.autoApproaching = false
	w.pendingGathering[command.PlayerID] = pendingGather{
		command: command, path: clonePositions(path),
		interactionPosition: interactionPosition,
		completionPosition:  completionPosition,
		traversal:           traversal,
	}
	return nil
}

func (w *World) startPendingGathering() {
	for id, pending := range w.pendingGathering {
		player, exists := w.players[id]
		if !exists || player.defeated {
			delete(w.pendingGathering, id)
			continue
		}
		spot, exists := w.gatheringSpots[pending.command.SpotID]
		if !exists {
			delete(w.pendingGathering, id)
			continue
		}
		player.lastGatherSequence = pending.command.Sequence
		player.gatheringSpot = spot.ID
		player.gatheringRecipe = pending.command.RecipeID
		player.gatheringInstanceID = pending.command.InstanceID
		player.productionBatch = productionBatch{}
		if usesProductionBatch(spot.Station) {
			player.productionBatch = newProductionBatch(pending.command.Quantity)
		}
		player.gatheringInteractionPosition = pending.interactionPosition
		player.gatheringCompletionPosition = pending.completionPosition
		player.gatheringTraversal = pending.traversal
		interval := timedActionInterval(spot, player.gatheringRecipe)
		player.nextGatherPulse = w.combat.pulse + interval
		interactionRange := spot.Range
		if pending.traversal {
			interactionRange = agilityEndpointTolerance
		}
		if distance(player.position, pending.interactionPosition) > interactionRange {
			player.destination = positionPointer(pending.path[len(pending.path)-1])
			player.path = clonePositions(pending.path)
			player.gatheringResult = "approaching"
		} else {
			stopPlayer(player)
			player.gatheringResult = spot.ActiveResult
		}
		delete(w.pendingGathering, id)
	}
}

func (w *World) resolveGatheringActions() {
	w.rechargeGatheringSpots()
	playerIDs := make([]PlayerID, 0, len(w.players))
	for id := range w.players {
		playerIDs = append(playerIDs, id)
	}
	sort.Slice(playerIDs, func(i, j int) bool { return playerIDs[i] < playerIDs[j] })
	for _, id := range playerIDs {
		player := w.players[id]
		if player.defeated {
			clearGathering(player, "")
			continue
		}
		if player.gatheringSpot == "" {
			continue
		}
		spot, exists := w.gatheringSpots[player.gatheringSpot]
		if !exists {
			clearGathering(player, "spot_missing")
			continue
		}
		interactionPosition := spot.Position
		interactionRange := spot.Range
		if player.gatheringTraversal {
			interactionPosition = player.gatheringInteractionPosition
			interactionRange = agilityEndpointTolerance
		}
		if distance(player.position, interactionPosition) > interactionRange {
			if player.destination == nil {
				clearGathering(player, "spot_unreachable")
			}
			continue
		}
		if player.destination != nil {
			stopPlayer(player)
		}
		if player.gatheringTraversal && player.gatheringResult == "approaching" {
			player.gatheringResult = spot.ActiveResult
			player.nextGatherPulse = w.combat.pulse + spot.Interval
			continue
		}
		if SkillLevelForExperience(player.skills[spot.Skill]) < spot.RequiredLevel {
			clearGathering(player, fmt.Sprintf("requires_%s_level_%d", spot.Skill, spot.RequiredLevel))
			continue
		}
		// Resolve tool requirements at the resource so the approach remains a
		// valid server-driven route even when the player lacks the tool.
		if spot.RequiredTool != "" && !player.inventory.contains(spot.RequiredTool) {
			clearGathering(player, spot.MissingResult)
			continue
		}
		if w.combat.pulse < player.nextGatherPulse {
			player.gatheringResult = spot.ActiveResult
			continue
		}
		switch spot.ActionKind {
		case timedActionGathering:
			w.resolveRenewableGathering(player, spot)
		case timedActionProduction:
			w.resolveProductionAction(player, spot)
		case timedActionWorldInteraction:
			w.resolveWorldInteraction(player, spot)
		default:
			clearGathering(player, "action_kind_missing")
		}
	}
}

func (w *World) rechargeGatheringSpots() {
	for id, spot := range w.gatheringSpots {
		if spot.MaximumAmount <= 0 || spot.CurrentAmount >= spot.MaximumAmount || spot.RechargePulses == 0 || spot.NextRecharge == 0 {
			continue
		}
		for spot.CurrentAmount < spot.MaximumAmount && w.combat.pulse >= spot.NextRecharge {
			spot.CurrentAmount++
			spot.NextRecharge += spot.RechargePulses
		}
		if spot.CurrentAmount >= spot.MaximumAmount {
			spot.NextRecharge = 0
		}
		w.gatheringSpots[id] = spot
	}
}

func cloneInventory(inventory inventoryState) inventoryState {
	return inventoryState{
		capacity: inventory.capacity,
		items:    append([]InventoryItem(nil), inventory.items...),
	}
}

func clearGathering(player *playerState, result string) {
	if player == nil {
		return
	}
	player.gatheringSpot = ""
	player.gatheringRecipe = ""
	player.gatheringInstanceID = ""
	player.productionBatch = productionBatch{}
	player.gatheringInteractionPosition = Position{}
	player.gatheringCompletionPosition = Position{}
	player.gatheringTraversal = false
	player.nextGatherPulse = 0
	player.gatheringResult = result
}

func (w *World) gatheringSpotSnapshots() []GatheringSpotSnapshot {
	spots := make([]GatheringSpotSnapshot, 0, len(w.gatheringSpots))
	for _, spot := range w.gatheringSpots {
		spots = append(spots, spot.GatheringSpotSnapshot)
	}
	sort.Slice(spots, func(i, j int) bool { return spots[i].ID < spots[j].ID })
	return spots
}
