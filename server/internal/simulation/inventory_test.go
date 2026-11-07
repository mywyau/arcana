package simulation

import (
	"reflect"
	"testing"
	"time"
)

func TestPlayerStartsUnarmedWithEmptyEightSlotInventory(t *testing.T) {
	world := NewWorld()
	if err := world.AddPlayer("player-one", Position{}, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}

	player := mustPlayer(t, world, "player-one")
	if got, want := player.Inventory.Capacity, 8; got != want {
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

func TestCookedMinnowsStackToTenPerSlot(t *testing.T) {
	inventory := inventoryState{capacity: 3}
	if !inventory.add(ItemCookedMinnow, 15) {
		t.Fatal("add() rejected minnow stacks with enough slots")
	}
	want := []InventoryItem{
		{Slot: 0, ItemID: ItemCookedMinnow, Quantity: 10},
		{Slot: 1, ItemID: ItemCookedMinnow, Quantity: 5},
	}
	if !reflect.DeepEqual(inventory.items, want) {
		t.Fatalf("minnow stacks = %+v, want %+v", inventory.items, want)
	}
	if !inventory.add(ItemCookedMinnow, 5) || inventory.items[1].Quantity != 10 {
		t.Fatalf("filled minnow stacks = %+v", inventory.items)
	}
	if !inventory.remove(ItemCookedMinnow, 12) || inventory.quantity(ItemCookedMinnow) != 8 {
		t.Fatalf("minnow stacks after removal = %+v", inventory.items)
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

func TestInventorySnapshotAllowsValidCookedMinnowStacks(t *testing.T) {
	snapshot := InventorySnapshot{
		Capacity: 2,
		Items: []InventoryItem{
			{Slot: 0, ItemID: ItemCookedMinnow, Quantity: 10},
			{Slot: 1, ItemID: ItemCookedMinnow, Quantity: 4},
		},
	}
	if _, err := inventoryFromSnapshot(snapshot); err != nil {
		t.Fatalf("inventoryFromSnapshot() rejected valid minnow stacks: %v", err)
	}
	snapshot.Items[1].Quantity = 11
	if _, err := inventoryFromSnapshot(snapshot); err == nil {
		t.Fatal("inventoryFromSnapshot() accepted a minnow stack above 10")
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

func TestMoveInventoryItemSwapsOccupiedSlots(t *testing.T) {
	world := NewWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	if err := world.QueueMoveInventoryItem(MoveInventoryItemCommand{
		PlayerID: "player-one", Sequence: 1, FromSlot: 0, ToSlot: 1,
	}); err != nil {
		t.Fatalf("QueueMoveInventoryItem() returned an unexpected error: %v", err)
	}
	world.Step(time.Millisecond)
	items := mustPlayer(t, world, "player-one").Inventory.Items
	if items[0].ItemID != string(WeaponDagger) || items[1].ItemID != string(WeaponSword) {
		t.Fatalf("swapped inventory = %+v", items)
	}
}

func TestMoveInventoryItemMovesIntoEmptySlot(t *testing.T) {
	world := NewWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 4)
	if err := world.QueueMoveInventoryItem(MoveInventoryItemCommand{
		PlayerID: "player-one", Sequence: 1, FromSlot: 0, ToSlot: 7,
	}); err != nil {
		t.Fatalf("QueueMoveInventoryItem() returned an unexpected error: %v", err)
	}
	world.Step(time.Millisecond)
	items := mustPlayer(t, world, "player-one").Inventory.Items
	if items[len(items)-1].Slot != 7 || items[len(items)-1].ItemID != string(WeaponSword) {
		t.Fatalf("moved inventory = %+v", items)
	}
}
