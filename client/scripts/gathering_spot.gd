class_name GatheringSpot
extends Node3D

@onready var body: StaticBody3D = $Body
@onready var label: Label3D = $Label
@onready var pond_visual: Node3D = $PondVisual
@onready var range_visual: Node3D = $RangeVisual
@onready var mine_visual: Node3D = $MineVisual
@onready var furnace_visual: Node3D = $FurnaceVisual
@onready var anvil_visual: Node3D = $AnvilVisual
@onready var ore_meshes: Array[MeshInstance3D] = [$MineVisual/CopperA, $MineVisual/CopperB]


func set_spot(id: String, display_name: String, skill: String, station: String = "") -> void:
	body.set_meta(&"gathering_spot_id", id)
	label.text = "%s\n%s" % [display_name, skill.capitalize()]
	var is_cooking_range: bool = skill == "cooking"
	var is_mine: bool = skill == "mining"
	var is_furnace: bool = station == "smelting"
	var is_anvil: bool = station == "forging"
	pond_visual.visible = not is_cooking_range and not is_mine and not is_furnace and not is_anvil
	range_visual.visible = is_cooking_range
	mine_visual.visible = is_mine
	furnace_visual.visible = is_furnace
	anvil_visual.visible = is_anvil
	label.position.y = 2.45 if is_furnace else 1.4 if is_anvil else 1.15
	if is_mine:
		var ore_colour: Color = {
			"willowbrook_tin_vein": Color(0.72, 0.75, 0.72),
			"willowbrook_iron_vein": Color(0.62, 0.67, 0.7),
			"willowbrook_coal_vein": Color(0.08, 0.09, 0.1),
			"willowbrook_mithril_vein": Color(0.18, 0.72, 0.78),
		}.get(id, Color(0.74, 0.31, 0.12))
		var ore_material := StandardMaterial3D.new()
		ore_material.albedo_color = ore_colour
		ore_material.metallic = 0.55
		ore_material.roughness = 0.4
		for ore_mesh: MeshInstance3D in ore_meshes:
			ore_mesh.material_override = ore_material


func set_amount(current_amount: int, maximum_amount: int) -> void:
	if maximum_amount <= 0 or not mine_visual.visible:
		return
	for ore_mesh: MeshInstance3D in ore_meshes:
		ore_mesh.visible = current_amount > 0
