extends Node3D

const PLAYER_SCENE: PackedScene = preload("res://scenes/player.tscn")
const GROUND_PLANE: Plane = Plane.PLANE_XZ

@onready var camera: Camera3D = $Camera3D
@onready var network: NetworkClient = $NetworkClient
@onready var players: Node3D = $Players
@onready var status_label: Label = $UI/Margin/Status
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
	if network.send_move(_next_sequence, destination):
		_next_sequence += 1
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
		var view: PlayerView = _player_views.get(id)
		if view == null:
			view = PLAYER_SCENE.instantiate() as PlayerView
			players.add_child(view)
			_player_views[id] = view
			view.set_local_player(id == _local_player_id)
		view.set_authoritative_position(position)
		active_ids[id] = true

	for id: String in _player_views.keys():
		if not active_ids.has(id):
			var departed: PlayerView = _player_views[id]
			departed.queue_free()
			_player_views.erase(id)


func _on_server_error(code: String, message: String) -> void:
	status_label.text = "%s: %s" % [code, message]
