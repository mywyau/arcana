package simulation

import (
	"reflect"
	"testing"
	"time"
)

func TestPlayerStartsUnarmedWithEmptyTwentySlotInventory(t *testing.T) {
	world := NewWorld()
	if err := world.AddPlayer("player-one", Position{}, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}

	player := mustPlayer(t, world, "player-one")
	if got, want := player.Inventory.Capacity, 20; got != want {
		t.Fatalf("Inventory.Capacity = %d, want %d", got, want)
	}
	if got, want := len(player.Inventory.Items), 0; got != want {
		t.Fatalf("len(Inventory.Items) = %d, want %d", got, want)
	}
	if player.EquippedWeapon != WeaponUnarmed || player.CombatStyle != CombatStyleCrush ||
		player.EquippedSpellbook != "" || player.ActiveSpellbook != "" || len(player.EquippedArmour) != 0 {
		t.Fatalf("fresh equipment state = %+v", player)
	}
}

func TestInventorySnapshotSupportsFiftySlotsAndRejectsMore(t *testing.T) {
	if _, err := inventoryFromSnapshot(InventorySnapshot{Capacity: MaxInventoryCapacity}); err != nil {
		t.Fatalf("inventoryFromSnapshot() rejected maximum capacity: %v", err)
	}
	_, err := inventoryFromSnapshot(InventorySnapshot{Capacity: MaxInventoryCapacity + 1})
	if err == nil {
		t.Fatal("inventoryFromSnapshot() error = nil for capacity above maximum")
	}
}

func TestCookedMinnowsStackToFivePerSlot(t *testing.T) {
	inventory := inventoryState{capacity: 3}
	if !inventory.add(ItemCookedMinnow, 8) {
		t.Fatal("add() rejected minnow stacks with enough slots")
	}
	want := []InventoryItem{
		{Slot: 0, ItemID: ItemCookedMinnow, Quantity: 5},
		{Slot: 1, ItemID: ItemCookedMinnow, Quantity: 3},
	}
	if !reflect.DeepEqual(inventory.items, want) {
		t.Fatalf("minnow stacks = %+v, want %+v", inventory.items, want)
	}
	if !inventory.add(ItemCookedMinnow, 2) || inventory.items[1].Quantity != 5 {
		t.Fatalf("filled minnow stacks = %+v", inventory.items)
	}
	if !inventory.remove(ItemCookedMinnow, 7) || inventory.quantity(ItemCookedMinnow) != 3 {
		t.Fatalf("minnow stacks after removal = %+v", inventory.items)
	}
}

func TestInventoryRevisionChangesOnlyAfterSuccessfulMutations(t *testing.T) {
	inventory := newEmptyInventory()
	initial := inventory.revision
	if !inventory.add(ItemCookedMinnow, 1) || inventory.revision <= initial {
		t.Fatalf("revision after add = %d, want greater than %d", inventory.revision, initial)
	}
	afterAdd := inventory.revision
	if !inventory.remove(ItemCookedMinnow, 1) || inventory.revision <= afterAdd {
		t.Fatalf("revision after remove = %d, want greater than %d", inventory.revision, afterAdd)
	}
	afterRemove := inventory.revision
	if inventory.remove(ItemCookedMinnow, 2) || inventory.revision != afterRemove {
		t.Fatalf("failed removal changed revision from %d to %d", afterRemove, inventory.revision)
	}
	if !inventory.add(ItemCookedMinnow, 1) {
		t.Fatal("could not restore item for slot-move check")
	}
	beforeMove := inventory.revision
	if !inventory.swapSlots(0, 1) || inventory.revision <= beforeMove {
		t.Fatalf("revision after slot move = %d, want greater than %d", inventory.revision, beforeMove)
	}
}

func TestEquipmentAndInventoryHaveIndependentRevisions(t *testing.T) {
	world := NewWorld()
	if err := world.AddPlayer("player-one", Position{}, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	player := world.players["player-one"]
	if !player.inventory.add(string(WeaponBronzeSword), 1) {
		t.Fatal("could not add bronze sword")
	}
	inventoryRevision := player.inventory.revision
	equipmentRevision := player.equipmentRevision
	if err := world.QueueEquip(EquipCommand{PlayerID: "player-one", Sequence: 1, ItemID: WeaponBronzeSword}); err != nil {
		t.Fatalf("QueueEquip() returned an unexpected error: %v", err)
	}
	world.Step(actionPulseInterval)
	if player.equipmentRevision <= equipmentRevision {
		t.Fatalf("equipment revision = %d, want greater than %d", player.equipmentRevision, equipmentRevision)
	}
	if player.inventory.revision <= inventoryRevision {
		t.Fatalf("inventory revision = %d, want greater than %d", player.inventory.revision, inventoryRevision)
	}
}

func TestPrayerPotionDosesStackToFivePerSlot(t *testing.T) {
	inventory := inventoryState{capacity: 2}
	if !inventory.add(ItemPrayerPotion, 6) {
		t.Fatal("add() rejected prayer potion doses with enough slots")
	}
	want := []InventoryItem{
		{Slot: 0, ItemID: ItemPrayerPotion, Quantity: 5},
		{Slot: 1, ItemID: ItemPrayerPotion, Quantity: 1},
	}
	if !reflect.DeepEqual(inventory.items, want) {
		t.Fatalf("prayer potion stacks = %+v, want %+v", inventory.items, want)
	}
}

func TestCookedMinnowAdditionIsAtomicWhenSlotsAreFull(t *testing.T) {
	inventory := inventoryState{capacity: 1}
	if inventory.add(ItemCookedMinnow, 11) {
		t.Fatal("add() accepted more minnows than the inventory can hold")
	}
	if len(inventory.items) != 0 {
		t.Fatalf("failed addition mutated inventory: %+v", inventory.items)
	}
}

func TestInventorySnapshotSplitsLegacyCookedMinnowStacks(t *testing.T) {
	snapshot := InventorySnapshot{
		Capacity: 3,
		Items: []InventoryItem{
			{Slot: 0, ItemID: ItemCookedMinnow, Quantity: 10},
			{Slot: 2, ItemID: ItemCookedMinnow, Quantity: 4},
		},
	}
	inventory, err := inventoryFromSnapshot(snapshot)
	if err != nil {
		t.Fatalf("inventoryFromSnapshot() rejected legacy minnow stacks: %v", err)
	}
	if inventory.quantity(ItemCookedMinnow) != 14 || len(inventory.items) != 3 {
		t.Fatalf("split legacy minnow inventory = %+v", inventory.items)
	}
	for _, item := range inventory.items {
		if item.Quantity > 5 {
			t.Fatalf("legacy minnow stack was not reduced to five: %+v", item)
		}
	}

	legacyFullInventory, err := inventoryFromSnapshot(InventorySnapshot{
		Capacity: 1,
		Items:    []InventoryItem{{Slot: 0, ItemID: ItemCookedMinnow, Quantity: 6}},
	})
	if err != nil || legacyFullInventory.quantity(ItemCookedMinnow) != 6 {
		t.Fatalf("full legacy inventory could not be restored safely: inventory %+v, error %v", legacyFullInventory, err)
	}
	if legacyFullInventory.add(ItemCookedMinnow, 1) {
		t.Fatal("inventory added food to a grandfathered stack above the current limit")
	}
}

func TestQueueEquipRejectsItemOutsideInventory(t *testing.T) {
	world := NewWorld()
	if err := world.AddPlayer("player-one", Position{}, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}

	err := world.QueueEquip(EquipCommand{PlayerID: "player-one", Sequence: 1, ItemID: WeaponDagger})
	if err == nil {
		t.Fatal("QueueEquip() error = nil for an item outside the player's inventory")
	}
}

func TestQueueEquipEnforcesAndThenAcceptsSkillRequirement(t *testing.T) {
	world := NewWorld()
	if err := world.AddPlayer("player-one", Position{}, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	player := world.players["player-one"]
	if !player.inventory.add(string(WeaponSteelSword), 1) {
		t.Fatal("could not add steel sword")
	}
	if err := world.QueueEquip(EquipCommand{PlayerID: "player-one", Sequence: 1, ItemID: WeaponSteelSword}); err == nil {
		t.Fatal("QueueEquip() error = nil below the Attack requirement")
	}
	player.skills[SkillAttack] = SkillExperienceForLevel(10)
	if err := world.QueueEquip(EquipCommand{PlayerID: "player-one", Sequence: 1, ItemID: WeaponSteelSword}); err != nil {
		t.Fatalf("QueueEquip() at the Attack requirement returned an unexpected error: %v", err)
	}
}

func TestMoveInventoryItemSwapsOccupiedSlots(t *testing.T) {
	world := NewWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	if err := world.QueueMoveInventoryItem(MoveInventoryItemCommand{
		PlayerID: "player-one", Sequence: 1, FromSlot: 1, ToSlot: 2,
	}); err != nil {
		t.Fatalf("QueueMoveInventoryItem() returned an unexpected error: %v", err)
	}
	world.Step(time.Millisecond)
	items := mustPlayer(t, world, "player-one").Inventory.Items
	if items[0].ItemID != string(WeaponBow) || items[1].ItemID != string(WeaponDagger) {
		t.Fatalf("swapped inventory = %+v", items)
	}
}

func TestMoveInventoryItemMovesIntoEmptySlot(t *testing.T) {
	world := NewWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	if err := world.QueueMoveInventoryItem(MoveInventoryItemCommand{
		PlayerID: "player-one", Sequence: 1, FromSlot: 1, ToSlot: 7,
	}); err != nil {
		t.Fatalf("QueueMoveInventoryItem() returned an unexpected error: %v", err)
	}
	world.Step(time.Millisecond)
	items := mustPlayer(t, world, "player-one").Inventory.Items
	if items[len(items)-1].Slot != 7 || items[len(items)-1].ItemID != string(WeaponDagger) {
		t.Fatalf("moved inventory = %+v", items)
	}
}
