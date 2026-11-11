package simulation

import "fmt"

const MaximumItemStackQuantity = 2_147_483_647

type ItemKind string

const (
	ItemKindWeapon     ItemKind = "weapon"
	ItemKindArmour     ItemKind = "armour"
	ItemKindSpellbook  ItemKind = "spellbook"
	ItemKindConsumable ItemKind = "consumable"
	ItemKindResource   ItemKind = "resource"

	ItemStandardSpellbook      = "standard_spellbook"
	ItemSkinBible              = "skin_bible"
	ItemHolyBook               = "holy_book"
	ItemClericsTome            = "clerics_tome"
	ItemBronzeArrow            = "bronze_arrow"
	ItemIronArrow              = "iron_arrow"
	ItemSteelArrow             = "steel_arrow"
	ItemMithrilArrow           = "mithril_arrow"
	ItemAdamantiteArrow        = "adamantite_arrow"
	ItemBronzeBolt             = "bronze_bolt"
	ItemIronBolt               = "iron_bolt"
	ItemSteelBolt              = "steel_bolt"
	ItemMithrilBolt            = "mithril_bolt"
	ItemAdamantiteBolt         = "adamantite_bolt"
	ItemWaterRune              = "water_rune"
	ItemEarthRune              = "earth_rune"
	ItemFireRune               = "fire_rune"
	ItemHolyRune               = "holy_rune"
	ItemDarkRune               = "dark_rune"
	ItemFocusRune              = "focus_rune"
	ItemBlankRune              = "blank_rune"
	ItemLeather                = "leather"
	ItemGoldOre                = "gold_ore"
	ItemSilverOre              = "silver_ore"
	ItemClay                   = "clay"
	ItemAdamantiteOre          = "adamantite_ore"
	ItemUncutSapphire          = "uncut_sapphire"
	ItemSapphire               = "sapphire"
	ItemUncutEmerald           = "uncut_emerald"
	ItemEmerald                = "emerald"
	ItemUncutRuby              = "uncut_ruby"
	ItemRuby                   = "ruby"
	ItemUncutDiamond           = "uncut_diamond"
	ItemDiamond                = "diamond"
	ItemGoldBar                = "gold_bar"
	ItemSilverBar              = "silver_bar"
	ItemSteelBar               = "steel_bar"
	ItemAdamantiteBar          = "adamantite_bar"
	ItemBones                  = "bones"
	ItemCoins                  = "coins"
	ItemBronzePickaxe          = "bronze_pickaxe"
	ItemBronzeAxe              = "bronze_axe"
	ItemKnife                  = "knife"
	ItemChisel                 = "chisel"
	ItemSewingNeedle           = "sewing_needle"
	ItemThread                 = "thread"
	ItemTinderbox              = "tinderbox"
	ItemHammer                 = "hammer"
	ItemRingMould              = "ring_mould"
	ItemAmuletMould            = "amulet_mould"
	ItemPrayerPotion           = "prayer_potion"
	ItemFishingRod             = "fishing_rod"
	ItemLobsterCage            = "lobster_cage"
	ItemVialOfWater            = "vial_of_water"
	ItemSunleaf                = "sunleaf"
	ItemMoonleaf               = "moonleaf"
	ItemSacredDust             = "sacred_dust"
	ItemFishOil                = "fish_oil"
	ItemMinorHealingPotion     = "minor_healing_potion"
	ItemMurkyBrew              = "murky_brew"
	ItemMeleePotion            = "melee_potion"
	ItemRangingPotion          = "ranging_potion"
	ItemMagicPotion            = "magic_potion"
	ItemArrowShafts            = "arrow_shafts"
	ItemFeathers               = "feathers"
	ItemBronzeArrowTips        = "bronze_arrow_tips"
	ItemIronArrowTips          = "iron_arrow_tips"
	ItemSteelArrowTips         = "steel_arrow_tips"
	ItemMithrilArrowTips       = "mithril_arrow_tips"
	ItemAdamantiteArrowTips    = "adamantite_arrow_tips"
	ItemBowstring              = "bowstring"
	ItemUnstrungShortbow       = "unstrung_shortbow"
	ItemUnstrungOakShortbow    = "unstrung_oak_shortbow"
	ItemUnstrungWillowShortbow = "unstrung_willow_shortbow"
	ItemUnstrungMapleShortbow  = "unstrung_maple_shortbow"
	ItemUnstrungYewShortbow    = "unstrung_yew_shortbow"
	ItemNormalLogs             = "normal_logs"
	ItemOakLogs                = "oak_logs"
	ItemWillowLogs             = "willow_logs"
	ItemMapleLogs              = "maple_logs"
	ItemYewLogs                = "yew_logs"
	ItemCopperOre              = "copper_ore"
	ItemTinOre                 = "tin_ore"
	ItemIronOre                = "iron_ore"
	ItemCoal                   = "coal"
	ItemMithrilOre             = "mithril_ore"
	ItemBronzeBar              = "bronze_bar"
	ItemIronBar                = "iron_bar"
	ItemMithrilBar             = "mithril_bar"
	ItemCookedMinnow           = "cooked_minnow"
	ItemRawMinnow              = "raw_minnow"
	ItemRawTrout               = "raw_trout"
	ItemCookedTrout            = "cooked_trout"
	ItemRawSalmon              = "raw_salmon"
	ItemCookedSalmon           = "cooked_salmon"
	ItemRawTuna                = "raw_tuna"
	ItemCookedTuna             = "cooked_tuna"
	ItemRawLobster             = "raw_lobster"
	ItemCookedLobster          = "cooked_lobster"
)

const (
	AmmunitionFamilyArrow = "arrow"
	AmmunitionFamilyBolt  = "bolt"
)

type ConsumableDefinition struct {
	Healing           int
	PrayerRestoration int
	CooldownPulses    uint64
	SkillBoosts       []SkillBoostDefinition
	BoostDuration     uint64
}

type SkillBoostDefinition struct {
	Skill  SkillID
	Levels int
}

type SkillRequirement struct {
	Skill SkillID
	Level int
}

// ItemDefinition is the authoritative gameplay definition for one stable item
// ID. Definitions are currently declared in Go; the registry boundary allows
// them to move to validated data files without changing combat or inventory.
type ItemDefinition struct {
	ID                 string
	Name               string
	Kind               ItemKind
	Virtual            bool
	StackLimit         int
	Ammunition         bool
	AmmunitionFamily   string
	AmmunitionStrength int
	EquipRequirements  []SkillRequirement
	Weapon             WeaponDefinition
	Armour             ArmourDefinition
	Spellbook          SpellbookID
	Consumable         ConsumableDefinition
}

var itemCatalog = mustBuildItemCatalog([]ItemDefinition{
	{
		ID: string(WeaponUnarmed), Name: "Unarmed", Kind: ItemKindWeapon, Virtual: true,
		Weapon: weaponWithStyles(WeaponDefinition{ID: WeaponUnarmed, AccuracyBonus: 0, MeleeStrength: 1, DamageVariance: 0, AttackInterval: swordAttackIntervalPulses, Range: swordRange, ApproachRange: swordApproachRange}, CombatStyleCrush,
			weaponStyle(CombatStyleCrush, DamageTypeCrush), weaponStyle(CombatStyleHeavyCrush, DamageTypeCrush), weaponStyle(CombatStyleDefensive, DamageTypeCrush)),
	},
	{
		ID: string(WeaponSword), Name: "Iron Sword", Kind: ItemKindWeapon,
		EquipRequirements: []SkillRequirement{{Skill: SkillAttack, Level: 5}},
		Weapon:            swordWeapon(WeaponDefinition{ID: WeaponSword, AccuracyBonus: 10, MeleeStrength: swordMeleeStrength, DamageVariance: swordDamageVariance, AttackInterval: swordAttackIntervalPulses, Range: swordRange, ApproachRange: swordApproachRange}),
	},
	{
		ID: string(WeaponBronzeSword), Name: "Bronze Sword", Kind: ItemKindWeapon,
		Weapon: swordWeapon(WeaponDefinition{ID: WeaponBronzeSword, AccuracyBonus: 8, MeleeStrength: 7, DamageVariance: swordDamageVariance, AttackInterval: swordAttackIntervalPulses, Range: swordRange, ApproachRange: swordApproachRange}),
	},
	{
		ID: string(WeaponSteelSword), Name: "Steel Sword", Kind: ItemKindWeapon,
		EquipRequirements: []SkillRequirement{{Skill: SkillAttack, Level: 10}},
		Weapon:            swordWeapon(WeaponDefinition{ID: WeaponSteelSword, AccuracyBonus: 12, MeleeStrength: steelSwordMeleeStrength, DamageVariance: swordDamageVariance, AttackInterval: swordAttackIntervalPulses, Range: swordRange, ApproachRange: swordApproachRange}),
	},
	{
		ID: string(WeaponMithrilSword), Name: "Mithril Sword", Kind: ItemKindWeapon,
		EquipRequirements: []SkillRequirement{{Skill: SkillAttack, Level: 17}},
		Weapon:            swordWeapon(WeaponDefinition{ID: WeaponMithrilSword, AccuracyBonus: 18, MeleeStrength: 18, DamageVariance: swordDamageVariance, AttackInterval: swordAttackIntervalPulses, Range: swordRange, ApproachRange: swordApproachRange}),
	},
	{
		ID: string(WeaponRuneSword), Name: "Rune Sword", Kind: ItemKindWeapon,
		EquipRequirements: []SkillRequirement{{Skill: SkillAttack, Level: 20}},
		Weapon:            swordWeapon(WeaponDefinition{ID: WeaponRuneSword, AccuracyBonus: 40, MeleeStrength: runeSwordMeleeStrength, DamageVariance: swordDamageVariance, AttackInterval: swordAttackIntervalPulses, Range: swordRange, ApproachRange: swordApproachRange}),
	},
	{
		ID: string(WeaponDagger), Name: "Dagger", Kind: ItemKindWeapon,
		Weapon: swordWeapon(WeaponDefinition{ID: WeaponDagger, AccuracyBonus: 14, MeleeStrength: daggerMeleeStrength, DamageVariance: daggerDamageVariance, AttackInterval: daggerAttackIntervalPulses, Range: daggerRange, ApproachRange: daggerApproachRange}),
	},
	{
		ID: string(WeaponBow), Name: "Shortbow", Kind: ItemKindWeapon,
		Weapon: weaponWithStyles(WeaponDefinition{ID: WeaponBow, TwoHanded: true, AccuracyBonus: 12, RangedStrength: bowRangedStrength, DamageVariance: bowDamageVariance, AttackInterval: bowAttackIntervalPulses, Range: bowRange, ApproachRange: bowApproachRange, Ammunition: ItemCost{ItemID: ItemIronArrow, Quantity: 1}, AmmunitionFamily: AmmunitionFamilyArrow}, CombatStyleAccurate,
			weaponStyle(CombatStyleRapid, DamageTypeRanged), weaponStyle(CombatStyleAccurate, DamageTypeRanged), weaponStyle(CombatStyleLongRange, DamageTypeRanged)),
	},
	{
		ID: string(WeaponOakShortbow), Name: "Oak Shortbow", Kind: ItemKindWeapon,
		EquipRequirements: []SkillRequirement{{Skill: SkillRange, Level: 5}},
		Weapon: weaponWithStyles(WeaponDefinition{ID: WeaponOakShortbow, TwoHanded: true, AccuracyBonus: 14, RangedStrength: 14, DamageVariance: bowDamageVariance, AttackInterval: bowAttackIntervalPulses, Range: bowRange, ApproachRange: bowApproachRange, Ammunition: ItemCost{ItemID: ItemIronArrow, Quantity: 1}, AmmunitionFamily: AmmunitionFamilyArrow}, CombatStyleAccurate,
			weaponStyle(CombatStyleRapid, DamageTypeRanged), weaponStyle(CombatStyleAccurate, DamageTypeRanged), weaponStyle(CombatStyleLongRange, DamageTypeRanged)),
	},
	{
		ID: string(WeaponWillowShortbow), Name: "Willow Shortbow", Kind: ItemKindWeapon,
		EquipRequirements: []SkillRequirement{{Skill: SkillRange, Level: 10}},
		Weapon: weaponWithStyles(WeaponDefinition{ID: WeaponWillowShortbow, TwoHanded: true, AccuracyBonus: 16, RangedStrength: 16, DamageVariance: bowDamageVariance, AttackInterval: bowAttackIntervalPulses, Range: bowRange, ApproachRange: bowApproachRange, Ammunition: ItemCost{ItemID: ItemIronArrow, Quantity: 1}, AmmunitionFamily: AmmunitionFamilyArrow}, CombatStyleAccurate,
			weaponStyle(CombatStyleRapid, DamageTypeRanged), weaponStyle(CombatStyleAccurate, DamageTypeRanged), weaponStyle(CombatStyleLongRange, DamageTypeRanged)),
	},
	{
		ID: string(WeaponMapleShortbow), Name: "Maple Shortbow", Kind: ItemKindWeapon,
		EquipRequirements: []SkillRequirement{{Skill: SkillRange, Level: 20}},
		Weapon: weaponWithStyles(WeaponDefinition{ID: WeaponMapleShortbow, TwoHanded: true, AccuracyBonus: 19, RangedStrength: 19, DamageVariance: bowDamageVariance, AttackInterval: bowAttackIntervalPulses, Range: bowRange, ApproachRange: bowApproachRange, Ammunition: ItemCost{ItemID: ItemIronArrow, Quantity: 1}, AmmunitionFamily: AmmunitionFamilyArrow}, CombatStyleAccurate,
			weaponStyle(CombatStyleRapid, DamageTypeRanged), weaponStyle(CombatStyleAccurate, DamageTypeRanged), weaponStyle(CombatStyleLongRange, DamageTypeRanged)),
	},
	{
		ID: string(WeaponYewShortbow), Name: "Yew Shortbow", Kind: ItemKindWeapon,
		EquipRequirements: []SkillRequirement{{Skill: SkillRange, Level: 30}},
		Weapon: weaponWithStyles(WeaponDefinition{ID: WeaponYewShortbow, TwoHanded: true, AccuracyBonus: 23, RangedStrength: 23, DamageVariance: bowDamageVariance, AttackInterval: bowAttackIntervalPulses, Range: bowRange, ApproachRange: bowApproachRange, Ammunition: ItemCost{ItemID: ItemIronArrow, Quantity: 1}, AmmunitionFamily: AmmunitionFamilyArrow}, CombatStyleAccurate,
			weaponStyle(CombatStyleRapid, DamageTypeRanged), weaponStyle(CombatStyleAccurate, DamageTypeRanged), weaponStyle(CombatStyleLongRange, DamageTypeRanged)),
	},
	{
		ID: string(WeaponMace), Name: "Mace", Kind: ItemKindWeapon,
		EquipRequirements: []SkillRequirement{{Skill: SkillAttack, Level: 5}},
		Weapon:            weaponWithStyles(WeaponDefinition{ID: WeaponMace, AccuracyBonus: 8, MeleeStrength: maceMeleeStrength, DamageVariance: maceDamageVariance, AttackInterval: maceAttackIntervalPulses, Range: maceRange, ApproachRange: maceApproachRange}, CombatStyleCrush, weaponStyle(CombatStyleCrush, DamageTypeCrush), weaponStyle(CombatStyleHeavyCrush, DamageTypeCrush), weaponStyle(CombatStyleDefensive, DamageTypeCrush)),
	},
	{
		ID: string(WeaponBattleAxe), Name: "Battle Axe", Kind: ItemKindWeapon,
		EquipRequirements: []SkillRequirement{{Skill: SkillAttack, Level: 10}},
		Weapon:            weaponWithStyles(WeaponDefinition{ID: WeaponBattleAxe, AccuracyBonus: 6, MeleeStrength: battleAxeMeleeStrength, DamageVariance: battleAxeDamageVariance, AttackInterval: battleAxeAttackIntervalPulses, Range: battleAxeRange, ApproachRange: battleAxeApproachRange}, CombatStyleSlash, weaponStyle(CombatStyleSlash, DamageTypeSlash), weaponStyle(CombatStyleHeavyCrush, DamageTypeCrush), weaponStyle(CombatStyleDefensive, DamageTypeSlash)),
	},
	{
		ID: string(WeaponStaff), Name: "Staff", Kind: ItemKindWeapon,
		Weapon: weaponWithStyles(WeaponDefinition{ID: WeaponStaff, AccuracyBonus: 11, MagicStrength: staffMagicStrength, DamageVariance: staffDamageVariance, AttackInterval: staffAttackIntervalPulses, Range: staffRange, ApproachRange: staffApproachRange}, CombatStyleAccurate,
			weaponStyle(CombatStyleRapid, DamageTypeMagic), weaponStyle(CombatStyleAccurate, DamageTypeMagic), weaponStyle(CombatStyleLongRange, DamageTypeMagic)),
	},
	{
		ID: string(WeaponSpear), Name: "Spear", Kind: ItemKindWeapon,
		Weapon: weaponWithStyles(WeaponDefinition{ID: WeaponSpear, AccuracyBonus: 14, MeleeStrength: spearMeleeStrength, DamageVariance: 0.20, AttackInterval: spearAttackIntervalPulses, Range: polearmRange, ApproachRange: polearmApproachRange}, CombatStyleStab,
			weaponStyle(CombatStyleStab, DamageTypeStab), weaponStyle(CombatStyleSlash, DamageTypeSlash), weaponStyle(CombatStyleDefensive, DamageTypeStab)),
	},
	{
		ID: string(WeaponHalberd), Name: "Halberd", Kind: ItemKindWeapon,
		Weapon: weaponWithStyles(WeaponDefinition{ID: WeaponHalberd, TwoHanded: true, AccuracyBonus: 10, MeleeStrength: halberdMeleeStrength, DamageVariance: 0.25, AttackInterval: halberdAttackIntervalPulses, Range: polearmRange, ApproachRange: polearmApproachRange}, CombatStyleSlash,
			weaponStyle(CombatStyleSlash, DamageTypeSlash), weaponStyle(CombatStyleStab, DamageTypeStab), weaponStyle(CombatStyleDefensive, DamageTypeSlash)),
	},
	{
		ID: string(WeaponCrossbow), Name: "One-handed Crossbow", Kind: ItemKindWeapon,
		Weapon: weaponWithStyles(WeaponDefinition{ID: WeaponCrossbow, AccuracyBonus: 18, RangedStrength: crossbowRangedStrength, DamageVariance: 0.20, AttackInterval: crossbowAttackIntervalPulses, Range: crossbowRange, ApproachRange: crossbowApproachRange, Ammunition: ItemCost{ItemID: ItemIronBolt, Quantity: 1}, AmmunitionFamily: AmmunitionFamilyBolt}, CombatStyleAccurate,
			weaponStyle(CombatStyleRapid, DamageTypeRanged), weaponStyle(CombatStyleAccurate, DamageTypeRanged), weaponStyle(CombatStyleLongRange, DamageTypeRanged)),
	},
	{
		ID: ItemTrainingHelm, Name: "Training Helm", Kind: ItemKindArmour,
		Armour: ArmourDefinition{Slot: ArmourSlotHead, HealthBonus: 2, Defence: DefenceBonuses{Stab: 2, Slash: 3, Crush: 2, Ranged: 1}},
	},
	{
		ID: ItemBronzeHelm, Name: "Bronze Helm", Kind: ItemKindArmour,
		Armour: ArmourDefinition{Slot: ArmourSlotHead, HealthBonus: 2, Defence: DefenceBonuses{Stab: 2, Slash: 2, Crush: 3, Ranged: 1, Magic: -1}},
	},
	{
		ID: ItemBronzePlate, Name: "Bronze Platebody", Kind: ItemKindArmour,
		Armour: ArmourDefinition{Slot: ArmourSlotBody, HealthBonus: 6, Defence: DefenceBonuses{Stab: 5, Slash: 6, Crush: 4, Ranged: 3, Magic: -2}},
	},
	{ID: ItemBronzeShield, Name: "Bronze Shield", Kind: ItemKindArmour, Armour: ArmourDefinition{Slot: ArmourSlotLeftHand, HealthBonus: 3, Defence: DefenceBonuses{Stab: 4, Slash: 5, Crush: 6, Ranged: 4, Magic: -2}}},
	{ID: ItemBronzeLegs, Name: "Bronze Legs", Kind: ItemKindArmour, Armour: ArmourDefinition{Slot: ArmourSlotLegs, HealthBonus: 4, Defence: DefenceBonuses{Stab: 4, Slash: 5, Crush: 3, Ranged: 2, Magic: -1}}},
	{ID: ItemBronzeGloves, Name: "Bronze Gloves", Kind: ItemKindArmour, Armour: ArmourDefinition{Slot: ArmourSlotHands, HealthBonus: 1, Defence: DefenceBonuses{Stab: 1, Slash: 2, Crush: 2, Ranged: 1}}},
	{ID: ItemBronzeBoots, Name: "Bronze Boots", Kind: ItemKindArmour, Armour: ArmourDefinition{Slot: ArmourSlotFeet, HealthBonus: 1, Defence: DefenceBonuses{Stab: 2, Slash: 2, Crush: 1, Ranged: 1}}},
	{
		ID: ItemIronHelm, Name: "Iron Helm", Kind: ItemKindArmour,
		EquipRequirements: []SkillRequirement{{Skill: SkillDefence, Level: 5}},
		Armour:            ArmourDefinition{Slot: ArmourSlotHead, HealthBonus: 4, Defence: DefenceBonuses{Stab: 4, Slash: 5, Crush: 4, Ranged: 2, Magic: -1}},
	},
	{
		ID: ItemIronPlate, Name: "Iron Platebody", Kind: ItemKindArmour,
		EquipRequirements: []SkillRequirement{{Skill: SkillDefence, Level: 5}},
		Armour:            ArmourDefinition{Slot: ArmourSlotBody, HealthBonus: 10, Defence: DefenceBonuses{Stab: 10, Slash: 12, Crush: 9, Ranged: 7, Magic: -3}},
	},
	{ID: ItemIronShield, Name: "Iron Shield", Kind: ItemKindArmour, EquipRequirements: []SkillRequirement{{Skill: SkillDefence, Level: 5}}, Armour: ArmourDefinition{Slot: ArmourSlotLeftHand, HealthBonus: 5, Defence: DefenceBonuses{Stab: 8, Slash: 9, Crush: 11, Ranged: 7, Magic: -3}}},
	{ID: ItemIronLegs, Name: "Iron Legs", Kind: ItemKindArmour, EquipRequirements: []SkillRequirement{{Skill: SkillDefence, Level: 5}}, Armour: ArmourDefinition{Slot: ArmourSlotLegs, HealthBonus: 7, Defence: DefenceBonuses{Stab: 7, Slash: 8, Crush: 6, Ranged: 5, Magic: -2}}},
	{ID: ItemIronGloves, Name: "Iron Gloves", Kind: ItemKindArmour, EquipRequirements: []SkillRequirement{{Skill: SkillDefence, Level: 5}}, Armour: ArmourDefinition{Slot: ArmourSlotHands, HealthBonus: 2, Defence: DefenceBonuses{Stab: 3, Slash: 3, Crush: 4, Ranged: 2, Magic: -1}}},
	{ID: ItemIronBoots, Name: "Iron Boots", Kind: ItemKindArmour, EquipRequirements: []SkillRequirement{{Skill: SkillDefence, Level: 5}}, Armour: ArmourDefinition{Slot: ArmourSlotFeet, HealthBonus: 2, Defence: DefenceBonuses{Stab: 3, Slash: 4, Crush: 3, Ranged: 2, Magic: -1}}},
	{
		ID: ItemMithrilHelm, Name: "Mithril Helm", Kind: ItemKindArmour,
		EquipRequirements: []SkillRequirement{{Skill: SkillDefence, Level: 17}},
		Armour:            ArmourDefinition{Slot: ArmourSlotHead, HealthBonus: 6, Defence: DefenceBonuses{Stab: 7, Slash: 8, Crush: 7, Ranged: 4, Magic: -2}},
	},
	{
		ID: ItemMithrilPlate, Name: "Mithril Platebody", Kind: ItemKindArmour,
		EquipRequirements: []SkillRequirement{{Skill: SkillDefence, Level: 17}},
		Armour:            ArmourDefinition{Slot: ArmourSlotBody, HealthBonus: 15, Defence: DefenceBonuses{Stab: 16, Slash: 18, Crush: 14, Ranged: 11, Magic: -5}},
	},
	{ID: ItemMithrilShield, Name: "Mithril Shield", Kind: ItemKindArmour, EquipRequirements: []SkillRequirement{{Skill: SkillDefence, Level: 17}}, Armour: ArmourDefinition{Slot: ArmourSlotLeftHand, HealthBonus: 8, Defence: DefenceBonuses{Stab: 13, Slash: 15, Crush: 17, Ranged: 11, Magic: -5}}},
	{ID: ItemMithrilLegs, Name: "Mithril Legs", Kind: ItemKindArmour, EquipRequirements: []SkillRequirement{{Skill: SkillDefence, Level: 17}}, Armour: ArmourDefinition{Slot: ArmourSlotLegs, HealthBonus: 10, Defence: DefenceBonuses{Stab: 11, Slash: 13, Crush: 10, Ranged: 8, Magic: -3}}},
	{ID: ItemMithrilGloves, Name: "Mithril Gloves", Kind: ItemKindArmour, EquipRequirements: []SkillRequirement{{Skill: SkillDefence, Level: 17}}, Armour: ArmourDefinition{Slot: ArmourSlotHands, HealthBonus: 3, Defence: DefenceBonuses{Stab: 5, Slash: 6, Crush: 6, Ranged: 4, Magic: -1}}},
	{ID: ItemMithrilBoots, Name: "Mithril Boots", Kind: ItemKindArmour, EquipRequirements: []SkillRequirement{{Skill: SkillDefence, Level: 17}}, Armour: ArmourDefinition{Slot: ArmourSlotFeet, HealthBonus: 3, Defence: DefenceBonuses{Stab: 5, Slash: 6, Crush: 5, Ranged: 4, Magic: -1}}},
	{
		ID: ItemTrainingPlate, Name: "Training Plate", Kind: ItemKindArmour,
		Armour: ArmourDefinition{Slot: ArmourSlotBody, HealthBonus: 5, Defence: DefenceBonuses{Stab: 8, Slash: 10, Crush: 7, Ranged: 6, Magic: -2}},
	},
	{
		ID: ItemSwiftHelm, Name: "Swift Helm", Kind: ItemKindArmour,
		EquipRequirements: []SkillRequirement{{Skill: SkillDefence, Level: 5}},
		Armour:            ArmourDefinition{Slot: ArmourSlotHead, Set: armourSetSwift, EvasionBonus: 1, HealthBonus: 1, Defence: DefenceBonuses{Stab: 1, Slash: 1, Crush: 1, Ranged: 2}},
	},
	{
		ID: ItemSwiftBody, Name: "Swift Body", Kind: ItemKindArmour,
		EquipRequirements: []SkillRequirement{{Skill: SkillDefence, Level: 5}},
		Armour:            ArmourDefinition{Slot: ArmourSlotBody, Set: armourSetSwift, EvasionBonus: 3, HealthBonus: 3, Defence: DefenceBonuses{Stab: 4, Slash: 4, Crush: 3, Ranged: 5, Magic: -1}},
	},
	{
		ID: ItemSwiftBoots, Name: "Swift Boots", Kind: ItemKindArmour,
		EquipRequirements: []SkillRequirement{{Skill: SkillDefence, Level: 5}},
		Armour:            ArmourDefinition{Slot: ArmourSlotFeet, Set: armourSetSwift, EvasionBonus: 1, HealthBonus: 1, Defence: DefenceBonuses{Stab: 1, Slash: 1, Crush: 1, Ranged: 2}},
	},
	{ID: ItemNoviceRobeHood, Name: "Novice Robe Hood", Kind: ItemKindArmour, Armour: ArmourDefinition{Slot: ArmourSlotHead, Defence: DefenceBonuses{Magic: 3}, Offence: OffenceBonuses{MagicAccuracy: 2, MagicStrength: 1}}},
	{ID: ItemNoviceRobeTop, Name: "Novice Robe Top", Kind: ItemKindArmour, Armour: ArmourDefinition{Slot: ArmourSlotBody, HealthBonus: 2, Defence: DefenceBonuses{Magic: 5}, Offence: OffenceBonuses{MagicAccuracy: 3, MagicStrength: 2}}},
	{ID: ItemNoviceRobeGloves, Name: "Novice Robe Gloves", Kind: ItemKindArmour, Armour: ArmourDefinition{Slot: ArmourSlotHands, Defence: DefenceBonuses{Magic: 1}, Offence: OffenceBonuses{MagicAccuracy: 1}}},
	{ID: ItemNoviceRobeBottoms, Name: "Novice Robe Bottoms", Kind: ItemKindArmour, Armour: ArmourDefinition{Slot: ArmourSlotLegs, HealthBonus: 1, Defence: DefenceBonuses{Magic: 3}, Offence: OffenceBonuses{MagicAccuracy: 2, MagicStrength: 1}}},
	{ID: ItemNoviceRobeBoots, Name: "Novice Robe Boots", Kind: ItemKindArmour, Armour: ArmourDefinition{Slot: ArmourSlotFeet, Defence: DefenceBonuses{Magic: 1}, Offence: OffenceBonuses{MagicAccuracy: 1, MagicStrength: 1}}},
	{ID: ItemNoviceRangerHood, Name: "Novice Ranger Hood", Kind: ItemKindArmour, Armour: ArmourDefinition{Slot: ArmourSlotHead, EvasionBonus: 1, Defence: DefenceBonuses{Stab: 1, Ranged: 2, Magic: 1}, Offence: OffenceBonuses{RangedAccuracy: 2, RangedStrength: 1}}},
	{ID: ItemNoviceRangerBody, Name: "Novice Ranger Body", Kind: ItemKindArmour, Armour: ArmourDefinition{Slot: ArmourSlotBody, EvasionBonus: 2, HealthBonus: 2, Defence: DefenceBonuses{Stab: 2, Slash: 1, Ranged: 4, Magic: 2}, Offence: OffenceBonuses{RangedAccuracy: 3, RangedStrength: 2}}},
	{ID: ItemNoviceRangerGloves, Name: "Novice Ranger Gloves", Kind: ItemKindArmour, Armour: ArmourDefinition{Slot: ArmourSlotHands, Defence: DefenceBonuses{Ranged: 1}, Offence: OffenceBonuses{RangedAccuracy: 1}}},
	{ID: ItemNoviceRangerLegs, Name: "Novice Ranger Legs", Kind: ItemKindArmour, Armour: ArmourDefinition{Slot: ArmourSlotLegs, EvasionBonus: 1, HealthBonus: 1, Defence: DefenceBonuses{Stab: 1, Ranged: 3, Magic: 1}, Offence: OffenceBonuses{RangedAccuracy: 2, RangedStrength: 1}}},
	{ID: ItemNoviceRangerBoots, Name: "Novice Ranger Boots", Kind: ItemKindArmour, Armour: ArmourDefinition{Slot: ArmourSlotFeet, EvasionBonus: 1, Defence: DefenceBonuses{Ranged: 1}, Offence: OffenceBonuses{RangedAccuracy: 1, RangedStrength: 1}}},
	{ID: ItemVanguardCape, Name: "Vanguard Cape", Kind: ItemKindArmour, Armour: ArmourDefinition{Slot: ArmourSlotCape, Defence: DefenceBonuses{Stab: 1, Slash: 1, Crush: 1}, Offence: OffenceBonuses{MeleeAccuracy: 3, MeleeStrength: 2}}},
	{ID: ItemWayfarerCape, Name: "Wayfarer Cape", Kind: ItemKindArmour, Armour: ArmourDefinition{Slot: ArmourSlotCape, EvasionBonus: 1, Defence: DefenceBonuses{Ranged: 1}, Offence: OffenceBonuses{RangedAccuracy: 3, RangedStrength: 2}}},
	{ID: ItemArcanistCape, Name: "Arcanist Cape", Kind: ItemKindArmour, Armour: ArmourDefinition{Slot: ArmourSlotCape, Defence: DefenceBonuses{Magic: 1}, Offence: OffenceBonuses{MagicAccuracy: 3, MagicStrength: 2}}},
	{
		ID: ItemAmuletFerocity, Name: "Ruby Amulet", Kind: ItemKindArmour,
		EquipRequirements: []SkillRequirement{{Skill: SkillAttack, Level: 5}},
		Armour:            ArmourDefinition{Slot: ArmourSlotNeck, Offence: OffenceBonuses{MeleeAccuracy: 5, MeleeStrength: 5}},
	},
	{
		ID: ItemAmuletHawk, Name: "Emerald Amulet", Kind: ItemKindArmour,
		EquipRequirements: []SkillRequirement{{Skill: SkillRange, Level: 5}},
		Armour:            ArmourDefinition{Slot: ArmourSlotNeck, Offence: OffenceBonuses{RangedAccuracy: 5, RangedStrength: 5}},
	},
	{
		ID: ItemAmuletEmbers, Name: "Sapphire Amulet", Kind: ItemKindArmour,
		EquipRequirements: []SkillRequirement{{Skill: SkillMagic, Level: 5}},
		Armour:            ArmourDefinition{Slot: ArmourSlotNeck, Offence: OffenceBonuses{MagicAccuracy: 5, MagicStrength: 5}},
	},
	{
		ID: ItemAmuletFury, Name: "Diamond Amulet", Kind: ItemKindArmour,
		EquipRequirements: []SkillRequirement{{Skill: SkillDefence, Level: 5}},
		Armour: ArmourDefinition{Slot: ArmourSlotNeck, Offence: OffenceBonuses{
			MeleeAccuracy: 2, RangedAccuracy: 2, MagicAccuracy: 2,
			MeleeStrength: 2, RangedStrength: 2, MagicStrength: 2,
		}},
	},
	{ID: ItemRubyRing, Name: "Ruby Ring", Kind: ItemKindArmour, Armour: ArmourDefinition{Slot: ArmourSlotRing, Offence: OffenceBonuses{MeleeStrength: 2}}},
	{ID: ItemEmeraldRing, Name: "Emerald Ring", Kind: ItemKindArmour, Armour: ArmourDefinition{Slot: ArmourSlotRing, Offence: OffenceBonuses{RangedStrength: 2}}},
	{ID: ItemSapphireRing, Name: "Sapphire Ring", Kind: ItemKindArmour, Armour: ArmourDefinition{Slot: ArmourSlotRing, Offence: OffenceBonuses{MagicStrength: 2}}},
	{ID: ItemDiamondRing, Name: "Diamond Ring", Kind: ItemKindArmour, Armour: ArmourDefinition{Slot: ArmourSlotRing, Offence: OffenceBonuses{
		MeleeAccuracy: 1, RangedAccuracy: 1, MagicAccuracy: 1,
		MeleeStrength: 1, RangedStrength: 1, MagicStrength: 1,
	}}},
	{ID: ItemTyrantsRing, Name: "Tyrant's Ring", Kind: ItemKindArmour, Armour: ArmourDefinition{
		Slot: ArmourSlotRing, Offence: OffenceBonuses{MeleeStrength: 100},
	}},
	{ID: ItemRangedTyrantsRing, Name: "Ranged Tyrant's Ring", Kind: ItemKindArmour, Armour: ArmourDefinition{
		Slot: ArmourSlotRing, Offence: OffenceBonuses{RangedStrength: 100},
	}},
	{ID: ItemMageTyrantsRing, Name: "Mage Tyrant's Ring", Kind: ItemKindArmour, Armour: ArmourDefinition{
		Slot: ArmourSlotRing, Offence: OffenceBonuses{MagicStrength: 100},
	}},
	{
		ID: ItemStandardSpellbook, Name: "Standard Spellbook", Kind: ItemKindSpellbook,
		Spellbook: SpellbookStandard,
	},
	{
		ID: ItemSkinBible, Name: "Skin Bible", Kind: ItemKindSpellbook,
		Spellbook: SpellbookDark,
	},
	{
		ID: ItemHolyBook, Name: "Holy Book", Kind: ItemKindSpellbook,
		Spellbook: SpellbookHoly,
	},
	{
		ID: ItemClericsTome, Name: "Cleric's Tome", Kind: ItemKindSpellbook,
		Spellbook: SpellbookCleric,
	},
	{ID: ItemAirRune, Name: "Air Rune", Kind: ItemKindResource},
	{ID: ItemWaterRune, Name: "Water Rune", Kind: ItemKindResource},
	{ID: ItemEarthRune, Name: "Earth Rune", Kind: ItemKindResource},
	{ID: ItemFireRune, Name: "Fire Rune", Kind: ItemKindResource},
	{ID: ItemHolyRune, Name: "Holy Rune", Kind: ItemKindResource},
	{ID: ItemDarkRune, Name: "Dark Rune", Kind: ItemKindResource},
	{ID: ItemFocusRune, Name: "Focus Rune", Kind: ItemKindResource},
	{ID: ItemBlankRune, Name: "Blank Rune", Kind: ItemKindResource},
	{ID: ItemLeather, Name: "Leather", Kind: ItemKindResource, StackLimit: 25},
	{ID: ItemGoldOre, Name: "Gold Ore", Kind: ItemKindResource},
	{ID: ItemSilverOre, Name: "Silver Ore", Kind: ItemKindResource},
	{ID: ItemClay, Name: "Clay", Kind: ItemKindResource},
	{ID: ItemAdamantiteOre, Name: "Adamantite Ore", Kind: ItemKindResource},
	{ID: ItemUncutSapphire, Name: "Uncut Sapphire", Kind: ItemKindResource, StackLimit: 10},
	{ID: ItemSapphire, Name: "Sapphire", Kind: ItemKindResource, StackLimit: 10},
	{ID: ItemUncutEmerald, Name: "Uncut Emerald", Kind: ItemKindResource, StackLimit: 10},
	{ID: ItemEmerald, Name: "Emerald", Kind: ItemKindResource, StackLimit: 10},
	{ID: ItemUncutRuby, Name: "Uncut Ruby", Kind: ItemKindResource, StackLimit: 10},
	{ID: ItemRuby, Name: "Ruby", Kind: ItemKindResource, StackLimit: 10},
	{ID: ItemUncutDiamond, Name: "Uncut Diamond", Kind: ItemKindResource, StackLimit: 10},
	{ID: ItemDiamond, Name: "Diamond", Kind: ItemKindResource, StackLimit: 10},
	{ID: ItemBronzeArrow, Name: "Bronze Arrow", Kind: ItemKindResource, Ammunition: true, AmmunitionFamily: AmmunitionFamilyArrow, AmmunitionStrength: -1},
	{ID: ItemIronArrow, Name: "Iron Arrow", Kind: ItemKindResource, Ammunition: true, AmmunitionFamily: AmmunitionFamilyArrow},
	{ID: ItemSteelArrow, Name: "Steel Arrow", Kind: ItemKindResource, Ammunition: true, AmmunitionFamily: AmmunitionFamilyArrow, AmmunitionStrength: 1},
	{ID: ItemMithrilArrow, Name: "Mithril Arrow", Kind: ItemKindResource, Ammunition: true, AmmunitionFamily: AmmunitionFamilyArrow, AmmunitionStrength: 2},
	{ID: ItemAdamantiteArrow, Name: "Adamantite Arrow", Kind: ItemKindResource, Ammunition: true, AmmunitionFamily: AmmunitionFamilyArrow, AmmunitionStrength: 3},
	{ID: ItemBronzeBolt, Name: "Bronze Bolt", Kind: ItemKindResource, Ammunition: true, AmmunitionFamily: AmmunitionFamilyBolt, AmmunitionStrength: -1},
	{ID: ItemIronBolt, Name: "Iron Bolt", Kind: ItemKindResource, Ammunition: true, AmmunitionFamily: AmmunitionFamilyBolt},
	{ID: ItemSteelBolt, Name: "Steel Bolt", Kind: ItemKindResource, Ammunition: true, AmmunitionFamily: AmmunitionFamilyBolt, AmmunitionStrength: 1},
	{ID: ItemMithrilBolt, Name: "Mithril Bolt", Kind: ItemKindResource, Ammunition: true, AmmunitionFamily: AmmunitionFamilyBolt, AmmunitionStrength: 2},
	{ID: ItemAdamantiteBolt, Name: "Adamantite Bolt", Kind: ItemKindResource, Ammunition: true, AmmunitionFamily: AmmunitionFamilyBolt, AmmunitionStrength: 3},
	{ID: ItemBones, Name: "Bones", Kind: ItemKindResource, StackLimit: 25},
	{ID: ItemCoins, Name: "Coins", Kind: ItemKindResource},
	{ID: ItemBronzePickaxe, Name: "Bronze Pickaxe", Kind: ItemKindResource},
	{ID: ItemBronzeAxe, Name: "Bronze Axe", Kind: ItemKindResource},
	{ID: ItemKnife, Name: "Knife", Kind: ItemKindResource, StackLimit: 1},
	{ID: ItemChisel, Name: "Chisel", Kind: ItemKindResource, StackLimit: 1},
	{ID: ItemSewingNeedle, Name: "Sewing Needle", Kind: ItemKindResource, StackLimit: 1},
	{ID: ItemThread, Name: "Thread", Kind: ItemKindResource, StackLimit: 100},
	{ID: ItemTinderbox, Name: "Tinderbox", Kind: ItemKindResource, StackLimit: 1},
	{ID: ItemHammer, Name: "Hammer", Kind: ItemKindResource, StackLimit: 1},
	{ID: ItemRingMould, Name: "Ring Mould", Kind: ItemKindResource, StackLimit: 1},
	{ID: ItemAmuletMould, Name: "Amulet Mould", Kind: ItemKindResource, StackLimit: 1},
	{ID: ItemFishingRod, Name: "Fishing Rod", Kind: ItemKindResource, StackLimit: 1},
	{ID: ItemLobsterCage, Name: "Lobster Cage", Kind: ItemKindResource, StackLimit: 1},
	{ID: ItemVialOfWater, Name: "Vial of Water", Kind: ItemKindResource, StackLimit: 25},
	{ID: ItemSunleaf, Name: "Sunleaf", Kind: ItemKindResource, StackLimit: 10},
	{ID: ItemMoonleaf, Name: "Moonleaf", Kind: ItemKindResource, StackLimit: 10},
	{ID: ItemSacredDust, Name: "Sacred Dust", Kind: ItemKindResource, StackLimit: 25},
	{ID: ItemFishOil, Name: "Fish Oil", Kind: ItemKindResource, StackLimit: 25},
	{ID: ItemArrowShafts, Name: "Arrow Shafts", Kind: ItemKindResource, StackLimit: MaximumItemStackQuantity},
	{ID: ItemFeathers, Name: "Feathers", Kind: ItemKindResource, StackLimit: 200},
	{ID: ItemBronzeArrowTips, Name: "Bronze Arrow Tips", Kind: ItemKindResource, StackLimit: 200},
	{ID: ItemIronArrowTips, Name: "Iron Arrow Tips", Kind: ItemKindResource, StackLimit: 200},
	{ID: ItemSteelArrowTips, Name: "Steel Arrow Tips", Kind: ItemKindResource, StackLimit: 200},
	{ID: ItemMithrilArrowTips, Name: "Mithril Arrow Tips", Kind: ItemKindResource, StackLimit: 200},
	{ID: ItemAdamantiteArrowTips, Name: "Adamantite Arrow Tips", Kind: ItemKindResource, StackLimit: 200},
	{ID: ItemBowstring, Name: "Bowstring", Kind: ItemKindResource, StackLimit: 50},
	{ID: ItemUnstrungShortbow, Name: "Unstrung Shortbow", Kind: ItemKindResource},
	{ID: ItemUnstrungOakShortbow, Name: "Unstrung Oak Shortbow", Kind: ItemKindResource},
	{ID: ItemUnstrungWillowShortbow, Name: "Unstrung Willow Shortbow", Kind: ItemKindResource},
	{ID: ItemUnstrungMapleShortbow, Name: "Unstrung Maple Shortbow", Kind: ItemKindResource},
	{ID: ItemUnstrungYewShortbow, Name: "Unstrung Yew Shortbow", Kind: ItemKindResource},
	{
		ID: ItemPrayerPotion, Name: "Prayer Potion", Kind: ItemKindConsumable, StackLimit: 5,
		Consumable: ConsumableDefinition{PrayerRestoration: 25, CooldownPulses: 2},
	},
	{
		ID: ItemMinorHealingPotion, Name: "Minor Healing Potion", Kind: ItemKindConsumable, StackLimit: 5,
		Consumable: ConsumableDefinition{Healing: 15, CooldownPulses: 2},
	},
	{
		ID: ItemMurkyBrew, Name: "Murky Brew", Kind: ItemKindConsumable, StackLimit: 5,
		Consumable: ConsumableDefinition{Healing: 5, CooldownPulses: 2},
	},
	{
		ID: ItemMeleePotion, Name: "Melee Potion", Kind: ItemKindConsumable, StackLimit: 5,
		Consumable: ConsumableDefinition{CooldownPulses: 2, BoostDuration: combatPotionDurationPulses, SkillBoosts: []SkillBoostDefinition{
			{Skill: SkillAttack, Levels: combatPotionBoostLevels},
			{Skill: SkillStrength, Levels: combatPotionBoostLevels},
			{Skill: SkillDefence, Levels: combatPotionBoostLevels},
		}},
	},
	{
		ID: ItemRangingPotion, Name: "Ranging Potion", Kind: ItemKindConsumable, StackLimit: 5,
		Consumable: ConsumableDefinition{CooldownPulses: 2, BoostDuration: combatPotionDurationPulses, SkillBoosts: []SkillBoostDefinition{
			{Skill: SkillRange, Levels: combatPotionBoostLevels},
		}},
	},
	{
		ID: ItemMagicPotion, Name: "Magic Potion", Kind: ItemKindConsumable, StackLimit: 5,
		Consumable: ConsumableDefinition{CooldownPulses: 2, BoostDuration: combatPotionDurationPulses, SkillBoosts: []SkillBoostDefinition{
			{Skill: SkillMagic, Levels: combatPotionBoostLevels},
		}},
	},
	{ID: ItemNormalLogs, Name: "Normal Logs", Kind: ItemKindResource, StackLimit: 10},
	{ID: ItemOakLogs, Name: "Oak Logs", Kind: ItemKindResource, StackLimit: 10},
	{ID: ItemWillowLogs, Name: "Willow Logs", Kind: ItemKindResource, StackLimit: 10},
	{ID: ItemMapleLogs, Name: "Maple Logs", Kind: ItemKindResource, StackLimit: 10},
	{ID: ItemYewLogs, Name: "Yew Logs", Kind: ItemKindResource, StackLimit: 10},
	{ID: ItemCopperOre, Name: "Copper Ore", Kind: ItemKindResource},
	{ID: ItemTinOre, Name: "Tin Ore", Kind: ItemKindResource},
	{ID: ItemIronOre, Name: "Iron Ore", Kind: ItemKindResource},
	{ID: ItemCoal, Name: "Coal", Kind: ItemKindResource},
	{ID: ItemMithrilOre, Name: "Mithril Ore", Kind: ItemKindResource},
	{ID: ItemBronzeBar, Name: "Bronze Bar", Kind: ItemKindResource},
	{ID: ItemIronBar, Name: "Iron Bar", Kind: ItemKindResource},
	{ID: ItemMithrilBar, Name: "Mithril Bar", Kind: ItemKindResource},
	{ID: ItemGoldBar, Name: "Gold Bar", Kind: ItemKindResource},
	{ID: ItemSilverBar, Name: "Silver Bar", Kind: ItemKindResource},
	{ID: ItemSteelBar, Name: "Steel Bar", Kind: ItemKindResource},
	{ID: ItemAdamantiteBar, Name: "Adamantite Bar", Kind: ItemKindResource},
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
})

type weaponStyleDefinition struct {
	ID         CombatStyleID
	DamageType DamageType
}

func weaponStyle(id CombatStyleID, damageType DamageType) weaponStyleDefinition {
	return weaponStyleDefinition{ID: id, DamageType: damageType}
}

func swordWeapon(definition WeaponDefinition) WeaponDefinition {
	return weaponWithStyles(definition, CombatStyleStab,
		weaponStyle(CombatStyleStab, DamageTypeStab),
		weaponStyle(CombatStyleSlash, DamageTypeSlash),
		weaponStyle(CombatStyleDefensive, DamageTypeSlash),
	)
}

func weaponWithStyles(definition WeaponDefinition, defaultStyle CombatStyleID, styles ...weaponStyleDefinition) WeaponDefinition {
	definition.DefaultStyle = defaultStyle
	definition.CombatStyles = make([]CombatStyleID, 0, len(styles))
	definition.StyleDamageTypes = make(map[CombatStyleID]DamageType, len(styles))
	for _, style := range styles {
		definition.CombatStyles = append(definition.CombatStyles, style.ID)
		definition.StyleDamageTypes[style.ID] = style.DamageType
	}
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
		if definition.Ammunition && (definition.Kind != ItemKindResource || definition.AmmunitionFamily == "") {
			return nil, fmt.Errorf("ammunition %q must be a resource with a family", definition.ID)
		}
		seenRequirements := make(map[SkillID]struct{}, len(definition.EquipRequirements))
		for _, requirement := range definition.EquipRequirements {
			if !skillIDValid(requirement.Skill) || requirement.Level < 1 || requirement.Level > MaximumSkillLevel {
				return nil, fmt.Errorf("item %q has invalid equipment requirement %+v", definition.ID, requirement)
			}
			if _, duplicate := seenRequirements[requirement.Skill]; duplicate {
				return nil, fmt.Errorf("item %q repeats its %q requirement", definition.ID, requirement.Skill)
			}
			seenRequirements[requirement.Skill] = struct{}{}
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
			if definition.Armour.Slot == "" || definition.Armour.HealthBonus < 0 {
				return nil, fmt.Errorf("armour %q requires a slot and non-negative health bonus", definition.ID)
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
			if definition.StackLimit == 0 || definition.Consumable.CooldownPulses == 0 ||
				definition.Consumable.Healing < 0 || definition.Consumable.PrayerRestoration < 0 ||
				(definition.Consumable.Healing == 0 && definition.Consumable.PrayerRestoration == 0 && len(definition.Consumable.SkillBoosts) == 0) {
				return nil, fmt.Errorf("consumable %q requires a stack limit, positive effect, and cooldown", definition.ID)
			}
			if len(definition.Consumable.SkillBoosts) > 0 && definition.Consumable.BoostDuration == 0 {
				return nil, fmt.Errorf("consumable %q has boosts without a duration", definition.ID)
			}
			for _, boost := range definition.Consumable.SkillBoosts {
				if !skillIDValid(boost.Skill) || boost.Levels <= 0 {
					return nil, fmt.Errorf("consumable %q has invalid skill boost %+v", definition.ID, boost)
				}
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
		if ammunition.ItemID == "" || ammunition.Quantity < 0 || !exists || ammunitionItem.Kind != ItemKindResource ||
			definition.Weapon.AmmunitionFamily == "" || ammunitionItem.AmmunitionFamily != definition.Weapon.AmmunitionFamily {
			return nil, fmt.Errorf("weapon %q has invalid ammunition %+v", definition.ID, ammunition)
		}
	}
	return catalog, nil
}

func unmetEquipRequirement(player *playerState, definition ItemDefinition) (SkillRequirement, bool) {
	for _, requirement := range definition.EquipRequirements {
		level := 1
		if player != nil {
			level = SkillLevelForExperience(player.skills[requirement.Skill])
		}
		if level < requirement.Level {
			return requirement, true
		}
	}
	return SkillRequirement{}, false
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
		damageType, hasDamageType := weapon.StyleDamageTypes[style]
		if !hasDamageType || !knownDamageType(damageType) {
			return fmt.Errorf("combat style %q requires a valid damage type", style)
		}
		defaultAllowed = defaultAllowed || style == weapon.DefaultStyle
	}
	if !defaultAllowed {
		return fmt.Errorf("default combat style %q is not allowed", weapon.DefaultStyle)
	}
	return nil
}

func knownDamageType(damageType DamageType) bool {
	switch damageType {
	case DamageTypeStab, DamageTypeSlash, DamageTypeCrush, DamageTypeRanged, DamageTypeMagic:
		return true
	default:
		return false
	}
}

func knownCombatStyle(style CombatStyleID) bool {
	switch style {
	case CombatStyleSlash, CombatStyleStab, CombatStyleRapid, CombatStyleAccurate,
		CombatStyleLongRange, CombatStyleCrush, CombatStyleHeavyCrush,
		CombatStyleDefensive:
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
