class_name FletchingWindow
extends Node

signal fletching_requested(recipe_id: String, quantity: int, maximum_available: int)

var _window: Window
var _recipe_grid: GridContainer
var _selection: Label
var _make_one: Button
var _make_x: Button
var _make_all: Button
var _quantity_dialog: ConfirmationDialog
var _quantity_input: SpinBox

var _item_icons: Dictionary
var _item_names: Dictionary
var _empty_slot_icon: Texture2D
var _recipes: Dictionary = {}
var _recipe_buttons: Dictionary = {}
var _inventory_items: Array = []
var _skill_level: int = 1
var _pending: bool = false
var _remaining: int = 0
var _selected_recipe: String = "arrow_shafts"
var _recipe_signature: String = ""
var _display_signature: String = ""


func configure(
	window: Window,
	item_icons: Dictionary,
	item_names: Dictionary,
	empty_slot_icon: Texture2D,
) -> void:
	_window = window
	_item_icons = item_icons
	_item_names = item_names
	_empty_slot_icon = empty_slot_icon
	_recipe_grid = _window.get_node("Panel/Margin/VBox/RecipeScroll/Recipes") as GridContainer
	_selection = _window.get_node("Panel/Margin/VBox/Selection") as Label
	_make_one = _window.get_node("Panel/Margin/VBox/Actions/Make1") as Button
	_make_x = _window.get_node("Panel/Margin/VBox/Actions/MakeX") as Button
	_make_all = _window.get_node("Panel/Margin/VBox/Actions/MakeAll") as Button
	var close := _window.get_node("Panel/Margin/VBox/Header/Close") as Button

	InterfaceStyle.configure_window(_window)
	InterfaceStyle.configure_close_button(close)
	close.pressed.connect(_window.hide)
	_window.close_requested.connect(_window.hide)
	_window.window_input.connect(_on_window_input)
	_make_one.pressed.connect(_emit_request.bind(1))
	_make_x.pressed.connect(_open_quantity_dialog)
	_make_all.pressed.connect(_emit_request.bind(0))
	_setup_quantity_dialog()


func update_data(
	recipes: Array,
	inventory_items: Array,
	skill_level: int,
	pending: bool,
	remaining: int,
) -> void:
	_inventory_items = inventory_items
	_skill_level = skill_level
	_pending = pending
	_remaining = remaining
	var recipe_signature := JSON.stringify(recipes)
	if recipe_signature != _recipe_signature:
		_recipe_signature = recipe_signature
		_rebuild_recipes(recipes)
	_refresh_if_changed()


func show_for_item(item_id: String = "") -> void:
	var matching_recipe := recipe_for_item(item_id)
	if not matching_recipe.is_empty():
		_selected_recipe = matching_recipe
	_display_signature = ""
	_refresh_if_changed()
	_window.popup_centered()


func is_item_input(item_id: String) -> bool:
	return not recipe_for_item(item_id).is_empty()


func recipe_for_item(item_id: String) -> String:
	if item_id.is_empty():
		return ""
	for recipe_id_value: Variant in _recipes:
		var recipe_id := str(recipe_id_value)
		var recipe := _recipes[recipe_id] as Dictionary
		if str(recipe.get("required_tool", "")) == item_id:
			return recipe_id
		for cost_data: Variant in recipe.get("costs", []) as Array:
			if typeof(cost_data) == TYPE_DICTIONARY and str((cost_data as Dictionary).get("item_id", "")) == item_id:
				return recipe_id
	return ""


func recipe_for_reward(item_id: String) -> Dictionary:
	for recipe_value: Variant in _recipes.values():
		var recipe := recipe_value as Dictionary
		if str(recipe.get("reward_item", "")) == item_id:
			return recipe
	return {}


func close_overlays() -> bool:
	if _quantity_dialog != null and _quantity_dialog.visible:
		_quantity_dialog.hide()
		return true
	return false


func _rebuild_recipes(recipes: Array) -> void:
	_recipes.clear()
	_recipe_buttons.clear()
	_clear_children(_recipe_grid)
	for payload: Variant in recipes:
		if typeof(payload) != TYPE_DICTIONARY:
			continue
		var recipe := (payload as Dictionary).duplicate(true)
		var recipe_id: String = str(recipe.get("id", ""))
		if recipe_id.is_empty():
			continue
		_recipes[recipe_id] = recipe
		var button := Button.new()
		button.custom_minimum_size = Vector2(280, 130)
		button.size_flags_horizontal = Control.SIZE_EXPAND_FILL
		button.toggle_mode = true
		button.focus_mode = Control.FOCUS_NONE
		button.text = _recipe_button_text(recipe)
		button.icon = _item_icons.get(str(recipe.get("reward_item", "")), _empty_slot_icon) as Texture2D
		button.expand_icon = true
		button.add_theme_constant_override("icon_max_width", 52)
		button.pressed.connect(_select_recipe.bind(recipe_id))
		_recipe_grid.add_child(button)
		_recipe_buttons[recipe_id] = button
	if not _recipes.has(_selected_recipe):
		_selected_recipe = str(_recipes.keys()[0]) if not _recipes.is_empty() else ""
	_display_signature = ""


func _recipe_button_text(recipe: Dictionary) -> String:
	var tool_id: String = str(recipe.get("required_tool", ""))
	var tool_text := ""
	if not tool_id.is_empty():
		tool_text = "\nTool: %s" % _item_name(tool_id)
	return "%s\nLv %d • %d XP\n%s%s" % [
		str(recipe.get("name", "Recipe")),
		int(recipe.get("required_level", 1)),
		int(recipe.get("reward_xp", 0)),
		_format_costs(recipe.get("costs", []) as Array),
		tool_text,
	]


func _select_recipe(recipe_id: String) -> void:
	if _pending or not _recipes.has(recipe_id):
		return
	_selected_recipe = recipe_id
	_display_signature = ""
	_refresh_if_changed()


func _refresh_if_changed() -> void:
	if _window == null:
		return
	var signature := "%s:%d:%s:%d:%s" % [
		_selected_recipe, _skill_level, _pending, _remaining, JSON.stringify(_inventory_items),
	]
	if signature == _display_signature:
		return
	_display_signature = signature
	var maximum := _maximum_batches()
	var recipe: Dictionary = _recipes.get(_selected_recipe, {}) as Dictionary
	var recipe_name: String = str(recipe.get("name", "Select a recipe"))
	for recipe_id_value: Variant in _recipe_buttons:
		var recipe_id := str(recipe_id_value)
		var button := _recipe_buttons[recipe_id] as Button
		button.set_pressed_no_signal(recipe_id == _selected_recipe)
		button.disabled = _pending
	if _pending:
		var interval: int = int(recipe.get("interval", 1))
		_selection.text = "Making %s every %.1f s • %d batch%s remaining" % [
			recipe_name, float(interval) * 0.6, _remaining, "" if _remaining == 1 else "es",
		]
	elif not recipe.is_empty() and _skill_level < int(recipe.get("required_level", 1)):
		_selection.text = "%s • Requires Fletching level %d" % [
			recipe_name, int(recipe.get("required_level", 1)),
		]
	else:
		_selection.text = "%s • %d XP per batch • materials for %d batch%s" % [
			recipe_name, int(recipe.get("reward_xp", 0)), maximum, "" if maximum == 1 else "es",
		]
	var unavailable := maximum < 1 or _pending
	_make_one.disabled = unavailable
	_make_x.disabled = unavailable
	_make_all.disabled = unavailable
	_make_one.text = "Fletching…" if _pending else "Make 1"


func _maximum_batches() -> int:
	var recipe: Dictionary = _recipes.get(_selected_recipe, {}) as Dictionary
	if recipe.is_empty() or _skill_level < int(recipe.get("required_level", 1)):
		return 0
	var tool_id: String = str(recipe.get("required_tool", ""))
	if not tool_id.is_empty() and _inventory_quantity(tool_id) < 1:
		return 0
	var costs: Array = recipe.get("costs", []) as Array
	if costs.is_empty():
		return 0
	var maximum: int = 2_147_483_647
	for cost_data: Variant in costs:
		if typeof(cost_data) != TYPE_DICTIONARY:
			return 0
		var cost := cost_data as Dictionary
		var quantity: int = int(cost.get("quantity", 0))
		if quantity < 1:
			return 0
		maximum = mini(maximum, floori(
			float(_inventory_quantity(str(cost.get("item_id", "")))) / float(quantity)
		))
	return maximum


func _inventory_quantity(item_id: String) -> int:
	var total: int = 0
	for payload: Variant in _inventory_items:
		if typeof(payload) == TYPE_DICTIONARY and str((payload as Dictionary).get("item_id", "")) == item_id:
			total += int((payload as Dictionary).get("quantity", 1))
	return total


func _format_costs(costs: Array) -> String:
	var labels: PackedStringArray = []
	for cost_data: Variant in costs:
		if typeof(cost_data) != TYPE_DICTIONARY:
			continue
		var cost := cost_data as Dictionary
		labels.append("%d %s" % [int(cost.get("quantity", 0)), _item_name(str(cost.get("item_id", "")))])
	return " + ".join(labels)


func _item_name(item_id: String) -> String:
	return str(_item_names.get(item_id, item_id.replace("_", " ").capitalize()))


func _setup_quantity_dialog() -> void:
	_quantity_dialog = ConfirmationDialog.new()
	InterfaceStyle.configure_window(_quantity_dialog)
	_quantity_dialog.title = "Fletch X batches"
	_quantity_dialog.dialog_text = "Choose how many batches to make:"
	_quantity_dialog.min_size = Vector2i(380, 180)
	_quantity_dialog.get_ok_button().text = "Start fletching"
	_quantity_input = SpinBox.new()
	_quantity_input.min_value = 1
	_quantity_input.step = 1
	_quantity_input.custom_minimum_size = Vector2(300, 44)
	_quantity_input.position = Vector2(40, 74)
	_quantity_dialog.add_child(_quantity_input)
	_quantity_dialog.confirmed.connect(_confirm_quantity)
	add_child(_quantity_dialog)


func _open_quantity_dialog() -> void:
	var maximum := _maximum_batches()
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
	var maximum := _maximum_batches()
	if _pending or maximum < 1 or _selected_recipe.is_empty():
		return
	fletching_requested.emit(_selected_recipe, quantity, maximum)


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
