class_name InventorySlot
extends Button

signal swap_requested(from_slot: int, to_slot: int)

var slot_index: int = -1
var contains_item: bool = false
var unlocked: bool = false


func configure(index: int, has_item: bool, is_unlocked: bool) -> void:
	slot_index = index
	contains_item = has_item
	unlocked = is_unlocked


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
