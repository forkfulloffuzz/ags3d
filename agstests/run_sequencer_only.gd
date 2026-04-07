#!/usr/bin/env -S godot --headless --script
extends SceneTree

const Reporter = preload("res://utils/reporter.gd")
const C_BOLD := "\u001b[1m"
const C_RESET := "\u001b[0m"

const ASYNC_SUITES: Array[String] = [
	"m_cut/test_sequencer.gd",
	"m_cut/test_sequencer_sync.gd",
	"m_cut/test_sequencer_timeout.gd",
	"m_cut/test_sequencer_fallback.gd",
]

func _init() -> void:
	call_deferred("_run_tests")

func _run_tests() -> void:
	var reporter := Reporter.new()
	print("")
	print("%sAGS3D Sequencer Tests%s" % [C_BOLD, C_RESET])
	print("-".repeat(50))
	for suite_path in ASYNC_SUITES:
		var script := load("res://" + suite_path) as GDScript
		var suite: Object = script.new()
		suite._tree = self
		var result: Dictionary = await suite.run_suite_async()
		reporter.record(result)
	reporter.print_summary()
	quit(0 if reporter.all_passed() else 1)
