package simulation

import (
	"testing"
	"time"
)

func TestBankDepositsWholeStackAndWithdrawsRequestedQuantity(t *testing.T) {
	world := NewWorld()
	if err := world.AddPlayer("player-one", starterBankChest.Position, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	player := world.players["player-one"]
	player.inventory.add(ItemAirRune, 3)

	if err := world.QueueBankTransfer(BankTransferCommand{
		PlayerID: "player-one", Sequence: 1, Deposit: true, ItemID: ItemAirRune, InventorySlot: 0,
	}); err != nil {
		t.Fatalf("QueueBankTransfer(deposit) returned an unexpected error: %v", err)
	}
	world.Step(time.Millisecond)
	snapshot := mustPlayer(t, world, "player-one")
	if snapshot.Inventory.itemsQuantity(ItemAirRune) != 0 || len(snapshot.Bank.Items) != 1 || snapshot.Bank.Items[0].Quantity != 3 {
		t.Fatalf("state after deposits = inventory %+v, bank %+v", snapshot.Inventory, snapshot.Bank)
	}

	if err := world.QueueBankTransfer(BankTransferCommand{
		PlayerID: "player-one", Sequence: 2, BankSlot: snapshot.Bank.Items[0].Slot, Quantity: 2,
	}); err != nil {
		t.Fatalf("QueueBankTransfer(withdraw) returned an unexpected error: %v", err)
	}
	world.Step(time.Millisecond)
	snapshot = mustPlayer(t, world, "player-one")
	if snapshot.Inventory.itemsQuantity(ItemAirRune) != 2 || snapshot.Bank.Items[0].Quantity != 1 || snapshot.LastBankSequence != 2 {
		t.Fatalf("state after withdrawal = %+v", snapshot)
	}
}

func TestBankResourceStackHasOneSlotAndIntegerMaximum(t *testing.T) {
	bank := newEmptyBank()
	if !bank.add(InventoryItem{ItemID: ItemAirRune, Quantity: MaximumBankStackQuantity - 1}) ||
		!bank.add(InventoryItem{ItemID: ItemAirRune, Quantity: 1}) {
		t.Fatal("bank rejected a resource stack within its maximum")
	}
	if len(bank.items) != 1 || bank.items[0].Quantity != MaximumBankStackQuantity {
		t.Fatalf("maximum bank stack = %+v", bank.items)
	}
	if bank.add(InventoryItem{ItemID: ItemAirRune, Quantity: 1}) {
		t.Fatal("bank accepted a resource stack above its maximum")
	}
	if len(bank.items) != 1 || bank.items[0].Quantity != MaximumBankStackQuantity {
		t.Fatalf("rejected deposit changed bank stack = %+v", bank.items)
	}
}

func TestBankRejectsInvalidWithdrawQuantityAtomically(t *testing.T) {
	world := NewWorld()
	if err := world.AddPlayer("player-one", starterBankChest.Position, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	player := world.players["player-one"]
	player.bank.add(InventoryItem{ItemID: ItemAirRune, Quantity: 3})
	if err := world.QueueBankTransfer(BankTransferCommand{
		PlayerID: "player-one", Sequence: 1, BankSlot: 0, Quantity: 4,
	}); err == nil {
		t.Fatal("QueueBankTransfer() error = nil above available quantity")
	}
	if player.inventory.quantity(ItemAirRune) != 0 || player.bank.items[0].Quantity != 3 {
		t.Fatalf("rejected withdrawal changed state: inventory %+v, bank %+v", player.inventory, player.bank)
	}
}

func TestBankSnapshotRejectsResourceSplitAcrossSlots(t *testing.T) {
	_, err := bankFromSnapshot(BankSnapshot{Capacity: BankCapacity, Items: []InventoryItem{
		{Slot: 0, ItemID: ItemAirRune, Quantity: 2},
		{Slot: 1, ItemID: ItemAirRune, Quantity: 3},
	}})
	if err == nil {
		t.Fatal("bankFromSnapshot() error = nil for a resource split across slots")
	}
}

func TestBankDepositsOnlyTheClickedInventoryStack(t *testing.T) {
	world := NewWorld()
	if err := world.AddPlayer("player-one", starterBankChest.Position, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	player := world.players["player-one"]
	player.inventory.add(ItemNormalLogs, 15)
	if err := world.QueueBankTransfer(BankTransferCommand{
		PlayerID: "player-one", Sequence: 1, Deposit: true, ItemID: ItemNormalLogs, InventorySlot: 1,
	}); err != nil {
		t.Fatalf("QueueBankTransfer() returned an unexpected error: %v", err)
	}
	world.Step(time.Millisecond)
	snapshot := mustPlayer(t, world, "player-one")
	if snapshot.Inventory.itemsQuantity(ItemNormalLogs) != 10 || len(snapshot.Bank.Items) != 1 || snapshot.Bank.Items[0].Quantity != 5 {
		t.Fatalf("state after clicked-stack deposit = inventory %+v, bank %+v", snapshot.Inventory, snapshot.Bank)
	}
}

func TestBankPreservesEquipmentInstanceMetadata(t *testing.T) {
	world := NewWorld()
	if err := world.AddPlayer("player-one", starterBankChest.Position, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	player := world.players["player-one"]
	player.inventory.add(string(WeaponBronzeSword), 1)
	item := player.inventory.items[0]
	player.inventory.items[0].Reinforcement = 1

	if err := world.QueueBankTransfer(BankTransferCommand{
		PlayerID: "player-one", Sequence: 1, Deposit: true,
		ItemID: item.ItemID, InstanceID: item.InstanceID,
	}); err != nil {
		t.Fatalf("QueueBankTransfer() returned an unexpected error: %v", err)
	}
	world.Step(time.Millisecond)
	snapshot := mustPlayer(t, world, "player-one")
	if len(snapshot.Bank.Items) != 1 || snapshot.Bank.Items[0].InstanceID != item.InstanceID || snapshot.Bank.Items[0].Reinforcement != 1 {
		t.Fatalf("banked equipment = %+v", snapshot.Bank.Items)
	}
}

func TestBankRejectsTransfersOutsideInteractionRange(t *testing.T) {
	world := NewWorld()
	if err := world.AddPlayer("player-one", Position{}, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	world.players["player-one"].inventory.add(ItemAirRune, 1)
	if err := world.QueueBankTransfer(BankTransferCommand{
		PlayerID: "player-one", Sequence: 1, Deposit: true, ItemID: ItemAirRune,
	}); err == nil {
		t.Fatal("QueueBankTransfer() error = nil outside bank range")
	}
}

func TestBankRejectsEquippedItem(t *testing.T) {
	world := NewWorld()
	if err := world.AddPlayer("player-one", starterBankChest.Position, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	player := world.players["player-one"]
	player.inventory.add(string(WeaponBronzeSword), 1)
	item := player.inventory.items[0]
	player.equippedWeapon = WeaponBronzeSword
	player.equippedWeaponInstanceID = item.InstanceID
	if err := world.QueueBankTransfer(BankTransferCommand{
		PlayerID: "player-one", Sequence: 1, Deposit: true,
		ItemID: item.ItemID, InstanceID: item.InstanceID,
	}); err == nil {
		t.Fatal("QueueBankTransfer() error = nil for equipped item")
	}
}

func TestBankRejectsTransfersDuringCombat(t *testing.T) {
	world := NewWorld()
	if err := world.AddPlayer("player-one", starterBankChest.Position, 4); err != nil {
		t.Fatalf("AddPlayer() returned an unexpected error: %v", err)
	}
	player := world.players["player-one"]
	player.inventory.add(ItemAirRune, 1)
	player.combatTarget = dummyTargetID
	if err := world.QueueBankTransfer(BankTransferCommand{
		PlayerID: "player-one", Sequence: 1, Deposit: true, ItemID: ItemAirRune,
	}); err == nil {
		t.Fatal("QueueBankTransfer() error = nil during combat")
	}
}
