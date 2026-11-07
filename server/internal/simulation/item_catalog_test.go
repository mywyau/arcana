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
		{string(WeaponMace), ItemKindWeapon},
		{string(WeaponBattleAxe), ItemKindWeapon},
		{string(WeaponStaff), ItemKindWeapon},
		{ItemTrainingHelm, ItemKindArmour},
		{ItemBronzeHelm, ItemKindArmour},
		{ItemBronzePlate, ItemKindArmour},
		{ItemIronHelm, ItemKindArmour},
		{ItemIronPlate, ItemKindArmour},
		{ItemMithrilHelm, ItemKindArmour},
		{ItemMithrilPlate, ItemKindArmour},
		{ItemTrainingPlate, ItemKindArmour},
		{ItemSwiftHelm, ItemKindArmour},
		{ItemSwiftBody, ItemKindArmour},
		{ItemSwiftBoots, ItemKindArmour},
		{ItemAmuletFerocity, ItemKindArmour},
		{ItemAmuletHawk, ItemKindArmour},
		{ItemAmuletEmbers, ItemKindArmour},
		{ItemAmuletFury, ItemKindArmour},
		{ItemStandardSpellbook, ItemKindSpellbook},
		{ItemAirRune, ItemKindResource},
		{ItemIronArrow, ItemKindResource},
		{ItemCoins, ItemKindResource},
		{ItemBronzePickaxe, ItemKindResource},
		{ItemBronzeAxe, ItemKindResource},
		{ItemKnife, ItemKindResource},
		{ItemChisel, ItemKindResource},
		{ItemSewingNeedle, ItemKindResource},
		{ItemThread, ItemKindResource},
		{ItemTinderbox, ItemKindResource},
		{ItemHammer, ItemKindResource},
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
		{ItemNormalLogs, ItemKindResource},
		{ItemOakLogs, ItemKindResource},
		{ItemCopperOre, ItemKindResource},
		{ItemTinOre, ItemKindResource},
		{ItemIronOre, ItemKindResource},
		{ItemCoal, ItemKindResource},
		{ItemMithrilOre, ItemKindResource},
		{ItemBronzeBar, ItemKindResource},
		{ItemIronBar, ItemKindResource},
		{ItemMithrilBar, ItemKindResource},
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
		{trainingShardItem, ItemKindResource},
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
	for _, itemID := range []string{ItemNormalLogs, ItemOakLogs} {
		if definition, _ := itemDefinition(itemID); definition.StackLimit != 10 {
			t.Errorf("%s stack limit = %d, want 10", itemID, definition.StackLimit)
		}
	}
	for _, itemID := range []string{ItemKnife, ItemChisel, ItemSewingNeedle, ItemTinderbox, ItemHammer, ItemFishingRod, ItemLobsterCage} {
		if definition, _ := itemDefinition(itemID); definition.StackLimit != 1 {
			t.Errorf("%s stack limit = %d, want 1", itemID, definition.StackLimit)
		}
	}
	if definition, _ := itemDefinition(ItemThread); definition.StackLimit != 100 {
		t.Errorf("%s stack limit = %d, want 100", ItemThread, definition.StackLimit)
	}
	if definition, _ := itemDefinition(ItemPrayerPotion); definition.StackLimit != 5 || definition.Consumable.PrayerRestoration != 25 {
		t.Errorf("prayer potion definition = %+v, want five-dose stacks restoring 25 prayer", definition)
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
		{WeaponStaff, CombatStyleAccurate, CombatStyleLongRange, CombatStyleCrush},
	}
	for _, test := range tests {
		if got := defaultCombatStyle(test.weapon); got != test.defaultStyle {
			t.Errorf("defaultCombatStyle(%q) = %q, want %q", test.weapon, got, test.defaultStyle)
		}
		if !combatStyleValidForWeapon(test.allowed, test.weapon) ||
			!combatStyleValidForWeapon(CombatStyleUnfocused, test.weapon) {
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
			weapon: WeaponDefinition{CombatStyles: []CombatStyleID{CombatStyleStab, CombatStyleUnfocused}},
		},
		{
			name: "default not allowed",
			weapon: WeaponDefinition{
				DefaultStyle: CombatStyleSlash,
				CombatStyles: []CombatStyleID{CombatStyleStab, CombatStyleUnfocused},
			},
		},
		{
			name: "duplicate style",
			weapon: WeaponDefinition{
				DefaultStyle: CombatStyleStab,
				CombatStyles: []CombatStyleID{CombatStyleStab, CombatStyleStab, CombatStyleUnfocused},
			},
		},
		{
			name: "unknown style",
			weapon: WeaponDefinition{
				DefaultStyle: CombatStyleStab,
				CombatStyles: []CombatStyleID{CombatStyleStab, "unknown", CombatStyleUnfocused},
			},
		},
		{
			name: "missing unfocused",
			weapon: WeaponDefinition{
				DefaultStyle: CombatStyleStab,
				CombatStyles: []CombatStyleID{CombatStyleStab, CombatStyleSlash},
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
