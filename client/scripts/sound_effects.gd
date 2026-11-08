class_name SoundEffects
extends Node

const VOICE_COUNT: int = 8

const SWORD_SWINGS: Array[AudioStream] = [
	preload("res://assets/sound_effects/sword_slash_01.ogg"),
	preload("res://assets/sound_effects/sword_slash_02.ogg"),
	preload("res://assets/sound_effects/sword_slash_03.ogg"),
]
const DAGGER_SWINGS: Array[AudioStream] = [
	preload("res://assets/sound_effects/dagger_slash_01.ogg"),
	preload("res://assets/sound_effects/dagger_slash_02.ogg"),
	preload("res://assets/sound_effects/dagger_slash_03.ogg"),
]
const BOW_RELEASES: Array[AudioStream] = [
	preload("res://assets/sound_effects/bow_release_01.ogg"),
	preload("res://assets/sound_effects/bow_release_02.ogg"),
	preload("res://assets/sound_effects/bow_release_03.ogg"),
]
const MELEE_IMPACTS: Array[AudioStream] = [
	preload("res://assets/sound_effects/damage_01.ogg"),
	preload("res://assets/sound_effects/damage_02.ogg"),
	preload("res://assets/sound_effects/damage_03.ogg"),
]
const RANGED_IMPACTS: Array[AudioStream] = [
	preload("res://assets/sound_effects/arrow_pierce_01.ogg"),
	preload("res://assets/sound_effects/arrow_pierce_02.ogg"),
	preload("res://assets/sound_effects/arrow_pierce_03.ogg"),
]

var _voices: Array[AudioStreamPlayer] = []
var _next_voice: int = 0
var _mining_sound: AudioStreamWAV
var _woodcutting_sound: AudioStreamWAV
var _cooking_sound: AudioStreamWAV


func _ready() -> void:
	for index: int in range(VOICE_COUNT):
		var voice := AudioStreamPlayer.new()
		voice.name = "Voice%d" % (index + 1)
		voice.bus = &"SFX"
		add_child(voice)
		_voices.append(voice)
	_mining_sound = _make_placeholder_sound(0.18, 920.0, 1480.0, 0.78, 0.04)
	_woodcutting_sound = _make_placeholder_sound(0.16, 150.0, 290.0, 0.7, 0.12)
	_cooking_sound = _make_placeholder_sound(0.24, 260.0, 520.0, 0.4, 0.34)


func play_weapon_attack(weapon_id: String) -> void:
	if weapon_id in ["bow", "crossbow"]:
		_play_variant(BOW_RELEASES, -1.0, 0.94, 1.06)
	elif weapon_id == "dagger" or weapon_id.ends_with("_dagger"):
		_play_variant(DAGGER_SWINGS, -2.0, 0.96, 1.08)
	elif weapon_id != "unarmed" and weapon_id != "staff":
		_play_variant(SWORD_SWINGS, -2.0, 0.9, 1.04)


func play_combat_impact(weapon_id: String) -> void:
	if weapon_id in ["bow", "crossbow"]:
		_play_variant(RANGED_IMPACTS, -1.5, 0.96, 1.06)
	else:
		_play_variant(MELEE_IMPACTS, -2.5, 0.94, 1.05)


func play_action(action: StringName) -> void:
	match action:
		&"mining": _play_stream(_mining_sound, -3.0, randf_range(0.94, 1.06))
		&"woodcutting": _play_stream(_woodcutting_sound, -3.0, randf_range(0.92, 1.04))
		&"cooking": _play_stream(_cooking_sound, -5.0, randf_range(0.96, 1.05))


func _play_variant(
	variants: Array[AudioStream],
	volume_db: float,
	minimum_pitch: float,
	maximum_pitch: float,
) -> void:
	if variants.is_empty():
		return
	_play_stream(
		variants[randi_range(0, variants.size() - 1)],
		volume_db,
		randf_range(minimum_pitch, maximum_pitch),
	)


func _play_stream(stream: AudioStream, volume_db: float, pitch_scale: float) -> void:
	if stream == null or _voices.is_empty():
		return
	var voice: AudioStreamPlayer = _voices[_next_voice]
	_next_voice = (_next_voice + 1) % _voices.size()
	voice.stream = stream
	voice.volume_db = volume_db
	voice.pitch_scale = pitch_scale
	voice.play()


func _make_placeholder_sound(
	duration: float,
	primary_frequency: float,
	secondary_frequency: float,
	tone_mix: float,
	noise_mix: float,
) -> AudioStreamWAV:
	const SAMPLE_RATE: int = 22050
	var sample_count: int = maxi(1, roundi(duration * SAMPLE_RATE))
	var samples := PackedByteArray()
	samples.resize(sample_count * 2)
	for index: int in range(sample_count):
		var time: float = float(index) / float(SAMPLE_RATE)
		var progress: float = float(index) / float(sample_count)
		var envelope: float = pow(1.0 - progress, 2.4)
		var tone: float = (
			sin(TAU * primary_frequency * time) * tone_mix
			+ sin(TAU * secondary_frequency * time) * (1.0 - tone_mix)
		)
		var noise: float = sin(float(index * 73 + 19) * 12.9898) * noise_mix
		var sample: int = clampi(roundi((tone + noise) * envelope * 19000.0), -32768, 32767)
		samples.encode_s16(index * 2, sample)
	var result := AudioStreamWAV.new()
	result.format = AudioStreamWAV.FORMAT_16_BITS
	result.mix_rate = SAMPLE_RATE
	result.stereo = false
	result.data = samples
	return result
