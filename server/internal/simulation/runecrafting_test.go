package simulation

import (
	"testing"
	"time"
)

func TestEveryRuneHasItsOwnAltar(t *testing.T) {
	expectedRunes := []string{
		ItemAirRune, ItemWaterRune, ItemEarthRune, ItemFireRune, ItemFocusRune,
		ItemNatureRune, ItemSpiritRune, ItemVitalRune, ItemHexRune, ItemHolyRune,
		ItemShadowRune, ItemAetherRune, ItemWrathRune,
	}
	if len(runecraftingAltars) != len(expectedRunes) {
		t.Fatalf("runecrafting altar count = %d, want %d", len(runecraftingAltars), len(expectedRunes))
	}
	spots := newGatheringSpots()
	seenRunes := make(map[string]string, len(expectedRunes))
	seenPositions := make(map[Position]string, len(expectedRunes))
	for altarID, altar := range runecraftingAltars {
		spot, exists := spots[altarID]
		if !exists {
			t.Errorf("altar %q has no gathering spot", altarID)
			continue
		}
		if spot.Station != craftingStationRunecrafting || spot.RewardItem != altar.RuneItem ||
			spot.RequiredLevel != altar.RequiredLevel || spot.RewardQuantity != altar.RewardQuantity ||
			len(spot.Costs) != 1 || spot.Costs[0] != (ItemCost{ItemID: ItemBlankRune, Quantity: 1}) {
			t.Errorf("altar spot %q = %+v, want definition %+v", altarID, spot, altar)
		}
		if previous, duplicate := seenRunes[altar.RuneItem]; duplicate {
			t.Errorf("rune %q is shared by altars %q and %q", altar.RuneItem, previous, altarID)
		}
		if previous, duplicate := seenPositions[altar.Position]; duplicate {
			t.Errorf("altars %q and %q share position %+v", previous, altarID, altar.Position)
		}
		seenRunes[altar.RuneItem] = altarID
		seenPositions[altar.Position] = altarID
	}
	for _, runeItem := range expectedRunes {
		if _, exists := seenRunes[runeItem]; !exists {
			t.Errorf("rune %q has no altar", runeItem)
		}
	}
}

func TestHighLevelRuneAltarEnforcesLevelAndCraftsItsRune(t *testing.T) {
	altar := runecraftingAltars[wrathRuneAltarID]
	world := NewWorldWithRollSource(fixedRollSource(0))
	if err := world.AddPlayer("player-one", altar.Position, 4); err != nil {
		t.Fatal(err)
	}
	player := world.players["player-one"]
	player.inventory.add(ItemBlankRune, 1)
	command := GatherCommand{PlayerID: "player-one", Sequence: 1, SpotID: altar.ID, Quantity: 1}
	if err := world.QueueGather(command); err == nil {
		t.Fatal("Wrath altar accepted a player below its Runecrafting requirement")
	}

	player.skills[SkillRunecrafting] = SkillExperienceForLevel(altar.RequiredLevel)
	command.Sequence++
	if err := world.QueueGather(command); err != nil {
		t.Fatalf("Wrath altar rejected an eligible player: %v", err)
	}
	world.Step(50 * time.Millisecond)
	world.Step(runecraftingInterval * actionPulseInterval)
	snapshot := mustPlayer(t, world, "player-one")
	if snapshot.Inventory.itemsQuantity(ItemBlankRune) != 0 ||
		snapshot.Inventory.itemsQuantity(ItemWrathRune) != altar.RewardQuantity {
		t.Fatalf("inventory after crafting Wrath Runes = %+v", snapshot.Inventory)
	}
	if skill := mustSkill(t, snapshot, SkillRunecrafting); skill.Experience != SkillExperienceForLevel(altar.RequiredLevel)+altar.RewardXP {
		t.Fatalf("Runecrafting after crafting Wrath Runes = %+v", skill)
	}
}

func TestRunecraftingOutputMultipliesEveryTenLevelsAndCaps(t *testing.T) {
	tests := []struct {
		currentLevel int
		want         int
	}{
		{currentLevel: 1, want: 1},
		{currentLevel: 10, want: 1},
		{currentLevel: 11, want: 2},
		{currentLevel: 21, want: 3},
		{currentLevel: 41, want: 5},
		{currentLevel: 99, want: 5},
	}
	for _, test := range tests {
		if got := runecraftingMultiplier(1, test.currentLevel); got != test.want {
			t.Errorf("multiplier at level %d = %d, want %d", test.currentLevel, got, test.want)
		}
	}
}

func TestHigherRunecraftingLevelProducesMultipleRunesPerRunestone(t *testing.T) {
	altar := runecraftingAltars[starterAirRuneAltarID]
	world := NewWorldWithRollSource(fixedRollSource(0))
	if err := world.AddPlayer("player-one", altar.Position, 4); err != nil {
		t.Fatal(err)
	}
	player := world.players["player-one"]
	player.skills[SkillRunecrafting] = SkillExperienceForLevel(21)
	player.inventory.add(ItemBlankRune, 1)
	if err := world.QueueGather(GatherCommand{
		PlayerID: "player-one", Sequence: 1, SpotID: altar.ID, Quantity: 1,
	}); err != nil {
		t.Fatal(err)
	}
	world.Step(50 * time.Millisecond)
	world.Step(runecraftingInterval * actionPulseInterval)
	snapshot := mustPlayer(t, world, "player-one")
	wantRunes := altar.RewardQuantity * 3
	if got := snapshot.Inventory.itemsQuantity(ItemAirRune); got != wantRunes {
		t.Fatalf("Air Runes at level 21 = %d, want %d", got, wantRunes)
	}
}
