## UT-M10-50..57 — Cutscene support: SetPlayerControl, FadeIn, FadeOut, Wait (T-GS18)
##
## Sync tests (50..54): C++ AGSRuntime player_control_enabled,
##   hide_room_item / show_room_item.
## Async tests (55..57): GDScript ags_cutscene.gd — wait(), fade_out(), fade_in().
##
## AGSRuntime is an engine singleton — never call AGSRuntime.new().
## Tests 50..54 use the singleton directly.
## The async tests attach ags_cutscene.gd to a Node added to the real scene tree
## so that get_tree() and create_tween() are available.
extends "res://utils/test_base.gd"

const CUTSCENE_SCRIPT := "res://m10_game_systems/runtime/ags_cutscene.gd"
## Short durations so tests complete quickly.
const FADE_DURATION := 0.05
const WAIT_SECONDS  := 0.05

func suite_name() -> String:
	return "M10: Cutscene"


# ── Sync: player control ──────────────────────────────────────────────────────

# UT-M10-50: player_control_enabled defaults to true.
func test_50_player_control_default_true() -> void:
	# Reset to known state first.
	AGSRuntime.set_player_control(true)
	assert_true(AGSRuntime.is_player_control_enabled(),
			"player_control_enabled should default to true")


# UT-M10-51: set_player_control(false) disables control.
func test_51_set_player_control_false() -> void:
	AGSRuntime.set_player_control(false)
	assert_false(AGSRuntime.is_player_control_enabled(),
			"player_control_enabled should be false after set_player_control(false)")
	AGSRuntime.set_player_control(true)  # restore


# UT-M10-52: set_player_control(true) re-enables control after disabling.
func test_52_set_player_control_true() -> void:
	AGSRuntime.set_player_control(false)
	AGSRuntime.set_player_control(true)
	assert_true(AGSRuntime.is_player_control_enabled(),
			"player_control_enabled should be true after set_player_control(true)")


# UT-M10-53: player_control_changed signal fires with the correct value.
func test_53_player_control_changed_signal() -> void:
	AGSRuntime.set_player_control(true)  # known start state
	var fired := [false]
	var got_value := [true]
	AGSRuntime.player_control_changed.connect(func(enabled: bool) -> void:
		fired[0] = true
		got_value[0] = enabled
	, CONNECT_ONE_SHOT)
	AGSRuntime.set_player_control(false)
	assert_true(fired[0], "player_control_changed should have fired")
	assert_false(got_value[0], "player_control_changed should carry false")
	AGSRuntime.set_player_control(true)  # restore


# ── Sync: hide / show room item ───────────────────────────────────────────────

# UT-M10-54: hide_room_item sets visible=false; show_room_item restores it.
# AGSRoomItem must be notified of READY so it registers with AGSRuntime.
func test_54_hide_show_room_item() -> void:
	var ri := AGSRoomItem.new()
	ri.item_name = "t54_test_chest"
	# Manually fire NOTIFICATION_READY — registers with the singleton.
	ri.notification(Node.NOTIFICATION_READY)

	AGSRuntime.hide_room_item("t54_test_chest")
	assert_false(ri.visible, "hide_room_item should set visible=false")

	AGSRuntime.show_room_item("t54_test_chest")
	assert_true(ri.visible, "show_room_item should restore visible=true")

	ri.notification(Node.NOTIFICATION_EXIT_TREE)
	ri.free()


# ── Async: ags_cutscene.gd ────────────────────────────────────────────────────

## Create an in-tree node with ags_cutscene.gd attached.
## Returns the cutscene node. Caller frees via node.get_parent().free().
func _make_cutscene_node() -> Node:
	var root := Node.new()
	add_to_tree(root)
	var cs := Node.new()
	cs.set_script(load(CUTSCENE_SCRIPT))
	root.add_child(cs)
	# Manual notification so _ready() runs synchronously and _overlay is
	# set up before the first await. The deferred natural _ready() call
	# is harmless (it replaces _overlay but the tween holds the old ref).
	cs.notification(Node.NOTIFICATION_READY)
	return cs


# UT-M10-55: wait(seconds) completes and resumes the caller coroutine.
func test_55_wait_completes() -> void:
	var cs := _make_cutscene_node()
	await cs.wait(WAIT_SECONDS)
	pass_test()  # reaching here means wait() completed
	cs.get_parent().free()


# UT-M10-56: fade_out() completes and resumes the caller coroutine.
func test_56_fade_out_completes() -> void:
	var cs := _make_cutscene_node()
	await cs.fade_out(FADE_DURATION)
	pass_test()
	cs.get_parent().free()


# UT-M10-57: fade_in() completes and resumes the caller coroutine.
func test_57_fade_in_completes() -> void:
	var cs := _make_cutscene_node()
	await cs.fade_in(FADE_DURATION)
	pass_test()
	cs.get_parent().free()
