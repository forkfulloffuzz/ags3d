#!/usr/bin/env -S godot --headless --script
## AGS3D master test runner.
##
## Usage (from repo root):
##   ./bin/godot.linuxbsd.editor.x86_64 --headless --path agstests --script run_tests.gd
##
## Exit code 0 = all tests passed. Exit code 1 = at least one failure.

extends SceneTree

const Reporter = preload("res://utils/reporter.gd")

## Register test suites here as they're implemented.
const SUITES: Array[String] = [
	"m1_module/test_script_language.gd",
]

func _init() -> void:
	var reporter := Reporter.new()

	print("")
	print("AGS3D Test Suite")
	print("-".repeat(50))

	for suite_path in SUITES:
		var script := load("res://" + suite_path) as GDScript
		if script == null:
			print("[ERROR] Could not load suite: %s" % suite_path)
			reporter.record({
				"suite": suite_path,
				"pass": 0,
				"fail": 1,
				"failures": ["  FAIL Failed to load test file: " + suite_path],
			})
			continue

		var suite: Object = script.new()
		if not suite.has_method("run_suite"):
			print("[ERROR] %s does not extend TestBase" % suite_path)
			continue

		var result: Dictionary = suite.run_suite()
		reporter.record(result)

	reporter.print_summary()
	quit(0 if reporter.all_passed() else 1)
