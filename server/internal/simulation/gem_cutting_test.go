package simulation

import "testing"

func TestCutGemConsumesOneGemAndPreservesChisel(t *testing.T) {
	world := NewWorld()
	if err := world.AddPlayer("player-one", Position{}, 4); err != nil {
		t.Fatal(err)
	}
	player := world.players["player-one"]
	player.inventory.add(ItemChisel, 1)
	player.inventory.add(ItemUncutRuby, 2)
	if err := world.CutGem("player-one", 1, ItemUncutRuby); err != nil {
		t.Fatalf("CutGem() returned an unexpected error: %v", err)
	}
	if player.inventory.quantity(ItemChisel) != 1 || player.inventory.quantity(ItemUncutRuby) != 1 || player.inventory.quantity(ItemRuby) != 1 {
		t.Fatalf("inventory after cutting = %+v", player.inventory)
	}
}

func TestCutGemRequiresChisel(t *testing.T) {
	world := NewWorld()
	if err := world.AddPlayer("player-one", Position{}, 4); err != nil {
		t.Fatal(err)
	}
	world.players["player-one"].inventory.add(ItemUncutSapphire, 1)
	if err := world.CutGem("player-one", 1, ItemUncutSapphire); err == nil {
		t.Fatal("CutGem() error = nil without a chisel")
	}
}
