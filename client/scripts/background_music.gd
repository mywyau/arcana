class_name BackgroundMusic
extends AudioStreamPlayer

const TRACKS: Array[AudioStream] = [
	preload("res://assets/audio/music/a_life_begins_revised_adventure.ogg"),
	preload("res://assets/audio/music/unnamed_area_theme.ogg"),
	preload("res://assets/audio/music/wanderers_glade.ogg"),
]

var _track_index: int = -1


func _ready() -> void:
	finished.connect(_play_next_track)
	_play_next_track()


func _play_next_track() -> void:
	if TRACKS.is_empty():
		return
	_track_index = (_track_index + 1) % TRACKS.size()
	stream = TRACKS[_track_index]
	if stream is AudioStreamOggVorbis:
		(stream as AudioStreamOggVorbis).loop = false
	play()
