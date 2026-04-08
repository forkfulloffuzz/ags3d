#!/usr/bin/env -S godot --headless --script
## AGS3D master test runner.
##
## Usage (from repo root):
##   ./bin/godot.linuxbsd.editor.x86_64 --headless --path agstests --script run_tests.gd
##
## Exit code 0 = all tests passed. Exit code 1 = at least one failure.

extends SceneTree

const Reporter = preload("res://utils/reporter.gd")

const C_RED   := "\u001b[31m"
const C_BOLD  := "\u001b[1m"
const C_RESET := "\u001b[0m"

## Synchronous test suites — each test_* method runs to completion in one call.
const SUITES: Array[String] = [
	"m1_module/test_script_language.gd",
	"m4_room/test_room_node.gd",
	"m4_room/test_walkable.gd",
	"m4_room/test_blocker.gd",
	"m4_room/test_points.gd",
	"m4_room/test_regions.gd",
	"m4_room/test_hotspots.gd",
	"m5_character/test_character_node.gd",
	"m5_character/test_navigation.gd",
	"m5_character/test_faceto.gd",
	"m5_character/test_spawnpoint.gd",
	"m6_bindings/test_event_binding.gd",
	"m6_bindings/test_source_map.gd",
	"m6_integration/test_script_wiring.gd",
	"m6_integration/test_runtime_api.gd",
	"m6_integration/test_event_routing.gd",
	"m10_game_systems/test_load_room.gd",  # T-GS10
	"m10_game_systems/test_globals.gd",    # T-GS08
	"m10_game_systems/test_item.gd",       # T-GS02
	"m10_game_systems/test_room_item.gd",  # T-GS03
	"m10_game_systems/test_audio.gd",     # T-GS12
	"m10_game_systems/test_save_load.gd",      # T-GS16
	"m10_game_systems/test_animation_player.gd", # T-GS28
	"m10_game_systems/test_billboard.gd",        # T-GS26
	"m10_game_systems/test_animation_player_2d.gd", # T-GS29
	"m10_game_systems/test_gui_runtime.gd",         # T-GS14
	"m_cut/test_event_bus.gd",                      # T-CUT10
]

## Async test suites — each test_* method is awaited so coroutine-based tests
## (multi-frame physics simulations) run to completion before the next test starts.
const ASYNC_SUITES: Array[String] = [
	"m_cut/test_event_bus_surface.gd",              # T-CUT11
	"m_cut/test_sequencer.gd",                      # T-CUT12
	"m_cut/test_sequencer_sync.gd",                 # T-CUT13
	"m_cut/test_sequencer_timeout.gd",              # T-CUT14
	"m_cut/test_sequencer_fallback.gd",             # T-CUT15
	"m_cut/test_char_commands.gd",                  # T-CUT17
	"m_cut/test_camera_commands.gd",               # T-CUT16
	"m_cut/test_audio_commands.gd",               # T-CUT18
	"m_cut/test_visual_commands.gd",              # T-CUT19
	"m_cut/test_flow_commands.gd",               # T-CUT20
	"m_cut/test_dialogue_commands.gd",          # T-CUT21
	"m_cut/test_skip_input.gd",                 # T-CUT22
	"m_cut/test_skip_system.gd",                # T-CUT23
	"m_dlg/test_dialogue_engine.gd",           # T-DLG14
	"m_dlg/test_localisation.gd",              # T-DLG17
	"m_dlg/test_dialogue_state.gd",            # T-DLG15
	"m_dlg/test_dialogue_ui.gd",               # T-DLG16
	"m6_integration/test_end_to_end.gd",
	"m10_game_systems/test_say.gd",       # T-GS01: say() / think() use SceneTree timers
	"m10_game_systems/test_cutscene.gd",  # T-GS18 (async subset: wait, fade_in, fade_out)
	"m10_game_systems/test_char_animation.gd",  # T-BL13 (async subset: say() drives anim)
]

func _init() -> void:
	# Defer test execution so SceneTree.initialize() has time to run first.
	# By the time _run_tests() is called, root.is_inside_tree() == true, which
	# allows add_to_tree() to properly propagate NOTIFICATION_ENTER_TREE to
	# nodes that require a live scene tree (e.g. CharacterBody3D, NavigationAgent3D).
	call_deferred("_run_tests")

func _run_tests() -> void:
	var reporter := Reporter.new()

	print("")
	print("%sAGS3D Test Suite%s" % [C_BOLD, C_RESET])
	print("-".repeat(50))

	for suite_path in SUITES:
		var script := load("res://" + suite_path) as GDScript
		if script == null:
			print("%s[ERROR]%s Could not load suite: %s" % [C_RED, C_RESET, suite_path])
			reporter.record({
				"suite": suite_path,
				"pass": 0,
				"fail": 1,
				"failures": ["  FAIL Failed to load test file: " + suite_path],
			})
			continue

		var suite: Object = script.new()
		if not suite.has_method("run_suite"):
			print("%s[ERROR]%s %s does not extend TestBase" % [C_RED, C_RESET, suite_path])
			continue

		suite._tree = self
		print("")
		print("── %s" % suite_path)
		var result: Dictionary = suite.run_suite()
		reporter.record(result)

	for suite_path in ASYNC_SUITES:
		var script := load("res://" + suite_path) as GDScript
		if script == null:
			print("%s[ERROR]%s Could not load async suite: %s" % [C_RED, C_RESET, suite_path])
			reporter.record({
				"suite": suite_path,
				"pass": 0,
				"fail": 1,
				"failures": ["  FAIL Failed to load test file: " + suite_path],
			})
			continue

		var suite: Object = script.new()
		if not suite.has_method("run_suite_async"):
			print("%s[ERROR]%s %s does not extend TestBase" % [C_RED, C_RESET, suite_path])
			continue

		suite._tree = self
		print("")
		print("── %s" % suite_path)
		var result: Dictionary = await suite.run_suite_async()
		reporter.record(result)

	reporter.print_summary()
	quit(0 if reporter.all_passed() else 1)
