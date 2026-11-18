class_name TyrantArenas
extends Node3D

# These dimensions mirror prototypeTyrantArenaObstacles in the Go navigation
# map. Keep both definitions aligned until world geometry becomes data-driven.
const CENTER_Z := 60.0
const WALL_HEIGHT := 2.6
const WALL_THICKNESS := 1.0
const OUTER_SIZE := 23.0
const INNER_SIZE := 21.0
const ENTRANCE_WIDTH := 6.0
const SOUTH_SEGMENT_LENGTH := (OUTER_SIZE - ENTRANCE_WIDTH) * 0.5

const ARENAS: Array[Dictionary] = [
	{
		"name": "RiftTyrantArena",
		"title": "RIFT TYRANT",
		"center_x": -48.0,
		"wall_color": Color("6f435f"),
		"floor_color": Color("403345"),
	},
	{
		"name": "RangedTyrantArena",
		"title": "RANGED TYRANT",
		"center_x": 0.0,
		"wall_color": Color("496851"),
		"floor_color": Color("34453a"),
	},
	{
		"name": "MageTyrantArena",
		"title": "MAGE TYRANT",
		"center_x": 48.0,
		"wall_color": Color("4d5279"),
		"floor_color": Color("343650"),
	},
]


func _ready() -> void:
	for definition: Dictionary in ARENAS:
		_build_arena(definition)


func _build_arena(definition: Dictionary) -> void:
	var arena := Node3D.new()
	arena.name = str(definition["name"])
	add_child(arena)
	var center_x := float(definition["center_x"])
	var wall_material := _material(definition["wall_color"] as Color, 0.92)
	var floor_material := _material(definition["floor_color"] as Color, 1.0)

	_add_floor(arena, Vector3(center_x, 0.025, CENTER_Z), floor_material)
	_add_wall(
		arena, "NorthWall", Vector3(center_x, WALL_HEIGHT * 0.5, CENTER_Z + INNER_SIZE * 0.5),
		Vector3(OUTER_SIZE, WALL_HEIGHT, WALL_THICKNESS), wall_material,
	)
	_add_wall(
		arena, "WestWall", Vector3(center_x - INNER_SIZE * 0.5, WALL_HEIGHT * 0.5, CENTER_Z),
		Vector3(WALL_THICKNESS, WALL_HEIGHT, OUTER_SIZE), wall_material,
	)
	_add_wall(
		arena, "EastWall", Vector3(center_x + INNER_SIZE * 0.5, WALL_HEIGHT * 0.5, CENTER_Z),
		Vector3(WALL_THICKNESS, WALL_HEIGHT, OUTER_SIZE), wall_material,
	)
	var south_segment_offset := ENTRANCE_WIDTH * 0.5 + SOUTH_SEGMENT_LENGTH * 0.5
	_add_wall(
		arena, "SouthWestWall",
		Vector3(center_x - south_segment_offset, WALL_HEIGHT * 0.5, CENTER_Z - INNER_SIZE * 0.5),
		Vector3(SOUTH_SEGMENT_LENGTH, WALL_HEIGHT, WALL_THICKNESS), wall_material,
	)
	_add_wall(
		arena, "SouthEastWall",
		Vector3(center_x + south_segment_offset, WALL_HEIGHT * 0.5, CENTER_Z - INNER_SIZE * 0.5),
		Vector3(SOUTH_SEGMENT_LENGTH, WALL_HEIGHT, WALL_THICKNESS), wall_material,
	)
	_add_title(arena, str(definition["title"]), Vector3(center_x, 3.15, CENTER_Z - 10.65))


func _add_floor(parent: Node3D, center: Vector3, material: StandardMaterial3D) -> void:
	var floor_mesh := BoxMesh.new()
	floor_mesh.size = Vector3(INNER_SIZE, 0.05, INNER_SIZE)
	var floor := MeshInstance3D.new()
	floor.name = "ArenaFloor"
	floor.position = center
	floor.mesh = floor_mesh
	floor.material_override = material
	parent.add_child(floor)


func _add_wall(
	parent: Node3D,
	wall_name: String,
	center: Vector3,
	size: Vector3,
	material: StandardMaterial3D,
) -> void:
	var wall_mesh := BoxMesh.new()
	wall_mesh.size = size
	var wall := MeshInstance3D.new()
	wall.name = wall_name
	wall.position = center
	wall.mesh = wall_mesh
	wall.material_override = material
	parent.add_child(wall)

	var body := StaticBody3D.new()
	body.name = wall_name + "Body"
	body.position = center
	body.set_meta(&"navigation_layer_id", "ground")
	var shape := CollisionShape3D.new()
	var box := BoxShape3D.new()
	box.size = size
	shape.shape = box
	body.add_child(shape)
	parent.add_child(body)


func _add_title(parent: Node3D, title: String, position: Vector3) -> void:
	var label := Label3D.new()
	label.name = "ArenaTitle"
	label.position = position
	label.text = title
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.no_depth_test = true
	label.font_size = 24
	label.outline_size = 8
	label.modulate = Color("e8e1f4")
	parent.add_child(label)


func _material(color: Color, roughness: float) -> StandardMaterial3D:
	var material := StandardMaterial3D.new()
	material.albedo_color = color
	material.roughness = roughness
	return material
