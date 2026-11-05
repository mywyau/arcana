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
	loaded, found, err := repository.Load(context.Background(), "player-one")
	if err != nil || !found {
		t.Fatalf("Load() = found %t, error %v", found, err)
	}
	if loaded.Inventory.Items[0].Quantity != 1 {
		t.Fatalf("stored quantity = %d, want 1", loaded.Inventory.Items[0].Quantity)
	}

	loaded.Inventory.Items[0].Quantity = 42
	reloaded, _, _ := repository.Load(context.Background(), "player-one")
	if reloaded.Inventory.Items[0].Quantity != 1 {
		t.Fatalf("reloaded quantity = %d, want 1", reloaded.Inventory.Items[0].Quantity)
	}
}
