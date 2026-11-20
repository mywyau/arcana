package simulation

import "testing"

func alchemyTestWorld(t *testing.T) (*World, *playerState) {
	t.Helper()
	w := NewWorld()
	mustAddEmptyPlayer(t, w, "mage", Position{})
	p := w.players["mage"]
	p.equippedSpellbook = ItemStandardSpellbook
	p.runePouch.add(ItemFocusRune, 30)
	return w, p
}

func TestAlchemyAutoConsumesOnlyTheSelectedStack(t *testing.T) {
	w, p := alchemyTestWorld(t)
	p.inventory.items = []InventoryItem{
		{Slot: 0, ItemID: ItemIronArrow, Quantity: 50},
		{Slot: 2, ItemID: ItemIronArrow, Quantity: 35},
	}
	if err := w.QueueAlchemy(p.id, 1, 2, ItemIronArrow, "", 21, true); err == nil {
		t.Fatal("accepted batch over 20")
	}
	if err := w.QueueAlchemy(p.id, 1, 2, ItemIronArrow, "", 20, true); err != nil {
		t.Fatal(err)
	}
	if err := w.QueueAlchemy(p.id, 2, 0, ItemIronArrow, "", 1, false); err == nil {
		t.Fatal("queued a manual cast during an automatic cast")
	}
	w.Step(4 * actionPulseInterval)
	if p.inventory.quantity(ItemCoins) != 0 {
		t.Fatal("auto cast fired early")
	}
	w.Step(actionPulseInterval)
	if p.inventory.quantity(ItemCoins) != 3 || p.alchemy.remaining != 19 {
		t.Fatal("first cast must convert exactly one")
	}
	w.Step(100 * actionPulseInterval)
	first, _ := p.inventory.itemAtSlot(0)
	clicked, _ := p.inventory.itemAtSlot(2)
	if first.Quantity != 50 || clicked.Quantity != 15 || p.inventory.quantity(ItemCoins) != 60 || p.runePouch.quantity(ItemFocusRune) != 10 || p.alchemy.remaining != 0 {
		t.Fatalf("unexpected final state: %+v, runes %v", p.inventory, p.runePouch)
	}
}

func TestAlchemyManualAndSingleItemAuto(t *testing.T) {
	for _, auto := range []bool{false, true} {
		w, p := alchemyTestWorld(t)
		p.inventory.add(ItemBronzeShield, 2)
		chosen, _ := p.inventory.itemAtSlot(1)
		if err := w.QueueAlchemy(p.id, 1, chosen.Slot, chosen.ItemID, chosen.InstanceID, 1, auto); err != nil {
			t.Fatal(err)
		}
		pulses := manualAlchemyPulses
		if auto {
			pulses = autoAlchemyPulses
		}
		for i := 1; i < pulses; i++ {
			w.Step(actionPulseInterval)
		}
		if p.inventory.quantity(ItemBronzeShield) != 2 {
			t.Fatal("cast fired early")
		}
		w.Step(actionPulseInterval)
		w.Step(10 * actionPulseInterval)
		if _, exists := p.inventory.itemByInstance(chosen.InstanceID); exists || p.inventory.quantity(ItemBronzeShield) != 1 || p.inventory.quantity(ItemCoins) != 22 || p.runePouch.quantity(ItemFocusRune) != 29 {
			t.Fatalf("wrong item or payout for auto=%v: %+v", auto, p.inventory)
		}
	}
}

func TestAlchemyRechecksTargetAndRunePayment(t *testing.T) {
	for _, change := range []string{"protect", "replace", "cancel", "disconnect", "remove book", "no runes", "no coin slot"} {
		t.Run(change, func(t *testing.T) {
			w, p := alchemyTestWorld(t)
			p.inventory.add(ItemBronzeBolt, 4)
			if err := w.QueueAlchemy(p.id, 1, 0, ItemBronzeBolt, "", 4, true); err != nil {
				t.Fatal(err)
			}
			switch change {
			case "protect":
				p.protectedInventory = map[int]protectedInventoryEntry{0: {Slot: 0, ItemID: ItemBronzeBolt}}
			case "replace":
				p.inventory.removeFromSlot(0, ItemBronzeBolt, 4)
				p.inventory.add(ItemBronzeBolt, 4)
			case "cancel":
				if err := w.QueueAlchemy(p.id, 2, 0, "", "", 0, false); err != nil {
					t.Fatal(err)
				}
			case "disconnect":
				if err := w.QueueSuspend(p.id); err != nil {
					t.Fatal(err)
				}
			case "remove book":
				p.equippedSpellbook = ""
			case "no runes":
				p.runePouch.remove(ItemFocusRune, 30)
			case "no coin slot":
				p.inventory.capacity = 1
			}
			runes := p.runePouch.quantity(ItemFocusRune)
			w.Step(6 * actionPulseInterval)
			if p.alchemy.remaining != 0 || p.inventory.quantity(ItemBronzeBolt) != 4 || p.inventory.quantity(ItemCoins) != 0 || p.runePouch.quantity(ItemFocusRune) != runes {
				t.Fatal("interrupted cast consumed an item or rune")
			}
		})
	}
}
