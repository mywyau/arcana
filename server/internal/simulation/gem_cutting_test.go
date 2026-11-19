package simulation

import "testing"

func TestEveryUncutGemHasACutOutput(t *testing.T) {
	want := map[string]string{
		ItemUncutSapphire:   ItemSapphire,
		ItemUncutEmerald:    ItemEmerald,
		ItemUncutRuby:       ItemRuby,
		ItemUncutDiamond:    ItemDiamond,
		ItemUncutOpal:       ItemOpal,
		ItemUncutTopaz:      ItemTopaz,
		ItemUncutAquamarine: ItemAquamarine,
		ItemUncutAmethyst:   ItemAmethyst,
		ItemUncutOnyx:       ItemOnyx,
	}
	if len(cutGemOutputs) != len(want) {
		t.Fatalf("cut gem output count = %d, want %d", len(cutGemOutputs), len(want))
	}
	for uncut, cut := range want {
		if got := cutGemOutputs[uncut]; got != cut {
			t.Errorf("cut output for %q = %q, want %q", uncut, got, cut)
		}
	}
}

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
