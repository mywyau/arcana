class_name MerchantWindow
extends Window

signal buy_requested(item_id: String, instance_id: String, quantity: int)

const CLOSE_ICON: Texture2D = preload("res://assets/icons/navigation/close_window.svg")
const EMPTY_ICON: Texture2D = preload("res://assets/icons/items/empty_slot.svg")
const BUY_ONE := 0
const BUY_X := 1
const BUY_ALL := 2

@onready var close_button: Button = $Panel/Margin/VBox/Header/Close
@onready var grid: GridContainer = $Panel/Margin/VBox/Scroll/Grid

var _context_menu: PopupMenu
var _quantity_dialog: ConfirmationDialog
var _quantity_input: SpinBox
var _selected_item_id: String = ""
var _selected_instance_id: String = ""
var _selected_quantity: int = 0


func _ready() -> void:
	close_button.icon = CLOSE_ICON
	close_button.expand_icon = true
	close_button.add_theme_constant_override("icon_max_width", 28)
	close_button.pressed.connect(hide)
	close_requested.connect(hide)
	window_input.connect(_on_window_input)
	_context_menu = PopupMenu.new()
	_context_menu.min_size = Vector2i(190, 0)
	_context_menu.add_theme_font_size_override("font_size", 18)
	_context_menu.add_theme_constant_override("item_start_padding", 18)
	_context_menu.add_theme_constant_override("item_end_padding", 22)
	_context_menu.add_theme_constant_override("v_separation", 8)
	_context_menu.id_pressed.connect(_on_context_action)
	add_child(_context_menu)

	_quantity_dialog = ConfirmationDialog.new()
	InterfaceStyle.configure_window(_quantity_dialog)
	_quantity_dialog.title = "Buy items"
	_quantity_dialog.dialog_text = "Choose an amount to buy:"
	_quantity_dialog.min_size = Vector2i(360, 180)
	_quantity_dialog.get_ok_button().text = "Buy"
	_quantity_input = SpinBox.new()
	_quantity_input.min_value = 1
	_quantity_input.step = 1
	_quantity_input.custom_minimum_size = Vector2(280, 44)
	_quantity_input.position = Vector2(40, 74)
	_quantity_dialog.add_child(_quantity_input)
	_quantity_dialog.confirmed.connect(_confirm_buy_quantity)
	add_child(_quantity_dialog)


func open_stock(items: Array) -> void:
	render_stock(items)
	popup_centered()


func render_stock(items: Array) -> void:
	_context_menu.hide()
	for child: Node in grid.get_children():
		grid.remove_child(child)
		child.queue_free()
	if items.is_empty():
		var empty_label := Label.new()
		empty_label.text = "Nothing is available to buy back yet."
		empty_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
		empty_label.custom_minimum_size = Vector2(420, 72)
		grid.add_child(empty_label)
		return
	for payload: Variant in items:
		if typeof(payload) != TYPE_DICTIONARY:
			continue
		var item := payload as Dictionary
		var price: int = int(item.get("buy_price", 0))
		if price <= 0:
			continue
		var item_id: String = str(item.get("item_id", ""))
		var quantity: int = int(item.get("quantity", 1))
		var instance_id: String = str(item.get("instance_id", ""))
		var reinforcement: int = int(item.get("reinforcement", 0))
		var button := Button.new()
		button.custom_minimum_size = Vector2(132, 112)
		button.focus_mode = Control.FOCUS_NONE
		button.tooltip_text = "%s%s • buy one for %d coins • %d available" % [
			item_id.replace("_", " ").capitalize(),
			" +%d" % reinforcement if reinforcement > 0 else "",
			price,
			quantity,
		]
		button.pressed.connect(buy_requested.emit.bind(item_id, instance_id, 1))
		button.gui_input.connect(_on_stock_input.bind(item_id, instance_id, quantity))
		grid.add_child(button)

		var icon := TextureRect.new()
		icon.texture = IconCatalog.centered_icon(
			IconCatalog.ITEMS.get(item_id, EMPTY_ICON) as Texture2D
		)
		icon.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
		icon.offset_left = 20
		icon.offset_top = 8
		icon.offset_right = -20
		icon.offset_bottom = -28
		icon.expand_mode = TextureRect.EXPAND_IGNORE_SIZE
		icon.stretch_mode = TextureRect.STRETCH_KEEP_ASPECT_CENTERED
		icon.mouse_filter = Control.MOUSE_FILTER_IGNORE
		button.add_child(icon)

		var price_label := Label.new()
		price_label.set_anchors_and_offsets_preset(Control.PRESET_BOTTOM_WIDE)
		price_label.offset_top = -28
		price_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
		price_label.text = "%d coins" % price
		price_label.mouse_filter = Control.MOUSE_FILTER_IGNORE
		button.add_child(price_label)

		if quantity > 1:
			var quantity_label := Label.new()
			quantity_label.set_anchors_and_offsets_preset(Control.PRESET_TOP_RIGHT)
			quantity_label.offset_left = -52
			quantity_label.offset_bottom = 26
			quantity_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_RIGHT
			quantity_label.text = "×%d" % quantity
			quantity_label.mouse_filter = Control.MOUSE_FILTER_IGNORE
			button.add_child(quantity_label)
		if reinforcement > 0:
			var reinforcement_label := Label.new()
			reinforcement_label.set_anchors_and_offsets_preset(Control.PRESET_TOP_LEFT)
			reinforcement_label.offset_right = 52
			reinforcement_label.offset_bottom = 26
			reinforcement_label.text = "+%d" % reinforcement
			reinforcement_label.mouse_filter = Control.MOUSE_FILTER_IGNORE
			button.add_child(reinforcement_label)


func _on_window_input(event: InputEvent) -> void:
	if event is InputEventKey:
		var key_event := event as InputEventKey
		if key_event.pressed and not key_event.echo and key_event.keycode == KEY_ESCAPE:
			if _context_menu.visible:
				_context_menu.hide()
				set_input_as_handled()
				return
			if _quantity_dialog.visible:
				_quantity_dialog.hide()
				set_input_as_handled()
				return
			hide()
			set_input_as_handled()


func _on_stock_input(
		event: InputEvent,
		item_id: String,
		instance_id: String,
		available_quantity: int,
) -> void:
	var mouse_button := event as InputEventMouseButton
	if mouse_button == null or not mouse_button.pressed or mouse_button.button_index != MOUSE_BUTTON_RIGHT:
		return
	_selected_item_id = item_id
	_selected_instance_id = instance_id
	_selected_quantity = available_quantity
	_context_menu.clear()
	_context_menu.add_item("Buy 1", BUY_ONE)
	_context_menu.add_item("Buy X", BUY_X)
	_context_menu.add_item("Buy All", BUY_ALL)
	_context_menu.position = Vector2i(mouse_button.global_position)
	_context_menu.popup()
	get_viewport().set_input_as_handled()


func _on_context_action(action: int) -> void:
	if _selected_item_id.is_empty() or _selected_quantity <= 0:
		return
	match action:
		BUY_ONE:
			buy_requested.emit(_selected_item_id, _selected_instance_id, 1)
		BUY_X:
			_quantity_input.max_value = _selected_quantity
			_quantity_input.value = mini(_selected_quantity, 10)
			_quantity_dialog.popup_centered()
			_quantity_input.grab_focus()
			_quantity_input.get_line_edit().select_all()
		BUY_ALL:
			buy_requested.emit(_selected_item_id, _selected_instance_id, _selected_quantity)


func _confirm_buy_quantity() -> void:
	if _selected_item_id.is_empty():
		return
	buy_requested.emit(
		_selected_item_id, _selected_instance_id, int(_quantity_input.value),
	)
