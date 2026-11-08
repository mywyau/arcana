package simulation

import (
	"testing"
	"time"
)

func TestRunePouchMovesWholeStacksAndRejectsOtherResources(t *testing.T) {
	world := NewWorld()
	if err := world.AddPlayer("player-one", Position{}, 4); err != nil {
		t.Fatal(err)
	}
	player := world.players["player-one"]
	player.inventory.add(ItemAirRune, 25)
	player.inventory.add(ItemCoins, 10)
	if err := world.QueueRuneTransfer(RuneTransferCommand{
		PlayerID: "player-one", Sequence: 1, Deposit: true, InventorySlot: 0,
	}); err != nil {
		t.Fatalf("deposit runes: %v", err)
	}
	world.Step(time.Millisecond)
	if player.runePouch.quantity(ItemAirRune) != 25 || player.inventory.quantity(ItemAirRune) != 0 {
		t.Fatalf("rune transfer state = pouch %+v, inventory %+v", player.runePouch, player.inventory)
	}
	if err := world.QueueRuneTransfer(RuneTransferCommand{
		PlayerID: "player-one", Sequence: 2, Deposit: true, InventorySlot: 1,
	}); err == nil {
		t.Fatal("non-rune deposit was accepted")
	}
}

func TestSpellCostsUseRunePouchBeforeInventory(t *testing.T) {
	player := &playerState{inventory: newEmptyInventory(), runePouch: newRunePouch()}
	player.runePouch.add(ItemFireRune, 1)
	player.inventory.add(ItemFireRune, 2)
	if !paySpellCosts(player, []ItemCost{{ItemID: ItemFireRune, Quantity: 2}}) {
		t.Fatal("paySpellCosts() rejected available runes")
	}
	if player.runePouch.quantity(ItemFireRune) != 0 || player.inventory.quantity(ItemFireRune) != 1 {
		t.Fatalf("remaining runes = pouch %d, inventory %d", player.runePouch.quantity(ItemFireRune), player.inventory.quantity(ItemFireRune))
	}

	before := player.inventory.quantity(ItemFireRune)
	if paySpellCosts(player, []ItemCost{{ItemID: ItemFireRune, Quantity: 2}}) {
		t.Fatal("paySpellCosts() accepted insufficient runes")
	}
	if player.inventory.quantity(ItemFireRune) != before {
		t.Fatal("failed payment consumed runes")
	}
}
