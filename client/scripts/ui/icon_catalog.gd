class_name IconCatalog
extends RefCounted

static var _centered_icon_cache: Dictionary = {}


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
const MACE: Texture2D = preload("res://assets/icons/equipment/weapons/mace.svg")
const BATTLE_AXE: Texture2D = preload("res://assets/icons/equipment/weapons/battle_axe.svg")
const STAFF: Texture2D = preload("res://assets/icons/equipment/weapons/staff.svg")
const SPEAR: Texture2D = SWORD
const HALBERD: Texture2D = BATTLE_AXE
const CROSSBOW: Texture2D = BOW
const TRAINING_HELM: Texture2D = preload("res://assets/icons/equipment/armour/training_helm.svg")
const BRONZE_HELM: Texture2D = preload("res://assets/icons/equipment/armour/bronze_helm.svg")
const BRONZE_PLATE: Texture2D = preload("res://assets/icons/equipment/armour/bronze_platebody.svg")
const IRON_HELM: Texture2D = preload("res://assets/icons/equipment/armour/iron_helm.svg")
const IRON_PLATE: Texture2D = preload("res://assets/icons/equipment/armour/iron_platebody.svg")
const MITHRIL_HELM: Texture2D = preload("res://assets/icons/equipment/armour/mithril_helm.svg")
const MITHRIL_PLATE: Texture2D = preload("res://assets/icons/equipment/armour/mithril_platebody.svg")
const TRAINING_PLATE: Texture2D = preload("res://assets/icons/equipment/armour/training_plate.svg")
const SWIFT_HELM: Texture2D = preload("res://assets/icons/equipment/armour/swift_helm.svg")
const SWIFT_BODY: Texture2D = preload("res://assets/icons/equipment/armour/swift_body.svg")
const SWIFT_BOOTS: Texture2D = preload("res://assets/icons/equipment/armour/swift_boots.svg")
const EMPTY_SLOT: Texture2D = preload("res://assets/icons/items/empty_slot.svg")
const COINS: Texture2D = preload("res://assets/icons/items/coins.png")
const BRONZE_PICKAXE: Texture2D = preload("res://assets/icons/items/bronze_pickaxe.svg")
const BRONZE_AXE: Texture2D = preload("res://assets/icons/items/bronze_axe.svg")
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
}

const PRAYER_PLACEHOLDERS: Array = [
	{"name": "Wisdom", "icon": preload("res://assets/icons/prayers/06_wisdom.png")},
	{"name": "Protection", "icon": preload("res://assets/icons/prayers/02_protection.png")},
	{"name": "Vision", "icon": preload("res://assets/icons/prayers/05_vision.png")},
	{"name": "Melee", "icon": preload("res://assets/icons/prayers/07_melee.png")},
	{"name": "Ranged", "icon": preload("res://assets/icons/prayers/08_ranged.png")},
	{"name": "Magic", "icon": preload("res://assets/icons/prayers/09_magic.png")},
	{"name": "Energy", "icon": preload("res://assets/icons/prayers/10_energy.png")},
	{"name": "Vitality", "icon": preload("res://assets/icons/prayers/11_vitality.png")},
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
	"mace": MACE,
	"battle_axe": BATTLE_AXE,
	"staff": STAFF,
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
	"mace": MACE,
	"battle_axe": BATTLE_AXE,
	"staff": STAFF,
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
	"blank_rune": preload("res://assets/icons/runes/blank_rune.svg"),
	"iron_arrow": preload("res://assets/icons/ranged/arrows/iron_arrow_150.png"),
	"iron_bolt": preload("res://assets/icons/ranged/arrows/iron_arrow_150.png"),
	"leather": preload("res://assets/icons/items/materials/leather.svg"),
	"bones": TAB_PRAYER,
	"arrow_shafts": preload("res://assets/icons/items/fletching/arrow_shafts.svg"),
	"feathers": preload("res://assets/icons/items/fletching/feathers.svg"),
	"iron_arrow_tips": preload("res://assets/icons/items/fletching/iron_arrow_tips.svg"),
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
	"coins": COINS,
	"bronze_pickaxe": BRONZE_PICKAXE,
	"bronze_axe": BRONZE_AXE,
	"knife": KNIFE,
	"chisel": CHISEL,
	"sewing_needle": SEWING_NEEDLE,
	"thread": THREAD,
	"tinderbox": TINDERBOX,
	"hammer": HAMMER,
	"prayer_potion": PRAYER_POTION,
	"fishing_rod": FISHING_ROD,
	"lobster_cage": LOBSTER_CAGE,
	"vial_of_water": preload("res://assets/icons/items/consumables/vial_of_water.svg"),
	"sunleaf": preload("res://assets/icons/items/ingredients/herb.svg"),
	"moonleaf": preload("res://assets/icons/items/ingredients/herb.svg"),
	"sacred_dust": preload("res://assets/icons/items/ingredients/sacred_dust.svg"),
	"fish_oil": preload("res://assets/icons/items/ingredients/fish_oil.svg"),
	"minor_healing_potion": preload("res://assets/icons/items/consumables/minor_healing_potion.svg"),
	"murky_brew": preload("res://assets/icons/items/consumables/murky_brew.svg"),
	"normal_logs": NORMAL_LOGS,
	"oak_logs": OAK_LOGS,
	"willow_logs": NORMAL_LOGS,
	"maple_logs": OAK_LOGS,
	"yew_logs": OAK_LOGS,
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
	"tin_ore": TIN_ORE,
	"iron_ore": IRON_ORE,
	"coal": COAL,
	"mithril_ore": MITHRIL_ORE,
	"bronze_bar": BRONZE_BAR,
	"iron_bar": IRON_BAR,
	"mithril_bar": MITHRIL_BAR,
	"gold_bar": preload("res://assets/icons/items/materials/gold_bar.svg"),
	"training_helm": TRAINING_HELM,
	"bronze_helm": BRONZE_HELM,
	"bronze_platebody": BRONZE_PLATE,
	"bronze_shield": EQUIPMENT_SLOTS["left_hand"], "bronze_legs": EQUIPMENT_SLOTS["legs"], "bronze_gloves": EQUIPMENT_SLOTS["hands"], "bronze_boots": EQUIPMENT_SLOTS["feet"],
	"iron_helm": IRON_HELM,
	"iron_platebody": IRON_PLATE,
	"iron_shield": EQUIPMENT_SLOTS["left_hand"], "iron_legs": EQUIPMENT_SLOTS["legs"], "iron_gloves": EQUIPMENT_SLOTS["hands"], "iron_boots": EQUIPMENT_SLOTS["feet"],
	"mithril_helm": MITHRIL_HELM,
	"mithril_platebody": MITHRIL_PLATE,
	"mithril_shield": EQUIPMENT_SLOTS["left_hand"], "mithril_legs": EQUIPMENT_SLOTS["legs"], "mithril_gloves": EQUIPMENT_SLOTS["hands"], "mithril_boots": EQUIPMENT_SLOTS["feet"],
	"training_plate": TRAINING_PLATE,
	"swift_helm": SWIFT_HELM,
	"swift_body": SWIFT_BODY,
	"swift_boots": SWIFT_BOOTS,
	"novice_robe_hood": TAB_MAGIC, "novice_robe_top": TAB_MAGIC, "novice_robe_gloves": TAB_MAGIC, "novice_robe_bottoms": TAB_MAGIC, "novice_robe_boots": TAB_MAGIC,
	"novice_ranger_hood": BOW, "novice_ranger_body": BOW, "novice_ranger_gloves": BOW, "novice_ranger_legs": BOW, "novice_ranger_boots": BOW,
	"vanguard_cape": preload("res://assets/icons/equipment/capes/vanguard.svg"),
	"wayfarer_cape": preload("res://assets/icons/equipment/capes/wayfarer.svg"),
	"arcanist_cape": preload("res://assets/icons/equipment/capes/arcanist.svg"),
	"amulet_ferocity": preload("res://assets/icons/equipment/amulets/ferocity.svg"),
	"amulet_hawk": preload("res://assets/icons/equipment/amulets/hawk.svg"),
	"amulet_embers": preload("res://assets/icons/equipment/amulets/embers.svg"),
	"amulet_fury": preload("res://assets/icons/equipment/amulets/fury.svg"),
	"ruby_amulet": preload("res://assets/icons/equipment/amulets/ferocity.svg"),
	"emerald_amulet": preload("res://assets/icons/equipment/amulets/hawk.svg"),
	"sapphire_amulet": preload("res://assets/icons/equipment/amulets/embers.svg"),
	"diamond_amulet": preload("res://assets/icons/equipment/amulets/fury.svg"),
	"standard_spellbook": STANDARD_SPELLBOOK,
	"skin_bible": preload("res://assets/icons/spells/miasma.png"),
	"holy_book": preload("res://assets/icons/spells/light.png"),
	"clerics_tome": preload("res://assets/icons/spells/shine.png"),
	"ruby_ring": preload("res://assets/icons/items/materials/ruby.svg"),
	"emerald_ring": preload("res://assets/icons/items/materials/emerald.svg"),
	"sapphire_ring": preload("res://assets/icons/items/materials/sapphire.svg"),
	"diamond_ring": preload("res://assets/icons/items/materials/diamond.svg"),
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
