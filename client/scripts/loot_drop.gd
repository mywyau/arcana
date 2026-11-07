class_name LootDrop
extends Node3D

const DEFAULT_ICON_WORLD_SIZE: float = 0.78
const ICON_PRESENTATION: Dictionary = {
	# World size is the longest visible edge in metres. Rotation is applied
	# within the ground plane, which makes long silhouettes less repetitive.
	"knife": {"world_size": 0.72, "rotation": -18.0},
	"chisel": {"world_size": 0.72, "rotation": 14.0},
	"sewing_needle": {"world_size": 0.62, "rotation": -28.0},
	"thread": {"world_size": 0.58, "rotation": 8.0},
	"tinderbox": {"world_size": 0.64, "rotation": -6.0},
	"hammer": {"world_size": 0.82, "rotation": 22.0},
	"fishing_rod": {"world_size": 0.9, "rotation": -18.0},
	"lobster_cage": {"world_size": 0.72, "rotation": 6.0},
	"prayer_potion": {"world_size": 0.58, "rotation": -4.0},
	"iron_arrow": {"world_size": 0.72, "rotation": -24.0},
	"air_rune": {"world_size": 0.56, "rotation": 0.0},
}

@onready var visual: Node3D = $Visual
@onready var mesh: MeshInstance3D = $Visual/Mesh
@onready var icon_sprite: Sprite3D = $Visual/IconSprite
@onready var body: StaticBody3D = $Body
@onready var item_label: Label3D = $ItemLabel

var _item_id: String = ""
var _quantity: int = 0

func set_drop_id(drop_id: String) -> void:
	body.set_meta(&"drop_id", drop_id)


func set_stack_top(is_top: bool) -> void:
	visual.visible = is_top
	item_label.visible = is_top


func set_item_id(item_id: String, quantity: int = 1) -> void:
	# Snapshots repeat unchanged drops frequently. A drop's item identity is
	# immutable, so only its inexpensive quantity label needs a possible update.
	# Rebuilding the visual here used to append coin/tool meshes every snapshot.
	if item_id == _item_id:
		if quantity != _quantity:
			_quantity = quantity
			_update_item_label()
		return
	_item_id = item_id
	_quantity = quantity
	_update_item_label()

	_build_item_visual(item_id)


func _update_item_label() -> void:
	item_label.text = (
		"%s ×%d" % [_item_id.capitalize(), _quantity]
		if _quantity > 1
		else _item_id.capitalize()
	)


func _build_item_visual(item_id: String) -> void:
	visual.position = Vector3.ZERO
	visual.rotation = Vector3.ZERO
	mesh.position = Vector3.ZERO
	mesh.rotation = Vector3.ZERO
	mesh.scale = Vector3.ONE
	mesh.visible = true
	icon_sprite.position = Vector3(0.0, 0.04, 0.0)
	icon_sprite.rotation_degrees = Vector3(-90.0, 0.0, 0.0)
	icon_sprite.scale = Vector3.ONE
	icon_sprite.offset = Vector2.ZERO
	icon_sprite.visible = false
	if item_id == "coins":
		visual.position.y = 0.06
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
		mesh.position.y = 0.38
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
		mesh.position.y = 0.19
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
		mesh.position.y = 0.22
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
		var item_icon: Texture2D = IconCatalog.ITEMS.get(item_id) as Texture2D
		if item_icon != null:
			mesh.visible = false
			_configure_icon_sprite(item_id, item_icon)
		return
	mesh.position.y = 0.42
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


func _configure_icon_sprite(item_id: String, texture: Texture2D) -> void:
	var centered_texture := IconCatalog.centered_icon(texture)
	icon_sprite.texture = centered_texture
	var visible_size := Vector2i(centered_texture.get_width(), centered_texture.get_height())
	var image := centered_texture.get_image()
	if image != null and not image.is_empty():
		var used_rect := image.get_used_rect()
		if used_rect.size.x > 0 and used_rect.size.y > 0:
			visible_size = used_rect.size
	var longest_edge: int = maxi(visible_size.x, visible_size.y)
	var presentation: Dictionary = ICON_PRESENTATION.get(item_id, {}) as Dictionary
	var world_size: float = float(presentation.get("world_size", DEFAULT_ICON_WORLD_SIZE))
	icon_sprite.pixel_size = world_size / float(maxi(longest_edge, 1))
	icon_sprite.rotation_degrees.z = float(presentation.get("rotation", 0.0))
	icon_sprite.visible = true
