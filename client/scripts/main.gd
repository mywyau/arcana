extends Node3D

const PLAYER_SCENE: PackedScene = preload("res://scenes/player.tscn")
const TARGET_SCENE: PackedScene = preload("res://scenes/target_dummy.tscn")
const LOOT_SCENE: PackedScene = preload("res://scenes/loot_drop.tscn")
const INVENTORY_SLOT_SCRIPT: Script = preload("res://scripts/inventory_slot.gd")
const RAY_LENGTH: float = 1000.0
const GROUND_COLLISION_MASK: int = 1
const UPPER_COLLISION_MASK: int = 2
const CONNECTOR_COLLISION_MASK: int = 4
const TARGET_COLLISION_MASK: int = 8
const LOOT_COLLISION_MASK: int = 16
const ALL_WORLD_COLLISION_MASK: int = 7
const HUD_LAYOUT_PATH: String = "user://hud_layout.cfg"
const HUD_MINIMUM_SCALE: float = 0.75
const HUD_MAXIMUM_SCALE: float = 1.5
const HUD_HANDLE_HEIGHT: float = 28.0
const HUD_RESIZE_HANDLE_SIZE: float = 30.0
const INVENTORY_DISPLAY_SLOTS: int = 16
const MENU_HUD_BASE_SIZE: Vector2 = Vector2(600.0, 650.0)

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
}
const ACTIVE_SKILL_NAMES: Dictionary = {
	"attack": "Attack",
	"strength": "Strength",
	"defence": "Defence",
	"range": "Range",
	"magic": "Magic",
	"agility": "Agility",
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
const PLACEHOLDER_SPELLS: Array = IconCatalog.SPELLS
const PLACEHOLDER_PRAYERS: Array = IconCatalog.PRAYERS


const COMBAT_STYLES: Dictionary = {
	"sword": [
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
	"bow": [
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
	"training_helm": "head",
	"training_plate": "body",
}
const ITEM_ICONS: Dictionary = IconCatalog.ITEMS
const EQUIPMENT_SLOT_ICONS: Dictionary = IconCatalog.EQUIPMENT_SLOTS
const SKILL_ICONS: Dictionary = IconCatalog.SKILLS

@onready var camera_rig: GuidedCamera = $CameraRig
@onready var camera: Camera3D = $CameraRig/Yaw/Pitch/Camera3D
@onready var ground: PrototypeTerrain = $Ground
@onready var network: NetworkClient = $NetworkClient
@onready var players: Node3D = $Players
@onready var targets: Node3D = $Targets
@onready var drops: Node3D = $Drops
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
var _last_action_sequence: int = 0
var _last_action_result: String = "none"
var _combat_target: String = ""
var _next_attack_pulse: int = 0
var _auto_approaching: bool = false
var _action_pulse: int = 0
var _target_views: Dictionary = {}
var _target_states: Dictionary = {}
var _drop_views: Dictionary = {}
var _player_health: int = 100
var _player_maximum_health: int = 100
var _player_prayer: int = 100
var _player_maximum_prayer: int = 100
var _active_prayers: Array[String] = []
var _last_prayer_sequence: int = 0
var _next_prayer_sequence: int = 1
var _queued_prayer_id: String = ""
var _queued_prayer_active: bool = false
var _selected_spell: String = ""
var _last_spell_sequence: int = 0
var _next_spell_sequence: int = 1
var _queued_spell: String = ""
var _spell_selection_pending: bool = false
var _player_defeated: bool = false
var _player_respawn_at_pulse: int = 0
var _player_dodge_rate: float = 0.1
var _equipped_weapon: String = "sword"
var _last_equipment_sequence: int = 0
var _queued_item: String = ""
var _equipped_armour: Dictionary = {}
var _combat_style: String = "slash"
var _last_combat_style_sequence: int = 0
var _next_combat_style_sequence: int = 1
var _queued_combat_style: String = ""
var _inventory_capacity: int = 8
var _inventory_items: Array = []
var _inventory_signature: String = ""
var _skill_progress: Dictionary = {
	"attack": {"experience": 0, "level": 1, "current": 0, "next": 100},
	"strength": {"experience": 0, "level": 1, "current": 0, "next": 100},
	"defence": {"experience": 0, "level": 1, "current": 0, "next": 100},
	"range": {"experience": 0, "level": 1, "current": 0, "next": 100},
	"magic": {"experience": 0, "level": 1, "current": 0, "next": 100},
	"agility": {"experience": 0, "level": 1, "current": 0, "next": 100},
}
var _skill_tiles: Dictionary = {}
var _combat_stats: Dictionary = {}
var _player_views: Dictionary = {}
var _last_combat_event_id: int = 0
var _combat_events_initialized: bool = false
var _pending_click: Vector2
var _has_pending_click: bool = false
var _target_click_tween: Tween
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
	if not mouse_event.pressed or mouse_event.button_index != MOUSE_BUTTON_LEFT:
		return
	if _player_defeated:
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
			if network.send_engage_target(_next_action_sequence, target_id):
				_next_action_sequence += 1
				var local_view: PlayerView = _player_views.get(_local_player_id)
				if local_view != null:
					local_view.stop_prediction()
				click_marker.visible = false
				_play_target_click_feedback(target_body.global_position)
				_update_combat_hud()
			return

	var loot_result: Dictionary = _top_loot_hit(ray_origin, ray_end)
	if not loot_result.is_empty():
		var loot_body := loot_result.get("collider") as CollisionObject3D
		if loot_body != null and loot_body.has_meta(&"drop_id"):
			var drop_id: String = str(loot_body.get_meta(&"drop_id"))
			if network.send_pickup_item(_next_pickup_sequence, drop_id):
				_next_pickup_sequence += 1
				var local_view: PlayerView = _player_views.get(_local_player_id)
				if local_view != null:
					local_view.stop_prediction()
				if not _combat_target.is_empty() or _last_action_sequence < _next_action_sequence - 1:
					_request_disengage()
				click_marker.visible = false
				_play_target_click_feedback(loot_body.global_position)
			return

	var query := PhysicsRayQueryParameters3D.create(ray_origin, ray_end)
	query.collision_mask = _active_collision_mask()
	var result: Dictionary = get_world_3d().direct_space_state.intersect_ray(query)
	if result.is_empty():
		return

	var clicked_body := result.get("collider") as CollisionObject3D
	var destination: Vector3 = result["position"]
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
			is_equal_approx(distance_squared, best_distance_squared) and drop_id < best_drop_id
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


func _request_equip(item_id: String) -> void:
	if _player_defeated:
		return
	var equipment_pending: bool = _last_equipment_sequence < _next_equipment_sequence - 1
	var is_armour: bool = ARMOUR_SLOTS.has(item_id)
	if item_id == _equipped_weapon and not is_armour and not equipment_pending:
		return
	if network.send_equip_item(_next_equipment_sequence, item_id):
		_queued_item = item_id
		_next_equipment_sequence += 1
		_update_combat_hud()
		_update_inventory_hud()


func _is_item_equipped(item_id: String) -> bool:
	if item_id == _equipped_weapon:
		return true
	var armour_slot: String = str(ARMOUR_SLOTS.get(item_id, ""))
	return not armour_slot.is_empty() and str(_equipped_armour.get(armour_slot, "")) == item_id


func _on_inventory_item_input(event: InputEvent, item_id: String, is_equipable: bool) -> void:
	if not is_equipable or _is_item_equipped(item_id) or _player_defeated:
		return
	if event is InputEventMouseButton:
		var mouse_button := event as InputEventMouseButton
		if mouse_button.pressed and mouse_button.button_index == MOUSE_BUTTON_RIGHT:
			if network.send_drop_item(_next_drop_sequence, item_id):
				_next_drop_sequence += 1
			get_viewport().set_input_as_handled()


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
			var equipped_weapon: String = str(player_data.get("equipped_weapon", "sword"))
			if equipped_weapon != _equipped_weapon:
				_equipped_weapon = equipped_weapon
				_update_equipment_hud()
				inventory_selection_changed = true
				combat_style_changed = true
			var equipped_armour_data: Variant = player_data.get("equipped_armour", {})
			if typeof(equipped_armour_data) == TYPE_DICTIONARY and equipped_armour_data != _equipped_armour:
				_equipped_armour = (equipped_armour_data as Dictionary).duplicate(true)
				_update_equipment_hud()
				inventory_selection_changed = true
			_last_equipment_sequence = int(player_data.get("last_equipment_sequence", 0))
			if _last_equipment_sequence >= _next_equipment_sequence - 1 and not _queued_item.is_empty():
				_queued_item = ""
				inventory_selection_changed = true
			var combat_style: String = str(player_data.get("combat_style", "slash"))
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
		view.set_item_id(
			str(drop_data.get("item_id", "")),
			int(drop_data.get("quantity", 1)),
		)
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


func _setup_hud_layout() -> void:
	var menu_panels: Array[Control] = [
		combat_hud, inventory_hud, skills_hud, equipment_hud, magic_hud, prayer_hud, styles_hud,
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
	styles_combat_tab_button.set_pressed_no_signal(false)
	styles_inventory_tab_button.set_pressed_no_signal(false)
	styles_skills_tab_button.set_pressed_no_signal(false)
	styles_equipment_tab_button.set_pressed_no_signal(false)
	styles_magic_tab_button.set_pressed_no_signal(false)
	styles_prayer_tab_button.set_pressed_no_signal(false)
	styles_active_tab_button.set_pressed_no_signal(true)
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
		else:
			skill_tile.tooltip_text = "%s • Level %d/%d • progression is not active yet" % [
				skill_name, starting_level, starting_level,
			]
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
	var skill_name: String = str(ACTIVE_SKILL_NAMES.get(skill_id, skill_id.capitalize()))
	var experience: int = int(progress.get("experience", 0))
	var level: int = int(progress.get("level", 1))
	var current_level_experience: int = int(progress.get("current", 0))
	var next_level_experience: int = int(progress.get("next", 100))
	level_label.text = "%d/%d" % [level, level]
	var training_description: String = "successful hits grant %s XP" % skill_name
	match skill_id:
		"attack":
			training_description = "Attack techniques grant Attack XP"
		"strength":
			training_description = "Strength techniques grant Strength XP"
		"defence":
			training_description = "Defensive techniques grant Defence XP"
		"range":
			training_description = "successful bow hits grant Range XP"
		"magic":
			training_description = "successful staff hits grant Magic XP"
		"agility":
			training_description = "Agility affects dodge chance • training is not active yet"
	if next_level_experience <= 0:
		tile.tooltip_text = "%s • Level %d/%d • %d XP • Maximum level • %s" % [
			skill_name, level, level, experience, training_description,
		]
		return
	var experience_remaining: int = maxi(0, next_level_experience - experience)
	var level_span: int = maxi(1, next_level_experience - current_level_experience)
	var level_progress: int = clampi(experience - current_level_experience, 0, level_span)
	var progress_percent: int = roundi(float(level_progress) / float(level_span) * 100.0)
	tile.tooltip_text = "%s • Level %d/%d • %d XP • XP equals damage dealt • %s • %d XP to level %d • %d%%" % [
		skill_name,
		level,
		level,
		experience,
		training_description,
		experience_remaining,
		level + 1,
		progress_percent,
	]


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
		var spell_button := Button.new()
		spell_button.custom_minimum_size = Vector2(104, 88)
		spell_button.focus_mode = Control.FOCUS_NONE
		spell_button.toggle_mode = not spell_id.is_empty()
		spell_button.disabled = spell_id.is_empty() or _player_defeated or magic_level < required_level
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
			spell_button.pressed.connect(_request_spell.bind(spell_id))
		magic_grid.add_child(spell_button)


func _populate_placeholder_prayers() -> void:
	_update_prayer_hud()


func _update_prayer_hud() -> void:
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
		prayer_button.icon = prayer.get("icon") as Texture2D
		prayer_button.expand_icon = true
		if prayer_id == "evasion":
			var displayed_active: bool = _active_prayers.has(prayer_id)
			if _queued_prayer_id == prayer_id:
				displayed_active = _queued_prayer_active
			prayer_button.set_pressed_no_signal(displayed_active)
			prayer_button.tooltip_text = "Evasion\n+5% dodge chance\nDrains 1 prayer point every 6 seconds"
			prayer_button.pressed.connect(_request_prayer.bind(prayer_id))
		elif prayer_id == "strength":
			var displayed_active: bool = _active_prayers.has(prayer_id)
			if _queued_prayer_id == prayer_id:
				displayed_active = _queued_prayer_active
			prayer_button.set_pressed_no_signal(displayed_active)
			prayer_button.tooltip_text = "Strength Boost\n+10% dodge chance\nDrains 1 prayer point every 6 seconds"
			prayer_button.pressed.connect(_request_prayer.bind(prayer_id))
		else:
			prayer_button.tooltip_text = "%s (coming later)" % prayer.get("name", "Prayer")
		prayer_grid.add_child(prayer_button)


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
		{"slot": "head", "name": "Head", "item": str(_equipped_armour.get("head", ""))},
		{"spacer": true},
		{"slot": "cape", "name": "Cape", "item": ""},
		{"slot": "neck", "name": "Neck", "item": ""},
		{"slot": "ring", "name": "Ring", "item": ""},
		{"slot": "main_hand", "name": "Main Hand", "item": _equipped_weapon},
		{"slot": "body", "name": "Body", "item": str(_equipped_armour.get("body", ""))},
		{"slot": "hands", "name": "Hands", "item": ""},
		{"spacer": true},
		{"slot": "legs", "name": "Legs", "item": ""},
		{"spacer": true},
		{"spacer": true},
		{"slot": "feet", "name": "Feet", "item": ""},
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
		var slot_id: String = str(slot.get("slot", ""))
		var slot_panel := PanelContainer.new()
		slot_panel.custom_minimum_size = Vector2(168, 82)
		slot_panel.mouse_filter = Control.MOUSE_FILTER_STOP
		slot_panel.tooltip_text = "%s: %s" % [
			slot.get("name", "Equipment"),
			item_id.capitalize() if not item_id.is_empty() else "No item equipped",
		]
		var slot_style := StyleBoxFlat.new()
		slot_style.bg_color = Color(0.055, 0.07, 0.085, 0.94)
		slot_style.border_width_left = 2
		slot_style.border_width_top = 2
		slot_style.border_width_right = 2
		slot_style.border_width_bottom = 2
		slot_style.border_color = Color(0.28, 0.34, 0.4, 0.95)
		slot_style.corner_radius_top_left = 7
		slot_style.corner_radius_top_right = 7
		slot_style.corner_radius_bottom_left = 7
		slot_style.corner_radius_bottom_right = 7
		slot_panel.add_theme_stylebox_override("panel", slot_style)
		var icon := TextureRect.new()
		icon.custom_minimum_size = Vector2(56, 56)
		icon.expand_mode = TextureRect.EXPAND_IGNORE_SIZE
		icon.stretch_mode = TextureRect.STRETCH_KEEP_ASPECT_CENTERED
		icon.texture = (
			ITEM_ICONS.get(item_id) as Texture2D
			if not item_id.is_empty()
			else EQUIPMENT_SLOT_ICONS.get(slot_id, EMPTY_SLOT_ICON) as Texture2D
		)
		icon.mouse_filter = Control.MOUSE_FILTER_IGNORE
		slot_panel.add_child(icon)
		equipment_grid.add_child(slot_panel)


func _show_detailed_combat_stats() -> void:
	_update_detailed_combat_stats()
	combat_stats_hud.visible = true
	combat_stats_close_button.grab_focus()


func _hide_detailed_combat_stats() -> void:
	combat_stats_hud.visible = false
	detailed_stats_button.grab_focus()


func _update_detailed_combat_stats() -> void:
	if _combat_stats.is_empty():
		combat_stats_label.text = "Waiting for the server to send combat stats…"
		return
	var defence_data: Variant = _combat_stats.get("defence_bonuses", {})
	var defence: Dictionary = defence_data if typeof(defence_data) == TYPE_DICTIONARY else {}
	var spell_id: String = str(_combat_stats.get("spell", ""))
	var loadout_text: String = "%s  •  %s style" % [
		str(_combat_stats.get("weapon", "unknown")).capitalize(),
		str(_combat_stats.get("style", "unknown")).capitalize(),
	]
	var accuracy_breakdown: String = "Weapon accuracy: %s\nStyle modifier: %s" % [
		_signed_stat(int(_combat_stats.get("weapon_accuracy_bonus", 0))),
		_signed_stat(int(_combat_stats.get("style_accuracy_bonus", 0))),
	]
	var damage_breakdown: String = "Weapon strength: %s" % _signed_stat(
		int(_combat_stats.get("strength_bonus", 0)),
	)
	if not spell_id.is_empty():
		loadout_text = "%s  •  %s equipped" % [
			spell_id.capitalize(), str(_combat_stats.get("weapon", "unknown")).capitalize(),
		]
		accuracy_breakdown = "Spell accuracy: %s\nEquipment magic accuracy: %s" % [
			_signed_stat(int(_combat_stats.get("spell_accuracy_bonus", 0))),
			_signed_stat(int(_combat_stats.get("weapon_accuracy_bonus", 0))),
		]
		damage_breakdown = "Spell power: %s\nEquipment magic strength: %s" % [
			_signed_stat(int(_combat_stats.get("spell_power", 0))),
			_signed_stat(int(_combat_stats.get("strength_bonus", 0))),
		]
	combat_stats_label.text = (
		"CURRENT LOADOUT\n"
		+"%s\n\n" % loadout_text
		+"ACCURACY\n"
		+"Total accuracy score: %d\n" % int(_combat_stats.get("total_accuracy", 0))
		+"%s level: %d\n" % [
			str(_combat_stats.get("accuracy_skill", "attack")).capitalize(),
			int(_combat_stats.get("accuracy_level", 1)),
		]
		+"%s\n\n" % accuracy_breakdown
		+"DAMAGE\n"
		+"Current damage range: %d–%d\n" % [
			int(_combat_stats.get("minimum_damage", 0)),
			int(_combat_stats.get("maximum_damage", 0)),
		]
		+"%s level: %d\n" % [
			str(_combat_stats.get("damage_skill", "strength")).capitalize(),
			int(_combat_stats.get("damage_level", 1)),
		]
		+"%s\n" % damage_breakdown
		+"Skill damage: +%.1f\n\n" % float(_combat_stats.get("skill_damage_bonus", 0.0))
		+"DEFENCE\n"
		+"Defence level: %d\n" % int(_combat_stats.get("defence_level", 1))
		+"Stab %s    Slash %s    Crush %s\n" % [
			_signed_stat(int(defence.get("stab", 0))),
			_signed_stat(int(defence.get("slash", 0))),
			_signed_stat(int(defence.get("crush", 0))),
		]
		+"Ranged %s    Magic %s\n\n" % [
			_signed_stat(int(defence.get("ranged", 0))),
			_signed_stat(int(defence.get("magic", 0))),
		]
		+"Equipment defence bonuses are informational;\n"
		+"damage mitigation is not active yet."
	)


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
	_update_inventory_hud()
	_update_magic_hud()


func _inventory_quantity(item_id: String) -> int:
	for payload: Variant in _inventory_items:
		if typeof(payload) != TYPE_DICTIONARY:
			continue
		var item: Dictionary = payload
		if str(item.get("item_id", "")) == item_id:
			return int(item.get("quantity", 0))
	return 0


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

	inventory_summary.text = "Inventory  •  %d/%d slots used" % [items_by_slot.size(), _inventory_capacity]
	var displayed_slots: int = maxi(_inventory_capacity, INVENTORY_DISPLAY_SLOTS)
	for slot: int in range(displayed_slots):
		var slot_button: Variant = INVENTORY_SLOT_SCRIPT.new()
		slot_button.custom_minimum_size = Vector2(124, 112)
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
			var is_equipable: bool = EQUIPMENT_ICONS.has(item_id) or ARMOUR_SLOTS.has(item_id)
			var equipment_pending: bool = _last_equipment_sequence < _next_equipment_sequence - 1
			var is_queued: bool = equipment_pending and _queued_item == item_id
			var is_equipped: bool = _is_item_equipped(item_id)
			slot_button.toggle_mode = is_equipable
			slot_button.set_pressed_no_signal(is_queued or is_equipped)
			slot_button.icon = ITEM_ICONS.get(item_id, EMPTY_SLOT_ICON) as Texture2D
			slot_button.expand_icon = true
			var item_status: String = ""
			if is_queued:
				item_status = "Queued"
			elif is_equipped:
				item_status = "Equipped"
			var quantity_text: String = " ×%d" % quantity if quantity > 1 else ""
			var status_text: String = " • %s" % item_status if not item_status.is_empty() else ""
			var equip_action: String = "left-click to unequip" if is_equipped and ARMOUR_SLOTS.has(item_id) else "left-click to equip"
			slot_button.tooltip_text = "%s%s%s • slot %d%s" % [
				item_id.capitalize(), quantity_text, status_text,
				slot + 1,
				" — %s • right-click to drop when unequipped" % equip_action if is_equipable else "",
			]
			slot_button.disabled = _player_defeated and is_equipable
			if is_equipped or is_queued:
				var selection_style := StyleBoxFlat.new()
				selection_style.bg_color = Color(0.06, 0.09, 0.12, 0.96)
				selection_style.border_width_left = 4
				selection_style.border_width_top = 4
				selection_style.border_width_right = 4
				selection_style.border_width_bottom = 4
				selection_style.border_color = Color(0.35, 0.7, 1.0) if is_queued else Color(0.95, 0.72, 0.2)
				selection_style.corner_radius_top_left = 6
				selection_style.corner_radius_top_right = 6
				selection_style.corner_radius_bottom_left = 6
				selection_style.corner_radius_bottom_right = 6
				slot_button.add_theme_stylebox_override("pressed", selection_style)
				slot_button.add_theme_stylebox_override("hover_pressed", selection_style)
			if is_equipable:
				slot_button.pressed.connect(_request_equip.bind(item_id))
				slot_button.gui_input.connect(_on_inventory_item_input.bind(item_id, is_equipable))
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
	target_health_bar.max_value = maxi(1, displayed_maximum_health)
	target_health_bar.value = displayed_health
	target_health_label.text = "%s  %d/%d" % [target_name, displayed_health, displayed_maximum_health]

	var has_aggro: bool = displayed_target_id == _local_player_id
	aggro_label.text = "ENEMY AGGRO" if has_aggro else "No aggro"
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

	var weapon_interval: int = 4
	if _equipped_weapon == "dagger":
		weapon_interval = 3
	elif _equipped_weapon == "bow":
		weapon_interval = 4 if _combat_style == "rapid" else 5
	elif _equipped_weapon == "mace" or _equipped_weapon == "staff":
		weapon_interval = 5
	elif _equipped_weapon == "battle_axe":
		weapon_interval = 6
	if not _selected_spell.is_empty():
		weapon_interval = 5
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
