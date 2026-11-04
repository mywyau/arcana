class_name FloorVisibility
extends Node3D

@export var layer_id: String = "upper"
@export_range(0.0, 1.0, 0.05) var faded_alpha: float = 0.2
@export_range(0.0, 1.0, 0.05) var connector_alpha: float = 0.55
@export var fade_speed: float = 3.0

var _materials: Array[StandardMaterial3D] = []
var _meshes: Array[MeshInstance3D] = []
var _current_alpha: float = 1.0
var _target_alpha: float = 1.0


func _ready() -> void:
	_collect_materials(self)


func _process(delta: float) -> void:
	if is_equal_approx(_current_alpha, _target_alpha):
		return
	_current_alpha = move_toward(_current_alpha, _target_alpha, fade_speed * delta)
	_apply_alpha()


func set_navigation_location(kind: String, id: String) -> void:
	if kind == "connector":
		_target_alpha = connector_alpha
	elif kind == "layer" and id != layer_id:
		_target_alpha = faded_alpha
	else:
		_target_alpha = 1.0


func _collect_materials(node: Node) -> void:
	for child: Node in node.get_children():
		if child is MeshInstance3D:
			var mesh_instance := child as MeshInstance3D
			var source := mesh_instance.get_active_material(0) as StandardMaterial3D
			if source != null:
				var material := source.duplicate() as StandardMaterial3D
				material.transparency = BaseMaterial3D.TRANSPARENCY_ALPHA
				mesh_instance.material_override = material
				_materials.append(material)
				_meshes.append(mesh_instance)
		_collect_materials(child)


func _apply_alpha() -> void:
	for material: StandardMaterial3D in _materials:
		var color: Color = material.albedo_color
		color.a = _current_alpha
		material.albedo_color = color
	var cast_shadows: bool = _current_alpha > 0.95
	for mesh_instance: MeshInstance3D in _meshes:
		mesh_instance.cast_shadow = (
			GeometryInstance3D.SHADOW_CASTING_SETTING_ON
			if cast_shadows
			else GeometryInstance3D.SHADOW_CASTING_SETTING_OFF
		)
