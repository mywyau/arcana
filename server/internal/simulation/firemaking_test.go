package simulation

import (
	"testing"
	"time"
)

func TestLightFireConsumesOneLogAndPreservesTinderbox(t *testing.T) {
	world := NewWorld()
	if err := world.AddPlayer("player-one", Position{X: 4, Z: 3}, 4); err != nil {
		t.Fatal(err)
	}
	player := world.players["player-one"]
	player.inventory.add(ItemTinderbox, 1)
	player.inventory.add(ItemOakLogs, 2)
	if err := world.LightFire("player-one", 1, ItemOakLogs); err != nil {
		t.Fatalf("LightFire() returned an unexpected error: %v", err)
	}
	if player.inventory.quantity(ItemTinderbox) != 1 || player.inventory.quantity(ItemOakLogs) != 1 {
		t.Fatalf("inventory after firemaking = %+v", player.inventory)
	}
	fires := world.Snapshot().Fires
	if len(fires) != 1 || fires[0].LogItem != ItemOakLogs || fires[0].Position != player.position {
		t.Fatalf("fires after firemaking = %+v", fires)
	}
}

func TestLightFireRequiresTinderbox(t *testing.T) {
	world := NewWorld()
	if err := world.AddPlayer("player-one", Position{}, 4); err != nil {
		t.Fatal(err)
	}
	world.players["player-one"].inventory.add(ItemNormalLogs, 1)
	if err := world.LightFire("player-one", 1, ItemNormalLogs); err == nil {
		t.Fatal("LightFire() error = nil without a tinderbox")
	}
}

func TestFireExpiresAfterOneMinute(t *testing.T) {
	world := NewWorld()
	if err := world.AddPlayer("player-one", Position{}, 4); err != nil {
		t.Fatal(err)
	}
	player := world.players["player-one"]
	player.inventory.add(ItemTinderbox, 1)
	player.inventory.add(ItemNormalLogs, 1)
	if err := world.LightFire("player-one", 1, ItemNormalLogs); err != nil {
		t.Fatal(err)
	}
	world.Step(basicFireLifetime - time.Millisecond)
	if len(world.Snapshot().Fires) != 1 {
		t.Fatal("fire expired too early")
	}
	world.Step(time.Millisecond)
	if len(world.Snapshot().Fires) != 0 {
		t.Fatal("fire did not expire")
	}
}
