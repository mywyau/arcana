class_name PlayerView
extends Node3D

@export var interpolation_speed: float = 14.0

@onready var mesh: MeshInstance3D = $Mesh

var _target_position: Vector3
var _has_position: bool = false


func _process(delta: float) -> void:
	if not _has_position:
		return
	var weight: float = 1.0 - exp(-interpolation_speed * delta)
	global_position = global_position.lerp(_target_position, weight)


func set_authoritative_position(position: Vector3) -> void:
	_target_position = position
	if not _has_position:
		global_position = position
		_has_position = true


func set_local_player(is_local: bool) -> void:
	var material := StandardMaterial3D.new()
	material.albedo_color = Color("58d68d") if is_local else Color("5dade2")
	material.roughness = 0.8
	mesh.material_override = material
