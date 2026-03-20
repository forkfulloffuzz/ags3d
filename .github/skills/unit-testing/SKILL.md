---
name: unit-testing
description: "Use when writing, implementing, or running AGS3D unit and integration tests. Covers test infrastructure setup, writing GDScript test suites (UT-Mxx), golden file tests, scene-based tests, and closing TEST-* GitHub issues. Use for any task involving agstests/ directory, test_base.gd, run_tests.gd, or verifying implemented features."
argument-hint: "Test task ID or description, e.g. 'TEST-M1-01' or 'write lexer tests'"
---

# AGS3D Unit Testing

## Architecture

AGS3D has two separate test layers:

| Layer                | Location    | Runner                                      | Purpose                               |
| -------------------- | ----------- | ------------------------------------------- | ------------------------------------- |
| Godot C++ tests      | `tests/`    | `scons tests=yes`                           | Godot engine internals (do not touch) |
| AGS3D GDScript tests | `agstests/` | `--headless --script agstests/run_tests.gd` | AGS3D feature tests                   |

**AGS3D tests are GDScript files run headlessly by the built Godot binary. They are NOT compiled into the engine.**

## Directory Structure

```
agstests/
  project.godot          # minimal Godot project so scene-based tests work
  run_tests.gd           # master runner — instantiates all suites, collects results, exits 0/1
  utils/
    test_base.gd         # base class with assert helpers
    reporter.gd          # collects pass/fail, prints summary, sets exit code
  m1_module/
    test_script_language.gd
  m2_parser/
    test_lexer.gd
    test_parser.gd
    test_symbols.gd
    test_errors.gd
    fixtures/            # .agscript fixture files
  m3_emitter/
    test_emit_statements.gd
    test_emit_await.gd
    test_sourcemaps.gd
    fixtures/            # .agscript inputs + .gd golden files
  m4_room/
    test_room_node.gd
    test_walkable.gd
    test_regions.gd
    test_hotspots.gd
    scenes/              # .tscn scene files for node tests
  m5_character/
    test_character_node.gd
    test_navigation.gd
    test_walkto.gd
    test_faceto.gd
    scenes/
  m6_integration/
    test_script_wiring.gd
    test_runtime.gd
    test_event_binding.gd
    test_end_to_end.gd
    scenes/
    scripts/             # .agscript files used as integration inputs
```

## Running Tests

```sh
# Run full suite
./bin/godot.linuxbsd.editor.x86_64 --headless --path agstests --script run_tests.gd

# Run a single suite (pass suite filename as arg)
./bin/godot.linuxbsd.editor.x86_64 --headless --path agstests --script run_tests.gd m1_module/test_script_language.gd
```

**Exit code 0 = all tests passed. Exit code 1 = at least one failure.**

## test_base.gd — Assert API

Every test class extends `test_base.gd`. Available helpers:

```gdscript
assert_eq(actual, expected, msg := "")       # fail if actual != expected
assert_ne(actual, expected, msg := "")       # fail if actual == expected
assert_true(condition, msg := "")            # fail if not condition
assert_false(condition, msg := "")           # fail if condition
assert_not_null(value, msg := "")            # fail if value == null
assert_null(value, msg := "")               # fail if value != null
assert_no_crash(callable, msg := "")         # run callable, fail if it throws
```

## Writing a Test Suite

```gdscript
# agstests/m1_module/test_script_language.gd
extends "res://utils/test_base.gd"

func suite_name() -> String:
    return "M1: ScriptLanguage"

func test_language_registered() -> void:
    var found := false
    for i in ScriptServer.get_language_count():
        if ScriptServer.get_language(i).get_name() == "AGSScript":
            found = true
    assert_true(found, "AGSScriptLanguage not registered with ScriptServer")

func test_extension_recognised() -> void:
    assert_true(
        ScriptServer.is_global_class_type(""),  # placeholder — replace with correct call
        ".agscript not recognised"
    )
```

**Naming rules:**

- File: `test_<topic>.gd`
- Every method starting with `test_` is auto-discovered and run
- `setUp()` / `tearDown()` run before/after each test (optional)
- `setUpSuite()` / `tearDownSuite()` run once per class (optional)

## run_tests.gd — Registering Suites

```gdscript
# In run_tests.gd, add each suite:
var suites := [
    "m1_module/test_script_language.gd",
    # add more here as milestones complete
]
```

## Golden File Tests (M3 Emitter)

```gdscript
func test_if_stmt_emit() -> void:
    var source := FileAccess.get_file_as_string("res://m3_emitter/fixtures/if_stmt.agscript")
    var result := AGSScriptLanguage.get_singleton().emit(source)
    var golden := FileAccess.get_file_as_string("res://m3_emitter/fixtures/if_stmt.gd")
    assert_eq(result.strip_edges(), golden.strip_edges(), "if_stmt emit mismatch")
```

Golden files live next to their `.agscript` fixtures. To regenerate all goldens:

```sh
./bin/godot.linuxbsd.editor.x86_64 --headless --path agstests --script utils/update_goldens.gd
```

## Scene-Based Tests (M4, M5, M6)

```gdscript
# Instantiate a packed scene and check properties
func test_room_instantiates() -> void:
    var scene := load("res://m4_room/scenes/test_room_basic.tscn") as PackedScene
    assert_not_null(scene, "Scene failed to load")
    var room := scene.instantiate()
    assert_not_null(room, "Scene instantiation returned null")
    assert_true(room is AGSRoom, "Root node is not AGSRoom")
    room.queue_free()
```

For tests requiring frame ticks (navigation, signals), use a coroutine with a timeout:

```gdscript
func test_character_moves() -> void:
    var room := preload("res://m5_character/scenes/test_nav_room.tscn").instantiate()
    add_child(room)
    var character := room.get_node("Player") as AGSCharacter
    character.walk_to("door_left")

    var deadline := Time.get_ticks_msec() + 5000
    while Time.get_ticks_msec() < deadline:
        await get_tree().process_frame
        if character.global_position.distance_to(room.get_point("door_left")) < 0.5:
            pass_test()
            room.queue_free()
            return

    fail("Character did not reach door_left within 5 seconds")
    room.queue_free()
```

## Test → Issue Mapping

| Test task      | Issue #                                                 |
| -------------- | ------------------------------------------------------- |
| TEST-INFRA-01  | varies — check `gh issue list --search "TEST-INFRA-01"` |
| TEST-INFRA-02  | varies                                                  |
| TEST-M1-01     | check with gh                                           |
| TEST-M2-01..04 | check with gh                                           |
| TEST-M3-01..02 | check with gh                                           |
| TEST-M4-01..02 | check with gh                                           |
| TEST-M5-01..02 | check with gh                                           |
| TEST-M6-01..03 | check with gh                                           |

Use: `gh issue list --search "TEST-M1-01" --state all` to find the number.

## Workflow: Implementing a Test Task

1. Find the test task issue: `gh issue list --search "TEST-Mx-xx" --state open`
2. Comment it started: `gh issue comment <N> --body "Work started."`
3. Create the test file(s) in `agstests/`
4. Register new suites in `agstests/run_tests.gd`
5. Run the tests and confirm they pass:
    ```sh
    ./bin/godot.linuxbsd.editor.x86_64 --headless --path agstests --script run_tests.gd
    ```
6. Commit: `git add agstests/ && git commit -m "TEST-Mx-xx: <description> — closes #<N>"`
7. Close the issue with a summary

## Dependency Order

TEST-INFRA-01 must be implemented before any milestone test tasks.
TEST-INFRA-02 (CI) can be done at any time after INFRA-01.
Milestone test tasks depend on the corresponding implementation tasks being done first (see each issue body for exact deps).

## agstests/project.godot

Minimal project file, no main scene, just enough for scene loading to work:

```ini
[application]
config/name="AGS3D Tests"
config/features=PackedStringArray("4.x")
```
