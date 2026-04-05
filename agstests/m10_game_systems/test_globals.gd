## UT-M10-20..24 — AGSRuntime global variable store (T-GS08)
##
## Tests get_global / set_global / init_globals on the AGSRuntime singleton.
## All tests are synchronous — no SceneTree timers needed.
##
## Uses unique per-test key prefixes (t20_, t21_, …) so tests are isolated
## even though they share the singleton's global Dictionary.
extends "res://utils/test_base.gd"

func suite_name() -> String:
	return "M10: Globals"


# UT-M10-20: set_global stores a value; get_global retrieves it.
func test_20_set_and_get_global() -> void:
	AGSRuntime.set_global("t20_score", 42)
	assert_eq(AGSRuntime.get_global("t20_score"), 42,
			"get_global should return value set by set_global")


# UT-M10-21: get_global returns null for an unknown key (not a hard error).
func test_21_get_global_unknown_returns_null() -> void:
	var v = AGSRuntime.get_global("t21_nonexistent")
	assert_true(v == null, "get_global for unknown key should return null")


# UT-M10-22: init_globals sets default values for all provided keys.
func test_22_init_globals_sets_defaults() -> void:
	AGSRuntime.init_globals({"t22_score": 0, "t22_door_unlocked": false, "t22_player_name": ""})
	assert_eq(AGSRuntime.get_global("t22_score"), 0, "score default should be 0")
	assert_eq(AGSRuntime.get_global("t22_door_unlocked"), false, "door_unlocked default should be false")
	assert_eq(AGSRuntime.get_global("t22_player_name"), "", "player_name default should be empty string")


# UT-M10-23: init_globals does not overwrite values already set.
func test_23_init_globals_no_overwrite() -> void:
	AGSRuntime.set_global("t23_score", 100)
	AGSRuntime.init_globals({"t23_score": 0})
	assert_eq(AGSRuntime.get_global("t23_score"), 100,
			"init_globals should not overwrite existing value")


# UT-M10-24: set_global can update a value multiple times.
func test_24_set_global_overwrite() -> void:
	AGSRuntime.set_global("t24_score", 10)
	AGSRuntime.set_global("t24_score", 20)
	assert_eq(AGSRuntime.get_global("t24_score"), 20,
			"set_global should overwrite previous value")
