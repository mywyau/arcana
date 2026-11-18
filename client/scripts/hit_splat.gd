class_name HitSplat
extends Node3D

@onready var value_label: Label3D = $Value

var _animation: Tween
var _base_position: Vector3


func _ready() -> void:
	_base_position = position
	visible = false


func show_damage(amount: int, splat_color: Color) -> void:
	_show_splat(str(maxi(0, amount)), splat_color, false)


func show_status(status: String, splat_color: Color) -> void:
	_show_splat(status, splat_color, true)


func show_healing(amount: int) -> void:
	_show_splat("+%d" % maxi(0, amount), Color(0.16, 0.62, 0.28, 1.0), false)


func _show_splat(text: String, splat_color: Color, compact: bool) -> void:
	if _animation != null:
		_animation.kill()
	position = _base_position
	scale = Vector3.ONE * 0.72
	value_label.modulate = splat_color
	value_label.text = text
	value_label.font_size = 42 if compact else 60
	visible = true

	_animation = create_tween()
	_animation.set_parallel(true)
	_animation.set_trans(Tween.TRANS_BACK).set_ease(Tween.EASE_OUT)
	_animation.tween_property(self, "scale", Vector3.ONE, 0.12)
	_animation.tween_property(self, "position:y", _base_position.y + 0.16, 0.62)
	_animation.tween_property(value_label, "modulate:a", 0.0, 0.28).set_delay(0.38)
	_animation.chain().tween_callback(func() -> void: visible = false)
