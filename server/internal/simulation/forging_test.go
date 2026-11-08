package simulation

import (
	"errors"
	"strings"
	"testing"
)

func TestForgingRecipesCreateExpectedBronzeEquipment(t *testing.T) {
	tests := []struct {
		id     ForgingRecipeID
		bars   int
		output string
	}{
		{ForgingBronzeSword, 2, string(WeaponBronzeSword)},
		{ForgingBronzeHelm, 2, ItemBronzeHelm},
		{ForgingBronzePickaxe, 3, ItemBronzePickaxe},
	}
	for _, test := range tests {
		recipe, exists := forgingRecipe(test.id)
		if !exists || recipe.RequiredLevel != 1 || recipe.RewardItem != test.output ||
			len(recipe.Costs) != 1 || recipe.Costs[0] != (ItemCost{ItemID: ItemBronzeBar, Quantity: test.bars}) {
			t.Errorf("forgingRecipe(%q) = %+v, %t", test.id, recipe, exists)
		}
	}
}

func TestApplyForgingRecipeAtomicallyConsumesBarsAndGrantsXP(t *testing.T) {
	player := &playerState{inventory: newEmptyInventory(), skills: newDefaultSkills()}
	player.inventory.add(ItemBronzeBar, 2)
	recipe, _ := forgingRecipe(ForgingBronzeHelm)
	if err := applyForgingRecipe(player, recipe); err != nil {
		t.Fatalf("applyForgingRecipe() returned an unexpected error: %v", err)
	}
	if player.inventory.quantity(ItemBronzeBar) != 0 || player.inventory.quantity(ItemBronzeHelm) != 1 {
		t.Fatalf("inventory after forging = %+v", player.inventory)
	}
	if got := player.skills[SkillSmithing]; got != recipe.RewardXP {
		t.Fatalf("Smithing XP = %d, want %d", got, recipe.RewardXP)
	}
}

func TestGemAmuletRecipesUseGoldAndMatchingCutGem(t *testing.T) {
	tests := []struct {
		id          ForgingRecipeID
		gem, output string
	}{
		{ForgingRubyAmulet, ItemRuby, ItemAmuletFerocity},
		{ForgingEmeraldAmulet, ItemEmerald, ItemAmuletHawk},
		{ForgingSapphireAmulet, ItemSapphire, ItemAmuletEmbers},
		{ForgingDiamondAmulet, ItemDiamond, ItemAmuletFury},
	}
	for _, test := range tests {
		recipe, exists := forgingRecipe(test.id)
		if !exists || recipe.RewardItem != test.output || recipe.RequiredTool != ItemAmuletMould || len(recipe.Costs) != 2 ||
			recipe.Costs[0] != (ItemCost{ItemID: ItemGoldBar, Quantity: 1}) ||
			recipe.Costs[1] != (ItemCost{ItemID: test.gem, Quantity: 1}) {
			t.Errorf("forgingRecipe(%q) = %+v, %t", test.id, recipe, exists)
		}
	}
}

func TestGemRingRecipesUseReusableMouldGoldAndMatchingCutGem(t *testing.T) {
	tests := []struct {
		id          ForgingRecipeID
		gem, output string
	}{
		{ForgingRubyRing, ItemRuby, ItemRubyRing},
		{ForgingEmeraldRing, ItemEmerald, ItemEmeraldRing},
		{ForgingSapphireRing, ItemSapphire, ItemSapphireRing},
		{ForgingDiamondRing, ItemDiamond, ItemDiamondRing},
	}
	for _, test := range tests {
		recipe, exists := forgingRecipe(test.id)
		if !exists || recipe.RewardItem != test.output || recipe.RequiredTool != ItemRingMould || len(recipe.Costs) != 2 ||
			recipe.Costs[0] != (ItemCost{ItemID: ItemGoldBar, Quantity: 1}) ||
			recipe.Costs[1] != (ItemCost{ItemID: test.gem, Quantity: 1}) {
			t.Errorf("forgingRecipe(%q) = %+v, %t", test.id, recipe, exists)
		}
	}
}

func TestJewelleryMouldIsRequiredButNotConsumed(t *testing.T) {
	recipe, _ := forgingRecipe(ForgingSapphireRing)
	player := &playerState{inventory: newEmptyInventory(), skills: newDefaultSkills()}
	player.skills[SkillSmithing] = SkillExperienceForLevel(recipe.RequiredLevel)
	player.inventory.add(ItemGoldBar, 1)
	player.inventory.add(ItemSapphire, 1)
	if err := applyForgingRecipe(player, recipe); !errors.Is(err, errMissingForgingTool) {
		t.Fatalf("applyForgingRecipe() without mould error = %v", err)
	}
	player.inventory.add(ItemRingMould, 1)
	if err := applyForgingRecipe(player, recipe); err != nil {
		t.Fatalf("applyForgingRecipe() with mould returned error: %v", err)
	}
	if player.inventory.quantity(ItemRingMould) != 1 || player.inventory.quantity(ItemSapphireRing) != 1 {
		t.Fatalf("inventory after crafting ring = %+v", player.inventory)
	}
}

func TestForgeIronArrowTipsCreatesTwentyFromOneBar(t *testing.T) {
	player := &playerState{inventory: newEmptyInventory(), skills: newDefaultSkills()}
	player.skills[SkillSmithing] = SkillExperienceForLevel(2)
	player.inventory.add(ItemIronBar, 1)
	recipe, exists := forgingRecipe(ForgingIronArrowTips)
	if !exists {
		t.Fatal("iron arrow tip recipe is missing")
	}
	if err := applyForgingRecipe(player, recipe); err != nil {
		t.Fatalf("applyForgingRecipe() returned an unexpected error: %v", err)
	}
	if player.inventory.quantity(ItemIronBar) != 0 || player.inventory.quantity(ItemIronArrowTips) != 20 {
		t.Fatalf("inventory after forging arrow tips = %+v", player.inventory)
	}
	if got := player.skills[SkillSmithing]; got != SkillExperienceForLevel(2)+recipe.RewardXP {
		t.Fatalf("Smithing XP = %d, want %d", got, SkillExperienceForLevel(2)+recipe.RewardXP)
	}
}

func TestApplyForgingRecipeDoesNotConsumePartialInputs(t *testing.T) {
	player := &playerState{inventory: newEmptyInventory(), skills: newDefaultSkills()}
	player.inventory.add(ItemBronzeBar, 1)
	recipe, _ := forgingRecipe(ForgingBronzePickaxe)
	if err := applyForgingRecipe(player, recipe); err == nil || !strings.Contains(err.Error(), "inputs") {
		t.Fatalf("applyForgingRecipe() error = %v, want missing inputs", err)
	}
	if player.inventory.quantity(ItemBronzeBar) != 1 {
		t.Fatalf("failed forge consumed bars: %+v", player.inventory)
	}
}

func TestForgingRecipesCoverMetalEquipmentProgression(t *testing.T) {
	tests := []struct {
		id     ForgingRecipeID
		level  int
		bar    string
		bars   int
		output string
	}{
		{ForgingBronzePlate, 2, ItemBronzeBar, 5, ItemBronzePlate},
		{ForgingIronSword, 3, ItemIronBar, 2, string(WeaponSword)},
		{ForgingIronHelm, 3, ItemIronBar, 2, ItemIronHelm},
		{ForgingIronPlate, 4, ItemIronBar, 5, ItemIronPlate},
		{ForgingMithrilSword, 6, ItemMithrilBar, 3, string(WeaponMithrilSword)},
		{ForgingMithrilHelm, 6, ItemMithrilBar, 3, ItemMithrilHelm},
		{ForgingMithrilPlate, 7, ItemMithrilBar, 6, ItemMithrilPlate},
	}
	for _, test := range tests {
		recipe, exists := forgingRecipe(test.id)
		if !exists || recipe.RequiredLevel != test.level || recipe.RewardItem != test.output ||
			len(recipe.Costs) != 1 || recipe.Costs[0] != (ItemCost{ItemID: test.bar, Quantity: test.bars}) {
			t.Errorf("forgingRecipe(%q) = %+v, %t", test.id, recipe, exists)
		}
	}
}

func TestForgingRecipesCoverCompleteMetalArmourProgression(t *testing.T) {
	tests := []struct {
		id     ForgingRecipeID
		output string
	}{
		{ForgingBronzeShield, ItemBronzeShield}, {ForgingBronzeLegs, ItemBronzeLegs}, {ForgingBronzeGloves, ItemBronzeGloves}, {ForgingBronzeBoots, ItemBronzeBoots},
		{ForgingIronShield, ItemIronShield}, {ForgingIronLegs, ItemIronLegs}, {ForgingIronGloves, ItemIronGloves}, {ForgingIronBoots, ItemIronBoots},
		{ForgingMithrilShield, ItemMithrilShield}, {ForgingMithrilLegs, ItemMithrilLegs}, {ForgingMithrilGloves, ItemMithrilGloves}, {ForgingMithrilBoots, ItemMithrilBoots},
	}
	for _, test := range tests {
		recipe, exists := forgingRecipe(test.id)
		if !exists || recipe.RewardItem != test.output || recipe.RewardQuantity != 1 {
			t.Errorf("forgingRecipe(%q) = %+v, %t", test.id, recipe, exists)
		}
	}
}
