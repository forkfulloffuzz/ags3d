## T34 — Source map translation tests.
##
## Verifies that AGSRuntime.register_source_map() stores a parsed .agmap and
## that translate_script_error() maps generated GDScript lines back to the
## originating AGS-spirit file and line.
extends "res://utils/test_base.gd"

func suite_name() -> String:
	return "M6: SourceMap"


# T34-01: register_source_map() followed by translate_script_error() returns
# the correct AGS-spirit file and line for an exact match.
func test_01_translate_exact_line() -> void:
	var gd_path := "res://.engine/generated/rooms/start.agscript.gd"
	# [[gd_line, agscript_file, agscript_line], ...]
	AGSRuntime.register_source_map(gd_path, [[2, "rooms/start.agscript", 5]])
	var loc: Dictionary = AGSRuntime.translate_script_error(gd_path, 2)
	assert_eq(loc.get("file"), "rooms/start.agscript",
			"translate_script_error should return the agscript file")
	assert_eq(loc.get("line"), 5,
			"translate_script_error should return the agscript line")


# T34-02: translate_script_error() returns an empty Dictionary for an unknown path.
func test_02_unknown_path_returns_empty() -> void:
	var loc: Dictionary = AGSRuntime.translate_script_error(
			"res://.engine/generated/rooms/unknown.agscript.gd", 1)
	assert_true(loc.is_empty(),
			"translate_script_error for an unknown path should return an empty Dictionary")


# T34-03: translate_script_error() returns the best matching entry when the
# queried line falls between two source map entries (takes the last entry
# whose gd_line <= queried line).
func test_03_translate_between_entries() -> void:
	var gd_path := "res://.engine/generated/rooms/multi.agscript.gd"
	AGSRuntime.register_source_map(gd_path, [
		[1, "rooms/multi.agscript", 10],
		[3, "rooms/multi.agscript", 12],
		[6, "rooms/multi.agscript", 15],
	])
	# Line 4 falls between entries at lines 3 and 6 — should map to the entry at 3.
	var loc: Dictionary = AGSRuntime.translate_script_error(gd_path, 4)
	assert_eq(loc.get("line"), 12,
			"translate_script_error should use the last entry with gd_line <= queried line")


# T34-04: translate_script_error() returns empty when the queried line is
# before all entries in the source map.
func test_04_line_before_all_entries_returns_empty() -> void:
	var gd_path := "res://.engine/generated/rooms/early.agscript.gd"
	AGSRuntime.register_source_map(gd_path, [[5, "rooms/early.agscript", 1]])
	var loc: Dictionary = AGSRuntime.translate_script_error(gd_path, 2)
	assert_true(loc.is_empty(),
			"translate_script_error should return empty when line is before all entries")


# T34-05: A second register_source_map() call for the same path replaces the
# previous map (no stale entries).
func test_05_re_register_replaces_map() -> void:
	var gd_path := "res://.engine/generated/rooms/replace.agscript.gd"
	AGSRuntime.register_source_map(gd_path, [[1, "rooms/replace.agscript", 99]])
	AGSRuntime.register_source_map(gd_path, [[1, "rooms/replace.agscript", 42]])
	var loc: Dictionary = AGSRuntime.translate_script_error(gd_path, 1)
	assert_eq(loc.get("line"), 42,
			"re-registering a source map should replace the old one")
