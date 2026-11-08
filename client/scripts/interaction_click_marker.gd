class_name InteractionClickMarker
extends Control

var _progress: float = 1.0
var _tween: Tween
var _colour: Color = Color(0.78, 0.32, 1.0)


func play_at(screen_position: Vector2, colour: Color = Color(0.78, 0.32, 1.0)) -> void:
	if _tween != null and _tween.is_valid():
		_tween.kill()
	position = screen_position - size * 0.5
	_colour = colour
	_progress = 0.0
	visible = true
	queue_redraw()
	_tween = create_tween()
	_tween.tween_method(_set_progress, 0.0, 1.0, 0.48).set_trans(Tween.TRANS_QUART).set_ease(Tween.EASE_OUT)
	_tween.tween_callback(hide)


func _set_progress(value: float) -> void:
	_progress = value
	queue_redraw()


func _draw() -> void:
	var centre := size * 0.5
	var alpha := 1.0 if _progress < 0.58 else 1.0 - inverse_lerp(0.58, 1.0, _progress)
	var radius := lerpf(7.0, 11.0, _progress)
	var marker_colour := Color(_colour.r, _colour.g, _colour.b, alpha)
	var shadow := Color(0.06, 0.01, 0.09, alpha)
	var top_left := centre + Vector2(-radius, -radius)
	var bottom_right := centre + Vector2(radius, radius)
	var top_right := centre + Vector2(radius, -radius)
	var bottom_left := centre + Vector2(-radius, radius)
	draw_line(top_left, bottom_right, shadow, 6.0, true)
	draw_line(top_right, bottom_left, shadow, 6.0, true)
	draw_line(top_left, bottom_right, marker_colour, 3.0, true)
	draw_line(top_right, bottom_left, marker_colour, 3.0, true)
