class_name DeathProtectionWindow
extends Node

signal selection_requested(slots: Array[int], keys: Array[String])
signal feedback_requested(message: String)

const MAXIMUM_PROTECTED_ITEMS: int = 4
const InventorySlotScript: Script = preload("res://scripts/inventory_slot.gd")

var _window: Window
var _pending := false
var _summary: Label
var _inventory_grid: GridContainer
var _equipment_grid: GridContainer
var _equipped: Array = []
var _selected_keys: Array[String] = []
var _items: Array = []
var _selected_slots: Array[int] = []
var _icons: Dictionary = {}
var _display_names: Dictionary = {}
var _empty_icon: Texture2D


func configure(icons: Dictionary, display_names: Dictionary, empty_icon: Texture2D) -> void:
	_icons = icons
	_display_names = display_names
	_empty_icon = empty_icon
	_build_window()


func set_pending(value: bool) -> void:
	if value == _pending:
		return
	_pending = value
	if is_open():
		_render()


func open() -> void:
	_render()
	_window.popup_centered()


func close() -> void:
	if _window != null:
		_window.hide()


func is_open() -> bool:
	return _window != null and _window.visible


func update_data(items: Array, selected_slots: Array[int], equipped: Array, keys: Array[String]) -> void:
	if _items == items and _selected_slots == selected_slots and _equipped == equipped and _selected_keys == keys:
		return
	_equipped = equipped.duplicate(true)
	_selected_keys = keys.duplicate()
	_items = items.duplicate(true)
	_selected_slots = selected_slots.duplicate()
	_selected_slots.sort()
	if _window != null and _window.visible:
		_render()


func _build_window() -> void:
	_window = Window.new()
	_window.title = "Items Kept on Death"
	_window.borderless = true
	_window.min_size = Vector2i(520, 470)
	_window.size = Vector2i(520, 470)
	_window.visible = false
	_window.close_requested.connect(close)
	_window.window_input.connect(_on_window_input)
	add_child(_window)

	var panel := PanelContainer.new()
	panel.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	panel.add_theme_stylebox_override("panel", InterfaceStyle.make_panel_style())
	_window.add_child(panel)
	var margin := MarginContainer.new()
	margin.add_theme_constant_override("margin_left", 18)
	margin.add_theme_constant_override("margin_top", 14)
	margin.add_theme_constant_override("margin_right", 14)
	margin.add_theme_constant_override("margin_bottom", 18)
	panel.add_child(margin)
	var layout := VBoxContainer.new()
	layout.add_theme_constant_override("separation", 12)
	margin.add_child(layout)

	var header := HBoxContainer.new()
	layout.add_child(header)
	var title := Label.new()
	title.text = "Items Kept on Death"
	title.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	title.add_theme_font_size_override("font_size", 23)
	header.add_child(title)
	var close_button := Button.new()
	InterfaceStyle.configure_close_button(close_button)
	close_button.tooltip_text = "Close"
	close_button.pressed.connect(close)
	header.add_child(close_button)

	_summary = Label.new()
	_summary.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	layout.add_child(_summary)
	var scroll := ScrollContainer.new()
	scroll.size_flags_vertical = Control.SIZE_EXPAND_FILL
	scroll.horizontal_scroll_mode = ScrollContainer.SCROLL_MODE_DISABLED
	layout.add_child(scroll)
	var sections := VBoxContainer.new()
	scroll.add_child(sections)
	_equipment_grid = _add_section(sections, "Equipped")
	_inventory_grid = _add_section(sections, "Inventory")


func _add_section(sections: VBoxContainer, title: String) -> GridContainer:
	var label := Label.new()
	label.text = title
	sections.add_child(label)
	var grid := GridContainer.new()
	grid.columns = 4
	grid.add_theme_constant_override("h_separation", 8)
	grid.add_theme_constant_override("v_separation", 8)
	sections.add_child(grid)
	return grid


func _render() -> void:
	for grid: GridContainer in [_equipment_grid, _inventory_grid]:
		for child: Node in grid.get_children():
			grid.remove_child(child)
			child.queue_free()
	_summary.text = "Updating protection…" if _pending else "%d/%d protected" % [_selected_slots.size() + _selected_keys.size(), MAXIMUM_PROTECTED_ITEMS]
	for item: Dictionary in _equipped:
		_add_item(_equipment_grid, item)
	var ordered := _items.duplicate(true)
	ordered.sort_custom(func(left: Dictionary, right: Dictionary) -> bool:
		return int(left.get("slot", -1)) < int(right.get("slot", -1))
	)
	for payload: Variant in ordered:
		if typeof(payload) != TYPE_DICTIONARY:
			continue
		var item := payload as Dictionary
		_add_item(_inventory_grid, item)


func _add_item(grid: GridContainer, item: Dictionary) -> void:
	var slot: int = int(item.get("slot", -1))
	var item_id: String = str(item.get("item_id", ""))
	var quantity: int = int(item.get("quantity", 0))
	var key := str(item.get("key", ""))
	if (slot < 0 and key.is_empty()) or item_id.is_empty() or quantity <= 0:
		return
	var selected: bool = key in _selected_keys if not key.is_empty() else slot in _selected_slots
	var button: Button = InventorySlotScript.new()
	button.custom_minimum_size = Vector2(108, 92)
	button.focus_mode = Control.FOCUS_NONE
	button.disabled = _pending
	button.configure(slot, true, true)
	button.icon = _icons.get(item_id, _empty_icon) as Texture2D
	button.expand_icon = true
	button.set_stack_quantity(quantity)
	button.set_reinforcement(int(item.get("reinforcement", 0)))
	button.tooltip_text = "%s%s • %s" % [
		str(_display_names.get(item_id, item_id.capitalize())),
		" ×%d" % quantity if quantity > 1 else "",
		"protected" if selected else "click to protect",
	]
	var style := StyleBoxFlat.new()
	style.bg_color = Color(0.22, 0.23, 0.38, 0.62)
	style.set_corner_radius_all(8)
	style.set_border_width_all(3 if selected else 1)
	style.border_color = Color(0.98, 0.78, 0.28, 1.0) if selected else Color(0.5, 0.48, 0.72, 0.65)
	button.add_theme_stylebox_override("normal", style)
	var hover := style.duplicate() as StyleBoxFlat
	hover.bg_color = Color(0.32, 0.3, 0.5, 0.85)
	button.add_theme_stylebox_override("hover", hover)
	button.pressed.connect(_toggle_slot.bind(slot, key))
	grid.add_child(button)


func _toggle_slot(slot: int, key: String) -> void:
	var slots := _selected_slots.duplicate()
	var keys := _selected_keys.duplicate()
	if not key.is_empty() and key in keys:
		keys.erase(key)
	elif key.is_empty() and slot in slots:
		slots.erase(slot)
	elif slots.size() + keys.size() >= MAXIMUM_PROTECTED_ITEMS:
		feedback_requested.emit("You can protect up to four items on death.")
		return
	elif key.is_empty():
		slots.append(slot)
	else:
		keys.append(key)
	selection_requested.emit(slots, keys)


func _on_window_input(event: InputEvent) -> void:
	if event is InputEventKey:
		var key_event := event as InputEventKey
		if key_event.pressed and not key_event.echo and key_event.keycode == KEY_ESCAPE:
			close()
			_window.set_input_as_handled()
