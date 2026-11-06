class_name IconCatalog
extends RefCounted

# Equipment and item icons.
const SWORD: Texture2D = preload("res://assets/icons/equipment/weapons/sword.svg")
const DAGGER: Texture2D = preload("res://assets/icons/equipment/weapons/dagger.svg")
const BOW: Texture2D = preload("res://assets/icons/equipment/weapons/bow.svg")
const MACE: Texture2D = preload("res://assets/icons/equipment/weapons/mace.svg")
const BATTLE_AXE: Texture2D = preload("res://assets/icons/equipment/weapons/battle_axe.svg")
const STAFF: Texture2D = preload("res://assets/icons/equipment/weapons/staff.svg")
const TRAINING_HELM: Texture2D = preload("res://assets/icons/equipment/armour/training_helm.svg")
const TRAINING_PLATE: Texture2D = preload("res://assets/icons/equipment/armour/training_plate.svg")
const SWIFT_HELM: Texture2D = preload("res://assets/icons/equipment/armour/swift_helm.svg")
const SWIFT_BODY: Texture2D = preload("res://assets/icons/equipment/armour/swift_body.svg")
const SWIFT_BOOTS: Texture2D = preload("res://assets/icons/equipment/armour/swift_boots.svg")
const EMPTY_SLOT: Texture2D = preload("res://assets/icons/items/empty_slot.svg")
const TRAINING_SHARD: Texture2D = preload("res://assets/icons/items/training_shard.svg")
const STANDARD_SPELLBOOK: Texture2D = preload("res://assets/icons/navigation/tab_magic.svg")

const EQUIPMENT_SLOTS: Dictionary = {
	"head": preload("res://assets/icons/equipment/slots/head.svg"),
	"cape": preload("res://assets/icons/equipment/slots/cape.svg"),
	"neck": preload("res://assets/icons/equipment/slots/neck.svg"),
	"ring": preload("res://assets/icons/equipment/slots/ring.svg"),
	"main_hand": preload("res://assets/icons/equipment/slots/main_hand.svg"),
	"body": preload("res://assets/icons/equipment/slots/body.svg"),
	"hands": preload("res://assets/icons/equipment/slots/hands.svg"),
	"legs": preload("res://assets/icons/equipment/slots/legs.svg"),
	"feet": preload("res://assets/icons/equipment/slots/feet.svg"),
	"spellbook": STANDARD_SPELLBOOK,
}

# Main HUD navigation icons.
const TAB_STYLES: Texture2D = preload("res://assets/icons/navigation/tab_styles.svg")
const TAB_INVENTORY: Texture2D = preload("res://assets/icons/navigation/tab_inventory.svg")
const TAB_SKILLS: Texture2D = preload("res://assets/icons/navigation/tab_skills.svg")
const TAB_EQUIPMENT: Texture2D = preload("res://assets/icons/navigation/tab_equipment.svg")
const TAB_MAGIC: Texture2D = preload("res://assets/icons/navigation/tab_magic.svg")
const TAB_PRAYER: Texture2D = preload("res://assets/icons/navigation/tab_prayer.svg")

const SPELLS: Array = [
	{"id": "wind_bolt", "book_id": "standard", "name": "Wind Bolt", "required_magic_level": 1, "cost_item_id": "air_rune", "cost_quantity": 1, "icon": preload("res://assets/icons/spells/wind_blade.svg")},
	{"name": "Gale", "icon": preload("res://assets/icons/spells/gale.svg")},
	{"name": "Tempest", "icon": preload("res://assets/icons/spells/tempest.svg")},
	{"name": "Ember Bolt", "icon": preload("res://assets/icons/spells/ember.svg")},
	{"name": "Inferno", "icon": preload("res://assets/icons/spells/inferno.svg")},
	{"name": "Frost Shard", "icon": preload("res://assets/icons/spells/frost_shard.svg")},
	{"name": "Ice Lance", "icon": preload("res://assets/icons/spells/ice_lance.svg")},
	{"name": "Frost Nova", "icon": preload("res://assets/icons/spells/frost_nova.svg")},
	{"name": "Glacier", "icon": preload("res://assets/icons/spells/glacier.svg")},
	{"name": "Absolute Zero", "icon": preload("res://assets/icons/spells/absolute_zero.svg")},
	{"name": "Pebble", "icon": preload("res://assets/icons/spells/pebble.svg")},
	{"name": "Stone Spike", "icon": preload("res://assets/icons/spells/stone_spike.svg")},
	{"name": "Rock Fall", "icon": preload("res://assets/icons/spells/rock_fall.svg")},
	{"name": "Earthquake", "icon": preload("res://assets/icons/spells/earthquake.svg")},
	{"name": "Mountain Wrath", "icon": preload("res://assets/icons/spells/mountain_wrath.svg")},
	{"name": "Light", "icon": preload("res://assets/icons/spells/light.png")},
	{"name": "Shine", "icon": preload("res://assets/icons/spells/shine.png")},
	{"name": "Divine", "icon": preload("res://assets/icons/spells/divine.png")},
	{"name": "Luce", "icon": preload("res://assets/icons/spells/luce.png")},
	{"name": "Aura", "icon": preload("res://assets/icons/spells/aura.png")},
	{"name": "Bio", "icon": preload("res://assets/icons/spells/bio.png")},
	{"name": "Acid", "icon": preload("res://assets/icons/spells/acid.png")},
	{"name": "Miasma", "icon": preload("res://assets/icons/spells/miasma.png")},
	{"name": "Plague Bloom", "icon": preload("res://assets/icons/spells/plague_bloom.png")},
]

const PRAYERS: Array = [
	{"id": "evasion", "name": "Evasion", "description": "+5% dodge chance\nDrains 1 prayer point every 6 seconds", "icon": preload("res://assets/icons/prayers/01_prayer.png")},
	{"name": "Protection", "icon": preload("res://assets/icons/prayers/02_protection.png")},
	{"id": "strength", "name": "Strength", "description": "+10% melee damage\nDrains 1 prayer point every 6 seconds", "icon": preload("res://assets/icons/prayers/03_strength.png")},
	{"name": "Agility", "icon": preload("res://assets/icons/prayers/04_agility.png")},
	{"name": "Vision", "icon": preload("res://assets/icons/prayers/05_vision.png")},
	{"name": "Wisdom", "icon": preload("res://assets/icons/prayers/06_wisdom.png")},
	{"name": "Melee", "icon": preload("res://assets/icons/prayers/07_melee.png")},
	{"name": "Ranged", "icon": preload("res://assets/icons/prayers/08_ranged.png")},
	{"name": "Magic", "icon": preload("res://assets/icons/prayers/09_magic.png")},
	{"name": "Energy", "icon": preload("res://assets/icons/prayers/10_energy.png")},
	{"name": "Vitality", "icon": preload("res://assets/icons/prayers/11_vitality.png")},
	{"name": "Holy Light", "icon": preload("res://assets/icons/prayers/12_holy_light.png")},
	{"name": "Melee Boost", "icon": preload("res://assets/icons/prayers/13_melee_boost.png")},
	{"id": "defence", "name": "Defence", "description": "10% incoming damage reduction\nDrains 1 prayer point every 6 seconds", "icon": preload("res://assets/icons/prayers/14_defence_boost.png")},
	{"name": "Strength Boost", "icon": preload("res://assets/icons/prayers/15_strength_boost.png")},
	{"name": "Ranged Boost", "icon": preload("res://assets/icons/prayers/16_ranged_boost.png")},
	{"name": "Magic Boost", "icon": preload("res://assets/icons/prayers/17_magic_boost.png")},
	{"id": "accuracy", "name": "Accuracy", "description": "+5 melee accuracy score\nDrains 1 prayer point every 6 seconds", "icon": preload("res://assets/icons/prayers/18_accuracy.png")},
	{"name": "Wrath", "icon": preload("res://assets/icons/prayers/19_wrath.png")},
]

const WEAPONS: Dictionary = {
	"sword": SWORD,
	"steel_sword": SWORD,
	"rune_sword": SWORD,
	"dagger": DAGGER,
	"bow": BOW,
	"mace": MACE,
	"battle_axe": BATTLE_AXE,
	"staff": STAFF,
}

const ITEMS: Dictionary = {
	"sword": SWORD,
	"steel_sword": SWORD,
	"rune_sword": SWORD,
	"dagger": DAGGER,
	"bow": BOW,
	"mace": MACE,
	"battle_axe": BATTLE_AXE,
	"staff": STAFF,
	"air_rune": preload("res://assets/icons/spells/breeze.svg"),
	"training_shard": TRAINING_SHARD,
	"training_helm": TRAINING_HELM,
	"training_plate": TRAINING_PLATE,
	"swift_helm": SWIFT_HELM,
	"swift_body": SWIFT_BODY,
	"swift_boots": SWIFT_BOOTS,
	"standard_spellbook": STANDARD_SPELLBOOK,
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
	"Runecrafting": preload("res://assets/icons/spells/placeholders/spell_frost.svg"),
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
