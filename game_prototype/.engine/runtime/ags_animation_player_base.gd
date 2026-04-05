## ags_animation_player_base.gd — Abstract animation player API for AGS3D (T-GS28)
##
## AGSAnimationPlayerBase defines the common interface all animation player
## implementations must satisfy. Attach a concrete subclass to an AGSCharacterBase
## node; the character runtime calls this interface without knowing whether the
## character is 3D, 2D billboard, or puppet.
##
## States correspond to character behaviour:
##   "idle"  — character is stationary
##   "walk"  — character is moving
##   "talk"  — character is speaking (Say / Think)
##
## Concrete implementations:
##   AGSAnimationPlayer3D  — drives a Godot AnimationPlayer on a 3D mesh
##   AGSAnimationPlayer2D  — drives Sprite3D frame selection for billboard chars (T-GS29)
extends Node

## Play a named animation clip immediately (interrupts current clip).
## [param clip_name] must match a clip name known to this implementation.
func play_clip(clip_name: String) -> void:
	push_error("AGSAnimationPlayerBase.play_clip: not implemented (clip='%s')" % clip_name)


## Stop playback and freeze on the current frame.
func stop() -> void:
	push_error("AGSAnimationPlayerBase.stop: not implemented")


## Switch to a named state ("idle", "walk", "talk").
## The implementation maps the state to the appropriate clip and plays it.
func set_state(state: String) -> void:
	push_error("AGSAnimationPlayerBase.set_state: not implemented (state='%s')" % state)


## Called by the AnimationPlayer method-call track (T-BL16) when a tagged frame fires.
## [param label] format: "sound:<name>", "signal:<name>", or "script:<fn>".
func on_anim_event(label: String) -> void:
	pass  # Default: ignore unhandled events; subclasses may override.
