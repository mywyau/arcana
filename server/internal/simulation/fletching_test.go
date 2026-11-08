package simulation

import (
	"testing"
	"time"
)

func TestFletchingShaftsProducesTwentyAndGrantsXP(t *testing.T) {
	world := NewWorld()
	if err := world.AddPlayer("player-one", Position{}, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	player := world.players["player-one"]
	player.inventory.add(ItemKnife, 1)
	player.inventory.add(ItemNormalLogs, 1)
	if err := world.QueueFletch(FletchCommand{
		PlayerID: "player-one", Sequence: 1, RecipeID: FletchArrowShafts, Quantity: 1,
	}); err != nil {
		t.Fatalf("QueueFletch() returned an unexpected error: %v", err)
	}
	world.Step(600 * time.Millisecond)
	snapshot := mustPlayer(t, world, "player-one")
	if snapshot.Inventory.itemsQuantity(ItemArrowShafts) != 20 ||
		snapshot.Inventory.itemsQuantity(ItemKnife) != 1 ||
		mustSkill(t, snapshot, SkillFletching).Experience != 10 {
		t.Fatalf("player after fletching shafts = %+v", snapshot)
	}
}

func TestFletchingIronArrowsConsumesBundlesOfTwenty(t *testing.T) {
	world := NewWorld()
	if err := world.AddPlayer("player-one", Position{}, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	player := world.players["player-one"]
	player.inventory.add(ItemArrowShafts, 40)
	player.inventory.add(ItemFeathers, 40)
	player.inventory.add(ItemIronArrowTips, 40)
	if err := world.QueueFletch(FletchCommand{
		PlayerID: "player-one", Sequence: 1, RecipeID: FletchIronArrows, Quantity: 0,
	}); err != nil {
		t.Fatalf("QueueFletch(make all) returned an unexpected error: %v", err)
	}
	world.Step(600 * time.Millisecond)
	first := mustPlayer(t, world, "player-one")
	if first.Inventory.itemsQuantity(ItemIronArrow) != 20 || first.FletchRemaining != 1 {
		t.Fatalf("first arrow batch = %+v", first)
	}
	world.Step(600 * time.Millisecond)
	second := mustPlayer(t, world, "player-one")
	if second.Inventory.itemsQuantity(ItemIronArrow) != 40 ||
		mustSkill(t, second, SkillFletching).Experience != 40 || second.FletchRemaining != 0 {
		t.Fatalf("completed arrow batches = %+v", second)
	}
}

func TestFletchingExperienceRaisesFletchingLevel(t *testing.T) {
	world := NewWorld()
	if err := world.AddPlayer("player-one", Position{}, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	player := world.players["player-one"]
	player.inventory.add(ItemKnife, 1)
	player.inventory.add(ItemNormalLogs, 10)
	if err := world.QueueFletch(FletchCommand{
		PlayerID: "player-one", Sequence: 1, RecipeID: FletchArrowShafts, Quantity: 0,
	}); err != nil {
		t.Fatalf("QueueFletch(make all) returned an unexpected error: %v", err)
	}
	world.Step(10 * actionPulseInterval)
	snapshot := mustPlayer(t, world, "player-one")
	fletching := mustSkill(t, snapshot, SkillFletching)
	if fletching.Experience != 100 || fletching.Level != 2 {
		t.Fatalf("Fletching progress = %+v, want 100 XP and level 2", fletching)
	}
}

func TestBasicFletchingChainProducesUsableIronArrows(t *testing.T) {
	world := NewWorld()
	if err := world.AddPlayer("player-one", Position{}, 8); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	player := world.players["player-one"]
	player.skills[SkillSmithing] = SkillExperienceForLevel(2)
	player.inventory.add(ItemKnife, 1)
	player.inventory.add(ItemNormalLogs, 1)
	player.inventory.add(ItemIronBar, 1)
	player.inventory.add(ItemFeathers, 20)

	tipRecipe, _ := forgingRecipe(ForgingIronArrowTips)
	if err := applyForgingRecipe(player, tipRecipe); err != nil {
		t.Fatalf("forge arrow tips: %v", err)
	}
	if err := world.QueueFletch(FletchCommand{
		PlayerID: "player-one", Sequence: 1, RecipeID: FletchArrowShafts, Quantity: 1,
	}); err != nil {
		t.Fatalf("fletch arrow shafts: %v", err)
	}
	world.Step(actionPulseInterval)
	if err := world.QueueFletch(FletchCommand{
		PlayerID: "player-one", Sequence: 2, RecipeID: FletchIronArrows, Quantity: 1,
	}); err != nil {
		t.Fatalf("assemble iron arrows: %v", err)
	}
	world.Step(actionPulseInterval)

	snapshot := mustPlayer(t, world, "player-one")
	if snapshot.Inventory.itemsQuantity(ItemIronArrow) != 20 ||
		snapshot.Inventory.itemsQuantity(ItemNormalLogs) != 0 ||
		snapshot.Inventory.itemsQuantity(ItemIronBar) != 0 ||
		snapshot.Inventory.itemsQuantity(ItemFeathers) != 0 ||
		snapshot.Inventory.itemsQuantity(ItemIronArrowTips) != 0 ||
		snapshot.Inventory.itemsQuantity(ItemArrowShafts) != 0 {
		t.Fatalf("inventory after complete Fletching chain = %+v", snapshot.Inventory)
	}
	if mustSkill(t, snapshot, SkillFletching).Experience != 30 {
		t.Fatalf("Fletching XP after complete chain = %+v", mustSkill(t, snapshot, SkillFletching))
	}
}
