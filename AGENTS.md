# AGS3D — Agent Instructions

AGS3D is a 3D/2D adventure game framework built on Godot 4, with a Go CLI for
building and validating project files, a GDScript runtime, and a Blender addon
for asset export.

> **Claude Code users:** The primary instructions are in `CLAUDE.md`. This file
> provides an equivalent summary for other agent systems.

## Workflow rules

- Work tasks from `TODO.md` in order; mark each `[x]` immediately on completion.
- After marking a task complete, append a manual-test checklist to `TODO_TESTS.md`.
- Commit after every task — only the files changed by that task.
- Run `.dev/test.sh` or `.dev/test-all.sh` before committing. Never skip hooks.
- Do exactly what the task asks — no scope creep, no pre-emptive refactoring.
- Keep `.dev/ag.sh` in sync whenever a new `ag` command or viz stage is added.

## Batch completion summary format

When all tasks in a batch are done, print one line per task:

```
- **T-ID** — [capability that now exists]
  - Tests: `path/to/test.gd` (N tests), `path/to/test_test.go` (N tests)
```

Omit the Tests sub-line only when a task added no automated tests.

## Project layout

```
docs/                    Design documents and milestone specs
tools/ag/                Go CLI — ag build, ag validate, ag viz
tools/blender_addon/     Blender addon — .glb + .aganim export
game_prototype/          Godot project used for prototyping
  .engine/runtime/       GDScript AutoLoad singletons and base classes
  .engine/generated/     ag build output — do not edit manually
agstests/                Headless GDScript test suite
.dev/                    Developer scripts (test.sh, ag.sh, build.sh, …)
```

## Area-specific instructions

| Area | AGENTS file |
|------|------------|
| Go CLI | [tools/ag/AGENTS.md](tools/ag/AGENTS.md) |
| Blender addon | [tools/blender_addon/AGENTS.md](tools/blender_addon/AGENTS.md) |
| GDScript test suite | [agstests/AGENTS.md](agstests/AGENTS.md) |
| Game prototype / runtime | [game_prototype/AGENTS.md](game_prototype/AGENTS.md) |
| Runtime AutoLoads | [game_prototype/.engine/runtime/AGENTS.md](game_prototype/.engine/runtime/AGENTS.md) |

## Active milestones

| ID  | Name | Status |
|-----|------|--------|
| M9  | Scene Generator & Runtime Core | Essentially complete |
| M10 | Game Systems | Active |
| M11 | Blender Integration | Active |
| M12 | Custom Editor | **Deferred** |
