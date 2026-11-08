class_name XPTracker
extends MarginContainer

@onready var total_label: Label = $Margin/VBox/Total
@onready var gain_row: HBoxContainer = $Margin/VBox/GainRow
@onready var gain_label: Label = $Margin/VBox/GainRow/Gain
@onready var skill_icon: TextureRect = $Margin/VBox/GainRow/Icon

var _gain_generation: int = 0


func show_gain(skill_name: String, icon: Texture2D, gained_xp: int, total_xp: int) -> void:
	_gain_generation += 1
	var generation: int = _gain_generation
	skill_icon.texture = icon
	skill_icon.tooltip_text = skill_name
	total_label.text = "Total %s XP  %d" % [skill_name, total_xp]
	gain_label.text = "+%d XP" % gained_xp
	gain_row.modulate = Color.WHITE

	await get_tree().create_timer(3.0).timeout
	if generation != _gain_generation or not is_instance_valid(gain_row):
		return
	var fade := create_tween()
	fade.tween_property(gain_row, "modulate:a", 0.0, 0.35)
