class_name LootDrop
extends Node3D

@onready var visual: Node3D = $Visual
@onready var mesh: MeshInstance3D = $Visual/Mesh
@onready var body: StaticBody3D = $Body
@onready var item_label: Label3D = $ItemLabel

var _phase: float = 0.0

func _process(delta: float) -> void:
	visual.rotation.y += delta * 1.8
	_phase += delta * 2.5
	visual.position.y = 0.35 + sin(_phase) * 0.08


func set_drop_id(drop_id: String) -> void:
	body.set_meta(&"drop_id", drop_id)
	_phase = float(abs(drop_id.hash()) % 100) / 100.0 * TAU


func set_item_id(item_id: String, quantity: int = 1) -> void:
	item_label.text = "%s ×%d" % [item_id.capitalize(), quantity] if quantity > 1 else item_id.capitalize()
	if item_id != "steel_sword":
		return
	var blade_mesh := BoxMesh.new()
	blade_mesh.size = Vector3(0.14, 1.0, 0.08)
	var blade_material := StandardMaterial3D.new()
	blade_material.albedo_color = Color(0.72, 0.82, 0.94)
	blade_material.metallic = 0.8
	blade_material.roughness = 0.2
	blade_material.emission_enabled = true
	blade_material.emission = Color(0.12, 0.28, 0.48)
	blade_material.emission_energy_multiplier = 0.7
	blade_mesh.material = blade_material
	mesh.mesh = blade_mesh
	mesh.rotation_degrees.z = -35.0
