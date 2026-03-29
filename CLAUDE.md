# AGS3D — Claude Code Instructions

## Development Workflow

### Task tracking

Development tasks are tracked in `TODO.md` at the project root.

- Work through tasks **in order** unless a task is explicitly blocked.
- Mark a task complete (`[x]`) in `TODO.md` **immediately** after finishing it.
- Work **one task at a time**: finish and mark complete before starting the next.
- Do exactly what the task asks — no scope creep, no pre-emptive refactoring.

When **all tasks in `TODO.md` are completed**:
1. Update the final task to `[x]`.
2. Stop and ask the user: *"All tasks are done. Shall I analyse the milestones and pick the next 10 tasks?"*
3. Only proceed to select new tasks after explicit user approval.
4. When approved: review open GitHub issues across active milestones, apply the same ordering logic (critical path first, then unblocked parallel work), write the new batch to `TODO.md`, and report the list to the user.

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

## Project layout

```
docs/           Design documents and milestone specs
tools/ag/       Go CLI (ag build, ag validate, ...)
.engine/        GDScript runtime (deployed into Godot projects)
.dev/           Developer scripts (test.sh, ag.sh, ...)
```

## Active milestones

| ID  | Name                    | Doc |
|-----|-------------------------|-----|
| M9  | AG Studio Editor        | docs/editor-milestone.md |
| M10 | Game Systems            | docs/game-systems-milestone.md |
| M11 | Blender Integration     | docs/blender-integration-milestone.md |

## Key architecture decisions

- **AG Studio** = Godot fork running in editor mode with `EditorPlugin` replacing Godot UI. NOT the Wails app.
- **`--godot-editor` flag** = skips AG Studio plugin, boots standard Godot editor for debugging.
- **`ag build`** = Go CLI that parses `.agroom`/`.agchar`/`.agscript` and generates `.tscn` scenes. Scene generation is NOT done via `EditorImportPlugin`.
- **Character types**: `AGSCharacterBase` (C++) → `AGSCharacter3D` / `AGSCharacter2D`; animation players inherit from `AGSAnimationPlayerBase`.
