package character

import (
	"context"
	"testing"

	"github.com/mywyau/arcana/server/internal/simulation"
)

func TestMemoryRepositoryCopiesInventoryState(t *testing.T) {
	repository := NewMemoryRepository()
	state := simulation.DefaultPersistentPlayerState()
	state.Inventory.Items = []simulation.InventoryItem{{Slot: 0, ItemID: simulation.ItemAirRune, Quantity: 1}}
	state.Bank.Items = []simulation.InventoryItem{{Slot: 0, ItemID: simulation.ItemAirRune, Quantity: 100}}
	state.AmmunitionPouch.Items = []simulation.InventoryItem{{Slot: 0, ItemID: simulation.ItemIronArrow, Quantity: 50}}
	state.RingRack.Items = []simulation.InventoryItem{{Slot: 0, ItemID: simulation.ItemRubyRing, Quantity: 1, InstanceID: "ring-1"}}
	state.RunePouch.Items = []simulation.InventoryItem{{Slot: 0, ItemID: simulation.ItemAirRune, Quantity: 75}}
	state.EquippedArmourReinforcements[simulation.ArmourSlotHead] = 1
	if err := repository.Save(context.Background(), "player-one", state); err != nil {
		t.Fatalf("Save() returned an unexpected error: %v", err)
	}

	state.Inventory.Items[0].Quantity = 99
	state.Bank.Items[0].Quantity = 999
	state.AmmunitionPouch.Items[0].Quantity = 999
	state.RingRack.Items[0].ItemID = simulation.ItemEmeraldRing
	state.RunePouch.Items[0].Quantity = 999
	state.Skills[0].Experience = 99
	state.EquippedArmourReinforcements[simulation.ArmourSlotHead] = 0
	loaded, found, err := repository.Load(context.Background(), "player-one")
	if err != nil || !found {
		t.Fatalf("Load() = found %t, error %v", found, err)
	}
	if loaded.Inventory.Items[0].Quantity != 1 {
		t.Fatalf("stored quantity = %d, want 1", loaded.Inventory.Items[0].Quantity)
	}
	if loaded.Bank.Items[0].Quantity != 100 {
		t.Fatalf("stored bank quantity = %d, want 100", loaded.Bank.Items[0].Quantity)
	}
	if loaded.AmmunitionPouch.Items[0].Quantity != 50 {
		t.Fatalf("stored ammunition quantity = %d, want 50", loaded.AmmunitionPouch.Items[0].Quantity)
	}
	if loaded.RingRack.Items[0].ItemID != simulation.ItemRubyRing {
		t.Fatalf("stored ring = %q, want %q", loaded.RingRack.Items[0].ItemID, simulation.ItemRubyRing)
	}
	if loaded.RunePouch.Items[0].Quantity != 75 {
		t.Fatalf("stored rune quantity = %d, want 75", loaded.RunePouch.Items[0].Quantity)
	}
	if loaded.Skills[0].Experience != 0 {
		t.Fatalf("stored Attack XP = %d, want 0", loaded.Skills[0].Experience)
	}
	if loaded.EquippedArmourReinforcements[simulation.ArmourSlotHead] != 1 {
		t.Fatal("stored equipment reinforcement was not copied")
	}

	loaded.Inventory.Items[0].Quantity = 42
	loaded.Bank.Items[0].Quantity = 42
	loaded.AmmunitionPouch.Items[0].Quantity = 42
	loaded.RingRack.Items[0].ItemID = simulation.ItemEmeraldRing
	loaded.RunePouch.Items[0].Quantity = 42
	loaded.Skills[0].Experience = 42
	loaded.EquippedArmourReinforcements[simulation.ArmourSlotHead] = 0
	reloaded, _, _ := repository.Load(context.Background(), "player-one")
	if reloaded.Inventory.Items[0].Quantity != 1 {
		t.Fatalf("reloaded quantity = %d, want 1", reloaded.Inventory.Items[0].Quantity)
	}
	if reloaded.Bank.Items[0].Quantity != 100 {
		t.Fatalf("reloaded bank quantity = %d, want 100", reloaded.Bank.Items[0].Quantity)
	}
	if reloaded.AmmunitionPouch.Items[0].Quantity != 50 {
		t.Fatalf("reloaded ammunition quantity = %d, want 50", reloaded.AmmunitionPouch.Items[0].Quantity)
	}
	if reloaded.RingRack.Items[0].ItemID != simulation.ItemRubyRing {
		t.Fatalf("reloaded ring = %q, want %q", reloaded.RingRack.Items[0].ItemID, simulation.ItemRubyRing)
	}
	if reloaded.RunePouch.Items[0].Quantity != 75 {
		t.Fatalf("reloaded rune quantity = %d, want 75", reloaded.RunePouch.Items[0].Quantity)
	}
	if reloaded.Skills[0].Experience != 0 {
		t.Fatalf("reloaded Attack XP = %d, want 0", reloaded.Skills[0].Experience)
	}
	if reloaded.EquippedArmourReinforcements[simulation.ArmourSlotHead] != 1 {
		t.Fatal("reloaded equipment reinforcement was not copied")
	}
}

func TestMemoryRepositoryDeletesCharacter(t *testing.T) {
	repository := NewMemoryRepository()
	if err := repository.Save(
		context.Background(), "player-one", simulation.DefaultPersistentPlayerState(),
	); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if err := repository.Delete(context.Background(), "player-one"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, found, err := repository.Load(context.Background(), "player-one"); err != nil || found {
		t.Fatalf("Load() after delete = found %t, error %v", found, err)
	}
}
