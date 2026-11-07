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

func TestCopperVeinRequiresPickaxeAndConsumesSharedCapacity(t *testing.T) {
	world := NewWorld()
	spot := world.gatheringSpots[starterCopperVeinID]
	if err := world.AddPlayer("player-one", spot.Position, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	command := GatherCommand{PlayerID: "player-one", Sequence: 1, SpotID: starterCopperVeinID}
	if err := world.QueueGather(command); err == nil {
		t.Fatal("QueueGather() error = nil without a Bronze Pickaxe")
	}
	if !world.players["player-one"].inventory.add(ItemBronzePickaxe, 1) {
		t.Fatal("failed to add Bronze Pickaxe")
	}
	if err := world.QueueGather(command); err != nil {
		t.Fatalf("QueueGather() returned an unexpected error: %v", err)
	}
	world.Step(50 * time.Millisecond)
	world.Step(miningInterval * actionPulseInterval)
	player := mustPlayer(t, world, "player-one")
	if player.Inventory.itemsQuantity(ItemCopperOre) != 1 || mustSkill(t, player, SkillMining).Experience != miningExperience {
		t.Fatalf("player after mining = %+v", player)
	}
	if got := world.gatheringSpots[starterCopperVeinID].CurrentAmount; got != copperVeinCapacity-1 {
		t.Fatalf("vein amount = %d, want %d", got, copperVeinCapacity-1)
	}
}

func TestCopperVeinDepletesAndGraduallyRecharges(t *testing.T) {
	world := NewWorld()
	spot := world.gatheringSpots[starterCopperVeinID]
	spot.CurrentAmount = 1
	world.gatheringSpots[starterCopperVeinID] = spot
	if err := world.AddPlayer("player-one", spot.Position, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	world.players["player-one"].inventory.add(ItemBronzePickaxe, 1)
	if err := world.QueueGather(GatherCommand{PlayerID: "player-one", Sequence: 1, SpotID: starterCopperVeinID}); err != nil {
		t.Fatalf("QueueGather() returned an unexpected error: %v", err)
	}
	world.Step(50 * time.Millisecond)
	world.Step(miningInterval * actionPulseInterval)
	if got := world.gatheringSpots[starterCopperVeinID].CurrentAmount; got != 0 {
		t.Fatalf("depleted vein amount = %d, want 0", got)
	}
	world.Step(copperRechargePulses * actionPulseInterval)
	if got := world.gatheringSpots[starterCopperVeinID].CurrentAmount; got != 1 {
		t.Fatalf("recharged vein amount = %d, want 1", got)
	}
}

func TestMiningVeinProgressionDefinitions(t *testing.T) {
	tests := []struct {
		id       string
		level    int
		reward   string
		capacity int
	}{
		{starterCopperVeinID, 1, ItemCopperOre, copperVeinCapacity},
		{starterTinVeinID, 1, ItemTinOre, tinVeinCapacity},
		{starterIronVeinID, 2, ItemIronOre, ironVeinCapacity},
		{starterCoalVeinID, 3, ItemCoal, coalVeinCapacity},
		{starterMithrilVeinID, 5, ItemMithrilOre, mithrilVeinCapacity},
	}
	spots := newGatheringSpots()
	for _, test := range tests {
		spot := spots[test.id]
		if spot.RequiredLevel != test.level || spot.RewardItem != test.reward || spot.CurrentAmount != test.capacity || spot.MaximumAmount != test.capacity {
			t.Errorf("mining vein %q = %+v", test.id, spot)
		}
	}
}

func TestFurnaceRepeatedlySmeltsSelectedRecipe(t *testing.T) {
	world := NewWorld()
	spot := world.gatheringSpots[starterFurnaceID]
	if err := world.AddPlayer("player-one", spot.Position, 6); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	player := world.players["player-one"]
	player.inventory.add(ItemCopperOre, 2)
	player.inventory.add(ItemTinOre, 2)
	recipe, _ := smeltingRecipe(SmeltingBronzeBar)
	if err := world.QueueGather(GatherCommand{
		PlayerID: "player-one", Sequence: 1, SpotID: starterFurnaceID, RecipeID: string(SmeltingBronzeBar),
	}); err != nil {
		t.Fatalf("QueueGather() returned an unexpected error: %v", err)
	}

	world.Step(50 * time.Millisecond)
	started := mustPlayer(t, world, "player-one")
	if started.GatheringSpot != starterFurnaceID || started.GatheringResult != "smelting" {
		t.Fatalf("started smelting state = %+v", started)
	}
	world.Step(time.Duration(recipe.Interval) * actionPulseInterval)
	playerSnapshot := mustPlayer(t, world, "player-one")
	if playerSnapshot.Inventory.itemsQuantity(ItemBronzeBar) != 1 ||
		playerSnapshot.Inventory.itemsQuantity(ItemCopperOre) != 1 ||
		playerSnapshot.Inventory.itemsQuantity(ItemTinOre) != 1 {
		t.Fatalf("inventory after first smelt = %+v", playerSnapshot.Inventory)
	}
	if smithing := mustSkill(t, playerSnapshot, SkillSmithing); smithing.Experience != recipe.RewardXP {
		t.Fatalf("Smithing after first smelt = %+v", smithing)
	}
	if playerSnapshot.GatheringResult != "smelted_bronze_bar" {
		t.Fatalf("gathering result = %q", playerSnapshot.GatheringResult)
	}

	world.Step(time.Duration(recipe.Interval) * actionPulseInterval)
	world.Step(time.Duration(recipe.Interval) * actionPulseInterval)
	playerSnapshot = mustPlayer(t, world, "player-one")
	if playerSnapshot.Inventory.itemsQuantity(ItemBronzeBar) != 2 || playerSnapshot.GatheringSpot != "" ||
		playerSnapshot.GatheringResult != "missing_smelting_inputs" {
		t.Fatalf("player after exhausting smelting inputs = %+v", playerSnapshot)
	}
}

func TestFurnaceRequiresAValidAffordableRecipe(t *testing.T) {
	world := NewWorld()
	spot := world.gatheringSpots[starterFurnaceID]
	if err := world.AddPlayer("player-one", spot.Position, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	if err := world.QueueGather(GatherCommand{
		PlayerID: "player-one", Sequence: 1, SpotID: starterFurnaceID,
	}); err == nil {
		t.Fatal("QueueGather() error = nil without a recipe")
	}
	if err := world.QueueGather(GatherCommand{
		PlayerID: "player-one", Sequence: 1, SpotID: starterFurnaceID, RecipeID: string(SmeltingCopperBar),
	}); err == nil {
		t.Fatal("QueueGather() error = nil without recipe inputs")
	}
}

func TestAnvilRepeatedlyForgesSelectedRecipe(t *testing.T) {
	world := NewWorld()
	spot := world.gatheringSpots[starterAnvilID]
	if err := world.AddPlayer("player-one", spot.Position, 6); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	world.players["player-one"].inventory.add(ItemBronzeBar, 4)
	recipe, _ := forgingRecipe(ForgingBronzeSword)
	if err := world.QueueGather(GatherCommand{
		PlayerID: "player-one", Sequence: 1, SpotID: starterAnvilID, RecipeID: string(ForgingBronzeSword),
	}); err != nil {
		t.Fatalf("QueueGather() returned an unexpected error: %v", err)
	}

	world.Step(50 * time.Millisecond)
	started := mustPlayer(t, world, "player-one")
	if started.GatheringSpot != starterAnvilID || started.GatheringResult != "forging" {
		t.Fatalf("started forging state = %+v", started)
	}
	world.Step(time.Duration(recipe.Interval) * actionPulseInterval)
	player := mustPlayer(t, world, "player-one")
	if player.Inventory.itemsQuantity(string(WeaponBronzeSword)) != 1 || player.Inventory.itemsQuantity(ItemBronzeBar) != 2 ||
		mustSkill(t, player, SkillSmithing).Experience != recipe.RewardXP {
		t.Fatalf("player after first forge = %+v", player)
	}

	world.Step(time.Duration(recipe.Interval) * actionPulseInterval)
	world.Step(time.Duration(recipe.Interval) * actionPulseInterval)
	player = mustPlayer(t, world, "player-one")
	if player.Inventory.itemsQuantity(string(WeaponBronzeSword)) != 2 || player.GatheringSpot != "" ||
		player.GatheringResult != "missing_forging_inputs" {
		t.Fatalf("player after exhausting forging inputs = %+v", player)
	}
}

func TestAnvilRequiresAValidAffordableRecipe(t *testing.T) {
	world := NewWorld()
	spot := world.gatheringSpots[starterAnvilID]
	if err := world.AddPlayer("player-one", spot.Position, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	if err := world.QueueGather(GatherCommand{
		PlayerID: "player-one", Sequence: 1, SpotID: starterAnvilID,
	}); err == nil {
		t.Fatal("QueueGather() error = nil without a recipe")
	}
	if err := world.QueueGather(GatherCommand{
		PlayerID: "player-one", Sequence: 1, SpotID: starterAnvilID, RecipeID: string(ForgingBronzeHelm),
	}); err == nil {
		t.Fatal("QueueGather() error = nil without Bronze Bars")
	}
}

func TestAnvilReinforcesOneEquipmentInstanceAndStops(t *testing.T) {
	world := NewWorld()
	spot := world.gatheringSpots[starterAnvilID]
	if err := world.AddPlayer("player-one", spot.Position, 6); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	player := world.players["player-one"]
	player.inventory.add(string(WeaponBronzeSword), 1)
	player.inventory.add(ItemBronzeBar, 3)
	player.skills[SkillSmithing] = SkillExperienceForLevel(2)
	instanceID := player.inventory.items[0].InstanceID
	recipe, _ := reinforcementRecipe("reinforce_bronze_sword")
	if err := world.QueueGather(GatherCommand{
		PlayerID: "player-one", Sequence: 1, SpotID: starterAnvilID,
		RecipeID: recipe.ID, InstanceID: instanceID,
	}); err != nil {
		t.Fatalf("QueueGather() returned an unexpected error: %v", err)
	}
	world.Step(50 * time.Millisecond)
	world.Step(time.Duration(recipe.Interval) * actionPulseInterval)
	item, exists := world.players["player-one"].inventory.itemByInstance(instanceID)
	if !exists || item.Reinforcement != 1 || world.players["player-one"].inventory.quantity(ItemBronzeBar) != 0 {
		t.Fatalf("inventory after reinforcement = %+v", world.players["player-one"].inventory.items)
	}
	snapshot := mustPlayer(t, world, "player-one")
	if snapshot.GatheringSpot != "" || snapshot.GatheringResult != "reinforced_bronze_sword" {
		t.Fatalf("gathering state after reinforcement = %+v", snapshot)
	}
}
