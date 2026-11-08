class_name GatheringSpot
extends Node3D

@onready var body: StaticBody3D = $Body
@onready var default_collision: CollisionShape3D = $Body/CollisionShape3D
@onready var tree_collision: CollisionShape3D = $Body/TreeCollision
@onready var label: Label3D = $Label
@onready var pond_visual: Node3D = $PondVisual
@onready var range_visual: Node3D = $RangeVisual
@onready var mine_visual: Node3D = $MineVisual
@onready var furnace_visual: Node3D = $FurnaceVisual
@onready var anvil_visual: Node3D = $AnvilVisual
@onready var tree_visual: Node3D = $TreeVisual
@onready var agility_visual: Node3D = $AgilityVisual
@onready var altar_visual: Node3D = $AltarVisual
@onready var full_tree_visual: Node3D = $TreeVisual/FullTree
@onready var stump_visual: Node3D = $TreeVisual/Stump
@onready var ore_meshes: Array[MeshInstance3D] = [$MineVisual/CopperA, $MineVisual/CopperB]
@onready var foliage_meshes: Array[MeshInstance3D] = [$TreeVisual/FullTree/CanopyA, $TreeVisual/FullTree/CanopyB]


func set_spot(id: String, display_name: String, skill: String, station: String = "") -> void:
	body.set_meta(&"gathering_spot_id", id)
	label.text = "%s\n%s" % [display_name, skill.capitalize()]
	var is_cooking_range: bool = skill == "cooking"
	var is_mine: bool = skill == "mining"
	var is_furnace: bool = station == "smelting"
	var is_anvil: bool = station == "forging"
	var is_tree: bool = skill == "woodcutting"
	var is_agility: bool = skill == "agility"
	var is_prayer_altar: bool = station == "prayer_altar"
	pond_visual.visible = not is_cooking_range and not is_mine and not is_furnace and not is_anvil and not is_tree and not is_agility and not is_prayer_altar
	range_visual.visible = is_cooking_range
	mine_visual.visible = is_mine
	furnace_visual.visible = is_furnace
	anvil_visual.visible = is_anvil
	tree_visual.visible = is_tree
	agility_visual.visible = is_agility
	altar_visual.visible = is_prayer_altar
	default_collision.disabled = is_tree
	tree_collision.disabled = not is_tree
	label.position.y = 3.65 if is_tree else 2.45 if is_furnace else 2.15 if is_prayer_altar else 1.6 if is_agility else 1.4 if is_anvil else 1.15
	if is_tree:
		var foliage_material := StandardMaterial3D.new()
		match id:
			"starter_oak_tree": foliage_material.albedo_color = Color(0.24, 0.48, 0.16)
			"starter_willow_tree": foliage_material.albedo_color = Color(0.38, 0.62, 0.32)
			"starter_maple_tree": foliage_material.albedo_color = Color(0.68, 0.38, 0.12)
			"starter_yew_tree": foliage_material.albedo_color = Color(0.12, 0.35, 0.18)
			_: foliage_material.albedo_color = Color(0.31, 0.59, 0.2)
		foliage_material.roughness = 0.95
		for foliage_mesh: MeshInstance3D in foliage_meshes:
			foliage_mesh.material_override = foliage_material
	if is_mine:
		var ore_colour: Color = {
			"starter_tin_vein": Color(0.72, 0.75, 0.72),
			"starter_iron_vein": Color(0.62, 0.67, 0.7),
			"starter_coal_vein": Color(0.08, 0.09, 0.1),
			"starter_mithril_vein": Color(0.18, 0.72, 0.78),
		}.get(id, Color(0.74, 0.31, 0.12))
		var ore_material := StandardMaterial3D.new()
		ore_material.albedo_color = ore_colour
		ore_material.metallic = 0.55
		ore_material.roughness = 0.4
		for ore_mesh: MeshInstance3D in ore_meshes:
			ore_mesh.material_override = ore_material


func set_amount(current_amount: int, maximum_amount: int) -> void:
	if maximum_amount <= 0:
		return
	if mine_visual.visible:
		for ore_mesh: MeshInstance3D in ore_meshes:
			ore_mesh.visible = current_amount > 0
	if tree_visual.visible:
		full_tree_visual.visible = current_amount > 0
		stump_visual.visible = current_amount <= 0
