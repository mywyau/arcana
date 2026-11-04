class_name TargetDummy
extends Node3D

const REST_COLOR := Color(0.75, 0.16, 0.12, 1.0)

@onready var visual: Node3D = $Visual
@onready var mesh: MeshInstance3D = $Visual/Mesh
@onready var damage_label: Label3D = $DamageLabel
@onready var target_body: StaticBody3D = $TargetBody
@onready var collision_shape: CollisionShape3D = $TargetBody/CollisionShape3D

var _material: StandardMaterial3D
var _hit_tween: Tween
var _defeated: bool = false


func _ready() -> void:
	var source := mesh.get_active_material(0) as StandardMaterial3D
	if source != null:
		_material = source.duplicate() as StandardMaterial3D
		mesh.material_override = _material


func play_hit(damage: int, attacker_position: Vector3) -> void:
	if _hit_tween != null:
		_hit_tween.kill()

	visual.position = Vector3.ZERO
	visual.scale = Vector3.ONE
	damage_label.text = "-%d" % damage
	damage_label.position = Vector3(0.0, 2.1, 0.0)
	damage_label.modulate = Color(1.0, 0.25, 0.18, 1.0)
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
		_hit_tween.tween_property(_material, "albedo_color", REST_COLOR, 0.18)
	_hit_tween.chain().tween_callback(func() -> void: damage_label.visible = false)


func set_defeated(defeated: bool) -> void:
	if _defeated == defeated:
		return
	_defeated = defeated
	visual.rotation_degrees.z = 90.0 if defeated else 0.0
	target_body.collision_layer = 0 if defeated else 8
	collision_shape.set_deferred("disabled", defeated)
	if not defeated and _material != null:
		_material.albedo_color = REST_COLOR
