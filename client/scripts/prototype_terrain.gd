class_name PrototypeTerrain
extends MeshInstance3D

const MINIMUM_COORDINATE: float = -50.0
const MAXIMUM_COORDINATE: float = 50.0
const CELL_SIZE: float = 1.0
const HILL_CENTER: Vector2 = Vector2(-12.0, 10.0)
const HILL_RADIUS: float = 4.0
const HILL_HEIGHT: float = 2.0

@export var collision_shape_path: NodePath


func _ready() -> void:
	var terrain_mesh := _build_mesh()
	mesh = terrain_mesh
	var collision_shape := get_node_or_null(collision_shape_path) as CollisionShape3D
	if collision_shape != null:
		collision_shape.shape = terrain_mesh.create_trimesh_shape()


func surface_path(start: Vector3, destination: Vector3) -> Array[Vector3]:
	var horizontal_distance := Vector2(
		destination.x - start.x,
		destination.z - start.z,
	).length()
	var steps: int = maxi(1, ceili(horizontal_distance / CELL_SIZE))
	var path: Array[Vector3] = []
	for index: int in range(1, steps + 1):
		var weight: float = float(index) / float(steps)
		var point := Vector3(
			lerpf(start.x, destination.x, weight),
			0.0,
			lerpf(start.z, destination.z, weight),
		)
		point.y = height_at(point.x, point.z)
		path.append(point)
	path[path.size() - 1] = destination
	return path


static func height_at(x: float, z: float) -> float:
	var distance: float = Vector2(x, z).distance_to(HILL_CENTER)
	if distance >= HILL_RADIUS:
		return 0.0
	return HILL_HEIGHT * 0.5 * (1.0 + cos(PI * distance / HILL_RADIUS))


func _build_mesh() -> ArrayMesh:
	var surface := SurfaceTool.new()
	surface.begin(Mesh.PRIMITIVE_TRIANGLES)
	if material_override != null:
		surface.set_material(material_override)

	var cells: int = roundi((MAXIMUM_COORDINATE - MINIMUM_COORDINATE) / CELL_SIZE)
	for z_index: int in range(cells):
		var z0: float = MINIMUM_COORDINATE + float(z_index) * CELL_SIZE
		var z1: float = z0 + CELL_SIZE
		for x_index: int in range(cells):
			var x0: float = MINIMUM_COORDINATE + float(x_index) * CELL_SIZE
			var x1: float = x0 + CELL_SIZE
			var bottom_left := Vector3(x0, height_at(x0, z0), z0)
			var top_left := Vector3(x0, height_at(x0, z1), z1)
			var bottom_right := Vector3(x1, height_at(x1, z0), z0)
			var top_right := Vector3(x1, height_at(x1, z1), z1)
			_add_triangle(surface, bottom_left, bottom_right, top_left)
			_add_triangle(surface, bottom_right, top_right, top_left)

	surface.generate_normals()
	return surface.commit()


func _add_triangle(surface: SurfaceTool, first: Vector3, second: Vector3, third: Vector3) -> void:
	surface.add_vertex(first)
	surface.add_vertex(second)
	surface.add_vertex(third)
