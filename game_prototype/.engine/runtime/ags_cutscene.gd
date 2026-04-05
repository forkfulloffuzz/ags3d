## ags_cutscene.gd — Cutscene helpers for AGS3D (T-GS18)
##
## Add as an AutoLoad named AGSCutscene so room scripts can call:
##   await AGSCutscene.fade_out()
##   await AGSCutscene.fade_in()
##   await AGSCutscene.wait(2.0)
##   AGSRuntime.set_player_control(false)
##
## The fade overlay is a full-screen black ColorRect in a high-priority
## CanvasLayer (layer 100) so it draws on top of the scene.
extends Node

var _overlay: ColorRect = null


func _ready() -> void:
	var layer := CanvasLayer.new()
	layer.layer = 100
	add_child(layer)

	_overlay = ColorRect.new()
	_overlay.color = Color.BLACK
	_overlay.visible = false
	_overlay.mouse_filter = Control.MOUSE_FILTER_IGNORE
	layer.add_child(_overlay)
	# Full-rect preset requires a connected viewport — skip in headless mode.
	if _overlay.get_viewport() != null:
		_overlay.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)


## Fade the screen to black over [param duration] seconds (blocking).
func fade_out(duration: float = 0.5) -> void:
	_overlay.color = Color(0.0, 0.0, 0.0, 0.0)
	_overlay.visible = true
	var tween := create_tween()
	tween.tween_property(_overlay, "color:a", 1.0, duration)
	await tween.finished


## Fade the screen from black to clear over [param duration] seconds (blocking).
func fade_in(duration: float = 0.5) -> void:
	_overlay.color = Color(0.0, 0.0, 0.0, 1.0)
	_overlay.visible = true
	var tween := create_tween()
	tween.tween_property(_overlay, "color:a", 0.0, duration)
	await tween.finished
	_overlay.visible = false


## Pause execution for [param seconds] seconds (blocking).
func wait(seconds: float) -> void:
	await get_tree().create_timer(seconds).timeout
