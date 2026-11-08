class_name WorldFire
extends Node3D

var _flame: MeshInstance3D
var _light: OmniLight3D
var _phase: float = 0.0


func _ready() -> void:
	for index: int in 2:
		var log := MeshInstance3D.new()
		var mesh := CylinderMesh.new()
		mesh.top_radius = 0.11
		mesh.bottom_radius = 0.14
		mesh.height = 1.0
		var material := StandardMaterial3D.new()
		material.albedo_color = Color(0.25, 0.11, 0.035)
		mesh.material = material
		log.mesh = mesh
		log.rotation_degrees = Vector3(90.0, 45.0 if index == 0 else -45.0, 0.0)
		log.position.y = 0.14
		add_child(log)
	_flame = MeshInstance3D.new()
	var flame_mesh := SphereMesh.new()
	flame_mesh.radius = 0.3
	flame_mesh.height = 0.85
	var flame_material := StandardMaterial3D.new()
	flame_material.shading_mode = BaseMaterial3D.SHADING_MODE_UNSHADED
	flame_material.albedo_color = Color(1.0, 0.36, 0.04, 0.92)
	flame_material.emission_enabled = true
	flame_material.emission = Color(1.0, 0.18, 0.01)
	flame_material.emission_energy_multiplier = 2.0
	flame_mesh.material = flame_material
	_flame.mesh = flame_mesh
	_flame.position.y = 0.55
	add_child(_flame)
	_light = OmniLight3D.new()
	_light.light_color = Color(1.0, 0.48, 0.18)
	_light.light_energy = 0.8
	_light.omni_range = 4.0
	_light.shadow_enabled = false
	_light.position.y = 0.8
	add_child(_light)


func _process(delta: float) -> void:
	_phase += delta * 7.0
	var flicker := 1.0 + sin(_phase) * 0.08 + sin(_phase * 2.31) * 0.04
	_flame.scale = Vector3(0.85, flicker, 0.85)
	_light.light_energy = 0.75 + sin(_phase * 1.73) * 0.08
