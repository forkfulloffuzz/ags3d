# AGS3D — Current Development Tasks

This file tracks the active batch of tasks. Update status as work progresses.
When all tasks are done, ask Claude to pick the next 10.

## Phase 9 — Infrastructure, Localisation & Runtime

Dependencies: TEST-INFRA-02 unblocks CI; T-CUT30 unblocks T-LOC16; T40/T-FINAL are runtime build tasks.

- [x] **TEST-INFRA-02** — CI: GitHub Actions workflow that builds AGS3D and runs the full test suite on every push and PR. Create `.github/workflows/test.yml`. Run `scons platform=linuxbsd`, then `agstests/run_tests.gd` headlessly. Cache SCons build objects. *(independent; done — commit 67ac49d150)*

 - [x] **T-CUT30** — Go: cutscene localisation pipeline. `<<line>>`, `<<title_card>>`, `<<subtitle>>`, `<<choice>>` commands in `.agcut` files participate in the loc_key pipeline using cutscene title as namespace. Lines appear in `ag export --locale` output. `voice_session` header groups lines in `ag export --voicescript`. Tests in `tools/ag/internal/cut/`. *(done — verified already implemented in commit 9aefb928f7)*

- [x] **T-LOC16** — Pipeline + Editor: voice file connection and recording coverage tracking. Connect recorded voice audio files to their loc_keys so the pipeline knows which lines have been recorded, which are missing, and which are stale (recorded against old source text). Machine-readable `voice_coverage.json` listing `(loc_key, file_path, duration_ms, hash)` for each recorded line. Tests in `tools/ag/internal/cut/`. *(depends T-CUT30; VoiceCoverageEntry/BuildVoiceCoverageReport/ScanVoiceDirectory added; ag voice coverage CLI command)*

 - [x] **T-LOC05** — Go: interactive standalone `ag-loc` translation tool. A terminal/TUI tool for authoring translations directly in `.agstrings` files with live validation feedback. Shows source text, character, scene context; marks entries complete on save. Partial implementation exists (check/report/import done). Full interactive TUI still needed. Tests in `tools/ag/`. *(done — Bubble Tea TUI in tools/ag/internal/loc/tui/; ag loc tui command)*

- [ ] **T40** — C++/GDScript: disable AGSRuntime trace in production builds. Add `#if DEBUG` guards around `trace()` calls so they are compiled out in release. *(independent)*

- [ ] **T-FINAL** — C++ build: embed `.engine/runtime/` GDScripts into the C++ module at build time so they are part of the module binary and don't need to be installed separately. *(depends on having all runtime GDScripts finalized)*

- [ ] **T-E19** — GDScript: billboard camera warnings as gizmo overlays. Show elevation angle warning (>30°), arc width warning for 4-angle sprites, and sprite_locked indicator in the Room editor 3D viewport. Tests in `agstests/`. *(depends T-CE09 — AG Studio Room editor main screen)*

- [ ] **T-BL14** — GDScript: Room editor "Re-import from Blender" button. Detects when `.glb` is newer than `.tscn`, runs `ag build`, shows conflict banner when `.agroom` was edited in AG Studio after last Blender export. Tests in `agstests/`. *(depends T-CE08)*

- [ ] **T-GS21** — GDScript: Room editor `RoomItem` gizmo + placement. Toolbar "Add Item" button; sprite icon appears at chosen position. Inspector shows item reference + visible toggle. `HideRoomItem`/`ShowRoomItem` script autocomplete. Tests in `agstests/`. *(depends T-CE07)*

- [ ] **T-BL16** — Blender: animation frame tags — trigger events (sounds, signals) on specific keyframes. Exports `.aganim` sidecar with frame tag events. Feasibility not yet assessed — mark as STUB if not viable. *(depends T-BL12, T-BL13)*

## Notes

- TEST-INFRA-02 is independent (CI infrastructure).
- T40 and T-FINAL are runtime build tasks — require C++ Godot fork changes (deferred until fork access).
- T-E19, T-BL14, T-GS21 depend on AG Studio (M12) editor features.
- T-BL16 is a STUB — assess feasibility first.
- After Phase 9: M12 AG Studio remaining tasks, then M13 (audio system).

## Phase 9 complete: TEST-INFRA-02, T-CUT30, T-LOC16, T-LOC05
