class_name MinimapHUD
extends Control

const BACKGROUND_COLOR := Color(0.025, 0.055, 0.065, 0.94)
const INNER_COLOR := Color(0.08, 0.16, 0.13, 0.96)
const BORDER_COLOR := Color(0.62, 0.48, 0.2, 1.0)
const GRID_COLOR := Color(0.75, 0.9, 0.8, 0.1)
const PLAYER_COLOR := Color(0.7, 0.94, 1.0, 1.0)

@export_range(10.0, 100.0, 1.0) var world_radius: float = 36.0

var _world_center: Vector3 = Vector3.ZERO
var _facing: Vector2 = Vector2.UP
var _markers: Array[Dictionary] = []


func _ready() -> void:
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	queue_redraw()


func set_world_state(
	world_center: Vector3,
	facing: Vector2,
	markers: Array[Dictionary],
) -> void:
	_world_center = world_center
	_facing = facing.normalized() if not facing.is_zero_approx() else Vector2.UP
	_markers = markers
	queue_redraw()


func map_offset_for(world_position: Vector3) -> Vector2:
	var delta: Vector3 = world_position - _world_center
	return Vector2(delta.x, delta.z) * (_map_radius() / maxf(1.0, world_radius))


func _draw() -> void:
	var center: Vector2 = size * 0.5
	var radius: float = _map_radius()
	draw_circle(center, radius + 7.0, BACKGROUND_COLOR)
	draw_circle(center, radius, INNER_COLOR)
	_draw_grid(center, radius)

	for marker: Dictionary in _markers:
		var position_value: Variant = marker.get("position")
		if not position_value is Vector3:
			continue
		var position := position_value as Vector3
		var offset: Vector2 = map_offset_for(position)
		if offset.length_squared() > radius * radius:
			continue
		_draw_marker(center + offset, str(marker.get("kind", "resource")), position.y - _world_center.y)

	draw_arc(center, radius + 3.5, 0.0, TAU, 96, BORDER_COLOR, 3.0, true)
	_draw_north_marker(center, radius)
	_draw_player_marker(center)


func _map_radius() -> float:
	return maxf(1.0, minf(size.x, size.y) * 0.5 - 12.0)


func _draw_grid(center: Vector2, radius: float) -> void:
	var spacing: float = radius / 3.0
	for step: int in range(-2, 3):
		var offset: float = float(step) * spacing
		var half_chord: float = sqrt(maxf(0.0, radius * radius - offset * offset))
		draw_line(
			center + Vector2(offset, -half_chord),
			center + Vector2(offset, half_chord),
			GRID_COLOR,
			1.0,
			true,
		)
		draw_line(
			center + Vector2(-half_chord, offset),
			center + Vector2(half_chord, offset),
			GRID_COLOR,
			1.0,
			true,
		)


func _draw_marker(position: Vector2, kind: String, elevation: float) -> void:
	var color: Color = _marker_color(kind)
	match kind:
		"target":
			draw_circle(position, 4.5, color)
			draw_arc(position, 6.0, 0.0, TAU, 20, Color(color, 0.45), 1.5, true)
		"drop":
			draw_colored_polygon(PackedVector2Array([
				position + Vector2(0.0, -4.0), position + Vector2(4.0, 0.0),
				position + Vector2(0.0, 4.0), position + Vector2(-4.0, 0.0),
			]), color)
		"bank", "merchant":
			draw_rect(Rect2(position - Vector2(3.5, 3.5), Vector2(7.0, 7.0)), color, true)
		_:
			draw_circle(position, 3.5, color)

	if absf(elevation) > 1.5:
		var direction: float = -1.0 if elevation > 0.0 else 1.0
		var indicator_center := position + Vector2(7.0, direction * 2.0)
		draw_colored_polygon(PackedVector2Array([
			indicator_center + Vector2(0.0, direction * -3.0),
			indicator_center + Vector2(-2.5, direction * 2.0),
			indicator_center + Vector2(2.5, direction * 2.0),
		]), Color(color, 0.8))


func _marker_color(kind: String) -> Color:
	match kind:
		"player":
			return Color(0.35, 0.9, 1.0, 1.0)
		"target":
			return Color(0.95, 0.25, 0.23, 1.0)
		"drop":
			return Color(1.0, 0.78, 0.18, 1.0)
		"fire":
			return Color(1.0, 0.38, 0.08, 1.0)
		"bank":
			return Color(0.45, 0.75, 1.0, 1.0)
		"merchant":
			return Color(1.0, 0.64, 0.15, 1.0)
		_:
			return Color(0.45, 0.9, 0.42, 1.0)


func _draw_north_marker(center: Vector2, radius: float) -> void:
	var top := center + Vector2(0.0, -radius - 1.0)
	draw_colored_polygon(PackedVector2Array([
		top + Vector2(0.0, -6.0),
		top + Vector2(-5.0, 3.0),
		top + Vector2(5.0, 3.0),
	]), Color(0.98, 0.82, 0.3, 1.0))


func _draw_player_marker(center: Vector2) -> void:
	var angle: float = _facing.angle() + PI * 0.5
	var points := PackedVector2Array([
		Vector2(0.0, -10.0).rotated(angle) + center,
		Vector2(6.5, 7.0).rotated(angle) + center,
		Vector2(0.0, 4.0).rotated(angle) + center,
		Vector2(-6.5, 7.0).rotated(angle) + center,
	])
	draw_colored_polygon(points, PLAYER_COLOR)
	draw_polyline(PackedVector2Array([points[0], points[1], points[2], points[3], points[0]]), Color(0.03, 0.08, 0.1, 1.0), 1.5, true)
