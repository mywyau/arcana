class_name InventoryWindow
extends Node

signal item_pressed(inventory_slot: int)
signal swap_requested(from_slot: int, to_slot: int)
signal context_requested(inventory_slot: int, screen_position: Vector2)

const INVENTORY_SLOT_SCRIPT: Script = preload("res://scripts/inventory_slot.gd")

var _summary: Label
var _previous_button: Button
var _page_label: Label
var _next_button: Button
var _grid: GridContainer
var _empty_slot_icon: Texture2D
var _total_slots: int = 0
var _page_size: int = 1
var _page: int = 0
var _capacity: int = 0
var _presentations: Dictionary = {}


func configure(
		summary: Label,
		previous_button: Button,
		page_label: Label,
		next_button: Button,
		grid: GridContainer,
		empty_slot_icon: Texture2D,
		total_slots: int,
		page_size: int,
) -> void:
	_summary = summary
	_previous_button = previous_button
	_page_label = page_label
	_next_button = next_button
	_grid = grid
	_empty_slot_icon = empty_slot_icon
	_total_slots = maxi(1, total_slots)
	_page_size = clampi(page_size, 1, _total_slots)
	_previous_button.pressed.connect(_change_page.bind(-1))
	_next_button.pressed.connect(_change_page.bind(1))


func render(presentations: Dictionary, capacity: int) -> void:
	_presentations = presentations
	_capacity = maxi(0, capacity)
	_rebuild()


func _change_page(direction: int) -> void:
	var page_count := ceili(float(_total_slots) / float(_page_size))
	_page = clampi(_page + direction, 0, page_count - 1)
	_rebuild()


func _rebuild() -> void:
	if _grid == null:
		return
	for child: Node in _grid.get_children():
		_grid.remove_child(child)
		child.queue_free()
	var unlocked_slots := mini(_capacity, _total_slots)
	var page_count := ceili(float(_total_slots) / float(_page_size))
	_page = clampi(_page, 0, page_count - 1)
	var first_slot := _page * _page_size
	var final_slot := mini(first_slot + _page_size, _total_slots)
	_summary.text = "Inventory  •  %d/%d slots used" % [_presentations.size(), unlocked_slots]
	_page_label.text = "Page %d/%d" % [_page + 1, page_count]
	_previous_button.disabled = _page == 0
	_next_button.disabled = _page >= page_count - 1
	for slot: int in range(first_slot, final_slot):
		_grid.add_child(_create_slot(slot, _presentations.get(slot, {}) as Dictionary))


func _create_slot(slot: int, presentation: Dictionary) -> Button:
	var slot_button: Variant = INVENTORY_SLOT_SCRIPT.new()
	slot_button.custom_minimum_size = Vector2(96, 96)
	slot_button.size_flags_horizontal = Control.SIZE_FILL
	slot_button.focus_mode = Control.FOCUS_NONE
	slot_button.mouse_filter = Control.MOUSE_FILTER_STOP
	slot_button.text = ""
	_style_slot(slot_button)
	if slot >= _capacity:
		slot_button.configure(slot, false, false)
		slot_button.icon = _empty_slot_icon
		slot_button.expand_icon = true
		slot_button.disabled = true
		slot_button.modulate = Color(0.28, 0.31, 0.34)
		slot_button.tooltip_text = "Locked inventory slot"
	elif presentation.is_empty():
		slot_button.configure(slot, false, true)
		slot_button.icon = _empty_slot_icon
		slot_button.expand_icon = true
		slot_button.modulate = Color(0.9, 0.93, 0.96)
	else:
		_configure_item_slot(slot_button, slot, presentation)
	slot_button.swap_requested.connect(_on_swap_requested)
	slot_button.context_requested.connect(_on_context_requested)
	return slot_button


func _configure_item_slot(slot_button: Variant, slot: int, presentation: Dictionary) -> void:
	slot_button.configure(slot, true, true)
	slot_button.icon = presentation.get("icon", _empty_slot_icon) as Texture2D
	slot_button.expand_icon = true
	slot_button.set_stack_quantity(int(presentation.get("quantity", 1)))
	slot_button.set_reinforcement(int(presentation.get("reinforcement", 0)))
	slot_button.toggle_mode = bool(presentation.get("toggle_mode", false))
	slot_button.set_pressed_no_signal(bool(presentation.get("pressed", false)))
	slot_button.tooltip_text = str(presentation.get("tooltip", ""))
	slot_button.disabled = bool(presentation.get("disabled", false))
	var highlight: String = str(presentation.get("highlight", ""))
	if not highlight.is_empty():
		_apply_highlight(slot_button, highlight)
	slot_button.pressed.connect(_on_item_pressed.bind(slot))


func _style_slot(slot_button: Button) -> void:
	var slot_style := StyleBoxFlat.new()
	slot_style.bg_color = Color(0.055, 0.065, 0.075, 0.94)
	slot_style.border_width_left = 1
	slot_style.border_width_top = 1
	slot_style.border_width_right = 1
	slot_style.border_width_bottom = 1
	slot_style.border_color = Color(0.2, 0.23, 0.26, 0.95)
	slot_style.corner_radius_top_left = 5
	slot_style.corner_radius_top_right = 5
	slot_style.corner_radius_bottom_left = 5
	slot_style.corner_radius_bottom_right = 5
	slot_button.add_theme_stylebox_override("normal", slot_style)
	slot_button.add_theme_stylebox_override("hover", slot_style.duplicate())


func _apply_highlight(slot_button: Button, highlight: String) -> void:
	var selection_style := StyleBoxFlat.new()
	selection_style.bg_color = Color(0.06, 0.09, 0.12, 0.96)
	selection_style.border_width_left = 4
	selection_style.border_width_top = 4
	selection_style.border_width_right = 4
	selection_style.border_width_bottom = 4
	selection_style.border_color = (
		Color(0.35, 0.7, 1.0) if highlight == "pending" else Color(0.95, 0.72, 0.2)
	)
	selection_style.corner_radius_top_left = 6
	selection_style.corner_radius_top_right = 6
	selection_style.corner_radius_bottom_left = 6
	selection_style.corner_radius_bottom_right = 6
	slot_button.add_theme_stylebox_override("pressed", selection_style)
	slot_button.add_theme_stylebox_override("hover_pressed", selection_style)


func _on_item_pressed(inventory_slot: int) -> void:
	item_pressed.emit(inventory_slot)


func _on_swap_requested(from_slot: int, to_slot: int) -> void:
	swap_requested.emit(from_slot, to_slot)


func _on_context_requested(inventory_slot: int, screen_position: Vector2) -> void:
	context_requested.emit(inventory_slot, screen_position)
