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

## Register test suites here as they're implemented.
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
		var result: Dictionary = suite.run_suite()
		reporter.record(result)

	reporter.print_summary()
	quit(0 if reporter.all_passed() else 1)
