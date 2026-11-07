class_name GatheringSpot
extends Node3D

@onready var body: StaticBody3D = $Body
@onready var label: Label3D = $Label


func set_spot(id: String, display_name: String, skill: String) -> void:
	body.set_meta(&"gathering_spot_id", id)
	label.text = "%s\n%s" % [display_name, skill.capitalize()]
