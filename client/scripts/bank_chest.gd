class_name BankChest
extends Node3D

@onready var body: StaticBody3D = $Body
@onready var label: Label3D = $Label


func set_chest(id: String, display_name: String, interaction_range: float) -> void:
	body.set_meta(&"bank_chest_id", id)
	body.set_meta(&"bank_range", interaction_range)
	label.text = display_name
	if id.begins_with("test_"):
		label.position.y = 1.72
		label.font_size = 40
		label.outline_size = 12
		label.modulate = Color(1.0, 0.9, 0.58)
