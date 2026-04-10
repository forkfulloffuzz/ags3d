# AGS3D — Current Development Tasks

This file tracks the active batch of tasks. Update status as work progresses.
When all tasks are done, ask Claude to pick the next 10.

## Phase 11 — Blender Addon Polish + 2D Character Scene Gen (complete)

- [x] **T-BL06** — Blender: camera look_at eyedropper operator. Added `AGS3D_OT_EyedropLookAt`
  modal operator (`panels.py:91-132`); button with `EYEDROPPER` icon now appears next to
  the `prop_search` widget in the Camera panel (`panels.py:203`). Left-click to pick,
  right-click or Esc to cancel. *(done)*
- [x] **T-GS25** — Go: billboard `.agchar` properties in generated 2D scene.
  `generate2DCharScene` now emits `AGSBillboardController` and `AGSAnimationPlayer2D`
  child nodes. *(done)*
- [x] **T-GS26** — GDScript: billboard direction runtime. `ags_billboard_controller.gd`
  implemented; Go scene gen wiring done via T-GS25. *(done)*
- [x] **T-GS29** — GDScript: `AGSAnimationPlayer2D`. `ags_animation_player_2d.gd`
  implemented; Go scene gen wiring done via T-GS25. *(done)*
- [x] **T-BL07** — Blender: export merge mode. Implemented in `operators.py:405-440`
  (`_append_existing_blocks`, `merge_mode` property). *(done — was already in code)*
- [x] **T-BL08** — Blender: import operator. Implemented in `operators.py:269-369`
  (`AGS3D_OT_ImportRoom` with full `.agroom` parser and object creation). *(done)*
- [x] **T-BL09** — Blender: NavMesh baking. Implemented in `operators.py:657-739`
  (`_bake_navmesh`, `AGS3D_OT_BakeNavMesh`). *(done)*

## Phase 12 — Go CLI Validation + Runtime Completions (complete)

- [x] **T-E19 (Go part)** — `ag validate` billboard camera warnings. Added `validateBillboardCameraWarnings`
  in `validate.go:879-960`: W1 (camera elevation >30° with billboard chars), W3 (camera
  arc >45° with 4-angle sprites). 6 new tests in `validate_test.go`. *(done)*
- [x] **T-VAL01** — `ag validate`: `HideRoomItem`/`ShowRoomItem` → hotspot name cross-check.
  Added `checkScriptRoomItemRefs` in `validate.go:772-815`. 4 new tests. *(done)*
- [x] **T-VAL02** — `ag validate`: `GoToRoom` room-name cross-check. Added
  `checkScriptGoToRoomRefs` in `validate.go:818-855`. Resolves `GoToRoom("name")` against
  `rooms/<name>/<name.agroom>`. 4 new tests. *(done)*
- [x] **T-GS28** — GDScript: `AGSAnimationPlayer3D`. Already fully implemented in
  `ags_animation_player_3d.gd` (77 lines: `play_clip`, `stop`, `set_state`, `on_anim_event`).
  *(done — was already in code)*
- [x] **T-GS18 (GDScript part)** — `fade_in`/`fade_out` runtime. Already implemented in
  `ags_cutscene.gd` (Tween-based ColorRect fade in CanvasLayer). *(done)*

## Phase 13 — AGSRuntime AutoLoad + Room Transitions (complete)

- [x] **T-GS10** — GDScript: `AGSRuntime` AutoLoad with `load_room()` + `room_change_requested`
  signal. Created `ags_runtime.gd` with the full AGSRuntime API surface (room transitions,
  player control, audio signals, HUD, inventory, save/load). Wired `AGSRuntime`, `AGSSaveLoad`,
  `AGSAudio`, `AGSRoomManager` as AutoLoads in `project.godot`. `ags_room_manager.gd` updated
  to use `get_node("/root/AGSRuntime")` instead of `Engine.get_singleton`. *(done)*

## Phase 9 — Infrastructure, Localisation & Runtime (complete)

- [x] **TEST-INFRA-02** — CI: GitHub Actions workflow (done — commit 67ac49d150)
- [x] **T-CUT30** — Go: cutscene localisation pipeline (done — commit 9aefb928f7)
- [x] **T-LOC16** — Pipeline + Editor: voice file connection and recording coverage (done — commit d63c73185d)
- [x] **T-LOC05** — Go: interactive standalone `ag-loc` translation tool (done — commit 021a30455e)
- [x] **T-LOC18** — Go: multi-source-language authoring + `ag export` stub generation (done — commit dabf8ce3a3)
- [x] **T-LOC-NORM** — Normalise all `.agcut` command arguments with `#` prefix (done — commit 2ab3eb6aef)

## Deferred / Blocked

### C++ Godot fork required
- **T-GS27** — C++: split `AGSCharacter` → `AGSCharacterBase` / `AGSCharacter3D` / `AGSCharacter2D`
- **T-GS24** — C++: `AGSCharacter.visual_mode` property
- **T-GS30** — Go: generate `AGSCharacter3D` vs `AGSCharacter2D` `.tscn` (mostly done; C++ split pending)
- **T40** — C++/GDScript: disable AGSRuntime trace in production builds
- **T-FINAL** — C++: embed `.engine/runtime/` GDScripts into the C++ module at build time
- M10 C++ node tasks: T-GS01, T-GS02, T-GS08, T-GS12, T-GS14, T-GS16, T-GS18

### M12 AG Studio required
- **T-E19 (editor part)** — GDScript: billboard camera gizmo overlays in Room editor
- **T-BL14** — GDScript: AG Studio "Re-import from Blender" button
- **T-GS21** — GDScript: Room editor `RoomItem` gizmo + placement
