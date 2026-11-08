class_name SpellSlot
extends Button

var _spell_name: String = "Spell"
var _required_level: int = 1
var _costs: Array[Dictionary] = []
var _placeholder: bool = false


func _ready() -> void:
	# Godot wraps custom tooltips in its own TooltipPanel. Removing that
	# background leaves the rounded card below as the only visible container.
	var tooltip_theme := Theme.new()
	tooltip_theme.set_stylebox("panel", "TooltipPanel", StyleBoxEmpty.new())
	theme = tooltip_theme


func configure_spell_tooltip(
	spell_name: String,
	required_level: int,
	costs: Array[Dictionary],
	placeholder: bool,
) -> void:
	_spell_name = spell_name
	_required_level = required_level
	_costs = costs.duplicate(true)
	_placeholder = placeholder
	# Godot only requests a custom tooltip when tooltip_text is non-empty.
	tooltip_text = spell_name


func _make_custom_tooltip(_for_text: String) -> Object:
	var panel := PanelContainer.new()
	panel.custom_minimum_size = Vector2(320.0, 0.0)
	var panel_style := StyleBoxFlat.new()
	panel_style.bg_color = Color(0.045, 0.052, 0.065, 0.98)
	panel_style.border_width_left = 1
	panel_style.border_width_top = 1
	panel_style.border_width_right = 1
	panel_style.border_width_bottom = 1
	panel_style.border_color = Color(0.26, 0.3, 0.36, 1.0)
	panel_style.corner_radius_top_left = 7
	panel_style.corner_radius_top_right = 7
	panel_style.corner_radius_bottom_left = 7
	panel_style.corner_radius_bottom_right = 7
	panel.add_theme_stylebox_override("panel", panel_style)

	var margin := MarginContainer.new()
	margin.add_theme_constant_override("margin_left", 12)
	margin.add_theme_constant_override("margin_top", 10)
	margin.add_theme_constant_override("margin_right", 12)
	margin.add_theme_constant_override("margin_bottom", 10)
	panel.add_child(margin)

	var content := VBoxContainer.new()
	content.add_theme_constant_override("separation", 6)
	margin.add_child(content)

	var header := HBoxContainer.new()
	header.add_theme_constant_override("separation", 10)
	content.add_child(header)
	var title := Label.new()
	title.text = _spell_name
	title.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	title.add_theme_font_size_override("font_size", 19)
	title.add_theme_color_override("font_color", Color(0.93, 0.78, 0.35))
	header.add_child(title)
	if not _placeholder:
		var level_label := Label.new()
		level_label.text = "Lv %d" % _required_level
		level_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_RIGHT
		level_label.vertical_alignment = VERTICAL_ALIGNMENT_CENTER
		level_label.add_theme_color_override("font_color", Color(0.72, 0.78, 0.86))
		header.add_child(level_label)

	if _placeholder:
		var coming_later := Label.new()
		coming_later.text = "Coming later"
		coming_later.add_theme_color_override("font_color", Color(0.68, 0.72, 0.78))
		content.add_child(coming_later)
		return panel

	var cost_row := HBoxContainer.new()
	cost_row.add_theme_constant_override("separation", 16)
	cost_row.alignment = BoxContainer.ALIGNMENT_CENTER
	content.add_child(cost_row)
	var missing_names := PackedStringArray()
	for cost: Dictionary in _costs:
		var rune_entry := PanelContainer.new()
		rune_entry.custom_minimum_size = Vector2(118.0, 64.0)
		var rune_style := StyleBoxFlat.new()
		rune_style.bg_color = Color(0.075, 0.09, 0.11, 1.0)
		rune_style.border_width_left = 2
		rune_style.border_width_top = 2
		rune_style.border_width_right = 2
		rune_style.border_width_bottom = 2
		rune_style.border_color = (
			Color(0.85, 0.2, 0.18, 1.0)
			if bool(cost.get("missing", false))
			else Color(0.34, 0.55, 0.68, 1.0)
		)
		rune_style.corner_radius_top_left = 6
		rune_style.corner_radius_top_right = 6
		rune_style.corner_radius_bottom_left = 6
		rune_style.corner_radius_bottom_right = 6
		rune_entry.add_theme_stylebox_override("panel", rune_style)
		var rune_margin := MarginContainer.new()
		rune_margin.add_theme_constant_override("margin_left", 8)
		rune_margin.add_theme_constant_override("margin_top", 6)
		rune_margin.add_theme_constant_override("margin_right", 8)
		rune_margin.add_theme_constant_override("margin_bottom", 6)
		rune_entry.add_child(rune_margin)
		var rune_contents := HBoxContainer.new()
		rune_contents.add_theme_constant_override("separation", 7)
		rune_margin.add_child(rune_contents)
		var rune_icon := TextureRect.new()
		rune_icon.custom_minimum_size = Vector2(46.0, 46.0)
		rune_icon.texture = cost.get("icon") as Texture2D
		rune_icon.expand_mode = TextureRect.EXPAND_IGNORE_SIZE
		rune_icon.stretch_mode = TextureRect.STRETCH_KEEP_ASPECT_CENTERED
		rune_icon.tooltip_text = str(cost.get("name", "Rune"))
		rune_contents.add_child(rune_icon)
		var quantity := Label.new()
		quantity.text = "×%d" % int(cost.get("quantity", 0))
		quantity.vertical_alignment = VERTICAL_ALIGNMENT_CENTER
		quantity.add_theme_font_size_override("font_size", 20)
		if bool(cost.get("missing", false)):
			quantity.add_theme_color_override("font_color", Color(1.0, 0.42, 0.38))
			missing_names.append(str(cost.get("name", "Rune")))
		else:
			quantity.add_theme_color_override("font_color", Color(0.92, 0.94, 0.98))
		rune_contents.add_child(quantity)
		cost_row.add_child(rune_entry)

	if not missing_names.is_empty():
		var missing_label := Label.new()
		missing_label.text = "Missing: %s" % " + ".join(missing_names)
		missing_label.add_theme_color_override("font_color", Color(1.0, 0.42, 0.38))
		content.add_child(missing_label)

	return panel
