class_name InventoryContextMenu
extends Node

signal action_requested(inventory_slot: int, action_kind: String)
signal sell_requested(inventory_slot: int, quantity: int)

const ACTION_PRIMARY := "primary"
const ACTION_DROP := "drop"
const ACTION_SELL_ONE := "sell_one"
const ACTION_SELL_X := "sell_x"
const ACTION_SELL_ALL := "sell_all"

var _popup: PopupPanel
var _rows: VBoxContainer
var _inventory_slot: int = -1
var _available_quantity: int = 0
var _quantity_dialog: ConfirmationDialog
var _quantity_input: SpinBox


func _ready() -> void:
	_popup = PopupPanel.new()
	_popup.name = "InventoryContextMenu"
	var panel_style := StyleBoxFlat.new()
	panel_style.bg_color = Color(0.055, 0.06, 0.075, 1.0)
	_popup.add_theme_stylebox_override("panel", panel_style)
	add_child(_popup)
	var margin := MarginContainer.new()
	margin.add_theme_constant_override("margin_left", 6)
	margin.add_theme_constant_override("margin_top", 6)
	margin.add_theme_constant_override("margin_right", 6)
	margin.add_theme_constant_override("margin_bottom", 6)
	_popup.add_child(margin)
	_rows = VBoxContainer.new()
	_rows.add_theme_constant_override("separation", 6)
	margin.add_child(_rows)

	_quantity_dialog = ConfirmationDialog.new()
	InterfaceStyle.configure_window(_quantity_dialog)
	_quantity_dialog.title = "Sell items"
	_quantity_dialog.dialog_text = "Choose an amount to sell:"
	_quantity_dialog.min_size = Vector2i(360, 180)
	_quantity_dialog.get_ok_button().text = "Sell"
	_quantity_input = SpinBox.new()
	_quantity_input.min_value = 1
	_quantity_input.step = 1
	_quantity_input.custom_minimum_size = Vector2(280, 44)
	_quantity_input.position = Vector2(40, 74)
	_quantity_dialog.add_child(_quantity_input)
	_quantity_dialog.confirmed.connect(_confirm_sell_quantity)
	add_child(_quantity_dialog)


func open_menu(
		inventory_slot: int,
		screen_position: Vector2,
		item_name: String,
		quantity: int,
		primary_action: String,
		consumable_heals: bool,
		sell_price: int = 0,
) -> void:
	_inventory_slot = inventory_slot
	_available_quantity = quantity
	_clear_rows()
	if sell_price > 0:
		_add_action("Sell 1", item_name, ACTION_SELL_ONE)
		_add_action("Sell X", item_name, ACTION_SELL_X)
		_add_action("Sell All", "%s • %d coins" % [item_name, sell_price], ACTION_SELL_ALL)
	elif not primary_action.is_empty():
		_add_action(
			_primary_verb(primary_action, consumable_heals), item_name, ACTION_PRIMARY,
		)
	if sell_price <= 0:
		_add_action("Drop All" if quantity > 1 else "Drop", item_name, ACTION_DROP)
	var menu_height: int = 24 + maxi(0, _rows.get_child_count() - 1) * 6
	menu_height += _rows.get_child_count() * 48
	_popup.popup(Rect2i(Vector2i(screen_position), Vector2i(280, menu_height)))


func close_overlay() -> bool:
	if not _popup.visible:
		if not _quantity_dialog.visible:
			return false
		_quantity_dialog.hide()
		return true
	_popup.hide()
	return true


func _primary_verb(primary_action: String, consumable_heals: bool) -> String:
	match primary_action:
		"equip", "ring": return "Equip"
		"ammunition": return "Load"
		"consume": return "Eat" if consumable_heals else "Drink"
		_: return "Use"


func _clear_rows() -> void:
	for child: Node in _rows.get_children():
		_rows.remove_child(child)
		child.queue_free()


func _add_action(verb: String, item_name: String, action_kind: String) -> void:
	var row := Button.new()
	row.custom_minimum_size = Vector2(260, 44)
	row.focus_mode = Control.FOCUS_NONE
	row.text = ""
	var normal_style := StyleBoxFlat.new()
	normal_style.bg_color = Color(0, 0, 0, 0)
	normal_style.content_margin_left = 16
	normal_style.content_margin_top = 8
	normal_style.content_margin_right = 16
	normal_style.content_margin_bottom = 8
	row.add_theme_stylebox_override("normal", normal_style)
	row.add_theme_stylebox_override("focus", normal_style)
	var hover_style := normal_style.duplicate() as StyleBoxFlat
	hover_style.bg_color = Color(1.0, 1.0, 1.0, 0.08)
	row.add_theme_stylebox_override("hover", hover_style)
	var pressed_style := normal_style.duplicate() as StyleBoxFlat
	pressed_style.bg_color = Color(1.0, 0.78, 0.22, 0.12)
	row.add_theme_stylebox_override("pressed", pressed_style)
	row.pressed.connect(_on_action_pressed.bind(action_kind))
	var labels := HBoxContainer.new()
	labels.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	labels.offset_left = 16
	labels.offset_top = 8
	labels.offset_right = -16
	labels.offset_bottom = -8
	labels.mouse_filter = Control.MOUSE_FILTER_IGNORE
	labels.add_theme_constant_override("separation", 5)
	row.add_child(labels)
	var verb_label := Label.new()
	verb_label.text = verb
	verb_label.add_theme_color_override("font_color", Color(0.9, 0.68, 0.16, 1.0))
	verb_label.mouse_filter = Control.MOUSE_FILTER_IGNORE
	labels.add_child(verb_label)
	var target_label := Label.new()
	target_label.text = item_name
	target_label.mouse_filter = Control.MOUSE_FILTER_IGNORE
	target_label.text_overrun_behavior = TextServer.OVERRUN_TRIM_ELLIPSIS
	target_label.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	labels.add_child(target_label)
	_rows.add_child(row)


func _on_action_pressed(action_kind: String) -> void:
	_popup.hide()
	match action_kind:
		ACTION_SELL_ONE:
			sell_requested.emit(_inventory_slot, 1)
			return
		ACTION_SELL_X:
			_open_sell_quantity_dialog()
			return
		ACTION_SELL_ALL:
			sell_requested.emit(_inventory_slot, _available_quantity)
			return
	action_requested.emit(_inventory_slot, action_kind)


func _open_sell_quantity_dialog() -> void:
	if _inventory_slot < 0 or _available_quantity <= 0:
		return
	_quantity_input.max_value = _available_quantity
	_quantity_input.value = mini(_available_quantity, 10)
	_quantity_dialog.popup_centered()
	_quantity_input.grab_focus()
	_quantity_input.get_line_edit().select_all()


func _confirm_sell_quantity() -> void:
	if _inventory_slot < 0:
		return
	sell_requested.emit(_inventory_slot, int(_quantity_input.value))
