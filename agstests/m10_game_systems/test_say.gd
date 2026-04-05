## UT-M10-10..14 — AGSCharacter say_text / say_completed / say() / think() (T-GS01)
##
## Split into two groups:
##   UT-M10-10..11  synchronous — test C++ property and signal registration only.
##   UT-M10-12..14  async — test the GDScript say() / think() runtime methods.
##
## The async tests attach the runtime script from:
##   res://m10_game_systems/runtime/ags_character.gd
## (a verbatim copy of game_prototype/.engine/runtime/ags_character.gd).
##
## All async tests use a short duration (0.05 s) so they complete quickly.
##
## NOTE: lambdas capture primitives (bool, int) by value in GDScript 4.
## All captured signal flags must use Array wrappers — same pattern as
## test_end_to_end.gd's `var fired := [false]`.
extends "res://utils/test_base.gd"

const CHAR_SCRIPT := "res://m10_game_systems/runtime/ags_character.gd"
## How long say() runs in these tests.
const SAY_DURATION  := 0.05
## How long to wait before asserting — generous headroom over SAY_DURATION.
const SAY_WAIT      := 0.3

func suite_name() -> String:
	return "M10: Say"


# ── Synchronous C++ tests ─────────────────────────────────────────────────────

# UT-M10-10: say_text property is readable and writable on a bare AGSCharacter.
func test_10_say_text_property_readable_writable() -> void:
	var ch := AGSCharacter3D.new()
	assert_eq(ch.say_text, "", "say_text should default to empty string")
	ch.say_text = "Hello world"
	assert_eq(ch.say_text, "Hello world", "say_text did not store the assigned value")
	ch.free()


# UT-M10-11: say_completed signal is registered on AGSCharacter.
func test_11_say_completed_signal_exists() -> void:
	var ch := AGSCharacter3D.new()
	assert_true(ch.has_signal("say_completed"), "AGSCharacter does not have a say_completed signal")
	ch.free()


# ── Async GDScript runtime tests ──────────────────────────────────────────────

## Create an AGSCharacter with the runtime script attached and add it to the tree.
## Caller must free the returned node's parent (get_parent().free()).
func _make_runtime_char(char_name: String) -> AGSCharacterBase:
	var root := Node.new()
	add_to_tree(root)
	var ch := AGSCharacter3D.new()
	ch.set_script(load(CHAR_SCRIPT))
	ch.character_name = char_name
	root.add_child(ch)
	ch.notification(Node.NOTIFICATION_READY)
	return ch


# UT-M10-12: say() sets say_text immediately, clears it, and emits say_completed.
func test_12_say_sets_clears_text_and_emits_signal() -> void:
	var ch := _make_runtime_char("say_char_12")

	var completed := [false]
	ch.say_completed.connect(func() -> void: completed[0] = true, CONNECT_ONE_SHOT)

	ch.say("Hello world", SAY_DURATION)  # fire-and-forget coroutine
	assert_eq(ch.say_text, "Hello world", "say_text not set immediately after say() call")

	await _tree.create_timer(SAY_WAIT).timeout

	assert_true(completed[0], "say_completed did not fire within %.1f s" % SAY_WAIT)
	assert_eq(ch.say_text, "", "say_text was not cleared after say() completed")

	ch.get_parent().free()


# UT-M10-13: think() emits say_completed (delegates to say()).
func test_13_think_emits_say_completed() -> void:
	var ch := _make_runtime_char("say_char_13")

	var completed := [false]
	ch.say_completed.connect(func() -> void: completed[0] = true, CONNECT_ONE_SHOT)

	ch.think("Hmm…", SAY_DURATION)
	await _tree.create_timer(SAY_WAIT).timeout

	assert_true(completed[0], "say_completed did not fire after think()")

	ch.get_parent().free()


# UT-M10-14: say_completed fires only once per say() call.
func test_14_say_completed_fires_once() -> void:
	var ch := _make_runtime_char("say_char_14")

	var count := [0]
	ch.say_completed.connect(func() -> void: count[0] += 1)

	ch.say("Once", SAY_DURATION)
	await _tree.create_timer(SAY_WAIT).timeout

	assert_eq(count[0], 1, "say_completed should fire exactly once, fired %d times" % count[0])

	ch.get_parent().free()  # frees ch and cleans up signal connections
