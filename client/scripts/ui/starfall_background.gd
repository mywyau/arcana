class_name StarfallBackground
extends Control

const STAR_COUNT: int = 90
const GRADIENT_STEPS: int = 32
const MINIMUM_HEIGHT: float = 1.0

var _stars: Array[Dictionary] = []
var _random := RandomNumberGenerator.new()


func _ready() -> void:
	mouse_filter = Control.MOUSE_FILTER_IGNORE
	_random.seed = 0xA7CA4A
	resized.connect(_reset_stars)
	_reset_stars()
	set_process(is_visible_in_tree())


func _notification(what: int) -> void:
	if what == NOTIFICATION_VISIBILITY_CHANGED and is_node_ready():
		set_process(is_visible_in_tree())


func _process(delta: float) -> void:
	if size.x <= 0.0 or size.y <= 0.0:
		return
	for index: int in range(_stars.size()):
		var star: Dictionary = _stars[index]
		var position: Vector2 = star["position"] as Vector2
		position.x += float(star["drift"]) * delta / maxf(size.x, 1.0)
		position.y += float(star["speed"]) * delta / maxf(size.y, MINIMUM_HEIGHT)
		if position.y > 1.08 or position.x > 1.08:
			position = Vector2(
				_random.randf_range(-0.08, 1.0),
				_random.randf_range(-0.18, -0.02),
			)
		star["position"] = position
		_stars[index] = star
	queue_redraw()


func _draw() -> void:
	if size.x <= 0.0 or size.y <= 0.0:
		return
	_draw_night_gradient()
	_draw_stars()


func _draw_night_gradient() -> void:
	var top := Color("090721")
	var middle := Color("17103a")
	var bottom := Color("080a1c")
	var strip_height: float = size.y / float(GRADIENT_STEPS)
	for index: int in range(GRADIENT_STEPS):
		var progress: float = float(index) / float(GRADIENT_STEPS - 1)
		var colour: Color
		if progress < 0.58:
			colour = top.lerp(middle, progress / 0.58)
		else:
			colour = middle.lerp(bottom, (progress - 0.58) / 0.42)
		draw_rect(
			Rect2(0.0, strip_height * index, size.x, strip_height + 1.0),
			colour,
		)

	# Broad, low-alpha colour pools prevent the background feeling flat while
	# keeping the login card readable.
	draw_circle(Vector2(size.x * 0.18, size.y * 0.25), size.y * 0.24, Color(0.28, 0.16, 0.52, 0.08))
	draw_circle(Vector2(size.x * 0.82, size.y * 0.72), size.y * 0.30, Color(0.16, 0.28, 0.58, 0.07))


func _draw_stars() -> void:
	for star: Dictionary in _stars:
		var normalized_position: Vector2 = star["position"] as Vector2
		var position := Vector2(normalized_position.x * size.x, normalized_position.y * size.y)
		var brightness: float = float(star["brightness"])
		var radius: float = float(star["radius"])
		var speed: float = float(star["speed"])
		var drift: float = float(star["drift"])
		var colour := Color(0.82, 0.86, 1.0, brightness)
		if bool(star["warm"]):
			colour = Color(0.96, 0.82, 1.0, brightness)
		var travel := Vector2(drift, speed).normalized()
		var trail_length: float = float(star["trail"])
		if trail_length > 0.0:
			draw_line(position - travel * trail_length, position, Color(colour, brightness * 0.32), radius * 0.7, true)
		draw_circle(position, radius, colour)


func _reset_stars() -> void:
	_stars.clear()
	for index: int in range(STAR_COUNT):
		var is_falling_star: bool = index % 13 == 0
		_stars.append({
			"position": Vector2(_random.randf_range(-0.06, 1.02), _random.randf_range(-0.12, 1.02)),
			"speed": _random.randf_range(22.0, 48.0) if not is_falling_star else _random.randf_range(95.0, 145.0),
			"drift": _random.randf_range(3.0, 10.0) if not is_falling_star else _random.randf_range(34.0, 55.0),
			"radius": _random.randf_range(0.7, 1.6) if not is_falling_star else _random.randf_range(1.4, 2.1),
			"brightness": _random.randf_range(0.38, 0.82) if not is_falling_star else _random.randf_range(0.78, 1.0),
			"trail": 0.0 if not is_falling_star else _random.randf_range(26.0, 54.0),
			"warm": _random.randf() < 0.24,
		})
	queue_redraw()
