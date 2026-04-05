## UT-M10-70..79 — Save / Load game state (T-GS16)
##
## Tests cover:
##   - game_saved() returns false for non-existent slot
##   - save_game() creates the file; game_saved() then returns true
##   - load_game() restores global variables
##   - load_game() restores character inventory
##   - load_game() restores room item visibility
##   - load_game() emits load_game_requested signal with the saved data
##   - Saving to different slots is independent
##
## AGSRuntime is an engine singleton — never call AGSRuntime.new().
## Each test cleans up its save file in user:// after running.
extends "res://utils/test_base.gd"

const SLOT_70 := 70
const SLOT_71 := 71
const SLOT_72 := 72
const SLOT_73 := 73
const SLOT_74 := 74
const SLOT_75 := 75
const SLOT_76 := 76

func suite_name() -> String:
	return "M10: SaveLoad"


func _cleanup_slot(slot: int) -> void:
	var path := "user://save_%d.json" % slot
	if FileAccess.file_exists(path):
		DirAccess.remove_absolute(path)


# ── game_saved() ──────────────────────────────────────────────────────────────

# UT-M10-70: game_saved returns false for a slot that has never been saved.
func test_70_game_saved_false_for_new_slot() -> void:
	_cleanup_slot(SLOT_70)
	assert_false(AGSRuntime.game_saved(SLOT_70),
			"game_saved should be false for slot %d before any save" % SLOT_70)


# UT-M10-71: save_game creates the file; game_saved then returns true.
func test_71_save_creates_file() -> void:
	_cleanup_slot(SLOT_71)
	AGSRuntime.save_game(SLOT_71)
	assert_true(AGSRuntime.game_saved(SLOT_71),
			"game_saved should be true after save_game")
	_cleanup_slot(SLOT_71)


# ── Globals round-trip ────────────────────────────────────────────────────────

# UT-M10-72: save_game + load_game restores global variables.
func test_72_globals_round_trip() -> void:
	_cleanup_slot(SLOT_72)
	AGSRuntime.set_global("t72_score", 99)
	AGSRuntime.set_global("t72_flag", true)

	AGSRuntime.save_game(SLOT_72)

	# Overwrite with different values so we know load actually restored them.
	AGSRuntime.set_global("t72_score", 0)
	AGSRuntime.set_global("t72_flag", false)

	AGSRuntime.load_game(SLOT_72)

	assert_eq(AGSRuntime.get_global("t72_score"), 99,
			"load_game should restore t72_score = 99")
	assert_eq(AGSRuntime.get_global("t72_flag"), true,
			"load_game should restore t72_flag = true")
	_cleanup_slot(SLOT_72)


# ── Character inventory round-trip ────────────────────────────────────────────

# UT-M10-73: save_game + load_game restores a character's inventory.
func test_73_inventory_round_trip() -> void:
	_cleanup_slot(SLOT_73)

	# Create a character with a GDScript extending AGSCharacter.
	var char_script := load("res://m10_game_systems/runtime/ags_character.gd") as GDScript
	var char_node: Node = char_script.new()
	char_node.character_name = "t73_hero"
	add_to_tree(char_node)

	char_node.add_inventory("rusty_key")
	char_node.add_inventory("torch")

	AGSRuntime.save_game(SLOT_73)

	# Clear inventory so we can verify load restores it.
	char_node.lose_inventory("rusty_key")
	char_node.lose_inventory("torch")
	assert_false(char_node.has_inventory("rusty_key"), "pre-load: rusty_key should be gone")

	AGSRuntime.load_game(SLOT_73)

	assert_true(char_node.has_inventory("rusty_key"),
			"load_game should restore rusty_key in inventory")
	assert_true(char_node.has_inventory("torch"),
			"load_game should restore torch in inventory")

	char_node.notification(Node.NOTIFICATION_EXIT_TREE)
	char_node.queue_free()
	_cleanup_slot(SLOT_73)


# ── Room item visibility round-trip ───────────────────────────────────────────

# UT-M10-74: save_game + load_game restores a room item's visibility.
func test_74_room_item_visibility_round_trip() -> void:
	_cleanup_slot(SLOT_74)

	var ri := AGSRoomItem.new()
	ri.item_name = "t74_chest"
	add_to_tree(ri)

	# Hide the item, then save.
	AGSRuntime.hide_room_item("t74_chest")
	assert_false(ri.visible, "pre-save: chest should be hidden")

	AGSRuntime.save_game(SLOT_74)

	# Show it again so we can verify load hides it.
	AGSRuntime.show_room_item("t74_chest")
	assert_true(ri.visible, "pre-load: chest should be visible again")

	AGSRuntime.load_game(SLOT_74)

	assert_false(ri.visible,
			"load_game should restore chest to hidden")

	ri.notification(Node.NOTIFICATION_EXIT_TREE)
	ri.queue_free()
	_cleanup_slot(SLOT_74)


# ── load_game_requested signal ────────────────────────────────────────────────

# UT-M10-75: load_game emits load_game_requested with the saved dictionary.
func test_75_load_game_requested_signal() -> void:
	_cleanup_slot(SLOT_75)
	AGSRuntime.set_global("t75_key", "hello")
	AGSRuntime.save_game(SLOT_75)

	var fired := [false]
	var got_data := [{}]
	AGSRuntime.load_game_requested.connect(func(data: Dictionary) -> void:
		fired[0] = true
		got_data[0] = data
	, CONNECT_ONE_SHOT)

	AGSRuntime.load_game(SLOT_75)

	assert_true(fired[0], "load_game_requested should fire")
	assert_true(got_data[0].has("globals"), "data should have globals key")
	_cleanup_slot(SLOT_75)


# ── Slot independence ─────────────────────────────────────────────────────────

# UT-M10-76: saves to different slots do not interfere.
func test_76_independent_slots() -> void:
	_cleanup_slot(76)
	_cleanup_slot(77)

	AGSRuntime.set_global("t76_val", 1)
	AGSRuntime.save_game(76)

	AGSRuntime.set_global("t76_val", 2)
	AGSRuntime.save_game(77)

	# Load slot 76 and verify it restores 1.
	AGSRuntime.load_game(76)
	assert_eq(AGSRuntime.get_global("t76_val"), 1,
			"slot 76 should restore value 1")

	# Load slot 77 and verify it restores 2.
	AGSRuntime.load_game(77)
	assert_eq(AGSRuntime.get_global("t76_val"), 2,
			"slot 77 should restore value 2")

	_cleanup_slot(76)
	_cleanup_slot(77)
