class_name XPTracker
extends PanelContainer

@onready var total_label: Label = $Margin/VBox/TotalRow/Total
@onready var total_skill_icon: TextureRect = $Margin/VBox/TotalRow/Icon
@onready var gain_row: HBoxContainer = $Margin/VBox/GainRow
@onready var gain_label: Label = $Margin/VBox/GainRow/Gain
@onready var skill_icon: TextureRect = $Margin/VBox/GainRow/Icon
@onready var drop_layer: Control = $DropLayer

var _active_drops: Array[Control] = []


func show_gain(skill_name: String, icon: Texture2D, gained_xp: int, total_xp: int) -> void:
	total_label.text = "%d XP" % total_xp
	total_skill_icon.texture = icon
	total_skill_icon.tooltip_text = skill_name
	_active_drops = _active_drops.filter(func(drop: Control) -> bool: return is_instance_valid(drop))

	var drop := gain_row.duplicate() as HBoxContainer
	drop.name = "XPDrop"
	drop.visible = true
	drop.mouse_filter = Control.MOUSE_FILTER_IGNORE
	drop_layer.add_child(drop)
	var drop_label := drop.get_node("Gain") as Label
	var drop_icon := drop.get_node("Icon") as TextureRect
	drop_label.text = "+%d XP" % gained_xp
	drop_icon.texture = icon
	drop_icon.tooltip_text = skill_name

	var stack_offset: float = float(mini(_active_drops.size(), 5)) * 9.0
	var start_y: float = 68.0 + stack_offset
	drop.set_anchors_preset(Control.PRESET_TOP_RIGHT)
	drop.offset_left = -150.0
	drop.offset_right = -12.0
	drop.offset_top = start_y
	drop.offset_bottom = start_y + 22.0
	_active_drops.append(drop)

	var rise := create_tween()
	rise.tween_property(drop, "position:y", start_y - 48.0, 1.5).set_trans(Tween.TRANS_QUAD).set_ease(Tween.EASE_OUT)
	var fade := create_tween()
	fade.tween_interval(0.8)
	fade.tween_property(drop, "modulate:a", 0.0, 0.7)
	fade.finished.connect(_remove_drop.bind(drop))


func _remove_drop(drop: Control) -> void:
	_active_drops.erase(drop)
	if is_instance_valid(drop):
		drop.queue_free()
