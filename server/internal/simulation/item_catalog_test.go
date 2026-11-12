package simulation

import (
	"strings"
	"testing"
)

func TestItemCatalogContainsExistingGameplayItems(t *testing.T) {
	tests := []struct {
		itemID string
		kind   ItemKind
	}{
		{string(WeaponUnarmed), ItemKindWeapon},
		{string(WeaponSword), ItemKindWeapon},
		{string(WeaponBronzeSword), ItemKindWeapon},
		{string(WeaponSteelSword), ItemKindWeapon},
		{string(WeaponMithrilSword), ItemKindWeapon},
		{string(WeaponRuneSword), ItemKindWeapon},
		{string(WeaponDagger), ItemKindWeapon},
		{string(WeaponBow), ItemKindWeapon},
		{string(WeaponOakShortbow), ItemKindWeapon}, {string(WeaponWillowShortbow), ItemKindWeapon},
		{string(WeaponMapleShortbow), ItemKindWeapon}, {string(WeaponYewShortbow), ItemKindWeapon},
		{string(WeaponMace), ItemKindWeapon},
		{string(WeaponBattleAxe), ItemKindWeapon},
		{string(WeaponWand), ItemKindWeapon},
		{string(WeaponMoonspunWand), ItemKindWeapon},
		{string(WeaponStaff), ItemKindWeapon},
		{string(WeaponEnchantedStaff), ItemKindWeapon},
		{string(WeaponStardewStaff), ItemKindWeapon},
		{string(WeaponSpear), ItemKindWeapon},
		{string(WeaponHalberd), ItemKindWeapon},
		{string(WeaponCrossbow), ItemKindWeapon},
		{ItemTrainingHelm, ItemKindArmour},
		{ItemBronzeHelm, ItemKindArmour},
		{ItemBronzePlate, ItemKindArmour},
		{ItemBronzeShield, ItemKindArmour}, {ItemBronzeLegs, ItemKindArmour}, {ItemBronzeGloves, ItemKindArmour}, {ItemBronzeBoots, ItemKindArmour},
		{ItemIronHelm, ItemKindArmour},
		{ItemIronPlate, ItemKindArmour},
		{ItemIronShield, ItemKindArmour}, {ItemIronLegs, ItemKindArmour}, {ItemIronGloves, ItemKindArmour}, {ItemIronBoots, ItemKindArmour},
		{ItemSteelHelm, ItemKindArmour}, {ItemSteelPlate, ItemKindArmour},
		{ItemSteelShield, ItemKindArmour}, {ItemSteelLegs, ItemKindArmour}, {ItemSteelGloves, ItemKindArmour}, {ItemSteelBoots, ItemKindArmour},
		{ItemMithrilHelm, ItemKindArmour},
		{ItemMithrilPlate, ItemKindArmour},
		{ItemMithrilShield, ItemKindArmour}, {ItemMithrilLegs, ItemKindArmour}, {ItemMithrilGloves, ItemKindArmour}, {ItemMithrilBoots, ItemKindArmour},
		{ItemAdamantiteHelm, ItemKindArmour}, {ItemAdamantitePlate, ItemKindArmour},
		{ItemAdamantiteShield, ItemKindArmour}, {ItemAdamantiteLegs, ItemKindArmour}, {ItemAdamantiteGloves, ItemKindArmour}, {ItemAdamantiteBoots, ItemKindArmour},
		{ItemTrainingPlate, ItemKindArmour},
		{ItemSwiftHelm, ItemKindArmour},
		{ItemSwiftBody, ItemKindArmour},
		{ItemSwiftBoots, ItemKindArmour},
		{ItemNoviceRobeHood, ItemKindArmour}, {ItemNoviceRobeTop, ItemKindArmour}, {ItemNoviceRobeGloves, ItemKindArmour}, {ItemNoviceRobeBottoms, ItemKindArmour}, {ItemNoviceRobeBoots, ItemKindArmour},
		{ItemMageRobeHood, ItemKindArmour}, {ItemMageRobeTop, ItemKindArmour}, {ItemMageRobeGloves, ItemKindArmour}, {ItemMageRobeBottoms, ItemKindArmour}, {ItemMageRobeBoots, ItemKindArmour},
		{ItemMoonspunHood, ItemKindArmour}, {ItemMoonspunTop, ItemKindArmour}, {ItemMoonspunGloves, ItemKindArmour}, {ItemMoonspunBottoms, ItemKindArmour}, {ItemMoonspunBoots, ItemKindArmour},
		{ItemEnchantedHood, ItemKindArmour}, {ItemEnchantedTop, ItemKindArmour}, {ItemEnchantedGloves, ItemKindArmour}, {ItemEnchantedBottoms, ItemKindArmour}, {ItemEnchantedBoots, ItemKindArmour},
		{ItemStardewHood, ItemKindArmour}, {ItemStardewTop, ItemKindArmour}, {ItemStardewGloves, ItemKindArmour}, {ItemStardewBottoms, ItemKindArmour}, {ItemStardewBoots, ItemKindArmour},
		{ItemNoviceRangerHood, ItemKindArmour}, {ItemNoviceRangerBody, ItemKindArmour}, {ItemNoviceRangerGloves, ItemKindArmour}, {ItemNoviceRangerLegs, ItemKindArmour}, {ItemNoviceRangerBoots, ItemKindArmour},
		{ItemScoutHood, ItemKindArmour}, {ItemScoutBody, ItemKindArmour}, {ItemScoutGloves, ItemKindArmour}, {ItemScoutLegs, ItemKindArmour}, {ItemScoutBoots, ItemKindArmour},
		{ItemTrackerHood, ItemKindArmour}, {ItemTrackerBody, ItemKindArmour}, {ItemTrackerGloves, ItemKindArmour}, {ItemTrackerLegs, ItemKindArmour}, {ItemTrackerBoots, ItemKindArmour},
		{ItemHunterHood, ItemKindArmour}, {ItemHunterBody, ItemKindArmour}, {ItemHunterGloves, ItemKindArmour}, {ItemHunterLegs, ItemKindArmour}, {ItemHunterBoots, ItemKindArmour},
		{ItemMarksmanHood, ItemKindArmour}, {ItemMarksmanBody, ItemKindArmour}, {ItemMarksmanGloves, ItemKindArmour}, {ItemMarksmanLegs, ItemKindArmour}, {ItemMarksmanBoots, ItemKindArmour},
		{ItemVanguardCape, ItemKindArmour}, {ItemWayfarerCape, ItemKindArmour}, {ItemArcanistCape, ItemKindArmour},
		{ItemAmuletFerocity, ItemKindArmour},
		{ItemAmuletHawk, ItemKindArmour},
		{ItemAmuletEmbers, ItemKindArmour},
		{ItemAmuletFury, ItemKindArmour},
		{ItemDiamondAmulet, ItemKindArmour},
		{ItemRubyRing, ItemKindArmour}, {ItemEmeraldRing, ItemKindArmour},
		{ItemSapphireRing, ItemKindArmour}, {ItemDiamondRing, ItemKindArmour},
		{ItemTyrantsRing, ItemKindArmour}, {ItemRangedTyrantsRing, ItemKindArmour}, {ItemMageTyrantsRing, ItemKindArmour},
		{ItemStandardSpellbook, ItemKindSpellbook},
		{ItemSkinBible, ItemKindSpellbook},
		{ItemHolyBook, ItemKindSpellbook},
		{ItemClericsTome, ItemKindSpellbook},
		{ItemAirRune, ItemKindResource},
		{ItemWaterRune, ItemKindResource},
		{ItemEarthRune, ItemKindResource},
		{ItemFireRune, ItemKindResource},
		{ItemHolyRune, ItemKindResource},
		{ItemDarkRune, ItemKindResource},
		{ItemFocusRune, ItemKindResource},
		{ItemBlankRune, ItemKindResource},
		{ItemLeather, ItemKindResource},
		{ItemGoldOre, ItemKindResource},
		{ItemSilverOre, ItemKindResource}, {ItemClay, ItemKindResource}, {ItemAdamantiteOre, ItemKindResource},
		{ItemUncutSapphire, ItemKindResource}, {ItemSapphire, ItemKindResource},
		{ItemUncutEmerald, ItemKindResource}, {ItemEmerald, ItemKindResource},
		{ItemUncutRuby, ItemKindResource}, {ItemRuby, ItemKindResource},
		{ItemUncutDiamond, ItemKindResource}, {ItemDiamond, ItemKindResource},
		{ItemIronArrow, ItemKindResource},
		{ItemBronzeArrow, ItemKindResource}, {ItemSteelArrow, ItemKindResource},
		{ItemMithrilArrow, ItemKindResource}, {ItemAdamantiteArrow, ItemKindResource},
		{ItemIronBolt, ItemKindResource},
		{ItemBronzeBolt, ItemKindResource}, {ItemSteelBolt, ItemKindResource},
		{ItemMithrilBolt, ItemKindResource}, {ItemAdamantiteBolt, ItemKindResource},
		{ItemBones, ItemKindResource},
		{ItemCoins, ItemKindResource},
		{ItemBronzePickaxe, ItemKindResource},
		{ItemIronPickaxe, ItemKindResource}, {ItemSteelPickaxe, ItemKindResource}, {ItemMithrilPickaxe, ItemKindResource},
		{ItemBronzeAxe, ItemKindResource},
		{ItemIronAxe, ItemKindResource}, {ItemSteelAxe, ItemKindResource}, {ItemMithrilAxe, ItemKindResource},
		{ItemKnife, ItemKindResource},
		{ItemChisel, ItemKindResource},
		{ItemSewingNeedle, ItemKindResource},
		{ItemThread, ItemKindResource},
		{ItemTinderbox, ItemKindResource},
		{ItemHammer, ItemKindResource},
		{ItemRingMould, ItemKindResource}, {ItemAmuletMould, ItemKindResource},
		{ItemPrayerPotion, ItemKindConsumable},
		{ItemFishingRod, ItemKindResource},
		{ItemLobsterCage, ItemKindResource},
		{ItemVialOfWater, ItemKindResource},
		{ItemSunleaf, ItemKindResource},
		{ItemMoonleaf, ItemKindResource},
		{ItemSacredDust, ItemKindResource},
		{ItemFishOil, ItemKindResource},
		{ItemMinorHealingPotion, ItemKindConsumable},
		{ItemMurkyBrew, ItemKindConsumable},
		{ItemMeleePotion, ItemKindConsumable}, {ItemRangingPotion, ItemKindConsumable}, {ItemMagicPotion, ItemKindConsumable},
		{ItemArrowShafts, ItemKindResource},
		{ItemFeathers, ItemKindResource},
		{ItemIronArrowTips, ItemKindResource},
		{ItemBronzeArrowTips, ItemKindResource}, {ItemSteelArrowTips, ItemKindResource},
		{ItemMithrilArrowTips, ItemKindResource}, {ItemAdamantiteArrowTips, ItemKindResource},
		{ItemBowstring, ItemKindResource}, {ItemUnstrungShortbow, ItemKindResource},
		{ItemUnstrungOakShortbow, ItemKindResource}, {ItemUnstrungWillowShortbow, ItemKindResource},
		{ItemUnstrungMapleShortbow, ItemKindResource}, {ItemUnstrungYewShortbow, ItemKindResource},
		{ItemNormalLogs, ItemKindResource},
		{ItemOakLogs, ItemKindResource},
		{ItemWillowLogs, ItemKindResource},
		{ItemMapleLogs, ItemKindResource},
		{ItemYewLogs, ItemKindResource},
		{ItemCopperOre, ItemKindResource},
		{ItemTinOre, ItemKindResource},
		{ItemIronOre, ItemKindResource},
		{ItemCoal, ItemKindResource},
		{ItemMithrilOre, ItemKindResource},
		{ItemBronzeBar, ItemKindResource},
		{ItemIronBar, ItemKindResource},
		{ItemMithrilBar, ItemKindResource},
		{ItemGoldBar, ItemKindResource},
		{ItemSilverBar, ItemKindResource}, {ItemSteelBar, ItemKindResource}, {ItemAdamantiteBar, ItemKindResource},
		{ItemCookedMinnow, ItemKindConsumable},
		{ItemRawMinnow, ItemKindResource},
		{ItemRawTrout, ItemKindResource},
		{ItemCookedTrout, ItemKindConsumable},
		{ItemRawSalmon, ItemKindResource},
		{ItemCookedSalmon, ItemKindConsumable},
		{ItemRawTuna, ItemKindResource},
		{ItemCookedTuna, ItemKindConsumable},
		{ItemRawLobster, ItemKindResource},
		{ItemCookedLobster, ItemKindConsumable},
	}
	if got, want := len(itemCatalog), len(tests); got != want {
		t.Fatalf("item catalogue size = %d, want %d", got, want)
	}
	for _, test := range tests {
		definition, exists := itemDefinition(test.itemID)
		if !exists || definition.Kind != test.kind || definition.Name == "" {
			t.Errorf("itemDefinition(%q) = %+v, %t; want kind %q", test.itemID, definition, exists, test.kind)
		}
	}
	if unarmed, exists := itemDefinition(string(WeaponUnarmed)); !exists || !unarmed.Virtual {
		t.Fatalf("Unarmed definition = %+v, %t; want virtual item", unarmed, exists)
	}
	for _, itemID := range []string{ItemNormalLogs, ItemOakLogs, ItemWillowLogs, ItemMapleLogs, ItemYewLogs} {
		if definition, _ := itemDefinition(itemID); definition.StackLimit != 10 {
			t.Errorf("%s stack limit = %d, want 10", itemID, definition.StackLimit)
		}
	}
	for _, itemID := range []string{ItemKnife, ItemChisel, ItemSewingNeedle, ItemTinderbox, ItemHammer, ItemFishingRod, ItemLobsterCage} {
		if definition, _ := itemDefinition(itemID); definition.StackLimit != 1 {
			t.Errorf("%s stack limit = %d, want 1", itemID, definition.StackLimit)
		}
	}
	toolBonuses := map[string]float64{
		ItemBronzePickaxe: 0, ItemIronPickaxe: 0.05, ItemSteelPickaxe: 0.10, ItemMithrilPickaxe: 0.15,
		ItemBronzeAxe: 0, ItemIronAxe: 0.05, ItemSteelAxe: 0.10, ItemMithrilAxe: 0.15,
	}
	for itemID, wantBonus := range toolBonuses {
		definition, _ := itemDefinition(itemID)
		if definition.StackLimit != 1 || definition.GatheringTool.Family == "" || definition.GatheringTool.SuccessBonus != wantBonus {
			t.Errorf("gathering tool %q = %+v, want bonus %.2f", itemID, definition, wantBonus)
		}
	}
	if definition, _ := itemDefinition(ItemThread); definition.StackLimit != 100 {
		t.Errorf("%s stack limit = %d, want 100", ItemThread, definition.StackLimit)
	}
	if definition, _ := itemDefinition(ItemArrowShafts); definition.StackLimit != MaximumItemStackQuantity {
		t.Errorf("%s stack limit = %d, want %d", ItemArrowShafts, definition.StackLimit, MaximumItemStackQuantity)
	}
	if definition, _ := itemDefinition(ItemPrayerPotion); definition.StackLimit != 5 || definition.Consumable.PrayerRestoration != 25 {
		t.Errorf("prayer potion definition = %+v, want five-dose stacks restoring 25 prayer", definition)
	}
}

func TestGemAmuletsUsePlayerFacingGemNames(t *testing.T) {
	wantNames := map[string]string{
		ItemAmuletFerocity: "Ruby Amulet",
		ItemAmuletHawk:     "Emerald Amulet",
		ItemAmuletEmbers:   "Sapphire Amulet",
		ItemAmuletFury:     "Amulet of Fury",
		ItemDiamondAmulet:  "Diamond Amulet",
	}
	for itemID, want := range wantNames {
		definition, exists := itemDefinition(itemID)
		if !exists || definition.Name != want {
			t.Errorf("itemDefinition(%q).Name = %q, want %q", itemID, definition.Name, want)
		}
	}
}

func TestEquipmentProgressionUsesFiveLevelTiers(t *testing.T) {
	tests := []struct {
		itemIDs []string
		skill   SkillID
		level   int
	}{
		{[]string{string(WeaponBronzeSword)}, SkillAttack, 1},
		{[]string{string(WeaponSword)}, SkillAttack, 5},
		{[]string{string(WeaponSteelSword)}, SkillAttack, 10},
		{[]string{string(WeaponMithrilSword)}, SkillAttack, 15},
		{[]string{string(WeaponRuneSword)}, SkillAttack, 25},
		{[]string{string(WeaponBow)}, SkillRange, 1},
		{[]string{string(WeaponOakShortbow)}, SkillRange, 5},
		{[]string{string(WeaponWillowShortbow)}, SkillRange, 10},
		{[]string{string(WeaponMapleShortbow)}, SkillRange, 15},
		{[]string{string(WeaponYewShortbow)}, SkillRange, 20},
		{[]string{ItemBronzeHelm, ItemBronzePlate, ItemBronzeShield, ItemBronzeLegs, ItemBronzeGloves, ItemBronzeBoots}, SkillDefence, 1},
		{[]string{ItemIronHelm, ItemIronPlate, ItemIronShield, ItemIronLegs, ItemIronGloves, ItemIronBoots}, SkillDefence, 5},
		{[]string{ItemSteelHelm, ItemSteelPlate, ItemSteelShield, ItemSteelLegs, ItemSteelGloves, ItemSteelBoots}, SkillDefence, 10},
		{[]string{ItemMithrilHelm, ItemMithrilPlate, ItemMithrilShield, ItemMithrilLegs, ItemMithrilGloves, ItemMithrilBoots}, SkillDefence, 15},
		{[]string{ItemAdamantiteHelm, ItemAdamantitePlate, ItemAdamantiteShield, ItemAdamantiteLegs, ItemAdamantiteGloves, ItemAdamantiteBoots}, SkillDefence, 20},
		{[]string{ItemBronzeArrow, ItemBronzeBolt}, SkillRange, 1},
		{[]string{ItemIronArrow, ItemIronBolt}, SkillRange, 5},
		{[]string{ItemSteelArrow, ItemSteelBolt}, SkillRange, 10},
		{[]string{ItemMithrilArrow, ItemMithrilBolt}, SkillRange, 15},
		{[]string{ItemAdamantiteArrow, ItemAdamantiteBolt}, SkillRange, 20},
		{[]string{ItemBronzePickaxe}, SkillMining, 1},
		{[]string{ItemIronPickaxe}, SkillMining, 5},
		{[]string{ItemSteelPickaxe}, SkillMining, 10},
		{[]string{ItemMithrilPickaxe}, SkillMining, 15},
		{[]string{ItemBronzeAxe}, SkillWoodcutting, 1},
		{[]string{ItemIronAxe}, SkillWoodcutting, 5},
		{[]string{ItemSteelAxe}, SkillWoodcutting, 10},
		{[]string{ItemMithrilAxe}, SkillWoodcutting, 15},
	}

	for _, test := range tests {
		for _, itemID := range test.itemIDs {
			definition, exists := itemDefinition(itemID)
			if !exists || len(definition.EquipRequirements) != 1 ||
				definition.EquipRequirements[0] != (SkillRequirement{Skill: test.skill, Level: test.level}) {
				t.Errorf("item %q requirements = %+v, want %s level %d", itemID, definition.EquipRequirements, test.skill, test.level)
			}
		}
	}
}

func TestBuildItemCatalogRejectsDuplicateIDs(t *testing.T) {
	definitions := []ItemDefinition{
		{ID: "duplicate", Name: "First", Kind: ItemKindResource},
		{ID: "duplicate", Name: "Second", Kind: ItemKindResource},
	}
	if _, err := buildItemCatalog(definitions); err == nil || !strings.Contains(err.Error(), "duplicated") {
		t.Fatalf("buildItemCatalog() error = %v, want duplicate ID error", err)
	}
}

func TestBuildItemCatalogRejectsMalformedEquipment(t *testing.T) {
	tests := []struct {
		name       string
		definition ItemDefinition
	}{
		{
			name: "weapon without combat values",
			definition: ItemDefinition{
				ID: "broken_weapon", Name: "Broken Weapon", Kind: ItemKindWeapon,
				Weapon: WeaponDefinition{ID: "broken_weapon"},
			},
		},
		{
			name:       "armour without slot",
			definition: ItemDefinition{ID: "broken_armour", Name: "Broken Armour", Kind: ItemKindArmour},
		},
		{
			name: "armour with negative health bonus",
			definition: ItemDefinition{
				ID: "broken_health_armour", Name: "Broken Health Armour", Kind: ItemKindArmour,
				Armour: ArmourDefinition{Slot: ArmourSlotHead, HealthBonus: -1},
			},
		},
		{
			name:       "spellbook item without book",
			definition: ItemDefinition{ID: "broken_book", Name: "Broken Book", Kind: ItemKindSpellbook},
		},
		{
			name:       "consumable without healing",
			definition: ItemDefinition{ID: "broken_food", Name: "Broken Food", Kind: ItemKindConsumable},
		},
		{
			name: "armour with unknown set",
			definition: ItemDefinition{
				ID: "unknown_set_armour", Name: "Unknown Set Armour", Kind: ItemKindArmour,
				Armour: ArmourDefinition{Slot: ArmourSlotHead, Set: "missing_set"},
			},
		},
		{
			name: "equipment with invalid level requirement",
			definition: ItemDefinition{
				ID: "invalid_requirement", Name: "Invalid Requirement", Kind: ItemKindArmour,
				Armour:            ArmourDefinition{Slot: ArmourSlotHead},
				EquipRequirements: []SkillRequirement{{Skill: SkillAttack, Level: MaximumSkillLevel + 1}},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := buildItemCatalog([]ItemDefinition{test.definition}); err == nil {
				t.Fatal("buildItemCatalog() error = nil, want validation error")
			}
		})
	}
}

func TestInventorySnapshotRejectsUnknownCatalogItem(t *testing.T) {
	_, err := inventoryFromSnapshot(InventorySnapshot{
		Capacity: 1,
		Items:    []InventoryItem{{Slot: 0, ItemID: "unknown_item", Quantity: 1}},
	})
	if err == nil || !strings.Contains(err.Error(), "unknown item") {
		t.Fatalf("inventoryFromSnapshot() error = %v, want unknown item error", err)
	}
}

func TestCatalogWeaponStylesDriveDefaultsAndValidation(t *testing.T) {
	tests := []struct {
		weapon       WeaponID
		defaultStyle CombatStyleID
		allowed      CombatStyleID
		rejected     CombatStyleID
	}{
		{WeaponSword, CombatStyleStab, CombatStyleSlash, CombatStyleRapid},
		{WeaponUnarmed, CombatStyleCrush, CombatStyleHeavyCrush, CombatStyleStab},
		{WeaponSteelSword, CombatStyleStab, CombatStyleDefensive, CombatStyleRapid},
		{WeaponMithrilSword, CombatStyleStab, CombatStyleSlash, CombatStyleAccurate},
		{WeaponRuneSword, CombatStyleStab, CombatStyleSlash, CombatStyleAccurate},
		{WeaponDagger, CombatStyleStab, CombatStyleDefensive, CombatStyleLongRange},
		{WeaponBow, CombatStyleAccurate, CombatStyleRapid, CombatStyleSlash},
		{WeaponMace, CombatStyleCrush, CombatStyleHeavyCrush, CombatStyleStab},
		{WeaponBattleAxe, CombatStyleSlash, CombatStyleDefensive, CombatStyleRapid},
		{WeaponWand, CombatStyleAccurate, CombatStyleRapid, CombatStyleCrush},
		{WeaponMoonspunWand, CombatStyleAccurate, CombatStyleLongRange, CombatStyleCrush},
		{WeaponStaff, CombatStyleAccurate, CombatStyleLongRange, CombatStyleCrush},
		{WeaponEnchantedStaff, CombatStyleAccurate, CombatStyleRapid, CombatStyleCrush},
		{WeaponStardewStaff, CombatStyleAccurate, CombatStyleLongRange, CombatStyleCrush},
	}
	for _, test := range tests {
		if got := defaultCombatStyle(test.weapon); got != test.defaultStyle {
			t.Errorf("defaultCombatStyle(%q) = %q, want %q", test.weapon, got, test.defaultStyle)
		}
		if !combatStyleValidForWeapon(test.allowed, test.weapon) {
			t.Errorf("weapon %q does not allow configured styles", test.weapon)
		}
		if combatStyleValidForWeapon(test.rejected, test.weapon) {
			t.Errorf("weapon %q unexpectedly allows style %q", test.weapon, test.rejected)
		}
	}
}

func TestValidateWeaponStylesRejectsMalformedStyleMetadata(t *testing.T) {
	tests := []struct {
		name   string
		weapon WeaponDefinition
	}{
		{
			name:   "missing default",
			weapon: WeaponDefinition{CombatStyles: []CombatStyleID{CombatStyleStab}},
		},
		{
			name: "default not allowed",
			weapon: WeaponDefinition{
				DefaultStyle: CombatStyleSlash,
				CombatStyles: []CombatStyleID{CombatStyleStab},
			},
		},
		{
			name: "duplicate style",
			weapon: WeaponDefinition{
				DefaultStyle: CombatStyleStab,
				CombatStyles: []CombatStyleID{CombatStyleStab, CombatStyleStab},
			},
		},
		{
			name: "unknown style",
			weapon: WeaponDefinition{
				DefaultStyle: CombatStyleStab,
				CombatStyles: []CombatStyleID{CombatStyleStab, "unknown"},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := validateWeaponStyles(test.weapon); err == nil {
				t.Fatal("validateWeaponStyles() error = nil, want validation error")
			}
		})
	}
}

func TestWeaponFamiliesExposeOnlyTheirDamageTypes(t *testing.T) {
	tests := []struct {
		weapon WeaponID
		style  CombatStyleID
		want   DamageType
	}{
		{WeaponSword, CombatStyleStab, DamageTypeStab},
		{WeaponSword, CombatStyleSlash, DamageTypeSlash},
		{WeaponDagger, CombatStyleStab, DamageTypeStab},
		{WeaponDagger, CombatStyleSlash, DamageTypeSlash},
		{WeaponMace, CombatStyleCrush, DamageTypeCrush},
		{WeaponMace, CombatStyleHeavyCrush, DamageTypeCrush},
		{WeaponBattleAxe, CombatStyleSlash, DamageTypeSlash},
		{WeaponBattleAxe, CombatStyleHeavyCrush, DamageTypeCrush},
		{WeaponSpear, CombatStyleStab, DamageTypeStab},
		{WeaponSpear, CombatStyleSlash, DamageTypeSlash},
		{WeaponHalberd, CombatStyleSlash, DamageTypeSlash},
		{WeaponHalberd, CombatStyleStab, DamageTypeStab},
		{WeaponBow, CombatStyleAccurate, DamageTypeRanged},
		{WeaponCrossbow, CombatStyleAccurate, DamageTypeRanged},
	}
	for _, test := range tests {
		weapon, _ := weaponDefinition(test.weapon)
		if got, exists := weaponDamageTypeForStyle(weapon, test.style); !exists || got != test.want {
			t.Errorf("weaponDamageTypeForStyle(%q, %q) = %q, %t; want %q", test.weapon, test.style, got, exists, test.want)
		}
	}
	if combatStyleValidForWeapon(CombatStyleSlash, WeaponMace) {
		t.Fatal("mace unexpectedly allows slash")
	}
	if combatStyleValidForWeapon(CombatStyleCrush, WeaponSword) {
		t.Fatal("sword unexpectedly allows crush")
	}
}
