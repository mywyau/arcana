package simulation

import "testing"

func TestPlayerStartsWithEightSlotInventory(t *testing.T) {
	world := NewWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 4)

	player := mustPlayer(t, world, "player-one")
	if got, want := player.Inventory.Capacity, 8; got != want {
		t.Fatalf("Inventory.Capacity = %d, want %d", got, want)
	}
	if got, want := len(player.Inventory.Items), 6; got != want {
		t.Fatalf("len(Inventory.Items) = %d, want %d", got, want)
	}
	if player.Inventory.Items[0].ItemID != "sword" || player.Inventory.Items[0].Slot != 0 {
		t.Fatalf("first inventory item = %+v, want sword in slot 0", player.Inventory.Items[0])
	}
	if player.Inventory.Items[1].ItemID != "dagger" || player.Inventory.Items[1].Slot != 1 {
		t.Fatalf("second inventory item = %+v, want dagger in slot 1", player.Inventory.Items[1])
	}
	if player.Inventory.Items[2].ItemID != "bow" || player.Inventory.Items[2].Slot != 2 {
		t.Fatalf("third inventory item = %+v, want bow in slot 2", player.Inventory.Items[2])
	}
	wantAdditional := []string{"mace", "battle_axe", "staff"}
	for index, itemID := range wantAdditional {
		item := player.Inventory.Items[index+3]
		if item.ItemID != itemID || item.Slot != index+3 {
			t.Fatalf("inventory item %d = %+v, want %s in slot %d", index+3, item, itemID, index+3)
		}
	}

	player.Inventory.Items[0].ItemID = "changed"
	fresh := mustPlayer(t, world, "player-one")
	if fresh.Inventory.Items[0].ItemID != "sword" {
		t.Fatalf("authoritative inventory changed through snapshot: %+v", fresh.Inventory.Items)
	}
}

func TestQueueEquipRejectsItemOutsideInventory(t *testing.T) {
	world := NewWorld()
	mustAddPlayer(t, world, "player-one", Position{}, 4)

	world.mu.Lock()
	world.players["player-one"].inventory.items = world.players["player-one"].inventory.items[:1]
	world.mu.Unlock()

	err := world.QueueEquip(EquipCommand{PlayerID: "player-one", Sequence: 1, ItemID: WeaponDagger})
	if err == nil {
		t.Fatal("QueueEquip() error = nil for an item outside the player's inventory")
	}
}
