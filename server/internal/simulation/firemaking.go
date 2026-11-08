package simulation

import (
	"fmt"
	"sort"
	"time"
)

const (
	basicFireLifetime  = 60 * time.Second
	minimumFireSpacing = 1.5
)

var firemakingLogs = map[string]bool{
	ItemNormalLogs: true, ItemOakLogs: true, ItemWillowLogs: true,
	ItemMapleLogs: true, ItemYewLogs: true,
}

type FireSnapshot struct {
	ID       string
	Position Position
	LogItem  string
}

func (w *World) LightFire(playerID PlayerID, sequence uint64, logItem string) error {
	if sequence == 0 {
		return fmt.Errorf("firemaking sequence must be positive")
	}
	if !firemakingLogs[logItem] {
		return fmt.Errorf("item %q cannot be burned", logItem)
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	player, exists := w.players[playerID]
	if !exists {
		player, exists = w.pendingJoins[playerID]
	}
	if !exists || player.defeated {
		return fmt.Errorf("player %q cannot light a fire", playerID)
	}
	if sequence <= player.lastFireSequence {
		return fmt.Errorf("firemaking sequence %d is not newer than %d", sequence, player.lastFireSequence)
	}
	if !player.inventory.contains(ItemTinderbox) || !player.inventory.contains(logItem) {
		return fmt.Errorf("a tinderbox and %s are required", logItem)
	}
	for _, fire := range w.fires {
		if distance(player.position, fire.Position) < minimumFireSpacing {
			return fmt.Errorf("another fire is too close")
		}
	}
	candidate := cloneInventory(player.inventory)
	if !candidate.remove(logItem, 1) {
		return fmt.Errorf("a log is required")
	}
	w.nextFireID++
	id := fmt.Sprintf("fire-%d", w.nextFireID)
	w.fires[id] = FireSnapshot{ID: id, Position: player.position, LogItem: logItem}
	w.fireLifetimes[id] = basicFireLifetime
	player.inventory = candidate
	player.lastFireSequence = sequence
	return nil
}

func (w *World) advanceFireLifetimes(elapsed time.Duration) {
	for id, remaining := range w.fireLifetimes {
		remaining -= elapsed
		if remaining <= 0 {
			delete(w.fireLifetimes, id)
			delete(w.fires, id)
			continue
		}
		w.fireLifetimes[id] = remaining
	}
}

func (w *World) fireSnapshots() []FireSnapshot {
	ids := make([]string, 0, len(w.fires))
	for id := range w.fires {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	result := make([]FireSnapshot, 0, len(ids))
	for _, id := range ids {
		result = append(result, w.fires[id])
	}
	return result
}
