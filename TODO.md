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
- [x] **T-GS19** — Go: grammar + emitter — `SetPlayerControl`, `FadeIn`, `FadeOut`, `Wait`

## M10 — Game Systems (Batch 2)

- [x] **T-GS11** — Go: grammar + emitter — `PlayMusic`, `StopMusic`, `PlaySound` (non-blocking; map to `AGSRuntime.play_music`, `stop_music`, `play_sound`)
- [x] **T-GS12** — GDScript: `AGSRuntime` — audio manager (`AudioStreamPlayer` for music + sfx pool; `play_music(name)`, `stop_music()`, `play_sound(name)`; audio files in `audio/music/` and `audio/sfx/`)
- [x] **T-GS16** — GDScript: `AGSRuntime` — `save_game(slot)` / `load_game(slot)` (serialise globals, room name, character inventories, room item visibility to `user://save_<slot>.json`)
- [x] **T-GS17** — Go: grammar + emitter — `SaveGame`, `LoadGame`, `GameSaved` (non-blocking; map to `AGSRuntime.save_game`, `load_game`, `game_saved`)
- [ ] **T-GS27** — C++: split `AGSCharacter` → `AGSCharacterBase` (signals + shared props) + `AGSCharacter3D` + `AGSCharacter2D`; preserve all existing signal/property interface
- [ ] **T-GS28** — GDScript: `AGSAnimationPlayerBase` (common API: `play_clip`, `stop`, `set_state`, `on_anim_event`) + `AGSAnimationPlayer3D` wrapping existing `AnimationPlayer` *(depends on T-GS27)*
- [ ] **T-GS24** — C++: `AGSCharacter` — add `visual_mode` property (`"mesh"` | `"billboard"`); no scene gen change yet
- [ ] **T-GS25** — Go: `.agchar` billboard properties (`visual_mode`, `sprite_sheet`, `sprite_angles`, `frame_size`, `frames_per_angle`); `ag build` outputs `Sprite3D`-rooted `.tscn` when `visual_mode = "billboard"` *(depends on T-GS24)*
- [ ] **T-BL01** — Python: Blender add-on scaffold — `tools/blender_addon/`, `blender_manifest.toml`, register/unregister hooks, installable in Blender 4.x; no UI yet
- [ ] **T-BL02** — Python: AGS3D object type panel — Object Properties sidebar + N-panel; type dropdown (None/WalkableSurface/BlockerVolume/Point/Camera/Hotspot/TriggerRegion/SpawnPoint/NavMesh); stores `AGS_type`/`AGS_name` as custom properties on the Blender object *(depends on T-BL01)*

## Notes

- T-GS11 + T-GS12: audio pair — do T-GS11 (emitter) before T-GS12 (runtime).
- T-GS16 + T-GS17: save/load pair — T-GS07/T-GS08 (globals) already done, no other blockers.
- T-GS27 → T-GS28: character split must come before animation player refactor.
- T-GS24 → T-GS25: visual_mode C++ property must exist before Go scene gen can branch on it.
- T-BL01 → T-BL02: add-on must be installable before any panels are built.
- T-GS29 (AGSAnimationPlayer2D), T-GS30 (ag build char type routing), T-BL03–T-BL09 follow in the next batch.
