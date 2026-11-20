extends Window

signal cast_requested(item: Dictionary, quantity: int, automatic: bool)

var picker := OptionButton.new()
var detail := Label.new()
var fast := Button.new()
var auto_cast := Button.new()
var stop := Button.new()
var slots: Dictionary = {}
var chosen := -1
var runes := 0
var remaining := 0
var waiting := false


func _ready() -> void:
	title = "Alchemy"
	size = Vector2i(440, 185)
	transient = true
	InterfaceStyle.configure_window(self)
	var panel := PanelContainer.new()
	panel.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	add_child(panel)
	var column := VBoxContainer.new()
	column.add_theme_constant_override("separation", 10)
	panel.add_child(column)
	column.add_child(picker)
	column.add_child(detail)
	var buttons := HBoxContainer.new()
	column.add_child(buttons)
	for button: Button in [fast, auto_cast, stop]:
		buttons.add_child(button)
	fast.text = "Cast 1"
	fast.tooltip_text = "One item per click • 1.2 seconds"
	auto_cast.tooltip_text = "One item from this slot every 3 seconds • up to 20"
	stop.text = "Stop"
	picker.item_selected.connect(_choose)
	fast.pressed.connect(_cast.bind(false))
	auto_cast.pressed.connect(_cast.bind(true))
	stop.pressed.connect(func() -> void: cast_requested.emit({}, 0, false))
	close_requested.connect(hide)
	window_input.connect(_on_input)
	hide()


func _on_input(event: InputEvent) -> void:
	if event is InputEventKey and event.pressed and event.keycode == KEY_ESCAPE:
		hide()
		set_input_as_handled()


func refresh(items: Array, protected_slots: Array, available_runes: int, queued: int, awaiting_ack: bool) -> void:
	var updated: Dictionary = {}
	for item: Dictionary in items:
		var coins := int(item.get("alchemy_coins", 0))
		if coins <= 0 or protected_slots.has(int(item.get("slot", -1))):
			continue
		updated[int(item.slot)] = item
	if updated != slots or picker.item_count == 0:
		if chosen >= 0 and slots.has(chosen):
			var previous: Dictionary = slots[chosen]
			var next: Dictionary = updated.get(chosen, {})
			if previous.get("item_id") != next.get("item_id") or previous.get("instance_id") != next.get("instance_id"):
				chosen = -1
		slots = updated
		picker.clear()
		picker.add_item("Choose an inventory slot" if not slots.is_empty() else "No eligible items in inventory")
		picker.select(0)
		for slot: int in slots:
			var item: Dictionary = slots[slot]
			picker.add_item("Slot %d • %s ×%d" % [slot + 1, IconCatalog.item_display_name(item.item_id), item.quantity])
			var index := picker.item_count - 1
			picker.set_item_metadata(index, slot)
			if slot == chosen:
				picker.select(index)
	runes = available_runes
	remaining = queued
	waiting = awaiting_ack
	_refresh_actions()


func _choose(index: int) -> void:
	chosen = int(picker.get_item_metadata(index)) if index > 0 else -1
	_refresh_actions()


func select_slot(slot: int) -> void:
	for index in range(1, picker.item_count):
		if int(picker.get_item_metadata(index)) == slot:
			picker.select(index)
			_choose(index)
			return


func _refresh_actions() -> void:
	var entry: Dictionary = slots.get(chosen, {})
	var quantity := int(entry.get("quantity", 0))
	var count := mini(20, mini(quantity, runes))
	var busy := remaining > 0 or waiting
	picker.disabled = busy or slots.is_empty()
	fast.disabled = busy or count < 1
	auto_cast.disabled = fast.disabled
	auto_cast.text = "Auto ×%d" % count
	stop.disabled = remaining <= 0 or waiting
	if remaining > 0:
		detail.text = "%d cast%s remaining" % [remaining, "" if remaining == 1 else "s"]
	elif not entry.is_empty():
		detail.text = "%d coins each • 1 Focus Rune per item (%d held)" % [entry.alchemy_coins, runes]
	else:
		detail.text = "1 Focus Rune per item • %d held" % runes


func _cast(automatic: bool) -> void:
	if fast.disabled:
		return
	var item: Dictionary = slots[chosen]
	var quantity := int(item.quantity)
	var count := mini(20, mini(quantity, runes)) if automatic else 1
	cast_requested.emit(item, count, automatic)
