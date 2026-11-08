class_name InterfaceStyle
extends RefCounted

const CLOSE_WINDOW_ICON: Texture2D = preload("res://assets/icons/navigation/close_window.svg")


static func configure_window(window: Window) -> void:
	var interface_theme := Theme.new()
	interface_theme.set_icon("close", "Window", CLOSE_WINDOW_ICON)
	interface_theme.set_icon("close_pressed", "Window", CLOSE_WINDOW_ICON)
	window.theme = interface_theme


static func configure_close_button(button: Button) -> void:
	button.text = ""
	button.icon = CLOSE_WINDOW_ICON
	button.expand_icon = true
	button.icon_alignment = HORIZONTAL_ALIGNMENT_CENTER
	button.vertical_icon_alignment = VERTICAL_ALIGNMENT_CENTER
	button.custom_minimum_size = Vector2(44, 44)
	button.focus_mode = Control.FOCUS_NONE
	button.flat = true
	button.add_theme_constant_override("icon_max_width", 28)
	button.modulate = Color(1.0, 1.0, 1.0, 0.92)
	button.mouse_entered.connect(func() -> void: button.modulate = Color.WHITE)
	button.mouse_exited.connect(func() -> void: button.modulate = Color(1.0, 1.0, 1.0, 0.92))
	var transparent := StyleBoxEmpty.new()
	for state: String in ["normal", "hover", "pressed", "focus"]:
		button.add_theme_stylebox_override(state, transparent)
