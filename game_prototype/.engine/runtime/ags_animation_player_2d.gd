## ags_animation_player_2d.gd — Billboard animation player for AGS3D (T-GS29)
##
## Implements the AGSAnimationPlayerBase API for 2D billboard characters.
## Rather than AnimationPlayer clips, animation states correspond to contiguous
## column segments within each direction row of the sprite sheet.
##
## Sprite sheet column layout (all states have equal column width = frames_per_state):
##
##   col 0 .. fps-1               → "idle"  (slot 0)
##   col fps .. 2*fps-1           → "walk"  (slot 1)
##   col 2*fps .. 3*fps-1         → "talk"  (slot 2)
##
## The Sprite3D.hframes must equal frames_per_state * len(states).
##
## Usage:
##   Add as a child of AGSCharacter2D alongside AGSBillboardController.
##   Set frames_per_state and fps to match the sprite sheet dimensions.
##   Wire state via set_state("idle"|"walk"|"talk").

extends "res://m10_game_systems/runtime/ags_animation_player_base.gd"

## Frames per state per direction row (e.g. 6 = 6 idle + 6 walk + 6 talk columns).
@export var frames_per_state: int = 1

## Animation playback speed (frames per second).
@export var fps: float = 8.0

## NodePath to the sibling AGSBillboardController.
@export var controller_path: NodePath = NodePath("../AGSBillboardController")

## Maps state name → column slot index (0 = idle, 1 = walk, 2 = talk).
const STATE_SLOTS: Dictionary = {
	"idle": 0,
	"walk": 1,
	"talk": 2,
}

var _controller = null
var _current_state: String = "idle"
var _stopped: bool = false


func _ready() -> void:
	_controller = get_node_or_null(controller_path)
	if _controller == null:
		push_error(
			"AGSAnimationPlayer2D: no AGSBillboardController at '%s'" % str(controller_path))
	_apply_state("idle")


## Play a clip by name (treated as a state name).
func play_clip(clip_name: String) -> void:
	_stopped = false
	_apply_state(clip_name if STATE_SLOTS.has(clip_name) else "idle")


## Freeze the sprite on the current frame.
func stop() -> void:
	_stopped = true
	if _controller != null:
		_controller.fps = 0.0


## Switch to a named state.  No-op if the state is already active.
func set_state(state: String) -> void:
	if state == _current_state and not _stopped:
		return
	_stopped = false
	_apply_state(state)


## Forward animation events (sound/signal prefixes).
func on_anim_event(label: String) -> void:
	if label.begins_with("sound:"):
		AGSRuntime.play_sound(label.substr(6))
	elif label.begins_with("signal:"):
		var signal_name := label.substr(7)
		var parent := get_parent()
		if parent != null and parent.has_signal(signal_name):
			parent.emit_signal(signal_name)


# ─── private ──────────────────────────────────────────────────────────────────

func _apply_state(state: String) -> void:
	_current_state = state
	if _controller == null:
		return

	var slot: int = STATE_SLOTS.get(state, 0)
	_controller.frame_offset = slot * frames_per_state
	_controller.frames_per_angle = frames_per_state
	_controller.fps = fps
	# Reset cycling to the start of the new state immediately.
	_controller._current_anim_frame = 0
	_controller._frame_timer = 0.0
