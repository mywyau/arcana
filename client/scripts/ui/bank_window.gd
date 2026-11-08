class_name BankWindow
extends Node

signal withdrawal_requested(bank_slot: int, quantity: int)
signal item_context_requested(event: InputEvent, bank_slot: int, quantity: int)
signal expansion_requested
signal supply_requested(drop_id: String)

const INVENTORY_SLOT_SCRIPT: Script = preload("res://scripts/inventory_slot.gd")

var _hud: PanelContainer
var _title: Label
var _hint: Label
var _summary: Label
var _expand_button: Button
var _grid: GridContainer
var _item_icons: Dictionary
var _empty_slot_icon: Texture2D
var _test_supply_signature: String = ""


func configure(
		hud: PanelContainer,
		title: Label,
		hint: Label,
		summary: Label,
		expand_button: Button,
		grid: GridContainer,
		item_icons: Dictionary,
		empty_slot_icon: Texture2D,
) -> void:
	_hud = hud
	_title = title
	_hint = hint
	_summary = summary
	_expand_button = expand_button
	_grid = grid
	_item_icons = item_icons
	_empty_slot_icon = empty_slot_icon
	if not _expand_button.pressed.is_connected(_on_expansion_pressed):
		_expand_button.pressed.connect(_on_expansion_pressed)


func show_personal(
		items: Array,
		capacity: int,
		next_expansion_cost: int,
		inventory_coin_quantity: int,
) -> void:
	_title.text = "Personal Bank"
	_hint.text = "Click an inventory item to deposit one • click a bank item to withdraw one"
	_hud.visible = true
	_expand_button.visible = true
	_test_supply_signature = ""
	render_personal(items, capacity, next_expansion_cost, inventory_coin_quantity)


func show_test_supply(chest_name: String, supplies: Array[Dictionary], force: bool = false) -> void:
	_title.text = chest_name
	_hint.text = "Click a supply to take its stack • supplies return when the server restarts"
	_hud.visible = true
	_expand_button.visible = false
	var signature := "%s:%s" % [chest_name, JSON.stringify(supplies)]
	if not force and signature == _test_supply_signature:
		return
	_test_supply_signature = signature
	_clear_grid()
	_summary.text = "%d test suppl%s remaining" % [
		supplies.size(),
		"y" if supplies.size() == 1 else "ies",
	]
	for supply: Dictionary in supplies:
		_grid.add_child(_create_supply_slot(supply))


func render_personal(
		items: Array,
		capacity: int,
		next_expansion_cost: int,
		inventory_coin_quantity: int,
) -> void:
	if _grid == null:
		return
	_clear_grid()
	var items_by_slot: Dictionary = {}
	var bank_coin_quantity: int = 0
	for payload: Variant in items:
		if typeof(payload) != TYPE_DICTIONARY:
			continue
		var item: Dictionary = payload
		var item_id: String = str(item.get("item_id", ""))
		if item_id == "coins":
			bank_coin_quantity += int(item.get("quantity", 0))
		var slot: int = int(item.get("slot", -1))
		if slot >= 0 and slot < capacity:
			items_by_slot[slot] = item
	_summary.text = "%d/%d slots used" % [items_by_slot.size(), capacity]
	_update_expansion_button(
		next_expansion_cost,
		inventory_coin_quantity + bank_coin_quantity,
	)
	for slot: int in range(capacity):
		_grid.add_child(_create_slot(slot, items_by_slot.get(slot, {}) as Dictionary))


func _clear_grid() -> void:
	for child: Node in _grid.get_children():
		_grid.remove_child(child)
		child.queue_free()


func _update_expansion_button(next_expansion_cost: int, available_coins: int) -> void:
	if next_expansion_cost <= 0:
		_expand_button.text = "Maximum bank capacity reached"
		_expand_button.disabled = true
		return
	_expand_button.text = "Buy +50 bank slots • %d coins" % next_expansion_cost
	_expand_button.disabled = available_coins < next_expansion_cost


func _create_slot(slot: int, item: Dictionary) -> Button:
	var slot_button: Variant = INVENTORY_SLOT_SCRIPT.new()
	slot_button.custom_minimum_size = Vector2(112, 86)
	slot_button.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	slot_button.focus_mode = Control.FOCUS_NONE
	slot_button.text = ""
	if item.is_empty():
		slot_button.configure(slot, false, true)
		slot_button.icon = _empty_slot_icon
		slot_button.expand_icon = true
		slot_button.disabled = true
		slot_button.modulate = Color(0.6, 0.64, 0.68)
		return slot_button
	var item_id: String = str(item.get("item_id", ""))
	var quantity: int = int(item.get("quantity", 1))
	var reinforcement: int = int(item.get("reinforcement", 0))
	slot_button.configure(slot, true, true)
	slot_button.icon = _item_icons.get(item_id, _empty_slot_icon) as Texture2D
	slot_button.expand_icon = true
	slot_button.set_stack_quantity(quantity)
	slot_button.set_reinforcement(reinforcement)
	slot_button.tooltip_text = "%s%s ×%d • left-click: take 1 • right-click: options" % [
		item_id.capitalize(),
		" +%d" % reinforcement if reinforcement > 0 else "",
		quantity,
	]
	slot_button.pressed.connect(_on_slot_pressed.bind(slot))
	slot_button.gui_input.connect(_on_slot_input.bind(slot, quantity))
	return slot_button


func _create_supply_slot(supply: Dictionary) -> Button:
	var drop_id: String = str(supply.get("drop_id", ""))
	var item_id: String = str(supply.get("item_id", ""))
	var quantity: int = int(supply.get("quantity", 1))
	var slot_button: Variant = INVENTORY_SLOT_SCRIPT.new()
	slot_button.custom_minimum_size = Vector2(112, 86)
	slot_button.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	slot_button.focus_mode = Control.FOCUS_NONE
	slot_button.text = ""
	slot_button.configure(0, true, true)
	slot_button.icon = _item_icons.get(item_id, _empty_slot_icon) as Texture2D
	slot_button.expand_icon = true
	slot_button.set_stack_quantity(quantity)
	slot_button.tooltip_text = "%s ×%d • click to take the stack" % [item_id.capitalize(), quantity]
	slot_button.pressed.connect(_on_supply_pressed.bind(drop_id))
	return slot_button


func _on_slot_pressed(bank_slot: int) -> void:
	withdrawal_requested.emit(bank_slot, 1)


func _on_slot_input(event: InputEvent, bank_slot: int, quantity: int) -> void:
	item_context_requested.emit(event, bank_slot, quantity)


func _on_expansion_pressed() -> void:
	expansion_requested.emit()


func _on_supply_pressed(drop_id: String) -> void:
	supply_requested.emit(drop_id)
