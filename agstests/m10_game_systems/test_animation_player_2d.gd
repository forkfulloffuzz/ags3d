## UT-M10-100..107 — AGSAnimationPlayer2D (T-GS29)
##
## Tests verify the state→slot mapping, controller property updates,
## stop/resume behaviour, and the STATE_SLOTS constant.
extends "res://utils/test_base.gd"

const PLAYER2D_SCRIPT := "res://m10_game_systems/runtime/ags_animation_player_2d.gd"
const CTRL_SCRIPT     := "res://m10_game_systems/runtime/ags_billboard_controller.gd"
const BASE_SCRIPT     := "res://m10_game_systems/runtime/ags_animation_player_base.gd"

func suite_name() -> String:
	return "M10: AnimationPlayer2D"


# ── Helpers ───────────────────────────────────────────────────────────────────

# Build a minimal tree: parent Node → controller + player2d.
# Returns [parent, controller, player2d].
func _make_pair() -> Array:
	var parent := Node.new()

	var ctrl_script := load(CTRL_SCRIPT) as GDScript
	var ctrl: Node = ctrl_script.new()
	ctrl.name = "AGSBillboardController"
	ctrl.sprite_angles = 1
	ctrl.frames_per_angle = 1
	ctrl.fps = 8.0

	var p2d_script := load(PLAYER2D_SCRIPT) as GDScript
	var p2d: Node = p2d_script.new()
	p2d.frames_per_state = 4
	p2d.fps = 12.0
	# Default controller_path "../AGSBillboardController" navigates to sibling — keep it.

	parent.add_child(ctrl)
	parent.add_child(p2d)

	# Manually trigger _ready() so controller_path is resolved.
	p2d.notification(Node.NOTIFICATION_READY)

	return [parent, ctrl, p2d]


# ── Instantiation ─────────────────────────────────────────────────────────────

# UT-M10-100: AGSAnimationPlayer2D instantiates without crash.
func test_100_instantiates() -> void:
	var script := load(PLAYER2D_SCRIPT) as GDScript
	assert_not_null(script, "Failed to load ags_animation_player_2d.gd")
	var node: Node = script.new()
	assert_not_null(node, "script.new() returned null")
	assert_eq(node.frames_per_state, 1, "frames_per_state defaults to 1")
	assert_eq(node.fps, 8.0, "fps defaults to 8.0")
	node.free()


# UT-M10-101: STATE_SLOTS constant has idle=0, walk=1, talk=2.
func test_101_state_slots_constant() -> void:
	var script := load(PLAYER2D_SCRIPT) as GDScript
	var node: Node = script.new()
	assert_eq(node.STATE_SLOTS["idle"], 0, "idle slot = 0")
	assert_eq(node.STATE_SLOTS["walk"], 1, "walk slot = 1")
	assert_eq(node.STATE_SLOTS["talk"], 2, "talk slot = 2")
	node.free()


# ── State → controller property mapping ──────────────────────────────────────

# UT-M10-102: set_state("idle") sets controller frame_offset = 0.
func test_102_idle_sets_offset_0() -> void:
	var arr := _make_pair()
	var ctrl = arr[1]
	var p2d = arr[2]

	p2d.set_state("walk")  # change to walk first
	p2d.set_state("idle")
	assert_eq(ctrl.frame_offset, 0, "idle → frame_offset = 0")
	assert_eq(ctrl.frames_per_angle, p2d.frames_per_state, "frames_per_angle = frames_per_state")
	assert_eq(ctrl.fps, p2d.fps, "controller fps updated")

	arr[0].free()


# UT-M10-103: set_state("walk") sets frame_offset = frames_per_state.
func test_103_walk_sets_offset() -> void:
	var arr := _make_pair()
	var ctrl = arr[1]
	var p2d = arr[2]

	p2d.set_state("walk")
	assert_eq(ctrl.frame_offset, p2d.frames_per_state, "walk → frame_offset = frames_per_state")

	arr[0].free()


# UT-M10-104: set_state("talk") sets frame_offset = 2 * frames_per_state.
func test_104_talk_sets_offset() -> void:
	var arr := _make_pair()
	var ctrl = arr[1]
	var p2d = arr[2]

	p2d.set_state("talk")
	assert_eq(ctrl.frame_offset, 2 * p2d.frames_per_state, "talk → frame_offset = 2*fps")

	arr[0].free()


# ── Idempotent set_state ──────────────────────────────────────────────────────

# UT-M10-105: set_state with same state does not reset controller frame counter.
func test_105_same_state_noop() -> void:
	var arr := _make_pair()
	var ctrl = arr[1]
	var p2d = arr[2]

	p2d.set_state("walk")
	ctrl._current_anim_frame = 3  # simulate mid-animation
	p2d.set_state("walk")         # same state — should not reset
	assert_eq(ctrl._current_anim_frame, 3, "same state call does not reset frame counter")

	arr[0].free()


# ── stop / resume ─────────────────────────────────────────────────────────────

# UT-M10-106: stop() sets controller fps to 0.
func test_106_stop_freezes_fps() -> void:
	var arr := _make_pair()
	var ctrl = arr[1]
	var p2d = arr[2]

	p2d.set_state("walk")
	p2d.stop()
	assert_eq(ctrl.fps, 0.0, "stop() sets controller fps to 0")

	arr[0].free()


# UT-M10-107: After stop, set_state resumes playback.
func test_107_set_state_after_stop_resumes() -> void:
	var arr := _make_pair()
	var ctrl = arr[1]
	var p2d = arr[2]

	p2d.set_state("walk")
	p2d.stop()
	p2d.set_state("walk")  # same state, but was stopped → should resume
	assert_eq(ctrl.fps, p2d.fps, "set_state after stop resumes fps")

	arr[0].free()
