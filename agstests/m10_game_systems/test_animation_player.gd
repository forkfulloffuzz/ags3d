## UT-M10-80..87 — AGSAnimationPlayerBase + AGSAnimationPlayer3D (T-GS28)
##
## Tests cover the abstract base API contract and the concrete 3D implementation.
## The 3D implementation wraps an AnimationPlayer — tests use a real AnimationPlayer
## with a simple animation clip added at runtime.
extends "res://utils/test_base.gd"

const BASE_SCRIPT := "res://m10_game_systems/runtime/ags_animation_player_base.gd"
const PLAYER3D_SCRIPT := "res://m10_game_systems/runtime/ags_animation_player_3d.gd"

func suite_name() -> String:
	return "M10: AnimationPlayer"


# ── Base class API ────────────────────────────────────────────────────────────

# UT-M10-80: AGSAnimationPlayerBase can be instantiated.
func test_80_base_instantiates() -> void:
	var script := load(BASE_SCRIPT) as GDScript
	var node: Node = script.new()
	assert_not_null(node, "AGSAnimationPlayerBase.new() returned null")
	node.free()


# UT-M10-81: AGSAnimationPlayerBase exposes play_clip, stop, set_state, on_anim_event.
func test_81_base_has_required_methods() -> void:
	var script := load(BASE_SCRIPT) as GDScript
	var node: Node = script.new()
	assert_true(node.has_method("play_clip"),      "base must have play_clip()")
	assert_true(node.has_method("stop"),           "base must have stop()")
	assert_true(node.has_method("set_state"),      "base must have set_state()")
	assert_true(node.has_method("on_anim_event"),  "base must have on_anim_event()")
	node.free()


# ── AGSAnimationPlayer3D ──────────────────────────────────────────────────────

func _make_player3d_with_anim() -> Array:
	# Returns [parent_node, player3d_node, anim_player_node].
	# parent is added to tree so _ready() runs on both nodes.
	var parent := Node3D.new()
	add_to_tree(parent)

	var anim_player := AnimationPlayer.new()
	anim_player.name = "AnimationPlayer"
	parent.add_child(anim_player)

	# Add a minimal Idle animation with one key.
	var anim := Animation.new()
	anim.length = 1.0
	var anim_lib := AnimationLibrary.new()
	anim_lib.add_animation("Idle", anim)
	anim_player.add_animation_library("", anim_lib)

	var script := load(PLAYER3D_SCRIPT) as GDScript
	var player3d: Node = script.new()
	player3d.animation_player_path = NodePath("../AnimationPlayer")
	parent.add_child(player3d)

	return [parent, player3d, anim_player]


# UT-M10-82: AGSAnimationPlayer3D instantiates.
func test_82_player3d_instantiates() -> void:
	var script := load(PLAYER3D_SCRIPT) as GDScript
	var node: Node = script.new()
	assert_not_null(node, "AGSAnimationPlayer3D.new() returned null")
	node.free()


# UT-M10-83: AGSAnimationPlayer3D finds the AnimationPlayer via animation_player_path.
func test_83_player3d_finds_animation_player() -> void:
	var parts := _make_player3d_with_anim()
	var parent: Node = parts[0]
	var player3d: Node = parts[1]
	var anim_player: AnimationPlayer = parts[2]

	# _ready() should have resolved the path.
	assert_not_null(player3d.get("_anim_player"),
			"AGSAnimationPlayer3D._anim_player should not be null after _ready()")

	parent.free()


# UT-M10-84: play_clip starts a known animation.
func test_84_play_clip_plays_animation() -> void:
	var parts := _make_player3d_with_anim()
	var parent: Node = parts[0]
	var player3d: Node = parts[1]
	var anim_player: AnimationPlayer = parts[2]

	player3d.play_clip("Idle")
	assert_true(anim_player.is_playing(), "AnimationPlayer should be playing after play_clip('Idle')")

	parent.free()


# UT-M10-85: stop() halts playback.
func test_85_stop_halts_playback() -> void:
	var parts := _make_player3d_with_anim()
	var parent: Node = parts[0]
	var player3d: Node = parts[1]
	var anim_player: AnimationPlayer = parts[2]

	player3d.play_clip("Idle")
	player3d.stop()
	assert_false(anim_player.is_playing(), "AnimationPlayer should not be playing after stop()")

	parent.free()


# UT-M10-86: set_state("idle") maps to the Idle clip.
func test_86_set_state_idle_plays_idle_clip() -> void:
	var parts := _make_player3d_with_anim()
	var parent: Node = parts[0]
	var player3d: Node = parts[1]
	var anim_player: AnimationPlayer = parts[2]

	player3d.set_state("idle")
	assert_true(anim_player.is_playing(),
			"AnimationPlayer should be playing after set_state('idle')")
	assert_eq(anim_player.current_animation, "Idle",
			"current_animation should be 'Idle' after set_state('idle')")

	parent.free()


# UT-M10-87: set_state called twice with the same state does not restart the clip.
func test_87_set_state_duplicate_is_noop() -> void:
	var parts := _make_player3d_with_anim()
	var parent: Node = parts[0]
	var player3d: Node = parts[1]
	var anim_player: AnimationPlayer = parts[2]

	player3d.set_state("idle")
	# Seek to mid-clip so we can detect a restart.
	anim_player.seek(0.5, true)
	var pos_before: float = anim_player.current_animation_position

	player3d.set_state("idle")  # duplicate call
	var pos_after: float = anim_player.current_animation_position

	assert_eq(pos_before, pos_after,
			"set_state with same state should not restart the clip")

	parent.free()
