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
	{"name": "Blades", "level": 1},
	{"name": "Vitality", "level": 1},
	{"name": "Agility", "level": 1},
	{"name": "Arcana", "level": 1},
	{"name": "Devotion", "level": 1},
	{"name": "Crafting", "level": 1},
]
const SWORD_ICON: Texture2D = preload("res://assets/icons/sword.svg")
const DAGGER_ICON: Texture2D = preload("res://assets/icons/dagger.svg")
const BOW_ICON: Texture2D = preload("res://assets/icons/bow.svg")
const EMPTY_SLOT_ICON: Texture2D = preload("res://assets/icons/empty_slot.svg")
const TRAINING_SHARD_ICON: Texture2D = preload("res://assets/icons/training_shard.svg")
const EQUIPMENT_ICONS: Dictionary = {
	"sword": SWORD_ICON,
	"dagger": DAGGER_ICON,
	"bow": BOW_ICON,
}
const ITEM_ICONS: Dictionary = {
	"sword": SWORD_ICON,
	"dagger": DAGGER_ICON,
	"bow": BOW_ICON,
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
@onready var combat_tab_button: Button = $UI/CombatHUD/Margin/VBox/TabRow/Combat
@onready var inventory_tab_button: Button = $UI/CombatHUD/Margin/VBox/TabRow/Inventory
@onready var combat_skills_tab_button: Button = $UI/CombatHUD/Margin/VBox/TabRow/Skills
@onready var combat_equipment_tab_button: Button = $UI/CombatHUD/Margin/VBox/TabRow/Equipment
@onready var inventory_combat_tab_button: Button = $UI/InventoryHUD/Margin/VBox/TabRow/Combat
@onready var inventory_active_tab_button: Button = $UI/InventoryHUD/Margin/VBox/TabRow/Inventory
@onready var inventory_skills_tab_button: Button = $UI/InventoryHUD/Margin/VBox/TabRow/Skills
@onready var inventory_equipment_tab_button: Button = $UI/InventoryHUD/Margin/VBox/TabRow/Equipment
@onready var inventory_summary: Label = $UI/InventoryHUD/Margin/VBox/Summary
@onready var inventory_grid: GridContainer = $UI/InventoryHUD/Margin/VBox/Grid
@onready var skills_combat_tab_button: Button = $UI/SkillsHUD/Margin/VBox/TabRow/Combat
@onready var skills_inventory_tab_button: Button = $UI/SkillsHUD/Margin/VBox/TabRow/Inventory
@onready var skills_active_tab_button: Button = $UI/SkillsHUD/Margin/VBox/TabRow/Skills
@onready var skills_equipment_tab_button: Button = $UI/SkillsHUD/Margin/VBox/TabRow/Equipment
@onready var skills_grid: GridContainer = $UI/SkillsHUD/Margin/VBox/Grid
@onready var equipment_combat_tab_button: Button = $UI/EquipmentHUD/Margin/VBox/TabRow/Combat
@onready var equipment_inventory_tab_button: Button = $UI/EquipmentHUD/Margin/VBox/TabRow/Inventory
@onready var equipment_skills_tab_button: Button = $UI/EquipmentHUD/Margin/VBox/TabRow/Skills
@onready var equipment_active_tab_button: Button = $UI/EquipmentHUD/Margin/VBox/TabRow/Equipment
@onready var equipment_grid: GridContainer = $UI/EquipmentHUD/Margin/VBox/Grid
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
var _inventory_capacity: int = 8
var _inventory_items: Array = []
var _inventory_signature: String = ""
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
	inventory_combat_tab_button.pressed.connect(_show_combat_tab)
	inventory_active_tab_button.pressed.connect(_show_inventory_tab)
	inventory_skills_tab_button.pressed.connect(_show_skills_tab)
	inventory_equipment_tab_button.pressed.connect(_show_equipment_tab)
	skills_combat_tab_button.pressed.connect(_show_combat_tab)
	skills_inventory_tab_button.pressed.connect(_show_inventory_tab)
	skills_active_tab_button.pressed.connect(_show_skills_tab)
	skills_equipment_tab_button.pressed.connect(_show_equipment_tab)
	equipment_combat_tab_button.pressed.connect(_show_combat_tab)
	equipment_inventory_tab_button.pressed.connect(_show_inventory_tab)
	equipment_skills_tab_button.pressed.connect(_show_skills_tab)
	equipment_active_tab_button.pressed.connect(_show_equipment_tab)
	_update_combat_hud()
	_update_inventory_hud()
	_populate_placeholder_skills()
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
			var inventory_data: Variant = player_data.get("inventory", {})
			if typeof(inventory_data) == TYPE_DICTIONARY:
				_apply_inventory_snapshot(inventory_data as Dictionary)
			var equipped_weapon: String = str(player_data.get("equipped_weapon", "sword"))
			if equipped_weapon != _equipped_weapon:
				_equipped_weapon = equipped_weapon
				_update_equipment_hud()
				inventory_selection_changed = true
			_last_equipment_sequence = int(player_data.get("last_equipment_sequence", 0))
			if _last_equipment_sequence >= _next_equipment_sequence - 1 and not _queued_weapon.is_empty():
				_queued_weapon = ""
				inventory_selection_changed = true
			if inventory_selection_changed:
				_update_inventory_hud()
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
	if code == "command_rejected":
		var local_view: PlayerView = _player_views.get(_local_player_id)
		if local_view != null:
			local_view.cancel_prediction()


func _show_combat_tab() -> void:
	combat_hud.visible = true
	inventory_hud.visible = false
	skills_hud.visible = false
	equipment_hud.visible = false
	combat_tab_button.set_pressed_no_signal(true)
	inventory_tab_button.set_pressed_no_signal(false)
	combat_skills_tab_button.set_pressed_no_signal(false)
	combat_equipment_tab_button.set_pressed_no_signal(false)


func _show_inventory_tab() -> void:
	combat_hud.visible = false
	inventory_hud.visible = true
	skills_hud.visible = false
	equipment_hud.visible = false
	inventory_combat_tab_button.set_pressed_no_signal(false)
	inventory_active_tab_button.set_pressed_no_signal(true)
	inventory_skills_tab_button.set_pressed_no_signal(false)
	inventory_equipment_tab_button.set_pressed_no_signal(false)


func _show_skills_tab() -> void:
	combat_hud.visible = false
	inventory_hud.visible = false
	skills_hud.visible = true
	equipment_hud.visible = false
	skills_combat_tab_button.set_pressed_no_signal(false)
	skills_inventory_tab_button.set_pressed_no_signal(false)
	skills_active_tab_button.set_pressed_no_signal(true)
	skills_equipment_tab_button.set_pressed_no_signal(false)


func _show_equipment_tab() -> void:
	combat_hud.visible = false
	inventory_hud.visible = false
	skills_hud.visible = false
	equipment_hud.visible = true
	equipment_combat_tab_button.set_pressed_no_signal(false)
	equipment_inventory_tab_button.set_pressed_no_signal(false)
	equipment_skills_tab_button.set_pressed_no_signal(false)
	equipment_active_tab_button.set_pressed_no_signal(true)


func _populate_placeholder_skills() -> void:
	for skill_data: Variant in PLACEHOLDER_SKILLS:
		var skill: Dictionary = skill_data
		var skill_tile := Button.new()
		skill_tile.custom_minimum_size = Vector2(145, 72)
		skill_tile.focus_mode = Control.FOCUS_NONE
		skill_tile.mouse_filter = Control.MOUSE_FILTER_IGNORE
		skill_tile.text = "%s\nLevel %d" % [skill.get("name", "Skill"), skill.get("level", 1)]
		skill_tile.tooltip_text = "Placeholder skill — progression is not active yet"
		skills_grid.add_child(skill_tile)


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

	var weapon_interval: int = 4
	if _equipped_weapon == "dagger":
		weapon_interval = 3
	elif _equipped_weapon == "bow":
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
