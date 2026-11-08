class_name BankWithdrawControls
extends Node

signal withdrawal_requested(bank_slot: int, quantity: int)

const TAKE_ONE := 0
const TAKE_X := 1
const TAKE_ALL := 2

var _menu: PopupMenu
var _dialog: ConfirmationDialog
var _amount: SpinBox
var _slot: int = -1
var _available: int = 0


func _ready() -> void:
	_menu = PopupMenu.new()
	_menu.name = "BankItemMenu"
	_menu.min_size = Vector2i(190, 0)
	_menu.add_theme_font_size_override("font_size", 18)
	_menu.add_theme_constant_override("item_start_padding", 18)
	_menu.add_theme_constant_override("item_end_padding", 22)
	_menu.add_theme_constant_override("v_separation", 8)
	_menu.id_pressed.connect(_on_menu_action)
	add_child(_menu)

	_dialog = ConfirmationDialog.new()
	InterfaceStyle.configure_window(_dialog)
	_dialog.title = "Withdraw items"
	_dialog.dialog_text = "Choose an amount to withdraw:"
	_dialog.min_size = Vector2i(360, 180)
	_dialog.get_ok_button().text = "Withdraw"
	_amount = SpinBox.new()
	_amount.min_value = 1
	_amount.step = 1
	_amount.custom_minimum_size = Vector2(280, 44)
	_amount.position = Vector2(40, 74)
	_dialog.add_child(_amount)
	_dialog.confirmed.connect(_confirm_amount)
	add_child(_dialog)


func open_context_menu(event: InputEvent, bank_slot: int, available_quantity: int) -> void:
	var mouse_button := event as InputEventMouseButton
	if mouse_button == null or not mouse_button.pressed or mouse_button.button_index != MOUSE_BUTTON_RIGHT:
		return
	_slot = bank_slot
	_available = available_quantity
	_menu.clear()
	_menu.add_item("Take 1", TAKE_ONE)
	_menu.add_item("Take X", TAKE_X)
	_menu.add_item("Take All", TAKE_ALL)
	_menu.position = Vector2i(mouse_button.global_position)
	_menu.popup()
	get_viewport().set_input_as_handled()


func close_overlays() -> bool:
	if _menu.visible:
		_menu.hide()
		return true
	if _dialog.visible:
		_dialog.hide()
		return true
	return false


func dismiss() -> void:
	_menu.hide()
	_dialog.hide()
	_slot = -1
	_available = 0


func _on_menu_action(action: int) -> void:
	if _slot < 0 or _available < 1:
		return
	match action:
		TAKE_ONE: withdrawal_requested.emit(_slot, 1)
		TAKE_X: _open_amount_dialog()
		TAKE_ALL: withdrawal_requested.emit(_slot, _available)


func _open_amount_dialog() -> void:
	_amount.max_value = _available
	_amount.value = mini(_available, 10)
	_dialog.popup_centered()
	_amount.grab_focus()
	_amount.get_line_edit().select_all()


func _confirm_amount() -> void:
	if _slot < 0:
		return
	withdrawal_requested.emit(_slot, int(_amount.value))
