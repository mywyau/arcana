class_name BrewingWindow
extends Node

signal brewing_requested(item_ids: Array[String], quantity: int, maximum_available: int)

var _window: Window
var _slots: Array[Button] = []
var _result: Label
var _ingredient_grid: GridContainer
var _clear: Button
var _make_one: Button
var _make_x: Button
var _make_all: Button
var _quantity_dialog: ConfirmationDialog
var _quantity_input: SpinBox

var _ingredients: Dictionary
var _known_brews: Dictionary
var _item_icons: Dictionary
var _empty_slot_icon: Texture2D
var _inventory_items: Array = []
var _selection: Array[String] = []
var _pending: bool = false
var _remaining: int = 0
var _display_signature: String = ""


func configure(
	window: Window,
	ingredients: Dictionary,
	known_brews: Dictionary,
	item_icons: Dictionary,
	empty_slot_icon: Texture2D,
) -> void:
	_window = window
	_ingredients = ingredients
	_known_brews = known_brews
	_item_icons = item_icons
	_empty_slot_icon = empty_slot_icon
	_slots = [
		_window.get_node("Panel/Margin/VBox/Selected/Slot1") as Button,
		_window.get_node("Panel/Margin/VBox/Selected/Slot2") as Button,
		_window.get_node("Panel/Margin/VBox/Selected/Slot3") as Button,
	]
	_result = _window.get_node("Panel/Margin/VBox/Result") as Label
	_ingredient_grid = _window.get_node("Panel/Margin/VBox/IngredientScroll/Grid") as GridContainer
	_clear = _window.get_node("Panel/Margin/VBox/Actions/Clear") as Button
	_make_one = _window.get_node("Panel/Margin/VBox/Actions/Brew") as Button
	_make_x = _window.get_node("Panel/Margin/VBox/Actions/MakeX") as Button
	_make_all = _window.get_node("Panel/Margin/VBox/Actions/MakeAll") as Button
	var close := _window.get_node("Panel/Margin/VBox/Header/Close") as Button

	InterfaceStyle.configure_window(_window)
	InterfaceStyle.configure_close_button(close)
	close.pressed.connect(_window.hide)
	_window.close_requested.connect(_window.hide)
	_window.window_input.connect(_on_window_input)
	for slot_index: int in range(_slots.size()):
		_slots[slot_index].pressed.connect(_remove_ingredient.bind(slot_index))
	_clear.pressed.connect(clear_selection)
	_make_one.pressed.connect(_emit_request.bind(1))
	_make_x.pressed.connect(_open_quantity_dialog)
	_make_all.pressed.connect(_emit_request.bind(0))
	_setup_quantity_dialog()


func update_data(inventory_items: Array, pending: bool, remaining: int) -> void:
	_inventory_items = inventory_items
	_pending = pending
	_remaining = remaining
	_refresh_if_changed()


func show_for_item(item_id: String = "") -> void:
	if not _window.visible:
		_selection.clear()
	if not item_id.is_empty():
		_add_ingredient(item_id)
	_display_signature = ""
	_refresh_if_changed()
	_window.popup_centered()


func is_ingredient(item_id: String) -> bool:
	return _ingredients.has(item_id)


func clear_selection() -> void:
	if _pending:
		return
	_selection.clear()
	_display_signature = ""
	_refresh_if_changed()


func close_overlays() -> bool:
	if _quantity_dialog != null and _quantity_dialog.visible:
		_quantity_dialog.hide()
		return true
	return false


func _add_ingredient(item_id: String) -> void:
	if _pending or _selection.size() >= 3 or not _ingredients.has(item_id):
		return
	if _selection.count(item_id) >= _inventory_quantity(item_id):
		return
	_selection.append(item_id)
	_display_signature = ""
	_refresh_if_changed()


func _remove_ingredient(slot_index: int) -> void:
	if _pending or slot_index < 0 or slot_index >= _selection.size():
		return
	_selection.remove_at(slot_index)
	_display_signature = ""
	_refresh_if_changed()


func _refresh_if_changed() -> void:
	if _window == null:
		return
	var signature := "%s:%s:%d:%s" % [
		JSON.stringify(_selection), _pending, _remaining, JSON.stringify(_inventory_items),
	]
	if signature == _display_signature:
		return
	_display_signature = signature
	_refresh_slots()
	_refresh_ingredients()
	_refresh_result()


func _refresh_slots() -> void:
	for slot_index: int in range(_slots.size()):
		var slot := _slots[slot_index]
		if slot_index < _selection.size():
			var item_id: String = _selection[slot_index]
			slot.text = str((_ingredients[item_id] as Dictionary).get("name", item_id.capitalize()))
			slot.icon = _item_icons.get(item_id, _empty_slot_icon) as Texture2D
			slot.expand_icon = true
			slot.tooltip_text = "Click to remove"
		else:
			slot.text = "+ Optional" if slot_index == 2 else "+ Ingredient"
			slot.icon = null
			slot.tooltip_text = ""


func _refresh_ingredients() -> void:
	_clear_children(_ingredient_grid)
	for item_id_value: Variant in _ingredients.keys():
		var item_id := str(item_id_value)
		var quantity := _inventory_quantity(item_id)
		if quantity <= 0:
			continue
		var ingredient := _ingredients[item_id] as Dictionary
		var button := Button.new()
		button.custom_minimum_size = Vector2(158, 96)
		button.size_flags_horizontal = Control.SIZE_EXPAND_FILL
		button.text = "%s\n%s ×%d" % [
			ingredient.get("name", item_id.capitalize()), ingredient.get("role", "Ingredient"), quantity,
		]
		button.icon = _item_icons.get(item_id, _empty_slot_icon) as Texture2D
		button.expand_icon = true
		button.add_theme_constant_override("icon_max_width", 48)
		button.disabled = _pending or _selection.size() >= 3 or _selection.count(item_id) >= quantity
		button.pressed.connect(_add_ingredient.bind(item_id))
		_ingredient_grid.add_child(button)


func _refresh_result() -> void:
	var valid := _selection_is_valid()
	if valid:
		var key_items := _selection.duplicate()
		key_items.sort()
		var mixture_key := "+".join(key_items)
		_result.text = (
			"Known result: %s" % _known_brews[mixture_key]
			if _known_brews.has(mixture_key)
			else "Unknown mixture: experimentation may create a Murky Brew"
		)
	else:
		_result.text = "Requires exactly one Vial of Water and at least one herb"
	if _pending:
		_result.text = "Brewing one potion every 600 ms • %d remaining" % _remaining
	_make_one.disabled = not valid or _pending
	_make_one.text = "Brewing…" if _pending else "Make 1"
	_make_x.disabled = not valid or _pending
	_make_all.disabled = not valid or _pending
	_clear.disabled = _pending or _selection.is_empty()


func _selection_is_valid() -> bool:
	if _selection.size() < 2 or _selection.size() > 3 or _selection.count("vial_of_water") != 1:
		return false
	return _selection.has("sunleaf") or _selection.has("moonleaf")


func _maximum_brews() -> int:
	if _selection.is_empty():
		return 0
	var costs: Dictionary = {}
	for item_id: String in _selection:
		costs[item_id] = int(costs.get(item_id, 0)) + 1
	var maximum: int = 2_147_483_647
	for item_id_value: Variant in costs.keys():
		var item_id := str(item_id_value)
		maximum = mini(maximum, floori(
			float(_inventory_quantity(item_id)) / float(int(costs[item_id]))
		))
	return 0 if maximum == 2_147_483_647 else maximum


func _inventory_quantity(item_id: String) -> int:
	var total: int = 0
	for payload: Variant in _inventory_items:
		if typeof(payload) == TYPE_DICTIONARY and str((payload as Dictionary).get("item_id", "")) == item_id:
			total += int((payload as Dictionary).get("quantity", 1))
	return total


func _setup_quantity_dialog() -> void:
	_quantity_dialog = ConfirmationDialog.new()
	InterfaceStyle.configure_window(_quantity_dialog)
	_quantity_dialog.title = "Make X potions"
	_quantity_dialog.dialog_text = "Choose how many potions to brew:"
	_quantity_dialog.min_size = Vector2i(360, 180)
	_quantity_dialog.get_ok_button().text = "Start brewing"
	_quantity_input = SpinBox.new()
	_quantity_input.min_value = 1
	_quantity_input.step = 1
	_quantity_input.custom_minimum_size = Vector2(280, 44)
	_quantity_input.position = Vector2(40, 74)
	_quantity_dialog.add_child(_quantity_input)
	_quantity_dialog.confirmed.connect(_confirm_quantity)
	add_child(_quantity_dialog)


func _open_quantity_dialog() -> void:
	var maximum := _maximum_brews()
	if maximum < 1 or _pending:
		return
	_quantity_input.max_value = maximum
	_quantity_input.value = mini(maximum, 5)
	_quantity_dialog.popup_centered()
	_quantity_input.grab_focus()
	_quantity_input.get_line_edit().select_all()


func _confirm_quantity() -> void:
	_emit_request(int(_quantity_input.value))


func _emit_request(quantity: int) -> void:
	var maximum := _maximum_brews()
	if _pending or not _selection_is_valid() or maximum < 1:
		return
	brewing_requested.emit(_selection.duplicate(), quantity, maximum)


func _on_window_input(event: InputEvent) -> void:
	if event is InputEventKey:
		var key_event := event as InputEventKey
		if key_event.pressed and not key_event.echo and key_event.keycode == KEY_ESCAPE:
			_window.hide()
			_window.set_input_as_handled()


func _clear_children(parent: Node) -> void:
	for child: Node in parent.get_children():
		parent.remove_child(child)
		child.queue_free()
