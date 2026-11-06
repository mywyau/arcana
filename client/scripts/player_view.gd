class_name PlayerView
extends Node3D

@export var interpolation_speed: float = 14.0
@export var movement_speed: float = 6.0
@export var correction_threshold: float = 0.75
@export var snap_threshold: float = 3.0
@export var correction_speed: float = 1.5

@onready var visual_root: Node3D = $VisualRoot
@onready var mesh: MeshInstance3D = $VisualRoot/Mesh
@onready var sword_pivot: Node3D = $VisualRoot/SwordPivot
@onready var blade: MeshInstance3D = $VisualRoot/SwordPivot/Blade
@onready var hilt: MeshInstance3D = $VisualRoot/SwordPivot/Hilt
@onready var bow_pivot: Node3D = $VisualRoot/BowPivot
@onready var mace_pivot: Node3D = $VisualRoot/MacePivot
@onready var battle_axe_pivot: Node3D = $VisualRoot/BattleAxePivot
@onready var staff_pivot: Node3D = $VisualRoot/StaffPivot
@onready var damage_label: Label3D = $DamageLabel

var _is_local: bool = false
var _has_position: bool = false
var _remote_target: Vector3
var _authoritative_position: Vector3
var _predicted_path: Array[Vector3] = []
var _pending_correction: Vector3
var _last_route_sequence: int = -1
var _attack_tween: Tween
var _hit_tween: Tween
var _body_material: StandardMaterial3D
var _rest_color: Color
var _defeated: bool = false
var _following_server_path: bool = false
var _equipped_weapon: String = "sword"


func _process(delta: float) -> void:
	if not _has_position:
		return

	if not _is_local:
		var weight: float = 1.0 - exp(-interpolation_speed * delta)
		global_position = global_position.lerp(_remote_target, weight)
		return

	if not _predicted_path.is_empty():
		global_position = global_position.move_toward(
			_predicted_path[0],
			movement_speed * delta,
		)
		if global_position.is_equal_approx(_predicted_path[0]) and _predicted_path.size() > 1:
			_predicted_path.pop_front()

	if not _pending_correction.is_zero_approx():
		var correction_step: Vector3 = _pending_correction.limit_length(correction_speed * delta)
		global_position += correction_step
		_pending_correction -= correction_step


func predict_move(destination: Vector3) -> void:
	if not _is_local or not _has_position:
		return
	_predicted_path = [destination]


func predict_path(path: Array[Vector3]) -> void:
	if not _is_local or not _has_position:
		return
	_predicted_path = path.duplicate()


func apply_authoritative_state(
	position: Vector3,
	path: Array[Vector3],
	speed: float,
	acknowledged_sequence: int,
	latest_sent_sequence: int,
	server_driven_path: bool,
) -> float:
	_authoritative_position = position
	movement_speed = speed

	if not _has_position:
		global_position = position
		_remote_target = position
		_has_position = true

	if not _is_local:
		_remote_target = position
		return global_position.distance_to(position)

	var correction_distance: float = global_position.distance_to(position)
	if correction_distance >= snap_threshold:
		global_position = position
		_pending_correction = Vector3.ZERO
	elif correction_distance >= correction_threshold:
		_pending_correction = position - global_position
	else:
		_pending_correction = Vector3.ZERO

	if server_driven_path:
		_predicted_path = path.duplicate()
	elif _following_server_path:
		_predicted_path = path.duplicate()
	elif acknowledged_sequence >= latest_sent_sequence and acknowledged_sequence > _last_route_sequence:
		_predicted_path = path.duplicate()
		_last_route_sequence = acknowledged_sequence
	_following_server_path = server_driven_path

	return correction_distance


func cancel_prediction() -> void:
	_predicted_path.clear()
	_pending_correction = _authoritative_position - global_position


func stop_prediction() -> void:
	_predicted_path.clear()
	_pending_correction = Vector3.ZERO


func play_weapon_attack(target_position: Vector3) -> float:
	if _defeated:
		return 0.0
	var flat_target := Vector3(target_position.x, global_position.y, target_position.z)
	if global_position.distance_squared_to(flat_target) > 0.001:
		visual_root.look_at(flat_target, Vector3.UP, true)
	if _equipped_weapon == "bow":
		return _play_bow_attack(target_position)
	if _equipped_weapon == "staff":
		return _play_staff_attack(target_position)
	if _equipped_weapon == "mace":
		_play_melee_attack(mace_pivot, -50.0, 65.0, 0.2, 0.28)
		return 0.0
	if _equipped_weapon == "battle_axe":
		_play_melee_attack(battle_axe_pivot, -55.0, 75.0, 0.24, 0.34)
		return 0.0

	var forward_duration: float = 0.11 if _equipped_weapon == "dagger" else 0.16
	var return_duration: float = 0.14 if _equipped_weapon == "dagger" else 0.22
	_play_melee_attack(sword_pivot, -55.0, 70.0, forward_duration, return_duration)
	return 0.0


func _play_melee_attack(
	pivot: Node3D,
	rest_angle: float,
	attack_angle: float,
	forward_duration: float,
	return_duration: float,
) -> void:
	if _attack_tween != null:
		_attack_tween.kill()
	pivot.rotation_degrees.z = rest_angle
	_attack_tween = create_tween()
	_attack_tween.set_trans(Tween.TRANS_QUAD).set_ease(Tween.EASE_OUT)
	_attack_tween.tween_property(pivot, "rotation_degrees:z", attack_angle, forward_duration)
	_attack_tween.set_ease(Tween.EASE_IN_OUT)
	_attack_tween.tween_property(pivot, "rotation_degrees:z", rest_angle, return_duration)


func _play_bow_attack(target_position: Vector3) -> float:
	if _attack_tween != null:
		_attack_tween.kill()
	bow_pivot.rotation_degrees.z = -10.0
	_attack_tween = create_tween()
	_attack_tween.set_trans(Tween.TRANS_QUAD).set_ease(Tween.EASE_OUT)
	_attack_tween.tween_property(bow_pivot, "rotation_degrees:z", 8.0, 0.12)
	_attack_tween.set_ease(Tween.EASE_IN_OUT)
	_attack_tween.tween_property(bow_pivot, "rotation_degrees:z", -10.0, 0.18)

	var arrow := MeshInstance3D.new()
	var arrow_mesh := BoxMesh.new()
	arrow_mesh.size = Vector3(0.045, 0.045, 0.55)
	var arrow_material := StandardMaterial3D.new()
	arrow_material.albedo_color = Color(0.78, 0.58, 0.3)
	arrow_material.roughness = 0.8
	arrow_mesh.material = arrow_material
	arrow.mesh = arrow_mesh
	var projectile_parent: Node = get_tree().current_scene
	if projectile_parent == null:
		projectile_parent = get_tree().root
	projectile_parent.add_child(arrow)
	var arrow_start := global_position + Vector3.UP * 1.15
	var arrow_target := target_position + Vector3.UP * 1.0
	var horizontal_distance := Vector2(
		arrow_target.x - arrow_start.x,
		arrow_target.z - arrow_start.z,
	).length()
	var arc_height: float = maxf(1.25, horizontal_distance * 0.2)
	var arc_control := (arrow_start + arrow_target) * 0.5 + Vector3.UP * arc_height
	arrow.global_position = arrow_start
	arrow.look_at(arc_control, Vector3.UP, true)
	var flight_time: float = clampf(arrow_start.distance_to(arrow_target) / 18.0, 0.18, 0.55)
	var projectile_tween := arrow.create_tween()
	projectile_tween.tween_method(
		_move_arrow_along_arc.bind(arrow, arrow_start, arc_control, arrow_target),
		0.0,
		1.0,
		flight_time,
	).set_trans(Tween.TRANS_LINEAR)
	projectile_tween.tween_callback(arrow.queue_free)
	return flight_time


func _move_arrow_along_arc(
	progress: float,
	arrow: MeshInstance3D,
	start: Vector3,
	control: Vector3,
	target: Vector3,
) -> void:
	if not is_instance_valid(arrow):
		return
	var first_leg := start.lerp(control, progress)
	var second_leg := control.lerp(target, progress)
	var position := first_leg.lerp(second_leg, progress)
	var look_progress: float = minf(1.0, progress + 0.025)
	var look_first := start.lerp(control, look_progress)
	var look_second := control.lerp(target, look_progress)
	var look_position := look_first.lerp(look_second, look_progress)
	arrow.global_position = position
	if not position.is_equal_approx(look_position):
		arrow.look_at(look_position, Vector3.UP, true)


func _play_staff_attack(target_position: Vector3) -> float:
	if _attack_tween != null:
		_attack_tween.kill()
	staff_pivot.rotation_degrees.z = -15.0
	_attack_tween = create_tween()
	_attack_tween.set_trans(Tween.TRANS_QUAD).set_ease(Tween.EASE_OUT)
	_attack_tween.tween_property(staff_pivot, "rotation_degrees:z", 10.0, 0.14)
	_attack_tween.set_ease(Tween.EASE_IN_OUT)
	_attack_tween.tween_property(staff_pivot, "rotation_degrees:z", -15.0, 0.2)

	var bolt := MeshInstance3D.new()
	var bolt_mesh := SphereMesh.new()
	bolt_mesh.radius = 0.11
	bolt_mesh.height = 0.22
	var bolt_material := StandardMaterial3D.new()
	bolt_material.albedo_color = Color(0.45, 0.86, 1.0)
	bolt_material.emission_enabled = true
	bolt_material.emission = Color(0.25, 0.55, 1.0)
	bolt_material.emission_energy_multiplier = 3.0
	bolt_mesh.material = bolt_material
	bolt.mesh = bolt_mesh
	var projectile_parent: Node = get_tree().current_scene
	if projectile_parent == null:
		projectile_parent = get_tree().root
	projectile_parent.add_child(bolt)
	bolt.global_position = global_position + Vector3.UP * 1.35
	var bolt_target := target_position + Vector3.UP * 1.0
	var flight_time: float = clampf(bolt.global_position.distance_to(bolt_target) / 16.0, 0.2, 0.65)
	var projectile_tween := bolt.create_tween()
	projectile_tween.tween_property(bolt, "global_position", bolt_target, flight_time)
	projectile_tween.tween_callback(bolt.queue_free)
	return flight_time


func set_weapon(weapon_id: String) -> void:
	_equipped_weapon = weapon_id
	var dagger_equipped: bool = weapon_id == "dagger"
	_update_weapon_visibility()
	blade.scale = Vector3(0.8, 0.55, 1.0) if dagger_equipped else Vector3.ONE
	blade.position.y = 0.3 if dagger_equipped else 0.5
	hilt.scale = Vector3(0.75, 0.85, 0.85) if dagger_equipped else Vector3.ONE


func _update_weapon_visibility() -> void:
	sword_pivot.visible = not _defeated and (
		_equipped_weapon == "sword" or _equipped_weapon == "steel_sword" or _equipped_weapon == "dagger"
	)
	bow_pivot.visible = not _defeated and _equipped_weapon == "bow"
	mace_pivot.visible = not _defeated and _equipped_weapon == "mace"
	battle_axe_pivot.visible = not _defeated and _equipped_weapon == "battle_axe"
	staff_pivot.visible = not _defeated and _equipped_weapon == "staff"


func play_hit(damage: int) -> void:
	if _hit_tween != null:
		_hit_tween.kill()
	visual_root.scale = Vector3.ONE
	damage_label.text = "-%d" % damage
	damage_label.position = Vector3(0.0, 2.1, 0.0)
	damage_label.modulate = Color(1.0, 0.25, 0.18, 1.0)
	damage_label.visible = true
	if _body_material != null:
		_body_material.albedo_color = Color(1.0, 0.9, 0.82, 1.0)

	_hit_tween = create_tween()
	_hit_tween.set_parallel(true)
	_hit_tween.tween_property(visual_root, "scale", Vector3(1.08, 0.88, 1.08), 0.08)
	_hit_tween.tween_property(visual_root, "scale", Vector3.ONE, 0.14).set_delay(0.08)
	_hit_tween.tween_property(damage_label, "position:y", 2.8, 0.55)
	_hit_tween.tween_property(damage_label, "modulate:a", 0.0, 0.55).set_delay(0.15)
	if _body_material != null:
		_hit_tween.tween_property(_body_material, "albedo_color", _rest_color, 0.18)
	_hit_tween.chain().tween_callback(func() -> void: damage_label.visible = false)


func play_dodge() -> void:
	if _hit_tween != null:
		_hit_tween.kill()
	damage_label.text = "DODGE"
	damage_label.position = Vector3(0.0, 2.1, 0.0)
	damage_label.modulate = Color(0.4, 0.85, 1.0, 1.0)
	damage_label.visible = true
	_hit_tween = create_tween()
	_hit_tween.set_parallel(true)
	_hit_tween.tween_property(damage_label, "position:y", 2.8, 0.55)
	_hit_tween.tween_property(damage_label, "modulate:a", 0.0, 0.55).set_delay(0.15)
	_hit_tween.chain().tween_callback(func() -> void: damage_label.visible = false)


func set_defeated(defeated: bool) -> void:
	if _defeated == defeated:
		return
	_defeated = defeated
	visual_root.rotation_degrees.z = 90.0 if defeated else 0.0
	_update_weapon_visibility()
	if defeated:
		_predicted_path.clear()
		_pending_correction = Vector3.ZERO


func set_local_player(is_local: bool) -> void:
	_is_local = is_local
	_rest_color = Color("58d68d") if is_local else Color("5dade2")
	_body_material = StandardMaterial3D.new()
	_body_material.albedo_color = _rest_color
	_body_material.roughness = 0.8
	mesh.material_override = _body_material
