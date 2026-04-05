## ags_audio.gd — Audio manager for AGS3D (T-GS12)
##
## Add as an AutoLoad named AGSAudio. This node manages two AudioStreamPlayers:
##   - one for music (looping, one active at a time)
##   - one pool of players for sound effects (one-shot, fire and forget)
##
## Audio files are loaded from:
##   res://audio/music/<name>.* for music
##   res://audio/sfx/<name>.*   for sound effects
##
## Supported extensions (tried in order): .ogg, .mp3, .wav
##
## Scripts call audio via AGSRuntime:
##   AGSRuntime.play_music("theme_main")   → loads audio/music/theme_main.*
##   AGSRuntime.stop_music()
##   AGSRuntime.play_sound("door_creak")   → loads audio/sfx/door_creak.*
##
## AGSRuntime emits signals; this node connects to them and handles playback.
extends Node

const MUSIC_DIRS := ["res://audio/music/"]
const SFX_DIRS   := ["res://audio/sfx/"]
const EXTENSIONS := ["ogg", "mp3", "wav"]

## Maximum simultaneous sound effect players in the pool.
const SFX_POOL_SIZE := 8

var _music_player: AudioStreamPlayer
var _sfx_pool: Array[AudioStreamPlayer] = []


func _ready() -> void:
	_music_player = AudioStreamPlayer.new()
	_music_player.bus = "Music"
	add_child(_music_player)

	for i in range(SFX_POOL_SIZE):
		var p := AudioStreamPlayer.new()
		p.bus = "SFX"
		add_child(p)
		_sfx_pool.append(p)

	AGSRuntime.play_music_requested.connect(_on_play_music)
	AGSRuntime.stop_music_requested.connect(_on_stop_music)
	AGSRuntime.play_sound_requested.connect(_on_play_sound)


## Load and play a music track by name. Stops any currently playing music first.
func _on_play_music(name: String) -> void:
	var stream := _load_audio(name, MUSIC_DIRS)
	if stream == null:
		push_warning("AGSAudio: music '%s' not found in audio/music/" % name)
		return
	_music_player.stream = stream
	_music_player.stream.loop = true if stream.has_method("set_loop") else false
	_music_player.play()


## Stop the music player immediately.
func _on_stop_music() -> void:
	_music_player.stop()


## Play a one-shot sound effect using the next available pool player.
func _on_play_sound(name: String) -> void:
	var stream := _load_audio(name, SFX_DIRS)
	if stream == null:
		push_warning("AGSAudio: sound '%s' not found in audio/sfx/" % name)
		return
	for player in _sfx_pool:
		if not player.playing:
			player.stream = stream
			player.play()
			return
	# All pool slots busy — use the first slot (oldest sound cut off).
	_sfx_pool[0].stream = stream
	_sfx_pool[0].play()


## Try each supported extension in each directory and return the first stream found.
func _load_audio(name: String, dirs: Array) -> AudioStream:
	for dir in dirs:
		for ext in EXTENSIONS:
			var path := "%s%s.%s" % [dir, name, ext]
			if ResourceLoader.exists(path):
				return ResourceLoader.load(path) as AudioStream
	return null
