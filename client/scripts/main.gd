extends Node3D

const PLAYER_SCENE: PackedScene = preload("res://scenes/player.tscn")
const RAY_LENGTH: float = 1000.0
const GROUND_COLLISION_MASK: int = 1
const UPPER_COLLISION_MASK: int = 2
const CONNECTOR_COLLISION_MASK: int = 4
const TARGET_COLLISION_MASK: int = 8
const ALL_WORLD_COLLISION_MASK: int = 7

@onready var camera_rig: GuidedCamera = $CameraRig
@onready var camera: Camera3D = $CameraRig/Yaw/Pitch/Camera3D
@onready var network: NetworkClient = $NetworkClient
@onready var players: Node3D = $Players
@onready var status_label: Label = $UI/Margin/Status
@onready var diagnostics_label: Label = $UI/DiagnosticsMargin/Diagnostics
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
@onready var disengage_button: Button = $UI/CombatHUD/Margin/VBox/WeaponRow/Disengage
@onready var click_marker: MeshInstance3D = $ClickMarker
@onready var upper_level: FloorVisibility = $UpperLevel
@onready var target_dummy: TargetDummy = $TargetDummy

var _local_player_id: String = ""
var _local_navigation_kind: String = "unknown"
var _local_navigation_id: String = ""
var _next_sequence: int = 1
var _next_action_sequence: int = 1
var _next_equipment_sequence: int = 1
var _last_action_sequence: int = 0
var _last_action_result: String = "none"
var _combat_target: String = ""
var _next_attack_pulse: int = 0
var _auto_approaching: bool = false
var _action_pulse: int = 0
var _dummy_health: int = 100
var _dummy_maximum_health: int = 100
var _dummy_defeated: bool = false
var _dummy_respawn_at_pulse: int = 0
var _dummy_target_id: String = ""
var _player_health: int = 100
var _player_maximum_health: int = 100
var _player_defeated: bool = false
var _player_respawn_at_pulse: int = 0
var _player_dodge_rate: float = 0.1
var _equipped_weapon: String = "sword"
var _last_equipment_sequence: int = 0
var _queued_weapon: String = ""
var _player_views: Dictionary = {}
var _last_combat_event_id: int = 0
var _combat_events_initialized: bool = false
var _pending_click: Vector2
var _has_pending_click: bool = false


func _ready() -> void:
	network.status_changed.connect(_on_status_changed)
	network.welcome_received.connect(_on_welcome_received)
	network.snapshot_received.connect(_on_snapshot_received)
	network.server_error.connect(_on_server_error)
	sword_button.pressed.connect(_request_equip.bind("sword"))
	dagger_button.pressed.connect(_request_equip.bind("dagger"))
	disengage_button.pressed.connect(_request_disengage)
	_update_combat_hud()


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
	var query := PhysicsRayQueryParameters3D.create(ray_origin, ray_end)
	query.collision_mask = _active_collision_mask() | TARGET_COLLISION_MASK
	var result: Dictionary = get_world_3d().direct_space_state.intersect_ray(query)
	if result.is_empty():
		return

	var clicked_body := result.get("collider") as CollisionObject3D
	if clicked_body != null and clicked_body.has_meta(&"combat_target_id"):
		var target_id: String = str(clicked_body.get_meta(&"combat_target_id"))
		if network.send_engage_target(_next_action_sequence, target_id):
			_next_action_sequence += 1
			click_marker.visible = false
			_update_combat_hud()
		return

	var destination: Vector3 = result["position"]
	var sequence: int = _next_sequence
	if network.send_move(sequence, destination):
		_next_sequence += 1
		if not _combat_target.is_empty() or _last_action_sequence < _next_action_sequence - 1:
			_request_disengage()
		var local_view: PlayerView = _player_views.get(_local_player_id)
		if local_view != null and _can_predict_directly(local_view, destination, clicked_body):
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


func _request_equip(item_id: String) -> void:
	if _player_defeated:
		return
	if network.send_equip_item(_next_equipment_sequence, item_id):
		_queued_weapon = item_id
		_next_equipment_sequence += 1
		_update_combat_hud()


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
	var dummy_data: Variant = snapshot.get("dummy", {})
	if typeof(dummy_data) == TYPE_DICTIONARY:
		var dummy: Dictionary = dummy_data
		_dummy_health = int(dummy.get("health", _dummy_health))
		_dummy_maximum_health = int(dummy.get("maximum_health", _dummy_maximum_health))
		_dummy_defeated = bool(dummy.get("defeated", false))
		_dummy_respawn_at_pulse = int(dummy.get("respawn_at_pulse", 0))
		_dummy_target_id = str(dummy.get("target_id", ""))
		target_dummy.set_defeated(_dummy_defeated)
		var dummy_position_data: Variant = dummy.get("position", {})
		if typeof(dummy_position_data) == TYPE_DICTIONARY:
			var dummy_position: Dictionary = dummy_position_data
			target_dummy.position = Vector3(
				float(dummy_position.get("x", 0.0)),
				float(dummy_position.get("y", 0.0)),
				float(dummy_position.get("z", 0.0)),
			)
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
			_equipped_weapon = str(player_data.get("equipped_weapon", "sword"))
			_last_equipment_sequence = int(player_data.get("last_equipment_sequence", 0))
			if _last_equipment_sequence >= _next_equipment_sequence - 1:
				_queued_weapon = ""
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
		if target_id == "dummy":
			if source == null:
				continue
			source.play_weapon_attack(target_dummy.global_position)
			if event_type == "damage":
				target_dummy.play_hit(amount, source.global_position)
			else:
				target_dummy.play_dodge()
		elif target != null:
			if source != null:
				source.play_weapon_attack(target.global_position)
			elif source_id == "dummy":
				target_dummy.play_attack(target.global_position)
			if event_type == "damage":
				target.play_hit(amount)
			else:
				target.play_dodge()


func _on_server_error(code: String, message: String) -> void:
	status_label.text = "%s: %s" % [code, message]
	if code == "command_rejected":
		var local_view: PlayerView = _player_views.get(_local_player_id)
		if local_view != null:
			local_view.cancel_prediction()


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
	target_health_bar.max_value = maxi(1, _dummy_maximum_health)
	target_health_bar.value = _dummy_health
	target_health_label.text = "Target Dummy  %d/%d" % [_dummy_health, _dummy_maximum_health]

	var has_aggro: bool = _dummy_target_id == _local_player_id
	aggro_label.text = "ENEMY AGGRO" if has_aggro else "No aggro"
	aggro_label.modulate = Color(1.0, 0.42, 0.32) if has_aggro else Color(0.65, 0.72, 0.78)
	if _player_defeated:
		combat_state_label.text = "Defeated — %s" % state
	elif _dummy_defeated:
		combat_state_label.text = "Target respawns in %d pulses" % maxi(0, _dummy_respawn_at_pulse - _action_pulse)
	elif _auto_approaching:
		combat_state_label.text = "Closing distance to target"
	elif not _combat_target.is_empty():
		combat_state_label.text = "Auto-attacking %s  •  Last: %s" % [_combat_target, state]
	elif has_aggro:
		combat_state_label.text = "Not attacking — create distance to escape"
	else:
		combat_state_label.text = "Out of combat"

	var weapon_interval: int = 3 if _equipped_weapon == "dagger" else 4
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
	sword_button.text = "1  Sword%s" % ("  • queued" if equipment_pending and _queued_weapon == "sword" else "")
	dagger_button.text = "2  Dagger%s" % ("  • queued" if equipment_pending and _queued_weapon == "dagger" else "")
	sword_button.disabled = _player_defeated
	dagger_button.disabled = _player_defeated
	disengage_button.disabled = _player_defeated or (_combat_target.is_empty() and _last_action_sequence >= _next_action_sequence - 1)
