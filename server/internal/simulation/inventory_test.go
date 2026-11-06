package simulation

import (
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
