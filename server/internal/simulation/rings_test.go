package simulation

import (
	"testing"
	"time"
)

func TestRingRackEquipsFiveDistinctRingInstances(t *testing.T) {
	world := NewWorld()
	if err := world.AddPlayer("player-one", Position{}, 4); err != nil {
		t.Fatal(err)
	}
	player := world.players["player-one"]
	for _, itemID := range []string{ItemRubyRing, ItemEmeraldRing, ItemSapphireRing, ItemDiamondRing, ItemRubyRing} {
		if !player.inventory.add(itemID, 1) {
			t.Fatalf("could not add %q", itemID)
		}
	}
	for sequence := 1; sequence <= RingRackCapacity; sequence++ {
		item := player.inventory.items[0]
		if err := world.QueueRingTransfer(RingTransferCommand{
			PlayerID: "player-one", Sequence: uint64(sequence), Deposit: true,
			ItemID: item.ItemID, InstanceID: item.InstanceID,
		}); err != nil {
			t.Fatalf("QueueRingTransfer(%d): %v", sequence, err)
		}
		world.Step(time.Millisecond)
	}
	snapshot := mustPlayer(t, world, "player-one")
	if len(snapshot.RingRack.Items) != RingRackCapacity || len(snapshot.Inventory.Items) != 0 {
		t.Fatalf("ring state = rack %+v, inventory %+v", snapshot.RingRack, snapshot.Inventory)
	}

	if err := world.QueueRingTransfer(RingTransferCommand{
		PlayerID: "player-one", Sequence: 6, RackSlot: 0,
	}); err != nil {
		t.Fatalf("withdraw ring: %v", err)
	}
	world.Step(time.Millisecond)
	snapshot = mustPlayer(t, world, "player-one")
	if len(snapshot.RingRack.Items) != 4 || len(snapshot.Inventory.Items) != 1 || snapshot.LastRingSequence != 6 {
		t.Fatalf("state after withdrawal = %+v", snapshot)
	}
}

func TestEquippedRingsContributeTheirCombatBonuses(t *testing.T) {
	player := &playerState{
		equippedArmour: make(map[ArmourSlot]string),
		ringRack: inventoryState{capacity: RingRackCapacity, items: []InventoryItem{
			{Slot: 0, ItemID: ItemRubyRing, Quantity: 1, InstanceID: "ruby"},
			{Slot: 1, ItemID: ItemDiamondRing, Quantity: 1, InstanceID: "diamond"},
		}},
	}
	accuracy, strength := equipmentOffenceForWeapon(player, WeaponDefinition{MeleeStrength: 1})
	if accuracy != 1 || strength != 3 {
		t.Fatalf("melee ring bonuses = accuracy %d, strength %d", accuracy, strength)
	}
}

func TestTyrantsRingGrantsOneHundredMeleeStrength(t *testing.T) {
	player := &playerState{
		equippedArmour: make(map[ArmourSlot]string),
		ringRack: inventoryState{capacity: RingRackCapacity, items: []InventoryItem{
			{Slot: 0, ItemID: ItemTyrantsRing, Quantity: 1, InstanceID: "tyrant"},
		}},
	}
	accuracy, strength := equipmentOffenceForWeapon(player, WeaponDefinition{MeleeStrength: 1})
	if accuracy != 0 || strength != 100 {
		t.Fatalf("Tyrant's Ring bonuses = accuracy %d, strength %d; want 0, 100", accuracy, strength)
	}
}
