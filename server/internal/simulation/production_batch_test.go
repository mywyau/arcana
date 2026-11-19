package simulation

import (
	"testing"
	"time"
)

func TestProductionBatchCompletion(t *testing.T) {
	limited := newProductionBatch(2)
	if limited.completeStep() {
		t.Fatal("two-step batch completed after its first step")
	}
	if !limited.completeStep() {
		t.Fatal("two-step batch did not complete after its second step")
	}

	maximum := newProductionBatch(0)
	for step := 1; step <= maximumProductionBatchSize; step++ {
		completed := maximum.completeStep()
		if completed != (step == maximumProductionBatchSize) {
			t.Fatalf("maximum batch completion after step %d = %t", step, completed)
		}
	}
}

func TestProductionBatchRejectsMoreThanTenLots(t *testing.T) {
	world := NewWorld()
	spot := world.gatheringSpots[starterCookingRangeID]
	if err := world.AddPlayer("player-one", spot.Position, 4); err != nil {
		t.Fatal(err)
	}
	world.players["player-one"].inventory.add(ItemRawMinnow, maximumProductionBatchSize+1)
	err := world.QueueGather(GatherCommand{
		PlayerID: "player-one", Sequence: 1, SpotID: starterCookingRangeID,
		RecipeID: ItemCookedMinnow, Quantity: maximumProductionBatchSize + 1,
	})
	if err == nil {
		t.Fatal("QueueGather() accepted a production batch larger than ten")
	}
}

func TestCookingFixedBatchStopsAtRequestedQuantity(t *testing.T) {
	world := NewWorldWithRollSource(fixedRollSource(0))
	spot := world.gatheringSpots[starterCookingRangeID]
	if err := world.AddPlayer("player-one", spot.Position, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	player := world.players["player-one"]
	if !player.inventory.add(ItemRawMinnow, 3) {
		t.Fatal("failed to add Raw Minnows to test inventory")
	}
	if err := world.QueueGather(GatherCommand{
		PlayerID: "player-one", Sequence: 1, SpotID: starterCookingRangeID,
		RecipeID: ItemCookedMinnow, Quantity: 2,
	}); err != nil {
		t.Fatalf("QueueGather() returned an unexpected error: %v", err)
	}

	world.Step(50 * time.Millisecond)
	world.Step(cookingInterval * actionPulseInterval)
	first := mustPlayer(t, world, "player-one")
	if first.GatheringSpot != starterCookingRangeID || first.Inventory.itemsQuantity(ItemCookedMinnow) != 1 {
		t.Fatalf("player after first cooking step = %+v", first)
	}

	world.Step(cookingInterval * actionPulseInterval)
	finished := mustPlayer(t, world, "player-one")
	if finished.GatheringSpot != "" || finished.GatheringResult != ItemCookedMinnow ||
		finished.Inventory.itemsQuantity(ItemCookedMinnow) != 2 || finished.Inventory.itemsQuantity(ItemRawMinnow) != 1 {
		t.Fatalf("player after fixed cooking batch = %+v", finished)
	}
}

func TestRunecraftingFixedBatchStopsAtRequestedQuantity(t *testing.T) {
	world := NewWorld()
	spot := world.gatheringSpots[starterAirRuneAltarID]
	if err := world.AddPlayer("player-one", spot.Position, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	if !world.players["player-one"].inventory.add(ItemBlankRune, 2) {
		t.Fatal("failed to add Runestones to test inventory")
	}
	if err := world.QueueGather(GatherCommand{
		PlayerID: "player-one", Sequence: 1, SpotID: starterAirRuneAltarID, Quantity: 1,
	}); err != nil {
		t.Fatalf("QueueGather() returned an unexpected error: %v", err)
	}

	world.Step(50 * time.Millisecond)
	world.Step(runecraftingInterval * actionPulseInterval)
	finished := mustPlayer(t, world, "player-one")
	if finished.GatheringSpot != "" || finished.GatheringResult != "crafted_air_runes" ||
		finished.Inventory.itemsQuantity(ItemBlankRune) != 1 || finished.Inventory.itemsQuantity(ItemAirRune) != 5 {
		t.Fatalf("player after fixed Runecrafting batch = %+v", finished)
	}
}
