# AGS3D — Current Development Tasks

This file tracks the active batch of tasks. Update status as work progresses.
When all tasks are done, ask Claude to pick the next 10.

## M9 — AG Studio Editor (Batch 2)

- [x] **#94 T-E11** — GDScript: AGS Inspector plugin — property forms for all AGS node types
- [x] **#95 T-E15** — GDScript: Build Log dock — RichTextLabel + clickable error links
- [ ] **#96 T-E16** — GDScript: Play button wired to `ag build` + `play_main_scene()` *(needs T-E15)*
- [ ] **#97 T-E13** — GDScript: Character editor main screen — property form for `.agchar`
- [ ] **#98 T-E14** — GDScript: Script editor main screen — CodeEdit + AGS syntax highlighting
- [ ] **#99 T-E12** — GDScript: `.agroom` ↔ `.tscn` sync (gizmo edits → write back to `.agroom`) *(needs T-E11)*
- [ ] **#100 T-E17** — GDScript: Project wizard — new project scaffold via AG Studio menu
- [ ] **#101 T-E18** — Integration: prototype migration — delete hand-maintained `.tscn`, verify `ag build` regenerates it
- [ ] **#111 T-GS10** — GDScript: `AGSRuntime.load_room()` + `room_change_requested` signal
- [ ] **#103 T-GS01** — C++: `AGSCharacter` `say_completed` signal + `say()` / `think()` in runtime script

## Notes

- Critical path to first playable build: T-E15 → T-E16
- T-E12 blocked on T-E11; T-E16 blocked on T-E15 — do those first
- T-E18 validates the full pipeline end-to-end
- T-GS10 + T-GS01 are M10 Game Systems seeds — independent of M9 editor work
