class_name MerchantWindow
extends Window

signal buy_requested(item_id: String, instance_id: String)

const CLOSE_ICON: Texture2D = preload("res://assets/icons/navigation/close_window.svg")
const EMPTY_ICON: Texture2D = preload("res://assets/icons/items/empty_slot.svg")

@onready var close_button: Button = $Panel/Margin/VBox/Header/Close
@onready var grid: GridContainer = $Panel/Margin/VBox/Scroll/Grid


func _ready() -> void:
	close_button.icon = CLOSE_ICON
	close_button.expand_icon = true
	close_button.add_theme_constant_override("icon_max_width", 28)
	close_button.pressed.connect(hide)
	close_requested.connect(hide)
	window_input.connect(_on_window_input)


func open_stock(items: Array) -> void:
	render_stock(items)
	popup_centered()


func render_stock(items: Array) -> void:
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
		button.icon = IconCatalog.centered_icon(
			IconCatalog.ITEMS.get(item_id, EMPTY_ICON) as Texture2D
		)
		button.expand_icon = true
		button.add_theme_constant_override("icon_max_width", 52)
		button.text = "%s%s%s\n%d coins" % [
			item_id.replace("_", " ").capitalize(),
			" +%d" % reinforcement if reinforcement > 0 else "",
			" ×%d" % quantity if quantity > 1 else "",
			price,
		]
		button.tooltip_text = "Buy one for %d coins • %d available" % [price, quantity]
		button.pressed.connect(buy_requested.emit.bind(item_id, instance_id))
		grid.add_child(button)


func _on_window_input(event: InputEvent) -> void:
	if event is InputEventKey:
		var key_event := event as InputEventKey
		if key_event.pressed and not key_event.echo and key_event.keycode == KEY_ESCAPE:
			hide()
			set_input_as_handled()
