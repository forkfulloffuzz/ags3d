## UT-M10-20..24 — AGSRuntime global variable store (T-GS08)
##
## Tests get_global / set_global / init_globals on the AGSRuntime singleton.
## All tests are synchronous — no SceneTree timers needed.
extends "res://utils/test_base.gd"

func suite_name() -> String:
	return "M10: Globals"


# UT-M10-20: set_global stores a value; get_global retrieves it.
func test_20_set_and_get_global() -> void:
	var rt := AGSRuntime.new()
	rt.set_global("score", 42)
	assert_eq(rt.get_global("score"), 42, "get_global should return value set by set_global")
	rt.free()


# UT-M10-21: get_global returns null for an unknown key (not a hard error).
func test_21_get_global_unknown_returns_null() -> void:
	var rt := AGSRuntime.new()
	var v = rt.get_global("nonexistent")
	assert_true(v == null, "get_global for unknown key should return null")
	rt.free()


# UT-M10-22: init_globals sets default values for all provided keys.
func test_22_init_globals_sets_defaults() -> void:
	var rt := AGSRuntime.new()
	rt.init_globals({"score": 0, "door_unlocked": false, "player_name": ""})
	assert_eq(rt.get_global("score"), 0, "score default should be 0")
	assert_eq(rt.get_global("door_unlocked"), false, "door_unlocked default should be false")
	assert_eq(rt.get_global("player_name"), "", "player_name default should be empty string")
	rt.free()


# UT-M10-23: init_globals does not overwrite values already set.
func test_23_init_globals_no_overwrite() -> void:
	var rt := AGSRuntime.new()
	rt.set_global("score", 100)
	rt.init_globals({"score": 0})
	assert_eq(rt.get_global("score"), 100, "init_globals should not overwrite existing value")
	rt.free()


# UT-M10-24: set_global can update a value multiple times.
func test_24_set_global_overwrite() -> void:
	var rt := AGSRuntime.new()
	rt.set_global("score", 10)
	rt.set_global("score", 20)
	assert_eq(rt.get_global("score"), 20, "set_global should overwrite previous value")
	rt.free()
