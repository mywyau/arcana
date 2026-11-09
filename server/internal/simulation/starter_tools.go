package simulation

import "time"

const starterToolRespawnTime = time.Minute

type starterToolSpawn struct {
	Position Position
	ItemID   string
}

var starterToolSpawns = []starterToolSpawn{
	{Position: Position{X: -11, Z: 14}, ItemID: ItemFishingRod},
	{Position: Position{X: -21, Z: -14}, ItemID: ItemLobsterCage},
	{Position: Position{X: 16, Z: 12}, ItemID: ItemBronzePickaxe},
	{Position: Position{X: -13, Z: -2}, ItemID: ItemBronzeAxe},
	{Position: Position{X: 7, Z: 8}, ItemID: ItemHammer},
	{Position: Position{X: -13, Z: -5}, ItemID: ItemTinderbox},
	{Position: Position{X: -13, Z: 1}, ItemID: ItemKnife},
	{Position: Position{X: 16, Z: 15}, ItemID: ItemChisel},
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
