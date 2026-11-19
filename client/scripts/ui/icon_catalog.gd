class_name IconCatalog
extends RefCounted

const ITEM_DISPLAY_NAMES: Dictionary = {
	"gatherers_tonic": "Gatherer's Tonic",
	"ember_pepper": "Fire Pepper", "cinnamon_bark": "Cinnabark",
	"dark_rune": "Shadow Rune",
	"blank_rune": "Runestone",
	"sunleaf": "Dawnpetal", "moonleaf": "Nightbloom",
	"amulet_ferocity": "Ruby Amulet", "amulet_hawk": "Emerald Amulet",
	"amulet_embers": "Sapphire Amulet", "amulet_fury": "Amulet of Fury",
	"diamond_amulet": "Diamond Amulet",
	"skin_bible": "Skin Bible", "holy_book": "Holy Book", "clerics_tome": "Cleric's Tome",
	"ruby_ring": "Ruby Ring", "emerald_ring": "Emerald Ring",
	"sapphire_ring": "Sapphire Ring", "diamond_ring": "Diamond Ring",
	"ring_of_might": "Ring of Might", "ring_of_marksmanship": "Ring of Marksmanship",
	"ring_of_sorcery": "Ring of Sorcery",
	"tyrants_ring": "Tyrant's Ring",
	"ranged_tyrants_ring": "Ranged Tyrant's Ring",
	"mage_tyrants_ring": "Mage Tyrant's Ring",
	"vanguard_cape": "Vanguard Cape", "wayfarer_cape": "Wayfarer Cape", "arcanist_cape": "Arcanist Cape",
	"wand": "Apprentice Wand", "moonspun_wand": "Moonspun Wand",
	"staff": "Mage's Staff", "enchanted_staff": "Enchanted Staff", "stardew_staff": "Stardew Staff",
}

static var _centered_icon_cache: Dictionary = {}


static func item_display_name(item_id: String) -> String:
	return str(ITEM_DISPLAY_NAMES.get(item_id, item_id.capitalize()))


# Some source icons have asymmetric transparent padding. Keep their original
# canvas size and visible scale, but move the painted pixels to its centre.
static func centered_icon(texture: Texture2D) -> Texture2D:
	if texture == null:
		return texture
	var cache_key: String = texture.resource_path
	if cache_key.is_empty():
		cache_key = str(texture.get_instance_id())
	if _centered_icon_cache.has(cache_key):
		return _centered_icon_cache[cache_key] as Texture2D
	var source: Image = texture.get_image()
	if source == null or source.is_empty():
		return texture
	source.convert(Image.FORMAT_RGBA8)
	var used_rect: Rect2i = source.get_used_rect()
	if used_rect.size.x <= 0 or used_rect.size.y <= 0:
		return texture
	var centered := Image.create(
		source.get_width(),
		source.get_height(),
		false,
		Image.FORMAT_RGBA8,
	)
	centered.fill(Color.TRANSPARENT)
	var destination := Vector2i(
		(source.get_width() - used_rect.size.x) / 2,
		(source.get_height() - used_rect.size.y) / 2,
	)
	centered.blit_rect(source, used_rect, destination)
	var result := ImageTexture.create_from_image(centered)
	_centered_icon_cache[cache_key] = result
	return result

# Equipment and item icons.
const SWORD: Texture2D = preload("res://assets/icons/equipment/weapons/iron_sword.png")
const BRONZE_SWORD: Texture2D = preload("res://assets/icons/equipment/weapons/bronze_sword.png")
const STEEL_SWORD: Texture2D = preload("res://assets/icons/equipment/weapons/steel_sword.png")
const MITHRIL_SWORD: Texture2D = preload("res://assets/icons/equipment/weapons/mithril_sword.png")
const RUNE_SWORD: Texture2D = preload("res://assets/icons/equipment/weapons/sword.svg")
const DAGGER: Texture2D = preload("res://assets/icons/equipment/weapons/dagger.svg")
const BOW: Texture2D = preload("res://assets/icons/equipment/weapons/bow.svg")
const OAK_SHORTBOW: Texture2D = preload("res://assets/icons/generated/placeholders/oak_shortbow.svg")
const WILLOW_SHORTBOW: Texture2D = preload("res://assets/icons/generated/placeholders/willow_shortbow.svg")
const MAPLE_SHORTBOW: Texture2D = preload("res://assets/icons/generated/placeholders/maple_shortbow.svg")
const YEW_SHORTBOW: Texture2D = preload("res://assets/icons/generated/placeholders/yew_shortbow.svg")
const MACE: Texture2D = preload("res://assets/icons/equipment/weapons/mace.svg")
const BATTLE_AXE: Texture2D = preload("res://assets/icons/equipment/weapons/battle_axe.svg")
const STAFF: Texture2D = preload("res://assets/icons/equipment/weapons/staff.svg")
const WAND: Texture2D = preload("res://assets/icons/generated/placeholders/wand.svg")
const MOONSPUN_WAND: Texture2D = preload("res://assets/icons/generated/placeholders/moonspun_wand.svg")
const ENCHANTED_STAFF: Texture2D = preload("res://assets/icons/generated/placeholders/enchanted_staff.svg")
const STARDEW_STAFF: Texture2D = preload("res://assets/icons/generated/placeholders/stardew_staff.svg")
const SPEAR: Texture2D = preload("res://assets/icons/generated/placeholders/spear.svg")
const HALBERD: Texture2D = preload("res://assets/icons/generated/placeholders/halberd.svg")
const CROSSBOW: Texture2D = preload("res://assets/icons/generated/placeholders/crossbow.svg")
const TRAINING_HELM: Texture2D = preload("res://assets/icons/equipment/armour/training_helm.svg")
const BRONZE_HELM: Texture2D = preload("res://assets/icons/equipment/armour/bronze_helm.svg")
const BRONZE_PLATE: Texture2D = preload("res://assets/icons/equipment/armour/bronze_platebody.svg")
const IRON_HELM: Texture2D = preload("res://assets/icons/equipment/armour/iron_helm.svg")
const IRON_PLATE: Texture2D = preload("res://assets/icons/equipment/armour/iron_platebody.svg")
const STEEL_HELM: Texture2D = preload("res://assets/icons/equipment/armour/steel_helm.svg")
const STEEL_PLATE: Texture2D = preload("res://assets/icons/equipment/armour/steel_platebody.svg")
const MITHRIL_HELM: Texture2D = preload("res://assets/icons/equipment/armour/mithril_helm.svg")
const MITHRIL_PLATE: Texture2D = preload("res://assets/icons/equipment/armour/mithril_platebody.svg")
const ADAMANTITE_HELM: Texture2D = preload("res://assets/icons/equipment/armour/adamantite_helm.svg")
const ADAMANTITE_PLATE: Texture2D = preload("res://assets/icons/equipment/armour/adamantite_platebody.svg")
const TRAINING_PLATE: Texture2D = preload("res://assets/icons/equipment/armour/training_plate.svg")
const SWIFT_HELM: Texture2D = preload("res://assets/icons/equipment/armour/swift_helm.svg")
const SWIFT_BODY: Texture2D = preload("res://assets/icons/equipment/armour/swift_body.svg")
const SWIFT_BOOTS: Texture2D = preload("res://assets/icons/equipment/armour/swift_boots.svg")
const EMPTY_SLOT: Texture2D = preload("res://assets/icons/items/empty_slot.svg")
const COINS: Texture2D = preload("res://assets/icons/items/coins.png")
const BRONZE_PICKAXE: Texture2D = preload("res://assets/icons/items/bronze_pickaxe.svg")
const IRON_PICKAXE: Texture2D = preload("res://assets/icons/items/iron_pickaxe.svg")
const STEEL_PICKAXE: Texture2D = preload("res://assets/icons/items/steel_pickaxe.svg")
const MITHRIL_PICKAXE: Texture2D = preload("res://assets/icons/items/mithril_pickaxe.svg")
const BRONZE_AXE: Texture2D = preload("res://assets/icons/items/bronze_axe.svg")
const IRON_AXE: Texture2D = preload("res://assets/icons/items/iron_axe.svg")
const STEEL_AXE: Texture2D = preload("res://assets/icons/items/steel_axe.svg")
const MITHRIL_AXE: Texture2D = preload("res://assets/icons/items/mithril_axe.svg")
const KNIFE: Texture2D = preload("res://assets/icons/items/tools/knife.svg")
const CHISEL: Texture2D = preload("res://assets/icons/items/tools/chisel.svg")
const SEWING_NEEDLE: Texture2D = preload("res://assets/icons/items/tools/sewing_needle.svg")
const THREAD: Texture2D = preload("res://assets/icons/items/tools/thread.svg")
const TINDERBOX: Texture2D = preload("res://assets/icons/items/tools/tinderbox.svg")
const HAMMER: Texture2D = preload("res://assets/icons/items/tools/hammer.svg")
const PRAYER_POTION: Texture2D = preload("res://assets/icons/items/consumables/prayer_potion.svg")
const FISHING_ROD: Texture2D = preload("res://assets/icons/items/tools/fishing_rod.svg")
const LOBSTER_CAGE: Texture2D = preload("res://assets/icons/items/tools/lobster_cage.svg")
const NORMAL_LOGS: Texture2D = preload("res://assets/icons/logs/normal_logs_150.png")
const OAK_LOGS: Texture2D = preload("res://assets/icons/logs/oak_logs_150.png")
const COPPER_ORE: Texture2D = preload("res://assets/icons/ore/Pixel-Art Copper Ore Cluster-1.png")
const TIN_ORE: Texture2D = preload("res://assets/icons/ore/Stylized Tin Ore Pile Icon-2.png")
const IRON_ORE: Texture2D = preload("res://assets/icons/ore/OSRS-Style Iron Ore Cluster-3.png")
const COAL: Texture2D = preload("res://assets/icons/ore/Pixel-Art Coal Ore Cluster-4.png")
const MITHRIL_ORE: Texture2D = preload("res://assets/icons/ore/OSRS-Style Mithril Rock Cluster-5.png")
const SILVER_ORE: Texture2D = preload("res://assets/icons/items/materials/silver_ore.svg")
const CLAY: Texture2D = preload("res://assets/icons/items/materials/clay.svg")
const ADAMANTITE_ORE: Texture2D = preload("res://assets/icons/items/materials/adamantite_ore.svg")
const BRONZE_BAR: Texture2D = preload("res://assets/icons/bars/Bronze Ingot Stack Icon-1.png")
const IRON_BAR: Texture2D = preload("res://assets/icons/bars/Stacked Iron Ingots-2.png")
const MITHRIL_BAR: Texture2D = preload("res://assets/icons/bars/Mithril Ingot Stack Icon-4.png")
const STANDARD_SPELLBOOK: Texture2D = preload("res://assets/icons/navigation/tab_magic.svg")

const EQUIPMENT_SLOTS: Dictionary = {
	"head": preload("res://assets/icons/equipment/slots/head.svg"),
	"cape": preload("res://assets/icons/equipment/slots/cape.svg"),
	"neck": preload("res://assets/icons/equipment/slots/neck.svg"),
	"ring": preload("res://assets/icons/equipment/slots/ring.svg"),
	"main_hand": preload("res://assets/icons/equipment/slots/main_hand.svg"),
	"left_hand": preload("res://assets/icons/equipment/slots/left_hand.svg"),
	"body": preload("res://assets/icons/equipment/slots/body.svg"),
	"hands": preload("res://assets/icons/equipment/slots/hands.svg"),
	"legs": preload("res://assets/icons/equipment/slots/legs.svg"),
	"feet": preload("res://assets/icons/equipment/slots/feet.svg"),
	"spellbook": preload("res://assets/icons/equipment/slots/spellbook.svg"),
	"ammunition": preload("res://assets/icons/ranged/arrows/iron_arrow_150.png"),
	"rune_pouch": preload("res://assets/icons/runes/air_rune.png"),
}

# Main HUD navigation icons.
const TAB_STYLES: Texture2D = preload("res://assets/icons/navigation/tab_styles.svg")
const TAB_INVENTORY: Texture2D = preload("res://assets/icons/navigation/tab_inventory.svg")
const TAB_SKILLS: Texture2D = preload("res://assets/icons/navigation/tab_skills.svg")
const TAB_EQUIPMENT: Texture2D = preload("res://assets/icons/navigation/tab_equipment.svg")
const TAB_MAGIC: Texture2D = preload("res://assets/icons/navigation/tab_magic.svg")
const TAB_PRAYER: Texture2D = preload("res://assets/icons/navigation/tab_prayer.svg")
const TAB_SETTINGS: Texture2D = preload("res://assets/icons/navigation/tab_settings.svg")
const TAB_ACCOUNT: Texture2D = preload("res://assets/icons/navigation/tab_account.svg")

const SPELL_ICONS: Dictionary = {
	"wind_bolt": preload("res://assets/icons/spells/wind_blade.svg"),
	"water_bolt": preload("res://assets/icons/spells/frost_shard.svg"),
	"earth_bolt": preload("res://assets/icons/spells/pebble.svg"),
	"fire_bolt": preload("res://assets/icons/spells/ember.svg"),
	"holy_bolt": preload("res://assets/icons/spells/light.png"),
	"dark_bolt": preload("res://assets/icons/spells/miasma.png"),
}

const SPELL_PLACEHOLDERS: Array = [
	{"book_id": "dark", "name": "Flux", "icon": preload("res://assets/icons/spells/acid.png")},
	{"book_id": "dark", "name": "Eclipse", "icon": preload("res://assets/icons/spells/absolute_zero.svg")},
	{"book_id": "dark", "name": "Nightmare", "icon": preload("res://assets/icons/spells/plague_bloom.png")},
	{"book_id": "dark", "name": "Apocalypse", "icon": preload("res://assets/icons/spells/inferno.svg")},
	{"book_id": "cleric", "name": "Mend", "icon": preload("res://assets/icons/spells/shine.png")},
	{"book_id": "cleric", "name": "Cleanse", "icon": preload("res://assets/icons/spells/aura.png")},
	{"book_id": "cleric", "name": "Sanctuary", "icon": preload("res://assets/icons/spells/divine.png")},
	{"book_id": "cleric", "name": "Renew", "icon": preload("res://assets/icons/spells/luce.png")},
	{"book_id": "holy", "name": "Shine", "icon": preload("res://assets/icons/spells/shine.png")},
	{"book_id": "holy", "name": "Divine", "icon": preload("res://assets/icons/spells/divine.png")},
	{"book_id": "holy", "name": "Aura", "icon": preload("res://assets/icons/spells/aura.png")},
	{"book_id": "holy", "name": "Luce", "icon": preload("res://assets/icons/spells/luce.png")},
]

const PRAYER_ICONS: Dictionary = {
	"accuracy": preload("res://assets/icons/prayers/18_accuracy.png"),
	"strength": preload("res://assets/icons/prayers/03_strength.png"),
	"defence": preload("res://assets/icons/prayers/14_defence_boost.png"),
	"evasion": preload("res://assets/icons/prayers/04_agility.png"),
	"ranged_strength": preload("res://assets/icons/prayers/16_ranged_boost.png"),
	"magic_strength": preload("res://assets/icons/prayers/17_magic_boost.png"),
	"regeneration": preload("res://assets/icons/prayers/20_regeneration.svg"),
	"protect_melee": preload("res://assets/icons/prayers/07_melee.png"),
	"protect_ranged": preload("res://assets/icons/prayers/08_ranged.png"),
	"protect_magic": preload("res://assets/icons/prayers/09_magic.png"),
	"melee_mastery": preload("res://assets/icons/prayers/13_melee_boost.png"),
	"ranged_mastery": preload("res://assets/icons/prayers/16_ranged_boost.png"),
	"magic_mastery": preload("res://assets/icons/prayers/17_magic_boost.png"),
	"mercy": preload("res://assets/icons/prayers/11_vitality.png"),
	"wrath": preload("res://assets/icons/prayers/19_wrath.png"),
}

const PRAYER_PLACEHOLDERS: Array = [
	{"name": "Wisdom", "icon": preload("res://assets/icons/prayers/06_wisdom.png")},
	{"name": "Protection", "icon": preload("res://assets/icons/prayers/02_protection.png")},
	{"name": "Vision", "icon": preload("res://assets/icons/prayers/05_vision.png")},
	{"name": "Energy", "icon": preload("res://assets/icons/prayers/10_energy.png")},
	{"name": "Holy Light", "icon": preload("res://assets/icons/prayers/12_holy_light.png")},
	{"name": "Melee Boost", "icon": preload("res://assets/icons/prayers/13_melee_boost.png")},
	{"name": "Strength Boost", "icon": preload("res://assets/icons/prayers/15_strength_boost.png")},
	{"name": "Wrath", "icon": preload("res://assets/icons/prayers/19_wrath.png")},
]

const WEAPONS: Dictionary = {
	"sword": SWORD,
	"bronze_sword": BRONZE_SWORD,
	"steel_sword": STEEL_SWORD,
	"mithril_sword": MITHRIL_SWORD,
	"rune_sword": RUNE_SWORD,
	"dagger": DAGGER,
	"bow": BOW,
	"oak_shortbow": OAK_SHORTBOW,
	"willow_shortbow": WILLOW_SHORTBOW,
	"maple_shortbow": MAPLE_SHORTBOW,
	"yew_shortbow": YEW_SHORTBOW,
	"mace": MACE,
	"battle_axe": BATTLE_AXE,
	"wand": WAND,
	"moonspun_wand": MOONSPUN_WAND,
	"staff": STAFF,
	"enchanted_staff": ENCHANTED_STAFF,
	"stardew_staff": STARDEW_STAFF,
	"spear": SPEAR,
	"halberd": HALBERD,
	"crossbow": CROSSBOW,
}

const ITEMS: Dictionary = {
	"sword": SWORD,
	"bronze_sword": BRONZE_SWORD,
	"steel_sword": STEEL_SWORD,
	"mithril_sword": MITHRIL_SWORD,
	"rune_sword": RUNE_SWORD,
	"dagger": DAGGER,
	"bow": BOW,
	"oak_shortbow": OAK_SHORTBOW,
	"willow_shortbow": WILLOW_SHORTBOW,
	"maple_shortbow": MAPLE_SHORTBOW,
	"yew_shortbow": YEW_SHORTBOW,
	"mace": MACE,
	"battle_axe": BATTLE_AXE,
	"wand": WAND,
	"moonspun_wand": MOONSPUN_WAND,
	"staff": STAFF,
	"enchanted_staff": ENCHANTED_STAFF,
	"stardew_staff": STARDEW_STAFF,
	"spear": SPEAR,
	"halberd": HALBERD,
	"crossbow": CROSSBOW,
	"air_rune": preload("res://assets/icons/runes/air_rune.png"),
	"water_rune": preload("res://assets/icons/runes/water_rune.png"),
	"earth_rune": preload("res://assets/icons/runes/earth_rune.png"),
	"fire_rune": preload("res://assets/icons/runes/fire_rune.png"),
	"holy_rune": preload("res://assets/icons/runes/holy_rune.png"),
	"dark_rune": preload("res://assets/icons/runes/death_rune.png"),
	"focus_rune": preload("res://assets/icons/runes/mind_rune.png"),
	"wrath_rune": preload("res://assets/icons/runes/wrath_rune.svg"),
	"hex_rune": preload("res://assets/icons/runes/hex_rune.svg"),
	"nature_rune": preload("res://assets/icons/runes/nature_rune.svg"),
	"aether_rune": preload("res://assets/icons/runes/aether_rune.svg"),
	"spirit_rune": preload("res://assets/icons/runes/spirit_rune.svg"),
	"vital_rune": preload("res://assets/icons/runes/vital_rune.svg"),
	"blank_rune": preload("res://assets/icons/runes/blank_rune.svg"),
	"bronze_arrow": preload("res://assets/icons/ranged/arrows/bronze_arrow_150.png"),
	"iron_arrow": preload("res://assets/icons/ranged/arrows/iron_arrow_150.png"),
	"steel_arrow": preload("res://assets/icons/ranged/arrows/steel_arrow_150.png"),
	"mithril_arrow": preload("res://assets/icons/ranged/arrows/mithril_arrow_150.png"),
	"adamantite_arrow": preload("res://assets/icons/ranged/arrows/adamant_arrow_150.png"),
	"bronze_bolt": preload("res://assets/icons/generated/placeholders/bronze_bolt.svg"),
	"iron_bolt": preload("res://assets/icons/generated/placeholders/iron_bolt.svg"),
	"steel_bolt": preload("res://assets/icons/generated/placeholders/steel_bolt.svg"),
	"mithril_bolt": preload("res://assets/icons/generated/placeholders/mithril_bolt.svg"),
	"adamantite_bolt": preload("res://assets/icons/generated/placeholders/adamantite_bolt.svg"),
	"leather": preload("res://assets/icons/items/materials/leather.svg"),
	"bones": preload("res://assets/icons/generated/placeholders/bones.svg"),
	"arrow_shafts": preload("res://assets/icons/items/fletching/arrow_shafts.svg"),
	"feathers": preload("res://assets/icons/items/fletching/feathers.svg"),
	"iron_arrow_tips": preload("res://assets/icons/generated/placeholders/iron_arrow_tips.svg"),
	"bronze_arrow_tips": preload("res://assets/icons/generated/placeholders/bronze_arrow_tips.svg"),
	"steel_arrow_tips": preload("res://assets/icons/generated/placeholders/steel_arrow_tips.svg"),
	"mithril_arrow_tips": preload("res://assets/icons/generated/placeholders/mithril_arrow_tips.svg"),
	"adamantite_arrow_tips": preload("res://assets/icons/generated/placeholders/adamantite_arrow_tips.svg"),
	"bowstring": preload("res://assets/icons/generated/placeholders/bowstring.svg"),
	"unstrung_shortbow": preload("res://assets/icons/generated/placeholders/unstrung_shortbow.svg"),
	"unstrung_oak_shortbow": preload("res://assets/icons/generated/placeholders/unstrung_oak_shortbow.svg"),
	"unstrung_willow_shortbow": preload("res://assets/icons/generated/placeholders/unstrung_willow_shortbow.svg"),
	"unstrung_maple_shortbow": preload("res://assets/icons/generated/placeholders/unstrung_maple_shortbow.svg"),
	"unstrung_yew_shortbow": preload("res://assets/icons/generated/placeholders/unstrung_yew_shortbow.svg"),
	"raw_minnow": preload("res://assets/icons/items/fish/raw_minnow.svg"),
	"cooked_minnow": preload("res://assets/icons/items/fish/cooked_minnow.svg"),
	"raw_trout": preload("res://assets/icons/items/fish/raw_trout.png"),
	"cooked_trout": preload("res://assets/icons/items/fish/cooked_trout.svg"),
	"raw_salmon": preload("res://assets/icons/items/fish/raw_salmon.png"),
	"cooked_salmon": preload("res://assets/icons/items/fish/cooked_salmon.svg"),
	"raw_tuna": preload("res://assets/icons/items/fish/raw_tuna.png"),
	"cooked_tuna": preload("res://assets/icons/items/fish/cooked_tuna.svg"),
	"raw_lobster": preload("res://assets/icons/items/fish/raw_lobster.svg"),
	"cooked_lobster": preload("res://assets/icons/items/fish/cooked_lobster.png"),
	"burnt_fish": preload("res://assets/icons/items/fish/burnt_fish.svg"),
	"coins": COINS,
	"bronze_pickaxe": BRONZE_PICKAXE,
	"iron_pickaxe": IRON_PICKAXE,
	"steel_pickaxe": STEEL_PICKAXE,
	"mithril_pickaxe": MITHRIL_PICKAXE,
	"bronze_axe": BRONZE_AXE,
	"iron_axe": IRON_AXE,
	"steel_axe": STEEL_AXE,
	"mithril_axe": MITHRIL_AXE,
	"knife": KNIFE,
	"chisel": CHISEL,
	"sewing_needle": SEWING_NEEDLE,
	"thread": THREAD,
	"tinderbox": TINDERBOX,
	"hammer": HAMMER,
	"ring_mould": preload("res://assets/icons/items/tools/ring_mould.svg"),
	"amulet_mould": preload("res://assets/icons/items/tools/amulet_mould.svg"),
	"prayer_potion": PRAYER_POTION,
	"melee_potion": preload("res://assets/icons/items/consumables/melee_potion.svg"),
	"greater_melee_potion": preload("res://assets/icons/items/consumables/greater_melee_potion.svg"),
	"ranging_potion": preload("res://assets/icons/items/consumables/ranging_potion.svg"),
	"greater_ranging_potion": preload("res://assets/icons/items/consumables/greater_ranging_potion.svg"),
	"magic_potion": preload("res://assets/icons/items/consumables/magic_potion.svg"),
	"greater_magic_potion": preload("res://assets/icons/items/consumables/greater_magic_potion.svg"),
	"mending_draught": preload("res://assets/icons/items/consumables/mending_draught.svg"),
	"surefoot_tonic": preload("res://assets/icons/items/consumables/surefoot_tonic.svg"),
	"gatherers_tonic": preload("res://assets/icons/items/consumables/gatherers_tonic.svg"),
	"fishing_rod": FISHING_ROD,
	"lobster_cage": LOBSTER_CAGE,
	"vial_of_water": preload("res://assets/icons/items/consumables/vial_of_water.svg"),
	"flax": preload("res://assets/icons/items/materials/flax.svg"),
	"glowcap": preload("res://assets/icons/items/materials/mushrooms/glowcap.svg"),
	"emberel": preload("res://assets/icons/items/materials/mushrooms/emberel.svg"),
	"ghost_ink": preload("res://assets/icons/items/materials/mushrooms/ghost_ink.svg"),
	"miremane": preload("res://assets/icons/items/materials/mushrooms/miremane.svg"),
	"ironstalk": preload("res://assets/icons/items/materials/mushrooms/ironstalk.svg"),
	"dreamveil": preload("res://assets/icons/items/materials/mushrooms/dreamveil.svg"),
	"black_trumpet": preload("res://assets/icons/items/materials/mushrooms/black_trumpet.svg"),
	"sunleaf": preload("res://assets/icons/generated/placeholders/sunleaf.svg"),
	"moonleaf": preload("res://assets/icons/generated/placeholders/moonleaf.svg"),
	"ashmint": preload("res://assets/icons/items/materials/herbs/ashmint.svg"),
	"ember_pepper": preload("res://assets/icons/items/materials/spices/ember_pepper.svg"),
	"cinnamon_bark": preload("res://assets/icons/items/materials/spices/cinnamon_bark.svg"),
	"star_anise": preload("res://assets/icons/items/materials/spices/star_anise.svg"),
	"golden_saffron": preload("res://assets/icons/items/materials/spices/golden_saffron.svg"),
	"swiftleaf": preload("res://assets/icons/items/materials/herbs/swiftleaf.svg"),
	"bloodfern": preload("res://assets/icons/items/materials/herbs/bloodfern.svg"),
	"silverthorn": preload("res://assets/icons/items/materials/herbs/silverthorn.svg"),
	"stormsage": preload("res://assets/icons/items/materials/herbs/stormsage.svg"),
	"witchbloom": preload("res://assets/icons/items/materials/herbs/witchbloom.svg"),
	"starwort": preload("res://assets/icons/items/materials/herbs/starwort.svg"),
	"whisperweed": preload("res://assets/icons/items/materials/herbs/whisperweed.svg"),
	"sacred_dust": preload("res://assets/icons/items/ingredients/sacred_dust.svg"),
	"fish_oil": preload("res://assets/icons/items/ingredients/fish_oil.svg"),
	"pine_resin": preload("res://assets/icons/items/ingredients/pine_resin.svg"),
	"pearlscale": preload("res://assets/icons/items/ingredients/pearlscale.svg"),
	"hollow_marrow": preload("res://assets/icons/items/ingredients/hollow_marrow.svg"),
	"minor_healing_potion": preload("res://assets/icons/items/consumables/minor_healing_potion.svg"),
	"murky_brew": preload("res://assets/icons/items/consumables/murky_brew.svg"),
	"normal_logs": NORMAL_LOGS,
	"oak_logs": OAK_LOGS,
	"willow_logs": preload("res://assets/icons/logs/willow_logs_150.png"),
	"maple_logs": preload("res://assets/icons/logs/maple_logs_150.png"),
	"yew_logs": preload("res://assets/icons/logs/yew_logs_150.png"),
	"copper_ore": COPPER_ORE,
	"gold_ore": preload("res://assets/icons/items/materials/gold_ore.svg"),
	"uncut_sapphire": preload("res://assets/icons/items/materials/uncut_sapphire.svg"),
	"sapphire": preload("res://assets/icons/items/materials/sapphire.svg"),
	"uncut_emerald": preload("res://assets/icons/items/materials/uncut_emerald.svg"),
	"emerald": preload("res://assets/icons/items/materials/emerald.svg"),
	"uncut_ruby": preload("res://assets/icons/items/materials/uncut_ruby.svg"),
	"ruby": preload("res://assets/icons/items/materials/ruby.svg"),
	"uncut_diamond": preload("res://assets/icons/items/materials/uncut_diamond.svg"),
	"diamond": preload("res://assets/icons/items/materials/diamond.svg"),
	"uncut_opal": preload("res://assets/icons/items/materials/uncut_opal.svg"),
	"opal": preload("res://assets/icons/items/materials/opal.svg"),
	"uncut_topaz": preload("res://assets/icons/items/materials/uncut_topaz.svg"),
	"topaz": preload("res://assets/icons/items/materials/topaz.svg"),
	"uncut_aquamarine": preload("res://assets/icons/items/materials/uncut_aquamarine.svg"),
	"aquamarine": preload("res://assets/icons/items/materials/aquamarine.svg"),
	"uncut_amethyst": preload("res://assets/icons/items/materials/uncut_amethyst.svg"),
	"amethyst": preload("res://assets/icons/items/materials/amethyst.svg"),
	"uncut_onyx": preload("res://assets/icons/items/materials/uncut_onyx.svg"),
	"onyx": preload("res://assets/icons/items/materials/onyx.svg"),
	"tin_ore": TIN_ORE,
	"iron_ore": IRON_ORE,
	"coal": COAL,
	"mithril_ore": MITHRIL_ORE,
	"silver_ore": SILVER_ORE,
	"clay": CLAY,
	"adamantite_ore": ADAMANTITE_ORE,
	"bronze_bar": BRONZE_BAR,
	"iron_bar": IRON_BAR,
	"mithril_bar": MITHRIL_BAR,
	"gold_bar": preload("res://assets/icons/items/materials/gold_bar.svg"),
	"silver_bar": preload("res://assets/icons/bars/Stacked Silver Ingots Icon-3.png"),
	"steel_bar": preload("res://assets/icons/items/materials/steel_bar.svg"),
	"adamantite_bar": preload("res://assets/icons/items/materials/adamantite_bar.svg"),
	"training_helm": TRAINING_HELM,
	"bronze_helm": BRONZE_HELM,
	"bronze_platebody": BRONZE_PLATE,
	"bronze_shield": preload("res://assets/icons/generated/placeholders/bronze_shield.svg"), "bronze_legs": preload("res://assets/icons/generated/placeholders/bronze_legs.svg"), "bronze_gloves": preload("res://assets/icons/generated/placeholders/bronze_gloves.svg"), "bronze_boots": preload("res://assets/icons/generated/placeholders/bronze_boots.svg"),
	"iron_helm": IRON_HELM,
	"iron_platebody": IRON_PLATE,
	"iron_shield": preload("res://assets/icons/generated/placeholders/iron_shield.svg"), "iron_legs": preload("res://assets/icons/generated/placeholders/iron_legs.svg"), "iron_gloves": preload("res://assets/icons/generated/placeholders/iron_gloves.svg"), "iron_boots": preload("res://assets/icons/generated/placeholders/iron_boots.svg"),
	"steel_helm": STEEL_HELM,
	"steel_platebody": STEEL_PLATE,
	"steel_shield": preload("res://assets/icons/generated/placeholders/steel_shield.svg"), "steel_legs": preload("res://assets/icons/generated/placeholders/steel_legs.svg"), "steel_gloves": preload("res://assets/icons/generated/placeholders/steel_gloves.svg"), "steel_boots": preload("res://assets/icons/generated/placeholders/steel_boots.svg"),
	"mithril_helm": MITHRIL_HELM,
	"mithril_platebody": MITHRIL_PLATE,
	"mithril_shield": preload("res://assets/icons/generated/placeholders/mithril_shield.svg"), "mithril_legs": preload("res://assets/icons/generated/placeholders/mithril_legs.svg"), "mithril_gloves": preload("res://assets/icons/generated/placeholders/mithril_gloves.svg"), "mithril_boots": preload("res://assets/icons/generated/placeholders/mithril_boots.svg"),
	"adamantite_helm": ADAMANTITE_HELM,
	"adamantite_platebody": ADAMANTITE_PLATE,
	"adamantite_shield": preload("res://assets/icons/generated/placeholders/adamantite_shield.svg"), "adamantite_legs": preload("res://assets/icons/generated/placeholders/adamantite_legs.svg"), "adamantite_gloves": preload("res://assets/icons/generated/placeholders/adamantite_gloves.svg"), "adamantite_boots": preload("res://assets/icons/generated/placeholders/adamantite_boots.svg"),
	"training_plate": TRAINING_PLATE,
	"swift_helm": SWIFT_HELM,
	"swift_body": SWIFT_BODY,
	"swift_boots": SWIFT_BOOTS,
	"novice_robe_hood": preload("res://assets/icons/generated/placeholders/novice_robe_hood.svg"), "novice_robe_top": preload("res://assets/icons/generated/placeholders/novice_robe_top.svg"), "novice_robe_gloves": preload("res://assets/icons/generated/placeholders/novice_robe_gloves.svg"), "novice_robe_bottoms": preload("res://assets/icons/generated/placeholders/novice_robe_bottoms.svg"), "novice_robe_boots": preload("res://assets/icons/generated/placeholders/novice_robe_boots.svg"),
	"mage_robe_hood": preload("res://assets/icons/generated/placeholders/mage_robe_hood.svg"), "mage_robe_top": preload("res://assets/icons/generated/placeholders/mage_robe_top.svg"), "mage_robe_gloves": preload("res://assets/icons/generated/placeholders/mage_robe_gloves.svg"), "mage_robe_bottoms": preload("res://assets/icons/generated/placeholders/mage_robe_bottoms.svg"), "mage_robe_boots": preload("res://assets/icons/generated/placeholders/mage_robe_boots.svg"),
	"moonspun_hood": preload("res://assets/icons/generated/placeholders/moonspun_hood.svg"), "moonspun_robe_top": preload("res://assets/icons/generated/placeholders/moonspun_robe_top.svg"), "moonspun_gloves": preload("res://assets/icons/generated/placeholders/moonspun_gloves.svg"), "moonspun_bottoms": preload("res://assets/icons/generated/placeholders/moonspun_bottoms.svg"), "moonspun_boots": preload("res://assets/icons/generated/placeholders/moonspun_boots.svg"),
	"enchanted_hood": preload("res://assets/icons/generated/placeholders/enchanted_hood.svg"), "enchanted_robe_top": preload("res://assets/icons/generated/placeholders/enchanted_robe_top.svg"), "enchanted_gloves": preload("res://assets/icons/generated/placeholders/enchanted_gloves.svg"), "enchanted_bottoms": preload("res://assets/icons/generated/placeholders/enchanted_bottoms.svg"), "enchanted_boots": preload("res://assets/icons/generated/placeholders/enchanted_boots.svg"),
	"stardew_hood": preload("res://assets/icons/generated/placeholders/stardew_hood.svg"), "stardew_robe_top": preload("res://assets/icons/generated/placeholders/stardew_robe_top.svg"), "stardew_gloves": preload("res://assets/icons/generated/placeholders/stardew_gloves.svg"), "stardew_bottoms": preload("res://assets/icons/generated/placeholders/stardew_bottoms.svg"), "stardew_boots": preload("res://assets/icons/generated/placeholders/stardew_boots.svg"),
	"novice_ranger_hood": preload("res://assets/icons/generated/placeholders/novice_ranger_hood.svg"), "novice_ranger_body": preload("res://assets/icons/generated/placeholders/novice_ranger_body.svg"), "novice_ranger_gloves": preload("res://assets/icons/generated/placeholders/novice_ranger_gloves.svg"), "novice_ranger_legs": preload("res://assets/icons/generated/placeholders/novice_ranger_legs.svg"), "novice_ranger_boots": preload("res://assets/icons/generated/placeholders/novice_ranger_boots.svg"),
	"scout_hood": preload("res://assets/icons/generated/placeholders/scout_hood.svg"), "scout_body": preload("res://assets/icons/generated/placeholders/scout_body.svg"), "scout_gloves": preload("res://assets/icons/generated/placeholders/scout_gloves.svg"), "scout_legs": preload("res://assets/icons/generated/placeholders/scout_legs.svg"), "scout_boots": preload("res://assets/icons/generated/placeholders/scout_boots.svg"),
	"tracker_hood": preload("res://assets/icons/generated/placeholders/tracker_hood.svg"), "tracker_body": preload("res://assets/icons/generated/placeholders/tracker_body.svg"), "tracker_gloves": preload("res://assets/icons/generated/placeholders/tracker_gloves.svg"), "tracker_legs": preload("res://assets/icons/generated/placeholders/tracker_legs.svg"), "tracker_boots": preload("res://assets/icons/generated/placeholders/tracker_boots.svg"),
	"hunter_hood": preload("res://assets/icons/generated/placeholders/hunter_hood.svg"), "hunter_body": preload("res://assets/icons/generated/placeholders/hunter_body.svg"), "hunter_gloves": preload("res://assets/icons/generated/placeholders/hunter_gloves.svg"), "hunter_legs": preload("res://assets/icons/generated/placeholders/hunter_legs.svg"), "hunter_boots": preload("res://assets/icons/generated/placeholders/hunter_boots.svg"),
	"marksman_hood": preload("res://assets/icons/generated/placeholders/marksman_hood.svg"), "marksman_body": preload("res://assets/icons/generated/placeholders/marksman_body.svg"), "marksman_gloves": preload("res://assets/icons/generated/placeholders/marksman_gloves.svg"), "marksman_legs": preload("res://assets/icons/generated/placeholders/marksman_legs.svg"), "marksman_boots": preload("res://assets/icons/generated/placeholders/marksman_boots.svg"),
	"vanguard_cape": preload("res://assets/icons/equipment/capes/vanguard.svg"),
	"wayfarer_cape": preload("res://assets/icons/equipment/capes/wayfarer.svg"),
	"arcanist_cape": preload("res://assets/icons/equipment/capes/arcanist.svg"),
	"amulet_ferocity": preload("res://assets/icons/equipment/amulets/ferocity.svg"),
	"amulet_hawk": preload("res://assets/icons/equipment/amulets/hawk.svg"),
	"amulet_embers": preload("res://assets/icons/equipment/amulets/embers.svg"),
	"amulet_fury": preload("res://assets/icons/equipment/amulets/fury.svg"),
	"ruby_amulet": preload("res://assets/icons/generated/placeholders/ruby_amulet.svg"),
	"emerald_amulet": preload("res://assets/icons/generated/placeholders/emerald_amulet.svg"),
	"sapphire_amulet": preload("res://assets/icons/generated/placeholders/sapphire_amulet.svg"),
	"diamond_amulet": preload("res://assets/icons/generated/placeholders/diamond_amulet.svg"),
	"standard_spellbook": preload("res://assets/icons/generated/placeholders/standard_spellbook.svg"),
	"skin_bible": preload("res://assets/icons/generated/placeholders/skin_bible.svg"),
	"holy_book": preload("res://assets/icons/generated/placeholders/holy_book.svg"),
	"clerics_tome": preload("res://assets/icons/generated/placeholders/clerics_tome.svg"),
	"ruby_ring": preload("res://assets/icons/generated/placeholders/ruby_ring.svg"),
	"emerald_ring": preload("res://assets/icons/generated/placeholders/emerald_ring.svg"),
	"sapphire_ring": preload("res://assets/icons/generated/placeholders/sapphire_ring.svg"),
	"diamond_ring": preload("res://assets/icons/generated/placeholders/diamond_ring.svg"),
	"ring_of_might": preload("res://assets/icons/generated/placeholders/ring_of_might.svg"),
	"ring_of_marksmanship": preload("res://assets/icons/generated/placeholders/ring_of_marksmanship.svg"),
	"ring_of_sorcery": preload("res://assets/icons/generated/placeholders/ring_of_sorcery.svg"),
	"tyrants_ring": preload("res://assets/icons/generated/placeholders/tyrants_ring.svg"),
	"ranged_tyrants_ring": preload("res://assets/icons/generated/placeholders/ranged_tyrants_ring.svg"),
	"mage_tyrants_ring": preload("res://assets/icons/generated/placeholders/mage_tyrants_ring.svg"),
}

const SKILLS: Dictionary = {
	"Attack": preload("res://assets/icons/skills/attack.png"),
	"Strength": preload("res://assets/icons/skills/strength.png"),
	"Defence": preload("res://assets/icons/skills/defence.png"),
	"Range": preload("res://assets/icons/skills/ranged.png"),
	"Magic": preload("res://assets/icons/skills/magic.png"),
	"Prayer": TAB_PRAYER,
	"Vitality": preload("res://assets/icons/skills/health.png"),
	"Agility": preload("res://assets/icons/skills/agility.png"),
	"Mining": preload("res://assets/icons/skills/mining.png"),
	"Smithing": preload("res://assets/icons/skills/smithing.png"),
	"Crafting": preload("res://assets/icons/skills/crafting.png"),
	"Woodcutting": preload("res://assets/icons/skills/woodcutting.png"),
	"Runecrafting": preload("res://assets/icons/skills/runecrafting.png"),
	"Herblore": preload("res://assets/icons/skills/herblore.png"),
	"Fletching": preload("res://assets/icons/skills/fletching.png"),
	"Fishing": preload("res://assets/icons/skills/fishing.png"),
	"Cooking": preload("res://assets/icons/skills/cooking.png"),
	"Farming": preload("res://assets/icons/skills/farming.png"),
	"Thieving": preload("res://assets/icons/skills/thieving.png"),
	"Construction": TAB_EQUIPMENT,
	"Firemaking": preload("res://assets/icons/skills/firemaking.png"),
	"Hunter": preload("res://assets/icons/skills/hunter.png"),
	"Slayer": preload("res://assets/icons/skills/slayer.png"),
}
