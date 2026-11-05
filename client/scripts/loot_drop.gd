class_name LootDrop
extends Node3D

@onready var visual: Node3D = $Visual
@onready var body: StaticBody3D = $Body

var _phase: float = 0.0

func _process(delta: float) -> void:
	visual.rotation.y += delta * 1.8
	_phase += delta * 2.5
	visual.position.y = 0.35 + sin(_phase) * 0.08


func set_drop_id(drop_id: String) -> void:
	body.set_meta(&"drop_id", drop_id)
	_phase = float(abs(drop_id.hash()) % 100) / 100.0 * TAU
