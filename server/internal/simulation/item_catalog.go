package simulation

import "fmt"

type ItemKind string

const (
	ItemKindWeapon   ItemKind = "weapon"
	ItemKindArmour   ItemKind = "armour"
	ItemKindResource ItemKind = "resource"
)

// ItemDefinition is the authoritative gameplay definition for one stable item
// ID. Definitions are currently declared in Go; the registry boundary allows
// them to move to validated data files without changing combat or inventory.
type ItemDefinition struct {
	ID      string
	Name    string
	Kind    ItemKind
	Virtual bool
	Weapon  WeaponDefinition
	Armour  ArmourDefinition
}

var itemCatalog = mustBuildItemCatalog([]ItemDefinition{
	{
		ID: string(WeaponUnarmed), Name: "Unarmed", Kind: ItemKindWeapon, Virtual: true,
		Weapon: weaponWithStyles(WeaponDefinition{ID: WeaponUnarmed, AccuracyBonus: 0, MeleeStrength: 1, DamageVariance: 0, AttackInterval: swordAttackIntervalPulses, Range: swordRange, ApproachRange: swordApproachRange}, CombatStyleCrush, CombatStyleCrush, CombatStyleHeavyCrush, CombatStyleDefensive),
	},
	{
		ID: string(WeaponSword), Name: "Sword", Kind: ItemKindWeapon,
		Weapon: weaponWithStyles(WeaponDefinition{ID: WeaponSword, AccuracyBonus: 10, MeleeStrength: swordMeleeStrength, DamageVariance: swordDamageVariance, AttackInterval: swordAttackIntervalPulses, Range: swordRange, ApproachRange: swordApproachRange}, CombatStyleStab, CombatStyleStab, CombatStyleSlash, CombatStyleDefensive),
	},
	{
		ID: string(WeaponSteelSword), Name: "Steel Sword", Kind: ItemKindWeapon,
		Weapon: weaponWithStyles(WeaponDefinition{ID: WeaponSteelSword, AccuracyBonus: 12, MeleeStrength: steelSwordMeleeStrength, DamageVariance: swordDamageVariance, AttackInterval: swordAttackIntervalPulses, Range: swordRange, ApproachRange: swordApproachRange}, CombatStyleStab, CombatStyleStab, CombatStyleSlash, CombatStyleDefensive),
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
		Weapon: weaponWithStyles(WeaponDefinition{ID: WeaponBow, AccuracyBonus: 12, RangedStrength: bowRangedStrength, DamageVariance: bowDamageVariance, AttackInterval: bowAttackIntervalPulses, Range: bowRange, ApproachRange: bowApproachRange}, CombatStyleAccurate, CombatStyleRapid, CombatStyleAccurate, CombatStyleLongRange),
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
	{ID: ItemAirRune, Name: "Air Rune", Kind: ItemKindResource},
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
		case ItemKindResource:
		default:
			return nil, fmt.Errorf("item %q has unknown kind %q", definition.ID, definition.Kind)
		}
		catalog[definition.ID] = definition
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

func itemIsEquipment(itemID string) bool {
	definition, exists := itemDefinition(itemID)
	return exists && !definition.Virtual && (definition.Kind == ItemKindWeapon || definition.Kind == ItemKindArmour)
}
