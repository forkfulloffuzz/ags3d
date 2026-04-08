# agstests — GDScript Test Suite Agent Instructions

Headless GDScript tests for the AGS3D runtime. Tests run inside a Godot
`SceneTree` without a display, driven by `run_tests.gd`.

## Running tests

```sh
.dev/test.sh        # GDScript tests only
.dev/test-all.sh    # GDScript + Go tests
```

## Test runner

`agstests/run_tests.gd` is the entry point. It declares two arrays:

```gdscript
const SUITES = [...]        # Synchronous test suites
const ASYNC_SUITES = [...]  # Suites with async (await) methods
```

**Every new test file must be registered in one of these arrays.**

## Module layout

| Directory | What it tests |
|-----------|--------------|
| `m_cut/` | Cutscene sequencer, audio cleanup, save blocking, dialogue ducking, skip system |
| `m_dlg/` | Dialogue engine, state machine, UI, localisation |
| `m10_game_systems/` | Animation players, audio, character, GUI, room, save/load, say |
| `m1_module/` | Base module system |
| `m4_room/` | Room parsing and scene generation |
| `m5_character/` | Character parsing |
| `m6_bindings/` | GDScript ↔ AGS-spirit bindings |
| `m6_integration/` | End-to-end integration |

## Adding a test file

1. Create `agstests/<module>/test_<feature>.gd`
2. Extend `AGSTest` (sync) or `AGSAsyncTest` (async)
3. Name test methods `test_<description>()`
4. Register in `run_tests.gd` under `SUITES` or `ASYNC_SUITES`

## Async test rules

- Always `await` async calls, including `_cleanup()`
- Call `_cleanup()` in `_after_each()` or at the end of each test
- Missing `await` on cleanup causes ordering flakiness across tests

## AutoLoad injection pattern

AutoLoad singletons are unavailable in headless tests. Inject dependencies via
node metadata:

```gdscript
# In test setup:
save_load.set_meta("_seq_override", local_sequencer_node)

# In the singleton under test:
func _get_sequencer() -> Node:
    if has_meta("_seq_override"):
        return get_meta("_seq_override")
    return get_node("/root/AGSSequencer")
```

Use this pattern for any singleton that the code under test needs to reach.

## SceneTree constraints

- Nodes that call `add_child()` or emit signals tied to the tree must be added
  to the tree with `add_child(node)` before the test runs
- Nodes added to the tree must be removed in cleanup: `node.queue_free()` +
  `await get_tree().process_frame`
- Out-of-tree nodes: call `_ready()` manually if needed; no signal processing
