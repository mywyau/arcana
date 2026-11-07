class_name GatheringSpot
extends Node3D

@onready var body: StaticBody3D = $Body
@onready var label: Label3D = $Label
@onready var pond_visual: Node3D = $PondVisual
@onready var range_visual: Node3D = $RangeVisual


func set_spot(id: String, display_name: String, skill: String) -> void:
	body.set_meta(&"gathering_spot_id", id)
	label.text = "%s\n%s" % [display_name, skill.capitalize()]
	var is_cooking_range: bool = skill == "cooking"
	pond_visual.visible = not is_cooking_range
	range_visual.visible = is_cooking_range
