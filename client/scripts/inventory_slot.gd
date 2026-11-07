class_name InventorySlot
extends Button

signal swap_requested(from_slot: int, to_slot: int)

var slot_index: int = -1
var contains_item: bool = false
var unlocked: bool = false
var _quantity_label: Label


func configure(index: int, has_item: bool, is_unlocked: bool) -> void:
	slot_index = index
	contains_item = has_item
	unlocked = is_unlocked


func set_stack_quantity(quantity: int) -> void:
	if _quantity_label == null:
		_quantity_label = Label.new()
		_quantity_label.mouse_filter = Control.MOUSE_FILTER_IGNORE
		_quantity_label.set_anchors_preset(Control.PRESET_BOTTOM_RIGHT)
		_quantity_label.offset_left = -72.0
		_quantity_label.offset_top = -31.0
		_quantity_label.offset_right = -7.0
		_quantity_label.offset_bottom = -5.0
		_quantity_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_RIGHT
		_quantity_label.vertical_alignment = VERTICAL_ALIGNMENT_BOTTOM
		_quantity_label.add_theme_font_size_override("font_size", 18)
		_quantity_label.add_theme_color_override("font_color", Color(1.0, 0.94, 0.65))
		_quantity_label.add_theme_color_override("font_outline_color", Color(0.02, 0.025, 0.03, 1.0))
		_quantity_label.add_theme_constant_override("outline_size", 5)
		_quantity_label.z_index = 1
		add_child(_quantity_label)
	_quantity_label.text = _compact_quantity(quantity)
	_quantity_label.visible = quantity > 1


func _compact_quantity(quantity: int) -> String:
	if quantity >= 1_000_000:
		return _compact_quantity_unit(quantity, 1_000_000, "m")
	if quantity >= 1_000:
		return _compact_quantity_unit(quantity, 1_000, "k")
	return str(quantity)


func _compact_quantity_unit(quantity: int, divisor: int, suffix: String) -> String:
	if quantity % divisor == 0:
		return "%d%s" % [quantity / divisor, suffix]
	var abbreviated: float = floorf(float(quantity) / float(divisor) * 10.0) / 10.0
	return "%.1f%s" % [abbreviated, suffix]


func _get_drag_data(_at_position: Vector2) -> Variant:
	if not unlocked or not contains_item or icon == null:
		return null
	var preview := TextureRect.new()
	preview.custom_minimum_size = Vector2(64, 64)
	preview.texture = icon
	preview.expand_mode = TextureRect.EXPAND_IGNORE_SIZE
	preview.stretch_mode = TextureRect.STRETCH_KEEP_ASPECT_CENTERED
	set_drag_preview(preview)
	return {"inventory_slot": slot_index}


func _can_drop_data(_at_position: Vector2, data: Variant) -> bool:
	return unlocked and data is Dictionary and (data as Dictionary).has("inventory_slot")


func _drop_data(_at_position: Vector2, data: Variant) -> void:
	var from_slot: int = int((data as Dictionary).get("inventory_slot", -1))
	if from_slot >= 0 and from_slot != slot_index:
		swap_requested.emit(from_slot, slot_index)
