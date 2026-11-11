class_name ProductionWindows
extends Node

signal production_requested(station: String, recipe_id: String, instance_id: String, quantity: int)

const GROUP_ORDER: Array[String] = ["bronze", "iron", "steel", "mithril", "adamantite"]
const MAXIMUM_REINFORCEMENT: int = 5

var _smithing_window: Window
var _smithing_tabs: TabContainer
var _smelting_window: Window
var _smelting_grid: GridContainer
var _cooking_window: Window
var _cooking_grid: GridContainer
var _cooking_details_toggle: Button

var _item_icons: Dictionary
var _item_names: Dictionary
var _empty_slot_icon: Texture2D
var _reinforcement_tiers: Dictionary
var _armour_slots: Dictionary
var _equipment_icons: Dictionary

var _recipes: Dictionary = {}
var _skill_levels: Dictionary = {}
var _inventory_items: Array = []
var _cooking_details_visible: bool = false
var _batch_station: String = ""
var _batch_recipe_id: String = ""
var _action_menu: PopupMenu
var _quantity_dialog: ConfirmationDialog
var _quantity_input: SpinBox


func configure(
	smithing_window: Window,
	smelting_window: Window,
	cooking_window: Window,
	item_icons: Dictionary,
	item_names: Dictionary,
	empty_slot_icon: Texture2D,
	reinforcement_tiers: Dictionary,
	armour_slots: Dictionary,
	equipment_icons: Dictionary,
) -> void:
	_smithing_window = smithing_window
	_smelting_window = smelting_window
	_cooking_window = cooking_window
	_item_icons = item_icons
	_item_names = item_names
	_empty_slot_icon = empty_slot_icon
	_reinforcement_tiers = reinforcement_tiers
	_armour_slots = armour_slots
	_equipment_icons = equipment_icons

	_smithing_tabs = _smithing_window.get_node("Panel/Margin/VBox/Tabs") as TabContainer
	_smelting_grid = _smelting_window.get_node("Panel/Margin/VBox/RecipeScroll/Grid") as GridContainer
	_cooking_grid = _cooking_window.get_node("Panel/Margin/VBox/RecipeScroll/Grid") as GridContainer
	_cooking_details_toggle = _cooking_window.get_node("Panel/Margin/VBox/IngredientsToggle") as Button

	for window: Window in [_smithing_window, _smelting_window, _cooking_window]:
		InterfaceStyle.configure_window(window)
		window.close_requested.connect(window.hide)
		window.window_input.connect(_on_window_input.bind(window))
		var close := window.get_node("Panel/Margin/VBox/Header/Close") as Button
		InterfaceStyle.configure_close_button(close)
		close.pressed.connect(window.hide)
	_cooking_details_toggle.toggled.connect(_on_cooking_details_toggled)
	_setup_batch_controls()


func update_data(recipes: Dictionary, skill_progress: Dictionary, inventory_items: Array) -> void:
	_recipes = recipes
	_inventory_items = inventory_items
	_skill_levels = {
		"cooking": _skill_level(skill_progress, "cooking"),
		"smithing": _skill_level(skill_progress, "smithing"),
	}
	refresh_visible()


func show_station(station: String) -> void:
	match station:
		"cooking":
			_render_cooking()
			_cooking_window.popup_centered()
		"smelting":
			_render_smelting()
			_smelting_window.popup_centered()
		"forging":
			_render_forging()
			_smithing_window.popup_centered()


func refresh_visible() -> void:
	if _cooking_window != null and _cooking_window.visible:
		_render_cooking()
	if _smelting_window != null and _smelting_window.visible:
		_render_smelting()
	if _smithing_window != null and _smithing_window.visible:
		_render_forging()


func request_batch(station: String, recipe_id: String, quantity: int) -> void:
	_batch_station = station
	_batch_recipe_id = recipe_id
	if quantity < 0:
		_open_quantity_dialog()
	else:
		_emit_batch(quantity)


func close_overlays() -> bool:
	if _action_menu != null and _action_menu.visible:
		_action_menu.hide()
		return true
	if _quantity_dialog != null and _quantity_dialog.visible:
		_quantity_dialog.hide()
		return true
	return false


func _render_cooking() -> void:
	_clear_children(_cooking_grid)
	_cooking_details_toggle.set_pressed_no_signal(_cooking_details_visible)
	var cooking_level: int = int(_skill_levels.get("cooking", 1))
	for recipe_data: Variant in _station_recipes("cooking"):
		var recipe := recipe_data as Dictionary
		var recipe_id: String = str(recipe.get("id", ""))
		var reward_item: String = str(recipe.get("reward_item", recipe_id))
		var costs: Array = recipe.get("costs", []) as Array
		var required_level: int = int(recipe.get("required_level", 1))
		var available: int = _maximum_batches(costs)
		if available < 1 or cooking_level < required_level:
			continue
		var input_text: String = _format_costs(costs, " + ")
		var button := _recipe_button(reward_item, 62)
		button.custom_minimum_size = Vector2(260, 116 if _cooking_details_visible else 92)
		button.text = "%s\nLv %d" % [recipe.get("name", recipe_id.capitalize()), required_level]
		if _cooking_details_visible:
			button.text += "\n%s" % input_text
		button.tooltip_text = "Cook %s until you run out" % recipe.get("name", recipe_id.capitalize())
		button.pressed.connect(_request_default.bind("cooking", recipe_id, ""))
		button.gui_input.connect(_on_recipe_input.bind("cooking", recipe_id))
		_cooking_grid.add_child(button)


func _render_smelting() -> void:
	_clear_children(_smelting_grid)
	var smithing_level: int = int(_skill_levels.get("smithing", 1))
	for recipe_data: Variant in _station_recipes("smelting"):
		var recipe := recipe_data as Dictionary
		var recipe_id: String = str(recipe.get("id", ""))
		var reward_item: String = str(recipe.get("reward_item", recipe_id))
		var required_level: int = int(recipe.get("required_level", 1))
		var costs: Array = recipe.get("costs", []) as Array
		var input_lines: String = _format_costs(costs, "\n")
		var required_tool: String = str(recipe.get("required_tool", ""))
		var has_tool: bool = required_tool.is_empty() or _inventory_quantity(required_tool) > 0
		if _maximum_batches(costs) < 1 or not has_tool or smithing_level < required_level:
			continue
		var tool_name: String = _item_name(required_tool)
		if not required_tool.is_empty():
			input_lines += "\nTool: %s" % tool_name
		var button := _recipe_button(reward_item, 62)
		button.custom_minimum_size = Vector2(260, 132)
		button.text = "%s\n%s\nSmithing %d" % [
			recipe.get("name", recipe_id.capitalize()), input_lines, required_level,
		]
		button.tooltip_text = "%s %s until materials run out" % [
			"Craft" if str(recipe.get("group", "")) == "jewellery" else "Smelt",
			recipe.get("name", recipe_id.capitalize()),
		]
		button.pressed.connect(_request_default.bind("smelting", recipe_id, ""))
		button.gui_input.connect(_on_recipe_input.bind("smelting", recipe_id))
		_smelting_grid.add_child(button)


func _render_forging() -> void:
	_clear_children(_smithing_tabs)
	var recipes: Array = []
	for recipe_data: Variant in _station_recipes("forging"):
		var recipe := recipe_data as Dictionary
		var required_tool: String = str(recipe.get("required_tool", ""))
		if int(_skill_levels.get("smithing", 1)) < int(recipe.get("required_level", 1)):
			continue
		if _maximum_batches(recipe.get("costs", []) as Array) < 1:
			continue
		if not required_tool.is_empty() and _inventory_quantity(required_tool) < 1:
			continue
		recipes.append(recipe)
	for group: String in _ordered_recipe_groups(recipes):
		var grid := _add_forging_tab(group.capitalize())
		for recipe_data: Variant in recipes:
			var recipe := recipe_data as Dictionary
			if str(recipe.get("group", "")) != group:
				continue
			grid.add_child(_forging_recipe_button(recipe))
	_add_upgrade_tab()


func _forging_recipe_button(recipe: Dictionary) -> Button:
	var recipe_id: String = str(recipe.get("id", ""))
	var required_level: int = int(recipe.get("required_level", 1))
	var button := _recipe_button(str(recipe.get("reward_item", recipe_id)), 50)
	button.custom_minimum_size = Vector2(300, 78)
	button.text = "%s\n%s  •  Lv %d" % [
		recipe.get("name", recipe_id.capitalize()),
		_format_costs(recipe.get("costs", []) as Array, " + "),
		required_level,
	]
	button.disabled = int(_skill_levels.get("smithing", 1)) < required_level
	button.tooltip_text = (
		"Requires Smithing level %d" % required_level
		if button.disabled
		else "Forge %s" % recipe.get("name", recipe_id.capitalize())
	)
	button.pressed.connect(_request_default.bind("forging", recipe_id, ""))
	return button


func _add_upgrade_tab() -> void:
	var scroll := ScrollContainer.new()
	scroll.name = "Upgrade"
	scroll.horizontal_scroll_mode = ScrollContainer.SCROLL_MODE_DISABLED
	var list := VBoxContainer.new()
	list.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	list.add_theme_constant_override("separation", 10)
	scroll.add_child(list)
	_smithing_tabs.add_child(scroll)
	var candidates: int = 0
	for payload: Variant in _inventory_items:
		if typeof(payload) != TYPE_DICTIONARY:
			continue
		var item := payload as Dictionary
		var item_id: String = str(item.get("item_id", ""))
		var instance_id: String = str(item.get("instance_id", ""))
		var reinforcement: int = int(item.get("reinforcement", 0))
		var data := _reinforcement_data(item_id)
		if instance_id.is_empty() or reinforcement >= MAXIMUM_REINFORCEMENT or data.is_empty():
			continue
		var required_level: int = int(data.get("level", 1))
		var bar_id: String = str(data.get("bar", ""))
		var bar_cost: int = int(data.get("cost", 0))
		if (
			bar_id.is_empty()
			or bar_cost < 1
			or _inventory_quantity(bar_id) < bar_cost
			or int(_skill_levels.get("smithing", 1)) < required_level
		):
			continue
		var button := _recipe_button(item_id, 48)
		button.custom_minimum_size = Vector2(0, 72)
		button.text = "Upgrade %s to +%d\n%d %s  •  Lv %d  •  slot %d" % [
			item_id.capitalize(), reinforcement + 1, int(data.get("cost", 0)),
			data.get("bar_name", "Bars"), required_level, int(item.get("slot", 0)) + 1,
		]
		button.disabled = int(_skill_levels.get("smithing", 1)) < required_level
		button.pressed.connect(_request_default.bind("forging", "reinforce_%s" % item_id, instance_id))
		list.add_child(button)
		candidates += 1
	if candidates == 0:
		var empty := Label.new()
		empty.text = "No eligible equipment to upgrade."
		empty.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
		empty.custom_minimum_size = Vector2(0, 80)
		list.add_child(empty)


func _add_forging_tab(title: String) -> GridContainer:
	var scroll := ScrollContainer.new()
	scroll.name = title
	scroll.horizontal_scroll_mode = ScrollContainer.SCROLL_MODE_DISABLED
	var grid := GridContainer.new()
	grid.columns = 2
	grid.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	grid.add_theme_constant_override("h_separation", 10)
	grid.add_theme_constant_override("v_separation", 10)
	scroll.add_child(grid)
	_smithing_tabs.add_child(scroll)
	return grid


func _ordered_recipe_groups(recipes: Array) -> Array[String]:
	var present: Dictionary = {}
	for recipe_data: Variant in recipes:
		var group: String = str((recipe_data as Dictionary).get("group", ""))
		if not group.is_empty() and group != "jewellery":
			present[group] = true
	var ordered: Array[String] = []
	for group: String in GROUP_ORDER:
		if present.erase(group):
			ordered.append(group)
	var remaining: Array = present.keys()
	remaining.sort()
	for group_value: Variant in remaining:
		ordered.append(str(group_value))
	return ordered


func _recipe_button(icon_id: String, icon_width: int) -> Button:
	var button := Button.new()
	button.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	button.focus_mode = Control.FOCUS_NONE
	button.alignment = HORIZONTAL_ALIGNMENT_LEFT
	button.icon = _item_icons.get(icon_id, _empty_slot_icon) as Texture2D
	button.expand_icon = true
	button.add_theme_constant_override("icon_max_width", icon_width)
	return button


func _on_recipe_input(event: InputEvent, station: String, recipe_id: String) -> void:
	var mouse_event := event as InputEventMouseButton
	if mouse_event == null or not mouse_event.pressed or mouse_event.button_index != MOUSE_BUTTON_RIGHT:
		return
	_batch_station = station
	_batch_recipe_id = recipe_id
	_action_menu.position = Vector2i(get_viewport().get_mouse_position())
	_action_menu.popup()
	get_viewport().set_input_as_handled()


func _request_default(station: String, recipe_id: String, instance_id: String) -> void:
	_hide_station(station)
	production_requested.emit(station, recipe_id, instance_id, 0)


func _setup_batch_controls() -> void:
	_action_menu = PopupMenu.new()
	_action_menu.name = "ProductionActionMenu"
	var panel_style := StyleBoxFlat.new()
	panel_style.bg_color = Color(0.055, 0.06, 0.075, 1.0)
	_action_menu.add_theme_stylebox_override("panel", panel_style)
	_action_menu.add_item("Make 1", 0)
	_action_menu.add_item("Make X", 1)
	_action_menu.add_item("Make All", 2)
	_action_menu.id_pressed.connect(_on_batch_action_selected)
	add_child(_action_menu)

	_quantity_dialog = ConfirmationDialog.new()
	InterfaceStyle.configure_window(_quantity_dialog)
	_quantity_dialog.dialog_text = "Choose how many to make:"
	_quantity_dialog.min_size = Vector2i(360, 180)
	_quantity_dialog.get_ok_button().text = "Start"
	_quantity_input = SpinBox.new()
	_quantity_input.min_value = 1
	_quantity_input.max_value = 10_000
	_quantity_input.step = 1
	_quantity_input.value = 5
	_quantity_input.custom_minimum_size = Vector2(280, 44)
	_quantity_input.position = Vector2(40, 74)
	_quantity_dialog.add_child(_quantity_input)
	_quantity_dialog.confirmed.connect(_confirm_quantity)
	add_child(_quantity_dialog)


func _on_batch_action_selected(action_id: int) -> void:
	match action_id:
		0: _emit_batch(1)
		1: _open_quantity_dialog()
		2: _emit_batch(0)


func _open_quantity_dialog() -> void:
	match _batch_station:
		"cooking": _quantity_dialog.title = "Cook X"
		"smelting": _quantity_dialog.title = "Smelt X"
		_: _quantity_dialog.title = "Craft X"
	_quantity_input.value = 5
	_quantity_dialog.popup_centered()
	_quantity_input.grab_focus()
	_quantity_input.get_line_edit().select_all()


func _confirm_quantity() -> void:
	_emit_batch(int(_quantity_input.value))


func _emit_batch(quantity: int) -> void:
	if _batch_station.is_empty() or _batch_recipe_id.is_empty():
		return
	_hide_station(_batch_station)
	production_requested.emit(_batch_station, _batch_recipe_id, "", quantity)


func _hide_station(station: String) -> void:
	match station:
		"cooking": _cooking_window.hide()
		"smelting": _smelting_window.hide()
		"forging": _smithing_window.hide()


func _on_cooking_details_toggled(show_details: bool) -> void:
	_cooking_details_visible = show_details
	_render_cooking()


func _on_window_input(event: InputEvent, window: Window) -> void:
	if event is InputEventKey:
		var key_event := event as InputEventKey
		if key_event.pressed and not key_event.echo and key_event.keycode == KEY_ESCAPE:
			window.hide()
			window.set_input_as_handled()


func _station_recipes(station: String) -> Array:
	var value: Variant = _recipes.get(station, [])
	return value as Array if typeof(value) == TYPE_ARRAY else []


func _skill_level(progress: Dictionary, skill_id: String) -> int:
	return int((progress.get(skill_id, {}) as Dictionary).get("level", 1))


func _inventory_quantity(item_id: String) -> int:
	var total: int = 0
	for payload: Variant in _inventory_items:
		if typeof(payload) == TYPE_DICTIONARY and str((payload as Dictionary).get("item_id", "")) == item_id:
			total += int((payload as Dictionary).get("quantity", 1))
	return total


func _maximum_batches(costs: Array) -> int:
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


func _format_costs(costs: Array, separator: String) -> String:
	var labels: PackedStringArray = []
	for cost_data: Variant in costs:
		if typeof(cost_data) != TYPE_DICTIONARY:
			continue
		var cost := cost_data as Dictionary
		var item_id: String = str(cost.get("item_id", ""))
		var quantity: int = int(cost.get("quantity", 0))
		if not item_id.is_empty() and quantity > 0:
			labels.append("%d %s" % [quantity, _item_name(item_id)])
	return separator.join(labels)


func _item_name(item_id: String) -> String:
	return str(_item_names.get(item_id, item_id.replace("_", " ").capitalize()))


func _reinforcement_data(item_id: String) -> Dictionary:
	for tier_value: Variant in _reinforcement_tiers.values():
		var tier := tier_value as Dictionary
		if not item_id in (tier.get("items", []) as Array):
			continue
		var result := tier.duplicate()
		var slot: String = str(_armour_slots.get(item_id, ""))
		var cost: int = 3 if _equipment_icons.has(item_id) else 2
		if slot == "body":
			cost = 4
		elif slot in ["legs", "left_hand"]:
			cost = 3
		result["cost"] = cost
		return result
	return {}


func _clear_children(parent: Node) -> void:
	for child: Node in parent.get_children():
		parent.remove_child(child)
		child.queue_free()
