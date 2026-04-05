# AGS3D — Current Development Tasks

This file tracks the active batch of tasks. Update status as work progresses.
When all tasks are done, ask Claude to pick the next 10.

## M9 — Scene Generator & Runtime Core

- [x] **T-E01** — Go: `.agroom` parser → `RoomData` struct
- [x] **T-E02** — Go: `RoomData` → `.tscn` serialiser
- [x] **T-E03** — Go: `.agchar` parser + `CharData` → `.tscn`
- [x] **T-E04** — Go: wire scene gen into `ag build` pipeline
- [x] **T-E18** — Integration: prototype migration
- [x] **T-E05** — Go: `ag validate` cross-reference checks (SpawnPoint→agchar, initial_camera, point names, game.agp paths)

## M10 — Game Systems (Batch 1)

- [x] **T-GS07** — Go: grammar + emitter — `global.NAME` read/write; `[globals]` section in `game.agp`
- [x] **T-GS08** — C++: `AGSRuntime` — global variable store (init from `game.agp`, get/set API, include in save data)
- [x] **T-GS09** — Go: grammar + emitter — `GoToRoom("room")` blocking call *(T-GS10 already done; this wires the emitter)*
- [x] **T-GS02** — C++: `AGSItem` node + `AGSRuntime.get_item()`
- [x] **T-GS03** — C++: `AGSRoomItem` node — `item_clicked` signal, room wiring
- [x] **T-GS04** — Go: `ag build` — `.agitem` parser + inventory/item validation in `ag validate`
- [x] **T-GS05** — Go: grammar + emitter — `Say`, `Think`, `AddInventory`, `LoseInventory`, `HasInventory`
- [x] **T-GS06** — Go: grammar + emitter — `HideRoomItem`, `ShowRoomItem`, `item_interact` handler
- [x] **T-GS18** — GDScript/C++: cutscene support — `SetPlayerControl`, `FadeIn`, `FadeOut`, `Wait`
- [ ] **T-GS19** — Go: grammar + emitter — `SetPlayerControl`, `FadeIn`, `FadeOut`, `Wait`

## Notes

- T-E05 unblocks better error messages for authors — do it first.
- T-GS07 + T-GS08 (globals) are needed by save/load (T-GS16/17) — do early.
- T-GS09: `GoToRoom` emitter — T-GS10 (runtime `load_room`) is already done; this task adds the grammar/emitter side.
- T-GS02→T-GS06 are the item/inventory chain — each depends on the previous.
- T-GS18 + T-GS19 (cutscenes) are independent of items — can be done in parallel with item chain.
- Custom Editor (M12) tasks are **not** in this batch — deferred until engine is stable.
