class_name LootDrop
extends Node3D

@onready var visual: Node3D = $Visual
@onready var mesh: MeshInstance3D = $Visual/Mesh
@onready var body: StaticBody3D = $Body
@onready var item_label: Label3D = $ItemLabel

var _phase: float = 0.0
var _floats: bool = true

func _process(delta: float) -> void:
	if not _floats:
		return
	visual.rotation.y += delta * 1.8
	_phase += delta * 2.5
	visual.position.y = 0.35 + sin(_phase) * 0.08


func set_drop_id(drop_id: String) -> void:
	body.set_meta(&"drop_id", drop_id)
	_phase = float(abs(drop_id.hash()) % 100) / 100.0 * TAU


func set_item_id(item_id: String, quantity: int = 1) -> void:
	item_label.text = "%s ×%d" % [item_id.capitalize(), quantity] if quantity > 1 else item_id.capitalize()
	if item_id == "coins":
		_floats = false
		visual.position.y = 0.06
		visual.rotation = Vector3.ZERO
		var coin_mesh := CylinderMesh.new()
		coin_mesh.top_radius = 0.18
		coin_mesh.bottom_radius = 0.18
		coin_mesh.height = 0.055
		coin_mesh.radial_segments = 12
		var coin_material := StandardMaterial3D.new()
		coin_material.albedo_color = Color(0.95, 0.67, 0.12)
		coin_material.metallic = 0.72
		coin_material.roughness = 0.3
		coin_material.emission_enabled = true
		coin_material.emission = Color(0.28, 0.12, 0.01)
		coin_material.emission_energy_multiplier = 0.35
		coin_mesh.material = coin_material
		mesh.mesh = coin_mesh
		mesh.position = Vector3(-0.13, 0.0, 0.05)
		mesh.rotation_degrees = Vector3(0.0, 12.0, 2.0)
		var coin_transforms: Array[Transform3D] = [
			Transform3D(Basis.from_euler(Vector3(0.02, -0.35, -0.03)), Vector3(0.12, 0.01, 0.07)),
			Transform3D(Basis.from_euler(Vector3(-0.03, 0.2, 0.02)), Vector3(-0.02, 0.055, -0.08)),
			Transform3D(Basis.from_euler(Vector3(0.04, 0.5, 0.0)), Vector3(0.17, 0.07, -0.08)),
			Transform3D(Basis.from_euler(Vector3(-0.02, -0.15, 0.04)), Vector3(-0.12, 0.11, 0.04)),
			Transform3D(Basis.from_euler(Vector3(0.03, 0.3, -0.02)), Vector3(0.05, 0.125, -0.01)),
		]
		for coin_transform: Transform3D in coin_transforms:
			var coin := MeshInstance3D.new()
			coin.mesh = coin_mesh
			coin.transform = coin_transform
			visual.add_child(coin)
		return
	if item_id in ["bronze_pickaxe", "bronze_axe"]:
		var handle_mesh := BoxMesh.new()
		handle_mesh.size = Vector3(0.09, 0.9, 0.09)
		var handle_material := StandardMaterial3D.new()
		handle_material.albedo_color = Color(0.34, 0.16, 0.06)
		handle_material.roughness = 0.9
		handle_mesh.material = handle_material
		mesh.mesh = handle_mesh
		mesh.rotation_degrees.z = -42.0
		var head := MeshInstance3D.new()
		var head_mesh := BoxMesh.new()
		head_mesh.size = Vector3(0.52 if item_id == "bronze_pickaxe" else 0.32, 0.13, 0.12)
		var bronze_material := StandardMaterial3D.new()
		bronze_material.albedo_color = Color(0.68, 0.32, 0.12)
		bronze_material.metallic = 0.65
		bronze_material.roughness = 0.35
		head_mesh.material = bronze_material
		head.mesh = head_mesh
		head.position = Vector3(0.0, 0.43, 0.0)
		if item_id == "bronze_axe":
			head.position.x = 0.12
		mesh.add_child(head)
		return
	if item_id in ["normal_logs", "oak_logs"]:
		var log_mesh := CylinderMesh.new()
		log_mesh.top_radius = 0.18
		log_mesh.bottom_radius = 0.18
		log_mesh.height = 0.75
		log_mesh.radial_segments = 8
		var log_material := StandardMaterial3D.new()
		log_material.albedo_color = Color(0.3, 0.15, 0.055) if item_id == "oak_logs" else Color(0.42, 0.24, 0.1)
		log_material.roughness = 0.95
		log_mesh.material = log_material
		mesh.mesh = log_mesh
		mesh.rotation_degrees = Vector3(0.0, 0.0, 90.0)
		return
	if item_id in ["copper_ore", "tin_ore", "iron_ore", "coal", "mithril_ore"]:
		var ore_mesh := SphereMesh.new()
		ore_mesh.radius = 0.27
		ore_mesh.height = 0.46
		ore_mesh.radial_segments = 7
		ore_mesh.rings = 4
		var ore_material := StandardMaterial3D.new()
		ore_material.albedo_color = {
			"tin_ore": Color(0.58, 0.61, 0.58),
			"iron_ore": Color(0.48, 0.52, 0.55),
			"coal": Color(0.055, 0.06, 0.065),
			"mithril_ore": Color(0.12, 0.55, 0.62),
		}.get(item_id, Color(0.32, 0.27, 0.23))
		ore_material.roughness = 0.9
		ore_material.metallic = 0.12
		ore_mesh.material = ore_material
		mesh.mesh = ore_mesh
		return
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
