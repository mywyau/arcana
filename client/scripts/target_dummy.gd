class_name TargetDummy
extends Node3D

const REST_COLOR := Color(0.75, 0.16, 0.12, 1.0)
const CREATURE_COLOR := Color(0.34, 0.52, 0.2, 1.0)
const BIRD_COLOR := Color(0.76, 0.62, 0.25, 1.0)
const COW_COLOR := Color(0.42, 0.25, 0.14, 1.0)
const BOSS_COLOR := Color(0.28, 0.08, 0.42, 1.0)
const RANGED_BOSS_COLOR := Color(0.12, 0.42, 0.2, 1.0)
const MAGE_BOSS_COLOR := Color(0.12, 0.28, 0.62, 1.0)
const DUST_COLOR := Color(0.76, 0.76, 0.68, 1.0)
const CINDER_COLOR := Color(0.86, 0.24, 0.08, 1.0)
const TIDE_COLOR := Color(0.12, 0.48, 0.82, 1.0)
const THORN_COLOR := Color(0.22, 0.5, 0.16, 1.0)

@onready var visual: Node3D = $Visual
@onready var attack_pivot: Node3D = $Visual/AttackPivot
@onready var mesh: MeshInstance3D = $Visual/AttackPivot/Mesh
@onready var hit_splat: Node3D = $HitSplat
@onready var target_body: StaticBody3D = $TargetBody
@onready var collision_shape: CollisionShape3D = $TargetBody/CollisionShape3D

var _material: StandardMaterial3D
var _hit_tween: Tween
var _attack_tween: Tween
var _defeated: bool = false
var _mobile: bool = false
var _has_position: bool = false
var _authoritative_position: Vector3
var _rest_color: Color = REST_COLOR
var _target_kind: String = "dummy"
var _rest_attack_scale: Vector3 = Vector3.ONE
var _telegraph_mesh: MeshInstance3D
var _telegraph_material: StandardMaterial3D
var _telegraph_disc: CylinderMesh
var _telegraph_line: BoxMesh
var _arrow_volley_meshes: Array[MeshInstance3D] = []
var _telegraph_active: bool = false
var _telegraph_time: float = 0.0
var _telegraph_color := Color(1.0, 0.12, 0.08, 1.0)
var _vulnerable_until_pulse: int = 0
var _vulnerability_time_remaining: float = 0.0


func _ready() -> void:
	var source := mesh.get_active_material(0) as StandardMaterial3D
	if source != null:
		_material = source.duplicate() as StandardMaterial3D
		mesh.material_override = _material
	_create_telegraph_visual()


func _process(delta: float) -> void:
	if _telegraph_active:
		_telegraph_time += delta
		var pulse: float = 0.58 + sin(_telegraph_time * 8.0) * 0.18
		_telegraph_material.albedo_color = Color(
			_telegraph_color.r, _telegraph_color.g, _telegraph_color.b, pulse
		)
	if _vulnerability_time_remaining > 0.0:
		_vulnerability_time_remaining = maxf(0.0, _vulnerability_time_remaining - delta)
		if _material != null:
			var glow: float = 0.48 + sin(Time.get_ticks_msec() * 0.012) * 0.16
			_material.albedo_color = _rest_color.lerp(Color(1.0, 0.72, 0.18, 1.0), glow)
	elif _vulnerable_until_pulse != 0:
		_vulnerable_until_pulse = 0
		if _material != null and not _defeated:
			_material.albedo_color = _rest_color
	if not _mobile or not _has_position:
		return
	var weight: float = 1.0 - exp(-14.0 * delta)
	global_position = global_position.lerp(_authoritative_position, weight)


func _create_telegraph_visual() -> void:
	_telegraph_mesh = MeshInstance3D.new()
	_telegraph_mesh.name = "AttackTelegraph"
	_telegraph_disc = CylinderMesh.new()
	_telegraph_disc.top_radius = 1.0
	_telegraph_disc.bottom_radius = 1.0
	_telegraph_disc.height = 0.035
	_telegraph_disc.radial_segments = 64
	_telegraph_line = BoxMesh.new()
	_telegraph_material = StandardMaterial3D.new()
	_telegraph_material.transparency = BaseMaterial3D.TRANSPARENCY_ALPHA
	_telegraph_material.shading_mode = BaseMaterial3D.SHADING_MODE_UNSHADED
	_telegraph_material.albedo_color = Color(1.0, 0.12, 0.08, 0.65)
	_telegraph_disc.material = _telegraph_material
	_telegraph_line.material = _telegraph_material
	_telegraph_mesh.mesh = _telegraph_disc
	_telegraph_mesh.cast_shadow = GeometryInstance3D.SHADOW_CASTING_SETTING_OFF
	add_child(_telegraph_mesh)
	_telegraph_mesh.top_level = true
	_telegraph_mesh.visible = false
	for index in range(9):
		var circle := MeshInstance3D.new()
		circle.name = "ArrowVolleyCircle%d" % index
		circle.mesh = _telegraph_disc
		circle.cast_shadow = GeometryInstance3D.SHADOW_CASTING_SETTING_OFF
		add_child(circle)
		circle.top_level = true
		circle.visible = false
		_arrow_volley_meshes.append(circle)


func set_attack_telegraph(
	telegraph_type: String,
	telegraph_position: Vector3,
	radius: float,
) -> void:
	if _telegraph_mesh == null:
		return
	var arrow_volley: bool = telegraph_type.begins_with("arrow_volley_")
	var elemental_zone: bool = telegraph_type.begins_with("elemental_zone_")
	var active: bool = (
		telegraph_type in ["heavy_strike", "piercing_shot"] or arrow_volley or elemental_zone
	) and radius > 0.0
	_telegraph_active = active
	_telegraph_mesh.visible = active and not arrow_volley and not elemental_zone
	for circle in _arrow_volley_meshes:
		circle.visible = false
	if not active:
		_telegraph_time = 0.0
		return
	_telegraph_mesh.rotation = Vector3.ZERO
	if elemental_zone:
		var wave := maxi(0, int(telegraph_type.trim_prefix("elemental_zone_")) - 1)
		var zone_colors := [Color(1.0, 0.24, 0.06), Color(0.1, 0.48, 1.0), Color(0.2, 0.72, 0.3)]
		_telegraph_color = zone_colors[wave % zone_colors.size()]
		var offsets := _elemental_zone_offsets(wave)
		for index in range(mini(offsets.size(), _arrow_volley_meshes.size())):
			var circle := _arrow_volley_meshes[index]
			circle.global_position = telegraph_position + offsets[index] + Vector3(0.0, 0.05, 0.0)
			circle.scale = Vector3(radius, 1.0, radius)
			circle.visible = true
	elif arrow_volley:
		_telegraph_color = Color(1.0, 0.12, 0.08, 1.0)
		var wave := maxi(0, int(telegraph_type.trim_prefix("arrow_volley_")) - 1)
		var offsets := _arrow_volley_offsets(wave)
		for index in range(mini(offsets.size(), _arrow_volley_meshes.size())):
			var circle := _arrow_volley_meshes[index]
			circle.global_position = telegraph_position + offsets[index] + Vector3(0.0, 0.05, 0.0)
			circle.scale = Vector3(radius, 1.0, radius)
			circle.visible = true
	elif telegraph_type == "piercing_shot":
		_telegraph_color = Color(1.0, 0.12, 0.08, 1.0)
		var start := global_position
		var offset := telegraph_position - start
		var length := Vector2(offset.x, offset.z).length()
		if length <= 0.01:
			_telegraph_mesh.visible = false
			return
		_telegraph_line.size = Vector3(radius, 0.035, length)
		_telegraph_mesh.mesh = _telegraph_line
		_telegraph_mesh.global_position = (start + telegraph_position) * 0.5 + Vector3(0.0, 0.05, 0.0)
		_telegraph_mesh.rotation.y = atan2(offset.x, offset.z)
		_telegraph_mesh.scale = Vector3.ONE
	else:
		_telegraph_color = Color(1.0, 0.12, 0.08, 1.0)
		_telegraph_mesh.mesh = _telegraph_disc
		_telegraph_mesh.global_position = telegraph_position + Vector3(0.0, 0.05, 0.0)
		_telegraph_mesh.scale = Vector3(radius, 1.0, radius)


# Mirrors server arrowVolleyPositions so each snapshot only needs the wave ID.
func _arrow_volley_offsets(wave: int) -> Array[Vector3]:
	var patterns: Array[Array] = [
		[Vector3.ZERO, Vector3(-4.8, 0, -2), Vector3(3.9, 0, -3.4), Vector3(-2.2, 0, 4.6), Vector3(5.2, 0, 2.1), Vector3(-6, 0, 3), Vector3(1.5, 0, 6.2), Vector3(6.5, 0, -0.8), Vector3(-1, 0, -6)],
		[Vector3.ZERO, Vector3(-5.5, 0, -1), Vector3(2.8, 0, -5), Vector3(-3.8, 0, 4.2), Vector3(5.8, 0, 3.4), Vector3(-6.2, 0, -4), Vector3(0.8, 0, 6.5), Vector3(6, 0, -2.5), Vector3(-1.5, 0, 2.8)],
		[Vector3.ZERO, Vector3(-4.2, 0, -4.8), Vector3(4.9, 0, -2), Vector3(-5.8, 0, 1.7), Vector3(3.2, 0, 5.5), Vector3(-1.2, 0, 6.7), Vector3(6.5, 0, 1.2), Vector3(1.1, 0, -6.4), Vector3(-2.7, 0, 3.2)],
	]
	var offsets: Array[Vector3] = []
	for offset: Vector3 in patterns[wave % patterns.size()]:
		offsets.append(offset)
	return offsets


# Mirrors server elementalZonePositions.
func _elemental_zone_offsets(wave: int) -> Array[Vector3]:
	var patterns: Array[Array] = [
		[Vector3(-12, 0, 0), Vector3(-6, 0, 0), Vector3.ZERO, Vector3(6, 0, 0), Vector3(12, 0, 0)],
		[Vector3(0, 0, -12), Vector3(0, 0, -6), Vector3.ZERO, Vector3(0, 0, 6), Vector3(0, 0, 12)],
		[Vector3(-10, 0, -10), Vector3(-5, 0, -5), Vector3.ZERO, Vector3(5, 0, 5), Vector3(10, 0, 10)],
	]
	var offsets: Array[Vector3] = []
	for offset: Vector3 in patterns[wave % patterns.size()]:
		offsets.append(offset)
	return offsets


func set_vulnerable_until(until_pulse: int, current_pulse: int) -> void:
	if until_pulse == _vulnerable_until_pulse:
		return
	_vulnerable_until_pulse = until_pulse
	_vulnerability_time_remaining = (
		float(maxi(0, until_pulse - current_pulse)) * 0.6
	)
	if _vulnerability_time_remaining <= 0.0 and _material != null and not _defeated:
		_material.albedo_color = _rest_color


func set_target_id(target_id: String) -> void:
	target_body.set_meta(&"combat_target_id", target_id)


func set_target_kind(kind: String) -> void:
	_target_kind = kind
	_mobile = kind in [
		"creature", "bird", "cow", "dust", "cinder", "tide", "thorn",
		"boss", "ranged_boss", "mage_boss",
	]
	if kind == "bird":
		_rest_color = BIRD_COLOR
		_rest_attack_scale = Vector3(0.82, 0.48, 1.0)
		collision_shape.position = Vector3(0.0, 0.55, 0.0)
		collision_shape.scale = Vector3(0.8, 0.5, 0.8)
	elif kind == "creature":
		_rest_color = CREATURE_COLOR
		_rest_attack_scale = Vector3(1.15, 0.75, 1.25)
		collision_shape.position = Vector3(0.0, 1.0, 0.0)
		collision_shape.scale = Vector3.ONE
	elif kind == "cow":
		_rest_color = COW_COLOR
		_rest_attack_scale = Vector3(1.35, 0.8, 1.65)
		collision_shape.position = Vector3(0.0, 0.8, 0.0)
		collision_shape.scale = Vector3(1.15, 0.75, 1.35)
	elif kind in ["dust", "cinder", "tide", "thorn"]:
		match kind:
			"dust": _rest_color = DUST_COLOR
			"cinder": _rest_color = CINDER_COLOR
			"tide": _rest_color = TIDE_COLOR
			"thorn": _rest_color = THORN_COLOR
		_rest_attack_scale = Vector3(0.95, 1.05, 0.95)
		collision_shape.position = Vector3(0.0, 0.9, 0.0)
		collision_shape.scale = Vector3(0.9, 0.9, 0.9)
	elif kind in ["boss", "ranged_boss", "mage_boss"]:
		if kind == "ranged_boss":
			_rest_color = RANGED_BOSS_COLOR
		elif kind == "mage_boss":
			_rest_color = MAGE_BOSS_COLOR
		else:
			_rest_color = BOSS_COLOR
		_rest_attack_scale = Vector3(2.1, 1.55, 2.1)
		collision_shape.position = Vector3(0.0, 1.45, 0.0)
		collision_shape.scale = Vector3(1.7, 1.45, 1.7)
	else:
		_rest_color = REST_COLOR
		_rest_attack_scale = Vector3.ONE
		collision_shape.position = Vector3(0.0, 1.0, 0.0)
		collision_shape.scale = Vector3.ONE
	attack_pivot.scale = _rest_attack_scale
	if _material != null and not _defeated:
		_material.albedo_color = _rest_color


func set_authoritative_position(authoritative_position: Vector3) -> void:
	_authoritative_position = authoritative_position
	if not _has_position or not _mobile or (
		_target_kind == "mage_boss" and global_position.distance_to(authoritative_position) > 6.0
	):
		global_position = authoritative_position
	_has_position = true


func play_hit(
	damage: int,
	attacker_position: Vector3,
	feedback_color: Color = Color(1.0, 0.25, 0.18, 1.0),
) -> void:
	if _hit_tween != null:
		_hit_tween.kill()

	visual.position = Vector3.ZERO
	visual.scale = Vector3.ONE
	hit_splat.call("show_damage", damage, feedback_color)
	if _material != null:
		_material.albedo_color = Color(1.0, 0.9, 0.82, 1.0)

	var recoil := global_position - attacker_position
	recoil.y = 0.0
	if recoil.is_zero_approx():
		recoil = Vector3.FORWARD
	recoil = recoil.normalized() * 0.22

	_hit_tween = create_tween()
	_hit_tween.set_parallel(true)
	_hit_tween.tween_property(visual, "position", recoil, 0.08)
	_hit_tween.tween_property(visual, "position", Vector3.ZERO, 0.14).set_delay(0.08)
	_hit_tween.tween_property(visual, "scale", Vector3(1.08, 0.9, 1.08), 0.08)
	_hit_tween.tween_property(visual, "scale", Vector3.ONE, 0.14).set_delay(0.08)
	if _material != null:
		_hit_tween.tween_property(_material, "albedo_color", _rest_color, 0.18)


func play_dodge(label_text: String = "MISS") -> void:
	hit_splat.call("show_status", label_text, Color(0.18, 0.48, 0.68, 1.0))


func play_block() -> void:
	hit_splat.call("show_status", "BLOCK", Color(0.45, 0.5, 0.56, 1.0))


func play_attack(target_position: Vector3) -> void:
	if _defeated:
		return
	if _attack_tween != null:
		_attack_tween.kill()
	attack_pivot.position = Vector3.ZERO
	attack_pivot.rotation_degrees = Vector3.ZERO
	attack_pivot.scale = _rest_attack_scale
	var direction: Vector3 = visual.to_local(target_position)
	direction.y = 0.0
	if direction.is_zero_approx():
		direction = Vector3.FORWARD
	direction = direction.normalized()
	var windup: Vector3 = -direction * (0.16 if _target_kind == "bird" else 0.22)
	var strike: Vector3 = direction * (0.28 if _target_kind == "bird" else 0.42)
	_attack_tween = create_tween()
	_attack_tween.set_trans(Tween.TRANS_QUAD).set_ease(Tween.EASE_OUT)
	_attack_tween.tween_property(attack_pivot, "position", windup, 0.12)
	_attack_tween.parallel().tween_property(
		attack_pivot,
		"scale",
		_rest_attack_scale * (
			Vector3(1.08, 0.88, 1.08)
			if _target_kind == "bird"
			else Vector3(0.94, 1.06, 0.94)
		),
		0.12,
	)
	_attack_tween.set_trans(Tween.TRANS_CUBIC).set_ease(Tween.EASE_IN)
	_attack_tween.tween_property(attack_pivot, "position", strike, 0.1)
	_attack_tween.parallel().tween_property(attack_pivot, "rotation_degrees:x", -18.0, 0.1)
	_attack_tween.set_trans(Tween.TRANS_QUAD).set_ease(Tween.EASE_OUT)
	_attack_tween.tween_property(attack_pivot, "position", Vector3.ZERO, 0.18)
	_attack_tween.parallel().tween_property(attack_pivot, "rotation_degrees", Vector3.ZERO, 0.18)
	_attack_tween.parallel().tween_property(attack_pivot, "scale", _rest_attack_scale, 0.18)


func set_defeated(defeated: bool) -> void:
	if _defeated == defeated:
		return
	_defeated = defeated
	visual.rotation_degrees.z = 90.0 if defeated else 0.0
	target_body.collision_layer = 0 if defeated else 8
	collision_shape.set_deferred("disabled", defeated)
	if not defeated and _material != null:
		_material.albedo_color = _rest_color
