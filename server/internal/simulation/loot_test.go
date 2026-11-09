package simulation

import (
	"testing"
	"time"
)

func TestSpawnedDropExpiresAfterTwoMinutes(t *testing.T) {
	world := NewWorld()
	world.spawnDrop(Position{}, ItemCoins, 1)

	world.Step(worldDropLifetime - time.Millisecond)
	if got := len(world.Snapshot().Drops); got != 1 {
		t.Fatalf("drop count before lifetime elapsed = %d, want 1", got)
	}

	world.Step(time.Millisecond)
	if got := len(world.Snapshot().Drops); got != 0 {
		t.Fatalf("drop count after lifetime elapsed = %d, want 0", got)
	}
}

func TestInitialWorldDropDoesNotExpire(t *testing.T) {
	world := NewWorld()
	if err := world.AddWorldDrop(Position{}, string(WeaponSteelSword), 1); err != nil {
		t.Fatalf("AddWorldDrop() returned an unexpected error: %v", err)
	}

	world.Step(worldDropLifetime + time.Minute)
	if got := len(world.Snapshot().Drops); got != 1 {
		t.Fatalf("initial world drop count after lifetime elapsed = %d, want 1", got)
	}
}

func TestRespawningWorldDropReturnsAfterPickupCooldown(t *testing.T) {
	world := NewWorld()
	position := Position{X: 2, Z: 3}
	if err := world.AddRespawningWorldDrop(position, ItemFishingRod, 1, time.Minute); err != nil {
		t.Fatalf("AddRespawningWorldDrop() returned an unexpected error: %v", err)
	}
	mustAddPlayer(t, world, "player-one", position, 4)
	if err := world.QueuePickup(PickupCommand{
		PlayerID: "player-one", Sequence: 1, DropID: "drop-1",
	}); err != nil {
		t.Fatalf("QueuePickup() returned an unexpected error: %v", err)
	}
	world.Step(50 * time.Millisecond)
	if got := len(world.Snapshot().Drops); got != 0 {
		t.Fatalf("drop count after pickup = %d, want 0", got)
	}

	world.Step(time.Minute - time.Millisecond)
	if got := len(world.Snapshot().Drops); got != 0 {
		t.Fatalf("drop count before respawn = %d, want 0", got)
	}
	world.Step(time.Millisecond)
	drops := world.Snapshot().Drops
	if len(drops) != 1 || drops[0].ID == "drop-1" || drops[0].ItemID != ItemFishingRod ||
		drops[0].Position != position || drops[0].Infinite {
		t.Fatalf("respawned drop = %+v", drops)
	}
}

func TestStarterToolsExistWhenDebugContentIsDisabled(t *testing.T) {
	world := NewWorld()
	world.SetDebugContentEnabled(false)
	if err := world.AddStarterToolSpawns(); err != nil {
		t.Fatalf("AddStarterToolSpawns() returned an unexpected error: %v", err)
	}

	drops := world.Snapshot().Drops
	if len(drops) != len(starterToolSpawns) {
		t.Fatalf("starter tool drop count = %d, want %d", len(drops), len(starterToolSpawns))
	}
	found := make(map[string]Position, len(drops))
	for _, drop := range drops {
		if drop.Quantity != 1 || drop.Infinite {
			t.Fatalf("starter tool drop = %+v", drop)
		}
		found[drop.ItemID] = drop.Position
	}
	for _, spawn := range starterToolSpawns {
		if position, exists := found[spawn.ItemID]; !exists || position != spawn.Position {
			t.Errorf("starter tool %q position = %+v, found %t; want %+v", spawn.ItemID, position, exists, spawn.Position)
		}
	}
}

func TestInfiniteTestSupplyCanBeTakenRepeatedly(t *testing.T) {
	world := NewWorld()
	if err := world.AddTestSupply(Position{}, ItemCoins, 5); err != nil {
		t.Fatalf("AddTestSupply() returned an unexpected error: %v", err)
	}
	mustAddPlayer(t, world, "player-one", Position{}, 4)

	for sequence := uint64(1); sequence <= 2; sequence++ {
		if err := world.QueuePickup(PickupCommand{PlayerID: "player-one", Sequence: sequence, DropID: "drop-1"}); err != nil {
			t.Fatalf("QueuePickup(%d) returned an unexpected error: %v", sequence, err)
		}
		world.Step(50 * time.Millisecond)
	}

	player := mustPlayer(t, world, "player-one")
	if got := player.Inventory.itemsQuantity(ItemCoins); got != 10 {
		t.Fatalf("coin quantity after repeated test-supply takes = %d, want 10", got)
	}
	drops := world.Snapshot().Drops
	if len(drops) != 1 || !drops[0].Infinite {
		t.Fatalf("test supply after repeated takes = %+v", drops)
	}
}

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
	world.players["player-one"].skills[SkillAttack] = SkillExperienceForLevel(10)
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

func TestWorldRuneSwordCanBePickedUpEquippedAndUsed(t *testing.T) {
	world := NewWorldWithRollSource(fixedRollSource(0.5))
	playerPosition := Position{X: -4, Z: -1}
	if err := world.AddWorldDrop(playerPosition, string(WeaponRuneSword), 1); err != nil {
		t.Fatalf("AddWorldDrop() returned an unexpected error: %v", err)
	}
	mustAddPlayer(t, world, "player-one", playerPosition, 4)
	world.players["player-one"].skills[SkillAttack] = SkillExperienceForLevel(20)
	if err := world.QueuePickup(PickupCommand{PlayerID: "player-one", Sequence: 1, DropID: "drop-1"}); err != nil {
		t.Fatalf("QueuePickup() returned an unexpected error: %v", err)
	}
	world.Step(50 * time.Millisecond)
	if err := world.QueueEquip(EquipCommand{PlayerID: "player-one", Sequence: 1, ItemID: WeaponRuneSword}); err != nil {
		t.Fatalf("QueueEquip() returned an unexpected error: %v", err)
	}
	world.Step(actionPulseInterval)
	player := mustPlayer(t, world, "player-one")
	if player.EquippedWeapon != WeaponRuneSword || player.CombatStyle != CombatStyleStab {
		t.Fatalf("equipped rune sword state = %+v", player)
	}

	if err := world.QueueTarget(TargetCommand{PlayerID: "player-one", Sequence: 1, TargetID: dummyTargetID}); err != nil {
		t.Fatalf("QueueTarget() returned an unexpected error: %v", err)
	}
	world.Step(actionPulseInterval)
	if got, want := mustTarget(t, world.Snapshot(), dummyTargetID).Health, dummyMaximumHealth-runeSwordMeleeStrength; got != want {
		t.Fatalf("dummy health after rune sword hit = %d, want %d", got, want)
	}
}

func TestPickupWithinRangeAddsShardAndRemovesDrop(t *testing.T) {
	world := NewWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	world.spawnDrop(Position{X: pickupRange}, ItemCoins, 1)

	if err := world.QueuePickup(PickupCommand{PlayerID: "player-one", Sequence: 1, DropID: "drop-1"}); err != nil {
		t.Fatalf("QueuePickup() returned an unexpected error: %v", err)
	}
	world.Step(50 * time.Millisecond)

	if got := len(world.Snapshot().Drops); got != 0 {
		t.Fatalf("drop count after pickup = %d, want 0", got)
	}
	player := mustPlayer(t, world, "player-one")
	if got, want := len(player.Inventory.Items), 7; got != want {
		t.Fatalf("inventory item count = %d, want %d", got, want)
	}
	shard := player.Inventory.Items[0]
	if shard.Slot != 0 || shard.ItemID != ItemCoins || shard.Quantity != 1 {
		t.Fatalf("picked up shard = %+v", shard)
	}
}

func TestPickupOutsideRangeApproachesThenCollects(t *testing.T) {
	world := NewWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	world.spawnDrop(Position{X: pickupRange + 1}, ItemCoins, 1)

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
	world.spawnDrop(Position{X: 8}, ItemCoins, 1)
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
	world.spawnDrop(Position{}, ItemCoins, 1)

	for _, id := range []PlayerID{"player-two", "player-one"} {
		if err := world.QueuePickup(PickupCommand{PlayerID: id, Sequence: 1, DropID: "drop-1"}); err != nil {
			t.Fatalf("QueuePickup(%q) returned an unexpected error: %v", id, err)
		}
	}
	world.Step(50 * time.Millisecond)

	if !snapshotInventoryContains(mustPlayer(t, world, "player-one").Inventory, ItemCoins) {
		t.Fatal("deterministic first player did not receive contested drop")
	}
	if snapshotInventoryContains(mustPlayer(t, world, "player-two").Inventory, ItemCoins) {
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

func TestResourceDropAndPickupMovesEntireStack(t *testing.T) {
	tests := []struct {
		itemID   string
		quantity int
	}{
		{itemID: ItemAirRune, quantity: 50},
		{itemID: ItemCoins, quantity: 7},
		{itemID: ItemCoins, quantity: 1250},
	}
	for _, test := range tests {
		t.Run(test.itemID, func(t *testing.T) {
			world := NewWorld()
			if err := world.AddPlayer("player-one", Position{}, 4); err != nil {
				t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
			}
			world.players["player-one"].inventory.add(test.itemID, test.quantity)

			if err := world.QueueDropItem(DropItemCommand{
				PlayerID: "player-one", Sequence: 1, ItemID: test.itemID,
			}); err != nil {
				t.Fatalf("QueueDropItem() returned an unexpected error: %v", err)
			}
			world.Step(50 * time.Millisecond)
			if got := mustPlayer(t, world, "player-one").Inventory.itemsQuantity(test.itemID); got != 0 {
				t.Fatalf("inventory quantity after drop = %d, want 0", got)
			}
			drops := world.Snapshot().Drops
			if len(drops) != 1 || drops[0].ItemID != test.itemID || drops[0].Quantity != test.quantity {
				t.Fatalf("resource drop = %+v", drops)
			}

			if err := world.QueuePickup(PickupCommand{
				PlayerID: "player-one", Sequence: 1, DropID: drops[0].ID,
			}); err != nil {
				t.Fatalf("QueuePickup() returned an unexpected error: %v", err)
			}
			world.Step(50 * time.Millisecond)
			if got := mustPlayer(t, world, "player-one").Inventory.itemsQuantity(test.itemID); got != test.quantity {
				t.Fatalf("inventory quantity after pickup = %d, want %d", got, test.quantity)
			}
		})
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

func TestSpellbookMustBeUnequippedBeforeDropping(t *testing.T) {
	world := NewWorld()
	if err := world.AddPlayer("player-one", Position{}, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	player := world.players["player-one"]
	player.equippedSpellbook = ItemStandardSpellbook
	player.selectedSpell = SpellWindBolt

	if err := world.QueueDropItem(DropItemCommand{
		PlayerID: "player-one", Sequence: 1, ItemID: ItemStandardSpellbook,
	}); err == nil {
		t.Fatal("QueueDropItem() error = nil for equipped spellbook")
	}
	if err := world.QueueUnequip(UnequipCommand{
		PlayerID: "player-one", Sequence: 1, ItemID: ItemStandardSpellbook,
	}); err != nil {
		t.Fatalf("QueueUnequip() returned an unexpected error: %v", err)
	}
	world.Step(actionPulseInterval)
	snapshot := mustPlayer(t, world, "player-one")
	if snapshot.EquippedSpellbook != "" || snapshot.ActiveSpellbook != "" || snapshot.SelectedSpell != "" {
		t.Fatalf("state after unequipping spellbook = %+v", snapshot)
	}
	if err := world.QueueDropItem(DropItemCommand{
		PlayerID: "player-one", Sequence: 1, ItemID: ItemStandardSpellbook,
	}); err != nil {
		t.Fatalf("QueueDropItem() after unequip returned an unexpected error: %v", err)
	}
	world.Step(50 * time.Millisecond)
	snapshot = mustPlayer(t, world, "player-one")
	if snapshotInventoryContains(snapshot.Inventory, ItemStandardSpellbook) {
		t.Fatal("dropped spellbook remains in inventory")
	}
	drops := world.Snapshot().Drops
	if len(drops) != 1 || drops[0].ItemID != ItemStandardSpellbook || drops[0].Quantity != 1 {
		t.Fatalf("spellbook drops = %+v", drops)
	}
}

func TestDropRejectsSpellbookWithPendingEquipmentChange(t *testing.T) {
	world := NewWorld()
	if err := world.AddPlayer("player-one", Position{}, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	world.players["player-one"].inventory.add(ItemStandardSpellbook, 1)
	if err := world.QueueEquip(EquipCommand{
		PlayerID: "player-one", Sequence: 1, ItemID: WeaponID(ItemStandardSpellbook),
	}); err != nil {
		t.Fatalf("QueueEquip() returned an unexpected error: %v", err)
	}
	if err := world.QueueDropItem(DropItemCommand{
		PlayerID: "player-one", Sequence: 1, ItemID: ItemStandardSpellbook,
	}); err == nil {
		t.Fatal("QueueDropItem() error = nil during pending spellbook equip")
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

func TestDefeatedTargetSpawnsCoins(t *testing.T) {
	world := newCombatTestWorld()
	mustAddPlayer(t, world, "player-one", Position{X: -4, Z: -1}, 4)
	world.combat.targets[dummyTargetID].Health = swordDamage
	if err := world.QueueTarget(TargetCommand{PlayerID: "player-one", Sequence: 1, TargetID: dummyTargetID}); err != nil {
		t.Fatalf("QueueTarget() returned an unexpected error: %v", err)
	}
	world.Step(actionPulseInterval)

	drops := world.Snapshot().Drops
	if len(drops) != 1 || drops[0].ItemID != ItemCoins || drops[0].Quantity != 1 {
		t.Fatalf("target drops = %+v", drops)
	}
	if drops[0].Position != world.combat.targets[dummyTargetID].Position {
		t.Fatalf("drop position = %+v, want target position %+v", drops[0].Position, world.combat.targets[dummyTargetID].Position)
	}
}

func TestUpperDummySpawnsTrainingHelm(t *testing.T) {
	world := newCombatTestWorld()
	target := world.combat.targets[upperDummyTargetID]
	world.defeatDummy(upperDummyTargetID, target)
	drops := world.Snapshot().Drops
	if len(drops) != 1 || drops[0].ItemID != ItemTrainingHelm || drops[0].Quantity != 1 {
		t.Fatalf("upper dummy drops = %+v", drops)
	}
}

func TestUpperDummyCanAlsoDropRuneSword(t *testing.T) {
	world := NewWorldWithRollSource(fixedRollSource(0.1))
	target := world.combat.targets[upperDummyTargetID]
	world.defeatDummy(upperDummyTargetID, target)
	drops := world.Snapshot().Drops
	if len(drops) != 2 || drops[0].ItemID != ItemTrainingHelm || drops[0].Quantity != 1 ||
		drops[1].ItemID != string(WeaponRuneSword) || drops[1].Quantity != 1 {
		t.Fatalf("upper dummy rare drops = %+v", drops)
	}
}

func TestHillWardenAlwaysDropsAirRunesInConfiguredRange(t *testing.T) {
	world := NewWorldWithRollSource(fixedRollSource(0.5))
	target := world.combat.targets[hillWardenTargetID]
	world.defeatDummy(hillWardenTargetID, target)
	drops := world.Snapshot().Drops
	if len(drops) != 2 || drops[0].ItemID != ItemBones || drops[0].Quantity != 1 ||
		drops[1].ItemID != ItemAirRune || drops[1].Quantity != 8 || drops[1].Position != target.Position {
		t.Fatalf("Hill Warden common drops = %+v, want 8 air runes", drops)
	}
}

func TestHillWardenCanAlsoDropTrainingPlate(t *testing.T) {
	world := NewWorldWithRollSource(fixedRollSource(0.1))
	target := world.combat.targets[hillWardenTargetID]
	world.defeatDummy(hillWardenTargetID, target)
	drops := world.Snapshot().Drops
	if len(drops) != 3 || drops[0].ItemID != ItemBones || drops[1].ItemID != ItemAirRune || drops[1].Quantity != 5 ||
		drops[2].ItemID != ItemTrainingPlate || drops[2].Quantity != 1 {
		t.Fatalf("Hill Warden rare drops = %+v", drops)
	}
}

func TestRiftTyrantRareRingDropRateAndLoot(t *testing.T) {
	entries := lootTableForTarget(riftTyrantTargetID)
	if len(entries) != 3 || entries[2].ItemID != ItemTyrantsRing || entries[2].DropChance != 1.0/512.0 {
		t.Fatalf("Rift Tyrant loot table = %+v", entries)
	}

	world := NewWorldWithRollSource(fixedRollSource(0.001))
	target := world.combat.targets[riftTyrantTargetID]
	world.defeatDummy(riftTyrantTargetID, target)
	drops := world.Snapshot().Drops
	if len(drops) != 3 || drops[0].ItemID != ItemBones || drops[1].ItemID != ItemCoins ||
		drops[2].ItemID != ItemTyrantsRing || drops[2].Quantity != 1 {
		t.Fatalf("Rift Tyrant rare drops = %+v", drops)
	}
}

func TestWildFowlDropsFeathersInConfiguredRange(t *testing.T) {
	world := NewWorldWithRollSource(fixedRollSource(0.5))
	target := world.combat.targets[wildFowlTargetID]
	world.defeatDummy(wildFowlTargetID, target)
	drops := world.Snapshot().Drops
	if len(drops) != 2 || drops[0].ItemID != ItemBones || drops[0].Quantity != 1 ||
		drops[1].ItemID != ItemFeathers || drops[1].Quantity != 13 || drops[1].Position != target.Position {
		t.Fatalf("Wild Fowl drops = %+v, want 13 feathers", drops)
	}
}

func TestPickedUpAirRunesStackInExistingInventorySlot(t *testing.T) {
	world := NewWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	world.spawnDrop(Position{}, ItemAirRune, 8)
	if err := world.QueuePickup(PickupCommand{PlayerID: "player-one", Sequence: 1, DropID: "drop-1"}); err != nil {
		t.Fatalf("QueuePickup() returned an unexpected error: %v", err)
	}
	world.Step(50 * time.Millisecond)
	player := mustPlayer(t, world, "player-one")
	if got := player.Inventory.itemsQuantity(ItemAirRune); got != 58 {
		t.Fatalf("stacked air rune quantity = %d, want 58", got)
	}
	if got := len(player.Inventory.Items); got != 6 {
		t.Fatalf("inventory slots after stacking runes = %d, want 6", got)
	}
}
