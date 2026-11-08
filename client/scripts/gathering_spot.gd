class_name GatheringSpot
extends Node3D

@onready var body: StaticBody3D = $Body
@onready var default_collision: CollisionShape3D = $Body/CollisionShape3D
@onready var tree_collision: CollisionShape3D = $Body/TreeCollision
@onready var large_tree_collision: CollisionShape3D = $Body/LargeTreeCollision
@onready var fishing_collision: CollisionShape3D = $Body/FishingCollision
@onready var agility_collision: CollisionShape3D = $Body/AgilityCollision
@onready var altar_collision: CollisionShape3D = $Body/AltarCollision
@onready var furnace_collision: CollisionShape3D = $Body/FurnaceCollision
@onready var cooking_range_collision: CollisionShape3D = $Body/CookingRangeCollision
@onready var label: Label3D = $Label
@onready var pond_visual: Node3D = $PondVisual
@onready var range_visual: Node3D = $RangeVisual
@onready var mine_visual: Node3D = $MineVisual
@onready var furnace_visual: Node3D = $FurnaceVisual
@onready var anvil_visual: Node3D = $AnvilVisual
@onready var tree_visual: Node3D = $TreeVisual
@onready var agility_visual: Node3D = $AgilityVisual
@onready var altar_visual: Node3D = $AltarVisual
@onready var altar_holy_mark_vertical: MeshInstance3D = $AltarVisual/HolyMarkVertical
@onready var altar_holy_mark_horizontal: MeshInstance3D = $AltarVisual/HolyMarkHorizontal
@onready var altar_rune_orb: MeshInstance3D = $AltarVisual/RuneOrb
@onready var altar_surfaces: Array[MeshInstance3D] = [
	$AltarVisual/TableTop,
	$AltarVisual/LeftSupport,
	$AltarVisual/RightSupport,
	$AltarVisual/ClothTop,
	$AltarVisual/ClothFront,
]
@onready var full_tree_visual: Node3D = $TreeVisual/FullTree
@onready var stump_visual: Node3D = $TreeVisual/Stump
@onready var tree_trunk: MeshInstance3D = $TreeVisual/FullTree/Trunk
@onready var canopy_a: MeshInstance3D = $TreeVisual/FullTree/CanopyA
@onready var canopy_b: MeshInstance3D = $TreeVisual/FullTree/CanopyB
@onready var yew_top: MeshInstance3D = $TreeVisual/FullTree/YewTop
@onready var willow_drapes: Array[MeshInstance3D] = [
	$TreeVisual/FullTree/WillowDrapeLeft,
	$TreeVisual/FullTree/WillowDrapeRight,
	$TreeVisual/FullTree/WillowDrapeRear,
]
@onready var ore_meshes: Array[MeshInstance3D] = [$MineVisual/CopperA, $MineVisual/CopperB]
@onready var foliage_meshes: Array[MeshInstance3D] = [
	canopy_a, canopy_b, yew_top,
	willow_drapes[0], willow_drapes[1], willow_drapes[2],
]


func set_spot(id: String, display_name: String, skill: String, station: String = "") -> void:
	body.set_meta(&"gathering_spot_id", id)
	label.text = "%s\n%s" % [display_name, skill.capitalize()]
	var is_cooking_range: bool = skill == "cooking"
	var is_mine: bool = skill == "mining"
	var is_furnace: bool = station == "smelting"
	var is_anvil: bool = station == "forging"
	var is_tree: bool = skill == "woodcutting"
	var is_fishing: bool = skill == "fishing"
	var is_large_tree: bool = is_tree and id in ["starter_oak_tree", "starter_willow_tree", "starter_yew_tree"]
	var is_agility: bool = skill == "agility"
	var is_prayer_altar: bool = station == "prayer_altar"
	var is_runecrafting_altar: bool = station == "runecrafting"
	var is_altar: bool = is_prayer_altar or is_runecrafting_altar
	pond_visual.visible = not is_cooking_range and not is_mine and not is_furnace and not is_anvil and not is_tree and not is_agility and not is_altar
	range_visual.visible = is_cooking_range
	mine_visual.visible = is_mine
	furnace_visual.visible = is_furnace
	anvil_visual.visible = is_anvil
	tree_visual.visible = is_tree
	agility_visual.visible = is_agility
	altar_visual.visible = is_altar
	default_collision.disabled = is_tree or is_agility or is_altar or is_furnace or is_fishing or is_cooking_range
	tree_collision.disabled = not is_tree or is_large_tree
	large_tree_collision.disabled = not is_large_tree
	fishing_collision.disabled = not is_fishing
	agility_collision.disabled = not is_agility
	altar_collision.disabled = not is_altar
	furnace_collision.disabled = not is_furnace
	cooking_range_collision.disabled = not is_cooking_range
	label.position.y = 3.65 if is_tree else 3.35 if is_furnace else 2.15 if is_altar else 1.6 if is_agility else 1.4 if is_anvil else 1.15
	altar_holy_mark_vertical.visible = is_prayer_altar
	altar_holy_mark_horizontal.visible = is_prayer_altar
	altar_rune_orb.visible = is_runecrafting_altar
	if is_runecrafting_altar:
		var altar_material := StandardMaterial3D.new()
		altar_material.albedo_color = Color(0.18, 0.42, 0.62)
		altar_material.metallic = 0.25
		altar_material.roughness = 0.5
		for altar_surface: MeshInstance3D in altar_surfaces:
			altar_surface.material_override = altar_material
		var orb_material := StandardMaterial3D.new()
		orb_material.albedo_color = Color(0.72, 0.92, 1.0)
		orb_material.emission_enabled = true
		orb_material.emission = Color(0.28, 0.72, 1.0)
		orb_material.emission_energy_multiplier = 2.0
		altar_rune_orb.material_override = orb_material
	if is_tree:
		_configure_tree_shape(id)
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
			"starter_gold_vein": Color(0.82, 0.65, 0.16),
			"starter_silver_vein": Color(0.78, 0.84, 0.88),
			"starter_clay_deposit": Color(0.58, 0.32, 0.22),
			"starter_adamantite_vein": Color(0.18, 0.62, 0.36),
		}.get(id, Color(0.74, 0.31, 0.12))
		var ore_material := StandardMaterial3D.new()
		ore_material.albedo_color = ore_colour
		ore_material.metallic = 0.55
		ore_material.roughness = 0.4
		for ore_mesh: MeshInstance3D in ore_meshes:
			ore_mesh.material_override = ore_material


func _configure_tree_shape(id: String) -> void:
	if id == "starter_oak_tree":
		tree_trunk.position = Vector3(0, 1.5, 0)
		tree_trunk.scale = Vector3(1.55, 1.35, 1.55)
		canopy_a.position = Vector3(-0.5, 2.85, 0)
		canopy_a.scale = Vector3(1.35, 1.1, 1.35)
		canopy_b.position = Vector3(0.65, 3.0, 0.2)
		canopy_b.scale = Vector3(1.15, 1.05, 1.15)
		label.position.y = 4.25
	elif id == "starter_willow_tree":
		full_tree_visual.rotation_degrees.z = 10.0
		tree_trunk.position = Vector3(0, 1.45, 0)
		tree_trunk.scale = Vector3(0.85, 1.35, 0.85)
		canopy_a.position = Vector3(-0.45, 2.8, 0)
		canopy_a.scale = Vector3(1.2, 0.9, 1.2)
		canopy_b.position = Vector3(0.55, 2.9, 0.16)
		canopy_b.scale = Vector3(1.08, 0.85, 1.08)
		for drape: MeshInstance3D in willow_drapes:
			drape.visible = true
		label.position.y = 4.2
	elif id == "starter_yew_tree":
		tree_trunk.position = Vector3(0, 1.9, 0)
		tree_trunk.scale = Vector3(1.3, 1.7, 1.3)
		canopy_a.position = Vector3(-0.55, 3.65, 0)
		canopy_a.scale = Vector3(1.45, 0.9, 1.45)
		canopy_b.position = Vector3(0.7, 3.9, 0.1)
		canopy_b.scale = Vector3(1.25, 0.85, 1.25)
		yew_top.visible = true
		label.position.y = 5.65


func set_amount(current_amount: int, maximum_amount: int) -> void:
	if maximum_amount <= 0:
		return
	if mine_visual.visible:
		for ore_mesh: MeshInstance3D in ore_meshes:
			ore_mesh.visible = current_amount > 0
	if tree_visual.visible:
		full_tree_visual.visible = current_amount > 0
		stump_visual.visible = current_amount <= 0
