package character

import (
	"context"
	"sync"

	"github.com/mywyau/arcana/server/internal/simulation"
)

// Repository stores the durable subset of a character independently from the
// active simulation world.
type Repository interface {
	Load(context.Context, string) (simulation.PersistentPlayerState, bool, error)
	Save(context.Context, string, simulation.PersistentPlayerState) error
}

type MemoryRepository struct {
	mu     sync.RWMutex
	states map[string]simulation.PersistentPlayerState
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{states: make(map[string]simulation.PersistentPlayerState)}
}

func (repository *MemoryRepository) Load(
	_ context.Context,
	playerID string,
) (simulation.PersistentPlayerState, bool, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	state, found := repository.states[playerID]
	return cloneState(state), found, nil
}

func (repository *MemoryRepository) Save(
	_ context.Context,
	playerID string,
	state simulation.PersistentPlayerState,
) error {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	repository.states[playerID] = cloneState(state)
	return nil
}

func cloneState(state simulation.PersistentPlayerState) simulation.PersistentPlayerState {
	state.Inventory.Items = append([]simulation.InventoryItem(nil), state.Inventory.Items...)
	state.Skills = append([]simulation.PersistentSkillState(nil), state.Skills...)
	armour := make(map[simulation.ArmourSlot]string, len(state.EquippedArmour))
	for slot, itemID := range state.EquippedArmour {
		armour[slot] = itemID
	}
	state.EquippedArmour = armour
	return state
}
