# AGS3D — Claude Code Instructions

## Development Workflow

### Task tracking

Development tasks are tracked in `TODO.md` at the project root.

- Work through tasks **in order** unless a task is explicitly blocked.
- Mark a task complete (`[x]`) in `TODO.md` **immediately** after finishing it.
- After marking a task complete, add a manual-test checklist for that task to `TODO_TESTS.md` (next to `TODO.md`). Each checklist must be concrete and runnable by a human — specific UI actions, expected outputs, error cases.
- Work **one task at a time**: finish and mark complete before starting the next.
- Do exactly what the task asks — no scope creep, no pre-emptive refactoring.

### Custom Editor UI tracking

The Custom Editor (AG Studio) milestone (M12) is deferred. Engine and runtime
features are built first and used through the normal Godot editor. The M12
milestone doc at `docs/custom-editor-milestone.md` tracks what UI will
eventually need to be built.

**Rule:** After completing any engine, runtime, or tooling task (C++ node,
Go CLI feature, GDScript runtime method, `.agroom` block type), add or update
a stub entry in the `## Pending UI Stubs` table in
`docs/custom-editor-milestone.md` describing what AG Studio UI will expose
that feature. Use prefix `T-CE-UI` and note the engine task it corresponds to.
This keeps M12 complete as engine work accumulates.

When **all tasks in `TODO.md` are completed**:
1. Update the final task to `[x]`.
2. Print a one-line summary for every completed task in the batch. Each line must state what capability exists now that did not exist before (format: `- **T-ID** — [what you can do now that you couldn't before]`). Under each task line, list every test file added or extended for that task with test count (format: `  - Tests: \`path/to/test_file.gd\` (N tests), \`path/to/test_file_test.go\` (N tests)`). Omit the Tests sub-line only when a task added no automated tests.
3. Stop and ask the user: *"All tasks are done. Shall I analyse the milestones and pick the next 10 tasks?"*
4. Only proceed to select new tasks after explicit user approval.
5. When approved: review open GitHub issues across active milestones, apply the same ordering logic (critical path first, then unblocked parallel work), write the new batch to `TODO.md`, and report the list to the user.

### Commits

- Commit after **every completed task** so each task has its own reviewable commit.
- Commit only the files changed by that task — no bundling unrelated changes.
- Follow the standard git commit protocol (status → diff → log → commit).
- Do not commit docs or unrelated files alongside task work.

### File references

Always use markdown link syntax for file/line references so they are clickable in the IDE:
`[filename.ts:42](src/filename.ts#L42)`

### ag.sh

Keep `.dev/ag.sh` in sync whenever a new `ag` command or visualisation stage is added.

### Testing

Run `.dev/test.sh` or `.dev/test-all.sh` to verify changes. Do not skip hooks.

When a task adds new C++ types, signals, properties, or GDScript runtime methods that can be exercised headlessly, add automated tests in `agstests/` and register them in `agstests/run_tests.gd`. Editor-only tasks (UI panels, docks, gizmos) go in `TODO_TESTS.md` as manual tests only.

## Project layout

```
docs/           Design documents and milestone specs
tools/ag/       Go CLI (ag build, ag validate, ...)
.engine/        GDScript runtime (deployed into Godot projects)
.dev/           Developer scripts (test.sh, ag.sh, ...)
```

## Active milestones

| ID  | Name                    | Doc | Status |
|-----|-------------------------|-----|--------|
| M9  | Scene Generator & Runtime Core | docs/editor-milestone.md | Essentially complete — only T-E05 (`ag validate`) remaining |
| M10 | Game Systems            | docs/game-systems-milestone.md | Active |
| M11 | Blender Integration     | docs/blender-integration-milestone.md | Active |
| M12 | Custom Editor           | docs/custom-editor-milestone.md | **Deferred** — build after engine features are stable |

## Key architecture decisions

- **AG Studio** = Godot fork running in editor mode with `EditorPlugin` replacing Godot UI. NOT the Wails app.
- **`--godot-editor` flag** = skips AG Studio plugin, boots standard Godot editor for debugging.
- **`ag build`** = Go CLI that parses `.agroom`/`.agchar`/`.agscript` and generates `.tscn` scenes. Scene generation is NOT done via `EditorImportPlugin`.
- **Character types**: `AGSCharacterBase` (C++) → `AGSCharacter3D` / `AGSCharacter2D`; animation players inherit from `AGSAnimationPlayerBase`.
