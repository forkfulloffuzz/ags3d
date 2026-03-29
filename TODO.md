# AGS3D — Current Development Tasks

This file tracks the active batch of tasks. Update status as work progresses.
When all tasks are done, ask Claude to pick the next 10.

## M9 — AG Studio Editor (Batch 1)

- [x] **#84 T-E01** — Go: `.agroom` parser → `RoomData` struct
- [x] **#85 T-E02** — Go: `RoomData` → Godot `.tscn` serialiser
- [x] **#83 T-E03** — Go: `.agchar` parser + `CharData` → `.tscn`
- [x] **#86 T-E04** — Go: wire scene generation into `ag build` pipeline
- [x] **#87 T-E05** — Go: `ag validate` cross-reference checks
- [x] **#152 T-UI-01** — Go/Wails: agui room/char parser, generator and validator panel
- [x] **#88 T-E06** — C++: `--godot-editor` launch flag
- [x] **#91 T-E07** — GDScript: `EditorPlugin` skeleton — hide Godot docks, register custom screens
- [x] **#90 T-E08** — GDScript: Project panel dock
- [ ] **#89 T-E09** — GDScript: Room editor main screen + 3D viewport embed
- [ ] **#93 T-E10** — GDScript: AGS gizmo plugins for all node types

## Notes

- Tasks 1–5 are pure Go CLI (no editor needed — can run/test standalone)
- Task 6 is isolated C++ (can be done in parallel with Go tasks)
- Tasks 7–10 require the EditorPlugin foundation (T-E07 first)
- Critical path: T-E01 → T-E02 → T-E04 → T-E07 → T-E09
