class_name TargetDummy
extends Node3D

const REST_COLOR := Color(0.75, 0.16, 0.12, 1.0)
const CREATURE_COLOR := Color(0.34, 0.52, 0.2, 1.0)
const BIRD_COLOR := Color(0.76, 0.62, 0.25, 1.0)
const COW_COLOR := Color(0.42, 0.25, 0.14, 1.0)
const BOSS_COLOR := Color(0.28, 0.08, 0.42, 1.0)

@onready var visual: Node3D = $Visual
@onready var attack_pivot: Node3D = $Visual/AttackPivot
@onready var mesh: MeshInstance3D = $Visual/AttackPivot/Mesh
@onready var damage_label: Label3D = $DamageLabel
@onready var target_body: StaticBody3D = $TargetBody
@onready var collision_shape: CollisionShape3D = $TargetBody/CollisionShape3D

var _material: StandardMaterial3D
var _hit_tween: Tween
var _attack_tween: Tween
var _defeated: bool = false
var _mobile: bool = false
var _has_position: bool = false
var _authoritative_position: Vector3
var _rest_color: Color = REST_COLOR
var _target_kind: String = "dummy"
var _rest_attack_scale: Vector3 = Vector3.ONE


func _ready() -> void:
	var source := mesh.get_active_material(0) as StandardMaterial3D
	if source != null:
		_material = source.duplicate() as StandardMaterial3D
		mesh.material_override = _material


func _process(delta: float) -> void:
	if not _mobile or not _has_position:
		return
	var weight: float = 1.0 - exp(-14.0 * delta)
	global_position = global_position.lerp(_authoritative_position, weight)


func set_target_id(target_id: String) -> void:
	target_body.set_meta(&"combat_target_id", target_id)


func set_target_kind(kind: String) -> void:
	_target_kind = kind
	_mobile = kind in ["creature", "bird", "cow", "boss"]
	if kind == "bird":
		_rest_color = BIRD_COLOR
		_rest_attack_scale = Vector3(0.82, 0.48, 1.0)
		collision_shape.position = Vector3(0.0, 0.55, 0.0)
		collision_shape.scale = Vector3(0.8, 0.5, 0.8)
	elif kind == "creature":
		_rest_color = CREATURE_COLOR
		_rest_attack_scale = Vector3(1.15, 0.75, 1.25)
		collision_shape.position = Vector3(0.0, 1.0, 0.0)
		collision_shape.scale = Vector3.ONE
	elif kind == "cow":
		_rest_color = COW_COLOR
		_rest_attack_scale = Vector3(1.35, 0.8, 1.65)
		collision_shape.position = Vector3(0.0, 0.8, 0.0)
		collision_shape.scale = Vector3(1.15, 0.75, 1.35)
	elif kind == "boss":
		_rest_color = BOSS_COLOR
		_rest_attack_scale = Vector3(2.1, 1.55, 2.1)
		collision_shape.position = Vector3(0.0, 1.45, 0.0)
		collision_shape.scale = Vector3(1.7, 1.45, 1.7)
	else:
		_rest_color = REST_COLOR
		_rest_attack_scale = Vector3.ONE
		collision_shape.position = Vector3(0.0, 1.0, 0.0)
		collision_shape.scale = Vector3.ONE
	attack_pivot.scale = _rest_attack_scale
	if _material != null and not _defeated:
		_material.albedo_color = _rest_color


func set_authoritative_position(authoritative_position: Vector3) -> void:
	_authoritative_position = authoritative_position
	if not _has_position or not _mobile:
		global_position = authoritative_position
	_has_position = true


func play_hit(
	damage: int,
	attacker_position: Vector3,
	feedback_color: Color = Color(1.0, 0.25, 0.18, 1.0),
) -> void:
	if _hit_tween != null:
		_hit_tween.kill()

	visual.position = Vector3.ZERO
	visual.scale = Vector3.ONE
	damage_label.text = "-%d" % damage
	damage_label.position = Vector3(0.0, 2.1, 0.0)
	damage_label.modulate = feedback_color
	damage_label.visible = true
	if _material != null:
		_material.albedo_color = Color(1.0, 0.9, 0.82, 1.0)

	var recoil := global_position - attacker_position
	recoil.y = 0.0
	if recoil.is_zero_approx():
		recoil = Vector3.FORWARD
	recoil = recoil.normalized() * 0.22

	_hit_tween = create_tween()
	_hit_tween.set_parallel(true)
	_hit_tween.tween_property(visual, "position", recoil, 0.08)
	_hit_tween.tween_property(visual, "position", Vector3.ZERO, 0.14).set_delay(0.08)
	_hit_tween.tween_property(visual, "scale", Vector3(1.08, 0.9, 1.08), 0.08)
	_hit_tween.tween_property(visual, "scale", Vector3.ONE, 0.14).set_delay(0.08)
	_hit_tween.tween_property(damage_label, "position:y", 2.8, 0.55)
	_hit_tween.tween_property(damage_label, "modulate:a", 0.0, 0.55).set_delay(0.15)
	if _material != null:
		_hit_tween.tween_property(_material, "albedo_color", _rest_color, 0.18)
	_hit_tween.chain().tween_callback(func() -> void: damage_label.visible = false)


func play_dodge(label_text: String = "DODGE") -> void:
	if _hit_tween != null:
		_hit_tween.kill()
	damage_label.text = label_text
	damage_label.position = Vector3(0.0, 2.1, 0.0)
	damage_label.modulate = Color(0.4, 0.85, 1.0, 1.0)
	damage_label.visible = true
	_hit_tween = create_tween()
	_hit_tween.set_parallel(true)
	_hit_tween.tween_property(damage_label, "position:y", 2.8, 0.55)
	_hit_tween.tween_property(damage_label, "modulate:a", 0.0, 0.55).set_delay(0.15)
	_hit_tween.chain().tween_callback(func() -> void: damage_label.visible = false)


func play_block() -> void:
	play_dodge("BLOCK")
	damage_label.modulate = Color(0.72, 0.78, 0.84, 1.0)


func play_attack(target_position: Vector3) -> void:
	if _defeated:
		return
	if _attack_tween != null:
		_attack_tween.kill()
	attack_pivot.position = Vector3.ZERO
	attack_pivot.rotation_degrees = Vector3.ZERO
	attack_pivot.scale = _rest_attack_scale
	var direction: Vector3 = visual.to_local(target_position)
	direction.y = 0.0
	if direction.is_zero_approx():
		direction = Vector3.FORWARD
	direction = direction.normalized()
	var windup: Vector3 = -direction * (0.16 if _target_kind == "bird" else 0.22)
	var strike: Vector3 = direction * (0.28 if _target_kind == "bird" else 0.42)
	_attack_tween = create_tween()
	_attack_tween.set_trans(Tween.TRANS_QUAD).set_ease(Tween.EASE_OUT)
	_attack_tween.tween_property(attack_pivot, "position", windup, 0.12)
	_attack_tween.parallel().tween_property(
		attack_pivot,
		"scale",
		_rest_attack_scale * (
			Vector3(1.08, 0.88, 1.08)
			if _target_kind == "bird"
			else Vector3(0.94, 1.06, 0.94)
		),
		0.12,
	)
	_attack_tween.set_trans(Tween.TRANS_CUBIC).set_ease(Tween.EASE_IN)
	_attack_tween.tween_property(attack_pivot, "position", strike, 0.1)
	_attack_tween.parallel().tween_property(attack_pivot, "rotation_degrees:x", -18.0, 0.1)
	_attack_tween.set_trans(Tween.TRANS_QUAD).set_ease(Tween.EASE_OUT)
	_attack_tween.tween_property(attack_pivot, "position", Vector3.ZERO, 0.18)
	_attack_tween.parallel().tween_property(attack_pivot, "rotation_degrees", Vector3.ZERO, 0.18)
	_attack_tween.parallel().tween_property(attack_pivot, "scale", _rest_attack_scale, 0.18)


func set_defeated(defeated: bool) -> void:
	if _defeated == defeated:
		return
	_defeated = defeated
	visual.rotation_degrees.z = 90.0 if defeated else 0.0
	target_body.collision_layer = 0 if defeated else 8
	collision_shape.set_deferred("disabled", defeated)
	if not defeated and _material != null:
		_material.albedo_color = _rest_color
