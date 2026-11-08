class_name GuidedCamera
extends Node3D

@export_range(25.0, 75.0, 1.0) var minimum_pitch_degrees: float = 35.0
@export_range(25.0, 75.0, 1.0) var maximum_pitch_degrees: float = 65.0
@export var initial_pitch_degrees: float = 50.0
@export var initial_yaw_degrees: float = 35.0
@export var minimum_distance: float = 5.0
@export var maximum_distance: float = 30.0
@export var initial_distance: float = 13.0
@export var focus_height: float = 0.8
@export var follow_speed: float = 10.0
@export var orbit_speed: float = 0.25
@export var keyboard_yaw_speed_degrees: float = 150.0
@export var keyboard_pitch_speed_degrees: float = 90.0
@export var zoom_step: float = 1.5
@export var collision_margin: float = 0.3
@export var minimum_collision_distance: float = 1.0

@onready var yaw: Node3D = $Yaw
@onready var pitch: Node3D = $Yaw/Pitch
@onready var camera: Camera3D = $Yaw/Pitch/Camera3D

var _target: Node3D
var _orbiting: bool = false
var _has_follow_position: bool = false
var _desired_yaw: float
var _desired_pitch: float
var _desired_distance: float
var _current_distance: float
var _collision_distance: float
var _world_collision_mask: int = 7
var _zoom_blocked: bool = false


func _ready() -> void:
	_desired_yaw = deg_to_rad(initial_yaw_degrees)
	_desired_pitch = clampf(initial_pitch_degrees, minimum_pitch_degrees, maximum_pitch_degrees)
	_desired_distance = clampf(initial_distance, minimum_distance, maximum_distance)
	_current_distance = _desired_distance
	_collision_distance = _current_distance
	yaw.rotation.y = _desired_yaw
	pitch.rotation.x = -deg_to_rad(_desired_pitch)
	camera.position = Vector3(0.0, 0.0, _collision_distance)


func _process(delta: float) -> void:
	_apply_keyboard_rotation(delta)
	if is_instance_valid(_target):
		var desired_focus: Vector3 = _target.global_position + Vector3.UP * focus_height
		if not _has_follow_position:
			global_position = desired_focus
			_has_follow_position = true
		else:
			var follow_weight: float = 1.0 - exp(-follow_speed * delta)
			global_position = global_position.lerp(desired_focus, follow_weight)

	var rotation_weight: float = 1.0 - exp(-12.0 * delta)
	yaw.rotation.y = lerp_angle(yaw.rotation.y, _desired_yaw, rotation_weight)
	pitch.rotation.x = lerp_angle(
		pitch.rotation.x,
		-deg_to_rad(_desired_pitch),
		rotation_weight,
	)
	var zoom_weight: float = 1.0 - exp(-10.0 * delta)
	_current_distance = lerpf(_current_distance, _desired_distance, zoom_weight)


func _apply_keyboard_rotation(delta: float) -> void:
	var focused_control: Control = get_viewport().gui_get_focus_owner()
	if focused_control is LineEdit or focused_control is SpinBox:
		return
	var yaw_direction: float = (
		float(Input.is_key_pressed(KEY_LEFT))
		- float(Input.is_key_pressed(KEY_RIGHT))
	)
	var pitch_direction: float = (
		float(Input.is_key_pressed(KEY_DOWN))
		- float(Input.is_key_pressed(KEY_UP))
	)
	_desired_yaw += deg_to_rad(yaw_direction * keyboard_yaw_speed_degrees * delta)
	_desired_pitch = clampf(
		_desired_pitch + pitch_direction * keyboard_pitch_speed_degrees * delta,
		minimum_pitch_degrees,
		maximum_pitch_degrees,
	)


func _physics_process(delta: float) -> void:
	var desired_camera_position: Vector3 = pitch.to_global(Vector3(0.0, 0.0, _current_distance))
	var query := PhysicsRayQueryParameters3D.create(global_position, desired_camera_position)
	query.collision_mask = _world_collision_mask
	var result: Dictionary = get_world_3d().direct_space_state.intersect_ray(query)
	var available_distance: float = _current_distance
	if not result.is_empty():
		var hit_position: Vector3 = result["position"]
		available_distance = maxf(
			minimum_collision_distance,
			global_position.distance_to(hit_position) - collision_margin,
		)

	if available_distance < _collision_distance:
		_collision_distance = available_distance
	else:
		var recovery_weight: float = 1.0 - exp(-8.0 * delta)
		_collision_distance = lerpf(_collision_distance, available_distance, recovery_weight)
	camera.position = Vector3(0.0, 0.0, _collision_distance)


func _unhandled_input(event: InputEvent) -> void:
	if event is InputEventMouseButton:
		var mouse_button := event as InputEventMouseButton
		match mouse_button.button_index:
			MOUSE_BUTTON_MIDDLE:
				_orbiting = mouse_button.pressed
				get_viewport().set_input_as_handled()
			MOUSE_BUTTON_WHEEL_UP:
				if mouse_button.pressed and not _zoom_blocked:
					_desired_distance = maxf(minimum_distance, _desired_distance - zoom_step)
					get_viewport().set_input_as_handled()
			MOUSE_BUTTON_WHEEL_DOWN:
				if mouse_button.pressed and not _zoom_blocked:
					_desired_distance = minf(maximum_distance, _desired_distance + zoom_step)
					get_viewport().set_input_as_handled()
	elif event is InputEventMouseMotion and _orbiting:
		var motion := event as InputEventMouseMotion
		_desired_yaw -= deg_to_rad(motion.relative.x * orbit_speed)
		_desired_pitch = clampf(
			_desired_pitch + motion.relative.y * orbit_speed,
			minimum_pitch_degrees,
			maximum_pitch_degrees,
		)
		get_viewport().set_input_as_handled()


func follow(target: Node3D) -> void:
	_target = target
	if not is_instance_valid(target):
		_has_follow_position = false


func set_world_collision_mask(collision_mask: int) -> void:
	_world_collision_mask = collision_mask


func set_zoom_blocked(blocked: bool) -> void:
	_zoom_blocked = blocked
