class_name InterfaceStyle
extends RefCounted

const CLOSE_WINDOW_ICON: Texture2D = preload("res://assets/icons/navigation/close_window.svg")

const PANEL_COLOR := Color(0.78, 0.80, 0.92, 0.99)
const SURFACE_COLOR := Color(0.86, 0.87, 0.96, 1.0)
const SURFACE_HOVER_COLOR := Color(0.91, 0.90, 0.98, 1.0)
const SURFACE_PRESSED_COLOR := Color(0.69, 0.70, 0.88, 1.0)
const BORDER_COLOR := Color(0.48, 0.46, 0.72, 1.0)
const TEXT_COLOR := Color(0.13, 0.14, 0.27, 1.0)
const MUTED_TEXT_COLOR := Color(0.29, 0.31, 0.47, 1.0)
const DISABLED_TEXT_COLOR := Color(0.43, 0.44, 0.56, 0.75)


static func configure_window(window: Window) -> void:
	window.theme = create_theme()


static func configure_popup_menu(popup: PopupMenu) -> void:
	popup.theme = create_theme()


static func configure_panel(panel: PanelContainer) -> void:
	panel.theme = create_theme()
	panel.add_theme_stylebox_override("panel", make_panel_style())


static func create_theme() -> Theme:
	var interface_theme := Theme.new()
	interface_theme.set_icon("close", "Window", CLOSE_WINDOW_ICON)
	interface_theme.set_icon("close_pressed", "Window", CLOSE_WINDOW_ICON)
	interface_theme.set_stylebox("embedded_border", "Window", make_panel_style())
	interface_theme.set_stylebox("panel", "Panel", make_panel_style())
	interface_theme.set_stylebox("panel", "PanelContainer", make_panel_style())
	interface_theme.set_stylebox("panel", "PopupMenu", make_panel_style())

	for type_name: String in [
		"Label", "Button", "OptionButton", "CheckButton", "CheckBox", "LineEdit",
		"SpinBox", "PopupMenu", "TabBar", "RichTextLabel",
	]:
		interface_theme.set_color("font_color", type_name, TEXT_COLOR)
		interface_theme.set_color("font_hover_color", type_name, TEXT_COLOR)
		interface_theme.set_color("font_pressed_color", type_name, TEXT_COLOR)
		interface_theme.set_color("font_focus_color", type_name, TEXT_COLOR)
		interface_theme.set_color("font_unselected_color", type_name, MUTED_TEXT_COLOR)
		interface_theme.set_color("font_selected_color", type_name, TEXT_COLOR)
		interface_theme.set_color("font_disabled_color", type_name, DISABLED_TEXT_COLOR)
	interface_theme.set_color("default_color", "RichTextLabel", TEXT_COLOR)

	var normal := make_surface_style(SURFACE_COLOR)
	var hover := make_surface_style(SURFACE_HOVER_COLOR)
	var pressed := make_surface_style(SURFACE_PRESSED_COLOR)
	var disabled := make_surface_style(Color(0.73, 0.75, 0.84, 0.72))
	for type_name: String in ["Button", "OptionButton"]:
		interface_theme.set_stylebox("normal", type_name, normal)
		interface_theme.set_stylebox("hover", type_name, hover)
		interface_theme.set_stylebox("pressed", type_name, pressed)
		interface_theme.set_stylebox("hover_pressed", type_name, pressed)
		interface_theme.set_stylebox("focus", type_name, hover)
		interface_theme.set_stylebox("disabled", type_name, disabled)
	interface_theme.set_stylebox("normal", "LineEdit", normal)
	interface_theme.set_stylebox("focus", "LineEdit", hover)
	interface_theme.set_stylebox("read_only", "LineEdit", disabled)
	interface_theme.set_stylebox("hover", "PopupMenu", hover)
	interface_theme.set_stylebox("tab_unselected", "TabBar", normal)
	interface_theme.set_stylebox("tab_hovered", "TabBar", hover)
	interface_theme.set_stylebox("tab_selected", "TabBar", pressed)
	return interface_theme


static func make_panel_style() -> StyleBoxFlat:
	var style := StyleBoxFlat.new()
	style.bg_color = PANEL_COLOR
	style.border_color = BORDER_COLOR
	style.set_border_width_all(2)
	style.set_corner_radius_all(10)
	return style


static func make_surface_style(color: Color) -> StyleBoxFlat:
	var style := StyleBoxFlat.new()
	style.bg_color = color
	style.border_color = Color(BORDER_COLOR, 0.72)
	style.set_border_width_all(1)
	style.set_corner_radius_all(7)
	style.content_margin_left = 10
	style.content_margin_top = 7
	style.content_margin_right = 10
	style.content_margin_bottom = 7
	return style


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
