package simulation

import "time"

const starterToolRespawnTime = time.Minute

type starterToolSpawn struct {
	Position Position
	ItemID   string
}

var starterToolSpawns = []starterToolSpawn{
	{Position: Position{X: -58, Z: 22}, ItemID: ItemFishingRod},
	{Position: Position{X: -34, Z: 10}, ItemID: ItemLobsterCage},
	{Position: Position{X: 28, Z: 9}, ItemID: ItemBronzePickaxe},
	{Position: Position{X: -58, Z: -5}, ItemID: ItemBronzeAxe},
	{Position: Position{X: 38, Z: -5}, ItemID: ItemHammer},
	{Position: Position{X: -52, Z: -8}, ItemID: ItemTinderbox},
	{Position: Position{X: -46, Z: -8}, ItemID: ItemKnife},
	{Position: Position{X: 28, Z: 29}, ItemID: ItemChisel},
}

// AddStarterToolSpawns makes essential skilling tools available in every
// server mode, including hosted worlds where development chests are hidden.
func (w *World) AddStarterToolSpawns() error {
	for _, spawn := range starterToolSpawns {
		if err := w.AddRespawningWorldDrop(
			spawn.Position, spawn.ItemID, 1, starterToolRespawnTime,
		); err != nil {
			return err
		}
	}
	return nil
}
