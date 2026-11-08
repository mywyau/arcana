class_name MerchantView
extends Node3D

@onready var body: StaticBody3D = $Body
@onready var label: Label3D = $Label


func configure(id: String, display_name: String, interaction_range: float) -> void:
	body.set_meta(&"merchant_id", id)
	body.set_meta(&"merchant_range", interaction_range)
	label.text = display_name
