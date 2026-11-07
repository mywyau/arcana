package simulation

import (
	"fmt"
	"sort"
)

const (
	starterFishingSpotID  = "willowbrook_pond"
	starterCookingRangeID = "willowbrook_range"
	gatheringRange        = 2.5
	fishingInterval       = 4
	fishingExperience     = 10
	cookingInterval       = 4
	cookingExperience     = 10
)

type GatheringSpotSnapshot struct {
	ID            string
	Name          string
	Skill         SkillID
	Position      Position
	Range         float64
	RequiredLevel int
}

type gatheringSpotDefinition struct {
	GatheringSpotSnapshot
	RewardItem     string
	RewardQuantity int
	RewardXP       int64
	Interval       uint64
	Costs          []ItemCost
	ActiveResult   string
	RewardResult   string
	MissingResult  string
}

type GatherCommand struct {
	PlayerID PlayerID
	Sequence uint64
	SpotID   string
}

type pendingGather struct {
	command GatherCommand
	path    []Position
}

func newGatheringSpots() map[string]gatheringSpotDefinition {
	return map[string]gatheringSpotDefinition{
		starterFishingSpotID: {
			GatheringSpotSnapshot: GatheringSpotSnapshot{
				ID: starterFishingSpotID, Name: "Willowbrook Pond", Skill: SkillFishing,
				Position: Position{X: -14, Z: 14}, Range: gatheringRange, RequiredLevel: 1,
			},
			RewardItem: ItemRawMinnow, RewardQuantity: 1, RewardXP: fishingExperience,
			Interval: fishingInterval, ActiveResult: "fishing", RewardResult: "caught_raw_minnow",
		},
		starterCookingRangeID: {
			GatheringSpotSnapshot: GatheringSpotSnapshot{
				ID: starterCookingRangeID, Name: "Willowbrook Cooking Range", Skill: SkillCooking,
				Position: Position{X: -8, Z: -8}, Range: gatheringRange, RequiredLevel: 1,
			},
			RewardItem: ItemCookedMinnow, RewardQuantity: 1, RewardXP: cookingExperience,
			Interval: cookingInterval, Costs: []ItemCost{{ItemID: ItemRawMinnow, Quantity: 1}},
			ActiveResult: "cooking", RewardResult: "cooked_minnow", MissingResult: "missing_raw_minnow",
		},
	}
}

func (w *World) QueueGather(command GatherCommand) error {
	if command.Sequence == 0 {
		return fmt.Errorf("gather sequence must be positive")
	}
	spot, exists := w.gatheringSpots[command.SpotID]
	if !exists {
		return fmt.Errorf("gathering spot %q does not exist", command.SpotID)
	}

	w.mu.Lock()
	defer w.mu.Unlock()
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
	if SkillLevelForExperience(player.skills[spot.Skill]) < spot.RequiredLevel {
		return fmt.Errorf("%s level %d is required", spot.Skill, spot.RequiredLevel)
	}
	if !canCompleteGatheringAction(player.inventory, spot) {
		if !canPayItemCosts(player.inventory, spot.Costs) {
			return fmt.Errorf("required items are missing")
		}
		return fmt.Errorf("inventory is full")
	}

	path := []Position{spot.Position}
	if distance(player.position, spot.Position) > spot.Range && w.pathfinder != nil {
		var err error
		path, err = w.pathfinder.FindPathWithinRange(player.position, spot.Position, spot.Range-0.25)
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
	w.pendingGathering[command.PlayerID] = pendingGather{command: command, path: clonePositions(path)}
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
		player.nextGatherPulse = w.combat.pulse + spot.Interval
		if distance(player.position, spot.Position) > spot.Range {
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
		if distance(player.position, spot.Position) > spot.Range {
			if player.destination == nil {
				clearGathering(player, "spot_unreachable")
			}
			continue
		}
		if player.destination != nil {
			stopPlayer(player)
		}
		if w.combat.pulse < player.nextGatherPulse {
			player.gatheringResult = spot.ActiveResult
			continue
		}
		candidate := cloneInventory(player.inventory)
		if !payItemCosts(&candidate, spot.Costs) {
			clearGathering(player, spot.MissingResult)
			continue
		}
		if !candidate.add(spot.RewardItem, spot.RewardQuantity) {
			clearGathering(player, "inventory_full")
			continue
		}
		player.inventory = candidate
		grantSkillExperience(player, spot.Skill, spot.RewardXP)
		player.gatheringResult = spot.RewardResult
		player.nextGatherPulse = w.combat.pulse + spot.Interval
	}
}

func canCompleteGatheringAction(inventory inventoryState, spot gatheringSpotDefinition) bool {
	candidate := cloneInventory(inventory)
	return payItemCosts(&candidate, spot.Costs) && candidate.add(spot.RewardItem, spot.RewardQuantity)
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
