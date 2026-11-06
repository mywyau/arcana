package simulation

import "math"

type lootTableEntry struct {
	ItemID          string
	MinimumQuantity int
	MaximumQuantity int
	DropChance      float64
}

func lootTableForTarget(targetID string) []lootTableEntry {
	switch targetID {
	case upperDummyTargetID:
		return []lootTableEntry{{
			ItemID: ItemTrainingHelm, MinimumQuantity: 1, MaximumQuantity: 1, DropChance: 1,
		}}
	case hillWardenTargetID:
		return []lootTableEntry{
			{ItemID: ItemAirRune, MinimumQuantity: 5, MaximumQuantity: 10, DropChance: 1},
			{ItemID: ItemTrainingPlate, MinimumQuantity: 1, MaximumQuantity: 1, DropChance: 0.25},
		}
	default:
		return []lootTableEntry{{
			ItemID: trainingShardItem, MinimumQuantity: 1, MaximumQuantity: 1, DropChance: 1,
		}}
	}
}

func (w *World) spawnTargetLoot(targetID string, position Position) {
	for _, entry := range lootTableForTarget(targetID) {
		if entry.ItemID == "" || entry.MinimumQuantity <= 0 || entry.MaximumQuantity < entry.MinimumQuantity ||
			entry.DropChance <= 0 {
			continue
		}
		if entry.DropChance < 1 && w.rollSource.Float64() >= entry.DropChance {
			continue
		}
		quantity := entry.MinimumQuantity
		if entry.MaximumQuantity > entry.MinimumQuantity {
			roll := max(0, min(w.rollSource.Float64(), math.Nextafter(1, 0)))
			quantity += int(math.Floor(roll * float64(entry.MaximumQuantity-entry.MinimumQuantity+1)))
		}
		w.spawnDrop(position, entry.ItemID, quantity)
	}
}
