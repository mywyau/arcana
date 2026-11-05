extends Node3D

const PLAYER_SCENE: PackedScene = preload("res://scenes/player.tscn")
const TARGET_SCENE: PackedScene = preload("res://scenes/target_dummy.tscn")
const LOOT_SCENE: PackedScene = preload("res://scenes/loot_drop.tscn")
const RAY_LENGTH: float = 1000.0
const GROUND_COLLISION_MASK: int = 1
const UPPER_COLLISION_MASK: int = 2
const CONNECTOR_COLLISION_MASK: int = 4
const TARGET_COLLISION_MASK: int = 8
const LOOT_COLLISION_MASK: int = 16
const ALL_WORLD_COLLISION_MASK: int = 7

const PLACEHOLDER_SKILLS: Array = [
	{"name": "Attack", "level": 1},
	{"name": "Magic", "level": 1},
	{"name": "Range", "level": 1},
	{"name": "Defence", "level": 1},
	{"name": "Strength", "level": 1},
	{"name": "Prayer", "level": 1},
	{"name": "Vitality", "level": 1},
	{"name": "Mana", "level": 1},
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
}
const ACTIVE_SKILL_NAMES: Dictionary = {
	"attack": "Attack",
	"strength": "Strength",
	"defence": "Defence",
	"range": "Range",
	"magic": "Magic",
}

const SWORD_ICON: Texture2D = preload("res://assets/icons/sword.svg")
const DAGGER_ICON: Texture2D = preload("res://assets/icons/dagger.svg")
const BOW_ICON: Texture2D = preload("res://assets/icons/bow.svg")
const MACE_ICON: Texture2D = preload("res://assets/icons/mace.svg")
const BATTLE_AXE_ICON: Texture2D = preload("res://assets/icons/battle_axe.svg")
const STAFF_ICON: Texture2D = preload("res://assets/icons/staff.svg")
const EMPTY_SLOT_ICON: Texture2D = preload("res://assets/icons/empty_slot.svg")
const TRAINING_SHARD_ICON: Texture2D = preload("res://assets/icons/training_shard.svg")
const PLACEHOLDER_SPELLS: Array = [
	{"name": "Ember Bolt", "icon": preload("res://assets/icons/spell_ember.svg")},
	{"name": "Frost Bind", "icon": preload("res://assets/icons/spell_frost.svg")},
	{"name": "Arcane Ward", "icon": preload("res://assets/icons/spell_ward.svg")},
	{"name": "Blink", "icon": preload("res://assets/icons/spell_blink.svg")},
]
const COMBAT_STYLES: Dictionary = {
	"sword": [
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
const EQUIPMENT_ICONS: Dictionary = {
	"sword": SWORD_ICON,
	"dagger": DAGGER_ICON,
	"bow": BOW_ICON,
	"mace": MACE_ICON,
	"battle_axe": BATTLE_AXE_ICON,
	"staff": STAFF_ICON,
}
const ITEM_ICONS: Dictionary = {
	"sword": SWORD_ICON,
	"dagger": DAGGER_ICON,
	"bow": BOW_ICON,
	"mace": MACE_ICON,
	"battle_axe": BATTLE_AXE_ICON,
	"staff": STAFF_ICON,
	"training_shard": TRAINING_SHARD_ICON,
}

@onready var camera_rig: GuidedCamera = $CameraRig
@onready var camera: Camera3D = $CameraRig/Yaw/Pitch/Camera3D
@onready var ground: PrototypeTerrain = $Ground
@onready var network: NetworkClient = $NetworkClient
@onready var players: Node3D = $Players
@onready var targets: Node3D = $Targets
@onready var drops: Node3D = $Drops
@onready var status_label: Label = $UI/Margin/Status
@onready var diagnostics_label: Label = $UI/DiagnosticsMargin/Diagnostics
@onready var combat_hud: PanelContainer = $UI/CombatHUD
@onready var inventory_hud: PanelContainer = $UI/InventoryHUD
@onready var skills_hud: PanelContainer = $UI/SkillsHUD
@onready var equipment_hud: PanelContainer = $UI/EquipmentHUD
@onready var magic_hud: PanelContainer = $UI/MagicHUD
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
@onready var inventory_styles_tab_button: Button = $UI/InventoryHUD/Margin/VBox/TabRow/Styles
@onready var inventory_summary: Label = $UI/InventoryHUD/Margin/VBox/Summary
@onready var inventory_grid: GridContainer = $UI/InventoryHUD/Margin/VBox/Grid
@onready var skills_combat_tab_button: Button = $UI/SkillsHUD/Margin/VBox/TabRow/Combat
@onready var skills_inventory_tab_button: Button = $UI/SkillsHUD/Margin/VBox/TabRow/Inventory
@onready var skills_active_tab_button: Button = $UI/SkillsHUD/Margin/VBox/TabRow/Skills
@onready var skills_equipment_tab_button: Button = $UI/SkillsHUD/Margin/VBox/TabRow/Equipment
@onready var skills_magic_tab_button: Button = $UI/SkillsHUD/Margin/VBox/TabRow/Magic
@onready var skills_styles_tab_button: Button = $UI/SkillsHUD/Margin/VBox/TabRow/Styles
@onready var skills_summary: Label = $UI/SkillsHUD/Margin/VBox/Summary
@onready var skills_grid: GridContainer = $UI/SkillsHUD/Margin/VBox/Grid
@onready var equipment_combat_tab_button: Button = $UI/EquipmentHUD/Margin/VBox/TabRow/Combat
@onready var equipment_inventory_tab_button: Button = $UI/EquipmentHUD/Margin/VBox/TabRow/Inventory
@onready var equipment_skills_tab_button: Button = $UI/EquipmentHUD/Margin/VBox/TabRow/Skills
@onready var equipment_active_tab_button: Button = $UI/EquipmentHUD/Margin/VBox/TabRow/Equipment
@onready var equipment_magic_tab_button: Button = $UI/EquipmentHUD/Margin/VBox/TabRow/Magic
@onready var equipment_styles_tab_button: Button = $UI/EquipmentHUD/Margin/VBox/TabRow/Styles
@onready var equipment_grid: GridContainer = $UI/EquipmentHUD/Margin/VBox/Grid
@onready var magic_combat_tab_button: Button = $UI/MagicHUD/Margin/VBox/TabRow/Combat
@onready var magic_inventory_tab_button: Button = $UI/MagicHUD/Margin/VBox/TabRow/Inventory
@onready var magic_skills_tab_button: Button = $UI/MagicHUD/Margin/VBox/TabRow/Skills
@onready var magic_equipment_tab_button: Button = $UI/MagicHUD/Margin/VBox/TabRow/Equipment
@onready var magic_active_tab_button: Button = $UI/MagicHUD/Margin/VBox/TabRow/Magic
@onready var magic_styles_tab_button: Button = $UI/MagicHUD/Margin/VBox/TabRow/Styles
@onready var magic_grid: GridContainer = $UI/MagicHUD/Margin/VBox/Grid
@onready var styles_combat_tab_button: Button = $UI/StylesHUD/Margin/VBox/TabRow/Combat
@onready var styles_inventory_tab_button: Button = $UI/StylesHUD/Margin/VBox/TabRow/Inventory
@onready var styles_skills_tab_button: Button = $UI/StylesHUD/Margin/VBox/TabRow/Skills
@onready var styles_equipment_tab_button: Button = $UI/StylesHUD/Margin/VBox/TabRow/Equipment
@onready var styles_magic_tab_button: Button = $UI/StylesHUD/Margin/VBox/TabRow/Magic
@onready var styles_active_tab_button: Button = $UI/StylesHUD/Margin/VBox/TabRow/Styles
@onready var styles_summary: Label = $UI/StylesHUD/Margin/VBox/Summary
@onready var styles_grid: GridContainer = $UI/StylesHUD/Margin/VBox/Grid
@onready var player_health_bar: ProgressBar = $UI/CombatHUD/Margin/VBox/PlayerHealth
@onready var player_health_label: Label = $UI/CombatHUD/Margin/VBox/PlayerRow/PlayerHealthLabel
@onready var target_health_bar: ProgressBar = $UI/CombatHUD/Margin/VBox/TargetHealth
@onready var target_health_label: Label = $UI/CombatHUD/Margin/VBox/TargetRow/TargetHealthLabel
@onready var aggro_label: Label = $UI/CombatHUD/Margin/VBox/TargetRow/AggroLabel
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
var _player_defeated: bool = false
var _player_respawn_at_pulse: int = 0
var _player_dodge_rate: float = 0.1
var _equipped_weapon: String = "sword"
var _last_equipment_sequence: int = 0
var _queued_weapon: String = ""
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
}
var _skill_tiles: Dictionary = {}
var _player_views: Dictionary = {}
var _last_combat_event_id: int = 0
var _combat_events_initialized: bool = false
var _pending_click: Vector2
var _has_pending_click: bool = false
var _target_click_tween: Tween


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
	inventory_styles_tab_button.pressed.connect(_show_styles_tab)
	skills_combat_tab_button.pressed.connect(_show_combat_tab)
	skills_inventory_tab_button.pressed.connect(_show_inventory_tab)
	skills_active_tab_button.pressed.connect(_show_skills_tab)
	skills_equipment_tab_button.pressed.connect(_show_equipment_tab)
	skills_magic_tab_button.pressed.connect(_show_magic_tab)
	skills_styles_tab_button.pressed.connect(_show_styles_tab)
	equipment_combat_tab_button.pressed.connect(_show_combat_tab)
	equipment_inventory_tab_button.pressed.connect(_show_inventory_tab)
	equipment_skills_tab_button.pressed.connect(_show_skills_tab)
	equipment_active_tab_button.pressed.connect(_show_equipment_tab)
	equipment_magic_tab_button.pressed.connect(_show_magic_tab)
	equipment_styles_tab_button.pressed.connect(_show_styles_tab)
	magic_combat_tab_button.pressed.connect(_show_combat_tab)
	magic_inventory_tab_button.pressed.connect(_show_inventory_tab)
	magic_skills_tab_button.pressed.connect(_show_skills_tab)
	magic_equipment_tab_button.pressed.connect(_show_equipment_tab)
	magic_active_tab_button.pressed.connect(_show_magic_tab)
	magic_styles_tab_button.pressed.connect(_show_styles_tab)
	styles_combat_tab_button.pressed.connect(_show_combat_tab)
	styles_inventory_tab_button.pressed.connect(_show_inventory_tab)
	styles_skills_tab_button.pressed.connect(_show_skills_tab)
	styles_equipment_tab_button.pressed.connect(_show_equipment_tab)
	styles_magic_tab_button.pressed.connect(_show_magic_tab)
	styles_active_tab_button.pressed.connect(_show_styles_tab)
	_update_combat_hud()
	_update_inventory_hud()
	_populate_placeholder_skills()
	_populate_placeholder_spells()
	_update_combat_styles_hud()
	_update_equipment_hud()


func _unhandled_input(event: InputEvent) -> void:
	if event is InputEventKey:
		var key_event := event as InputEventKey
		if key_event.pressed and not key_event.echo:
			match key_event.keycode:
				KEY_ESCAPE:
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

	var loot_query := PhysicsRayQueryParameters3D.create(ray_origin, ray_end)
	loot_query.collision_mask = LOOT_COLLISION_MASK
	var loot_result: Dictionary = get_world_3d().direct_space_state.intersect_ray(loot_query)
	if not loot_result.is_empty():
		var loot_body := loot_result.get("collider") as CollisionObject3D
		if loot_body != null and loot_body.has_meta(&"drop_id"):
			var drop_id: String = str(loot_body.get_meta(&"drop_id"))
			if network.send_pickup_item(_next_pickup_sequence, drop_id):
				_next_pickup_sequence += 1
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
	if item_id == _equipped_weapon and not equipment_pending:
		return
	if network.send_equip_item(_next_equipment_sequence, item_id):
		_queued_weapon = item_id
		_next_equipment_sequence += 1
		_update_combat_hud()
		_update_inventory_hud()


func _request_combat_style(combat_style: String) -> void:
	if _player_defeated:
		return
	var style_pending: bool = _last_combat_style_sequence < _next_combat_style_sequence - 1
	if combat_style == _combat_style and not style_pending:
		return
	if network.send_combat_style(_next_combat_style_sequence, combat_style):
		_queued_combat_style = combat_style
		_next_combat_style_sequence += 1
		_update_combat_styles_hud()


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
		var server_driven_path: bool = bool(player_data.get("auto_approaching", false))
		var correction: float = view.apply_authoritative_state(
			position,
			path,
			float(player_data.get("speed", 4.0)),
			acknowledged_sequence,
			_next_sequence - 1,
			server_driven_path,
		)
		view.set_weapon(str(player_data.get("equipped_weapon", "sword")))
		view.set_defeated(bool(player_data.get("defeated", false)))
		if id == _local_player_id:
			var inventory_selection_changed: bool = false
			var combat_style_changed: bool = false
			_last_action_sequence = int(player_data.get("last_action_sequence", 0))
			_last_action_result = str(player_data.get("last_action_result", "none"))
			_combat_target = str(player_data.get("combat_target", ""))
			_next_attack_pulse = int(player_data.get("next_attack_pulse", 0))
			_auto_approaching = bool(player_data.get("auto_approaching", false))
			_player_health = int(player_data.get("health", _player_health))
			_player_maximum_health = int(player_data.get("maximum_health", _player_maximum_health))
			_player_defeated = bool(player_data.get("defeated", false))
			_player_respawn_at_pulse = int(player_data.get("respawn_at_pulse", 0))
			_player_dodge_rate = float(player_data.get("dodge_rate", _player_dodge_rate))
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
			_last_equipment_sequence = int(player_data.get("last_equipment_sequence", 0))
			if _last_equipment_sequence >= _next_equipment_sequence - 1 and not _queued_weapon.is_empty():
				_queued_weapon = ""
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
		var source: PlayerView = _player_views.get(source_id)
		var target: PlayerView = _player_views.get(target_id)
		var dummy_view: TargetDummy = _dummy_view_for_id(target_id)
		if dummy_view != null:
			if source == null:
				continue
			var impact_delay: float = source.play_weapon_attack(dummy_view.global_position)
			_play_dummy_impact_after_delay(
				dummy_view,
				event_type,
				amount,
				source.global_position,
				impact_delay,
			)
		elif target != null:
			if source != null:
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
	if code == "command_rejected":
		var local_view: PlayerView = _player_views.get(_local_player_id)
		if local_view != null:
			local_view.cancel_prediction()


func _show_combat_tab() -> void:
	combat_hud.visible = true
	inventory_hud.visible = false
	skills_hud.visible = false
	equipment_hud.visible = false
	magic_hud.visible = false
	styles_hud.visible = false
	combat_tab_button.set_pressed_no_signal(true)
	inventory_tab_button.set_pressed_no_signal(false)
	combat_skills_tab_button.set_pressed_no_signal(false)
	combat_equipment_tab_button.set_pressed_no_signal(false)
	combat_magic_tab_button.set_pressed_no_signal(false)
	combat_styles_tab_button.set_pressed_no_signal(false)


func _show_inventory_tab() -> void:
	combat_hud.visible = false
	inventory_hud.visible = true
	skills_hud.visible = false
	equipment_hud.visible = false
	magic_hud.visible = false
	styles_hud.visible = false
	inventory_combat_tab_button.set_pressed_no_signal(false)
	inventory_active_tab_button.set_pressed_no_signal(true)
	inventory_skills_tab_button.set_pressed_no_signal(false)
	inventory_equipment_tab_button.set_pressed_no_signal(false)
	inventory_magic_tab_button.set_pressed_no_signal(false)
	inventory_styles_tab_button.set_pressed_no_signal(false)


func _show_skills_tab() -> void:
	combat_hud.visible = false
	inventory_hud.visible = false
	skills_hud.visible = true
	equipment_hud.visible = false
	magic_hud.visible = false
	styles_hud.visible = false
	skills_combat_tab_button.set_pressed_no_signal(false)
	skills_inventory_tab_button.set_pressed_no_signal(false)
	skills_active_tab_button.set_pressed_no_signal(true)
	skills_equipment_tab_button.set_pressed_no_signal(false)
	skills_magic_tab_button.set_pressed_no_signal(false)
	skills_styles_tab_button.set_pressed_no_signal(false)


func _show_equipment_tab() -> void:
	combat_hud.visible = false
	inventory_hud.visible = false
	skills_hud.visible = false
	equipment_hud.visible = true
	magic_hud.visible = false
	styles_hud.visible = false
	equipment_combat_tab_button.set_pressed_no_signal(false)
	equipment_inventory_tab_button.set_pressed_no_signal(false)
	equipment_skills_tab_button.set_pressed_no_signal(false)
	equipment_active_tab_button.set_pressed_no_signal(true)
	equipment_magic_tab_button.set_pressed_no_signal(false)
	equipment_styles_tab_button.set_pressed_no_signal(false)


func _show_magic_tab() -> void:
	combat_hud.visible = false
	inventory_hud.visible = false
	skills_hud.visible = false
	equipment_hud.visible = false
	magic_hud.visible = true
	styles_hud.visible = false
	magic_combat_tab_button.set_pressed_no_signal(false)
	magic_inventory_tab_button.set_pressed_no_signal(false)
	magic_skills_tab_button.set_pressed_no_signal(false)
	magic_equipment_tab_button.set_pressed_no_signal(false)
	magic_active_tab_button.set_pressed_no_signal(true)
	magic_styles_tab_button.set_pressed_no_signal(false)


func _show_styles_tab() -> void:
	combat_hud.visible = false
	inventory_hud.visible = false
	skills_hud.visible = false
	equipment_hud.visible = false
	magic_hud.visible = false
	styles_hud.visible = true
	styles_combat_tab_button.set_pressed_no_signal(false)
	styles_inventory_tab_button.set_pressed_no_signal(false)
	styles_skills_tab_button.set_pressed_no_signal(false)
	styles_equipment_tab_button.set_pressed_no_signal(false)
	styles_magic_tab_button.set_pressed_no_signal(false)
	styles_active_tab_button.set_pressed_no_signal(true)


func _populate_placeholder_skills() -> void:
	for skill_data: Variant in PLACEHOLDER_SKILLS:
		var skill: Dictionary = skill_data
		var skill_name: String = str(skill.get("name", "Skill"))
		var skill_tile := Button.new()
		skill_tile.custom_minimum_size = Vector2(145, 72)
		skill_tile.focus_mode = Control.FOCUS_NONE
		skill_tile.mouse_filter = Control.MOUSE_FILTER_IGNORE
		skill_tile.text = "%s\nLevel %d" % [skill_name, skill.get("level", 1)]
		var skill_id: String = str(ACTIVE_SKILL_IDS.get(skill_name, ""))
		if not skill_id.is_empty():
			_skill_tiles[skill_id] = skill_tile
		else:
			skill_tile.tooltip_text = "Placeholder skill — progression is not active yet"
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


func _update_active_skills_hud() -> void:
	for skill_id: String in ACTIVE_SKILL_NAMES:
		_update_progression_skill_tile(skill_id)
	var attack: Dictionary = _skill_progress["attack"] as Dictionary
	var strength: Dictionary = _skill_progress["strength"] as Dictionary
	var defence: Dictionary = _skill_progress["defence"] as Dictionary
	var ranged: Dictionary = _skill_progress["range"] as Dictionary
	var magic: Dictionary = _skill_progress["magic"] as Dictionary
	skills_summary.text = "Attack %d  •  Strength %d  •  Defence %d\nRange %d  •  Magic %d" % [
		int(attack.get("level", 1)),
		int(strength.get("level", 1)),
		int(defence.get("level", 1)),
		int(ranged.get("level", 1)),
		int(magic.get("level", 1)),
	]


func _update_progression_skill_tile(skill_id: String) -> void:
	var skill_tile: Button = _skill_tiles.get(skill_id) as Button
	if skill_tile == null:
		return
	var progress: Dictionary = _skill_progress[skill_id] as Dictionary
	var skill_name: String = str(ACTIVE_SKILL_NAMES.get(skill_id, skill_id.capitalize()))
	var experience: int = int(progress.get("experience", 0))
	var level: int = int(progress.get("level", 1))
	var current_level_experience: int = int(progress.get("current", 0))
	var next_level_experience: int = int(progress.get("next", 100))
	skill_tile.text = "%s\nLevel %d  •  %d XP" % [skill_name, level, experience]
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
	if next_level_experience <= 0:
		skill_tile.tooltip_text = "Maximum level • %s" % training_description
		return
	var experience_remaining: int = maxi(0, next_level_experience - experience)
	var level_span: int = maxi(1, next_level_experience - current_level_experience)
	var level_progress: int = experience - current_level_experience
	var progress_percent: int = roundi(float(level_progress) / float(level_span) * 100.0)
	skill_tile.tooltip_text = "XP equals damage dealt • %s • %d XP to level %d • %d%%" % [
		training_description,
		experience_remaining,
		level + 1,
		progress_percent,
	]


func _populate_placeholder_spells() -> void:
	for spell_data: Variant in PLACEHOLDER_SPELLS:
		var spell: Dictionary = spell_data
		var spell_button := Button.new()
		spell_button.custom_minimum_size = Vector2(104, 88)
		spell_button.focus_mode = Control.FOCUS_NONE
		spell_button.disabled = true
		spell_button.icon = spell.get("icon") as Texture2D
		spell_button.expand_icon = true
		spell_button.tooltip_text = "%s — placeholder spell" % spell.get("name", "Spell")
		magic_grid.add_child(spell_button)


func _update_combat_styles_hud() -> void:
	for child: Node in styles_grid.get_children():
		styles_grid.remove_child(child)
		child.queue_free()

	var style_options: Array = COMBAT_STYLES.get(_equipped_weapon, []) as Array
	var selected_style: String = _queued_combat_style if not _queued_combat_style.is_empty() else _combat_style
	var selected_training: String = ""
	for style_data: Variant in style_options:
		var candidate: Dictionary = style_data
		if str(candidate.get("id", "")) == selected_style:
			selected_training = str(candidate.get("training", ""))
			break
	if selected_style == "unfocused":
		styles_summary.text = "%s  •  COMBAT XP LOCKED" % _equipped_weapon.capitalize()
	elif _equipped_weapon == "bow":
		styles_summary.text = "Bow  •  trains %s  •  Rapid: faster  •  Long Range: 14 units" % selected_training
	else:
		styles_summary.text = "%s  •  selected technique trains %s" % [
			_equipped_weapon.capitalize(),
			selected_training,
		]
	for style_data: Variant in style_options:
		var style: Dictionary = style_data
		var style_id: String = str(style.get("id", ""))
		var training: String = str(style.get("training", ""))
		var style_button := Button.new()
		style_button.custom_minimum_size = Vector2(145, 64)
		style_button.toggle_mode = true
		style_button.disabled = _player_defeated
		style_button.text = "%s\n%s" % [
			style.get("name", style_id.capitalize()),
			"No XP" if training.is_empty() else "%s XP" % training,
		]
		style_button.tooltip_text = "Combat XP is locked" if training.is_empty() else "Successful hits train %s" % training
		style_button.set_pressed_no_signal(style_id == selected_style)
		style_button.pressed.connect(_request_combat_style.bind(style_id))
		styles_grid.add_child(style_button)


func _update_equipment_hud() -> void:
	for child: Node in equipment_grid.get_children():
		equipment_grid.remove_child(child)
		child.queue_free()

	var equipment_slots: Array = [
		{"name": "Main Hand", "item": _equipped_weapon},
		{"name": "Head", "item": ""},
		{"name": "Body", "item": ""},
		{"name": "Hands", "item": ""},
		{"name": "Legs", "item": ""},
		{"name": "Trinket", "item": ""},
	]
	for slot_data: Variant in equipment_slots:
		var slot: Dictionary = slot_data
		var item_id: String = str(slot.get("item", ""))
		var tile := VBoxContainer.new()
		tile.custom_minimum_size = Vector2(145, 92)
		tile.alignment = BoxContainer.ALIGNMENT_CENTER
		tile.mouse_filter = Control.MOUSE_FILTER_IGNORE
		var icon := TextureRect.new()
		icon.custom_minimum_size = Vector2(48, 58)
		icon.expand_mode = TextureRect.EXPAND_IGNORE_SIZE
		icon.stretch_mode = TextureRect.STRETCH_KEEP_ASPECT_CENTERED
		icon.texture = EQUIPMENT_ICONS.get(item_id, EMPTY_SLOT_ICON) as Texture2D
		icon.mouse_filter = Control.MOUSE_FILTER_IGNORE
		tile.add_child(icon)
		var label := Label.new()
		label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
		label.text = "%s\n%s" % [
			slot.get("name", "Slot"),
			item_id.capitalize() if not item_id.is_empty() else "Empty",
		]
		label.mouse_filter = Control.MOUSE_FILTER_IGNORE
		tile.add_child(label)
		equipment_grid.add_child(tile)


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
	for slot: int in range(_inventory_capacity):
		var slot_button := Button.new()
		slot_button.custom_minimum_size = Vector2(108, 72)
		slot_button.focus_mode = Control.FOCUS_NONE
		slot_button.mouse_filter = Control.MOUSE_FILTER_STOP
		if items_by_slot.has(slot):
			var item: Dictionary = items_by_slot[slot]
			var item_id: String = str(item.get("item_id", ""))
			var quantity: int = int(item.get("quantity", 1))
			var is_weapon: bool = EQUIPMENT_ICONS.has(item_id)
			var equipment_pending: bool = _last_equipment_sequence < _next_equipment_sequence - 1
			var is_queued: bool = equipment_pending and _queued_weapon == item_id
			var is_equipped: bool = not equipment_pending and _equipped_weapon == item_id
			slot_button.toggle_mode = is_weapon
			slot_button.set_pressed_no_signal(is_queued or is_equipped)
			slot_button.icon = ITEM_ICONS.get(item_id, EMPTY_SLOT_ICON) as Texture2D
			slot_button.expand_icon = true
			var item_status: String = ""
			if is_queued:
				item_status = "Queued"
			elif is_equipped:
				item_status = "Equipped"
			elif quantity > 1:
				item_status = "×%d" % quantity
			slot_button.text = "%s%s" % [
				item_id.capitalize(),
				"\n%s" % item_status if not item_status.is_empty() else "",
			]
			slot_button.tooltip_text = "%s — slot %d%s" % [
				item_id.capitalize(),
				slot + 1,
				" — click to equip" if is_weapon else "",
			]
			slot_button.disabled = _player_defeated or not is_weapon
			if is_weapon:
				slot_button.pressed.connect(_request_equip.bind(item_id))
		else:
			slot_button.text = "Empty"
			slot_button.icon = EMPTY_SLOT_ICON
			slot_button.expand_icon = true
			slot_button.disabled = true
			slot_button.modulate = Color(0.66, 0.7, 0.74)
		inventory_grid.add_child(slot_button)


func _update_combat_hud() -> void:
	var state: String = "queued" if _last_action_sequence < _next_action_sequence - 1 else _last_action_result
	if _player_defeated:
		state = "respawn in %d" % maxi(0, _player_respawn_at_pulse - _action_pulse)
	player_health_bar.max_value = maxi(1, _player_maximum_health)
	player_health_bar.value = _player_health
	player_health_label.text = "You  %d/%d  •  Dodge %d%%" % [
		_player_health,
		_player_maximum_health,
		roundi(_player_dodge_rate * 100.0),
	]
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
		if displayed_target_key.is_empty() and not target_ids.is_empty():
			displayed_target_key = str(target_ids[0])
	var displayed_target: Dictionary = _target_states.get(displayed_target_key, {})
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
	var displayed_weapon: String = _queued_weapon if equipment_pending else _equipped_weapon
	sword_button.set_pressed_no_signal(displayed_weapon == "sword")
	dagger_button.set_pressed_no_signal(displayed_weapon == "dagger")
	bow_button.set_pressed_no_signal(displayed_weapon == "bow")
	sword_button.text = "1  Sword%s" % ("  • queued" if equipment_pending and _queued_weapon == "sword" else "")
	dagger_button.text = "2  Dagger%s" % ("  • queued" if equipment_pending and _queued_weapon == "dagger" else "")
	bow_button.text = "3  Bow%s" % ("  • queued" if equipment_pending and _queued_weapon == "bow" else "")
	sword_button.disabled = _player_defeated
	dagger_button.disabled = _player_defeated
	bow_button.disabled = _player_defeated
	disengage_button.disabled = _player_defeated or (_combat_target.is_empty() and _last_action_sequence >= _next_action_sequence - 1)
