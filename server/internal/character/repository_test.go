package character

import (
	"context"
	"testing"

	"github.com/mywyau/arcana/server/internal/simulation"
)

func TestMemoryRepositoryCopiesInventoryState(t *testing.T) {
	repository := NewMemoryRepository()
	state := simulation.DefaultPersistentPlayerState()
	if err := repository.Save(context.Background(), "player-one", state); err != nil {
		t.Fatalf("Save() returned an unexpected error: %v", err)
	}

	state.Inventory.Items[0].Quantity = 99
	state.Skills[0].Experience = 99
	loaded, found, err := repository.Load(context.Background(), "player-one")
	if err != nil || !found {
		t.Fatalf("Load() = found %t, error %v", found, err)
	}
	if loaded.Inventory.Items[0].Quantity != 1 {
		t.Fatalf("stored quantity = %d, want 1", loaded.Inventory.Items[0].Quantity)
	}
	if loaded.Skills[0].Experience != 0 {
		t.Fatalf("stored Attack XP = %d, want 0", loaded.Skills[0].Experience)
	}

	loaded.Inventory.Items[0].Quantity = 42
	loaded.Skills[0].Experience = 42
	reloaded, _, _ := repository.Load(context.Background(), "player-one")
	if reloaded.Inventory.Items[0].Quantity != 1 {
		t.Fatalf("reloaded quantity = %d, want 1", reloaded.Inventory.Items[0].Quantity)
	}
	if reloaded.Skills[0].Experience != 0 {
		t.Fatalf("reloaded Attack XP = %d, want 0", reloaded.Skills[0].Experience)
	}
}
