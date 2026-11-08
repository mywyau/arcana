class_name MerchantWindow
extends Window

signal sell_requested(inventory_slot: int)

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


func show_inventory(items: Array) -> void:
	for child: Node in grid.get_children():
		grid.remove_child(child)
		child.queue_free()
	for payload: Variant in items:
		if typeof(payload) != TYPE_DICTIONARY:
			continue
		var item := payload as Dictionary
		var price: int = int(item.get("sell_price", 0))
		if price <= 0:
			continue
		var item_id: String = str(item.get("item_id", ""))
		var quantity: int = int(item.get("quantity", 1))
		var slot: int = int(item.get("slot", -1))
		var button := Button.new()
		button.custom_minimum_size = Vector2(132, 112)
		button.icon = IconCatalog.ITEMS.get(item_id, EMPTY_ICON) as Texture2D
		button.expand_icon = true
		button.add_theme_constant_override("icon_max_width", 52)
		button.text = "%s%s\n%d coins" % [
			item_id.replace("_", " ").capitalize(),
			" ×%d" % quantity if quantity > 1 else "",
			price,
		]
		button.tooltip_text = "Sell the whole stack for %d coins" % price
		button.pressed.connect(sell_requested.emit.bind(slot))
		grid.add_child(button)
	popup_centered()


func _on_window_input(event: InputEvent) -> void:
	if event is InputEventKey:
		var key_event := event as InputEventKey
		if key_event.pressed and not key_event.echo and key_event.keycode == KEY_ESCAPE:
			hide()
			set_input_as_handled()
