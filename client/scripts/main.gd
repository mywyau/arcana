extends Node3D

const PLAYER_SCENE: PackedScene = preload("res://scenes/player.tscn")
const TARGET_SCENE: PackedScene = preload("res://scenes/target_dummy.tscn")
const LOOT_SCENE: PackedScene = preload("res://scenes/loot_drop.tscn")
const GATHERING_SCENE: PackedScene = preload("res://scenes/gathering_spot.tscn")
const BANK_CHEST_SCENE: PackedScene = preload("res://scenes/bank_chest.tscn")
const INVENTORY_SLOT_SCRIPT: Script = preload("res://scripts/inventory_slot.gd")
const RAY_LENGTH: float = 1000.0
const GROUND_COLLISION_MASK: int = 1
const UPPER_COLLISION_MASK: int = 2
const CONNECTOR_COLLISION_MASK: int = 4
const TARGET_COLLISION_MASK: int = 8
const LOOT_COLLISION_MASK: int = 16
const GATHERING_COLLISION_MASK: int = 32
const BANK_COLLISION_MASK: int = 64
const ALL_WORLD_COLLISION_MASK: int = 7
const HUD_LAYOUT_PATH: String = "user://hud_layout.cfg"
const CLIENT_SETTINGS_PATH: String = "user://client_settings.cfg"
const HUD_MINIMUM_SCALE: float = 0.75
const HUD_MAXIMUM_SCALE: float = 1.5
const HUD_HANDLE_HEIGHT: float = 28.0
const HUD_RESIZE_HANDLE_SIZE: float = 30.0
const INVENTORY_TOTAL_SLOTS: int = 50
const INVENTORY_PAGE_SIZE: int = 25
const MENU_HUD_BASE_SIZE: Vector2 = Vector2(600.0, 650.0)
const CONTEXT_LOOT_RADIUS_PIXELS: float = 42.0
const BANK_TAKE_ONE: int = 0
const BANK_TAKE_X: int = 1
const BANK_TAKE_ALL: int = 2
const TEST_SUPPLY_CHEST_PREFIX: String = "test_"
const UNCUT_GEMS: Dictionary = {
	"uncut_sapphire": "Sapphire", "uncut_emerald": "Emerald",
	"uncut_ruby": "Ruby", "uncut_diamond": "Diamond",
}
const ITEM_DISPLAY_NAMES: Dictionary = {
	"amulet_ferocity": "Ruby Amulet", "amulet_hawk": "Emerald Amulet",
	"amulet_embers": "Sapphire Amulet", "amulet_fury": "Diamond Amulet",
}
const CRAFTING_RECIPES: Dictionary = {
	"fishing": [
		{"id": "raw_minnow", "name": "Minnow", "level": 1, "inputs": "Pond"},
		{"id": "raw_trout", "name": "Trout", "level": 2, "inputs": "Pond"},
		{"id": "raw_salmon", "name": "Salmon", "level": 3, "inputs": "Pond"},
		{"id": "raw_tuna", "name": "Tuna", "level": 4, "inputs": "Pond"},
		{"id": "raw_lobster", "name": "Lobster", "level": 5, "inputs": "Pond"},
	],
	"cooking": [
		{"id": "cooked_minnow", "name": "Minnow", "level": 1, "input_id": "raw_minnow", "inputs": "1 Raw Minnow"},
		{"id": "cooked_trout", "name": "Trout", "level": 2, "input_id": "raw_trout", "inputs": "1 Raw Trout"},
		{"id": "cooked_salmon", "name": "Salmon", "level": 3, "input_id": "raw_salmon", "inputs": "1 Raw Salmon"},
		{"id": "cooked_tuna", "name": "Tuna", "level": 4, "input_id": "raw_tuna", "inputs": "1 Raw Tuna"},
		{"id": "cooked_lobster", "name": "Lobster", "level": 5, "input_id": "raw_lobster", "inputs": "1 Raw Lobster"},
	],
	"smelting": [
		{"id": "bronze_bar", "name": "Bronze Bar", "level": 1, "inputs": "1 Copper Ore + 1 Tin Ore"},
		{"id": "iron_bar", "name": "Iron Bar", "level": 2, "inputs": "1 Iron Ore + 1 Coal"},
		{"id": "mithril_bar", "name": "Mithril Bar", "level": 5, "inputs": "1 Mithril Ore + 2 Coal"},
		{"id": "gold_bar", "name": "Gold Bar", "level": 4, "inputs": "1 Gold Ore"},
	],
	"forging": [
		{"id": "bronze_sword", "name": "Bronze Sword", "tier": "bronze", "level": 1, "inputs": "2 Bronze Bars"},
		{"id": "bronze_helm", "name": "Bronze Helm", "tier": "bronze", "level": 1, "inputs": "2 Bronze Bars"},
		{"id": "bronze_platebody", "name": "Bronze Platebody", "tier": "bronze", "level": 2, "inputs": "5 Bronze Bars"},
		{"id": "bronze_pickaxe", "name": "Bronze Pickaxe", "tier": "bronze", "level": 1, "inputs": "3 Bronze Bars"},
		{"id": "bronze_shield", "name": "Bronze Shield", "tier": "bronze", "level": 2, "inputs": "3 Bronze Bars"},
		{"id": "bronze_legs", "name": "Bronze Legs", "tier": "bronze", "level": 2, "inputs": "4 Bronze Bars"},
		{"id": "bronze_gloves", "name": "Bronze Gloves", "tier": "bronze", "level": 1, "inputs": "1 Bronze Bar"},
		{"id": "bronze_boots", "name": "Bronze Boots", "tier": "bronze", "level": 1, "inputs": "2 Bronze Bars"},
		{"id": "iron_sword", "name": "Iron Sword", "tier": "iron", "level": 3, "inputs": "2 Iron Bars"},
		{"id": "iron_arrow_tips", "name": "20 Iron Arrow Tips", "tier": "iron", "level": 2, "inputs": "1 Iron Bar"},
		{"id": "iron_helm", "name": "Iron Helm", "tier": "iron", "level": 3, "inputs": "2 Iron Bars"},
		{"id": "iron_platebody", "name": "Iron Platebody", "tier": "iron", "level": 4, "inputs": "5 Iron Bars"},
		{"id": "iron_shield", "name": "Iron Shield", "tier": "iron", "level": 4, "inputs": "3 Iron Bars"},
		{"id": "iron_legs", "name": "Iron Legs", "tier": "iron", "level": 4, "inputs": "4 Iron Bars"},
		{"id": "iron_gloves", "name": "Iron Gloves", "tier": "iron", "level": 3, "inputs": "1 Iron Bar"},
		{"id": "iron_boots", "name": "Iron Boots", "tier": "iron", "level": 3, "inputs": "2 Iron Bars"},
		{"id": "mithril_sword", "name": "Mithril Sword", "tier": "mithril", "level": 6, "inputs": "3 Mithril Bars"},
		{"id": "mithril_helm", "name": "Mithril Helm", "tier": "mithril", "level": 6, "inputs": "3 Mithril Bars"},
		{"id": "mithril_platebody", "name": "Mithril Platebody", "tier": "mithril", "level": 7, "inputs": "6 Mithril Bars"},
		{"id": "mithril_shield", "name": "Mithril Shield", "tier": "mithril", "level": 7, "inputs": "4 Mithril Bars"},
		{"id": "mithril_legs", "name": "Mithril Legs", "tier": "mithril", "level": 7, "inputs": "5 Mithril Bars"},
		{"id": "mithril_gloves", "name": "Mithril Gloves", "tier": "mithril", "level": 6, "inputs": "2 Mithril Bars"},
		{"id": "mithril_boots", "name": "Mithril Boots", "tier": "mithril", "level": 6, "inputs": "3 Mithril Bars"},
		{"id": "sapphire_amulet", "name": "Sapphire Amulet", "tier": "gold", "level": 3, "inputs": "1 Gold Bar + 1 Sapphire"},
		{"id": "emerald_amulet", "name": "Emerald Amulet", "tier": "gold", "level": 4, "inputs": "1 Gold Bar + 1 Emerald"},
		{"id": "ruby_amulet", "name": "Ruby Amulet", "tier": "gold", "level": 5, "inputs": "1 Gold Bar + 1 Ruby"},
		{"id": "diamond_amulet", "name": "Diamond Amulet", "tier": "gold", "level": 7, "inputs": "1 Gold Bar + 1 Diamond"},
	],
	"prayer_altar": [
		{"id": "restore_prayer", "name": "Restore Prayer", "level": 1, "inputs": "Restore all prayer points"},
		{"id": "sacrifice_bones", "name": "Sacrifice All Bones", "level": 1, "inputs": "1 bone every 600 ms • 10 Prayer XP each"},
		{"id": "adopt_common_litany", "name": "Adopt Common Litany", "level": 1, "inputs": "Swap prayer set", "altar_recipe": "adopt_common_litany"},
		{"id": "adopt_war_covenant", "name": "Adopt War Covenant", "level": 1, "inputs": "Swap prayer set", "altar_recipe": "adopt_war_covenant"},
	],
}

const PLACEHOLDER_SKILLS: Array = [
	{"name": "Attack", "level": 1},
	{"name": "Magic", "level": 1},
	{"name": "Range", "level": 1},
	{"name": "Defence", "level": 1},
	{"name": "Strength", "level": 1},
	{"name": "Prayer", "level": 1},
	{"name": "Vitality", "level": 1},
	{"name": "Agility", "level": 1},
	{"name": "Mining", "level": 1},
	{"name": "Smithing", "level": 1},
	{"name": "Crafting", "level": 1},
	{"name": "Woodcutting", "level": 1},
	{"name": "Runecrafting", "level": 1},
	{"name": "Herblore", "level": 1},
	{"name": "Fletching", "level": 1},
	{"name": "Fishing", "level": 1},
	{"name": "Cooking", "level": 1},
	{"name": "Farming", "level": 1},
	{"name": "Thieving", "level": 1},
	{"name": "Construction", "level": 1},
]
const ACTIVE_SKILL_IDS: Dictionary = {
	"Attack": "attack",
	"Strength": "strength",
	"Defence": "defence",
	"Range": "range",
	"Magic": "magic",
	"Agility": "agility",
	"Fishing": "fishing",
	"Cooking": "cooking",
	"Mining": "mining",
	"Smithing": "smithing",
	"Woodcutting": "woodcutting",
	"Herblore": "herblore",
	"Fletching": "fletching",
	"Prayer": "prayer",
}
const ACTIVE_SKILL_NAMES: Dictionary = {
	"attack": "Attack",
	"strength": "Strength",
	"defence": "Defence",
	"range": "Range",
	"magic": "Magic",
	"agility": "Agility",
	"fishing": "Fishing",
	"cooking": "Cooking",
	"mining": "Mining",
	"smithing": "Smithing",
	"woodcutting": "Woodcutting",
	"herblore": "Herblore",
	"fletching": "Fletching",
	"prayer": "Prayer",
}

const SWORD_ICON: Texture2D = IconCatalog.SWORD
const DAGGER_ICON: Texture2D = IconCatalog.DAGGER
const BOW_ICON: Texture2D = IconCatalog.BOW
const MACE_ICON: Texture2D = IconCatalog.MACE
const BATTLE_AXE_ICON: Texture2D = IconCatalog.BATTLE_AXE
const STAFF_ICON: Texture2D = IconCatalog.STAFF
const EMPTY_SLOT_ICON: Texture2D = IconCatalog.EMPTY_SLOT
const TAB_STYLES_ICON: Texture2D = IconCatalog.TAB_STYLES
const TAB_INVENTORY_ICON: Texture2D = IconCatalog.TAB_INVENTORY
const TAB_SKILLS_ICON: Texture2D = IconCatalog.TAB_SKILLS
const TAB_EQUIPMENT_ICON: Texture2D = IconCatalog.TAB_EQUIPMENT
const TAB_MAGIC_ICON: Texture2D = IconCatalog.TAB_MAGIC
const TAB_PRAYER_ICON: Texture2D = IconCatalog.TAB_PRAYER
const TAB_SETTINGS_ICON: Texture2D = IconCatalog.TAB_SETTINGS
const PLACEHOLDER_SPELLS: Array = IconCatalog.SPELLS
const PLACEHOLDER_PRAYERS: Array = IconCatalog.PRAYERS


const COMBAT_STYLES: Dictionary = {
	"unarmed": [
		{"id": "crush", "name": "Punch", "training": "Attack"},
		{"id": "heavy_crush", "name": "Heavy Punch", "training": "Strength"},
		{"id": "defensive", "name": "Defensive", "training": "Defence"},
		{"id": "unfocused", "name": "Unfocused", "training": ""},
	],
	"sword": [
		{"id": "stab", "name": "Stab", "training": "Attack"},
		{"id": "slash", "name": "Slash", "training": "Strength"},
		{"id": "defensive", "name": "Defensive", "training": "Defence"},
		{"id": "unfocused", "name": "Unfocused", "training": ""},
	],
	"bronze_sword": [
		{"id": "stab", "name": "Stab", "training": "Attack"},
		{"id": "slash", "name": "Slash", "training": "Strength"},
		{"id": "defensive", "name": "Defensive", "training": "Defence"},
		{"id": "unfocused", "name": "Unfocused", "training": ""},
	],
	"steel_sword": [
		{"id": "stab", "name": "Stab", "training": "Attack"},
		{"id": "slash", "name": "Slash", "training": "Strength"},
		{"id": "defensive", "name": "Defensive", "training": "Defence"},
		{"id": "unfocused", "name": "Unfocused", "training": ""},
	],
	"mithril_sword": [
		{"id": "stab", "name": "Stab", "training": "Attack"},
		{"id": "slash", "name": "Slash", "training": "Strength"},
		{"id": "defensive", "name": "Defensive", "training": "Defence"},
		{"id": "unfocused", "name": "Unfocused", "training": ""},
	],
	"rune_sword": [
		{"id": "stab", "name": "Stab", "training": "Attack"},
		{"id": "slash", "name": "Slash", "training": "Strength"},
		{"id": "defensive", "name": "Defensive", "training": "Defence"},
		{"id": "unfocused", "name": "Unfocused", "training": ""},
	],
	"dagger": [
		{"id": "stab", "name": "Stab", "training": "Attack"},
		{"id": "slash", "name": "Slash", "training": "Strength"},
		{"id": "defensive", "name": "Defensive", "training": "Defence"},
		{"id": "unfocused", "name": "Unfocused", "training": ""},
	],
	"mace": [
		{"id": "crush", "name": "Crush", "training": "Attack"},
		{"id": "heavy_crush", "name": "Heavy Crush", "training": "Strength"},
		{"id": "defensive", "name": "Defensive", "training": "Defence"},
		{"id": "unfocused", "name": "Unfocused", "training": ""},
	],
	"battle_axe": [
		{"id": "slash", "name": "Measured Slash", "training": "Attack"},
		{"id": "heavy_crush", "name": "Heavy Crush", "training": "Strength"},
		{"id": "defensive", "name": "Defensive", "training": "Defence"},
		{"id": "unfocused", "name": "Unfocused", "training": ""},
	],
	"spear": [
		{"id": "stab", "name": "Stab", "training": "Attack"},
		{"id": "slash", "name": "Sweep", "training": "Strength"},
		{"id": "defensive", "name": "Defensive Stab", "training": "Defence"},
		{"id": "unfocused", "name": "Unfocused", "training": ""},
	],
	"halberd": [
		{"id": "slash", "name": "Cleave", "training": "Strength"},
		{"id": "stab", "name": "Thrust", "training": "Attack"},
		{"id": "defensive", "name": "Defensive Cleave", "training": "Defence"},
		{"id": "unfocused", "name": "Unfocused", "training": ""},
	],
	"bow": [
		{"id": "rapid", "name": "Rapid", "training": "Range"},
		{"id": "accurate", "name": "Accurate", "training": "Range"},
		{"id": "long_range", "name": "Long Range", "training": "Range"},
		{"id": "unfocused", "name": "Unfocused", "training": ""},
	],
	"crossbow": [
		{"id": "rapid", "name": "Rapid", "training": "Range"},
		{"id": "accurate", "name": "Accurate", "training": "Range"},
		{"id": "long_range", "name": "Long Range", "training": "Range"},
		{"id": "unfocused", "name": "Unfocused", "training": ""},
	],
	"staff": [
		{"id": "rapid", "name": "Rapid", "training": "Magic"},
		{"id": "accurate", "name": "Accurate", "training": "Magic"},
		{"id": "long_range", "name": "Long Range", "training": "Magic"},
		{"id": "unfocused", "name": "Unfocused", "training": ""},
	],
}
const EQUIPMENT_ICONS: Dictionary = IconCatalog.WEAPONS
const ARMOUR_SLOTS: Dictionary = {
	"bronze_helm": "head",
	"bronze_platebody": "body",
	"bronze_shield": "left_hand", "bronze_legs": "legs", "bronze_gloves": "hands", "bronze_boots": "feet",
	"iron_helm": "head",
	"iron_platebody": "body",
	"iron_shield": "left_hand", "iron_legs": "legs", "iron_gloves": "hands", "iron_boots": "feet",
	"mithril_helm": "head",
	"mithril_platebody": "body",
	"mithril_shield": "left_hand", "mithril_legs": "legs", "mithril_gloves": "hands", "mithril_boots": "feet",
	"training_helm": "head",
	"training_plate": "body",
	"swift_helm": "head",
	"swift_body": "body",
	"swift_boots": "feet",
	"novice_robe_hood": "head", "novice_robe_top": "body", "novice_robe_gloves": "hands", "novice_robe_bottoms": "legs", "novice_robe_boots": "feet",
	"novice_ranger_hood": "head", "novice_ranger_body": "body", "novice_ranger_gloves": "hands", "novice_ranger_legs": "legs", "novice_ranger_boots": "feet",
	"amulet_ferocity": "neck",
	"amulet_hawk": "neck",
	"amulet_embers": "neck",
	"amulet_fury": "neck",
}
const REINFORCEMENT_TIERS: Dictionary = {
	"bronze": {"bar": "bronze_bar", "bar_name": "Bronze Bars", "level": 2, "items": ["bronze_sword", "bronze_helm", "bronze_platebody", "bronze_shield", "bronze_legs", "bronze_gloves", "bronze_boots"]},
	"iron": {"bar": "iron_bar", "bar_name": "Iron Bars", "level": 5, "items": ["sword", "iron_helm", "iron_platebody", "iron_shield", "iron_legs", "iron_gloves", "iron_boots"]},
	"mithril": {"bar": "mithril_bar", "bar_name": "Mithril Bars", "level": 8, "items": ["mithril_sword", "mithril_helm", "mithril_platebody", "mithril_shield", "mithril_legs", "mithril_gloves", "mithril_boots"]},
}
const MAXIMUM_REINFORCEMENT: int = 5
const AMMUNITION_ITEMS: Dictionary = {"iron_arrow": true}
const ITEM_COMBAT_BONUSES: Dictionary = {
	"amulet_ferocity": "+5 melee accuracy • +5 melee strength",
	"amulet_hawk": "+5 ranged accuracy • +5 ranged strength",
	"amulet_embers": "+5 magic accuracy • +5 magic strength",
	"amulet_fury": "+2 accuracy • +2 strength for melee, ranged, and magic",
}
const EQUIPMENT_REQUIREMENTS: Dictionary = {
	"steel_sword": "Attack level 5",
	"mithril_sword": "Attack level 10",
	"rune_sword": "Attack level 20",
	"mace": "Attack level 5",
	"battle_axe": "Attack level 10",
	"iron_helm": "Defence level 5",
	"iron_platebody": "Defence level 5",
	"iron_shield": "Defence level 5", "iron_legs": "Defence level 5", "iron_gloves": "Defence level 5", "iron_boots": "Defence level 5",
	"mithril_helm": "Defence level 10",
	"mithril_platebody": "Defence level 10",
	"mithril_shield": "Defence level 10", "mithril_legs": "Defence level 10", "mithril_gloves": "Defence level 10", "mithril_boots": "Defence level 10",
	"swift_helm": "Defence level 5",
	"swift_body": "Defence level 5",
	"swift_boots": "Defence level 5",
	"amulet_ferocity": "Attack level 5",
	"amulet_hawk": "Range level 5",
	"amulet_embers": "Magic level 5",
	"amulet_fury": "Defence level 5",
}
const SPELLBOOK_ITEMS: Dictionary = {
	"standard_spellbook": "standard",
}
const CONSUMABLE_ITEMS: Dictionary = {
	"cooked_minnow": {"healing": 8, "stack_limit": 10},
	"cooked_trout": {"healing": 12, "stack_limit": 10},
	"cooked_salmon": {"healing": 16, "stack_limit": 10},
	"cooked_tuna": {"healing": 20, "stack_limit": 10},
	"cooked_lobster": {"healing": 25, "stack_limit": 10},
	"prayer_potion": {"prayer_restore": 25, "stack_limit": 5},
	"minor_healing_potion": {"healing": 15, "stack_limit": 5},
	"murky_brew": {"healing": 5, "stack_limit": 5},
}
const BREWING_INGREDIENTS: Dictionary = {
	"vial_of_water": {"name": "Vial of Water", "role": "Base"},
	"sunleaf": {"name": "Sunleaf", "role": "Herb"},
	"moonleaf": {"name": "Moonleaf", "role": "Herb"},
	"sacred_dust": {"name": "Sacred Dust", "role": "Secondary"},
	"fish_oil": {"name": "Fish Oil", "role": "Secondary"},
}
const KNOWN_BREWS: Dictionary = {
	"sunleaf+vial_of_water": "Minor Healing Potion",
	"sacred_dust+sunleaf+vial_of_water": "Prayer Potion",
}
const FLETCHING_RECIPES: Dictionary = {
	"arrow_shafts": {
		"name": "20 Arrow Shafts",
		"costs": {"normal_logs": 1},
		"tool": "knife",
		"experience": 10,
	},
	"iron_arrows": {
		"name": "20 Iron Arrows",
		"costs": {"arrow_shafts": 20, "feathers": 20, "iron_arrow_tips": 20},
		"tool": "",
		"experience": 20,
	},
}
const FLETCHING_INPUTS: Dictionary = {
	"knife": true,
	"normal_logs": true,
	"arrow_shafts": true,
	"feathers": true,
	"iron_arrow_tips": true,
}
const ITEM_ICONS: Dictionary = IconCatalog.ITEMS
const EQUIPMENT_SLOT_ICONS: Dictionary = IconCatalog.EQUIPMENT_SLOTS
const SKILL_ICONS: Dictionary = IconCatalog.SKILLS

@onready var camera_rig: GuidedCamera = $CameraRig
@onready var camera: Camera3D = $CameraRig/Yaw/Pitch/Camera3D
@onready var ground: PrototypeTerrain = $Ground
@onready var network: NetworkClient = $NetworkClient
@onready var smithing_window: Window = $SmithingWindow
@onready var smithing_tabs: TabContainer = $SmithingWindow/Panel/Margin/VBox/Tabs
@onready var cooking_window: Window = $CookingWindow
@onready var cooking_grid: GridContainer = $CookingWindow/Panel/Margin/VBox/RecipeScroll/Grid
@onready var brewing_window: Window = $BrewingWindow
@onready var brewing_slots: Array[Button] = [
	$BrewingWindow/Panel/Margin/VBox/Selected/Slot1,
	$BrewingWindow/Panel/Margin/VBox/Selected/Slot2,
	$BrewingWindow/Panel/Margin/VBox/Selected/Slot3,
]
@onready var brewing_result: Label = $BrewingWindow/Panel/Margin/VBox/Result
@onready var brewing_grid: GridContainer = $BrewingWindow/Panel/Margin/VBox/IngredientScroll/Grid
@onready var brewing_clear_button: Button = $BrewingWindow/Panel/Margin/VBox/Actions/Clear
@onready var brewing_brew_button: Button = $BrewingWindow/Panel/Margin/VBox/Actions/Brew
@onready var brewing_make_x_button: Button = $BrewingWindow/Panel/Margin/VBox/Actions/MakeX
@onready var brewing_make_all_button: Button = $BrewingWindow/Panel/Margin/VBox/Actions/MakeAll
@onready var fletching_window: Window = $FletchingWindow
@onready var fletching_shafts_button: Button = $FletchingWindow/Panel/Margin/VBox/Recipes/Shafts
@onready var fletching_arrows_button: Button = $FletchingWindow/Panel/Margin/VBox/Recipes/Arrows
@onready var fletching_selection: Label = $FletchingWindow/Panel/Margin/VBox/Selection
@onready var fletching_make_one_button: Button = $FletchingWindow/Panel/Margin/VBox/Actions/Make1
@onready var fletching_make_x_button: Button = $FletchingWindow/Panel/Margin/VBox/Actions/MakeX
@onready var fletching_make_all_button: Button = $FletchingWindow/Panel/Margin/VBox/Actions/MakeAll
@onready var players: Node3D = $Players
@onready var targets: Node3D = $Targets
@onready var drops: Node3D = $Drops
@onready var gathering_spots: Node3D = $GatheringSpots
@onready var bank_chests: Node3D = $BankChests
@onready var status_label: Label = $UI/Margin/Status
@onready var diagnostics_label: Label = $UI/DiagnosticsMargin/Diagnostics
@onready var hud_edit_hint: Label = $UI/HudEditHint
@onready var player_vitals_hud: PanelContainer = $UI/PlayerVitalsHUD
@onready var combat_hud: PanelContainer = $UI/CombatHUD
@onready var inventory_hud: PanelContainer = $UI/InventoryHUD
@onready var skills_hud: PanelContainer = $UI/SkillsHUD
@onready var equipment_hud: PanelContainer = $UI/EquipmentHUD
@onready var magic_hud: PanelContainer = $UI/MagicHUD
@onready var prayer_hud: PanelContainer = $UI/PrayerHUD
@onready var styles_hud: PanelContainer = $UI/StylesHUD
@onready var settings_hud: PanelContainer = $UI/SettingsHUD
@onready var inventory_settings_button: Button = $UI/InventoryHUD/Margin/VBox/UtilityTabs/Margin/Row/Settings
@onready var settings_inventory_button: Button = $UI/SettingsHUD/Margin/VBox/UtilityTabs/Margin/Row/Inventory
@onready var settings_diagnostics_toggle: CheckButton = $UI/SettingsHUD/Margin/VBox/ShowDiagnostics
@onready var settings_music_slider: HSlider = $UI/SettingsHUD/Margin/VBox/MusicVolume/Slider
@onready var settings_music_value: Label = $UI/SettingsHUD/Margin/VBox/MusicVolume/Value
@onready var settings_edit_hud_button: Button = $UI/SettingsHUD/Margin/VBox/EditHUD
@onready var settings_reset_hud_button: Button = $UI/SettingsHUD/Margin/VBox/ResetHUD
@onready var bank_hud: PanelContainer = $UI/BankHUD
@onready var bank_title: Label = $UI/BankHUD/Margin/VBox/Header/Title
@onready var bank_hint: Label = $UI/BankHUD/Margin/VBox/Hint
@onready var bank_close_button: Button = $UI/BankHUD/Margin/VBox/Header/Close
@onready var bank_summary: Label = $UI/BankHUD/Margin/VBox/Summary
@onready var bank_grid: GridContainer = $UI/BankHUD/Margin/VBox/ContentScroll/Grid
@onready var combat_tab_button: Button = $UI/CombatHUD/Margin/VBox/TabRow/Combat
@onready var inventory_tab_button: Button = $UI/CombatHUD/Margin/VBox/TabRow/Inventory
@onready var combat_skills_tab_button: Button = $UI/CombatHUD/Margin/VBox/TabRow/Skills
@onready var combat_equipment_tab_button: Button = $UI/CombatHUD/Margin/VBox/TabRow/Equipment
@onready var combat_magic_tab_button: Button = $UI/CombatHUD/Margin/VBox/TabRow/Magic
@onready var combat_styles_tab_button: Button = $UI/CombatHUD/Margin/VBox/TabRow/Styles
@onready var inventory_combat_tab_button: Button = $UI/InventoryHUD/Margin/VBox/TabRow/Combat
@onready var inventory_active_tab_button: Button = $UI/InventoryHUD/Margin/VBox/TabRow/Inventory
@onready var inventory_skills_tab_button: Button = $UI/InventoryHUD/Margin/VBox/TabRow/Skills
@onready var inventory_equipment_tab_button: Button = $UI/InventoryHUD/Margin/VBox/TabRow/Equipment
@onready var inventory_magic_tab_button: Button = $UI/InventoryHUD/Margin/VBox/TabRow/Magic
@onready var inventory_prayer_tab_button: Button = $UI/InventoryHUD/Margin/VBox/TabRow/Prayer
@onready var inventory_styles_tab_button: Button = $UI/InventoryHUD/Margin/VBox/TabRow/Styles
@onready var inventory_summary: Label = $UI/InventoryHUD/Margin/VBox/Summary
@onready var inventory_previous_page_button: Button = $UI/InventoryHUD/Margin/VBox/PageRow/Previous
@onready var inventory_page_label: Label = $UI/InventoryHUD/Margin/VBox/PageRow/Label
@onready var inventory_next_page_button: Button = $UI/InventoryHUD/Margin/VBox/PageRow/Next
@onready var inventory_grid: GridContainer = $UI/InventoryHUD/Margin/VBox/ContentScroll/Grid
@onready var skills_combat_tab_button: Button = $UI/SkillsHUD/Margin/VBox/TabRow/Combat
@onready var skills_inventory_tab_button: Button = $UI/SkillsHUD/Margin/VBox/TabRow/Inventory
@onready var skills_active_tab_button: Button = $UI/SkillsHUD/Margin/VBox/TabRow/Skills
@onready var skills_equipment_tab_button: Button = $UI/SkillsHUD/Margin/VBox/TabRow/Equipment
@onready var skills_magic_tab_button: Button = $UI/SkillsHUD/Margin/VBox/TabRow/Magic
@onready var skills_prayer_tab_button: Button = $UI/SkillsHUD/Margin/VBox/TabRow/Prayer
@onready var skills_styles_tab_button: Button = $UI/SkillsHUD/Margin/VBox/TabRow/Styles
@onready var skills_summary: Label = $UI/SkillsHUD/Margin/VBox/Summary
@onready var skills_grid: GridContainer = $UI/SkillsHUD/Margin/VBox/ContentScroll/Grid
@onready var equipment_combat_tab_button: Button = $UI/EquipmentHUD/Margin/VBox/TabRow/Combat
@onready var equipment_inventory_tab_button: Button = $UI/EquipmentHUD/Margin/VBox/TabRow/Inventory
@onready var equipment_skills_tab_button: Button = $UI/EquipmentHUD/Margin/VBox/TabRow/Skills
@onready var equipment_active_tab_button: Button = $UI/EquipmentHUD/Margin/VBox/TabRow/Equipment
@onready var equipment_magic_tab_button: Button = $UI/EquipmentHUD/Margin/VBox/TabRow/Magic
@onready var equipment_prayer_tab_button: Button = $UI/EquipmentHUD/Margin/VBox/TabRow/Prayer
@onready var equipment_styles_tab_button: Button = $UI/EquipmentHUD/Margin/VBox/TabRow/Styles
@onready var equipment_grid: GridContainer = $UI/EquipmentHUD/Margin/VBox/ContentScroll/Grid
@onready var detailed_stats_button: Button = $UI/EquipmentHUD/Margin/VBox/DetailedStats
@onready var combat_stats_hud: PanelContainer = $UI/CombatStatsHUD
@onready var combat_stats_label: Label = $UI/CombatStatsHUD/Margin/VBox/Stats
@onready var combat_stats_close_button: Button = $UI/CombatStatsHUD/Margin/VBox/Header/Close
@onready var combat_stats_offence_button: Button = $UI/CombatStatsHUD/Margin/VBox/PageRow/Offence
@onready var combat_stats_defence_button: Button = $UI/CombatStatsHUD/Margin/VBox/PageRow/Defence
@onready var combat_stats_effects_button: Button = $UI/CombatStatsHUD/Margin/VBox/PageRow/Effects
@onready var magic_combat_tab_button: Button = $UI/MagicHUD/Margin/VBox/TabRow/Combat
@onready var magic_inventory_tab_button: Button = $UI/MagicHUD/Margin/VBox/TabRow/Inventory
@onready var magic_skills_tab_button: Button = $UI/MagicHUD/Margin/VBox/TabRow/Skills
@onready var magic_equipment_tab_button: Button = $UI/MagicHUD/Margin/VBox/TabRow/Equipment
@onready var magic_active_tab_button: Button = $UI/MagicHUD/Margin/VBox/TabRow/Magic
@onready var magic_prayer_tab_button: Button = $UI/MagicHUD/Margin/VBox/TabRow/Prayer
@onready var magic_styles_tab_button: Button = $UI/MagicHUD/Margin/VBox/TabRow/Styles
@onready var magic_grid: GridContainer = $UI/MagicHUD/Margin/VBox/ContentScroll/Grid
@onready var prayer_inventory_tab_button: Button = $UI/PrayerHUD/Margin/VBox/TabRow/Inventory
@onready var prayer_skills_tab_button: Button = $UI/PrayerHUD/Margin/VBox/TabRow/Skills
@onready var prayer_equipment_tab_button: Button = $UI/PrayerHUD/Margin/VBox/TabRow/Equipment
@onready var prayer_magic_tab_button: Button = $UI/PrayerHUD/Margin/VBox/TabRow/Magic
@onready var prayer_active_tab_button: Button = $UI/PrayerHUD/Margin/VBox/TabRow/Prayer
@onready var prayer_styles_tab_button: Button = $UI/PrayerHUD/Margin/VBox/TabRow/Styles
@onready var prayer_grid: GridContainer = $UI/PrayerHUD/Margin/VBox/ContentScroll/Grid
@onready var prayer_set_label: Label = $UI/PrayerHUD/Margin/VBox/PrayerSet
@onready var styles_combat_tab_button: Button = $UI/StylesHUD/Margin/VBox/TabRow/Combat
@onready var styles_inventory_tab_button: Button = $UI/StylesHUD/Margin/VBox/TabRow/Inventory
@onready var styles_skills_tab_button: Button = $UI/StylesHUD/Margin/VBox/TabRow/Skills
@onready var styles_equipment_tab_button: Button = $UI/StylesHUD/Margin/VBox/TabRow/Equipment
@onready var styles_magic_tab_button: Button = $UI/StylesHUD/Margin/VBox/TabRow/Magic
@onready var styles_prayer_tab_button: Button = $UI/StylesHUD/Margin/VBox/TabRow/Prayer
@onready var styles_active_tab_button: Button = $UI/StylesHUD/Margin/VBox/TabRow/Styles
@onready var styles_grid: GridContainer = $UI/StylesHUD/Margin/VBox/ContentScroll/Grid
@onready var target_hud: PanelContainer = $UI/TargetHUD
@onready var player_health_bar: ProgressBar = $UI/PlayerVitalsHUD/Margin/VBox/Health
@onready var player_health_label: Label = $UI/PlayerVitalsHUD/Margin/VBox/HealthLabel
@onready var player_prayer_bar: ProgressBar = $UI/PlayerVitalsHUD/Margin/VBox/Prayer
@onready var player_prayer_label: Label = $UI/PlayerVitalsHUD/Margin/VBox/PrayerLabel
@onready var target_health_bar: ProgressBar = $UI/TargetHUD/Margin/VBox/Health
@onready var target_health_label: Label = $UI/TargetHUD/Margin/VBox/Header/HealthLabel
@onready var aggro_label: Label = $UI/TargetHUD/Margin/VBox/Header/AggroLabel
@onready var combat_state_label: Label = $UI/CombatHUD/Margin/VBox/CombatState
@onready var cooldown_label: Label = $UI/CombatHUD/Margin/VBox/CooldownLabel
@onready var cooldown_bar: ProgressBar = $UI/CombatHUD/Margin/VBox/Cooldown
@onready var sword_button: Button = $UI/CombatHUD/Margin/VBox/WeaponRow/Sword
@onready var dagger_button: Button = $UI/CombatHUD/Margin/VBox/WeaponRow/Dagger
@onready var bow_button: Button = $UI/CombatHUD/Margin/VBox/WeaponRow/Bow
@onready var disengage_button: Button = $UI/CombatHUD/Margin/VBox/WeaponRow/Disengage
@onready var click_marker: MeshInstance3D = $ClickMarker
@onready var target_click_pulse: MeshInstance3D = $TargetClickPulse
@onready var upper_level: FloorVisibility = $UpperLevel

var _local_player_id: String = ""
var _local_navigation_kind: String = "unknown"
var _local_navigation_id: String = ""
var _next_sequence: int = 1
var _next_action_sequence: int = 1
var _next_equipment_sequence: int = 1
var _next_pickup_sequence: int = 1
var _next_drop_sequence: int = 1
var _next_inventory_move_sequence: int = 1
var _next_consume_sequence: int = 1
var _next_gather_sequence: int = 1
var _next_bank_sequence: int = 1
var _last_action_sequence: int = 0
var _last_action_result: String = "none"
var _combat_target: String = ""
var _next_attack_pulse: int = 0
var _auto_approaching: bool = false
var _action_pulse: int = 0
var _last_consume_sequence: int = 0
var _next_consume_pulse: int = 0
var _queued_consume_item: String = ""
var _last_gather_sequence: int = 0
var _gathering_spot: String = ""
var _gathering_result: String = ""
var _next_gather_pulse: int = 0
var _gathering_views: Dictionary = {}
var _gathering_states: Dictionary = {}
var _bank_chest_views: Dictionary = {}
var _bank_chest_states: Dictionary = {}
var _bank_items: Array = []
var _bank_capacity: int = 100
var _bank_signature: String = ""
var _test_supply_signature: String = ""
var _active_bank_chest_id: String = ""
var _pending_bank_chest_id: String = ""
var _bank_withdraw_dialog: ConfirmationDialog
var _bank_withdraw_amount: SpinBox
var _pending_bank_withdraw_slot: int = -1
var _bank_item_menu: PopupMenu
var _bank_menu_slot: int = -1
var _bank_menu_quantity: int = 0
var _ammunition_pouch_items: Array = []
var _ammunition_pouch_capacity: int = 10
var _ammunition_pouch_unlocked: int = 2
var _ammunition_pouch_signature: String = ""
var _next_ammunition_sequence: int = 1
var _ammunition_pouch_window: Window
var _ammunition_pouch_grid: GridContainer
var _ammunition_pouch_summary: Label
var _target_views: Dictionary = {}
var _target_states: Dictionary = {}
var _drop_views: Dictionary = {}
var _drop_states: Dictionary = {}
var _player_health: int = 100
var _player_maximum_health: int = 100
var _player_prayer: int = 100
var _player_maximum_prayer: int = 100
var _active_prayers: Array[String] = []
var _active_prayer_set: String = "common_litany"
var _last_prayer_sequence: int = 0
var _next_prayer_sequence: int = 1
var _queued_prayer_id: String = ""
var _queued_prayer_active: bool = false
var _next_gem_cut_sequence: int = 1
var _selected_spell: String = ""
var _last_spell_sequence: int = 0
var _next_spell_sequence: int = 1
var _queued_spell: String = ""
var _spell_selection_pending: bool = false
var _player_defeated: bool = false
var _player_respawn_at_pulse: int = 0
var _player_dodge_rate: float = 0.1
var _equipped_weapon: String = "unarmed"
var _equipped_weapon_instance_id: String = ""
var _equipped_weapon_reinforcement: int = 0
var _last_equipment_sequence: int = 0
var _queued_item: String = ""
var _queued_instance_id: String = ""
var _queued_unequip: bool = false
var _queued_drop_after_unequip: String = ""
var _equipped_armour: Dictionary = {}
var _equipped_armour_instances: Dictionary = {}
var _equipped_armour_reinforcements: Dictionary = {}
var _equipped_spellbook: String = ""
var _active_spellbook: String = ""
var _combat_style: String = "crush"
var _last_combat_style_sequence: int = 0
var _next_combat_style_sequence: int = 1
var _queued_combat_style: String = ""
var _inventory_capacity: int = 8
var _inventory_items: Array = []
var _inventory_signature: String = ""
var _inventory_page: int = 0
var _skill_progress: Dictionary = {
	"attack": {"experience": 0, "level": 1, "current": 0, "next": 100},
	"strength": {"experience": 0, "level": 1, "current": 0, "next": 100},
	"defence": {"experience": 0, "level": 1, "current": 0, "next": 100},
	"range": {"experience": 0, "level": 1, "current": 0, "next": 100},
	"magic": {"experience": 0, "level": 1, "current": 0, "next": 100},
	"agility": {"experience": 0, "level": 1, "current": 0, "next": 100},
	"fishing": {"experience": 0, "level": 1, "current": 0, "next": 100},
	"cooking": {"experience": 0, "level": 1, "current": 0, "next": 100},
	"mining": {"experience": 0, "level": 1, "current": 0, "next": 100},
	"smithing": {"experience": 0, "level": 1, "current": 0, "next": 100},
	"woodcutting": {"experience": 0, "level": 1, "current": 0, "next": 100},
	"herblore": {"experience": 0, "level": 1, "current": 0, "next": 100},
	"fletching": {"experience": 0, "level": 1, "current": 0, "next": 100},
	"prayer": {"experience": 0, "level": 1, "current": 0, "next": 100},
}
var _skill_tiles: Dictionary = {}
var _combat_stats: Dictionary = {}
var _combat_stats_page: String = "offence"
var _player_views: Dictionary = {}
var _last_combat_event_id: int = 0
var _combat_events_initialized: bool = false
var _pending_click: Vector2
var _has_pending_click: bool = false
var _target_click_tween: Tween
var _world_context_menu: PopupMenu
var _context_menu_actions: Dictionary = {}
var _crafting_menu: PopupMenu
var _crafting_menu_actions: Dictionary = {}
var _selected_station_id: String = ""
var _selected_station_position: Vector3 = Vector3.ZERO
var _brewing_selection: Array[String] = []
var _next_brew_sequence: int = 1
var _brew_pending: bool = false
var _brew_remaining: int = 0
var _next_brew_pulse: int = 0
var _last_brew_sequence: int = 0
var _brewing_result_text: String = ""
var _brew_quantity_dialog: ConfirmationDialog
var _brew_quantity_input: SpinBox
var _selected_fletching_recipe: String = "arrow_shafts"
var _next_fletch_sequence: int = 1
var _last_fletch_sequence: int = 0
var _fletch_pending: bool = false
var _fletch_remaining: int = 0
var _next_fletch_pulse: int = 0
var _fletching_result_text: String = ""
var _fletch_quantity_dialog: ConfirmationDialog
var _fletch_quantity_input: SpinBox
var _hud_groups: Dictionary = {}
var _hud_handles: Dictionary = {}
var _hud_edit_mode: bool = false
var _target_hud_active: bool = false
var _combat_debug_hud_visible: bool = false
var _hud_drag_group: String = ""
var _hud_scale_group: String = ""
var _hud_pointer_start: Vector2 = Vector2.ZERO
var _hud_position_start: Vector2 = Vector2.ZERO
var _hud_scale_start: float = 1.0


func _ready() -> void:
	_world_context_menu = PopupMenu.new()
	_world_context_menu.name = "WorldContextMenu"
	add_child(_world_context_menu)
	_world_context_menu.id_pressed.connect(_on_world_context_action)
	_crafting_menu = PopupMenu.new()
	_crafting_menu.name = "CraftingMenu"
	add_child(_crafting_menu)
	_crafting_menu.id_pressed.connect(_on_crafting_recipe_selected)
	smithing_window.close_requested.connect(smithing_window.hide)
	cooking_window.close_requested.connect(cooking_window.hide)
	brewing_window.close_requested.connect(brewing_window.hide)
	for slot_index: int in range(brewing_slots.size()):
		brewing_slots[slot_index].pressed.connect(_remove_brewing_ingredient.bind(slot_index))
	brewing_clear_button.pressed.connect(_clear_brewing_selection)
	brewing_brew_button.pressed.connect(_request_brew)
	brewing_make_x_button.pressed.connect(_open_brew_quantity_dialog)
	brewing_make_all_button.pressed.connect(_request_brew.bind(0))
	_setup_brew_quantity_dialog()
	fletching_window.close_requested.connect(fletching_window.hide)
	fletching_shafts_button.toggle_mode = true
	fletching_arrows_button.toggle_mode = true
	fletching_shafts_button.pressed.connect(_select_fletching_recipe.bind("arrow_shafts"))
	fletching_arrows_button.pressed.connect(_select_fletching_recipe.bind("iron_arrows"))
	fletching_make_one_button.pressed.connect(_request_fletch)
	fletching_make_x_button.pressed.connect(_open_fletch_quantity_dialog)
	fletching_make_all_button.pressed.connect(_request_fletch.bind(0))
	_setup_fletch_quantity_dialog()
	_bank_item_menu = PopupMenu.new()
	_bank_item_menu.name = "BankItemMenu"
	add_child(_bank_item_menu)
	_bank_item_menu.id_pressed.connect(_on_bank_item_menu_action)
	bank_close_button.pressed.connect(_close_bank)
	_setup_bank_withdraw_dialog()
	_setup_ammunition_pouch_window()
	network.status_changed.connect(_on_status_changed)
	network.welcome_received.connect(_on_welcome_received)
	network.snapshot_received.connect(_on_snapshot_received)
	network.server_error.connect(_on_server_error)
	sword_button.pressed.connect(_request_equip.bind("sword"))
	dagger_button.pressed.connect(_request_equip.bind("dagger"))
	bow_button.pressed.connect(_request_equip.bind("bow"))
	disengage_button.pressed.connect(_request_disengage)
	combat_tab_button.pressed.connect(_show_combat_tab)
	inventory_tab_button.pressed.connect(_show_inventory_tab)
	combat_skills_tab_button.pressed.connect(_show_skills_tab)
	combat_equipment_tab_button.pressed.connect(_show_equipment_tab)
	combat_magic_tab_button.pressed.connect(_show_magic_tab)
	combat_styles_tab_button.pressed.connect(_show_styles_tab)
	inventory_combat_tab_button.pressed.connect(_show_combat_tab)
	inventory_active_tab_button.pressed.connect(_show_inventory_tab)
	inventory_skills_tab_button.pressed.connect(_show_skills_tab)
	inventory_equipment_tab_button.pressed.connect(_show_equipment_tab)
	inventory_magic_tab_button.pressed.connect(_show_magic_tab)
	inventory_prayer_tab_button.pressed.connect(_show_prayer_tab)
	inventory_styles_tab_button.pressed.connect(_show_styles_tab)
	inventory_previous_page_button.pressed.connect(_change_inventory_page.bind(-1))
	inventory_next_page_button.pressed.connect(_change_inventory_page.bind(1))
	skills_combat_tab_button.pressed.connect(_show_combat_tab)
	skills_inventory_tab_button.pressed.connect(_show_inventory_tab)
	skills_active_tab_button.pressed.connect(_show_skills_tab)
	skills_equipment_tab_button.pressed.connect(_show_equipment_tab)
	skills_magic_tab_button.pressed.connect(_show_magic_tab)
	skills_prayer_tab_button.pressed.connect(_show_prayer_tab)
	skills_styles_tab_button.pressed.connect(_show_styles_tab)
	equipment_combat_tab_button.pressed.connect(_show_combat_tab)
	equipment_inventory_tab_button.pressed.connect(_show_inventory_tab)
	equipment_skills_tab_button.pressed.connect(_show_skills_tab)
	equipment_active_tab_button.pressed.connect(_show_equipment_tab)
	equipment_magic_tab_button.pressed.connect(_show_magic_tab)
	equipment_prayer_tab_button.pressed.connect(_show_prayer_tab)
	equipment_styles_tab_button.pressed.connect(_show_styles_tab)
	detailed_stats_button.pressed.connect(_show_detailed_combat_stats)
	combat_stats_close_button.pressed.connect(_hide_detailed_combat_stats)
	combat_stats_offence_button.pressed.connect(_show_combat_stats_page.bind("offence"))
	combat_stats_defence_button.pressed.connect(_show_combat_stats_page.bind("defence"))
	combat_stats_effects_button.pressed.connect(_show_combat_stats_page.bind("effects"))
	magic_combat_tab_button.pressed.connect(_show_combat_tab)
	magic_inventory_tab_button.pressed.connect(_show_inventory_tab)
	magic_skills_tab_button.pressed.connect(_show_skills_tab)
	magic_equipment_tab_button.pressed.connect(_show_equipment_tab)
	magic_active_tab_button.pressed.connect(_show_magic_tab)
	magic_prayer_tab_button.pressed.connect(_show_prayer_tab)
	magic_styles_tab_button.pressed.connect(_show_styles_tab)
	prayer_inventory_tab_button.pressed.connect(_show_inventory_tab)
	prayer_skills_tab_button.pressed.connect(_show_skills_tab)
	prayer_equipment_tab_button.pressed.connect(_show_equipment_tab)
	prayer_magic_tab_button.pressed.connect(_show_magic_tab)
	prayer_active_tab_button.pressed.connect(_show_prayer_tab)
	prayer_styles_tab_button.pressed.connect(_show_styles_tab)
	styles_combat_tab_button.pressed.connect(_show_combat_tab)
	styles_inventory_tab_button.pressed.connect(_show_inventory_tab)
	styles_skills_tab_button.pressed.connect(_show_skills_tab)
	styles_equipment_tab_button.pressed.connect(_show_equipment_tab)
	styles_magic_tab_button.pressed.connect(_show_magic_tab)
	styles_prayer_tab_button.pressed.connect(_show_prayer_tab)
	styles_active_tab_button.pressed.connect(_show_styles_tab)
	_load_client_settings()
	inventory_settings_button.pressed.connect(_show_settings_tab)
	settings_inventory_button.pressed.connect(_show_inventory_tab)
	settings_diagnostics_toggle.toggled.connect(_on_diagnostics_toggled)
	settings_music_slider.value_changed.connect(_on_music_volume_changed)
	settings_edit_hud_button.pressed.connect(_begin_hud_editing_from_settings)
	settings_reset_hud_button.pressed.connect(_reset_hud_layout_from_settings)
	for settings_button: Button in [inventory_settings_button, $UI/SettingsHUD/Margin/VBox/UtilityTabs/Margin/Row/Settings]:
		settings_button.icon = TAB_SETTINGS_ICON
		settings_button.expand_icon = true
		settings_button.add_theme_constant_override("icon_max_width", 42)
	_configure_player_menu()
	_update_combat_hud()
	_update_inventory_hud()
	_populate_placeholder_skills()
	_populate_placeholder_spells()
	_populate_placeholder_prayers()
	_update_combat_styles_hud()
	_update_equipment_hud()
	_setup_hud_layout.call_deferred()


func _unhandled_input(event: InputEvent) -> void:
	if event is InputEventKey:
		var key_event := event as InputEventKey
		if key_event.pressed and not key_event.echo:
			if key_event.keycode == KEY_F10:
				if key_event.shift_pressed:
					_reset_hud_layout()
				else:
					_set_hud_edit_mode(not _hud_edit_mode)
				get_viewport().set_input_as_handled()
				return
			if key_event.keycode == KEY_F9:
				_toggle_combat_debug_hud()
				get_viewport().set_input_as_handled()
				return
			match key_event.keycode:
				KEY_ESCAPE:
					if _hud_edit_mode:
						_set_hud_edit_mode(false)
						get_viewport().set_input_as_handled()
						return
					if fletching_window.visible:
						fletching_window.hide()
						get_viewport().set_input_as_handled()
						return
					if brewing_window.visible:
						brewing_window.hide()
						get_viewport().set_input_as_handled()
						return
					if smithing_window.visible:
						smithing_window.hide()
						get_viewport().set_input_as_handled()
						return
					if settings_hud.visible:
						_show_inventory_tab()
						get_viewport().set_input_as_handled()
						return
					if _ammunition_pouch_window != null and _ammunition_pouch_window.visible:
						_close_ammunition_pouch()
						get_viewport().set_input_as_handled()
						return
					if combat_stats_hud.visible:
						_hide_detailed_combat_stats()
						get_viewport().set_input_as_handled()
						return
					_request_disengage()
					get_viewport().set_input_as_handled()
				KEY_1:
					_request_equip("sword")
					get_viewport().set_input_as_handled()
				KEY_2:
					_request_equip("dagger")
					get_viewport().set_input_as_handled()
				KEY_3:
					_request_equip("bow")
					get_viewport().set_input_as_handled()
				KEY_4:
					_request_equip("mace")
					get_viewport().set_input_as_handled()
				KEY_5:
					_request_equip("battle_axe")
					get_viewport().set_input_as_handled()
				KEY_6:
					_request_equip("staff")
					get_viewport().set_input_as_handled()
		return
	if _hud_edit_mode:
		get_viewport().set_input_as_handled()
		return
	if not event is InputEventMouseButton:
		return
	var mouse_event := event as InputEventMouseButton
	if not mouse_event.pressed:
		return
	if _player_defeated:
		return
	if mouse_event.button_index == MOUSE_BUTTON_RIGHT:
		_open_world_context_menu(mouse_event.position)
		get_viewport().set_input_as_handled()
		return
	if mouse_event.button_index != MOUSE_BUTTON_LEFT:
		return

	_pending_click = mouse_event.position
	_has_pending_click = true


func _physics_process(_delta: float) -> void:
	if not _has_pending_click:
		return
	_has_pending_click = false

	var ray_origin: Vector3 = camera.project_ray_origin(_pending_click)
	var ray_end: Vector3 = ray_origin + camera.project_ray_normal(_pending_click) * RAY_LENGTH
	var target_query := PhysicsRayQueryParameters3D.create(ray_origin, ray_end)
	target_query.collision_mask = TARGET_COLLISION_MASK
	var target_result: Dictionary = get_world_3d().direct_space_state.intersect_ray(target_query)
	if not target_result.is_empty():
		var target_body := target_result.get("collider") as CollisionObject3D
		if target_body != null and target_body.has_meta(&"combat_target_id"):
			var target_id: String = str(target_body.get_meta(&"combat_target_id"))
			_request_engage_target(target_id, target_body.global_position)
			return

	var bank_query := PhysicsRayQueryParameters3D.create(ray_origin, ray_end)
	bank_query.collision_mask = BANK_COLLISION_MASK
	var bank_result: Dictionary = get_world_3d().direct_space_state.intersect_ray(bank_query)
	if not bank_result.is_empty():
		var bank_body := bank_result.get("collider") as CollisionObject3D
		if bank_body != null and bank_body.has_meta(&"bank_chest_id"):
			var bank_chest_id: String = str(bank_body.get_meta(&"bank_chest_id"))
			_request_open_bank(bank_chest_id, bank_body.global_position, float(bank_body.get_meta(&"bank_range", 2.5)))
			return

	var gathering_query := PhysicsRayQueryParameters3D.create(ray_origin, ray_end)
	gathering_query.collision_mask = GATHERING_COLLISION_MASK
	var gathering_result: Dictionary = get_world_3d().direct_space_state.intersect_ray(gathering_query)
	if not gathering_result.is_empty():
		var gathering_body := gathering_result.get("collider") as CollisionObject3D
		if gathering_body != null and gathering_body.has_meta(&"gathering_spot_id"):
			var spot_id: String = str(gathering_body.get_meta(&"gathering_spot_id"))
			var spot_state: Dictionary = _gathering_states.get(spot_id, {}) as Dictionary
			var station: String = str(spot_state.get("station", ""))
			if station == "fishing":
				_request_gather(spot_id, gathering_body.global_position, str(spot_state.get("recipe_id", "")))
			elif CRAFTING_RECIPES.has(station):
				_open_crafting_menu(spot_id, gathering_body.global_position, station, _pending_click)
			else:
				_request_gather(spot_id, gathering_body.global_position)
			return

	var loot_result: Dictionary = _top_loot_hit(ray_origin, ray_end)
	if not loot_result.is_empty():
		var loot_body := loot_result.get("collider") as CollisionObject3D
		if loot_body != null and loot_body.has_meta(&"drop_id"):
			var drop_id: String = str(loot_body.get_meta(&"drop_id"))
			_request_pickup_drop(drop_id, loot_body.global_position)
			return

	var query := PhysicsRayQueryParameters3D.create(ray_origin, ray_end)
	# Movement clicks can intentionally target a different navigation layer.
	# The authoritative pathfinder will route through the ramp when, for
	# example, an upper-level player clicks visible ground below.
	query.collision_mask = ALL_WORLD_COLLISION_MASK
	var result: Dictionary = get_world_3d().direct_space_state.intersect_ray(query)
	if result.is_empty():
		return

	var clicked_body := result.get("collider") as CollisionObject3D
	var destination: Vector3 = result["position"]
	if bank_hud.visible or not _pending_bank_chest_id.is_empty() or not _active_bank_chest_id.is_empty():
		_close_bank()
	var sequence: int = _next_sequence
	if network.send_move(sequence, destination):
		_next_sequence += 1
		if not _combat_target.is_empty() or _last_action_sequence < _next_action_sequence - 1:
			_request_disengage()
		var local_view: PlayerView = _player_views.get(_local_player_id)
		if local_view != null and _can_predict_directly(local_view, destination, clicked_body):
			if clicked_body.name == &"GroundBody":
				local_view.predict_path(ground.surface_path(local_view.global_position, destination))
			else:
				local_view.predict_move(destination)
		click_marker.position = destination + Vector3.UP * 0.03
		click_marker.visible = true


func _request_engage_target(target_id: String, target_position: Vector3) -> void:
	if target_id.is_empty() or _player_defeated:
		return
	if network.send_engage_target(_next_action_sequence, target_id):
		_next_action_sequence += 1
		var local_view: PlayerView = _player_views.get(_local_player_id)
		if local_view != null:
			local_view.stop_prediction()
		click_marker.visible = false
		_play_target_click_feedback(target_position)
		_update_combat_hud()


func _request_pickup_drop(drop_id: String, drop_position: Vector3) -> void:
	if drop_id.is_empty() or _player_defeated:
		return
	if network.send_pickup_item(_next_pickup_sequence, drop_id):
		_next_pickup_sequence += 1
		var local_view: PlayerView = _player_views.get(_local_player_id)
		if local_view != null:
			local_view.stop_prediction()
		if not _combat_target.is_empty() or _last_action_sequence < _next_action_sequence - 1:
			_request_disengage()
		click_marker.visible = false
		_play_target_click_feedback(drop_position)


func _request_gather(spot_id: String, spot_position: Vector3, recipe_id: String = "", instance_id: String = "") -> void:
	if spot_id.is_empty() or _player_defeated:
		return
	if network.send_gather(_next_gather_sequence, spot_id, recipe_id, instance_id):
		_next_gather_sequence += 1
		var local_view: PlayerView = _player_views.get(_local_player_id)
		if local_view != null:
			local_view.stop_prediction()
		click_marker.visible = false
		_play_target_click_feedback(spot_position)


func _request_open_bank(chest_id: String, chest_position: Vector3, interaction_range: float) -> void:
	if chest_id.is_empty() or _player_defeated:
		return
	_pending_bank_chest_id = chest_id
	var local_view: PlayerView = _player_views.get(_local_player_id)
	if local_view != null and local_view.global_position.distance_to(chest_position) <= interaction_range:
		_open_bank(chest_id)
		return
	if network.send_move(_next_sequence, chest_position):
		_next_sequence += 1
		status_label.text = "Walking to the bank chest…"
		if not _combat_target.is_empty() or _last_action_sequence < _next_action_sequence - 1:
			_request_disengage()


func _open_bank(chest_id: String) -> void:
	_active_bank_chest_id = chest_id
	_pending_bank_chest_id = ""
	if _is_test_supply_chest(chest_id):
		_show_inventory_tab()
		var chest: Dictionary = _bank_chest_states.get(chest_id, {}) as Dictionary
		bank_title.text = str(chest.get("name", "Test Supply Chest"))
		bank_hint.text = "Click a supply to take its stack • supplies return when the server restarts"
		bank_hud.visible = true
		_update_test_supply_hud(true)
		_update_inventory_hud()
		status_label.text = "Test supply chest opened"
		return
	_show_inventory_tab()
	bank_title.text = "Personal Bank"
	bank_hint.text = "Click an inventory item to deposit one • click a bank item to withdraw one"
	bank_hud.visible = true
	_update_bank_hud()
	_update_inventory_hud()
	status_label.text = "Bank opened"


func _is_test_supply_chest(chest_id: String) -> bool:
	return chest_id.begins_with(TEST_SUPPLY_CHEST_PREFIX) and chest_id.ends_with("_chest")


func _close_bank() -> void:
	_active_bank_chest_id = ""
	_pending_bank_chest_id = ""
	bank_hud.visible = false
	_pending_bank_withdraw_slot = -1
	_bank_menu_slot = -1
	_bank_menu_quantity = 0
	if _bank_item_menu != null:
		_bank_item_menu.hide()
	if _bank_withdraw_dialog != null:
		_bank_withdraw_dialog.hide()
	_update_inventory_hud()


func _update_bank_proximity(player_position: Vector3) -> void:
	if not _combat_target.is_empty():
		if bank_hud.visible or not _pending_bank_chest_id.is_empty() or not _active_bank_chest_id.is_empty():
			_close_bank()
		return
	var chest_id: String = _active_bank_chest_id if not _active_bank_chest_id.is_empty() else _pending_bank_chest_id
	if chest_id.is_empty():
		return
	var chest: Dictionary = _bank_chest_states.get(chest_id, {}) as Dictionary
	var chest_position_value: Variant = chest.get("position", {})
	if typeof(chest_position_value) != TYPE_DICTIONARY:
		_close_bank()
		return
	var coordinates: Dictionary = chest_position_value
	var chest_position := Vector3(
		float(coordinates.get("x", 0.0)),
		float(coordinates.get("y", 0.0)),
		float(coordinates.get("z", 0.0)),
	)
	var interaction_range: float = float(chest.get("range", 2.5))
	if player_position.distance_to(chest_position) <= interaction_range:
		if _active_bank_chest_id.is_empty():
			_open_bank(chest_id)
	elif not _active_bank_chest_id.is_empty():
		_close_bank()


func _test_supply_items() -> Array[Dictionary]:
	var supplies: Array[Dictionary] = []
	var chest: Dictionary = _bank_chest_states.get(_active_bank_chest_id, {}) as Dictionary
	var chest_position_data: Variant = chest.get("position", {})
	if typeof(chest_position_data) != TYPE_DICTIONARY:
		return supplies
	var chest_coordinates := chest_position_data as Dictionary
	var chest_position := Vector3(
		float(chest_coordinates.get("x", 0.0)),
		float(chest_coordinates.get("y", 0.0)),
		float(chest_coordinates.get("z", 0.0)),
	)
	for drop_id: String in _drop_states.keys():
		var state: Dictionary = _drop_states.get(drop_id, {}) as Dictionary
		var position_data: Variant = state.get("position", {})
		if typeof(position_data) != TYPE_DICTIONARY:
			continue
		var coordinates := position_data as Dictionary
		var position := Vector3(
			float(coordinates.get("x", 0.0)),
			float(coordinates.get("y", 0.0)),
			float(coordinates.get("z", 0.0)),
		)
		if position.distance_to(chest_position) > 0.1:
			continue
		supplies.append({
			"drop_id": drop_id,
			"item_id": str(state.get("item_id", "item")),
			"quantity": int(state.get("quantity", 1)),
		})
	supplies.sort_custom(func(left: Dictionary, right: Dictionary) -> bool:
		return str(left.get("item_id", "")) < str(right.get("item_id", ""))
	)
	return supplies


func _update_test_supply_hud(force: bool = false) -> void:
	if not _is_test_supply_chest(_active_bank_chest_id) or not bank_hud.visible:
		return
	var supplies := _test_supply_items()
	var signature := JSON.stringify(supplies)
	if not force and signature == _test_supply_signature:
		return
	_test_supply_signature = signature
	for child: Node in bank_grid.get_children():
		bank_grid.remove_child(child)
		child.queue_free()
	bank_summary.text = "%d test suppl%s remaining" % [
		supplies.size(),
		"y" if supplies.size() == 1 else "ies",
	]
	for supply: Dictionary in supplies:
		var drop_id: String = str(supply.get("drop_id", ""))
		var item_id: String = str(supply.get("item_id", ""))
		var quantity: int = int(supply.get("quantity", 1))
		var slot_button: Variant = INVENTORY_SLOT_SCRIPT.new()
		slot_button.custom_minimum_size = Vector2(112, 86)
		slot_button.size_flags_horizontal = Control.SIZE_EXPAND_FILL
		slot_button.focus_mode = Control.FOCUS_NONE
		slot_button.text = ""
		slot_button.configure(0, true, true)
		slot_button.icon = ITEM_ICONS.get(item_id, EMPTY_SLOT_ICON) as Texture2D
		slot_button.expand_icon = true
		slot_button.set_stack_quantity(quantity)
		slot_button.tooltip_text = "%s ×%d • click to take the stack" % [item_id.capitalize(), quantity]
		slot_button.pressed.connect(_take_test_supply.bind(drop_id))
		bank_grid.add_child(slot_button)


func _take_test_supply(drop_id: String) -> void:
	var drop_view := _drop_views.get(drop_id) as LootDrop
	if drop_view == null or not is_instance_valid(drop_view):
		status_label.text = "That test supply is no longer available"
		return
	_request_pickup_drop(drop_id, drop_view.global_position)


func _open_crafting_menu(spot_id: String, spot_position: Vector3, station: String, screen_position: Vector2) -> void:
	_selected_station_id = spot_id
	_selected_station_position = spot_position
	if station == "forging":
		_open_smithing_window()
		return
	if station == "cooking":
		_open_cooking_window()
		return
	_crafting_menu.clear()
	_crafting_menu_actions.clear()
	var recipes: Array = CRAFTING_RECIPES.get(station, []) as Array
	var altar_recipe: String = str((_gathering_states.get(spot_id, {}) as Dictionary).get("recipe_id", ""))
	for recipe_index: int in recipes.size():
		var recipe: Dictionary = recipes[recipe_index]
		var required_altar_recipe: String = str(recipe.get("altar_recipe", ""))
		if not required_altar_recipe.is_empty() and required_altar_recipe != altar_recipe:
			continue
		_crafting_menu.add_item(
			"%s  •  %s  •  Lv %d" % [recipe["name"], recipe["inputs"], recipe["level"]],
			recipe_index,
		)
		_crafting_menu_actions[recipe_index] = {"recipe_id": str(recipe["id"]), "instance_id": ""}
	_crafting_menu.position = Vector2i(screen_position)
	_crafting_menu.popup()


func _open_cooking_window() -> void:
	for child: Node in cooking_grid.get_children():
		cooking_grid.remove_child(child)
		child.queue_free()
	var cooking_level: int = int((_skill_progress.get("cooking", {}) as Dictionary).get("level", 1))
	var recipes: Array = CRAFTING_RECIPES.get("cooking", []) as Array
	for recipe_data: Variant in recipes:
		var recipe := recipe_data as Dictionary
		var recipe_id: String = str(recipe.get("id", ""))
		var input_id: String = str(recipe.get("input_id", ""))
		var required_level: int = int(recipe.get("level", 1))
		var available: int = _inventory_quantity(input_id)
		var recipe_button := Button.new()
		recipe_button.custom_minimum_size = Vector2(260, 116)
		recipe_button.size_flags_horizontal = Control.SIZE_EXPAND_FILL
		recipe_button.focus_mode = Control.FOCUS_NONE
		recipe_button.alignment = HORIZONTAL_ALIGNMENT_LEFT
		recipe_button.text = "%s\n%s  •  Cooking %d\nAvailable: %d" % [
			recipe.get("name", recipe_id.capitalize()), recipe.get("inputs", ""), required_level, available,
		]
		recipe_button.icon = ITEM_ICONS.get(recipe_id, EMPTY_SLOT_ICON) as Texture2D
		recipe_button.expand_icon = true
		recipe_button.add_theme_constant_override("icon_max_width", 62)
		recipe_button.disabled = cooking_level < required_level or available < 1
		if cooking_level < required_level:
			recipe_button.tooltip_text = "Requires Cooking level %d" % required_level
		elif available < 1:
			recipe_button.tooltip_text = "Requires %s" % recipe.get("inputs", "raw food")
		else:
			recipe_button.tooltip_text = "Cook %s until you run out" % recipe.get("name", recipe_id.capitalize())
		recipe_button.pressed.connect(_select_cooking_recipe.bind(recipe_id))
		cooking_grid.add_child(recipe_button)
	cooking_window.popup_centered()


func _select_cooking_recipe(recipe_id: String) -> void:
	cooking_window.hide()
	_request_gather(_selected_station_id, _selected_station_position, recipe_id)


func _open_smithing_window() -> void:
	for child: Node in smithing_tabs.get_children():
		smithing_tabs.remove_child(child)
		child.queue_free()
	var smithing_progress: Dictionary = _skill_progress.get("smithing", {}) as Dictionary
	var smithing_level: int = int(smithing_progress.get("level", 1))
	var recipes: Array = CRAFTING_RECIPES.get("forging", []) as Array
	for tier: String in ["bronze", "iron", "mithril", "gold"]:
		var tier_scroll := ScrollContainer.new()
		tier_scroll.name = tier.capitalize()
		tier_scroll.horizontal_scroll_mode = ScrollContainer.SCROLL_MODE_DISABLED
		var tier_grid := GridContainer.new()
		tier_grid.columns = 2
		tier_grid.size_flags_horizontal = Control.SIZE_EXPAND_FILL
		tier_grid.add_theme_constant_override("h_separation", 10)
		tier_grid.add_theme_constant_override("v_separation", 10)
		tier_scroll.add_child(tier_grid)
		smithing_tabs.add_child(tier_scroll)
		for recipe_data: Variant in recipes:
			var recipe := recipe_data as Dictionary
			if str(recipe.get("tier", "")) != tier:
				continue
			var recipe_id: String = str(recipe.get("id", ""))
			var required_level: int = int(recipe.get("level", 1))
			var recipe_button := Button.new()
			recipe_button.custom_minimum_size = Vector2(330, 104)
			recipe_button.size_flags_horizontal = Control.SIZE_EXPAND_FILL
			recipe_button.focus_mode = Control.FOCUS_NONE
			recipe_button.text = "%s\n%s  •  Smithing %d" % [
				recipe.get("name", recipe_id.capitalize()), recipe.get("inputs", ""), required_level,
			]
			recipe_button.alignment = HORIZONTAL_ALIGNMENT_LEFT
			var icon_id: String = "sword" if recipe_id == "iron_sword" else recipe_id
			recipe_button.icon = ITEM_ICONS.get(icon_id, EMPTY_SLOT_ICON) as Texture2D
			recipe_button.expand_icon = true
			recipe_button.add_theme_constant_override("icon_max_width", 62)
			recipe_button.disabled = smithing_level < required_level
			recipe_button.tooltip_text = (
				"Requires Smithing level %d" % required_level
				if recipe_button.disabled
				else "Forge %s" % recipe.get("name", recipe_id.capitalize())
			)
			recipe_button.pressed.connect(_select_smithing_recipe.bind(recipe_id, ""))
			tier_grid.add_child(recipe_button)
	_add_reinforcement_smithing_tab()
	smithing_window.popup_centered()


func _add_reinforcement_smithing_tab() -> void:
	var reinforce_scroll := ScrollContainer.new()
	reinforce_scroll.name = "Reinforce"
	reinforce_scroll.horizontal_scroll_mode = ScrollContainer.SCROLL_MODE_DISABLED
	var reinforce_list := VBoxContainer.new()
	reinforce_list.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	reinforce_list.add_theme_constant_override("separation", 10)
	reinforce_scroll.add_child(reinforce_list)
	smithing_tabs.add_child(reinforce_scroll)
	var candidates: int = 0
	var smithing_level: int = int((_skill_progress.get("smithing", {}) as Dictionary).get("level", 1))
	for payload: Variant in _inventory_items:
		if typeof(payload) != TYPE_DICTIONARY:
			continue
		var item := payload as Dictionary
		var item_id: String = str(item.get("item_id", ""))
		var instance_id: String = str(item.get("instance_id", ""))
		var reinforcement: int = int(item.get("reinforcement", 0))
		var reinforcement_data := _reinforcement_data(item_id)
		if instance_id.is_empty() or reinforcement >= MAXIMUM_REINFORCEMENT or reinforcement_data.is_empty():
			continue
		var bar_cost: int = int(reinforcement_data.get("cost", 0))
		var required_level: int = int(reinforcement_data.get("level", 1))
		var recipe_button := Button.new()
		recipe_button.custom_minimum_size = Vector2(0, 90)
		recipe_button.text = "Reinforce %s to +%d\n%d %s  •  Smithing %d  •  inventory slot %d" % [
			item_id.capitalize(), reinforcement + 1, bar_cost, reinforcement_data.get("bar_name", "Bars"),
			required_level, int(item.get("slot", 0)) + 1,
		]
		recipe_button.icon = ITEM_ICONS.get(item_id, EMPTY_SLOT_ICON) as Texture2D
		recipe_button.expand_icon = true
		recipe_button.add_theme_constant_override("icon_max_width", 56)
		recipe_button.disabled = smithing_level < required_level
		recipe_button.pressed.connect(_select_smithing_recipe.bind("reinforce_%s" % item_id, instance_id))
		reinforce_list.add_child(recipe_button)
		candidates += 1
	if candidates == 0:
		var empty_label := Label.new()
		empty_label.text = "No eligible equipment to reinforce."
		empty_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
		empty_label.custom_minimum_size = Vector2(0, 80)
		reinforce_list.add_child(empty_label)


func _reinforcement_data(item_id: String) -> Dictionary:
	for tier_data_value: Variant in REINFORCEMENT_TIERS.values():
		var tier_data := tier_data_value as Dictionary
		if not item_id in (tier_data.get("items", []) as Array):
			continue
		var result := tier_data.duplicate()
		var slot: String = str(ARMOUR_SLOTS.get(item_id, ""))
		var cost := 3 if EQUIPMENT_ICONS.has(item_id) else 2
		if slot == "body":
			cost = 4
		elif slot in ["legs", "left_hand"]:
			cost = 3
		result["cost"] = cost
		return result
	return {}


func _select_smithing_recipe(recipe_id: String, instance_id: String) -> void:
	smithing_window.hide()
	_request_gather(_selected_station_id, _selected_station_position, recipe_id, instance_id)


func _open_brewing_window(starting_item_id: String = "") -> void:
	if not brewing_window.visible:
		_brewing_selection.clear()
	if not starting_item_id.is_empty():
		_add_brewing_ingredient(starting_item_id)
	_refresh_brewing_window()
	brewing_window.popup_centered()


func _add_brewing_ingredient(item_id: String) -> void:
	if _brew_pending or _brewing_selection.size() >= 3 or not BREWING_INGREDIENTS.has(item_id):
		return
	var selected_count: int = _brewing_selection.count(item_id)
	if selected_count >= _inventory_quantity(item_id):
		return
	_brewing_selection.append(item_id)
	_refresh_brewing_window()


func _remove_brewing_ingredient(slot_index: int) -> void:
	if _brew_pending or slot_index < 0 or slot_index >= _brewing_selection.size():
		return
	_brewing_selection.remove_at(slot_index)
	_refresh_brewing_window()


func _clear_brewing_selection() -> void:
	if _brew_pending:
		return
	_brewing_selection.clear()
	_refresh_brewing_window()


func _refresh_brewing_window() -> void:
	for slot_index: int in range(brewing_slots.size()):
		var slot_button := brewing_slots[slot_index]
		if slot_index < _brewing_selection.size():
			var item_id: String = _brewing_selection[slot_index]
			slot_button.text = str(BREWING_INGREDIENTS[item_id].get("name", item_id.capitalize()))
			slot_button.icon = ITEM_ICONS.get(item_id, EMPTY_SLOT_ICON) as Texture2D
			slot_button.expand_icon = true
			slot_button.tooltip_text = "Click to remove"
		else:
			slot_button.text = "+ Optional" if slot_index == 2 else "+ Ingredient"
			slot_button.icon = null
			slot_button.tooltip_text = ""
	for child: Node in brewing_grid.get_children():
		brewing_grid.remove_child(child)
		child.queue_free()
	for item_id: String in BREWING_INGREDIENTS.keys():
		var quantity: int = _inventory_quantity(item_id)
		if quantity <= 0:
			continue
		var ingredient: Dictionary = BREWING_INGREDIENTS[item_id] as Dictionary
		var ingredient_button := Button.new()
		ingredient_button.custom_minimum_size = Vector2(158, 96)
		ingredient_button.size_flags_horizontal = Control.SIZE_EXPAND_FILL
		ingredient_button.text = "%s\n%s ×%d" % [ingredient.get("name", item_id.capitalize()), ingredient.get("role", "Ingredient"), quantity]
		ingredient_button.icon = ITEM_ICONS.get(item_id, EMPTY_SLOT_ICON) as Texture2D
		ingredient_button.expand_icon = true
		ingredient_button.add_theme_constant_override("icon_max_width", 48)
		ingredient_button.disabled = _brew_pending or _brewing_selection.size() >= 3 or _brewing_selection.count(item_id) >= quantity
		ingredient_button.pressed.connect(_add_brewing_ingredient.bind(item_id))
		brewing_grid.add_child(ingredient_button)
	var valid_roles := _brewing_selection.size() >= 2 and _brewing_selection.size() <= 3
	valid_roles = valid_roles and _brewing_selection.count("vial_of_water") == 1
	var has_herb: bool = _brewing_selection.has("sunleaf") or _brewing_selection.has("moonleaf")
	valid_roles = valid_roles and has_herb
	if valid_roles:
		var key_items := _brewing_selection.duplicate()
		key_items.sort()
		var mixture_key := "+".join(key_items)
		brewing_result.text = "Known result: %s" % KNOWN_BREWS[mixture_key] if KNOWN_BREWS.has(mixture_key) else "Unknown mixture: experimentation may create a Murky Brew"
	else:
		brewing_result.text = "Requires exactly one Vial of Water and at least one herb"
	if _brew_pending:
		brewing_result.text = "Brewing one potion every 600 ms • %d remaining" % _brew_remaining
	brewing_brew_button.disabled = not valid_roles or _brew_pending
	brewing_brew_button.text = "Brewing…" if _brew_pending else "Make 1"
	brewing_make_x_button.disabled = not valid_roles or _brew_pending
	brewing_make_all_button.disabled = not valid_roles or _brew_pending
	brewing_clear_button.disabled = _brew_pending or _brewing_selection.is_empty()


func _maximum_selected_brews() -> int:
	if _brewing_selection.is_empty():
		return 0
	var costs: Dictionary = {}
	for item_id: String in _brewing_selection:
		costs[item_id] = int(costs.get(item_id, 0)) + 1
	var maximum: int = 2_147_483_647
	for item_id: String in costs.keys():
		maximum = mini(maximum, floori(float(_inventory_quantity(item_id)) / float(int(costs[item_id]))))
	return 0 if maximum == 2_147_483_647 else maximum


func _setup_brew_quantity_dialog() -> void:
	_brew_quantity_dialog = ConfirmationDialog.new()
	_brew_quantity_dialog.title = "Make X potions"
	_brew_quantity_dialog.dialog_text = "Choose how many potions to brew:"
	_brew_quantity_dialog.min_size = Vector2i(360, 180)
	_brew_quantity_dialog.get_ok_button().text = "Start brewing"
	_brew_quantity_input = SpinBox.new()
	_brew_quantity_input.min_value = 1
	_brew_quantity_input.step = 1
	_brew_quantity_input.custom_minimum_size = Vector2(280, 44)
	_brew_quantity_input.position = Vector2(40, 74)
	_brew_quantity_dialog.add_child(_brew_quantity_input)
	_brew_quantity_dialog.confirmed.connect(_confirm_brew_quantity)
	add_child(_brew_quantity_dialog)


func _open_brew_quantity_dialog() -> void:
	var maximum := _maximum_selected_brews()
	if maximum < 1 or _brew_pending:
		return
	_brew_quantity_input.max_value = maximum
	_brew_quantity_input.value = mini(maximum, 5)
	_brew_quantity_dialog.popup_centered()
	_brew_quantity_input.grab_focus()
	_brew_quantity_input.get_line_edit().select_all()


func _confirm_brew_quantity() -> void:
	_request_brew(int(_brew_quantity_input.value))


func _request_brew(quantity: int = 1) -> void:
	if _brew_pending or _brewing_selection.size() < 2 or _brewing_selection.size() > 3:
		return
	if network.send_brew(_next_brew_sequence, _brewing_selection, quantity):
		_next_brew_sequence += 1
		_brew_pending = true
		_brew_remaining = _maximum_selected_brews() if quantity == 0 else quantity
		_refresh_brewing_window()


func _open_fletching_window(starting_item_id: String = "") -> void:
	if starting_item_id in ["arrow_shafts", "feathers", "iron_arrow_tips"]:
		_selected_fletching_recipe = "iron_arrows"
	elif starting_item_id in ["knife", "normal_logs"]:
		_selected_fletching_recipe = "arrow_shafts"
	_refresh_fletching_window()
	fletching_window.popup_centered()


func _select_fletching_recipe(recipe_id: String) -> void:
	if _fletch_pending or not FLETCHING_RECIPES.has(recipe_id):
		return
	_selected_fletching_recipe = recipe_id
	_refresh_fletching_window()


func _maximum_fletch_batches() -> int:
	var recipe: Dictionary = FLETCHING_RECIPES.get(_selected_fletching_recipe, {}) as Dictionary
	if recipe.is_empty():
		return 0
	var tool_id: String = str(recipe.get("tool", ""))
	if not tool_id.is_empty() and _inventory_quantity(tool_id) < 1:
		return 0
	var maximum: int = 2_147_483_647
	var costs: Dictionary = recipe.get("costs", {}) as Dictionary
	for item_id: String in costs:
		var required: int = int(costs[item_id])
		maximum = mini(maximum, floori(float(_inventory_quantity(item_id)) / float(required)))
	return 0 if maximum == 2_147_483_647 else maximum


func _refresh_fletching_window() -> void:
	var maximum := _maximum_fletch_batches()
	var recipe: Dictionary = FLETCHING_RECIPES.get(_selected_fletching_recipe, {}) as Dictionary
	var recipe_name: String = str(recipe.get("name", "Select a recipe"))
	var experience: int = int(recipe.get("experience", 0))
	fletching_shafts_button.button_pressed = _selected_fletching_recipe == "arrow_shafts"
	fletching_arrows_button.button_pressed = _selected_fletching_recipe == "iron_arrows"
	fletching_shafts_button.disabled = _fletch_pending
	fletching_arrows_button.disabled = _fletch_pending
	if _fletch_pending:
		fletching_selection.text = "Making %s every 600 ms • %d batch%s remaining" % [
			recipe_name, _fletch_remaining, "" if _fletch_remaining == 1 else "es",
		]
	else:
		fletching_selection.text = "%s • %d XP per batch • materials for %d batch%s" % [
			recipe_name, experience, maximum, "" if maximum == 1 else "es",
		]
	var unavailable := maximum < 1 or _fletch_pending
	fletching_make_one_button.disabled = unavailable
	fletching_make_x_button.disabled = unavailable
	fletching_make_all_button.disabled = unavailable
	fletching_make_one_button.text = "Fletching…" if _fletch_pending else "Make 1"


func _setup_fletch_quantity_dialog() -> void:
	_fletch_quantity_dialog = ConfirmationDialog.new()
	_fletch_quantity_dialog.title = "Fletch X batches"
	_fletch_quantity_dialog.dialog_text = "Choose how many batches of 20 to make:"
	_fletch_quantity_dialog.min_size = Vector2i(380, 180)
	_fletch_quantity_dialog.get_ok_button().text = "Start fletching"
	_fletch_quantity_input = SpinBox.new()
	_fletch_quantity_input.min_value = 1
	_fletch_quantity_input.step = 1
	_fletch_quantity_input.custom_minimum_size = Vector2(300, 44)
	_fletch_quantity_input.position = Vector2(40, 74)
	_fletch_quantity_dialog.add_child(_fletch_quantity_input)
	_fletch_quantity_dialog.confirmed.connect(_confirm_fletch_quantity)
	add_child(_fletch_quantity_dialog)


func _open_fletch_quantity_dialog() -> void:
	var maximum := _maximum_fletch_batches()
	if maximum < 1 or _fletch_pending:
		return
	_fletch_quantity_input.max_value = maximum
	_fletch_quantity_input.value = mini(maximum, 5)
	_fletch_quantity_dialog.popup_centered()
	_fletch_quantity_input.grab_focus()
	_fletch_quantity_input.get_line_edit().select_all()


func _confirm_fletch_quantity() -> void:
	_request_fletch(int(_fletch_quantity_input.value))


func _request_fletch(quantity: int = 1) -> void:
	if _fletch_pending or _maximum_fletch_batches() < 1:
		return
	if network.send_fletch(_next_fletch_sequence, _selected_fletching_recipe, quantity):
		_next_fletch_sequence += 1
		_fletch_pending = true
		_fletch_remaining = _maximum_fletch_batches() if quantity == 0 else quantity
		_refresh_fletching_window()


func _on_crafting_recipe_selected(menu_id: int) -> void:
	var action: Dictionary = _crafting_menu_actions.get(menu_id, {}) as Dictionary
	var recipe_id: String = str(action.get("recipe_id", ""))
	var instance_id: String = str(action.get("instance_id", ""))
	if recipe_id.is_empty() or _selected_station_id.is_empty():
		return
	_request_gather(_selected_station_id, _selected_station_position, recipe_id, instance_id)


func _open_world_context_menu(screen_position: Vector2) -> void:
	_world_context_menu.clear()
	_context_menu_actions.clear()
	var next_menu_id: int = 0
	var target_id: String = _target_id_at_screen_position(screen_position)
	if not target_id.is_empty():
		var target_state: Dictionary = _target_states.get(target_id, {})
		var target_name: String = str(target_state.get("name", target_id.capitalize()))
		_world_context_menu.add_item("Attack %s" % target_name, next_menu_id)
		_context_menu_actions[next_menu_id] = {"type": "attack", "target_id": target_id}
		next_menu_id += 1

	var loot_candidates: Array[Dictionary] = _loot_near_screen_position(screen_position)
	if not target_id.is_empty() and not loot_candidates.is_empty():
		_world_context_menu.add_separator()
	for candidate: Dictionary in loot_candidates:
		var quantity: int = int(candidate.get("quantity", 1))
		var item_name: String = str(candidate.get("item_id", "Item")).capitalize()
		var label: String = "Take %s" % item_name
		if quantity > 1:
			label += " ×%d" % quantity
		_world_context_menu.add_item(label, next_menu_id)
		_context_menu_actions[next_menu_id] = {
			"type": "take",
			"drop_id": str(candidate.get("drop_id", "")),
		}
		next_menu_id += 1

	if _context_menu_actions.is_empty():
		return
	_world_context_menu.position = Vector2i(screen_position)
	_world_context_menu.popup()


func _target_id_at_screen_position(screen_position: Vector2) -> String:
	var ray_origin: Vector3 = camera.project_ray_origin(screen_position)
	var ray_end: Vector3 = ray_origin + camera.project_ray_normal(screen_position) * RAY_LENGTH
	var query := PhysicsRayQueryParameters3D.create(ray_origin, ray_end)
	query.collision_mask = TARGET_COLLISION_MASK
	var result: Dictionary = get_world_3d().direct_space_state.intersect_ray(query)
	if result.is_empty():
		return ""
	var body := result.get("collider") as CollisionObject3D
	if body == null or not body.has_meta(&"combat_target_id"):
		return ""
	return str(body.get_meta(&"combat_target_id"))


func _loot_near_screen_position(screen_position: Vector2) -> Array[Dictionary]:
	var candidates: Array[Dictionary] = []
	for drop_id: String in _drop_views.keys():
		var view := _drop_views.get(drop_id) as LootDrop
		if view == null or not is_instance_valid(view) or camera.is_position_behind(view.global_position):
			continue
		var projected_position: Vector2 = camera.unproject_position(view.global_position + Vector3.UP * 0.35)
		var screen_distance: float = projected_position.distance_to(screen_position)
		if screen_distance > CONTEXT_LOOT_RADIUS_PIXELS:
			continue
		var state: Dictionary = _drop_states.get(drop_id, {})
		candidates.append({
			"drop_id": drop_id,
			"item_id": str(state.get("item_id", "item")),
			"quantity": int(state.get("quantity", 1)),
			"screen_distance": screen_distance,
		})
	candidates.sort_custom(_context_loot_candidate_before)
	return candidates


func _context_loot_candidate_before(left: Dictionary, right: Dictionary) -> bool:
	var left_distance: float = float(left.get("screen_distance", INF))
	var right_distance: float = float(right.get("screen_distance", INF))
	if not is_equal_approx(left_distance, right_distance):
		return left_distance < right_distance
	var left_id: String = str(left.get("drop_id", ""))
	var right_id: String = str(right.get("drop_id", ""))
	if _drop_sequence(left_id) != _drop_sequence(right_id):
		return _drop_sequence(left_id) > _drop_sequence(right_id)
	var left_name: String = str(left.get("item_id", ""))
	var right_name: String = str(right.get("item_id", ""))
	if left_name != right_name:
		return left_name < right_name
	return left_id > right_id


func _on_world_context_action(menu_id: int) -> void:
	var action: Dictionary = _context_menu_actions.get(menu_id, {})
	match str(action.get("type", "")):
		"attack":
			var target_id: String = str(action.get("target_id", ""))
			var target_view := _target_views.get(target_id) as TargetDummy
			if target_view != null and is_instance_valid(target_view):
				_request_engage_target(target_id, target_view.global_position)
		"take":
			var drop_id: String = str(action.get("drop_id", ""))
			var drop_view := _drop_views.get(drop_id) as LootDrop
			if drop_view != null and is_instance_valid(drop_view):
				_request_pickup_drop(drop_id, drop_view.global_position)


func _top_loot_hit(ray_origin: Vector3, ray_end: Vector3) -> Dictionary:
	var excluded: Array[RID] = []
	var best_hit: Dictionary = {}
	var best_distance_squared: float = INF
	var best_drop_id: String = ""
	for _index: int in range(32):
		var query := PhysicsRayQueryParameters3D.create(ray_origin, ray_end)
		query.collision_mask = LOOT_COLLISION_MASK
		query.exclude = excluded
		var hit: Dictionary = get_world_3d().direct_space_state.intersect_ray(query)
		if hit.is_empty():
			break
		var body := hit.get("collider") as CollisionObject3D
		if body == null:
			break
		excluded.append(body.get_rid())
		if not body.has_meta(&"drop_id"):
			continue
		var hit_position: Vector3 = hit.get("position", body.global_position)
		var distance_squared: float = ray_origin.distance_squared_to(hit_position)
		var drop_id: String = str(body.get_meta(&"drop_id"))
		if distance_squared < best_distance_squared or (
			is_equal_approx(distance_squared, best_distance_squared) and _drop_is_newer(drop_id, best_drop_id)
		):
			best_hit = hit
			best_distance_squared = distance_squared
			best_drop_id = drop_id
	return best_hit


func _can_predict_directly(
	local_view: PlayerView,
	destination: Vector3,
	clicked_body: CollisionObject3D,
) -> bool:
	if clicked_body == null:
		return false
	if clicked_body.name != &"GroundBody" and clicked_body.name != &"UpperFloorBody":
		return false

	var clearance := Vector3.UP * 0.5
	var query := PhysicsRayQueryParameters3D.create(
		local_view.global_position + clearance,
		destination + clearance,
	)
	query.collision_mask = _active_collision_mask()
	return get_world_3d().direct_space_state.intersect_ray(query).is_empty()


func _play_target_click_feedback(target_position: Vector3) -> void:
	if _target_click_tween != null and _target_click_tween.is_valid():
		_target_click_tween.kill()
	target_click_pulse.global_position = target_position + Vector3.UP * 0.06
	target_click_pulse.scale = Vector3.ONE * 0.55
	target_click_pulse.transparency = 0.0
	target_click_pulse.visible = true
	_target_click_tween = create_tween()
	_target_click_tween.set_parallel(true)
	_target_click_tween.tween_property(
		target_click_pulse,
		"scale",
		Vector3.ONE * 1.35,
		0.22,
	).set_trans(Tween.TRANS_QUART).set_ease(Tween.EASE_OUT)
	_target_click_tween.tween_property(target_click_pulse, "transparency", 1.0, 0.38)
	_target_click_tween.chain().tween_callback(target_click_pulse.hide)


func _request_equip(item_id: String, instance_id: String = "") -> void:
	if _player_defeated:
		return
	var equipment_pending: bool = _last_equipment_sequence < _next_equipment_sequence - 1
	if equipment_pending:
		return
	if _is_item_equipped(item_id, instance_id):
		_request_unequip(item_id, instance_id)
		return
	if network.send_equip_item(_next_equipment_sequence, item_id, instance_id):
		_queued_item = item_id
		_queued_instance_id = instance_id
		_queued_unequip = false
		_next_equipment_sequence += 1
		_update_combat_hud()
		_update_inventory_hud()


func _request_unequip(item_id: String, instance_id: String = "") -> bool:
	if _player_defeated or not _is_item_equipped(item_id, instance_id):
		return false
	var equipment_pending: bool = _last_equipment_sequence < _next_equipment_sequence - 1
	if equipment_pending:
		return false
	if network.send_unequip_item(_next_equipment_sequence, item_id, instance_id):
		_queued_item = item_id
		_queued_instance_id = instance_id
		_queued_unequip = true
		_next_equipment_sequence += 1
		_update_combat_hud()
		_update_inventory_hud()
		_update_equipment_hud()
		return true
	return false


func _is_item_equipped(item_id: String, instance_id: String = "") -> bool:
	if item_id == _equipped_weapon:
		return instance_id.is_empty() or instance_id == _equipped_weapon_instance_id
	if item_id == _equipped_spellbook:
		return true
	var armour_slot: String = str(ARMOUR_SLOTS.get(item_id, ""))
	return not armour_slot.is_empty() and str(_equipped_armour.get(armour_slot, "")) == item_id and (
		instance_id.is_empty() or instance_id == str(_equipped_armour_instances.get(armour_slot, ""))
	)


func _on_inventory_item_input(event: InputEvent, item_id: String, instance_id: String = "") -> void:
	if _player_defeated:
		return
	if bank_hud.visible:
		return
	if event is InputEventMouseButton:
		var mouse_button := event as InputEventMouseButton
		if mouse_button.pressed and mouse_button.button_index == MOUSE_BUTTON_RIGHT:
			var equipment_pending: bool = _last_equipment_sequence < _next_equipment_sequence - 1
			if equipment_pending and _queued_item == item_id:
				return
			if _is_item_equipped(item_id, instance_id):
				if not SPELLBOOK_ITEMS.has(item_id):
					return
				if _request_unequip(item_id, instance_id):
					_queued_drop_after_unequip = item_id
			else:
				_request_drop_item(item_id, instance_id)
			get_viewport().set_input_as_handled()


func _request_drop_item(item_id: String, instance_id: String = "") -> void:
	if item_id.is_empty() or _player_defeated or _is_item_equipped(item_id, instance_id):
		return
	if network.send_drop_item(_next_drop_sequence, item_id, instance_id):
		_next_drop_sequence += 1


func _request_consume_item(item_id: String) -> void:
	if _player_defeated or not CONSUMABLE_ITEMS.has(item_id) or not _queued_consume_item.is_empty():
		return
	var consumable: Dictionary = CONSUMABLE_ITEMS[item_id] as Dictionary
	var restores_health: bool = int(consumable.get("healing", 0)) > 0
	var restores_prayer: bool = int(consumable.get("prayer_restore", 0)) > 0
	var can_benefit: bool = (
		restores_health and _player_health < _player_maximum_health
	) or (
		restores_prayer and _player_prayer < _player_maximum_prayer
	)
	if not can_benefit:
		status_label.text = "You are already at full %s" % ("prayer" if restores_prayer else "health")
		return
	if _action_pulse < _next_consume_pulse:
		status_label.text = "Consumable cooldown: %d pulse(s) remaining" % (_next_consume_pulse - _action_pulse)
		return
	if network.send_consume_item(_next_consume_sequence, item_id):
		_queued_consume_item = item_id
		_next_consume_sequence += 1
		_update_inventory_hud()


func _request_inventory_swap(from_slot: int, to_slot: int) -> void:
	if _player_defeated or from_slot == to_slot:
		return
	if network.send_move_inventory_item(_next_inventory_move_sequence, from_slot, to_slot):
		_next_inventory_move_sequence += 1


func _request_combat_style(combat_style: String) -> void:
	if _player_defeated:
		return
	var style_pending: bool = _last_combat_style_sequence < _next_combat_style_sequence - 1
	if combat_style == _queued_combat_style:
		_update_combat_styles_hud()
		return
	if combat_style == _combat_style and not style_pending:
		_update_combat_styles_hud()
		return
	if network.send_combat_style(_next_combat_style_sequence, combat_style):
		_queued_combat_style = combat_style
		_next_combat_style_sequence += 1
		_update_combat_styles_hud()


func _request_prayer(prayer_id: String) -> void:
	if _player_defeated:
		return
	var active: bool = _active_prayers.has(prayer_id)
	if _queued_prayer_id == prayer_id:
		active = _queued_prayer_active
	var desired_active: bool = not active
	if desired_active and _player_prayer <= 0:
		return
	if network.send_prayer(_next_prayer_sequence, prayer_id, desired_active):
		_queued_prayer_id = prayer_id
		_queued_prayer_active = desired_active
		_next_prayer_sequence += 1
		_update_prayer_hud()


func _request_cut_gem(item_id: String) -> void:
	if network.send_cut_gem(_next_gem_cut_sequence, item_id):
		_next_gem_cut_sequence += 1
		status_label.text = "Cutting %s…" % str(UNCUT_GEMS.get(item_id, "gem"))


func _request_spell(spell_id: String) -> void:
	if _player_defeated:
		return
	if _spell_selection_pending:
		_update_magic_hud()
		return
	var desired_spell: String = "" if spell_id == _selected_spell else spell_id
	if network.send_spell(_next_spell_sequence, desired_spell):
		_queued_spell = desired_spell
		_spell_selection_pending = true
		_next_spell_sequence += 1
		_update_magic_hud()


func _request_disengage() -> void:
	if _player_defeated:
		return
	if network.send_disengage(_next_action_sequence):
		_next_action_sequence += 1
		_update_combat_hud()


func _active_collision_mask() -> int:
	if _local_navigation_kind == "layer":
		match _local_navigation_id:
			"ground":
				return GROUND_COLLISION_MASK | CONNECTOR_COLLISION_MASK
			"upper":
				return UPPER_COLLISION_MASK | CONNECTOR_COLLISION_MASK
	elif _local_navigation_kind == "connector":
		return ALL_WORLD_COLLISION_MASK
	return ALL_WORLD_COLLISION_MASK


func _on_status_changed(status: String) -> void:
	status_label.text = status


func _on_welcome_received(player_id: String) -> void:
	_local_player_id = player_id
	status_label.text = "Connected as %s — click a walkable surface to move" % player_id
	for id: String in _player_views:
		var view: PlayerView = _player_views[id]
		view.set_local_player(id == _local_player_id)
		if id == _local_player_id:
			camera_rig.follow(view)


func _on_snapshot_received(snapshot: Dictionary) -> void:
	var snapshot_tick: int = int(snapshot.get("tick", 0))
	_action_pulse = int(snapshot.get("action_pulse", 0))
	_sync_targets(snapshot.get("targets", []))
	_sync_drops(snapshot.get("drops", []))
	_sync_gathering_spots(snapshot.get("gathering_spots", []))
	_sync_bank_chests(snapshot.get("bank_chests", []))
	var player_payloads: Variant = snapshot.get("players", [])
	if typeof(player_payloads) != TYPE_ARRAY:
		return

	var active_ids: Dictionary = {}
	for payload: Variant in player_payloads:
		if typeof(payload) != TYPE_DICTIONARY:
			continue
		var player_data: Dictionary = payload
		var id: String = str(player_data.get("id", ""))
		var position_data: Variant = player_data.get("position")
		if id.is_empty() or typeof(position_data) != TYPE_DICTIONARY:
			continue

		var coordinates: Dictionary = position_data
		var position := Vector3(
			float(coordinates.get("x", 0.0)),
			float(coordinates.get("y", 0.0)),
			float(coordinates.get("z", 0.0)),
		)
		var path: Array[Vector3] = []
		var path_data: Variant = player_data.get("path", [])
		if typeof(path_data) == TYPE_ARRAY:
			for waypoint_data: Variant in path_data:
				if typeof(waypoint_data) != TYPE_DICTIONARY:
					continue
				var waypoint: Dictionary = waypoint_data
				path.append(Vector3(
					float(waypoint.get("x", 0.0)),
					float(waypoint.get("y", 0.0)),
					float(waypoint.get("z", 0.0)),
				))

		var view: PlayerView = _player_views.get(id)
		if view == null:
			view = PLAYER_SCENE.instantiate() as PlayerView
			players.add_child(view)
			_player_views[id] = view
			view.set_local_player(id == _local_player_id)
			if id == _local_player_id:
				camera_rig.follow(view)

		var acknowledged_sequence: int = int(player_data.get("last_command_sequence", 0))
		var server_driven_path: bool = (
			bool(player_data.get("auto_approaching", false))
			or bool(player_data.get("auto_picking_up", false))
			or not str(player_data.get("gathering_spot", "")).is_empty()
		)
		var correction: float = view.apply_authoritative_state(
			position,
			path,
			float(player_data.get("speed", 6.0)),
			acknowledged_sequence,
			_next_sequence - 1,
			server_driven_path,
		)
		view.set_weapon(str(player_data.get("equipped_weapon", "sword")))
		var world_armour_data: Variant = player_data.get("equipped_armour", {})
		if typeof(world_armour_data) == TYPE_DICTIONARY:
			view.set_armour(world_armour_data as Dictionary)
		view.set_defeated(bool(player_data.get("defeated", false)))
		if id == _local_player_id:
			var inventory_selection_changed: bool = false
			var combat_style_changed: bool = false
			var prayer_changed: bool = false
			_last_action_sequence = int(player_data.get("last_action_sequence", 0))
			_last_action_result = str(player_data.get("last_action_result", "none"))
			_combat_target = str(player_data.get("combat_target", ""))
			_next_attack_pulse = int(player_data.get("next_attack_pulse", 0))
			_auto_approaching = bool(player_data.get("auto_approaching", false))
			_player_health = int(player_data.get("health", _player_health))
			_player_maximum_health = int(player_data.get("maximum_health", _player_maximum_health))
			var prayer_points: int = int(player_data.get("prayer", _player_prayer))
			if prayer_points != _player_prayer:
				_player_prayer = prayer_points
				prayer_changed = true
			_player_maximum_prayer = int(player_data.get("maximum_prayer", _player_maximum_prayer))
			var active_prayer_set: String = str(player_data.get("active_prayer_set", _active_prayer_set))
			if active_prayer_set != _active_prayer_set:
				_active_prayer_set = active_prayer_set
				prayer_changed = true
			var active_prayers_data: Variant = player_data.get("active_prayers", [])
			if typeof(active_prayers_data) == TYPE_ARRAY:
				var active_prayers: Array[String] = []
				for prayer_id_data: Variant in active_prayers_data as Array:
					active_prayers.append(str(prayer_id_data))
				if active_prayers != _active_prayers:
					_active_prayers = active_prayers
					prayer_changed = true
			_last_prayer_sequence = int(player_data.get("last_prayer_sequence", 0))
			if _last_prayer_sequence >= _next_prayer_sequence - 1 and not _queued_prayer_id.is_empty():
				_queued_prayer_id = ""
				prayer_changed = true
			var selected_spell: String = str(player_data.get("selected_spell", ""))
			if selected_spell != _selected_spell:
				_selected_spell = selected_spell
				_update_magic_hud()
			_last_spell_sequence = int(player_data.get("last_spell_sequence", 0))
			if _spell_selection_pending and _last_spell_sequence >= _next_spell_sequence - 1:
				_spell_selection_pending = false
				_queued_spell = ""
				_update_magic_hud()
			var player_defeated: bool = bool(player_data.get("defeated", false))
			if player_defeated != _player_defeated:
				_player_defeated = player_defeated
				prayer_changed = true
				_update_magic_hud()
			_player_respawn_at_pulse = int(player_data.get("respawn_at_pulse", 0))
			_player_dodge_rate = float(player_data.get("dodge_rate", _player_dodge_rate))
			var combat_stats_data: Variant = player_data.get("combat_stats", {})
			if typeof(combat_stats_data) == TYPE_DICTIONARY and combat_stats_data != _combat_stats:
				_combat_stats = (combat_stats_data as Dictionary).duplicate(true)
				_update_detailed_combat_stats()
			var skills_data: Variant = player_data.get("skills", [])
			if typeof(skills_data) == TYPE_ARRAY:
				_apply_skills_snapshot(skills_data as Array)
			var inventory_data: Variant = player_data.get("inventory", {})
			if typeof(inventory_data) == TYPE_DICTIONARY:
				_apply_inventory_snapshot(inventory_data as Dictionary)
			var ammunition_data: Variant = player_data.get("ammunition_pouch", {})
			if typeof(ammunition_data) == TYPE_DICTIONARY:
				_apply_ammunition_pouch_snapshot(ammunition_data as Dictionary)
			var last_ammunition_sequence: int = int(player_data.get("last_ammunition_sequence", 0))
			_next_ammunition_sequence = maxi(_next_ammunition_sequence, last_ammunition_sequence + 1)
			var bank_data: Variant = player_data.get("bank", {})
			if typeof(bank_data) == TYPE_DICTIONARY:
				_apply_bank_snapshot(bank_data as Dictionary)
			var last_bank_sequence: int = int(player_data.get("last_bank_sequence", 0))
			_next_bank_sequence = maxi(_next_bank_sequence, last_bank_sequence + 1)
			_update_bank_proximity(position)
			_last_consume_sequence = int(player_data.get("last_consume_sequence", 0))
			_next_consume_pulse = int(player_data.get("next_consume_pulse", 0))
			_next_consume_sequence = maxi(_next_consume_sequence, _last_consume_sequence + 1)
			if _last_consume_sequence >= _next_consume_sequence - 1 and not _queued_consume_item.is_empty():
				_queued_consume_item = ""
				inventory_selection_changed = true
			_last_brew_sequence = int(player_data.get("last_brew_sequence", 0))
			_next_brew_sequence = maxi(_next_brew_sequence, _last_brew_sequence + 1)
			var previous_brew_remaining: int = _brew_remaining
			_brew_remaining = int(player_data.get("brew_remaining", 0))
			_next_brew_pulse = int(player_data.get("next_brew_pulse", 0))
			var brewing_result_text: String = str(player_data.get("brewing_result", ""))
			if _brew_pending and _brew_remaining <= 0 and _last_brew_sequence >= _next_brew_sequence - 1:
				_brew_pending = false
				_brewing_selection.clear()
				status_label.text = "Brewing complete • Herblore XP gained"
			elif _brew_remaining > 0:
				_brew_pending = true
				if brewing_result_text != _brewing_result_text or previous_brew_remaining != _brew_remaining:
					status_label.text = "Brewing… %d potion%s remaining" % [
						_brew_remaining, "" if _brew_remaining == 1 else "s",
					]
			_brewing_result_text = brewing_result_text
			if brewing_window.visible:
				_refresh_brewing_window()
			_last_fletch_sequence = int(player_data.get("last_fletch_sequence", 0))
			_next_fletch_sequence = maxi(_next_fletch_sequence, _last_fletch_sequence + 1)
			var previous_fletch_remaining: int = _fletch_remaining
			_fletch_remaining = int(player_data.get("fletch_remaining", 0))
			_next_fletch_pulse = int(player_data.get("next_fletch_pulse", 0))
			var fletching_result_text: String = str(player_data.get("fletching_result", ""))
			if _fletch_pending and _fletch_remaining <= 0 and _last_fletch_sequence >= _next_fletch_sequence - 1:
				_fletch_pending = false
				status_label.text = "Fletching complete • Fletching XP gained"
			elif _fletch_remaining > 0:
				_fletch_pending = true
				if fletching_result_text != _fletching_result_text or previous_fletch_remaining != _fletch_remaining:
					status_label.text = "Fletching… %d batch%s remaining" % [
						_fletch_remaining, "" if _fletch_remaining == 1 else "es",
					]
			_fletching_result_text = fletching_result_text
			if fletching_window.visible:
				_refresh_fletching_window()
			_last_gather_sequence = int(player_data.get("last_gather_sequence", 0))
			_next_gather_sequence = maxi(_next_gather_sequence, _last_gather_sequence + 1)
			_gathering_spot = str(player_data.get("gathering_spot", ""))
			_next_gather_pulse = int(player_data.get("next_gather_pulse", 0))
			var gathering_result_text: String = str(player_data.get("gathering_result", ""))
			if gathering_result_text != _gathering_result:
				var previous_gathering_result: String = _gathering_result
				_gathering_result = gathering_result_text
				match _gathering_result:
					"":
						if not previous_gathering_result.is_empty():
							status_label.text = "Gathering stopped"
					"approaching":
						status_label.text = "Walking to the resource…"
					"balancing":
						status_label.text = "Crossing the balance beam…"
					"completed_agility_obstacle":
						status_label.text = "Balance beam completed • Agility XP gained"
					"praying":
						status_label.text = "Praying at the altar…"
					"prayer_restored":
						status_label.text = "Prayer points fully restored"
					"prayer_set_common_litany":
						status_label.text = "Prayer set changed to Common Litany"
					"prayer_set_war_covenant":
						status_label.text = "Prayer set changed to War Covenant"
					"sacrificed_bone":
						status_label.text = "Sacrificing bones… • 10 Prayer XP"
					"missing_bones":
						status_label.text = "All bones sacrificed"
					"sacrificed_all_bones":
						status_label.text = "All bones sacrificed • Prayer XP gained"
					"fishing":
						status_label.text = "Fishing…"
					"missing_fishing_rod":
						status_label.text = "Fishing stopped: a Fishing Rod is required"
					"missing_lobster_cage":
						status_label.text = "Fishing stopped: a Lobster Cage is required"
					"caught_raw_minnow":
						status_label.text = "Caught a Raw Minnow • Fishing continues…"
					"cooking":
						status_label.text = "Cooking…"
					"cooked_minnow":
						status_label.text = "Cooked a Minnow • Cooking continues…"
					"missing_raw_minnow":
						status_label.text = "Cooking stopped: no Raw Minnows"
					"mining":
						status_label.text = "Mining…"
					"mined_copper_ore":
						status_label.text = "Mined Copper Ore • Mining continues…"
					"mined_tin_ore":
						status_label.text = "Mined Tin Ore • Mining continues…"
					"mined_iron_ore":
						status_label.text = "Mined Iron Ore • Mining continues…"
					"mined_coal":
						status_label.text = "Mined Coal • Mining continues…"
					"mined_mithril_ore":
						status_label.text = "Mined Mithril Ore • Mining continues…"
					"woodcutting":
						status_label.text = "Chopping…"
					"chopped_normal_logs":
						status_label.text = "Chopped Normal Logs • Woodcutting continues…"
					"chopped_oak_logs":
						status_label.text = "Chopped Oak Logs • Woodcutting continues…"
					"missing_bronze_axe":
						status_label.text = "Woodcutting stopped: a Bronze Axe is required"
					"tree_depleted":
						status_label.text = "Woodcutting stopped: the tree is depleted"
					"smelting":
						status_label.text = "Smelting…"
					"smelted_bronze_bar":
						status_label.text = "Smelted a Bronze Bar • Smithing continues…"
					"smelted_iron_bar":
						status_label.text = "Smelted an Iron Bar • Smithing continues…"
					"smelted_mithril_bar":
						status_label.text = "Smelted a Mithril Bar • Smithing continues…"
					"missing_smelting_inputs":
						status_label.text = "Smithing stopped: required ores are missing"
					"forging":
						status_label.text = "Forging…"
					"forged_bronze_sword":
						status_label.text = "Forged a Bronze Sword • Smithing continues…"
					"forged_bronze_helm":
						status_label.text = "Forged a Bronze Helm • Smithing continues…"
					"forged_bronze_pickaxe":
						status_label.text = "Forged a Bronze Pickaxe • Smithing continues…"
					"forged_iron_arrow_tips":
						status_label.text = "Forged 20 Iron Arrow Tips • Smithing continues…"
					"missing_forging_inputs":
						status_label.text = "Smithing stopped: required bars are missing"
					"reinforcement_failed":
						status_label.text = "Reinforcement stopped: item or bars are unavailable"
					"missing_bronze_pickaxe":
						status_label.text = "Mining stopped: a Bronze Pickaxe is required"
					"vein_depleted":
						status_label.text = "Mining stopped: the vein is depleted"
					"inventory_full":
						status_label.text = "Gathering stopped: inventory is full"
					_:
						if _gathering_result.begins_with("caught_raw_"):
							var caught_name: String = _gathering_result.trim_prefix("caught_").replace("_", " ").capitalize()
							status_label.text = "Caught a %s • Fishing continues…" % caught_name
						elif _gathering_result.begins_with("cooked_"):
							var cooked_name: String = _gathering_result.trim_prefix("cooked_").replace("_", " ").capitalize()
							status_label.text = "Cooked a %s • Cooking continues…" % cooked_name
						elif _gathering_result.begins_with("missing_raw_"):
							var missing_name: String = _gathering_result.trim_prefix("missing_").replace("_", " ").capitalize()
							status_label.text = "Cooking stopped: no %s" % missing_name
						elif _gathering_result.begins_with("forged_"):
							var forged_name: String = _gathering_result.trim_prefix("forged_").replace("_", " ").capitalize()
							status_label.text = "Forged a %s • Smithing continues…" % forged_name
						elif _gathering_result.begins_with("reinforced_"):
							var reinforcement_result := _gathering_result.trim_prefix("reinforced_")
							var result_parts := reinforcement_result.rsplit("_to_", true, 1)
							var reinforced_name: String = str(result_parts[0]).replace("_", " ").capitalize()
							var reinforced_level: String = str(result_parts[1]) if result_parts.size() > 1 else ""
							status_label.text = "%s reinforced to +%s" % [reinforced_name, reinforced_level]
			var equipped_weapon: String = str(player_data.get("equipped_weapon", "unarmed"))
			var equipped_weapon_instance_id: String = str(player_data.get("equipped_weapon_instance_id", ""))
			var equipped_weapon_reinforcement: int = int(player_data.get("equipped_weapon_reinforcement", 0))
			if equipped_weapon != _equipped_weapon or equipped_weapon_instance_id != _equipped_weapon_instance_id or equipped_weapon_reinforcement != _equipped_weapon_reinforcement:
				_equipped_weapon = equipped_weapon
				_equipped_weapon_instance_id = equipped_weapon_instance_id
				_equipped_weapon_reinforcement = equipped_weapon_reinforcement
				_update_equipment_hud()
				inventory_selection_changed = true
				combat_style_changed = true
			var equipped_armour_data: Variant = player_data.get("equipped_armour", {})
			var equipped_armour_instances_data: Variant = player_data.get("equipped_armour_instances", {})
			var equipped_armour_reinforcements_data: Variant = player_data.get("equipped_armour_reinforcements", {})
			if typeof(equipped_armour_data) == TYPE_DICTIONARY and (
				equipped_armour_data != _equipped_armour or equipped_armour_instances_data != _equipped_armour_instances or equipped_armour_reinforcements_data != _equipped_armour_reinforcements
			):
				_equipped_armour = (equipped_armour_data as Dictionary).duplicate(true)
				_equipped_armour_instances = (
					(equipped_armour_instances_data as Dictionary).duplicate(true)
					if typeof(equipped_armour_instances_data) == TYPE_DICTIONARY else {}
				)
				_equipped_armour_reinforcements = (
					(equipped_armour_reinforcements_data as Dictionary).duplicate(true)
					if typeof(equipped_armour_reinforcements_data) == TYPE_DICTIONARY else {}
				)
				_update_equipment_hud()
				inventory_selection_changed = true
			var equipped_spellbook: String = str(player_data.get("equipped_spellbook", ""))
			var active_spellbook: String = str(player_data.get("active_spellbook", ""))
			if equipped_spellbook != _equipped_spellbook or active_spellbook != _active_spellbook:
				_equipped_spellbook = equipped_spellbook
				_active_spellbook = active_spellbook
				_update_equipment_hud()
				_update_magic_hud()
				inventory_selection_changed = true
			_last_equipment_sequence = int(player_data.get("last_equipment_sequence", 0))
			_next_equipment_sequence = maxi(
				_next_equipment_sequence,
				_last_equipment_sequence + 1,
			)
			if _last_equipment_sequence >= _next_equipment_sequence - 1 and not _queued_item.is_empty():
				var completed_item: String = _queued_item
				var completed_unequip: bool = _queued_unequip
				_queued_item = ""
				_queued_instance_id = ""
				_queued_unequip = false
				inventory_selection_changed = true
				if completed_unequip and _queued_drop_after_unequip == completed_item and not _is_item_equipped(completed_item):
					_queued_drop_after_unequip = ""
					_request_drop_item(completed_item)
			var combat_style: String = str(player_data.get("combat_style", "crush"))
			if combat_style != _combat_style:
				_combat_style = combat_style
				combat_style_changed = true
			_last_combat_style_sequence = int(player_data.get("last_combat_style_sequence", 0))
			if _last_combat_style_sequence >= _next_combat_style_sequence - 1 and not _queued_combat_style.is_empty():
				_queued_combat_style = ""
				combat_style_changed = true
			if inventory_selection_changed:
				_update_inventory_hud()
			if combat_style_changed:
				_update_combat_styles_hud()
			if prayer_changed:
				_update_prayer_hud()
			if _last_action_result.is_empty():
				_last_action_result = "none"
			var navigation_data: Variant = player_data.get("navigation_location", {})
			if typeof(navigation_data) == TYPE_DICTIONARY:
				var navigation_location: Dictionary = navigation_data
				_local_navigation_kind = str(navigation_location.get("kind", "unknown"))
				_local_navigation_id = str(navigation_location.get("id", ""))
			upper_level.set_navigation_location(_local_navigation_kind, _local_navigation_id)
			camera_rig.set_world_collision_mask(_active_collision_mask())
			diagnostics_label.text = "tick %d  |  sent %d  |  acknowledged %d  |  correction %.3f  |  %s:%s" % [
				snapshot_tick,
				_next_sequence - 1,
				acknowledged_sequence,
				correction,
				_local_navigation_kind,
				_local_navigation_id,
			]
		active_ids[id] = true
	_update_combat_hud()

	for id: String in _player_views.keys():
		if not active_ids.has(id):
			var departed: PlayerView = _player_views[id]
			if id == _local_player_id:
				camera_rig.follow(null)
			departed.queue_free()
			_player_views.erase(id)
	_process_combat_events(snapshot)


func _sync_targets(target_payloads: Variant) -> void:
	if typeof(target_payloads) != TYPE_ARRAY:
		return
	var active_ids: Dictionary = {}
	for payload: Variant in target_payloads:
		if typeof(payload) != TYPE_DICTIONARY:
			continue
		var target_data: Dictionary = payload
		var target_id: String = str(target_data.get("id", ""))
		var position_data: Variant = target_data.get("position", {})
		if target_id.is_empty() or typeof(position_data) != TYPE_DICTIONARY:
			continue
		var view: TargetDummy = _target_views.get(target_id)
		if view == null:
			view = TARGET_SCENE.instantiate() as TargetDummy
			view.name = "Target_%s" % target_id
			targets.add_child(view)
			view.set_target_id(target_id)
			view.set_target_kind(str(target_data.get("kind", "dummy")))
			_target_views[target_id] = view
		var coordinates: Dictionary = position_data
		view.set_authoritative_position(Vector3(
			float(coordinates.get("x", 0.0)),
			float(coordinates.get("y", 0.0)),
			float(coordinates.get("z", 0.0)),
		))
		view.set_defeated(bool(target_data.get("defeated", false)))
		_target_states[target_id] = target_data.duplicate(true)
		active_ids[target_id] = true

	for target_id: String in _target_views.keys():
		if active_ids.has(target_id):
			continue
		var departed: TargetDummy = _target_views[target_id]
		departed.queue_free()
		_target_views.erase(target_id)
		_target_states.erase(target_id)


func _sync_drops(drop_payloads: Variant) -> void:
	if typeof(drop_payloads) != TYPE_ARRAY:
		return
	var active_ids: Dictionary = {}
	var stack_layout_changed: bool = false
	for payload: Variant in drop_payloads:
		if typeof(payload) != TYPE_DICTIONARY:
			continue
		var drop_data: Dictionary = payload
		var drop_id: String = str(drop_data.get("id", ""))
		var position_data: Variant = drop_data.get("position", {})
		if drop_id.is_empty() or typeof(position_data) != TYPE_DICTIONARY:
			continue
		var view: LootDrop = _drop_views.get(drop_id)
		if view == null:
			view = LOOT_SCENE.instantiate() as LootDrop
			view.name = "Drop_%s" % drop_id
			drops.add_child(view)
			view.set_drop_id(drop_id)
			_drop_views[drop_id] = view
			stack_layout_changed = true
		view.set_item_id(
			str(drop_data.get("item_id", "")),
			int(drop_data.get("quantity", 1)),
		)
		_drop_states[drop_id] = drop_data.duplicate(true)
		var coordinates: Dictionary = position_data
		view.position = Vector3(
			float(coordinates.get("x", 0.0)),
			float(coordinates.get("y", 0.0)),
			float(coordinates.get("z", 0.0)),
		)
		active_ids[drop_id] = true

	for drop_id: String in _drop_views.keys():
		if active_ids.has(drop_id):
			continue
		var collected: LootDrop = _drop_views[drop_id]
		collected.queue_free()
		_drop_views.erase(drop_id)
		_drop_states.erase(drop_id)
		stack_layout_changed = true
	if stack_layout_changed:
		_arrange_drop_stacks()
	_update_test_supply_hud()


func _arrange_drop_stacks() -> void:
	var stacks: Dictionary = {}
	for drop_id: String in _drop_views.keys():
		var state: Dictionary = _drop_states.get(drop_id, {}) as Dictionary
		var position_data: Variant = state.get("position", {})
		if typeof(position_data) != TYPE_DICTIONARY:
			continue
		var coordinates := position_data as Dictionary
		var key := "%.2f:%.2f:%.2f" % [
			float(coordinates.get("x", 0.0)),
			float(coordinates.get("y", 0.0)),
			float(coordinates.get("z", 0.0)),
		]
		var stack: Array = stacks.get(key, []) as Array
		stack.append(drop_id)
		stacks[key] = stack
	for stack_data: Variant in stacks.values():
		var stack := stack_data as Array
		stack.sort_custom(_drop_is_newer)
		for index: int in range(stack.size()):
			var view := _drop_views.get(str(stack[index])) as LootDrop
			if view != null and is_instance_valid(view):
				view.set_stack_top(index == 0)


func _drop_is_newer(left_id: String, right_id: String) -> bool:
	var left_sequence: int = _drop_sequence(left_id)
	var right_sequence: int = _drop_sequence(right_id)
	if left_sequence != right_sequence:
		return left_sequence > right_sequence
	return left_id > right_id


func _drop_sequence(drop_id: String) -> int:
	if not drop_id.begins_with("drop-"):
		return -1
	return int(drop_id.trim_prefix("drop-"))


func _sync_gathering_spots(spot_payloads: Variant) -> void:
	if typeof(spot_payloads) != TYPE_ARRAY:
		return
	var active_ids: Dictionary = {}
	for payload: Variant in spot_payloads:
		if typeof(payload) != TYPE_DICTIONARY:
			continue
		var spot_data: Dictionary = payload
		var spot_id: String = str(spot_data.get("id", ""))
		var position_data: Variant = spot_data.get("position", {})
		if spot_id.is_empty() or typeof(position_data) != TYPE_DICTIONARY:
			continue
		var view := _gathering_views.get(spot_id) as GatheringSpot
		if view == null:
			view = GATHERING_SCENE.instantiate() as GatheringSpot
			view.name = "Gathering_%s" % spot_id
			gathering_spots.add_child(view)
			view.set_spot(
				spot_id,
				str(spot_data.get("name", spot_id.capitalize())),
				str(spot_data.get("skill", "gathering")),
				str(spot_data.get("station", "")),
			)
			_gathering_views[spot_id] = view
		view.set_amount(int(spot_data.get("current_amount", 0)), int(spot_data.get("maximum_amount", 0)))
		var coordinates: Dictionary = position_data
		view.position = Vector3(
			float(coordinates.get("x", 0.0)),
			float(coordinates.get("y", 0.0)),
			float(coordinates.get("z", 0.0)),
		)
		_gathering_states[spot_id] = spot_data.duplicate(true)
		active_ids[spot_id] = true

	for spot_id: String in _gathering_views.keys():
		if active_ids.has(spot_id):
			continue
		var departed := _gathering_views.get(spot_id) as GatheringSpot
		if departed != null:
			departed.queue_free()
		_gathering_views.erase(spot_id)
		_gathering_states.erase(spot_id)


func _sync_bank_chests(chest_payloads: Variant) -> void:
	if typeof(chest_payloads) != TYPE_ARRAY:
		return
	var active_ids: Dictionary = {}
	for payload: Variant in chest_payloads:
		if typeof(payload) != TYPE_DICTIONARY:
			continue
		var chest_data: Dictionary = payload
		var chest_id: String = str(chest_data.get("id", ""))
		var position_data: Variant = chest_data.get("position", {})
		if chest_id.is_empty() or typeof(position_data) != TYPE_DICTIONARY:
			continue
		var view := _bank_chest_views.get(chest_id) as BankChest
		if view == null:
			view = BANK_CHEST_SCENE.instantiate() as BankChest
			view.name = "Bank_%s" % chest_id
			bank_chests.add_child(view)
			view.set_chest(chest_id, str(chest_data.get("name", "Bank Chest")), float(chest_data.get("range", 2.5)))
			_bank_chest_views[chest_id] = view
		var coordinates: Dictionary = position_data
		view.position = Vector3(
			float(coordinates.get("x", 0.0)),
			float(coordinates.get("y", 0.0)),
			float(coordinates.get("z", 0.0)),
		)
		_bank_chest_states[chest_id] = chest_data.duplicate(true)
		active_ids[chest_id] = true

	for chest_id: String in _bank_chest_views.keys():
		if active_ids.has(chest_id):
			continue
		var departed := _bank_chest_views.get(chest_id) as BankChest
		if departed != null:
			departed.queue_free()
		_bank_chest_views.erase(chest_id)
		_bank_chest_states.erase(chest_id)
		if chest_id == _active_bank_chest_id or chest_id == _pending_bank_chest_id:
			_close_bank()


func _process_combat_events(snapshot: Dictionary) -> void:
	var events_data: Variant = snapshot.get("combat_events", [])
	if typeof(events_data) != TYPE_ARRAY:
		return
	var events: Array = events_data

	if not _combat_events_initialized:
		for payload: Variant in events:
			if typeof(payload) == TYPE_DICTIONARY:
				_last_combat_event_id = maxi(
					_last_combat_event_id,
					int((payload as Dictionary).get("id", 0)),
				)
		_combat_events_initialized = true
		return

	for payload: Variant in events:
		if typeof(payload) != TYPE_DICTIONARY:
			continue
		var event: Dictionary = payload
		var event_id: int = int(event.get("id", 0))
		if event_id <= _last_combat_event_id:
			continue
		_last_combat_event_id = event_id
		var event_type: String = str(event.get("type", ""))
		if event_type != "damage" and event_type != "dodged":
			continue
		var source_id: String = str(event.get("source_id", ""))
		var target_id: String = str(event.get("target_id", ""))
		var amount: int = int(event.get("amount", 0))
		var spell_id: String = str(event.get("spell_id", ""))
		var source: PlayerView = _player_views.get(source_id)
		var target: PlayerView = _player_views.get(target_id)
		var dummy_view: TargetDummy = _dummy_view_for_id(target_id)
		if dummy_view != null:
			if source == null:
				continue
			var impact_delay: float = (
				source.play_spell_attack(dummy_view.global_position, spell_id)
				if not spell_id.is_empty()
				else source.play_weapon_attack(dummy_view.global_position)
			)
			_play_dummy_impact_after_delay(
				dummy_view,
				event_type,
				amount,
				source.global_position,
				impact_delay,
			)
		elif target != null:
			if source != null:
				if not spell_id.is_empty():
					source.play_spell_attack(target.global_position, spell_id)
				else:
					source.play_weapon_attack(target.global_position)
			elif _dummy_view_for_id(source_id) != null:
				var attacking_dummy: TargetDummy = _dummy_view_for_id(source_id)
				attacking_dummy.play_attack(target.global_position)
			if event_type == "damage":
				target.play_hit(amount)
			else:
				target.play_dodge()


func _dummy_view_for_id(target_id: String) -> TargetDummy:
	return _target_views.get(target_id) as TargetDummy


func _play_dummy_impact_after_delay(
	dummy_view: TargetDummy,
	event_type: String,
	amount: int,
	source_position: Vector3,
	delay: float,
) -> void:
	if delay > 0.0:
		await get_tree().create_timer(delay).timeout
	if not is_instance_valid(dummy_view):
		return
	if event_type == "damage":
		dummy_view.play_hit(amount, source_position)
	else:
		dummy_view.play_dodge()


func _on_server_error(code: String, message: String) -> void:
	status_label.text = "%s: %s" % [code, message]
	if code == "combat_style_rejected":
		_queued_combat_style = ""
		_update_combat_styles_hud()
	if code == "prayer_rejected":
		_queued_prayer_id = ""
		_update_prayer_hud()
	if code == "spell_rejected":
		_spell_selection_pending = false
		_queued_spell = ""
		_update_magic_hud()
	if code == "equipment_rejected":
		_next_equipment_sequence = _last_equipment_sequence + 1
		if _queued_drop_after_unequip == _queued_item:
			_queued_drop_after_unequip = ""
		_queued_item = ""
		_queued_unequip = false
		_update_inventory_hud()
		_update_equipment_hud()
	if code == "consume_rejected":
		_next_consume_sequence = _last_consume_sequence + 1
		_queued_consume_item = ""
		_update_inventory_hud()
	if code == "brew_rejected":
		_brew_pending = false
		_brew_remaining = 0
		_next_brew_sequence = _last_brew_sequence + 1
		status_label.text = "Brewing failed: %s" % message
		if brewing_window.visible:
			_refresh_brewing_window()
	if code == "fletch_rejected":
		_fletch_pending = false
		_fletch_remaining = 0
		_next_fletch_sequence = _last_fletch_sequence + 1
		status_label.text = "Fletching failed: %s" % message
		if fletching_window.visible:
			_refresh_fletching_window()
	if code == "gather_rejected":
		_next_gather_sequence = _last_gather_sequence + 1
	if code == "command_rejected":
		var local_view: PlayerView = _player_views.get(_local_player_id)
		if local_view != null:
			local_view.cancel_prediction()


func _configure_player_menu() -> void:
	var tab_icons: Dictionary = {
		"Styles": TAB_STYLES_ICON,
		"Inventory": TAB_INVENTORY_ICON,
		"Skills": TAB_SKILLS_ICON,
		"Equipment": TAB_EQUIPMENT_ICON,
		"Magic": TAB_MAGIC_ICON,
		"Prayer": TAB_PRAYER_ICON,
	}
	var tab_rows: Array[HBoxContainer] = [
		inventory_active_tab_button.get_parent() as HBoxContainer,
		skills_active_tab_button.get_parent() as HBoxContainer,
		equipment_active_tab_button.get_parent() as HBoxContainer,
		magic_active_tab_button.get_parent() as HBoxContainer,
		prayer_active_tab_button.get_parent() as HBoxContainer,
		styles_active_tab_button.get_parent() as HBoxContainer,
	]
	for tab_row: HBoxContainer in tab_rows:
		var combat_button := tab_row.get_node("Combat") as Button
		var styles_button := tab_row.get_node("Styles") as Button
		combat_button.visible = false
		tab_row.move_child(styles_button, 0)
		for tab_name_value: Variant in tab_icons.keys():
			var tab_name: String = str(tab_name_value)
			var tab_button := tab_row.get_node(tab_name) as Button
			tab_button.custom_minimum_size = Vector2(84, 64)
			tab_button.text = ""
			tab_button.icon = tab_icons[tab_name] as Texture2D
			tab_button.expand_icon = true
			tab_button.add_theme_constant_override("icon_max_width", 52)
			tab_button.tooltip_text = tab_name
	(combat_tab_button.get_parent() as HBoxContainer).visible = false
	_show_styles_tab()


func _toggle_combat_debug_hud() -> void:
	if _combat_debug_hud_visible:
		_show_styles_tab()
	else:
		_show_combat_tab()


func _load_client_settings() -> void:
	var config := ConfigFile.new()
	var music_volume: float = 40.0
	var show_diagnostics: bool = true
	if config.load(CLIENT_SETTINGS_PATH) == OK:
		music_volume = clampf(float(config.get_value("audio", "music_volume", music_volume)), 0.0, 100.0)
		show_diagnostics = bool(config.get_value("interface", "show_diagnostics", show_diagnostics))
	settings_music_slider.set_value_no_signal(music_volume)
	settings_diagnostics_toggle.set_pressed_no_signal(show_diagnostics)
	_apply_music_volume(music_volume)
	settings_music_value.text = "%d%%" % roundi(music_volume)
	diagnostics_label.visible = show_diagnostics


func _save_client_settings() -> void:
	var config := ConfigFile.new()
	config.set_value("audio", "music_volume", settings_music_slider.value)
	config.set_value("interface", "show_diagnostics", settings_diagnostics_toggle.button_pressed)
	config.save(CLIENT_SETTINGS_PATH)


func _on_music_volume_changed(volume: float) -> void:
	_apply_music_volume(volume)
	settings_music_value.text = "%d%%" % roundi(volume)
	_save_client_settings()


func _apply_music_volume(volume: float) -> void:
	var music_bus_index := AudioServer.get_bus_index(&"Music")
	if music_bus_index < 0:
		return
	var normalized_volume := clampf(volume / 100.0, 0.0, 1.0)
	AudioServer.set_bus_mute(music_bus_index, normalized_volume <= 0.0)
	if normalized_volume > 0.0:
		AudioServer.set_bus_volume_db(music_bus_index, linear_to_db(normalized_volume))


func _on_diagnostics_toggled(visible: bool) -> void:
	diagnostics_label.visible = visible
	_save_client_settings()


func _begin_hud_editing_from_settings() -> void:
	_set_hud_edit_mode(true)
	status_label.text = "HUD editing enabled • F10 saves and closes"


func _reset_hud_layout_from_settings() -> void:
	_reset_hud_layout()
	status_label.text = "HUD layout reset"


func _setup_hud_layout() -> void:
	var menu_panels: Array[Control] = [
		combat_hud, inventory_hud, skills_hud, equipment_hud, magic_hud, prayer_hud, styles_hud, settings_hud,
	]
	for panel: Control in menu_panels:
		panel.size = MENU_HUD_BASE_SIZE
	_register_hud_group("player", [player_vitals_hud], player_vitals_hud)
	_register_hud_group("target", [target_hud], target_hud)
	_register_hud_group(
		"menu",
		menu_panels,
		inventory_hud,
	)
	_load_hud_layout()
	_create_hud_edit_handles()
	get_viewport().size_changed.connect(_on_hud_viewport_resized)
	_clamp_all_hud_groups()
	_refresh_hud_edit_handles()


func _register_hud_group(group_id: String, panels: Array, reference_panel: Control) -> void:
	var viewport_size: Vector2 = get_viewport().get_visible_rect().size
	var position_ratio := Vector2(
		reference_panel.global_position.x / maxf(1.0, viewport_size.x),
		reference_panel.global_position.y / maxf(1.0, viewport_size.y),
	)
	_hud_groups[group_id] = {
		"panels": panels,
		"position": reference_panel.global_position,
		"scale": 1.0,
		"default_position_ratio": position_ratio,
		"default_scale": 1.0,
	}


func _create_hud_edit_handles() -> void:
	for group_id_value: Variant in _hud_groups.keys():
		var group_id: String = str(group_id_value)
		var move_handle := Button.new()
		move_handle.name = "HudMove_%s" % group_id
		move_handle.text = "MOVE  %s" % group_id.capitalize()
		move_handle.focus_mode = Control.FOCUS_NONE
		move_handle.mouse_default_cursor_shape = Control.CURSOR_MOVE
		move_handle.z_index = 100
		move_handle.visible = false
		move_handle.gui_input.connect(_on_hud_move_handle_input.bind(group_id))
		$UI.add_child(move_handle)

		var resize_handle := Button.new()
		resize_handle.name = "HudResize_%s" % group_id
		resize_handle.text = "↘"
		resize_handle.focus_mode = Control.FOCUS_NONE
		resize_handle.mouse_default_cursor_shape = Control.CURSOR_FDIAGSIZE
		resize_handle.z_index = 101
		resize_handle.visible = false
		resize_handle.gui_input.connect(_on_hud_resize_handle_input.bind(group_id))
		$UI.add_child(resize_handle)
		_hud_handles[group_id] = {"move": move_handle, "resize": resize_handle}


func _set_hud_edit_mode(enabled: bool) -> void:
	_hud_edit_mode = enabled
	hud_edit_hint.visible = enabled
	target_hud.visible = _target_hud_active or enabled
	if not enabled:
		_hud_drag_group = ""
		_hud_scale_group = ""
		_save_hud_layout()
	_refresh_hud_edit_handles()


func _refresh_hud_edit_handles() -> void:
	if _hud_handles.is_empty():
		return
	for group_id_value: Variant in _hud_handles.keys():
		var group_id: String = str(group_id_value)
		var handles: Dictionary = _hud_handles[group_id]
		var move_handle := handles["move"] as Button
		var resize_handle := handles["resize"] as Button
		var panel: Control = _get_visible_hud_panel(group_id)
		var show_handles: bool = _hud_edit_mode and panel != null
		move_handle.visible = show_handles
		resize_handle.visible = show_handles
		if not show_handles:
			continue
		var scaled_size: Vector2 = panel.size * panel.scale
		move_handle.global_position = panel.global_position
		move_handle.size = Vector2(scaled_size.x, HUD_HANDLE_HEIGHT)
		resize_handle.global_position = panel.global_position + scaled_size - Vector2.ONE * HUD_RESIZE_HANDLE_SIZE
		resize_handle.size = Vector2.ONE * HUD_RESIZE_HANDLE_SIZE


func _get_visible_hud_panel(group_id: String) -> Control:
	var group: Dictionary = _hud_groups.get(group_id, {})
	var panels: Array = group.get("panels", [])
	for panel_value: Variant in panels:
		var panel := panel_value as Control
		if panel != null and panel.visible:
			return panel
	return null


func _on_hud_move_handle_input(event: InputEvent, group_id: String) -> void:
	if event is InputEventMouseButton:
		var mouse_button := event as InputEventMouseButton
		if mouse_button.button_index != MOUSE_BUTTON_LEFT:
			return
		if mouse_button.pressed:
			_hud_drag_group = group_id
			_hud_scale_group = ""
			_hud_pointer_start = get_viewport().get_mouse_position()
			var group: Dictionary = _hud_groups[group_id]
			_hud_position_start = group["position"] as Vector2
		else:
			_hud_drag_group = ""
			_save_hud_layout()
		get_viewport().set_input_as_handled()
	elif event is InputEventMouseMotion and _hud_drag_group == group_id:
		var pointer_delta: Vector2 = get_viewport().get_mouse_position() - _hud_pointer_start
		_set_hud_group_layout(group_id, _hud_position_start + pointer_delta, -1.0)
		get_viewport().set_input_as_handled()


func _on_hud_resize_handle_input(event: InputEvent, group_id: String) -> void:
	if event is InputEventMouseButton:
		var mouse_button := event as InputEventMouseButton
		if mouse_button.button_index != MOUSE_BUTTON_LEFT:
			return
		if mouse_button.pressed:
			_hud_scale_group = group_id
			_hud_drag_group = ""
			_hud_pointer_start = get_viewport().get_mouse_position()
			var group: Dictionary = _hud_groups[group_id]
			_hud_position_start = group["position"] as Vector2
			_hud_scale_start = float(group["scale"])
		else:
			_hud_scale_group = ""
			_save_hud_layout()
		get_viewport().set_input_as_handled()
	elif event is InputEventMouseMotion and _hud_scale_group == group_id:
		var pointer_delta: Vector2 = get_viewport().get_mouse_position() - _hud_pointer_start
		var scale_delta: float = (pointer_delta.x + pointer_delta.y) / 600.0
		_set_hud_group_layout(group_id, _hud_position_start, _hud_scale_start + scale_delta)
		get_viewport().set_input_as_handled()


func _set_hud_group_layout(group_id: String, requested_position: Vector2, requested_scale: float) -> void:
	var group: Dictionary = _hud_groups[group_id]
	var layout_scale: float = float(group["scale"])
	if requested_scale >= 0.0:
		layout_scale = clampf(requested_scale, HUD_MINIMUM_SCALE, HUD_MAXIMUM_SCALE)
	var maximum_size := Vector2.ZERO
	var panels: Array = group["panels"]
	for panel_value: Variant in panels:
		var panel := panel_value as Control
		maximum_size.x = maxf(maximum_size.x, panel.size.x * layout_scale)
		maximum_size.y = maxf(maximum_size.y, panel.size.y * layout_scale)
	var viewport_size: Vector2 = get_viewport().get_visible_rect().size
	var clamped_position := Vector2(
		clampf(requested_position.x, 0.0, maxf(0.0, viewport_size.x - maximum_size.x)),
		clampf(requested_position.y, 0.0, maxf(0.0, viewport_size.y - maximum_size.y)),
	)
	group["position"] = clamped_position
	group["scale"] = layout_scale
	_hud_groups[group_id] = group
	for panel_value: Variant in panels:
		var panel := panel_value as Control
		panel.scale = Vector2.ONE * layout_scale
		panel.global_position = clamped_position
	_refresh_hud_edit_handles()


func _load_hud_layout() -> void:
	var config := ConfigFile.new()
	if config.load(HUD_LAYOUT_PATH) != OK:
		return
	var viewport_size: Vector2 = get_viewport().get_visible_rect().size
	for group_id_value: Variant in _hud_groups.keys():
		var group_id: String = str(group_id_value)
		var group: Dictionary = _hud_groups[group_id]
		var default_ratio: Vector2 = group["default_position_ratio"] as Vector2
		var ratio_value: Variant = config.get_value(group_id, "position_ratio", default_ratio)
		var position_ratio: Vector2 = ratio_value if ratio_value is Vector2 else default_ratio
		var layout_scale: float = float(config.get_value(group_id, "scale", group["default_scale"]))
		_set_hud_group_layout(group_id, position_ratio * viewport_size, layout_scale)


func _save_hud_layout() -> void:
	if _hud_groups.is_empty():
		return
	var config := ConfigFile.new()
	var viewport_size: Vector2 = get_viewport().get_visible_rect().size
	for group_id_value: Variant in _hud_groups.keys():
		var group_id: String = str(group_id_value)
		var group: Dictionary = _hud_groups[group_id]
		var position: Vector2 = group["position"] as Vector2
		var position_ratio := Vector2(
			position.x / maxf(1.0, viewport_size.x),
			position.y / maxf(1.0, viewport_size.y),
		)
		config.set_value(group_id, "position_ratio", position_ratio)
		config.set_value(group_id, "scale", float(group["scale"]))
	var save_error: Error = config.save(HUD_LAYOUT_PATH)
	if save_error != OK:
		push_warning("Could not save HUD layout: %s" % error_string(save_error))


func _reset_hud_layout() -> void:
	if _hud_groups.is_empty():
		return
	var viewport_size: Vector2 = get_viewport().get_visible_rect().size
	for group_id_value: Variant in _hud_groups.keys():
		var group_id: String = str(group_id_value)
		var group: Dictionary = _hud_groups[group_id]
		var default_ratio: Vector2 = group["default_position_ratio"] as Vector2
		_set_hud_group_layout(group_id, default_ratio * viewport_size, float(group["default_scale"]))
	_save_hud_layout()


func _clamp_all_hud_groups() -> void:
	for group_id_value: Variant in _hud_groups.keys():
		var group_id: String = str(group_id_value)
		var group: Dictionary = _hud_groups[group_id]
		_set_hud_group_layout(group_id, group["position"] as Vector2, float(group["scale"]))


func _on_hud_viewport_resized() -> void:
	_clamp_all_hud_groups()


func _show_combat_tab() -> void:
	_combat_debug_hud_visible = true
	combat_hud.visible = true
	inventory_hud.visible = false
	skills_hud.visible = false
	equipment_hud.visible = false
	magic_hud.visible = false
	prayer_hud.visible = false
	styles_hud.visible = false
	settings_hud.visible = false
	combat_tab_button.set_pressed_no_signal(true)
	inventory_tab_button.set_pressed_no_signal(false)
	combat_skills_tab_button.set_pressed_no_signal(false)
	combat_equipment_tab_button.set_pressed_no_signal(false)
	combat_magic_tab_button.set_pressed_no_signal(false)
	combat_styles_tab_button.set_pressed_no_signal(false)
	_refresh_hud_edit_handles()


func _show_inventory_tab() -> void:
	_combat_debug_hud_visible = false
	combat_hud.visible = false
	inventory_hud.visible = true
	skills_hud.visible = false
	equipment_hud.visible = false
	magic_hud.visible = false
	prayer_hud.visible = false
	styles_hud.visible = false
	settings_hud.visible = false
	inventory_combat_tab_button.set_pressed_no_signal(false)
	inventory_active_tab_button.set_pressed_no_signal(true)
	inventory_skills_tab_button.set_pressed_no_signal(false)
	inventory_equipment_tab_button.set_pressed_no_signal(false)
	inventory_magic_tab_button.set_pressed_no_signal(false)
	inventory_prayer_tab_button.set_pressed_no_signal(false)
	inventory_styles_tab_button.set_pressed_no_signal(false)
	_refresh_hud_edit_handles()


func _show_skills_tab() -> void:
	_combat_debug_hud_visible = false
	combat_hud.visible = false
	inventory_hud.visible = false
	skills_hud.visible = true
	equipment_hud.visible = false
	magic_hud.visible = false
	prayer_hud.visible = false
	styles_hud.visible = false
	settings_hud.visible = false
	skills_combat_tab_button.set_pressed_no_signal(false)
	skills_inventory_tab_button.set_pressed_no_signal(false)
	skills_active_tab_button.set_pressed_no_signal(true)
	skills_equipment_tab_button.set_pressed_no_signal(false)
	skills_magic_tab_button.set_pressed_no_signal(false)
	skills_prayer_tab_button.set_pressed_no_signal(false)
	skills_styles_tab_button.set_pressed_no_signal(false)
	_refresh_hud_edit_handles()


func _show_equipment_tab() -> void:
	_combat_debug_hud_visible = false
	combat_hud.visible = false
	inventory_hud.visible = false
	skills_hud.visible = false
	equipment_hud.visible = true
	magic_hud.visible = false
	prayer_hud.visible = false
	styles_hud.visible = false
	settings_hud.visible = false
	equipment_combat_tab_button.set_pressed_no_signal(false)
	equipment_inventory_tab_button.set_pressed_no_signal(false)
	equipment_skills_tab_button.set_pressed_no_signal(false)
	equipment_active_tab_button.set_pressed_no_signal(true)
	equipment_magic_tab_button.set_pressed_no_signal(false)
	equipment_prayer_tab_button.set_pressed_no_signal(false)
	equipment_styles_tab_button.set_pressed_no_signal(false)
	_refresh_hud_edit_handles()


func _show_magic_tab() -> void:
	_combat_debug_hud_visible = false
	combat_hud.visible = false
	inventory_hud.visible = false
	skills_hud.visible = false
	equipment_hud.visible = false
	magic_hud.visible = true
	prayer_hud.visible = false
	styles_hud.visible = false
	settings_hud.visible = false
	magic_combat_tab_button.set_pressed_no_signal(false)
	magic_inventory_tab_button.set_pressed_no_signal(false)
	magic_skills_tab_button.set_pressed_no_signal(false)
	magic_equipment_tab_button.set_pressed_no_signal(false)
	magic_active_tab_button.set_pressed_no_signal(true)
	magic_prayer_tab_button.set_pressed_no_signal(false)
	magic_styles_tab_button.set_pressed_no_signal(false)
	_refresh_hud_edit_handles()


func _show_prayer_tab() -> void:
	_combat_debug_hud_visible = false
	combat_hud.visible = false
	inventory_hud.visible = false
	skills_hud.visible = false
	equipment_hud.visible = false
	magic_hud.visible = false
	prayer_hud.visible = true
	styles_hud.visible = false
	settings_hud.visible = false
	prayer_inventory_tab_button.set_pressed_no_signal(false)
	prayer_skills_tab_button.set_pressed_no_signal(false)
	prayer_equipment_tab_button.set_pressed_no_signal(false)
	prayer_magic_tab_button.set_pressed_no_signal(false)
	prayer_active_tab_button.set_pressed_no_signal(true)
	prayer_styles_tab_button.set_pressed_no_signal(false)
	_refresh_hud_edit_handles()


func _show_styles_tab() -> void:
	_combat_debug_hud_visible = false
	combat_hud.visible = false
	inventory_hud.visible = false
	skills_hud.visible = false
	equipment_hud.visible = false
	magic_hud.visible = false
	prayer_hud.visible = false
	styles_hud.visible = true
	settings_hud.visible = false
	styles_combat_tab_button.set_pressed_no_signal(false)
	styles_inventory_tab_button.set_pressed_no_signal(false)
	styles_skills_tab_button.set_pressed_no_signal(false)
	styles_equipment_tab_button.set_pressed_no_signal(false)
	styles_magic_tab_button.set_pressed_no_signal(false)
	styles_prayer_tab_button.set_pressed_no_signal(false)
	styles_active_tab_button.set_pressed_no_signal(true)
	_refresh_hud_edit_handles()


func _show_settings_tab() -> void:
	if bank_hud.visible:
		_close_bank()
	_combat_debug_hud_visible = false
	combat_hud.visible = false
	inventory_hud.visible = false
	skills_hud.visible = false
	equipment_hud.visible = false
	magic_hud.visible = false
	prayer_hud.visible = false
	styles_hud.visible = false
	settings_hud.visible = true
	_refresh_hud_edit_handles()


func _populate_placeholder_skills() -> void:
	for skill_data: Variant in PLACEHOLDER_SKILLS:
		var skill: Dictionary = skill_data
		var skill_name: String = str(skill.get("name", "Skill"))
		var skill_tile := Button.new()
		skill_tile.custom_minimum_size = Vector2(124, 108)
		skill_tile.size_flags_horizontal = Control.SIZE_EXPAND_FILL
		skill_tile.focus_mode = Control.FOCUS_NONE
		skill_tile.icon = SKILL_ICONS.get(skill_name, TAB_SKILLS_ICON) as Texture2D
		skill_tile.expand_icon = true
		skill_tile.icon_alignment = HORIZONTAL_ALIGNMENT_CENTER
		skill_tile.vertical_icon_alignment = VERTICAL_ALIGNMENT_CENTER
		skill_tile.add_theme_constant_override("icon_max_width", 58)
		var level_label := Label.new()
		var starting_level: int = int(skill.get("level", 1))
		level_label.text = "%d/%d" % [starting_level, starting_level]
		level_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
		level_label.set_anchors_preset(Control.PRESET_BOTTOM_WIDE)
		level_label.offset_top = -26.0
		level_label.offset_bottom = -5.0
		level_label.mouse_filter = Control.MOUSE_FILTER_IGNORE
		skill_tile.add_child(level_label)
		var skill_id: String = str(ACTIVE_SKILL_IDS.get(skill_name, ""))
		if not skill_id.is_empty():
			_skill_tiles[skill_id] = {
				"tile": skill_tile,
				"level_label": level_label,
			}
			if skill_id == "herblore":
				skill_tile.pressed.connect(_open_brewing_window)
			elif skill_id == "fletching":
				skill_tile.pressed.connect(_open_fletching_window)
		else:
			skill_tile.tooltip_text = "Current XP: 0\nNext level XP: unavailable"
		skills_grid.add_child(skill_tile)
	_update_active_skills_hud()


func _apply_skills_snapshot(skills: Array) -> void:
	var changed: bool = false
	for payload: Variant in skills:
		if typeof(payload) != TYPE_DICTIONARY:
			continue
		var skill: Dictionary = payload
		var skill_id: String = str(skill.get("id", ""))
		if not _skill_progress.has(skill_id):
			continue
		var current_progress: Dictionary = _skill_progress[skill_id] as Dictionary
		var experience: int = int(skill.get("experience", current_progress.get("experience", 0)))
		var level: int = int(skill.get("level", current_progress.get("level", 1)))
		var current_level_experience: int = int(skill.get(
			"current_level_experience",
			current_progress.get("current", 0),
		))
		var next_level_experience: int = int(skill.get(
			"next_level_experience",
			current_progress.get("next", 100),
		))
		if experience == int(current_progress.get("experience", 0)) and \
			level == int(current_progress.get("level", 1)) and \
			current_level_experience == int(current_progress.get("current", 0)) and \
			next_level_experience == int(current_progress.get("next", 100)):
			continue
		_skill_progress[skill_id] = {
			"experience": experience,
			"level": level,
			"current": current_level_experience,
			"next": next_level_experience,
		}
		changed = true
	if changed:
		_update_active_skills_hud()
		_update_magic_hud()


func _update_active_skills_hud() -> void:
	for skill_id: String in ACTIVE_SKILL_NAMES:
		_update_progression_skill_tile(skill_id)


func _update_progression_skill_tile(skill_id: String) -> void:
	var skill_tile: Dictionary = _skill_tiles.get(skill_id, {}) as Dictionary
	if skill_tile.is_empty():
		return
	var tile := skill_tile["tile"] as Button
	var level_label := skill_tile["level_label"] as Label
	var progress: Dictionary = _skill_progress[skill_id] as Dictionary
	var experience: int = int(progress.get("experience", 0))
	var level: int = int(progress.get("level", 1))
	var next_level_experience: int = int(progress.get("next", 100))
	level_label.text = "%d/%d" % [level, level]
	if next_level_experience <= 0:
		tile.tooltip_text = "Current XP: %d\nNext level XP: maximum" % experience
		return
	tile.tooltip_text = "Current XP: %d\nNext level XP: %d" % [experience, next_level_experience]


func _populate_placeholder_spells() -> void:
	_update_magic_hud()


func _update_magic_hud() -> void:
	for child: Node in magic_grid.get_children():
		magic_grid.remove_child(child)
		child.queue_free()
	var magic_progress: Dictionary = _skill_progress.get("magic", {}) as Dictionary
	var magic_level: int = int(magic_progress.get("level", 1))
	for spell_data: Variant in PLACEHOLDER_SPELLS:
		var spell: Dictionary = spell_data
		var spell_id: String = str(spell.get("id", ""))
		var required_level: int = int(spell.get("required_magic_level", 1))
		var required_book: String = str(spell.get("book_id", ""))
		var spell_button := Button.new()
		spell_button.custom_minimum_size = Vector2(104, 88)
		spell_button.size_flags_horizontal = Control.SIZE_EXPAND_FILL
		spell_button.focus_mode = Control.FOCUS_NONE
		spell_button.toggle_mode = not spell_id.is_empty()
		spell_button.disabled = spell_id.is_empty() or _player_defeated or magic_level < required_level or (
			not required_book.is_empty() and required_book != _active_spellbook
		)
		spell_button.icon = spell.get("icon") as Texture2D
		spell_button.expand_icon = true
		if spell_id.is_empty():
			spell_button.tooltip_text = "%s (coming later)" % spell.get("name", "Spell")
		else:
			var displayed_spell: String = _queued_spell if _spell_selection_pending else _selected_spell
			var displayed_selected: bool = spell_id == displayed_spell
			spell_button.set_pressed_no_signal(displayed_selected)
			var cost_item_id: String = str(spell.get("cost_item_id", ""))
			var cost_quantity: int = int(spell.get("cost_quantity", 0))
			spell_button.tooltip_text = "%s • Magic level %d • costs %d %s per cast • you have %d • click again to stop autocasting" % [
				spell.get("name", "Spell"), required_level, cost_quantity,
				cost_item_id.capitalize(), _inventory_quantity(cost_item_id),
			]
			if not required_book.is_empty() and required_book != _active_spellbook:
				spell_button.tooltip_text += " • requires the %s spellbook" % required_book.capitalize()
			spell_button.pressed.connect(_request_spell.bind(spell_id))
		magic_grid.add_child(spell_button)


func _populate_placeholder_prayers() -> void:
	_update_prayer_hud()


func _update_prayer_hud() -> void:
	prayer_set_label.text = "Prayer set: %s" % _active_prayer_set.replace("_", " ").capitalize()
	for child: Node in prayer_grid.get_children():
		prayer_grid.remove_child(child)
		child.queue_free()
	for prayer_data: Variant in PLACEHOLDER_PRAYERS:
		var prayer: Dictionary = prayer_data
		var prayer_id: String = str(prayer.get("id", ""))
		var prayer_button := Button.new()
		prayer_button.custom_minimum_size = Vector2(104, 88)
		prayer_button.size_flags_horizontal = Control.SIZE_EXPAND_FILL
		prayer_button.focus_mode = Control.FOCUS_NONE
		prayer_button.toggle_mode = not prayer_id.is_empty()
		prayer_button.disabled = prayer_id.is_empty() or _player_defeated or (
			_player_prayer <= 0 and not _active_prayers.has(prayer_id)
		)
		prayer_button.icon = IconCatalog.centered_icon(prayer.get("icon") as Texture2D)
		prayer_button.expand_icon = true
		prayer_button.icon_alignment = HORIZONTAL_ALIGNMENT_CENTER
		prayer_button.vertical_icon_alignment = VERTICAL_ALIGNMENT_CENTER
		if not prayer_id.is_empty():
			var displayed_active: bool = _active_prayers.has(prayer_id)
			if _queued_prayer_id == prayer_id:
				displayed_active = _queued_prayer_active
			prayer_button.set_pressed_no_signal(displayed_active)
			prayer_button.tooltip_text = "%s • %s\n%s" % [
				prayer.get("name", "Prayer"),
				"ACTIVE • click to turn off" if displayed_active else "Inactive • click to turn on",
				prayer.get("description", ""),
			]
			if displayed_active:
				_style_active_prayer_button(prayer_button, _queued_prayer_id == prayer_id)
			prayer_button.pressed.connect(_request_prayer.bind(prayer_id))
		else:
			prayer_button.tooltip_text = "%s (coming later)" % prayer.get("name", "Prayer")
		prayer_grid.add_child(prayer_button)


func _style_active_prayer_button(prayer_button: Button, pending: bool) -> void:
	var active_color := Color(0.98, 0.78, 0.2) if not pending else Color(0.35, 0.78, 1.0)
	var active_style := StyleBoxFlat.new()
	active_style.bg_color = Color(0.16, 0.21, 0.08, 0.98) if not pending else Color(0.07, 0.16, 0.23, 0.98)
	active_style.border_width_left = 4
	active_style.border_width_top = 4
	active_style.border_width_right = 4
	active_style.border_width_bottom = 4
	active_style.border_color = active_color
	active_style.corner_radius_top_left = 7
	active_style.corner_radius_top_right = 7
	active_style.corner_radius_bottom_left = 7
	active_style.corner_radius_bottom_right = 7
	prayer_button.add_theme_stylebox_override("normal", active_style)
	prayer_button.add_theme_stylebox_override("pressed", active_style)
	prayer_button.add_theme_stylebox_override("hover_pressed", active_style)
	prayer_button.add_theme_stylebox_override("disabled", active_style)

	var active_badge := Label.new()
	active_badge.mouse_filter = Control.MOUSE_FILTER_IGNORE
	active_badge.set_anchors_and_offsets_preset(Control.PRESET_BOTTOM_WIDE)
	active_badge.offset_left = 5.0
	active_badge.offset_top = -25.0
	active_badge.offset_right = -5.0
	active_badge.offset_bottom = -4.0
	active_badge.text = "ACTIVATING…" if pending else "● ACTIVE"
	active_badge.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	active_badge.vertical_alignment = VERTICAL_ALIGNMENT_CENTER
	active_badge.add_theme_color_override("font_color", active_color)
	active_badge.add_theme_color_override("font_outline_color", Color(0.02, 0.03, 0.04, 0.95))
	active_badge.add_theme_constant_override("outline_size", 5)
	active_badge.add_theme_font_size_override("font_size", 13)
	prayer_button.add_child(active_badge)


func _update_combat_styles_hud() -> void:
	for child: Node in styles_grid.get_children():
		styles_grid.remove_child(child)
		child.queue_free()

	var style_options: Array = COMBAT_STYLES.get(_equipped_weapon, []) as Array
	var selected_style: String = _queued_combat_style if not _queued_combat_style.is_empty() else _combat_style
	for style_data: Variant in style_options:
		var style: Dictionary = style_data
		var style_id: String = str(style.get("id", ""))
		var training: String = str(style.get("training", ""))
		var style_button := Button.new()
		style_button.custom_minimum_size = Vector2(268, 160)
		style_button.focus_mode = Control.FOCUS_NONE
		style_button.toggle_mode = true
		style_button.disabled = _player_defeated
		style_button.add_theme_font_size_override("font_size", 22)
		style_button.add_theme_color_override("font_color", Color(0.82, 0.86, 0.9))
		style_button.add_theme_color_override("font_hover_color", Color(0.96, 0.98, 1.0))
		style_button.add_theme_color_override("font_pressed_color", Color(1.0, 1.0, 1.0))
		var normal_style := StyleBoxFlat.new()
		normal_style.bg_color = Color(0.055, 0.07, 0.085, 0.96)
		normal_style.border_width_left = 2
		normal_style.border_width_top = 2
		normal_style.border_width_right = 2
		normal_style.border_width_bottom = 2
		normal_style.border_color = Color(0.25, 0.31, 0.37, 0.95)
		normal_style.corner_radius_top_left = 8
		normal_style.corner_radius_top_right = 8
		normal_style.corner_radius_bottom_left = 8
		normal_style.corner_radius_bottom_right = 8
		var hover_style := normal_style.duplicate() as StyleBoxFlat
		hover_style.bg_color = Color(0.09, 0.13, 0.17, 0.98)
		hover_style.border_color = Color(0.42, 0.55, 0.66, 1.0)
		var selected_style_box := normal_style.duplicate() as StyleBoxFlat
		selected_style_box.bg_color = Color(0.08, 0.3, 0.46, 0.98)
		selected_style_box.border_width_left = 3
		selected_style_box.border_width_top = 3
		selected_style_box.border_width_right = 3
		selected_style_box.border_width_bottom = 3
		selected_style_box.border_color = Color(0.28, 0.78, 1.0, 1.0)
		style_button.add_theme_stylebox_override("normal", normal_style)
		style_button.add_theme_stylebox_override("hover", hover_style)
		style_button.add_theme_stylebox_override("pressed", selected_style_box)
		style_button.add_theme_stylebox_override("hover_pressed", selected_style_box)
		style_button.text = "%s\n%s" % [
			style.get("name", style_id.capitalize()),
			"No XP" if training.is_empty() else "%s XP" % training,
		]
		style_button.tooltip_text = "Combat XP is locked" if training.is_empty() else "Successful hits train %s" % training
		var accuracy_note: String = ""
		match style_id:
			"accurate":
				accuracy_note = "Improves accuracy"
			"stab", "crush":
				accuracy_note = "Small accuracy bonus"
			"rapid", "heavy_crush":
				accuracy_note = "Slightly lowers accuracy"
			"defensive":
				accuracy_note = "Slightly lowers accuracy while training Defence"
		if not accuracy_note.is_empty():
			style_button.tooltip_text += "\n%s" % accuracy_note
		style_button.set_pressed_no_signal(style_id == selected_style)
		style_button.pressed.connect(_request_combat_style.bind(style_id))
		styles_grid.add_child(style_button)


func _update_equipment_hud() -> void:
	for child: Node in equipment_grid.get_children():
		equipment_grid.remove_child(child)
		child.queue_free()

	# Three columns form a simple paper-doll silhouette. Spacer cells keep the
	# head, legs, and feet aligned down the centre of the body.
	var equipment_slots: Array = [
		{"spacer": true},
		{"slot": "head", "name": "Head", "item": str(_equipped_armour.get("head", "")), "instance": str(_equipped_armour_instances.get("head", "")), "reinforcement": int(_equipped_armour_reinforcements.get("head", 0))},
		{"spacer": true},
		{"slot": "cape", "name": "Cape", "item": ""},
		{"slot": "neck", "name": "Neck", "item": str(_equipped_armour.get("neck", "")), "instance": str(_equipped_armour_instances.get("neck", "")), "reinforcement": int(_equipped_armour_reinforcements.get("neck", 0))},
		{"slot": "spellbook", "name": "Spellbook", "item": _equipped_spellbook},
		{"slot": "main_hand", "name": "Main Hand", "item": "" if _equipped_weapon == "unarmed" else _equipped_weapon, "instance": _equipped_weapon_instance_id, "reinforcement": _equipped_weapon_reinforcement},
		{"slot": "body", "name": "Body", "item": str(_equipped_armour.get("body", "")), "instance": str(_equipped_armour_instances.get("body", "")), "reinforcement": int(_equipped_armour_reinforcements.get("body", 0))},
		{"slot": "left_hand", "name": "Left Hand", "item": str(_equipped_armour.get("left_hand", "")), "instance": str(_equipped_armour_instances.get("left_hand", "")), "reinforcement": int(_equipped_armour_reinforcements.get("left_hand", 0))},
		{"slot": "hands", "name": "Hands", "item": str(_equipped_armour.get("hands", "")), "instance": str(_equipped_armour_instances.get("hands", "")), "reinforcement": int(_equipped_armour_reinforcements.get("hands", 0))},
		{"slot": "legs", "name": "Legs", "item": str(_equipped_armour.get("legs", "")), "instance": str(_equipped_armour_instances.get("legs", "")), "reinforcement": int(_equipped_armour_reinforcements.get("legs", 0))},
		{"slot": "ring", "name": "Ring", "item": ""},
		{"slot": "ammunition", "name": "Ammunition Pouch", "item": ""},
		{"slot": "feet", "name": "Feet", "item": str(_equipped_armour.get("feet", "")), "instance": str(_equipped_armour_instances.get("feet", "")), "reinforcement": int(_equipped_armour_reinforcements.get("feet", 0))},
		{"spacer": true},
	]
	for slot_data: Variant in equipment_slots:
		var slot: Dictionary = slot_data
		if bool(slot.get("spacer", false)):
			var spacer := Control.new()
			spacer.custom_minimum_size = Vector2(168, 82)
			spacer.mouse_filter = Control.MOUSE_FILTER_IGNORE
			equipment_grid.add_child(spacer)
			continue
		var item_id: String = str(slot.get("item", ""))
		var instance_id: String = str(slot.get("instance", ""))
		var reinforcement: int = int(slot.get("reinforcement", 0))
		var slot_id: String = str(slot.get("slot", ""))
		var slot_panel := Button.new()
		slot_panel.custom_minimum_size = Vector2(168, 82)
		slot_panel.focus_mode = Control.FOCUS_NONE
		var unequip_pending: bool = (
			not item_id.is_empty()
			and _queued_unequip
			and _queued_item == item_id
			and (_queued_instance_id.is_empty() or _queued_instance_id == instance_id)
		)
		slot_panel.tooltip_text = "%s: %s%s" % [
			slot.get("name", "Equipment"),
			item_id.capitalize() if not item_id.is_empty() else "No item equipped",
			" +%d" % reinforcement if reinforcement > 0 else "",
		]
		if slot_id == "ammunition":
			slot_panel.tooltip_text = "Ammunition Pouch • %d/%d types stored • left-click to open" % [
				_ammunition_pouch_items.size(), _ammunition_pouch_unlocked,
			]
		elif ITEM_COMBAT_BONUSES.has(item_id):
			slot_panel.tooltip_text += " • %s" % str(ITEM_COMBAT_BONUSES[item_id])
		if EQUIPMENT_REQUIREMENTS.has(item_id):
			slot_panel.tooltip_text += " • requires %s" % str(EQUIPMENT_REQUIREMENTS[item_id])
		if unequip_pending:
			slot_panel.tooltip_text += " • unequipping…"
		elif not item_id.is_empty():
			slot_panel.tooltip_text += " • left-click to unequip"
		var slot_style := StyleBoxFlat.new()
		slot_style.bg_color = Color(0.055, 0.07, 0.085, 0.94)
		slot_style.border_width_left = 2
		slot_style.border_width_top = 2
		slot_style.border_width_right = 2
		slot_style.border_width_bottom = 2
		slot_style.border_color = Color(0.28, 0.34, 0.4, 0.95)
		if not item_id.is_empty():
			slot_style.border_width_left = 3
			slot_style.border_width_top = 3
			slot_style.border_width_right = 3
			slot_style.border_width_bottom = 3
			slot_style.border_color = Color(0.95, 0.72, 0.2)
		slot_style.corner_radius_top_left = 7
		slot_style.corner_radius_top_right = 7
		slot_style.corner_radius_bottom_left = 7
		slot_style.corner_radius_bottom_right = 7
		slot_panel.add_theme_stylebox_override("normal", slot_style)
		var hover_style := slot_style.duplicate() as StyleBoxFlat
		hover_style.bg_color = Color(0.09, 0.13, 0.17, 0.98)
		hover_style.border_color = Color(0.42, 0.55, 0.66, 1.0)
		slot_panel.add_theme_stylebox_override("hover", hover_style)
		var pressed_style := hover_style.duplicate() as StyleBoxFlat
		pressed_style.bg_color = Color(0.08, 0.3, 0.46, 0.98)
		slot_panel.add_theme_stylebox_override("pressed", pressed_style)
		if unequip_pending:
			var pending_style := slot_style.duplicate() as StyleBoxFlat
			pending_style.border_color = Color(0.35, 0.7, 1.0)
			slot_panel.add_theme_stylebox_override("disabled", pending_style)
		slot_panel.icon = (
			ITEM_ICONS.get(item_id) as Texture2D
			if not item_id.is_empty()
			else EQUIPMENT_SLOT_ICONS.get(slot_id, EMPTY_SLOT_ICON) as Texture2D
		)
		slot_panel.expand_icon = true
		slot_panel.icon_alignment = HORIZONTAL_ALIGNMENT_CENTER
		slot_panel.add_theme_constant_override("icon_max_width", 56)
		if slot_id == "ammunition":
			slot_panel.disabled = false
			slot_panel.pressed.connect(_toggle_ammunition_pouch)
		elif not item_id.is_empty():
			slot_panel.disabled = unequip_pending or _player_defeated
			slot_panel.pressed.connect(_request_unequip.bind(item_id, instance_id))
		else:
			slot_panel.disabled = true
			slot_panel.add_theme_stylebox_override("disabled", slot_style)
			slot_panel.add_theme_color_override(
				"icon_disabled_color",
				Color(0.42, 0.47, 0.52, 0.38),
			)
		equipment_grid.add_child(slot_panel)


func _show_detailed_combat_stats() -> void:
	_update_detailed_combat_stats()
	combat_stats_hud.visible = true
	combat_stats_close_button.grab_focus()


func _hide_detailed_combat_stats() -> void:
	combat_stats_hud.visible = false
	detailed_stats_button.grab_focus()


func _show_combat_stats_page(page: String) -> void:
	_combat_stats_page = page
	combat_stats_offence_button.set_pressed_no_signal(page == "offence")
	combat_stats_defence_button.set_pressed_no_signal(page == "defence")
	combat_stats_effects_button.set_pressed_no_signal(page == "effects")
	_update_detailed_combat_stats()


func _update_detailed_combat_stats() -> void:
	if _combat_stats.is_empty():
		combat_stats_label.text = "Waiting for the server to send combat stats…"
		return
	match _combat_stats_page:
		"defence":
			combat_stats_label.text = _detailed_defence_text()
		"effects":
			combat_stats_label.text = _detailed_effects_text()
		_:
			combat_stats_label.text = _detailed_offence_text()


func _detailed_offence_text() -> String:
	var spell_id: String = str(_combat_stats.get("spell", ""))
	var loadout_text: String = "%s  •  %s style" % [
		str(_combat_stats.get("weapon", "unarmed")).capitalize(),
		str(_combat_stats.get("style", "crush")).capitalize(),
	]
	var accuracy_breakdown: String = "Weapon accuracy: %s\nStyle modifier: %s" % [
		_signed_stat(int(_combat_stats.get("weapon_accuracy_bonus", 0))),
		_signed_stat(int(_combat_stats.get("style_accuracy_bonus", 0))),
	]
	accuracy_breakdown += "\nAccessory accuracy: %s" % _signed_stat(
		int(_combat_stats.get("equipment_accuracy_bonus", 0)),
	)
	var prayer_accuracy_bonus: int = int(_combat_stats.get("prayer_accuracy_bonus", 0))
	if prayer_accuracy_bonus != 0:
		accuracy_breakdown += "\nPrayer accuracy: %s" % _signed_stat(prayer_accuracy_bonus)
	var damage_breakdown: String = "Weapon strength: %s" % _signed_stat(
		int(_combat_stats.get("strength_bonus", 0)),
	)
	damage_breakdown += "\nAccessory strength: %s" % _signed_stat(
		int(_combat_stats.get("equipment_strength_bonus", 0)),
	)
	if not spell_id.is_empty():
		loadout_text = "%s  •  %s equipped" % [
			spell_id.capitalize(), str(_combat_stats.get("weapon", "unarmed")).capitalize(),
		]
		accuracy_breakdown = "Spell accuracy: %s\nWeapon focus accuracy: %s\nAccessory magic accuracy: %s" % [
			_signed_stat(int(_combat_stats.get("spell_accuracy_bonus", 0))),
			_signed_stat(int(_combat_stats.get("weapon_accuracy_bonus", 0))),
			_signed_stat(int(_combat_stats.get("equipment_accuracy_bonus", 0))),
		]
		damage_breakdown = "Spell power: %s\nWeapon focus strength: %s\nAccessory magic strength: %s" % [
			_signed_stat(int(_combat_stats.get("spell_power", 0))),
			_signed_stat(int(_combat_stats.get("strength_bonus", 0))),
			_signed_stat(int(_combat_stats.get("equipment_strength_bonus", 0))),
		]
	var prayer_damage_bonus: float = float(_combat_stats.get("prayer_damage_bonus", 0.0))
	if prayer_damage_bonus > 0.0 and spell_id.is_empty():
		damage_breakdown += "\nPrayer damage: +%.0f%%" % (prayer_damage_bonus * 100.0)
	var attack_interval: int = int(_combat_stats.get("attack_interval", 0))
	var base_attack_interval: int = int(_combat_stats.get("base_attack_interval", attack_interval))
	var attack_speed_text: String = "%d pulses (%.1f seconds)" % [
		attack_interval, attack_interval * 0.6,
	]
	if attack_interval < base_attack_interval:
		attack_speed_text += "  •  equipment modifier active"
	return (
		"CURRENT LOADOUT\n%s\n\n" % loadout_text
		+"ACCURACY\nTotal accuracy score: %d\n" % int(_combat_stats.get("total_accuracy", 0))
		+"%s level: %d\n%s\n\n" % [
			str(_combat_stats.get("accuracy_skill", "attack")).capitalize(),
			int(_combat_stats.get("accuracy_level", 1)), accuracy_breakdown,
		]
		+"DAMAGE\nCurrent damage range: %d–%d\n" % [
			int(_combat_stats.get("minimum_damage", 0)),
			int(_combat_stats.get("maximum_damage", 0)),
		]
		+"%s level: %d\n%s\nSkill damage: +%.1f\n\n" % [
			str(_combat_stats.get("damage_skill", "strength")).capitalize(),
			int(_combat_stats.get("damage_level", 1)), damage_breakdown,
			float(_combat_stats.get("skill_damage_bonus", 0.0)),
		]
		+"ATTACK SPEED\n%s" % attack_speed_text
	)


func _detailed_defence_text() -> String:
	var defence_data: Variant = _combat_stats.get("defence_bonuses", {})
	var defence: Dictionary = defence_data if typeof(defence_data) == TYPE_DICTIONARY else {}
	return (
		"DEFENCE OVERVIEW\n"
		+"Defence level: %d\n" % int(_combat_stats.get("defence_level", 1))
		+"Prayer damage reduction: %.0f%%\n" % (
			float(_combat_stats.get("prayer_damage_reduction", 0.0)) * 100.0
		)
		+"Armour reduction vs crush: %.1f%%\n" % (
			float(_combat_stats.get("armour_damage_reduction", 0.0)) * 100.0
		)
		+"Maximum health from armour: +%d\n" % int(_combat_stats.get("maximum_health_bonus", 0))
		+"\nEQUIPMENT DEFENCE BONUSES\n"
		+"Stab: %s\nSlash: %s\nCrush: %s\n" % [
			_signed_stat(int(defence.get("stab", 0))),
			_signed_stat(int(defence.get("slash", 0))),
			_signed_stat(int(defence.get("crush", 0))),
		]
		+"Ranged: %s\nMagic: %s\n\n" % [
			_signed_stat(int(defence.get("ranged", 0))),
			_signed_stat(int(defence.get("magic", 0))),
		]
		+"Matching defence uses diminishing returns for damage reduction."
	)


func _detailed_effects_text() -> String:
	var swift_pieces: int = 0
	for item_id: String in _equipped_armour.values():
		if item_id in ["swift_helm", "swift_body", "swift_boots"]:
			swift_pieces += 1
	var interval_modifier: int = int(_combat_stats.get("attack_interval_modifier", 0))
	var swift_status: String = "Active" if swift_pieces >= 3 else "Inactive"
	var spellbook_text: String = _equipped_spellbook.capitalize() if not _equipped_spellbook.is_empty() else "None"
	return (
		"EQUIPPED ITEMS\n"
		+"Main hand: %s\n" % _equipped_weapon.capitalize()
		+"Spellbook: %s\n" % spellbook_text
		+"Neck: %s\n" % _equipment_item_name("neck")
		+"Head: %s\n" % _equipment_item_name("head")
		+"Body: %s\n" % _equipment_item_name("body")
		+"Left hand: %s\n" % _equipment_item_name("left_hand")
		+"Hands: %s\n" % _equipment_item_name("hands")
		+"Legs: %s\n" % _equipment_item_name("legs")
		+"Feet: %s\n\n" % _equipment_item_name("feet")
		+"SET EFFECTS\n"
		+"Swift set: %s (%d/3 pieces)\n" % [swift_status, swift_pieces]
		+"Three pieces reduce the attack interval by 1 pulse.\n"
		+"Current equipment interval modifier: %s pulses\n\n" % _signed_stat(interval_modifier)
		+"Future equipment sets will be listed on this page."
	)


func _equipment_item_name(slot: String) -> String:
	var item_id: String = str(_equipped_armour.get(slot, ""))
	return item_id.capitalize() if not item_id.is_empty() else "None"


func _signed_stat(value: int) -> String:
	return "+%d" % value if value >= 0 else str(value)


func _apply_inventory_snapshot(inventory: Dictionary) -> void:
	var capacity: int = maxi(0, int(inventory.get("capacity", _inventory_capacity)))
	var items_data: Variant = inventory.get("items", [])
	if typeof(items_data) != TYPE_ARRAY:
		return
	var signature := "%d:%s" % [capacity, JSON.stringify(items_data)]
	if signature == _inventory_signature:
		return
	_inventory_signature = signature
	_inventory_capacity = capacity
	_inventory_items = (items_data as Array).duplicate(true)
	if brewing_window.visible:
		_refresh_brewing_window()
	_update_inventory_hud()
	_update_magic_hud()


func _apply_bank_snapshot(bank: Dictionary) -> void:
	var capacity: int = maxi(0, int(bank.get("capacity", _bank_capacity)))
	var items_data: Variant = bank.get("items", [])
	if typeof(items_data) != TYPE_ARRAY:
		return
	var signature := "%d:%s" % [capacity, JSON.stringify(items_data)]
	if signature == _bank_signature:
		return
	_bank_signature = signature
	_bank_capacity = capacity
	_bank_items = (items_data as Array).duplicate(true)
	_update_bank_hud()


func _apply_ammunition_pouch_snapshot(pouch: Dictionary) -> void:
	var capacity: int = int(pouch.get("capacity", 10))
	var unlocked: int = int(pouch.get("unlocked_slots", 2))
	var items_data: Variant = pouch.get("items", [])
	if typeof(items_data) != TYPE_ARRAY:
		return
	var signature := "%d:%d:%s" % [capacity, unlocked, JSON.stringify(items_data)]
	if signature == _ammunition_pouch_signature:
		return
	_ammunition_pouch_signature = signature
	_ammunition_pouch_capacity = capacity
	_ammunition_pouch_unlocked = unlocked
	_ammunition_pouch_items = (items_data as Array).duplicate(true)
	_update_ammunition_pouch_window()
	_update_equipment_hud()


func _setup_ammunition_pouch_window() -> void:
	_ammunition_pouch_window = Window.new()
	_ammunition_pouch_window.title = "Ammunition Pouch"
	_ammunition_pouch_window.min_size = Vector2i(570, 300)
	_ammunition_pouch_window.size = Vector2i(570, 300)
	_ammunition_pouch_window.visible = false
	_ammunition_pouch_window.close_requested.connect(_close_ammunition_pouch)
	add_child(_ammunition_pouch_window)
	var margin := MarginContainer.new()
	margin.add_theme_constant_override("margin_left", 18)
	margin.add_theme_constant_override("margin_top", 18)
	margin.add_theme_constant_override("margin_right", 18)
	margin.add_theme_constant_override("margin_bottom", 18)
	_ammunition_pouch_window.add_child(margin)
	var layout := VBoxContainer.new()
	margin.add_child(layout)
	var header := HBoxContainer.new()
	layout.add_child(header)
	_ammunition_pouch_summary = Label.new()
	_ammunition_pouch_summary.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	header.add_child(_ammunition_pouch_summary)
	var close_button := Button.new()
	close_button.text = "Close"
	close_button.focus_mode = Control.FOCUS_NONE
	close_button.pressed.connect(_close_ammunition_pouch)
	header.add_child(close_button)
	_ammunition_pouch_grid = GridContainer.new()
	_ammunition_pouch_grid.columns = 5
	_ammunition_pouch_grid.add_theme_constant_override("h_separation", 8)
	_ammunition_pouch_grid.add_theme_constant_override("v_separation", 8)
	layout.add_child(_ammunition_pouch_grid)
	_update_ammunition_pouch_window()


func _toggle_ammunition_pouch() -> void:
	if _ammunition_pouch_window.visible:
		_close_ammunition_pouch()
	else:
		_ammunition_pouch_window.popup_centered()
		_update_ammunition_pouch_window()
		_update_inventory_hud()


func _close_ammunition_pouch() -> void:
	_ammunition_pouch_window.hide()
	_update_inventory_hud()


func _request_ammunition_deposit(inventory_slot: int) -> void:
	if network.send_ammunition_deposit(_next_ammunition_sequence, inventory_slot):
		_next_ammunition_sequence += 1
		status_label.text = "Moving ammunition into the pouch…"


func _request_ammunition_withdraw(pouch_slot: int) -> void:
	if network.send_ammunition_withdraw(_next_ammunition_sequence, pouch_slot):
		_next_ammunition_sequence += 1
		status_label.text = "Moving ammunition into the inventory…"


func _update_ammunition_pouch_window() -> void:
	if _ammunition_pouch_grid == null:
		return
	for child: Node in _ammunition_pouch_grid.get_children():
		_ammunition_pouch_grid.remove_child(child)
		child.queue_free()
	var items_by_slot: Dictionary = {}
	for payload: Variant in _ammunition_pouch_items:
		if typeof(payload) == TYPE_DICTIONARY:
			items_by_slot[int((payload as Dictionary).get("slot", -1))] = payload
	_ammunition_pouch_summary.text = "%d/%d ammunition types stored • %d/%d slots unlocked" % [
		items_by_slot.size(), _ammunition_pouch_unlocked, _ammunition_pouch_unlocked, _ammunition_pouch_capacity,
	]
	for slot: int in range(_ammunition_pouch_capacity):
		var button: Variant = INVENTORY_SLOT_SCRIPT.new()
		button.custom_minimum_size = Vector2(96, 88)
		button.focus_mode = Control.FOCUS_NONE
		button.text = ""
		if slot >= _ammunition_pouch_unlocked:
			button.configure(slot, false, false)
			button.icon = EMPTY_SLOT_ICON
			button.expand_icon = true
			button.disabled = true
			button.modulate = Color(0.28, 0.31, 0.34)
			button.tooltip_text = "Locked ammunition slot"
		elif items_by_slot.has(slot):
			var item: Dictionary = items_by_slot[slot]
			var item_id: String = str(item.get("item_id", ""))
			var quantity: int = int(item.get("quantity", 0))
			button.configure(slot, true, true)
			button.icon = ITEM_ICONS.get(item_id, EMPTY_SLOT_ICON) as Texture2D
			button.expand_icon = true
			button.set_stack_quantity(quantity)
			button.tooltip_text = "%s ×%d • left-click to return the whole stack to inventory" % [item_id.capitalize(), quantity]
			button.pressed.connect(_request_ammunition_withdraw.bind(slot))
		else:
			button.configure(slot, false, true)
			button.icon = EMPTY_SLOT_ICON
			button.expand_icon = true
			button.tooltip_text = "Unlocked ammunition slot • click ammunition in your inventory to store it"
		_ammunition_pouch_grid.add_child(button)


func _request_bank_deposit(item_id: String, instance_id: String, inventory_slot: int, quantity: int) -> void:
	if not bank_hud.visible or _is_test_supply_chest(_active_bank_chest_id) or item_id.is_empty() or _player_defeated:
		return
	if _is_item_equipped(item_id, instance_id):
		status_label.text = "Unequip the item before depositing it"
		return
	if network.send_bank_deposit(_next_bank_sequence, item_id, inventory_slot, instance_id):
		_next_bank_sequence += 1
		status_label.text = "Depositing %s%s…" % [
			item_id.capitalize(),
			" ×%d" % quantity if quantity > 1 else "",
		]


func _request_bank_withdraw(bank_slot: int, quantity: int = 1) -> void:
	if not bank_hud.visible or _is_test_supply_chest(_active_bank_chest_id) or bank_slot < 0 or _player_defeated:
		return
	if network.send_bank_withdraw(_next_bank_sequence, bank_slot, quantity):
		_next_bank_sequence += 1
		status_label.text = "Withdrawing %d item%s…" % [quantity, "" if quantity == 1 else "s"]


func _setup_bank_withdraw_dialog() -> void:
	_bank_withdraw_dialog = ConfirmationDialog.new()
	_bank_withdraw_dialog.title = "Withdraw items"
	_bank_withdraw_dialog.dialog_text = "Choose an amount to withdraw:"
	_bank_withdraw_dialog.min_size = Vector2i(360, 180)
	_bank_withdraw_dialog.get_ok_button().text = "Withdraw"
	_bank_withdraw_amount = SpinBox.new()
	_bank_withdraw_amount.min_value = 1
	_bank_withdraw_amount.step = 1
	_bank_withdraw_amount.custom_minimum_size = Vector2(280, 44)
	_bank_withdraw_amount.position = Vector2(40, 74)
	_bank_withdraw_dialog.add_child(_bank_withdraw_amount)
	_bank_withdraw_dialog.confirmed.connect(_confirm_bank_withdraw)
	add_child(_bank_withdraw_dialog)


func _on_bank_item_pressed(bank_slot: int) -> void:
	_request_bank_withdraw(bank_slot, 1)


func _on_bank_item_input(event: InputEvent, bank_slot: int, available_quantity: int) -> void:
	if not event is InputEventMouseButton:
		return
	var mouse_button := event as InputEventMouseButton
	if not mouse_button.pressed or mouse_button.button_index != MOUSE_BUTTON_RIGHT:
		return
	_bank_menu_slot = bank_slot
	_bank_menu_quantity = available_quantity
	_bank_item_menu.clear()
	_bank_item_menu.add_item("Take 1", BANK_TAKE_ONE)
	_bank_item_menu.add_item("Take X", BANK_TAKE_X)
	_bank_item_menu.add_item("Take All", BANK_TAKE_ALL)
	_bank_item_menu.position = Vector2i(mouse_button.global_position)
	_bank_item_menu.popup()
	get_viewport().set_input_as_handled()


func _on_bank_item_menu_action(menu_id: int) -> void:
	if _bank_menu_slot < 0 or _bank_menu_quantity < 1:
		return
	match menu_id:
		BANK_TAKE_ONE:
			_request_bank_withdraw(_bank_menu_slot, 1)
		BANK_TAKE_X:
			_open_bank_withdraw_dialog(_bank_menu_slot, _bank_menu_quantity)
		BANK_TAKE_ALL:
			_request_bank_withdraw(_bank_menu_slot, _bank_menu_quantity)


func _open_bank_withdraw_dialog(bank_slot: int, available_quantity: int) -> void:
	_pending_bank_withdraw_slot = bank_slot
	_bank_withdraw_amount.max_value = available_quantity
	_bank_withdraw_amount.value = mini(available_quantity, 10)
	_bank_withdraw_dialog.popup_centered()
	_bank_withdraw_amount.grab_focus()
	_bank_withdraw_amount.get_line_edit().select_all()


func _confirm_bank_withdraw() -> void:
	if _pending_bank_withdraw_slot < 0:
		return
	var bank_slot: int = _pending_bank_withdraw_slot
	_pending_bank_withdraw_slot = -1
	_request_bank_withdraw(bank_slot, int(_bank_withdraw_amount.value))


func _update_bank_hud() -> void:
	if bank_grid == null:
		return
	for child: Node in bank_grid.get_children():
		bank_grid.remove_child(child)
		child.queue_free()
	var items_by_slot: Dictionary = {}
	for payload: Variant in _bank_items:
		if typeof(payload) != TYPE_DICTIONARY:
			continue
		var item: Dictionary = payload
		var slot: int = int(item.get("slot", -1))
		if slot >= 0 and slot < _bank_capacity:
			items_by_slot[slot] = item
	bank_summary.text = "%d/%d slots used" % [items_by_slot.size(), _bank_capacity]
	for slot: int in range(_bank_capacity):
		var slot_button: Variant = INVENTORY_SLOT_SCRIPT.new()
		slot_button.custom_minimum_size = Vector2(112, 86)
		slot_button.size_flags_horizontal = Control.SIZE_EXPAND_FILL
		slot_button.focus_mode = Control.FOCUS_NONE
		slot_button.text = ""
		if items_by_slot.has(slot):
			var item: Dictionary = items_by_slot[slot]
			var item_id: String = str(item.get("item_id", ""))
			var quantity: int = int(item.get("quantity", 1))
			var reinforcement: int = int(item.get("reinforcement", 0))
			slot_button.configure(slot, true, true)
			slot_button.icon = ITEM_ICONS.get(item_id, EMPTY_SLOT_ICON) as Texture2D
			slot_button.expand_icon = true
			slot_button.set_stack_quantity(quantity)
			slot_button.set_reinforcement(reinforcement)
			slot_button.tooltip_text = "%s%s ×%d • left-click: take 1 • right-click: options" % [
				item_id.capitalize(),
				" +%d" % reinforcement if reinforcement > 0 else "",
				quantity,
			]
			slot_button.pressed.connect(_on_bank_item_pressed.bind(slot))
			slot_button.gui_input.connect(_on_bank_item_input.bind(slot, quantity))
		else:
			slot_button.configure(slot, false, true)
			slot_button.icon = EMPTY_SLOT_ICON
			slot_button.expand_icon = true
			slot_button.disabled = true
			slot_button.modulate = Color(0.6, 0.64, 0.68)
		bank_grid.add_child(slot_button)


func _inventory_quantity(item_id: String) -> int:
	var total: int = 0
	for payload: Variant in _inventory_items:
		if typeof(payload) != TYPE_DICTIONARY:
			continue
		var item: Dictionary = payload
		if str(item.get("item_id", "")) == item_id:
			total += int(item.get("quantity", 0))
	return total


func _change_inventory_page(direction: int) -> void:
	var page_count: int = ceili(float(INVENTORY_TOTAL_SLOTS) / float(INVENTORY_PAGE_SIZE))
	_inventory_page = clampi(_inventory_page + direction, 0, page_count - 1)
	_update_inventory_hud()


func _update_inventory_hud() -> void:
	for child: Node in inventory_grid.get_children():
		inventory_grid.remove_child(child)
		child.queue_free()

	var items_by_slot: Dictionary = {}
	for payload: Variant in _inventory_items:
		if typeof(payload) != TYPE_DICTIONARY:
			continue
		var item: Dictionary = payload
		var slot: int = int(item.get("slot", -1))
		var item_id: String = str(item.get("item_id", ""))
		var quantity: int = int(item.get("quantity", 0))
		if slot >= 0 and slot < _inventory_capacity and not item_id.is_empty() and quantity > 0:
			items_by_slot[slot] = item

	var unlocked_slots: int = mini(_inventory_capacity, INVENTORY_TOTAL_SLOTS)
	var page_count: int = ceili(float(INVENTORY_TOTAL_SLOTS) / float(INVENTORY_PAGE_SIZE))
	_inventory_page = clampi(_inventory_page, 0, page_count - 1)
	var first_slot: int = _inventory_page * INVENTORY_PAGE_SIZE
	var final_slot: int = mini(first_slot + INVENTORY_PAGE_SIZE, INVENTORY_TOTAL_SLOTS)
	inventory_summary.text = "Inventory  •  %d/%d slots used" % [items_by_slot.size(), unlocked_slots]
	inventory_page_label.text = "Page %d/%d  •  %d/%d unlocked" % [
		_inventory_page + 1,
		page_count,
		unlocked_slots,
		INVENTORY_TOTAL_SLOTS,
	]
	inventory_previous_page_button.disabled = _inventory_page == 0
	inventory_next_page_button.disabled = _inventory_page >= page_count - 1
	var is_banking: bool = bank_hud.visible and not _is_test_supply_chest(_active_bank_chest_id)
	var is_storing_ammunition: bool = _ammunition_pouch_window != null and _ammunition_pouch_window.visible
	for slot: int in range(first_slot, final_slot):
		var slot_button: Variant = INVENTORY_SLOT_SCRIPT.new()
		slot_button.custom_minimum_size = Vector2(104, 92)
		slot_button.size_flags_horizontal = Control.SIZE_EXPAND_FILL
		slot_button.focus_mode = Control.FOCUS_NONE
		slot_button.mouse_filter = Control.MOUSE_FILTER_STOP
		slot_button.text = ""
		if slot >= _inventory_capacity:
			slot_button.configure(slot, false, false)
			slot_button.icon = EMPTY_SLOT_ICON
			slot_button.expand_icon = true
			slot_button.disabled = true
			slot_button.modulate = Color(0.28, 0.31, 0.34)
			slot_button.tooltip_text = "Locked inventory slot"
		elif items_by_slot.has(slot):
			slot_button.configure(slot, true, true)
			var item: Dictionary = items_by_slot[slot]
			var item_id: String = str(item.get("item_id", ""))
			var quantity: int = int(item.get("quantity", 1))
			var instance_id: String = str(item.get("instance_id", ""))
			var reinforcement: int = int(item.get("reinforcement", 0))
			slot_button.set_stack_quantity(quantity)
			slot_button.set_reinforcement(reinforcement)
			var is_equipable: bool = EQUIPMENT_ICONS.has(item_id) or ARMOUR_SLOTS.has(item_id) or SPELLBOOK_ITEMS.has(item_id)
			var is_consumable: bool = CONSUMABLE_ITEMS.has(item_id)
			var is_brewing_ingredient: bool = BREWING_INGREDIENTS.has(item_id)
			var is_fletching_input: bool = FLETCHING_INPUTS.has(item_id)
			var is_uncut_gem: bool = UNCUT_GEMS.has(item_id)
			var equipment_pending: bool = _last_equipment_sequence < _next_equipment_sequence - 1
			var is_queued: bool = equipment_pending and _queued_item == item_id and (
				_queued_instance_id.is_empty() or _queued_instance_id == instance_id
			)
			var is_consume_queued: bool = _queued_consume_item == item_id
			var is_equipped: bool = _is_item_equipped(item_id, instance_id)
			slot_button.toggle_mode = is_equipable and not is_banking
			var displayed_equipped: bool = not _queued_unequip if is_queued else is_equipped
			slot_button.set_pressed_no_signal(displayed_equipped)
			slot_button.icon = ITEM_ICONS.get(item_id, EMPTY_SLOT_ICON) as Texture2D
			slot_button.expand_icon = true
			var item_status: String = ""
			if is_queued:
				item_status = "Queued"
			elif is_consume_queued:
				item_status = "Using"
			elif is_equipped:
				item_status = "Equipped"
			var quantity_text: String = " ×%d" % quantity if quantity > 1 else ""
			var reinforcement_text: String = " +%d" % reinforcement if reinforcement > 0 else ""
			var status_text: String = " • %s" % item_status if not item_status.is_empty() else ""
			var equip_action: String = "left-click to unequip" if is_equipped else "left-click to equip"
			var interaction_text: String = "right-click to drop" if not instance_id.is_empty() else "right-click to drop all"
			if is_equipable:
				interaction_text = "%s • %s" % [equip_action, interaction_text]
			elif is_consumable:
				var consumable: Dictionary = CONSUMABLE_ITEMS[item_id] as Dictionary
				if int(consumable.get("healing", 0)) > 0:
					interaction_text = "left-click to eat (+%d health, max %d/slot) • %s" % [
						int(consumable.get("healing", 0)), int(consumable.get("stack_limit", 1)), interaction_text,
					]
				else:
					interaction_text = "left-click to drink one dose (+%d prayer, max %d doses/slot) • %s" % [
						int(consumable.get("prayer_restore", 0)), int(consumable.get("stack_limit", 1)), interaction_text,
					]
			elif is_brewing_ingredient:
				interaction_text = "left-click to open Experimental Brewing • %s" % interaction_text
			elif is_fletching_input:
				interaction_text = "left-click to open Fletching • %s" % interaction_text
			elif is_uncut_gem:
				interaction_text = "left-click to cut with a Chisel • %s" % interaction_text
			if is_equipped and SPELLBOOK_ITEMS.has(item_id):
				interaction_text = "left-click to unequip • right-click to unequip and drop"
			if is_banking:
				interaction_text = "left-click to deposit whole stack"
			elif is_storing_ammunition and AMMUNITION_ITEMS.has(item_id):
				interaction_text = "left-click to move the whole stack into the ammunition pouch"
			slot_button.tooltip_text = "%s%s%s%s • slot %d%s" % [
				str(ITEM_DISPLAY_NAMES.get(item_id, item_id.capitalize())), reinforcement_text, quantity_text, status_text,
				slot + 1,
				" — %s" % interaction_text,
			]
			if ITEM_COMBAT_BONUSES.has(item_id):
				slot_button.tooltip_text += " • %s" % str(ITEM_COMBAT_BONUSES[item_id])
			if EQUIPMENT_REQUIREMENTS.has(item_id):
				slot_button.tooltip_text += " • requires %s" % str(EQUIPMENT_REQUIREMENTS[item_id])
			slot_button.disabled = _player_defeated and is_equipable
			if is_equipped or is_queued or is_consume_queued:
				var selection_style := StyleBoxFlat.new()
				selection_style.bg_color = Color(0.06, 0.09, 0.12, 0.96)
				selection_style.border_width_left = 4
				selection_style.border_width_top = 4
				selection_style.border_width_right = 4
				selection_style.border_width_bottom = 4
				selection_style.border_color = Color(0.35, 0.7, 1.0) if is_queued or is_consume_queued else Color(0.95, 0.72, 0.2)
				selection_style.corner_radius_top_left = 6
				selection_style.corner_radius_top_right = 6
				selection_style.corner_radius_bottom_left = 6
				selection_style.corner_radius_bottom_right = 6
				slot_button.add_theme_stylebox_override("pressed", selection_style)
				slot_button.add_theme_stylebox_override("hover_pressed", selection_style)
			if is_banking:
				slot_button.pressed.connect(_request_bank_deposit.bind(item_id, instance_id, slot, quantity))
			elif is_storing_ammunition and AMMUNITION_ITEMS.has(item_id):
				slot_button.pressed.connect(_request_ammunition_deposit.bind(slot))
			elif is_equipable:
				slot_button.pressed.connect(_request_equip.bind(item_id, instance_id))
			elif is_consumable:
				slot_button.pressed.connect(_request_consume_item.bind(item_id))
			elif is_brewing_ingredient:
				slot_button.pressed.connect(_open_brewing_window.bind(item_id))
			elif is_fletching_input:
				slot_button.pressed.connect(_open_fletching_window.bind(item_id))
			elif is_uncut_gem:
				slot_button.pressed.connect(_request_cut_gem.bind(item_id))
			slot_button.gui_input.connect(_on_inventory_item_input.bind(item_id, instance_id))
		else:
			slot_button.configure(slot, false, true)
			slot_button.icon = EMPTY_SLOT_ICON
			slot_button.expand_icon = true
			slot_button.modulate = Color(0.9, 0.93, 0.96)
		slot_button.swap_requested.connect(_request_inventory_swap)
		inventory_grid.add_child(slot_button)


func _update_combat_hud() -> void:
	var state: String = "queued" if _last_action_sequence < _next_action_sequence - 1 else _last_action_result
	if _player_defeated:
		state = "respawn in %d" % maxi(0, _player_respawn_at_pulse - _action_pulse)
	player_health_bar.max_value = maxi(1, _player_maximum_health)
	player_health_bar.value = _player_health
	player_health_label.text = "Health  %d/%d  •  Dodge %d%%" % [
		_player_health,
		_player_maximum_health,
		roundi(_player_dodge_rate * 100.0),
	]
	player_prayer_bar.max_value = maxi(1, _player_maximum_prayer)
	player_prayer_bar.value = _player_prayer
	player_prayer_label.text = "Prayer  %d/%d" % [_player_prayer, _player_maximum_prayer]
	var target_ids: Array = _target_states.keys()
	target_ids.sort()
	var displayed_target_key: String = _combat_target
	if not _target_states.has(displayed_target_key):
		displayed_target_key = ""
		for target_id: String in target_ids:
			var candidate: Dictionary = _target_states[target_id]
			if str(candidate.get("target_id", "")) == _local_player_id:
				displayed_target_key = target_id
				break
	var displayed_target: Dictionary = _target_states.get(displayed_target_key, {})
	_target_hud_active = not displayed_target_key.is_empty() and not displayed_target.is_empty()
	target_hud.visible = _target_hud_active or _hud_edit_mode
	var displayed_health: int = int(displayed_target.get("health", 0))
	var displayed_maximum_health: int = int(displayed_target.get("maximum_health", 100))
	var displayed_defeated: bool = bool(displayed_target.get("defeated", false))
	var displayed_respawn_pulse: int = int(displayed_target.get("respawn_at_pulse", 0))
	var displayed_target_id: String = str(displayed_target.get("target_id", ""))
	var target_name: String = str(displayed_target.get("name", "No Target"))
	var target_attack_type: String = str(displayed_target.get("attack_type", "crush")).capitalize()
	target_health_bar.max_value = maxi(1, displayed_maximum_health)
	target_health_bar.value = displayed_health
	target_health_label.text = "%s  %d/%d" % [target_name, displayed_health, displayed_maximum_health]

	var has_aggro: bool = displayed_target_id == _local_player_id
	aggro_label.text = "%s • ENEMY AGGRO" % target_attack_type if has_aggro else "%s attacks • No aggro" % target_attack_type
	aggro_label.modulate = Color(1.0, 0.42, 0.32) if has_aggro else Color(0.65, 0.72, 0.78)
	if _player_defeated:
		combat_state_label.text = "Defeated — %s" % state
	elif displayed_defeated:
		combat_state_label.text = "Target respawns in %d pulses" % maxi(0, displayed_respawn_pulse - _action_pulse)
	elif _auto_approaching:
		combat_state_label.text = "Closing distance to target"
	elif _last_action_result == "missing_reagents" and not _selected_spell.is_empty():
		combat_state_label.text = "Cannot cast %s — missing Air Runes" % _selected_spell.capitalize()
	elif not _combat_target.is_empty():
		combat_state_label.text = "Auto-attacking %s  •  Last: %s" % [_combat_target, state]
	elif has_aggro:
		combat_state_label.text = "Not attacking — create distance to escape"
	else:
		combat_state_label.text = "Out of combat"
	if _combat_style == "unfocused":
		combat_state_label.text += "  •  COMBAT XP LOCKED"

	var weapon_interval: int = int(_combat_stats.get("attack_interval", 4))
	if _equipped_weapon == "dagger":
		weapon_interval = int(_combat_stats.get("attack_interval", 3))
	elif _equipped_weapon == "bow":
		weapon_interval = int(_combat_stats.get("attack_interval", 4 if _combat_style == "rapid" else 5))
	elif _equipped_weapon == "mace" or _equipped_weapon == "staff":
		weapon_interval = int(_combat_stats.get("attack_interval", 5))
	elif _equipped_weapon == "battle_axe":
		weapon_interval = int(_combat_stats.get("attack_interval", 6))
	if not _selected_spell.is_empty():
		weapon_interval = int(_combat_stats.get("attack_interval", 5))
	var remaining: int = maxi(0, _next_attack_pulse - _action_pulse)
	cooldown_bar.max_value = weapon_interval
	cooldown_bar.value = weapon_interval - mini(weapon_interval, remaining)
	if _auto_approaching:
		cooldown_label.text = "Attack: closing distance"
	elif _combat_target.is_empty():
		cooldown_label.text = "Attack: idle"
	elif remaining > 0:
		cooldown_label.text = "Next attack in %d pulse%s" % [remaining, "" if remaining == 1 else "s"]
	else:
		cooldown_label.text = "Attack ready"

	var equipment_pending: bool = _last_equipment_sequence < _next_equipment_sequence - 1
	var queued_weapon: String = _queued_item if EQUIPMENT_ICONS.has(_queued_item) else ""
	var displayed_weapon: String = queued_weapon if equipment_pending and not queued_weapon.is_empty() else _equipped_weapon
	sword_button.set_pressed_no_signal(displayed_weapon == "sword")
	dagger_button.set_pressed_no_signal(displayed_weapon == "dagger")
	bow_button.set_pressed_no_signal(displayed_weapon == "bow")
	sword_button.text = "1  Sword%s" % ("  • queued" if equipment_pending and queued_weapon == "sword" else "")
	dagger_button.text = "2  Dagger%s" % ("  • queued" if equipment_pending and queued_weapon == "dagger" else "")
	bow_button.text = "3  Bow%s" % ("  • queued" if equipment_pending and queued_weapon == "bow" else "")
	sword_button.disabled = _player_defeated
	dagger_button.disabled = _player_defeated
	bow_button.disabled = _player_defeated
	disengage_button.disabled = _player_defeated or (_combat_target.is_empty() and _last_action_sequence >= _next_action_sequence - 1)
