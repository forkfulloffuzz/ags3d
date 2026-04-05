## UT-M10-90..96 — ags_character.gd AnimationPlayer state driving (T-BL13)
##
## Verifies that ags_character.gd drives an AnimationPlayer child using the
## anim_idle / anim_walk / anim_talk export vars emitted by ag build (T-BL12).
##
## Sync tests: export var presence, no-crash without AnimationPlayer.
## Async tests: state transitions during say() drive the correct clips.
extends "res://utils/test_base.gd"

const CHAR_SCRIPT := "res://m10_game_systems/runtime/ags_character.gd"
const SAY_DURATION := 0.05
const SAY_WAIT     := 0.3

func suite_name() -> String:
	return "M10: CharAnimation (T-BL13)"


# ── Helpers ───────────────────────────────────────────────────────────────────

## Build an AGSCharacter3D with ags_character.gd script, an AnimationPlayer child,
## and three clips (Idle/Walk/Talk) pre-loaded. Returns [char_node, anim_player, root].
## Add the AnimationPlayer before tree entry so _ready() finds it.
func _make_animated_char(char_name: String) -> Array:
	var root := Node.new()
	add_to_tree(root)

	var ch := AGSCharacter3D.new()
	ch.set_script(load(CHAR_SCRIPT))
	ch.character_name = char_name
	ch.anim_idle = "Idle"
	ch.anim_walk = "Walk"
	ch.anim_talk = "Talk"

	var anim_player := AnimationPlayer.new()
	var lib := AnimationLibrary.new()
	for clip_name in ["Idle", "Walk", "Talk"]:
		var anim := Animation.new()
		anim.length = 1.0
		lib.add_animation(clip_name, anim)
	anim_player.add_animation_library("", lib)
	ch.add_child(anim_player)  # must be added before tree entry so _ready() finds it

	root.add_child(ch)  # _ready() fires here; find_child("AnimationPlayer") succeeds
	return [ch, anim_player, root]


# ── UT-M10-90: export vars exist ──────────────────────────────────────────────

func test_90_anim_export_vars_exist() -> void:
	var ch := AGSCharacter3D.new()
	ch.set_script(load(CHAR_SCRIPT))
	assert_true("anim_idle" in ch, "ags_character.gd must export anim_idle")
	assert_true("anim_walk" in ch, "ags_character.gd must export anim_walk")
	assert_true("anim_talk" in ch, "ags_character.gd must export anim_talk")
	ch.free()


# ── UT-M10-91: no crash without AnimationPlayer ───────────────────────────────

func test_91_no_crash_without_anim_player() -> void:
	var root := Node.new()
	add_to_tree(root)
	var ch := AGSCharacter3D.new()
	ch.set_script(load(CHAR_SCRIPT))
	ch.character_name = "bare_char"
	root.add_child(ch)  # _ready() runs; _anim_player stays null

	# Directly call the internal state driver — should not crash
	ch.call("_play_anim_state", "walk")
	ch.call("_play_anim_state", "idle")

	assert_true(true, "no crash when AnimationPlayer absent")
	root.free()


# ── UT-M10-92: idle after _on_navigation_finished ────────────────────────────

func test_92_nav_finished_sets_idle() -> void:
	var parts := _make_animated_char("nav_char_92")
	var ch: Node = parts[0]
	var anim_player: AnimationPlayer = parts[1]
	var root: Node = parts[2]

	# Manually put character in walk state first
	ch.call("_play_anim_state", "walk")
	assert_eq(anim_player.current_animation, "Walk", "pre-condition: should be in Walk")

	# Fire navigation_finished directly
	ch.call("_on_navigation_finished")
	assert_eq(anim_player.current_animation, "Idle",
			"_on_navigation_finished must switch AnimationPlayer to idle clip")

	root.free()


# ── UT-M10-93: duplicate state call is a no-op ───────────────────────────────

func test_93_duplicate_state_noop() -> void:
	var parts := _make_animated_char("dup_char_93")
	var ch: Node = parts[0]
	var anim_player: AnimationPlayer = parts[1]
	var root: Node = parts[2]

	ch.call("_play_anim_state", "idle")
	anim_player.seek(0.5, true)
	var pos_before: float = anim_player.current_animation_position

	ch.call("_play_anim_state", "idle")  # same state again
	var pos_after: float = anim_player.current_animation_position

	assert_eq(pos_before, pos_after, "duplicate state call must not restart the clip")
	root.free()


# ── UT-M10-94 (async): say() drives talk then returns to idle ─────────────────

func test_94_say_drives_talk_then_idle() -> void:
	var parts := _make_animated_char("say_anim_char_94")
	var ch: Node = parts[0]
	var anim_player: AnimationPlayer = parts[1]
	var root: Node = parts[2]

	# Pre-condition: not yet in talk state
	assert_ne(anim_player.current_animation, "Talk", "pre-condition: not talking yet")

	ch.call("say", "Hello", SAY_DURATION)  # fire-and-forget coroutine
	# After first frame the say() coroutine has set say_text and called _play_anim_state("talk")
	assert_eq(anim_player.current_animation, "Talk",
			"AnimationPlayer should play Talk clip immediately after say() starts")

	await _tree.create_timer(SAY_WAIT).timeout

	assert_eq(anim_player.current_animation, "Idle",
			"AnimationPlayer should return to Idle after say() completes")

	root.free()


# ── UT-M10-95: unknown clip name logs a warning, does not crash ───────────────

func test_95_unknown_clip_warns_not_crash() -> void:
	var root := Node.new()
	add_to_tree(root)

	var ch := AGSCharacter3D.new()
	ch.set_script(load(CHAR_SCRIPT))
	ch.character_name = "warn_char_95"
	ch.anim_idle = "DoesNotExist"

	var anim_player := AnimationPlayer.new()
	var lib := AnimationLibrary.new()
	# Only add Walk — Idle is intentionally missing
	var walk_anim := Animation.new(); walk_anim.length = 1.0
	lib.add_animation("Walk", walk_anim)
	anim_player.add_animation_library("", lib)
	ch.add_child(anim_player)
	root.add_child(ch)

	ch.call("_play_anim_state", "idle")  # clip "DoesNotExist" → warning, not crash
	assert_true(true, "no crash when clip name not found in AnimationPlayer")
	root.free()
