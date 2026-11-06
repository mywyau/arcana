package simulation

import (
	"testing"
	"time"
)

func TestWorldSteelSwordCanBePickedUpAndEquipped(t *testing.T) {
	world := NewWorld()
	if err := world.AddWorldDrop(Position{X: -2, Z: -5}, string(WeaponSteelSword), 1); err != nil {
		t.Fatalf("AddWorldDrop() returned an unexpected error: %v", err)
	}
	drops := world.Snapshot().Drops
	if len(drops) != 1 || drops[0].ItemID != string(WeaponSteelSword) || drops[0].Position != (Position{X: -2, Z: -5}) {
		t.Fatalf("initial steel sword drop = %+v", drops)
	}

	mustAddPlayer(t, world, "player-one", Position{X: -2, Z: -5}, 4)
	if err := world.QueuePickup(PickupCommand{PlayerID: "player-one", Sequence: 1, DropID: drops[0].ID}); err != nil {
		t.Fatalf("QueuePickup() returned an unexpected error: %v", err)
	}
	world.Step(50 * time.Millisecond)
	if !snapshotInventoryContains(mustPlayer(t, world, "player-one").Inventory, string(WeaponSteelSword)) {
		t.Fatal("picked-up steel sword is missing from inventory")
	}
	if err := world.QueueEquip(EquipCommand{PlayerID: "player-one", Sequence: 1, ItemID: WeaponSteelSword}); err != nil {
		t.Fatalf("QueueEquip() returned an unexpected error: %v", err)
	}
	world.Step(actionPulseInterval)
	player := mustPlayer(t, world, "player-one")
	if player.EquippedWeapon != WeaponSteelSword || player.CombatStyle != CombatStyleStab {
		t.Fatalf("equipped steel sword state = %+v", player)
	}
	if player.CombatStats.StrengthBonus != steelSwordMeleeStrength {
		t.Fatalf("steel sword strength = %d, want %d", player.CombatStats.StrengthBonus, steelSwordMeleeStrength)
	}
}

func TestPickupWithinRangeAddsShardAndRemovesDrop(t *testing.T) {
	world := NewWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	world.spawnDrop(Position{X: pickupRange}, trainingShardItem, 1)

	if err := world.QueuePickup(PickupCommand{PlayerID: "player-one", Sequence: 1, DropID: "drop-1"}); err != nil {
		t.Fatalf("QueuePickup() returned an unexpected error: %v", err)
	}
	world.Step(50 * time.Millisecond)

	if got := len(world.Snapshot().Drops); got != 0 {
		t.Fatalf("drop count after pickup = %d, want 0", got)
	}
	player := mustPlayer(t, world, "player-one")
	if got, want := len(player.Inventory.Items), 8; got != want {
		t.Fatalf("inventory item count = %d, want %d", got, want)
	}
	shard := player.Inventory.Items[7]
	if shard.Slot != 7 || shard.ItemID != trainingShardItem || shard.Quantity != 1 {
		t.Fatalf("picked up shard = %+v", shard)
	}
}

func TestPickupOutsideRangeApproachesThenCollects(t *testing.T) {
	world := NewWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	world.spawnDrop(Position{X: pickupRange + 1}, trainingShardItem, 1)

	if err := world.QueuePickup(PickupCommand{PlayerID: "player-one", Sequence: 1, DropID: "drop-1"}); err != nil {
		t.Fatalf("QueuePickup() returned an unexpected error: %v", err)
	}
	world.Step(50 * time.Millisecond)
	if player := mustPlayer(t, world, "player-one"); !player.AutoPickingUp || player.Destination == nil {
		t.Fatalf("player did not begin pickup approach: %+v", player)
	}

	for range 10 {
		world.Step(50 * time.Millisecond)
	}
	if got := len(world.Snapshot().Drops); got != 0 {
		t.Fatalf("drop count after pickup approach = %d, want 0", got)
	}
	if player := mustPlayer(t, world, "player-one"); player.AutoPickingUp || player.Destination != nil {
		t.Fatalf("player continued moving after pickup: %+v", player)
	}
}

func TestMoveCancelsPickupApproach(t *testing.T) {
	world := NewWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	world.spawnDrop(Position{X: 8}, trainingShardItem, 1)
	if err := world.QueuePickup(PickupCommand{PlayerID: "player-one", Sequence: 1, DropID: "drop-1"}); err != nil {
		t.Fatalf("QueuePickup() returned an unexpected error: %v", err)
	}
	world.Step(50 * time.Millisecond)
	mustQueueMove(t, world, "player-one", 1, Position{Z: 2})
	world.Step(50 * time.Millisecond)

	player := mustPlayer(t, world, "player-one")
	if player.AutoPickingUp {
		t.Fatal("player remains in pickup approach after a movement command")
	}
	if player.Destination == nil || *player.Destination != (Position{Z: 2}) {
		t.Fatalf("movement destination after cancel = %+v, want z=2", player.Destination)
	}
}

func TestCompetingPickupAwardsDropOnlyOnce(t *testing.T) {
	world := NewWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	mustAddPlayer(t, world, "player-two", Position{}, 4)
	world.spawnDrop(Position{}, trainingShardItem, 1)

	for _, id := range []PlayerID{"player-two", "player-one"} {
		if err := world.QueuePickup(PickupCommand{PlayerID: id, Sequence: 1, DropID: "drop-1"}); err != nil {
			t.Fatalf("QueuePickup(%q) returned an unexpected error: %v", id, err)
		}
	}
	world.Step(50 * time.Millisecond)

	if !snapshotInventoryContains(mustPlayer(t, world, "player-one").Inventory, trainingShardItem) {
		t.Fatal("deterministic first player did not receive contested drop")
	}
	if snapshotInventoryContains(mustPlayer(t, world, "player-two").Inventory, trainingShardItem) {
		t.Fatal("second player also received contested drop")
	}
}

func TestDropUnequippedEquipmentCreatesPickup(t *testing.T) {
	world := NewWorld()
	mustAddPlayer(t, world, "player-one", Position{X: 2, Z: 3}, 4)

	if err := world.QueueDropItem(DropItemCommand{
		PlayerID: "player-one", Sequence: 1, ItemID: string(WeaponDagger),
	}); err != nil {
		t.Fatalf("QueueDropItem() returned an unexpected error: %v", err)
	}
	world.Step(50 * time.Millisecond)

	player := mustPlayer(t, world, "player-one")
	if snapshotInventoryContains(player.Inventory, string(WeaponDagger)) {
		t.Fatal("dropped dagger remains in inventory")
	}
	drops := world.Snapshot().Drops
	if len(drops) != 1 || drops[0].ItemID != string(WeaponDagger) || drops[0].Position != player.Position {
		t.Fatalf("equipment drop = %+v", drops)
	}
}

func TestDropRejectsEquippedItem(t *testing.T) {
	world := NewWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	if err := world.QueueDropItem(DropItemCommand{
		PlayerID: "player-one", Sequence: 1, ItemID: string(WeaponSword),
	}); err == nil {
		t.Fatal("QueueDropItem() error = nil for equipped sword")
	}
}

func snapshotInventoryContains(inventory InventorySnapshot, itemID string) bool {
	for _, item := range inventory.Items {
		if item.ItemID == itemID && item.Quantity > 0 {
			return true
		}
	}
	return false
}

func TestDefeatedTargetSpawnsTrainingShard(t *testing.T) {
	world := newCombatTestWorld()
	mustAddPlayer(t, world, "player-one", Position{X: -4, Z: -1}, 4)
	world.combat.targets[dummyTargetID].Health = swordDamage
	if err := world.QueueTarget(TargetCommand{PlayerID: "player-one", Sequence: 1, TargetID: dummyTargetID}); err != nil {
		t.Fatalf("QueueTarget() returned an unexpected error: %v", err)
	}
	world.Step(actionPulseInterval)

	drops := world.Snapshot().Drops
	if len(drops) != 1 || drops[0].ItemID != trainingShardItem || drops[0].Quantity != 1 {
		t.Fatalf("target drops = %+v", drops)
	}
	if drops[0].Position != world.combat.targets[dummyTargetID].Position {
		t.Fatalf("drop position = %+v, want target position %+v", drops[0].Position, world.combat.targets[dummyTargetID].Position)
	}
}

func TestArmourTargetsSpawnTrainingArmour(t *testing.T) {
	tests := []struct {
		targetID string
		itemID   string
	}{
		{targetID: upperDummyTargetID, itemID: ItemTrainingHelm},
		{targetID: hillWardenTargetID, itemID: ItemTrainingPlate},
	}
	for _, test := range tests {
		t.Run(test.itemID, func(t *testing.T) {
			world := newCombatTestWorld()
			target := world.combat.targets[test.targetID]
			world.defeatDummy(test.targetID, target)
			drops := world.Snapshot().Drops
			if len(drops) != 1 || drops[0].ItemID != test.itemID {
				t.Fatalf("target drops = %+v, want %q", drops, test.itemID)
			}
		})
	}
}
