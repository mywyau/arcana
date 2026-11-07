package simulation

import (
	"testing"
	"time"
)

func TestCookedMinnowHealsAndConsumesOneWithoutDelayingAttack(t *testing.T) {
	world := NewWorld()
	if err := world.AddPlayer("player-one", Position{}, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	player := world.players["player-one"]
	player.health = 90
	player.nextAttackPulse = 9
	if !player.inventory.add(ItemCookedMinnow, 3) {
		t.Fatal("could not add cooked minnows")
	}
	if err := world.QueueConsumeItem(ConsumeItemCommand{
		PlayerID: "player-one", Sequence: 1, ItemID: ItemCookedMinnow,
	}); err != nil {
		t.Fatalf("QueueConsumeItem() returned an unexpected error: %v", err)
	}

	world.Step(50 * time.Millisecond)
	snapshot := mustPlayer(t, world, "player-one")
	if snapshot.Health != 98 || snapshot.Inventory.itemsQuantity(ItemCookedMinnow) != 2 {
		t.Fatalf("player after eating = %+v", snapshot)
	}
	if snapshot.LastConsumeSequence != 1 || snapshot.NextConsumePulse != 2 || snapshot.NextAttackPulse != 9 {
		t.Fatalf("consume cadence after eating = %+v", snapshot)
	}
}

func TestCookedMinnowCannotBeEatenAtFullHealthOrDuringCooldown(t *testing.T) {
	world := NewWorld()
	if err := world.AddPlayer("player-one", Position{}, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	player := world.players["player-one"]
	if !player.inventory.add(ItemCookedMinnow, 2) {
		t.Fatal("could not add cooked minnows")
	}
	if err := world.QueueConsumeItem(ConsumeItemCommand{
		PlayerID: "player-one", Sequence: 1, ItemID: ItemCookedMinnow,
	}); err == nil {
		t.Fatal("QueueConsumeItem() error = nil at full health")
	}

	player.health = 50
	if err := world.QueueConsumeItem(ConsumeItemCommand{
		PlayerID: "player-one", Sequence: 1, ItemID: ItemCookedMinnow,
	}); err != nil {
		t.Fatalf("QueueConsumeItem() returned an unexpected error: %v", err)
	}
	world.Step(50 * time.Millisecond)
	if err := world.QueueConsumeItem(ConsumeItemCommand{
		PlayerID: "player-one", Sequence: 2, ItemID: ItemCookedMinnow,
	}); err == nil {
		t.Fatal("QueueConsumeItem() error = nil during cooldown")
	}
}
