## ags_animation_player_3d.gd — 3D mesh animation player for AGS3D (T-GS28)
##
## Wraps a Godot AnimationPlayer that lives as a sibling node (or child of the
## character scene). The .agchar `animations` block provides the clip name mapping:
##   idle  → AnimationPlayer clip "Idle"
##   walk  → AnimationPlayer clip "Walk"
##   talk  → AnimationPlayer clip "Talk"
##
## Usage:
##   Add as a child of AGSCharacter3D. Set [member animation_player_path] to
##   point at the AnimationPlayer node. Wire clip names via [member clips].
##
## Animation is driven in T-BL13; this class provides the API bridge now so
## the character runtime can call set_state() without knowing the character type.
extends "res://m10_game_systems/runtime/ags_animation_player_base.gd"

## NodePath to the AnimationPlayer sibling/child (relative to this node's parent).
## Example: "../AnimationPlayer" or "AnimationPlayer"
@export var animation_player_path: NodePath = NodePath("../AnimationPlayer")

## Maps state names → AnimationPlayer clip names.
## Defaults match the .agchar `animations` block conventions.
@export var clips: Dictionary = {
	"idle": "Idle",
	"walk": "Walk",
	"talk": "Talk",
}

var _anim_player: AnimationPlayer = null
var _current_state: String = ""


func _ready() -> void:
	if not animation_player_path.is_empty():
		_anim_player = get_node_or_null(animation_player_path)


## Play the named clip directly on the AnimationPlayer.
func play_clip(clip_name: String) -> void:
	if _anim_player == null:
		return
	if _anim_player.has_animation(clip_name):
		_anim_player.play(clip_name)
	else:
		push_warning("AGSAnimationPlayer3D: clip '%s' not found in AnimationPlayer" % clip_name)


## Stop the AnimationPlayer and freeze on the current frame.
func stop() -> void:
	if _anim_player != null:
		_anim_player.stop()


## Map a character state name to its clip and play it.
## Ignores the call if the state is already active (avoids restarting a playing clip).
func set_state(state: String) -> void:
	if state == _current_state:
		return
	_current_state = state
	var clip: String = clips.get(state, "")
	if clip.is_empty():
		push_warning("AGSAnimationPlayer3D: no clip mapped for state '%s'" % state)
		return
	play_clip(clip)


## Forward animation frame events to AGSRuntime or the character's room script.
## Label prefixes: "sound:<name>", "signal:<name>".
func on_anim_event(label: String) -> void:
	if label.begins_with("sound:"):
		var sound_name := label.substr(6)
		AGSRuntime.play_sound(sound_name)
	elif label.begins_with("signal:"):
		var signal_name := label.substr(7)
		var parent := get_parent()
		if parent != null and parent.has_signal(signal_name):
			parent.emit_signal(signal_name)
