class_name DebugSkillWindow
extends Window

signal set_level_requested(skill_id: String, level: int)
signal set_experience_requested(skill_id: String, experience: int)

const WINDOW_SIZE := Vector2i(460, 300)

var _skill_progress: Dictionary = {}
var _skill_selector: OptionButton
var _skill_level: SpinBox
var _skill_experience: SpinBox


func _init() -> void:
	name = "DebugSkillWindow"
	title = ""
	size = WINDOW_SIZE
	unresizable = true
	transient = true


func _ready() -> void:
	InterfaceStyle.configure_window(self)
	_build_content()
	close_requested.connect(hide)
	window_input.connect(_on_window_input)
	hide()


func show_for_skills(skill_progress: Dictionary) -> void:
	_skill_progress = skill_progress
	if _skill_selector.item_count == 0:
		for skill_id: String in _skill_progress.keys():
			_add_skill(skill_id)
	_refresh_inputs()
	popup_centered(WINDOW_SIZE)


func _build_content() -> void:
	var panel := PanelContainer.new()
	panel.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	add_child(panel)

	var margin := MarginContainer.new()
	margin.add_theme_constant_override("margin_left", 20)
	margin.add_theme_constant_override("margin_top", 14)
	margin.add_theme_constant_override("margin_right", 20)
	margin.add_theme_constant_override("margin_bottom", 18)
	panel.add_child(margin)

	var content := VBoxContainer.new()
	content.add_theme_constant_override("separation", 12)
	margin.add_child(content)
	content.add_child(_build_header())

	_skill_selector = OptionButton.new()
	_skill_selector.item_selected.connect(_refresh_inputs.unbind(1))
	content.add_child(_skill_selector)
	content.add_child(_build_level_row())
	content.add_child(_build_experience_row())

	var warning := Label.new()
	warning.text = "Debug only — changes are saved to this character."
	warning.modulate = Color(0.95, 0.72, 0.38)
	content.add_child(warning)


func _build_header() -> HBoxContainer:
	var header := HBoxContainer.new()
	var heading := Label.new()
	heading.text = "Test Skill Shrine"
	heading.add_theme_font_size_override("font_size", 22)
	heading.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	header.add_child(heading)
	var close := Button.new()
	InterfaceStyle.configure_close_button(close)
	close.pressed.connect(hide)
	header.add_child(close)
	return header


func _build_level_row() -> HBoxContainer:
	var row := HBoxContainer.new()
	row.add_child(_field_label("Level"))
	_skill_level = SpinBox.new()
	_skill_level.min_value = 1
	_skill_level.max_value = 99
	_skill_level.rounded = true
	_skill_level.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	row.add_child(_skill_level)
	var apply := Button.new()
	apply.text = "Set level"
	apply.pressed.connect(_emit_level_request)
	row.add_child(apply)
	return row


func _build_experience_row() -> HBoxContainer:
	var row := HBoxContainer.new()
	row.add_child(_field_label("Exact XP"))
	_skill_experience = SpinBox.new()
	_skill_experience.min_value = 0
	_skill_experience.max_value = 10_000_000
	_skill_experience.rounded = true
	_skill_experience.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	row.add_child(_skill_experience)
	var apply := Button.new()
	apply.text = "Set XP"
	apply.pressed.connect(_emit_experience_request)
	row.add_child(apply)
	return row


func _field_label(text: String) -> Label:
	var field_label := Label.new()
	field_label.text = text
	field_label.custom_minimum_size.x = 90
	return field_label


func _add_skill(skill_id: String) -> void:
	_skill_selector.add_item(skill_id.capitalize())
	_skill_selector.set_item_metadata(_skill_selector.item_count - 1, skill_id)


func _selected_skill_id() -> String:
	if _skill_selector == null or _skill_selector.item_count == 0:
		return ""
	return str(_skill_selector.get_item_metadata(_skill_selector.selected))


func _refresh_inputs() -> void:
	if _skill_selector == null or _skill_selector.item_count == 0:
		return
	var progress: Dictionary = _skill_progress.get(_selected_skill_id(), {}) as Dictionary
	_skill_level.value = int(progress.get("level", 1))
	_skill_experience.value = int(progress.get("experience", 0))


func _emit_level_request() -> void:
	set_level_requested.emit(_selected_skill_id(), int(_skill_level.value))


func _emit_experience_request() -> void:
	set_experience_requested.emit(_selected_skill_id(), int(_skill_experience.value))


func _on_window_input(event: InputEvent) -> void:
	if event is InputEventKey:
		var key_event := event as InputEventKey
		if key_event.pressed and not key_event.echo and key_event.keycode == KEY_ESCAPE:
			hide()
			set_input_as_handled()
