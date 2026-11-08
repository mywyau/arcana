class_name PlayerView
extends Node3D

const BRONZE_COLOUR := Color("9a5b32")
const IRON_COLOUR := Color("4b535b")
const STEEL_COLOUR := Color("c8d0d8")
const MITHRIL_COLOUR := Color("2f9fda")
const RUNE_COLOUR := Color("3ca66b")
const TRAINING_COLOUR := Color("7b8790")

@export var interpolation_speed: float = 14.0
@export var movement_speed: float = 6.0
@export var correction_threshold: float = 0.75
@export var snap_threshold: float = 3.0
@export var correction_speed: float = 1.5

@onready var visual_root: Node3D = $VisualRoot
@onready var mesh: MeshInstance3D = $VisualRoot/Humanoid/Torso
@onready var left_arm: Node3D = $VisualRoot/Humanoid/LeftArm
@onready var right_arm: Node3D = $VisualRoot/Humanoid/RightArm
@onready var left_leg: Node3D = $VisualRoot/Humanoid/LeftLeg
@onready var right_leg: Node3D = $VisualRoot/Humanoid/RightLeg
@onready var left_sleeve: MeshInstance3D = $VisualRoot/Humanoid/LeftArm/UpperArm
@onready var right_sleeve: MeshInstance3D = $VisualRoot/Humanoid/RightArm/UpperArm
@onready var hair_top: MeshInstance3D = $VisualRoot/Humanoid/HairTop
@onready var hair_back: MeshInstance3D = $VisualRoot/Humanoid/HairBack
@onready var training_plate: MeshInstance3D = $VisualRoot/Humanoid/TrainingPlate
@onready var training_helm: Node3D = $VisualRoot/Humanoid/TrainingHelm
@onready var swift_body: MeshInstance3D = $VisualRoot/Humanoid/SwiftBody
@onready var swift_chest_trim: MeshInstance3D = $VisualRoot/Humanoid/SwiftChestTrim
@onready var swift_helm: Node3D = $VisualRoot/Humanoid/SwiftHelm
@onready var swift_boots: Node3D = $VisualRoot/Humanoid/SwiftBoots
@onready var cape: MeshInstance3D = $VisualRoot/Cape
@onready var weapon_anchor: Node3D = $VisualRoot/WeaponAnchor
@onready var sword_pivot: Node3D = $VisualRoot/WeaponAnchor/SwordPivot
@onready var blade: MeshInstance3D = $VisualRoot/WeaponAnchor/SwordPivot/Blade
@onready var hilt: MeshInstance3D = $VisualRoot/WeaponAnchor/SwordPivot/Hilt
@onready var bow_pivot: Node3D = $VisualRoot/WeaponAnchor/BowPivot
@onready var crossbow_pivot: Node3D = $VisualRoot/WeaponAnchor/CrossbowPivot
@onready var mace_pivot: Node3D = $VisualRoot/WeaponAnchor/MacePivot
@onready var mace_head: MeshInstance3D = $VisualRoot/WeaponAnchor/MacePivot/Head
@onready var battle_axe_pivot: Node3D = $VisualRoot/WeaponAnchor/BattleAxePivot
@onready var battle_axe_blade: MeshInstance3D = $VisualRoot/WeaponAnchor/BattleAxePivot/Blade
@onready var staff_pivot: Node3D = $VisualRoot/WeaponAnchor/StaffPivot
@onready var damage_label: Label3D = $DamageLabel

var _is_local: bool = false
var _has_position: bool = false
var _remote_target: Vector3
var _authoritative_position: Vector3
var _predicted_path: Array[Vector3] = []
var _pending_correction: Vector3
var _last_route_sequence: int = -1
var _attack_tween: Tween
var _hit_tween: Tween
var _body_material: StandardMaterial3D
var _rest_color: Color
var _defeated: bool = false
var _following_server_path: bool = false
var _equipped_weapon: String = "sword"
var _equipped_armour: Dictionary = {}
var _walk_time: float = 0.0
var _equipment_materials: Dictionary = {}


func _process(delta: float) -> void:
	if not _has_position:
		return
	var previous_position: Vector3 = global_position

	if not _is_local:
		var weight: float = 1.0 - exp(-interpolation_speed * delta)
		global_position = global_position.lerp(_remote_target, weight)
		_update_movement_pose(delta, previous_position)
		return

	if not _predicted_path.is_empty():
		global_position = global_position.move_toward(
			_predicted_path[0],
			movement_speed * delta,
		)
		if global_position.is_equal_approx(_predicted_path[0]) and _predicted_path.size() > 1:
			_predicted_path.pop_front()

	if not _pending_correction.is_zero_approx():
		var correction_step: Vector3 = _pending_correction.limit_length(correction_speed * delta)
		global_position += correction_step
		_pending_correction -= correction_step
	_update_movement_pose(delta, previous_position)


func _update_movement_pose(delta: float, previous_position: Vector3) -> void:
	var motion: Vector3 = global_position - previous_position
	motion.y = 0.0
	var moving: bool = motion.length_squared() > 0.000001 and not _defeated
	if moving:
		visual_root.look_at(visual_root.global_position + motion, Vector3.UP, true)
		_walk_time += delta * 10.0
		var swing: float = sin(_walk_time) * 26.0
		left_leg.rotation_degrees.x = swing
		right_leg.rotation_degrees.x = -swing
		left_arm.rotation_degrees.x = -swing * 0.65
		right_arm.rotation_degrees.x = swing * 0.65
		weapon_anchor.rotation_degrees.x = right_arm.rotation_degrees.x
		visual_root.position.y = absf(sin(_walk_time)) * 0.025
		return
	var settle: float = clampf(delta * 10.0, 0.0, 1.0)
	left_leg.rotation_degrees.x = lerpf(left_leg.rotation_degrees.x, 0.0, settle)
	right_leg.rotation_degrees.x = lerpf(right_leg.rotation_degrees.x, 0.0, settle)
	left_arm.rotation_degrees.x = lerpf(left_arm.rotation_degrees.x, 0.0, settle)
	right_arm.rotation_degrees.x = lerpf(right_arm.rotation_degrees.x, 0.0, settle)
	weapon_anchor.rotation_degrees.x = right_arm.rotation_degrees.x
	visual_root.position.y = lerpf(visual_root.position.y, 0.0, settle)


func predict_move(destination: Vector3) -> void:
	if not _is_local or not _has_position:
		return
	_predicted_path = [destination]


func predict_path(path: Array[Vector3]) -> void:
	if not _is_local or not _has_position:
		return
	_predicted_path = path.duplicate()


func apply_authoritative_state(
	position: Vector3,
	path: Array[Vector3],
	speed: float,
	acknowledged_sequence: int,
	latest_sent_sequence: int,
	server_driven_path: bool,
) -> float:
	_authoritative_position = position
	movement_speed = speed

	if not _has_position:
		global_position = position
		_remote_target = position
		_has_position = true

	if not _is_local:
		_remote_target = position
		return global_position.distance_to(position)

	var correction_distance: float = global_position.distance_to(position)
	if correction_distance >= snap_threshold:
		global_position = position
		_pending_correction = Vector3.ZERO
	elif correction_distance >= correction_threshold:
		_pending_correction = position - global_position
	else:
		_pending_correction = Vector3.ZERO

	if server_driven_path:
		_predicted_path = path.duplicate()
	elif _following_server_path:
		_predicted_path = path.duplicate()
	elif acknowledged_sequence >= latest_sent_sequence and acknowledged_sequence > _last_route_sequence:
		_predicted_path = path.duplicate()
		_last_route_sequence = acknowledged_sequence
	_following_server_path = server_driven_path

	return correction_distance


func cancel_prediction() -> void:
	_predicted_path.clear()
	_pending_correction = _authoritative_position - global_position


func stop_prediction() -> void:
	_predicted_path.clear()
	_pending_correction = Vector3.ZERO


func play_weapon_attack(target_position: Vector3) -> float:
	if _defeated:
		return 0.0
	var flat_target := Vector3(target_position.x, global_position.y, target_position.z)
	if global_position.distance_squared_to(flat_target) > 0.001:
		visual_root.look_at(flat_target, Vector3.UP, true)
	if _equipped_weapon == "unarmed":
		_play_unarmed_attack()
		return 0.0
	if _equipped_weapon == "bow" or _equipped_weapon == "crossbow":
		return _play_bow_attack(target_position)
	if _equipped_weapon == "staff":
		return _play_staff_attack(target_position)
	if _equipped_weapon == "mace" or _equipped_weapon.ends_with("_mace"):
		_play_melee_attack(mace_pivot, -50.0, 65.0, 0.2, 0.28)
		return 0.0
	if _equipped_weapon == "battle_axe" or _equipped_weapon.ends_with("_battle_axe"):
		_play_melee_attack(battle_axe_pivot, -55.0, 75.0, 0.24, 0.34)
		return 0.0
	if _equipped_weapon == "halberd":
		_play_melee_attack(battle_axe_pivot, -55.0, 75.0, 0.24, 0.34)
		return 0.0

	var is_dagger: bool = _equipped_weapon == "dagger" or _equipped_weapon.ends_with("_dagger")
	var forward_duration: float = 0.11 if is_dagger else 0.16
	var return_duration: float = 0.14 if is_dagger else 0.22
	_play_sword_attack(forward_duration, return_duration)
	return 0.0


func _play_unarmed_attack() -> void:
	if _attack_tween != null:
		_attack_tween.kill()
	left_arm.rotation_degrees.x = 0.0
	_attack_tween = create_tween()
	_attack_tween.set_trans(Tween.TRANS_QUAD).set_ease(Tween.EASE_OUT)
	_attack_tween.tween_property(left_arm, "rotation_degrees:x", -82.0, 0.1)
	_attack_tween.set_ease(Tween.EASE_IN_OUT)
	_attack_tween.tween_property(left_arm, "rotation_degrees:x", 0.0, 0.16)


func play_spell_attack(target_position: Vector3, spell_id: String) -> float:
	if _defeated:
		return 0.0
	var flat_target := Vector3(target_position.x, global_position.y, target_position.z)
	if global_position.distance_squared_to(flat_target) > 0.001:
		visual_root.look_at(flat_target, Vector3.UP, true)
	return _play_staff_attack(target_position, spell_id)


func _play_melee_attack(
	pivot: Node3D,
	rest_angle: float,
	attack_angle: float,
	forward_duration: float,
	return_duration: float,
) -> void:
	if _attack_tween != null:
		_attack_tween.kill()
	pivot.rotation_degrees.z = rest_angle
	_attack_tween = create_tween()
	_attack_tween.set_trans(Tween.TRANS_QUAD).set_ease(Tween.EASE_OUT)
	_attack_tween.tween_property(pivot, "rotation_degrees:z", attack_angle, forward_duration)
	_attack_tween.set_ease(Tween.EASE_IN_OUT)
	_attack_tween.tween_property(pivot, "rotation_degrees:z", rest_angle, return_duration)


func _play_sword_attack(forward_duration: float, return_duration: float) -> void:
	if _attack_tween != null:
		_attack_tween.kill()
	var rest_pose := Vector3(-10.0, 0.0, 135.0)
	var windup_pose := Vector3(-28.0, 0.0, 48.0)
	var strike_pose := Vector3(68.0, 0.0, 62.0)
	sword_pivot.rotation_degrees = rest_pose
	_attack_tween = create_tween()
	_attack_tween.set_trans(Tween.TRANS_QUAD).set_ease(Tween.EASE_OUT)
	_attack_tween.tween_property(sword_pivot, "rotation_degrees", windup_pose, forward_duration * 0.45)
	_attack_tween.set_trans(Tween.TRANS_CUBIC).set_ease(Tween.EASE_IN)
	_attack_tween.tween_property(sword_pivot, "rotation_degrees", strike_pose, forward_duration * 0.55)
	_attack_tween.set_trans(Tween.TRANS_QUAD).set_ease(Tween.EASE_IN_OUT)
	_attack_tween.tween_property(sword_pivot, "rotation_degrees", rest_pose, return_duration)


func _play_bow_attack(target_position: Vector3) -> float:
	if _attack_tween != null:
		_attack_tween.kill()
	var ranged_pivot: Node3D = crossbow_pivot if _equipped_weapon == "crossbow" else bow_pivot
	var rest_angle: float = -25.0 if _equipped_weapon == "crossbow" else -10.0
	ranged_pivot.rotation_degrees.z = rest_angle
	_attack_tween = create_tween()
	_attack_tween.set_trans(Tween.TRANS_QUAD).set_ease(Tween.EASE_OUT)
	_attack_tween.tween_property(ranged_pivot, "rotation_degrees:z", rest_angle + 18.0, 0.12)
	_attack_tween.set_ease(Tween.EASE_IN_OUT)
	_attack_tween.tween_property(ranged_pivot, "rotation_degrees:z", rest_angle, 0.18)

	var arrow := MeshInstance3D.new()
	var arrow_mesh := BoxMesh.new()
	arrow_mesh.size = Vector3(0.045, 0.045, 0.36 if _equipped_weapon == "crossbow" else 0.55)
	var arrow_material := StandardMaterial3D.new()
	arrow_material.albedo_color = Color(0.78, 0.58, 0.3)
	arrow_material.roughness = 0.8
	arrow_mesh.material = arrow_material
	arrow.mesh = arrow_mesh
	var projectile_parent: Node = get_tree().current_scene
	if projectile_parent == null:
		projectile_parent = get_tree().root
	projectile_parent.add_child(arrow)
	var arrow_start := global_position + Vector3.UP * 1.15
	var arrow_target := target_position + Vector3.UP * 1.0
	var horizontal_distance := Vector2(
		arrow_target.x - arrow_start.x,
		arrow_target.z - arrow_start.z,
	).length()
	var arc_height: float = maxf(1.25, horizontal_distance * 0.2)
	var arc_control := (arrow_start + arrow_target) * 0.5 + Vector3.UP * arc_height
	arrow.global_position = arrow_start
	arrow.look_at(arc_control, Vector3.UP, true)
	var flight_time: float = clampf(arrow_start.distance_to(arrow_target) / 18.0, 0.18, 0.55)
	var projectile_tween := arrow.create_tween()
	projectile_tween.tween_method(
		_move_arrow_along_arc.bind(arrow, arrow_start, arc_control, arrow_target),
		0.0,
		1.0,
		flight_time,
	).set_trans(Tween.TRANS_LINEAR)
	projectile_tween.tween_callback(arrow.queue_free)
	return flight_time


func _move_arrow_along_arc(
	progress: float,
	arrow: MeshInstance3D,
	start: Vector3,
	control: Vector3,
	target: Vector3,
) -> void:
	if not is_instance_valid(arrow):
		return
	var first_leg := start.lerp(control, progress)
	var second_leg := control.lerp(target, progress)
	var position := first_leg.lerp(second_leg, progress)
	var look_progress: float = minf(1.0, progress + 0.025)
	var look_first := start.lerp(control, look_progress)
	var look_second := control.lerp(target, look_progress)
	var look_position := look_first.lerp(look_second, look_progress)
	arrow.global_position = position
	if not position.is_equal_approx(look_position):
		arrow.look_at(look_position, Vector3.UP, true)


func _play_staff_attack(target_position: Vector3, spell_id: String = "") -> float:
	if _attack_tween != null:
		_attack_tween.kill()
	staff_pivot.rotation_degrees.z = -15.0
	_attack_tween = create_tween()
	_attack_tween.set_trans(Tween.TRANS_QUAD).set_ease(Tween.EASE_OUT)
	_attack_tween.tween_property(staff_pivot, "rotation_degrees:z", 10.0, 0.14)
	_attack_tween.set_ease(Tween.EASE_IN_OUT)
	_attack_tween.tween_property(staff_pivot, "rotation_degrees:z", -15.0, 0.2)

	var bolt := MeshInstance3D.new()
	var bolt_mesh := SphereMesh.new()
	bolt_mesh.radius = 0.17
	bolt_mesh.height = 0.34
	var projectile_color: Color = _spell_projectile_color(spell_id)
	var bolt_material := StandardMaterial3D.new()
	bolt_material.albedo_color = projectile_color
	bolt_material.emission_enabled = true
	bolt_material.emission = projectile_color
	bolt_material.emission_energy_multiplier = 1.8 if spell_id == "dark_bolt" else 3.0
	bolt_mesh.material = bolt_material
	bolt.mesh = bolt_mesh
	var projectile_parent: Node = get_tree().current_scene
	if projectile_parent == null:
		projectile_parent = get_tree().root
	projectile_parent.add_child(bolt)
	bolt.global_position = global_position + Vector3.UP * 1.35
	var bolt_target := target_position + Vector3.UP * 1.0
	var flight_time: float = clampf(bolt.global_position.distance_to(bolt_target) / 16.0, 0.2, 0.65)
	var projectile_tween := bolt.create_tween()
	projectile_tween.tween_property(bolt, "global_position", bolt_target, flight_time)
	projectile_tween.tween_callback(bolt.queue_free)
	return flight_time


func _spell_projectile_color(spell_id: String) -> Color:
	match spell_id:
		"wind_bolt": return Color(0.96, 0.98, 1.0)
		"water_bolt": return Color(0.08, 0.38, 1.0)
		"earth_bolt": return Color(0.16, 0.68, 0.24)
		"fire_bolt": return Color(1.0, 0.06, 0.025)
		"holy_bolt": return Color(1.0, 0.72, 0.02)
		"dark_bolt": return Color(0.3, 0.06, 0.42)
		_: return Color(0.45, 0.86, 1.0)


func set_weapon(weapon_id: String) -> void:
	_equipped_weapon = weapon_id
	var dagger_equipped: bool = weapon_id == "dagger" or weapon_id.ends_with("_dagger")
	var spear_equipped: bool = weapon_id == "spear"
	_update_weapon_visibility()
	_update_weapon_material()
	blade.scale = Vector3(0.8, 0.55, 1.0) if dagger_equipped else (Vector3(0.8, 1.9, 0.8) if spear_equipped else Vector3.ONE)
	blade.position.y = 0.3 if dagger_equipped else 0.5
	hilt.scale = Vector3(0.75, 0.85, 0.85) if dagger_equipped else Vector3.ONE
	sword_pivot.rotation_degrees = Vector3(-10.0, 0.0, 135.0)
	battle_axe_pivot.scale = Vector3(1.0, 1.65, 1.0) if weapon_id == "halberd" else Vector3.ONE


func equipped_weapon() -> String:
	return _equipped_weapon


func set_armour(equipped_armour: Dictionary) -> void:
	if equipped_armour == _equipped_armour:
		return
	_equipped_armour = equipped_armour.duplicate(true)
	var head_item: String = str(_equipped_armour.get("head", ""))
	var body_item: String = str(_equipped_armour.get("body", ""))
	var feet_item: String = str(_equipped_armour.get("feet", ""))
	var cape_item: String = str(_equipped_armour.get("cape", ""))
	training_helm.visible = head_item == "training_helm" or _is_tiered_item(head_item, "helm")
	swift_helm.visible = head_item == "swift_helm"
	var helmet_equipped: bool = training_helm.visible or swift_helm.visible
	hair_top.visible = not helmet_equipped
	hair_back.visible = not helmet_equipped
	training_plate.visible = body_item == "training_plate" or _is_tiered_item(body_item, "platebody")
	swift_body.visible = body_item == "swift_body"
	swift_chest_trim.visible = swift_body.visible
	swift_boots.visible = feet_item == "swift_boots"
	cape.visible = not cape_item.is_empty()
	if cape.visible:
		cape.material_override = _cape_material_for_item(cape_item)
	if training_helm.visible:
		_apply_material_to_meshes(training_helm, _material_for_item(head_item))
	if training_plate.visible:
		training_plate.material_override = _material_for_item(body_item)


func _update_weapon_visibility() -> void:
	var is_sword_or_dagger: bool = (
		_equipped_weapon == "sword" or
		_equipped_weapon == "dagger" or
		_equipped_weapon == "spear" or
		_equipped_weapon.ends_with("_sword") or
		_equipped_weapon.ends_with("_dagger")
	)
	sword_pivot.visible = not _defeated and is_sword_or_dagger
	bow_pivot.visible = not _defeated and _equipped_weapon == "bow"
	crossbow_pivot.visible = not _defeated and _equipped_weapon == "crossbow"
	mace_pivot.visible = not _defeated and (_equipped_weapon == "mace" or _equipped_weapon.ends_with("_mace"))
	battle_axe_pivot.visible = not _defeated and (_equipped_weapon in ["battle_axe", "halberd"] or _equipped_weapon.ends_with("_battle_axe"))
	staff_pivot.visible = not _defeated and _equipped_weapon == "staff"


func _update_weapon_material() -> void:
	var material: StandardMaterial3D = _material_for_item(_equipped_weapon)
	blade.material_override = material
	mace_head.material_override = material
	battle_axe_blade.material_override = material


func _is_tiered_item(item_id: String, item_suffix: String) -> bool:
	return item_id.ends_with("_" + item_suffix) and (
		item_id.begins_with("bronze_") or
		item_id.begins_with("iron_") or
		item_id.begins_with("steel_") or
		item_id.begins_with("mithril_") or
		item_id.begins_with("rune_")
	)


func _material_for_item(item_id: String) -> StandardMaterial3D:
	var tier: String = "training"
	var colour: Color = TRAINING_COLOUR
	if item_id.begins_with("bronze_"):
		tier = "bronze"
		colour = BRONZE_COLOUR
	elif item_id.begins_with("steel_"):
		tier = "steel"
		colour = STEEL_COLOUR
	elif item_id.begins_with("mithril_"):
		tier = "mithril"
		colour = MITHRIL_COLOUR
	elif item_id.begins_with("rune_"):
		tier = "rune"
		colour = RUNE_COLOUR
	elif item_id.begins_with("iron_") or item_id in ["sword", "dagger", "mace", "battle_axe"]:
		tier = "iron"
		colour = IRON_COLOUR
	if _equipment_materials.has(tier):
		return _equipment_materials[tier] as StandardMaterial3D
	var material := StandardMaterial3D.new()
	material.albedo_color = colour
	material.metallic = 0.68
	material.roughness = 0.3
	_equipment_materials[tier] = material
	return material


func _cape_material_for_item(item_id: String) -> StandardMaterial3D:
	var key := "cape:%s" % item_id
	if _equipment_materials.has(key):
		return _equipment_materials[key] as StandardMaterial3D
	var colour := Color(0.28, 0.2, 0.45)
	if item_id == "vanguard_cape":
		colour = Color(0.56, 0.18, 0.18)
	elif item_id == "wayfarer_cape":
		colour = Color(0.18, 0.43, 0.32)
	var material := StandardMaterial3D.new()
	material.albedo_color = colour
	material.roughness = 0.88
	_equipment_materials[key] = material
	return material


func _apply_material_to_meshes(node: Node, material: StandardMaterial3D) -> void:
	if node is MeshInstance3D:
		(node as MeshInstance3D).material_override = material
	for child: Node in node.get_children():
		_apply_material_to_meshes(child, material)


func play_hit(damage: int) -> void:
	if _hit_tween != null:
		_hit_tween.kill()
	visual_root.scale = Vector3.ONE
	damage_label.text = "-%d" % damage
	damage_label.position = Vector3(0.0, 2.1, 0.0)
	damage_label.modulate = Color(1.0, 0.25, 0.18, 1.0)
	damage_label.visible = true
	if _body_material != null:
		_body_material.albedo_color = Color(1.0, 0.9, 0.82, 1.0)

	_hit_tween = create_tween()
	_hit_tween.set_parallel(true)
	_hit_tween.tween_property(visual_root, "scale", Vector3(1.08, 0.88, 1.08), 0.08)
	_hit_tween.tween_property(visual_root, "scale", Vector3.ONE, 0.14).set_delay(0.08)
	_hit_tween.tween_property(damage_label, "position:y", 2.8, 0.55)
	_hit_tween.tween_property(damage_label, "modulate:a", 0.0, 0.55).set_delay(0.15)
	if _body_material != null:
		_hit_tween.tween_property(_body_material, "albedo_color", _rest_color, 0.18)
	_hit_tween.chain().tween_callback(func() -> void: damage_label.visible = false)


func play_dodge() -> void:
	if _hit_tween != null:
		_hit_tween.kill()
	damage_label.text = "DODGE"
	damage_label.position = Vector3(0.0, 2.1, 0.0)
	damage_label.modulate = Color(0.4, 0.85, 1.0, 1.0)
	damage_label.visible = true
	_hit_tween = create_tween()
	_hit_tween.set_parallel(true)
	_hit_tween.tween_property(damage_label, "position:y", 2.8, 0.55)
	_hit_tween.tween_property(damage_label, "modulate:a", 0.0, 0.55).set_delay(0.15)
	_hit_tween.chain().tween_callback(func() -> void: damage_label.visible = false)


func set_defeated(defeated: bool) -> void:
	if _defeated == defeated:
		return
	_defeated = defeated
	visual_root.rotation_degrees.z = 90.0 if defeated else 0.0
	_update_weapon_visibility()
	if defeated:
		_predicted_path.clear()
		_pending_correction = Vector3.ZERO


func set_local_player(is_local: bool) -> void:
	_is_local = is_local
	_rest_color = Color("58d68d") if is_local else Color("5dade2")
	_body_material = StandardMaterial3D.new()
	_body_material.albedo_color = _rest_color
	_body_material.roughness = 0.8
	mesh.material_override = _body_material
	left_sleeve.material_override = _body_material
	right_sleeve.material_override = _body_material
