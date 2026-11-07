package simulation

import (
	"testing"
	"time"
)

func TestStarterFishingSpotAutomaticallyCatchesRawMinnowsAndGrantsXP(t *testing.T) {
	world := NewWorld()
	spot := world.gatheringSpots[starterFishingSpotID]
	if err := world.AddPlayer("player-one", spot.Position, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	if err := world.QueueGather(GatherCommand{
		PlayerID: "player-one", Sequence: 1, SpotID: starterFishingSpotID,
	}); err != nil {
		t.Fatalf("QueueGather() returned an unexpected error: %v", err)
	}

	world.Step(50 * time.Millisecond)
	started := mustPlayer(t, world, "player-one")
	if started.GatheringSpot != starterFishingSpotID || started.GatheringResult != "fishing" {
		t.Fatalf("started gathering state = %+v", started)
	}
	world.Step(fishingInterval * actionPulseInterval)
	player := mustPlayer(t, world, "player-one")
	if got := player.Inventory.itemsQuantity(ItemRawMinnow); got != 1 {
		t.Fatalf("raw minnow quantity = %d, want 1", got)
	}
	if fishing := mustSkill(t, player, SkillFishing); fishing.Experience != fishingExperience {
		t.Fatalf("Fishing after first catch = %+v", fishing)
	}
	if player.GatheringResult != "caught_raw_minnow" || player.NextGatherPulse != 8 {
		t.Fatalf("gathering state after first catch = %+v", player)
	}

	world.Step(fishingInterval * actionPulseInterval)
	player = mustPlayer(t, world, "player-one")
	if player.Inventory.itemsQuantity(ItemRawMinnow) != 2 ||
		mustSkill(t, player, SkillFishing).Experience != 2*fishingExperience {
		t.Fatalf("player after repeated fishing = %+v", player)
	}
}

func TestMovementCancelsFishing(t *testing.T) {
	world := NewWorld()
	spot := world.gatheringSpots[starterFishingSpotID]
	if err := world.AddPlayer("player-one", spot.Position, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	if err := world.QueueGather(GatherCommand{
		PlayerID: "player-one", Sequence: 1, SpotID: starterFishingSpotID,
	}); err != nil {
		t.Fatalf("QueueGather() returned an unexpected error: %v", err)
	}
	world.Step(50 * time.Millisecond)
	if err := world.QueueMove(MoveCommand{
		PlayerID: "player-one", Sequence: 1, Destination: Position{X: -10, Z: 10},
	}); err != nil {
		t.Fatalf("QueueMove() returned an unexpected error: %v", err)
	}
	world.Step(50 * time.Millisecond)
	if player := mustPlayer(t, world, "player-one"); player.GatheringSpot != "" {
		t.Fatalf("gathering continued after movement: %+v", player)
	}
}

func TestFishingRejectsFullInventory(t *testing.T) {
	world := NewWorld()
	spot := world.gatheringSpots[starterFishingSpotID]
	if err := world.AddPlayer("player-one", spot.Position, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	world.players["player-one"].inventory = inventoryState{
		capacity: 1,
		items:    []InventoryItem{{Slot: 0, ItemID: ItemRawMinnow, Quantity: 10}},
	}
	if err := world.QueueGather(GatherCommand{
		PlayerID: "player-one", Sequence: 1, SpotID: starterFishingSpotID,
	}); err == nil {
		t.Fatal("QueueGather() error = nil with a full inventory")
	}
}

func TestCookingRangeConsumesRawMinnowsAndGrantsCookingXP(t *testing.T) {
	world := NewWorld()
	spot := world.gatheringSpots[starterCookingRangeID]
	if err := world.AddPlayer("player-one", spot.Position, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	if !world.players["player-one"].inventory.add(ItemRawMinnow, 2) {
		t.Fatal("failed to add Raw Minnows to test inventory")
	}
	if err := world.QueueGather(GatherCommand{
		PlayerID: "player-one", Sequence: 1, SpotID: starterCookingRangeID,
	}); err != nil {
		t.Fatalf("QueueGather() returned an unexpected error: %v", err)
	}

	world.Step(50 * time.Millisecond)
	started := mustPlayer(t, world, "player-one")
	if started.GatheringSpot != starterCookingRangeID || started.GatheringResult != "cooking" {
		t.Fatalf("started cooking state = %+v", started)
	}
	world.Step(cookingInterval * actionPulseInterval)
	player := mustPlayer(t, world, "player-one")
	if player.Inventory.itemsQuantity(ItemRawMinnow) != 1 ||
		player.Inventory.itemsQuantity(ItemCookedMinnow) != 1 {
		t.Fatalf("inventory after first cook = %+v", player.Inventory)
	}
	if cooking := mustSkill(t, player, SkillCooking); cooking.Experience != cookingExperience {
		t.Fatalf("Cooking after first cook = %+v", cooking)
	}
	if player.GatheringResult != "cooked_minnow" || player.NextGatherPulse != 8 {
		t.Fatalf("gathering state after first cook = %+v", player)
	}

	world.Step(cookingInterval * actionPulseInterval)
	player = mustPlayer(t, world, "player-one")
	if player.Inventory.itemsQuantity(ItemRawMinnow) != 0 ||
		player.Inventory.itemsQuantity(ItemCookedMinnow) != 2 ||
		mustSkill(t, player, SkillCooking).Experience != 2*cookingExperience {
		t.Fatalf("player after repeated cooking = %+v", player)
	}
	world.Step(cookingInterval * actionPulseInterval)
	player = mustPlayer(t, world, "player-one")
	if player.GatheringSpot != "" || player.GatheringResult != "missing_raw_minnow" {
		t.Fatalf("cooking did not stop after inputs ran out: %+v", player)
	}
}

func TestCookingRangeRequiresRawMinnow(t *testing.T) {
	world := NewWorld()
	spot := world.gatheringSpots[starterCookingRangeID]
	if err := world.AddPlayer("player-one", spot.Position, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	if err := world.QueueGather(GatherCommand{
		PlayerID: "player-one", Sequence: 1, SpotID: starterCookingRangeID,
	}); err == nil {
		t.Fatal("QueueGather() error = nil without a Raw Minnow")
	}
}

func TestCookingCanReplaceLastRawMinnowInFullInventory(t *testing.T) {
	world := NewWorld()
	spot := world.gatheringSpots[starterCookingRangeID]
	if err := world.AddPlayer("player-one", spot.Position, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	world.players["player-one"].inventory = inventoryState{
		capacity: 1,
		items:    []InventoryItem{{Slot: 0, ItemID: ItemRawMinnow, Quantity: 1}},
	}
	if err := world.QueueGather(GatherCommand{
		PlayerID: "player-one", Sequence: 1, SpotID: starterCookingRangeID,
	}); err != nil {
		t.Fatalf("QueueGather() rejected an atomic input replacement: %v", err)
	}

	world.Step(50 * time.Millisecond)
	world.Step(cookingInterval * actionPulseInterval)
	player := mustPlayer(t, world, "player-one")
	if player.Inventory.itemsQuantity(ItemRawMinnow) != 0 ||
		player.Inventory.itemsQuantity(ItemCookedMinnow) != 1 {
		t.Fatalf("inventory after cooking in the full slot = %+v", player.Inventory)
	}
}
