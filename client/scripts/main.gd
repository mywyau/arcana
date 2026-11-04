extends Node3D

const PLAYER_SCENE: PackedScene = preload("res://scenes/player.tscn")
const GROUND_PLANE: Plane = Plane.PLANE_XZ

@onready var camera: Camera3D = $Camera3D
@onready var network: NetworkClient = $NetworkClient
@onready var players: Node3D = $Players
@onready var status_label: Label = $UI/Margin/Status
@onready var diagnostics_label: Label = $UI/DiagnosticsMargin/Diagnostics
@onready var click_marker: MeshInstance3D = $ClickMarker

var _local_player_id: String = ""
var _next_sequence: int = 1
var _player_views: Dictionary = {}


func _ready() -> void:
	camera.look_at(Vector3.ZERO)
	network.status_changed.connect(_on_status_changed)
	network.welcome_received.connect(_on_welcome_received)
	network.snapshot_received.connect(_on_snapshot_received)
	network.server_error.connect(_on_server_error)


func _unhandled_input(event: InputEvent) -> void:
	if not event is InputEventMouseButton:
		return
	var mouse_event := event as InputEventMouseButton
	if not mouse_event.pressed or mouse_event.button_index != MOUSE_BUTTON_LEFT:
		return

	var ray_origin: Vector3 = camera.project_ray_origin(mouse_event.position)
	var ray_direction: Vector3 = camera.project_ray_normal(mouse_event.position)
	var intersection: Variant = GROUND_PLANE.intersects_ray(ray_origin, ray_direction)
	if intersection == null:
		return

	var destination := intersection as Vector3
	var sequence: int = _next_sequence
	if network.send_move(sequence, destination):
		_next_sequence += 1
		var local_view: PlayerView = _player_views.get(_local_player_id)
		if local_view != null:
			local_view.predict_move(destination)
		click_marker.position = Vector3(destination.x, 0.03, destination.z)
		click_marker.visible = true


func _on_status_changed(status: String) -> void:
	status_label.text = status


func _on_welcome_received(player_id: String) -> void:
	_local_player_id = player_id
	status_label.text = "Connected as %s — click the ground to move" % player_id
	for id: String in _player_views:
		var view: PlayerView = _player_views[id]
		view.set_local_player(id == _local_player_id)


func _on_snapshot_received(snapshot: Dictionary) -> void:
	var snapshot_tick: int = int(snapshot.get("tick", 0))
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
			0.0,
			float(coordinates.get("z", 0.0)),
		)
		var destination := Vector3.ZERO
		var has_destination: bool = false
		var destination_data: Variant = player_data.get("destination")
		if typeof(destination_data) == TYPE_DICTIONARY:
			var destination_coordinates: Dictionary = destination_data
			destination = Vector3(
				float(destination_coordinates.get("x", 0.0)),
				0.0,
				float(destination_coordinates.get("z", 0.0)),
			)
			has_destination = true

		var view: PlayerView = _player_views.get(id)
		if view == null:
			view = PLAYER_SCENE.instantiate() as PlayerView
			players.add_child(view)
			_player_views[id] = view
			view.set_local_player(id == _local_player_id)

		var acknowledged_sequence: int = int(player_data.get("last_command_sequence", 0))
		var correction: float = view.apply_authoritative_state(
			position,
			destination,
			has_destination,
			float(player_data.get("speed", 4.0)),
			acknowledged_sequence,
			_next_sequence - 1,
		)
		if id == _local_player_id:
			diagnostics_label.text = "tick %d  |  sent %d  |  acknowledged %d  |  correction %.3f" % [
				snapshot_tick,
				_next_sequence - 1,
				acknowledged_sequence,
				correction,
			]
		active_ids[id] = true

	for id: String in _player_views.keys():
		if not active_ids.has(id):
			var departed: PlayerView = _player_views[id]
			departed.queue_free()
			_player_views.erase(id)


func _on_server_error(code: String, message: String) -> void:
	status_label.text = "%s: %s" % [code, message]
	if code == "command_rejected":
		var local_view: PlayerView = _player_views.get(_local_player_id)
		if local_view != null:
			local_view.cancel_prediction()
