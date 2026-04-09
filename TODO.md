# AGS3D — Current Development Tasks

This file tracks the active batch of tasks. Update status as work progresses.
When all tasks are done, ask Claude to pick the next 10.

## Phase 10 — Game Systems: Go CLI Implementation (complete)

- [x] **T-BL16** — Animation frame tags (Blender + Go + runtime) *(done)*
- [x] **T-GS04** — `.agitem` parser + `ag validate` item checks *(done)*
- [x] **T-GS13** — `.agui` parser + GUI scene generator *(done)*
- [x] **T-GS11** — PlayMusic / StopMusic / PlaySound emitter *(done)*
- [x] **T-GS07** — `global.NAME` read/write emitter *(done)*
- [x] **T-GS17** — SaveGame / LoadGame / GameSaved emitter *(done)*
- [x] **T-GS15** — SetStatusText / SetActiveVerb / GetActiveVerb emitter *(done)*
- [x] **T-GS19** — SetPlayerControl / FadeIn / FadeOut / Wait emitter *(done)*
- [x] **T-GS05** — Say / Think / AddInventory / LoseInventory / HasInventory emitter *(done)*
- [x] **T-GS06** — HideRoomItem / ShowRoomItem / item_interact emitter *(done)*

## Phase 11 — Game Systems: Blender Addon + 2D Character Scene Gen

Blender addon tasks: all core features are implemented (T-BL01-10 done, T-BL11-13 done,
T-BL16 done). Remaining polish: camera eyedropper UI operator (T-BL06).

2D character tasks: the `char` parser, `generate2DCharScene`, and runtime GDScripts
(`ags_billboard_controller.gd`, `ags_animation_player_2d.gd`) all exist. The remaining
work is wiring the `AGSBillboardController` and `AGSAnimationPlayer2D` child nodes into
the Go-generated 2D character `.tscn`.

- [x] **T-BL06** — Blender: camera look_at eyedropper operator. Added `AGS3D_OT_EyedropLookAt`
  modal operator (`panels.py:91-132`); button with `EYEDROPPER` icon now appears next to
  the `prop_search` widget in the Camera panel (`panels.py:203`). Left-click to pick,
  right-click or Esc to cancel. *(done)*

- [x] **T-GS25** — Go: billboard `.agchar` properties in generated 2D scene.
  `generate2DCharScene` now emits `AGSBillboardController` and `AGSAnimationPlayer2D`
  child nodes with `@export` properties wired from `CharData`: `sprite_angles`,
  `frames_per_angle` → `hframes/vframes`, `sprite_path`, `controller_path`. Tests in
  `tools/ag/internal/scene/char_scene_test.go`. *(done)*

- [x] **T-GS26** — GDScript: billboard direction runtime. Runtime `ags_billboard_controller.gd`
  implemented; Go scene gen wiring done via T-GS25. *(done)*
- [x] **T-GS29** — GDScript: `AGSAnimationPlayer2D`. Runtime `ags_animation_player_2d.gd`
  implemented; Go scene gen wiring done via T-GS25. *(done)*

- [ ] **T-GS27** — C++: split `AGSCharacter` → `AGSCharacterBase` / `AGSCharacter3D` /
  `AGSCharacter2D`. Defers C++ Godot fork node split to later. *(blocked on C++ fork)*

- [ ] **T-GS24** — C++: `AGSCharacter.visual_mode` property — blocks billboard scene
  gen C++ wiring. *(blocked on C++ fork)*

- [ ] **T-GS30** — Go: generate `AGSCharacter3D` vs `AGSCharacter2D` `.tscn` based
  on `type` field in `.agchar`. Already partially done (`generate3DCharScene` vs
  `generate2DCharScene`); final C++ wiring defers to T-GS27. *(mostly done; C++
  node type split pending)*

## Phase 9 — Infrastructure, Localisation & Runtime (complete)

- [x] **TEST-INFRA-02** — CI: GitHub Actions workflow (done — commit 67ac49d150)
- [x] **T-CUT30** — Go: cutscene localisation pipeline (done — commit 9aefb928f7)
- [x] **T-LOC16** — Pipeline + Editor: voice file connection and recording coverage (done — commit d63c73185d)
- [x] **T-LOC05** — Go: interactive standalone `ag-loc` translation tool (done — commit 021a30455e)
- [x] **T-LOC18** — Go: multi-source-language authoring + `ag export` stub generation (done — commit dabf8ce3a3)
- [x] **T-LOC-NORM** — Normalise all `.agcut` command arguments with `#` prefix (done — commit 2ab3eb6aef)

## Notes

- Phase 11 has 3 unblocked tasks (T-BL06, T-GS25) and 4 deferred (T-GS26, T-GS29, T-GS27, T-GS24, T-GS30) — these defer to C++ Godot fork access.
- T40 and T-FINAL remain deferred until C++ Godot fork access is available.
- T-E19, T-BL14, T-GS21 remain deferred to M12 AG Studio.
- After Phase 11 unblocked work: M10 C++ node tasks (T-GS01, T-GS02, T-GS08, T-GS10, T-GS12, T-GS14, T-GS16, T-GS18) open up once C++ fork is available, followed by M12 AG Studio editor tasks.
