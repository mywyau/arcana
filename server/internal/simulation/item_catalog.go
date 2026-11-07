package simulation

import "fmt"

type ItemKind string

const (
	ItemKindWeapon     ItemKind = "weapon"
	ItemKindArmour     ItemKind = "armour"
	ItemKindSpellbook  ItemKind = "spellbook"
	ItemKindConsumable ItemKind = "consumable"
	ItemKindResource   ItemKind = "resource"

	ItemStandardSpellbook = "standard_spellbook"
	ItemIronArrow         = "iron_arrow"
	ItemCoins             = "coins"
	ItemBronzePickaxe     = "bronze_pickaxe"
	ItemCopperOre         = "copper_ore"
	ItemTinOre            = "tin_ore"
	ItemIronOre           = "iron_ore"
	ItemCoal              = "coal"
	ItemMithrilOre        = "mithril_ore"
	ItemCopperBar         = "copper_bar"
	ItemBronzeBar         = "bronze_bar"
	ItemIronBar           = "iron_bar"
	ItemMithrilBar        = "mithril_bar"
	ItemCookedMinnow      = "cooked_minnow"
	ItemRawMinnow         = "raw_minnow"
	ItemRawTrout          = "raw_trout"
	ItemCookedTrout       = "cooked_trout"
	ItemRawSalmon         = "raw_salmon"
	ItemCookedSalmon      = "cooked_salmon"
	ItemRawTuna           = "raw_tuna"
	ItemCookedTuna        = "cooked_tuna"
	ItemRawLobster        = "raw_lobster"
	ItemCookedLobster     = "cooked_lobster"
)

type ConsumableDefinition struct {
	Healing        int
	CooldownPulses uint64
}

// ItemDefinition is the authoritative gameplay definition for one stable item
// ID. Definitions are currently declared in Go; the registry boundary allows
// them to move to validated data files without changing combat or inventory.
type ItemDefinition struct {
	ID         string
	Name       string
	Kind       ItemKind
	Virtual    bool
	StackLimit int
	Weapon     WeaponDefinition
	Armour     ArmourDefinition
	Spellbook  SpellbookID
	Consumable ConsumableDefinition
}

var itemCatalog = mustBuildItemCatalog([]ItemDefinition{
	{
		ID: string(WeaponUnarmed), Name: "Unarmed", Kind: ItemKindWeapon, Virtual: true,
		Weapon: weaponWithStyles(WeaponDefinition{ID: WeaponUnarmed, AccuracyBonus: 0, MeleeStrength: 1, DamageVariance: 0, AttackInterval: swordAttackIntervalPulses, Range: swordRange, ApproachRange: swordApproachRange}, CombatStyleCrush, CombatStyleCrush, CombatStyleHeavyCrush, CombatStyleDefensive),
	},
	{
		ID: string(WeaponSword), Name: "Iron Sword", Kind: ItemKindWeapon,
		Weapon: weaponWithStyles(WeaponDefinition{ID: WeaponSword, AccuracyBonus: 10, MeleeStrength: swordMeleeStrength, DamageVariance: swordDamageVariance, AttackInterval: swordAttackIntervalPulses, Range: swordRange, ApproachRange: swordApproachRange}, CombatStyleStab, CombatStyleStab, CombatStyleSlash, CombatStyleDefensive),
	},
	{
		ID: string(WeaponBronzeSword), Name: "Bronze Sword", Kind: ItemKindWeapon,
		Weapon: weaponWithStyles(WeaponDefinition{ID: WeaponBronzeSword, AccuracyBonus: 8, MeleeStrength: 7, DamageVariance: swordDamageVariance, AttackInterval: swordAttackIntervalPulses, Range: swordRange, ApproachRange: swordApproachRange}, CombatStyleStab, CombatStyleStab, CombatStyleSlash, CombatStyleDefensive),
	},
	{
		ID: string(WeaponSteelSword), Name: "Steel Sword", Kind: ItemKindWeapon,
		Weapon: weaponWithStyles(WeaponDefinition{ID: WeaponSteelSword, AccuracyBonus: 12, MeleeStrength: steelSwordMeleeStrength, DamageVariance: swordDamageVariance, AttackInterval: swordAttackIntervalPulses, Range: swordRange, ApproachRange: swordApproachRange}, CombatStyleStab, CombatStyleStab, CombatStyleSlash, CombatStyleDefensive),
	},
	{
		ID: string(WeaponMithrilSword), Name: "Mithril Sword", Kind: ItemKindWeapon,
		Weapon: weaponWithStyles(WeaponDefinition{ID: WeaponMithrilSword, AccuracyBonus: 18, MeleeStrength: 18, DamageVariance: swordDamageVariance, AttackInterval: swordAttackIntervalPulses, Range: swordRange, ApproachRange: swordApproachRange}, CombatStyleStab, CombatStyleStab, CombatStyleSlash, CombatStyleDefensive),
	},
	{
		ID: string(WeaponRuneSword), Name: "Rune Sword", Kind: ItemKindWeapon,
		Weapon: weaponWithStyles(WeaponDefinition{ID: WeaponRuneSword, AccuracyBonus: 40, MeleeStrength: runeSwordMeleeStrength, DamageVariance: swordDamageVariance, AttackInterval: swordAttackIntervalPulses, Range: swordRange, ApproachRange: swordApproachRange}, CombatStyleStab, CombatStyleStab, CombatStyleSlash, CombatStyleDefensive),
	},
	{
		ID: string(WeaponDagger), Name: "Dagger", Kind: ItemKindWeapon,
		Weapon: weaponWithStyles(WeaponDefinition{ID: WeaponDagger, AccuracyBonus: 14, MeleeStrength: daggerMeleeStrength, DamageVariance: daggerDamageVariance, AttackInterval: daggerAttackIntervalPulses, Range: daggerRange, ApproachRange: daggerApproachRange}, CombatStyleStab, CombatStyleStab, CombatStyleSlash, CombatStyleDefensive),
	},
	{
		ID: string(WeaponBow), Name: "Bow", Kind: ItemKindWeapon,
		Weapon: weaponWithStyles(WeaponDefinition{ID: WeaponBow, AccuracyBonus: 12, RangedStrength: bowRangedStrength, DamageVariance: bowDamageVariance, AttackInterval: bowAttackIntervalPulses, Range: bowRange, ApproachRange: bowApproachRange, Ammunition: ItemCost{ItemID: ItemIronArrow, Quantity: 1}}, CombatStyleAccurate, CombatStyleRapid, CombatStyleAccurate, CombatStyleLongRange),
	},
	{
		ID: string(WeaponMace), Name: "Mace", Kind: ItemKindWeapon,
		Weapon: weaponWithStyles(WeaponDefinition{ID: WeaponMace, AccuracyBonus: 8, MeleeStrength: maceMeleeStrength, DamageVariance: maceDamageVariance, AttackInterval: maceAttackIntervalPulses, Range: maceRange, ApproachRange: maceApproachRange}, CombatStyleCrush, CombatStyleCrush, CombatStyleHeavyCrush, CombatStyleDefensive),
	},
	{
		ID: string(WeaponBattleAxe), Name: "Battle Axe", Kind: ItemKindWeapon,
		Weapon: weaponWithStyles(WeaponDefinition{ID: WeaponBattleAxe, AccuracyBonus: 6, MeleeStrength: battleAxeMeleeStrength, DamageVariance: battleAxeDamageVariance, AttackInterval: battleAxeAttackIntervalPulses, Range: battleAxeRange, ApproachRange: battleAxeApproachRange}, CombatStyleSlash, CombatStyleSlash, CombatStyleHeavyCrush, CombatStyleDefensive),
	},
	{
		ID: string(WeaponStaff), Name: "Staff", Kind: ItemKindWeapon,
		Weapon: weaponWithStyles(WeaponDefinition{ID: WeaponStaff, AccuracyBonus: 11, MagicStrength: staffMagicStrength, DamageVariance: staffDamageVariance, AttackInterval: staffAttackIntervalPulses, Range: staffRange, ApproachRange: staffApproachRange}, CombatStyleAccurate, CombatStyleRapid, CombatStyleAccurate, CombatStyleLongRange),
	},
	{
		ID: ItemTrainingHelm, Name: "Training Helm", Kind: ItemKindArmour,
		Armour: ArmourDefinition{Slot: ArmourSlotHead, Defence: DefenceBonuses{Stab: 2, Slash: 3, Crush: 2, Ranged: 1}},
	},
	{
		ID: ItemBronzeHelm, Name: "Bronze Helm", Kind: ItemKindArmour,
		Armour: ArmourDefinition{Slot: ArmourSlotHead, Defence: DefenceBonuses{Stab: 2, Slash: 2, Crush: 3, Ranged: 1, Magic: -1}},
	},
	{
		ID: ItemBronzePlate, Name: "Bronze Platebody", Kind: ItemKindArmour,
		Armour: ArmourDefinition{Slot: ArmourSlotBody, Defence: DefenceBonuses{Stab: 5, Slash: 6, Crush: 4, Ranged: 3, Magic: -2}},
	},
	{
		ID: ItemIronHelm, Name: "Iron Helm", Kind: ItemKindArmour,
		Armour: ArmourDefinition{Slot: ArmourSlotHead, Defence: DefenceBonuses{Stab: 4, Slash: 5, Crush: 4, Ranged: 2, Magic: -1}},
	},
	{
		ID: ItemIronPlate, Name: "Iron Platebody", Kind: ItemKindArmour,
		Armour: ArmourDefinition{Slot: ArmourSlotBody, Defence: DefenceBonuses{Stab: 10, Slash: 12, Crush: 9, Ranged: 7, Magic: -3}},
	},
	{
		ID: ItemMithrilHelm, Name: "Mithril Helm", Kind: ItemKindArmour,
		Armour: ArmourDefinition{Slot: ArmourSlotHead, Defence: DefenceBonuses{Stab: 7, Slash: 8, Crush: 7, Ranged: 4, Magic: -2}},
	},
	{
		ID: ItemMithrilPlate, Name: "Mithril Platebody", Kind: ItemKindArmour,
		Armour: ArmourDefinition{Slot: ArmourSlotBody, Defence: DefenceBonuses{Stab: 16, Slash: 18, Crush: 14, Ranged: 11, Magic: -5}},
	},
	{
		ID: ItemTrainingPlate, Name: "Training Plate", Kind: ItemKindArmour,
		Armour: ArmourDefinition{Slot: ArmourSlotBody, Defence: DefenceBonuses{Stab: 8, Slash: 10, Crush: 7, Ranged: 6, Magic: -2}},
	},
	{
		ID: ItemSwiftHelm, Name: "Swift Helm", Kind: ItemKindArmour,
		Armour: ArmourDefinition{Slot: ArmourSlotHead, Set: armourSetSwift, Defence: DefenceBonuses{Stab: 1, Slash: 1, Crush: 1, Ranged: 2}},
	},
	{
		ID: ItemSwiftBody, Name: "Swift Body", Kind: ItemKindArmour,
		Armour: ArmourDefinition{Slot: ArmourSlotBody, Set: armourSetSwift, Defence: DefenceBonuses{Stab: 4, Slash: 4, Crush: 3, Ranged: 5, Magic: -1}},
	},
	{
		ID: ItemSwiftBoots, Name: "Swift Boots", Kind: ItemKindArmour,
		Armour: ArmourDefinition{Slot: ArmourSlotFeet, Set: armourSetSwift, Defence: DefenceBonuses{Stab: 1, Slash: 1, Crush: 1, Ranged: 2}},
	},
	{
		ID: ItemStandardSpellbook, Name: "Standard Spellbook", Kind: ItemKindSpellbook,
		Spellbook: SpellbookStandard,
	},
	{ID: ItemAirRune, Name: "Air Rune", Kind: ItemKindResource},
	{ID: ItemIronArrow, Name: "Iron Arrow", Kind: ItemKindResource},
	{ID: ItemCoins, Name: "Coins", Kind: ItemKindResource},
	{ID: ItemBronzePickaxe, Name: "Bronze Pickaxe", Kind: ItemKindResource},
	{ID: ItemCopperOre, Name: "Copper Ore", Kind: ItemKindResource},
	{ID: ItemTinOre, Name: "Tin Ore", Kind: ItemKindResource},
	{ID: ItemIronOre, Name: "Iron Ore", Kind: ItemKindResource},
	{ID: ItemCoal, Name: "Coal", Kind: ItemKindResource},
	{ID: ItemMithrilOre, Name: "Mithril Ore", Kind: ItemKindResource},
	{ID: ItemCopperBar, Name: "Copper Bar", Kind: ItemKindResource},
	{ID: ItemBronzeBar, Name: "Bronze Bar", Kind: ItemKindResource},
	{ID: ItemIronBar, Name: "Iron Bar", Kind: ItemKindResource},
	{ID: ItemMithrilBar, Name: "Mithril Bar", Kind: ItemKindResource},
	{
		ID: ItemCookedMinnow, Name: "Cooked Minnow", Kind: ItemKindConsumable,
		StackLimit: 10,
		Consumable: ConsumableDefinition{Healing: 8, CooldownPulses: 2},
	},
	{ID: ItemRawMinnow, Name: "Raw Minnow", Kind: ItemKindResource, StackLimit: 10},
	{ID: ItemRawTrout, Name: "Raw Trout", Kind: ItemKindResource, StackLimit: 10},
	{ID: ItemCookedTrout, Name: "Cooked Trout", Kind: ItemKindConsumable, StackLimit: 10, Consumable: ConsumableDefinition{Healing: 12, CooldownPulses: 2}},
	{ID: ItemRawSalmon, Name: "Raw Salmon", Kind: ItemKindResource, StackLimit: 10},
	{ID: ItemCookedSalmon, Name: "Cooked Salmon", Kind: ItemKindConsumable, StackLimit: 10, Consumable: ConsumableDefinition{Healing: 16, CooldownPulses: 2}},
	{ID: ItemRawTuna, Name: "Raw Tuna", Kind: ItemKindResource, StackLimit: 10},
	{ID: ItemCookedTuna, Name: "Cooked Tuna", Kind: ItemKindConsumable, StackLimit: 10, Consumable: ConsumableDefinition{Healing: 20, CooldownPulses: 2}},
	{ID: ItemRawLobster, Name: "Raw Lobster", Kind: ItemKindResource, StackLimit: 10},
	{ID: ItemCookedLobster, Name: "Cooked Lobster", Kind: ItemKindConsumable, StackLimit: 10, Consumable: ConsumableDefinition{Healing: 25, CooldownPulses: 2}},
	{ID: trainingShardItem, Name: "Training Shard", Kind: ItemKindResource},
})

func weaponWithStyles(definition WeaponDefinition, defaultStyle CombatStyleID, styles ...CombatStyleID) WeaponDefinition {
	definition.DefaultStyle = defaultStyle
	definition.CombatStyles = append(append([]CombatStyleID(nil), styles...), CombatStyleUnfocused)
	return definition
}

func buildItemCatalog(definitions []ItemDefinition) (map[string]ItemDefinition, error) {
	catalog := make(map[string]ItemDefinition, len(definitions))
	for index, definition := range definitions {
		if definition.ID == "" || definition.Name == "" {
			return nil, fmt.Errorf("item definition %d requires an ID and name", index)
		}
		if _, exists := catalog[definition.ID]; exists {
			return nil, fmt.Errorf("item ID %q is duplicated", definition.ID)
		}
		if definition.StackLimit < 0 {
			return nil, fmt.Errorf("item %q has a negative stack limit", definition.ID)
		}
		switch definition.Kind {
		case ItemKindWeapon:
			if string(definition.Weapon.ID) != definition.ID || definition.Weapon.AttackInterval == 0 ||
				definition.Weapon.Range <= 0 || definition.Weapon.ApproachRange <= 0 {
				return nil, fmt.Errorf("weapon %q has an invalid weapon definition", definition.ID)
			}
			if err := validateWeaponStyles(definition.Weapon); err != nil {
				return nil, fmt.Errorf("weapon %q: %w", definition.ID, err)
			}
		case ItemKindArmour:
			if definition.Armour.Slot == "" {
				return nil, fmt.Errorf("armour %q requires a slot", definition.ID)
			}
			if definition.Armour.Set != "" {
				if _, exists := armourSetDefinitions[definition.Armour.Set]; !exists {
					return nil, fmt.Errorf("armour %q references unknown set %q", definition.ID, definition.Armour.Set)
				}
			}
		case ItemKindSpellbook:
			if definition.Spellbook == "" {
				return nil, fmt.Errorf("spellbook item %q requires a spellbook ID", definition.ID)
			}
		case ItemKindConsumable:
			if definition.StackLimit == 0 || definition.Consumable.Healing <= 0 || definition.Consumable.CooldownPulses == 0 {
				return nil, fmt.Errorf("consumable %q requires a stack limit, positive healing, and cooldown", definition.ID)
			}
		case ItemKindResource:
		default:
			return nil, fmt.Errorf("item %q has unknown kind %q", definition.ID, definition.Kind)
		}
		catalog[definition.ID] = definition
	}
	for _, definition := range catalog {
		ammunition := definition.Weapon.Ammunition
		if ammunition.Quantity == 0 {
			continue
		}
		ammunitionItem, exists := catalog[ammunition.ItemID]
		if ammunition.ItemID == "" || ammunition.Quantity < 0 || !exists || ammunitionItem.Kind != ItemKindResource {
			return nil, fmt.Errorf("weapon %q has invalid ammunition %+v", definition.ID, ammunition)
		}
	}
	return catalog, nil
}

func validateWeaponStyles(weapon WeaponDefinition) error {
	if weapon.DefaultStyle == "" || len(weapon.CombatStyles) == 0 {
		return fmt.Errorf("requires a default and at least one combat style")
	}
	seen := make(map[CombatStyleID]struct{}, len(weapon.CombatStyles))
	defaultAllowed := false
	for _, style := range weapon.CombatStyles {
		if !knownCombatStyle(style) {
			return fmt.Errorf("contains unknown combat style %q", style)
		}
		if _, duplicate := seen[style]; duplicate {
			return fmt.Errorf("contains duplicate combat style %q", style)
		}
		seen[style] = struct{}{}
		defaultAllowed = defaultAllowed || style == weapon.DefaultStyle
	}
	if !defaultAllowed {
		return fmt.Errorf("default combat style %q is not allowed", weapon.DefaultStyle)
	}
	if _, allowsUnfocused := seen[CombatStyleUnfocused]; !allowsUnfocused {
		return fmt.Errorf("must allow combat style %q", CombatStyleUnfocused)
	}
	return nil
}

func knownCombatStyle(style CombatStyleID) bool {
	switch style {
	case CombatStyleSlash, CombatStyleStab, CombatStyleRapid, CombatStyleAccurate,
		CombatStyleLongRange, CombatStyleCrush, CombatStyleHeavyCrush,
		CombatStyleDefensive, CombatStyleUnfocused:
		return true
	default:
		return false
	}
}

func mustBuildItemCatalog(definitions []ItemDefinition) map[string]ItemDefinition {
	catalog, err := buildItemCatalog(definitions)
	if err != nil {
		panic(err)
	}
	return catalog
}

func itemDefinition(itemID string) (ItemDefinition, bool) {
	definition, exists := itemCatalog[itemID]
	return definition, exists
}

func itemStackLimit(itemID string) int {
	definition, exists := itemDefinition(itemID)
	if !exists {
		return 0
	}
	return definition.StackLimit
}
