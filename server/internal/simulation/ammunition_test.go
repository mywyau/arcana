package simulation

import (
	"testing"
	"time"
)

func TestAmmunitionPouchStartsWithTwoOfTenSlotsUnlocked(t *testing.T) {
	snapshot := snapshotAmmunitionPouch(newAmmunitionPouch())
	if snapshot.Capacity != AmmunitionPouchCapacity || snapshot.UnlockedSlots != BaseAmmunitionPouchUnlocked {
		t.Fatalf("ammunition pouch = %+v", snapshot)
	}
}

func TestAmmunitionTransferMovesWholeStackBothWays(t *testing.T) {
	world := NewWorld()
	if err := world.AddPlayer("player-one", Position{}, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	player := world.players["player-one"]
	if !player.inventory.add(ItemIronArrow, 25) {
		t.Fatal("could not add iron arrows")
	}
	if err := world.QueueAmmunitionTransfer(AmmunitionTransferCommand{
		PlayerID: "player-one", Sequence: 1, Deposit: true, InventorySlot: 0,
	}); err != nil {
		t.Fatalf("QueueAmmunitionTransfer(deposit) returned an unexpected error: %v", err)
	}
	world.Step(time.Millisecond)
	snapshot := mustPlayer(t, world, "player-one")
	if snapshot.Inventory.itemsQuantity(ItemIronArrow) != 0 ||
		(InventorySnapshot{Items: snapshot.AmmunitionPouch.Items}).itemsQuantity(ItemIronArrow) != 25 {
		t.Fatalf("state after ammunition deposit = inventory %+v, pouch %+v", snapshot.Inventory, snapshot.AmmunitionPouch)
	}

	if err := world.QueueAmmunitionTransfer(AmmunitionTransferCommand{
		PlayerID: "player-one", Sequence: 2, PouchSlot: 0,
	}); err != nil {
		t.Fatalf("QueueAmmunitionTransfer(withdraw) returned an unexpected error: %v", err)
	}
	world.Step(time.Millisecond)
	snapshot = mustPlayer(t, world, "player-one")
	if snapshot.Inventory.itemsQuantity(ItemIronArrow) != 25 || len(snapshot.AmmunitionPouch.Items) != 0 {
		t.Fatalf("state after ammunition withdrawal = inventory %+v, pouch %+v", snapshot.Inventory, snapshot.AmmunitionPouch)
	}
}

func TestAmmunitionPouchRejectsNonAmmunition(t *testing.T) {
	world := NewWorld()
	if err := world.AddPlayer("player-one", Position{}, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	world.players["player-one"].inventory.add(ItemAirRune, 5)
	if err := world.QueueAmmunitionTransfer(AmmunitionTransferCommand{
		PlayerID: "player-one", Sequence: 1, Deposit: true, InventorySlot: 0,
	}); err == nil {
		t.Fatal("QueueAmmunitionTransfer() error = nil for a non-ammunition item")
	}
}
