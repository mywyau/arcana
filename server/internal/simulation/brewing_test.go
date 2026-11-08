package simulation

import (
	"testing"
	"time"
)

func TestKnownThreeItemBrewCreatesPrayerPotionAndGrantsHerbloreXP(t *testing.T) {
	world := NewWorld()
	if err := world.AddPlayer("player-one", Position{}, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	player := world.players["player-one"]
	for itemID, quantity := range map[string]int{ItemVialOfWater: 2, ItemSunleaf: 2, ItemSacredDust: 1} {
		if !player.inventory.add(itemID, quantity) {
			t.Fatalf("could not add %s", itemID)
		}
	}
	if err := world.QueueBrew(BrewCommand{
		PlayerID: "player-one", Sequence: 1,
		ItemIDs: []string{ItemSacredDust, ItemVialOfWater, ItemSunleaf},
	}); err != nil {
		t.Fatalf("QueueBrew() returned an unexpected error: %v", err)
	}
	world.Step(actionPulseInterval)
	snapshot := mustPlayer(t, world, "player-one")
	if snapshot.Inventory.itemsQuantity(ItemPrayerPotion) != 1 ||
		snapshot.Inventory.itemsQuantity(ItemVialOfWater) != 1 ||
		mustSkill(t, snapshot, SkillHerblore).Experience != 20 {
		t.Fatalf("player after brewing = %+v", snapshot)
	}
}

func TestMakeAllBrewsOnePotionPerSixHundredMillisecondPulse(t *testing.T) {
	world := NewWorld()
	if err := world.AddPlayer("player-one", Position{}, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	player := world.players["player-one"]
	player.inventory.add(ItemVialOfWater, 3)
	player.inventory.add(ItemSunleaf, 3)
	if err := world.QueueBrew(BrewCommand{
		PlayerID: "player-one", Sequence: 1, ItemIDs: []string{ItemVialOfWater, ItemSunleaf}, Quantity: 0,
	}); err != nil {
		t.Fatalf("QueueBrew(make all) returned an unexpected error: %v", err)
	}

	for made := 1; made <= 3; made++ {
		world.Step(600 * time.Millisecond)
		snapshot := mustPlayer(t, world, "player-one")
		if got := snapshot.Inventory.itemsQuantity(ItemMinorHealingPotion); got != made {
			t.Fatalf("potions after %d pulse(s) = %d, want %d", made, got, made)
		}
		if snapshot.BrewRemaining != 3-made {
			t.Fatalf("remaining after %d pulse(s) = %d, want %d", made, snapshot.BrewRemaining, 3-made)
		}
	}
}

func TestUnknownValidCombinationCreatesMurkyBrew(t *testing.T) {
	inventory := inventoryState{capacity: 3}
	inventory.add(ItemVialOfWater, 1)
	inventory.add(ItemMoonleaf, 1)
	inventory.add(ItemFishOil, 1)
	result, recipe, err := brewedInventoryAfter(inventory, []string{ItemVialOfWater, ItemMoonleaf, ItemFishOil})
	if err != nil {
		t.Fatalf("brewedInventoryAfter() returned an unexpected error: %v", err)
	}
	if result.quantity(ItemMurkyBrew) != 1 || recipe.Experience != 5 {
		t.Fatalf("unknown brew result = %+v, recipe %+v", result, recipe)
	}
}

func TestBrewRejectsInvalidRolesWithoutConsumingIngredients(t *testing.T) {
	inventory := inventoryState{capacity: 3}
	inventory.add(ItemVialOfWater, 1)
	inventory.add(ItemSacredDust, 1)
	if _, _, err := brewedInventoryAfter(inventory, []string{ItemVialOfWater, ItemSacredDust}); err == nil {
		t.Fatal("brewedInventoryAfter() error = nil without an herb")
	}
	if inventory.quantity(ItemVialOfWater) != 1 || inventory.quantity(ItemSacredDust) != 1 {
		t.Fatalf("failed brew mutated inventory: %+v", inventory)
	}
}
