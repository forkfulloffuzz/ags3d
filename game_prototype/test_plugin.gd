#!/usr/bin/env -S godot --headless --script
## AG Studio plugin test suite — run headlessly against the game_prototype project.
##
## Usage (from repo root):
##   ./bin/godot.linuxbsd.editor.x86_64 --headless --path game_prototype --script test_plugin.gd
##
## Tests:
##   1. All plugin .gd files compile (load() returns non-null).
##   2. RoomSync.write_agroom() guards return correct error codes.
##   3. RoomSync float formatter (_f) produces expected output.
##
## Exit code: 0 = all pass, 1 = any failure.

extends SceneTree

const C_GREEN := "\u001b[32m"
const C_RED   := "\u001b[31m"
const C_BOLD  := "\u001b[1m"
const C_RESET := "\u001b[0m"

var _pass := 0
var _fail := 0
var _failures: Array[String] = []

# All plugin scripts that must compile without errors.
const PLUGIN_SCRIPTS: Array[String] = [
	"res://addons/ag_studio/ag_studio.gd",
	"res://addons/ag_studio/ags_inspector_plugin.gd",
	"res://addons/ag_studio/build_log.gd",
	"res://addons/ag_studio/char_editor.gd",
	"res://addons/ag_studio/project_panel.gd",
	"res://addons/ag_studio/project_wizard.gd",
	"res://addons/ag_studio/room_editor.gd",
	"res://addons/ag_studio/room_sync.gd",
	"res://addons/ag_studio/script_editor.gd",
	"res://addons/ag_studio/gizmos/ags_gizmo_box.gd",
	"res://addons/ag_studio/gizmos/ags_blocker_gizmo.gd",
	"res://addons/ag_studio/gizmos/ags_camera_gizmo.gd",
	"res://addons/ag_studio/gizmos/ags_hotspot_gizmo.gd",
	"res://addons/ag_studio/gizmos/ags_point_gizmo.gd",
	"res://addons/ag_studio/gizmos/ags_spawn_gizmo.gd",
	"res://addons/ag_studio/gizmos/ags_trigger_gizmo.gd",
	"res://addons/ag_studio/gizmos/ags_walkable_gizmo.gd",
]


func _init() -> void:
	call_deferred("_run")


func _run() -> void:
	print("")
	print("%sAG Studio Plugin Tests%s" % [C_BOLD, C_RESET])
	print("-".repeat(50))

	_test_all_scripts_compile()
	_test_room_sync_guards()
	_test_room_sync_float_formatter()

	print("")
	print("-".repeat(50))
	if _fail == 0:
		print("%s[PASS]%s  %d passed, 0 failed" % [C_GREEN, C_RESET, _pass])
	else:
		print("%s[FAIL]%s  %d passed, %d failed" % [C_RED, C_RESET, _pass, _fail])
		for f in _failures:
			print("  %s✗%s %s" % [C_RED, C_RESET, f])
	print("")

	quit(0 if _fail == 0 else 1)


# ── 1. Script compilation ─────────────────────────────────────────────────────

func _test_all_scripts_compile() -> void:
	print("Plugin script compilation:")
	for path in PLUGIN_SCRIPTS:
		var script: GDScript = load(path)
		if script == null:
			_record_fail("compile", path, "load() returned null — script has parse/compile errors")
		else:
			_record_pass("compile", path.get_file())


# ── 2. RoomSync guards ────────────────────────────────────────────────────────

func _test_room_sync_guards() -> void:
	print("RoomSync guards:")
	var sync_script: GDScript = load("res://addons/ag_studio/room_sync.gd")
	if sync_script == null:
		_record_fail("room_sync", "load", "room_sync.gd did not compile — skipping guard tests")
		return

	# 2a: non-AGSRoom node returns ERR_INVALID_PARAMETER
	var plain := Node.new()
	var err_a: int = sync_script.write_agroom(plain)
	if err_a == ERR_INVALID_PARAMETER:
		_record_pass("room_sync", "write_agroom(non-AGSRoom) → ERR_INVALID_PARAMETER")
	else:
		_record_fail("room_sync", "write_agroom(non-AGSRoom)", "expected ERR_INVALID_PARAMETER (%d), got %d" % [ERR_INVALID_PARAMETER, err_a])
	plain.free()

	# 2b: AGSRoom with no scene_file_path returns ERR_FILE_NOT_FOUND
	var room := AGSRoom.new()
	room.room_name = "test_room"
	var err_b: int = sync_script.write_agroom(room)
	if err_b == ERR_FILE_NOT_FOUND:
		_record_pass("room_sync", "write_agroom(AGSRoom, no path) → ERR_FILE_NOT_FOUND")
	else:
		_record_fail("room_sync", "write_agroom(AGSRoom, no path)", "expected ERR_FILE_NOT_FOUND (%d), got %d" % [ERR_FILE_NOT_FOUND, err_b])
	room.free()


# ── 3. RoomSync float formatter ───────────────────────────────────────────────

func _test_room_sync_float_formatter() -> void:
	print("RoomSync _f() formatter:")
	var sync_script: GDScript = load("res://addons/ag_studio/room_sync.gd")
	if sync_script == null:
		_record_fail("room_sync", "_f", "room_sync.gd did not compile — skipping formatter tests")
		return

	# Call the static _f() via the script's static method.
	# We exercise it indirectly via _serialise → _vec3 → _f by checking a known
	# round-trip through write_agroom on a minimal in-memory room scene, but
	# since write_agroom needs a file path we test _f() directly via call().
	var cases: Array = [
		[1.0,   "1.0"],
		[1.5,   "1.5"],
		[1.50,  "1.5"],
		[10.0,  "10.0"],
		[0.1,   "0.1"],
		[3.14,  "3.14"],
		[3.10,  "3.1"],
		[-1.0,  "-1.0"],
		[-0.05, "-0.05"],
	]
	for c in cases:
		var input: float = c[0]
		var want: String  = c[1]
		var got: String   = sync_script._f(input)
		if got == want:
			_record_pass("_f", "_f(%s) = %s" % [input, want])
		else:
			_record_fail("_f", "_f(%s)" % input, "expected '%s', got '%s'" % [want, got])


# ── Helpers ───────────────────────────────────────────────────────────────────

func _record_pass(suite: String, label: String) -> void:
	_pass += 1
	print("  %s✓%s [%s] %s" % [C_GREEN, C_RESET, suite, label])


func _record_fail(suite: String, label: String, msg: String) -> void:
	_fail += 1
	print("  %s✗%s [%s] %s — %s" % [C_RED, C_RESET, suite, label, msg])
	_failures.append("[%s] %s — %s" % [suite, label, msg])
