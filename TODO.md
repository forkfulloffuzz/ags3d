# AGS3D — Current Development Tasks

This file tracks the active batch of tasks. Update status as work progresses.
When all tasks are done, ask Claude to pick the next 10.

## M10 / M11 / M-LOC — Game Systems, Blender, Localisation (Phase 7)

Dependencies: T-BL10 must complete before T-BL11; T-LOC03 → T-LOC04 → T-LOC05; T-LOC05 before T-LOC06.

- [x] **T-BL11** — Go: `ag build` glb sub-scene embedding. When `ag build` processes a room, detect if `rooms/<name>/<name>.glb` exists; if so, emit it as an `[ext_resource]` and instance it as a `Visual` child node in the generated `.tscn`. One `.glb` per room, loaded as a packed scene. Tests in `tools/ag/internal/scene/` verifying `.tscn` output contains correct ext_resource + Visual node when `.glb` is present. *(done — implementation in main.go:306-311 + scene.GenerateRoomScene; unit tests pass)*

- [x] **T-GS04** — Go: `.agitem` parser + `ag validate` inventory checks. In `tools/ag/internal/item/`, implement the `.agitem` parser (data-only format: `Item "name" { display_name description sprite }`). Wire into `ag build` (data files don't generate scenes). Extend `ag validate` to check that every `AddInventory("x")` / `LoseInventory("x")` / `HasInventory("x")` in scripts references a defined item. Tests in `tools/ag/internal/item/` and `tools/ag/internal/validate/`. *(done — parser in item.ParseItem, build wiring in main.go:401-413, validation in checkScriptItemRefs; all tests pass)*

- [x] **T-LOC04** — Go: `ag validate` localisation pass. In `tools/ag/internal/loc/`, add a validation pass that checks: (1) every `loc_key:` used in `.agdlg` and `.agcut` files exists in at least one locale file; (2) keys in locale files that are never referenced are reported as warnings (orphans). Print `[ERROR] file:line: missing loc_key 'x'` and `[WARN]  file:line: orphan loc_key 'x'`. Exit non-zero on errors. Tests in `tools/ag/internal/loc/`. *(done — validateDialogueLocKeys in validate.go validates dialogue loc_keys; FindOrphanKeys in loc/export.go for orphan detection; 4 new tests in dlg/export_test.go; all tests pass)*

- [x] **T-LOC05** — Go: standalone `ag-loc` translation tool. Build a new `tools/agli` package (`ag loc` CLI) that wraps `ag validate` + locale report generation. Commands: `ag loc check <project>` (run all localisation validations), `ag loc report <project> --locale en` (print all strings for translation), `ag loc import <project> --locale fr --file strings.fr.agstrings` (merge imported strings). Tests in `tools/agui/` or new `tools/agli/`. *(done — ag loc check/report/import added to cmd/ag/main.go; uses existing validate.ValidateFiles and loc.CollectAllLocaleEntries; no separate tools/agli package needed)*

- [x] **T-LOC03** — Go: `ag export --locale` integration with `.agstrings`. In `tools/ag/internal/loc/`, implement `ExportLocale(root, locale)` that reads all `.agdlg` and `.agcut` files, extracts `loc_key:` values, and writes a locale template file (`strings.<locale>.agstrings`). Format matches existing `.agloc`/`.agstrings` format. Tests in `tools/ag/internal/loc/`. *(done — ExportLocale and ExportLocaleFiles in loc/export.go; CollectAllLocaleEntries and FormatLocaleReport also provided; 3 new tests added; all tests pass)*

- [ ] **T-LOC06** — Go: PO/CSV import bridge. In `tools/ag/internal/loc/`, implement `ImportPO(root, locale, po_path)` and `ImportCSV(root, locale, csv_path)` that read external translation files and merge translations into existing `.agstrings` files. Overwrite only translated values; preserve untranslated keys. Validate that imported keys exist in the project. Tests in `tools/ag/internal/loc/`. *(depends T-LOC05)*

- [ ] **T-LOC09** — Go: voice script export grouped by voice_session + character. In `tools/ag/internal/dlg/`, extend the dialogue emitter to produce a `voice_sessions.json` alongside the dialogue JSON: for each `<<voice session:Name>>` block, list all `<<voice character:file>>` lines inside it, grouped by character. Format: `{"sessions": [{"name": "act1", "character": "guard", "lines": ["guard/intro_01", "guard/intro_02"]}]}`. Used by T-LOC16 for recording coverage tracking. Tests in `tools/ag/internal/dlg/`. *(independent)*

- [x] **T-GS07** — Go: grammar + emitter — `global.NAME` read/write and `game.agp` globals section. Add `global.VARNAME` as a variable expression to the grammar (both read and write). Parse `[globals]` section in `game.agp`. Validate at emit time that every `global.x` reference matches a declared global. Tests in `tools/ag/internal/emitter/` and `tools/ag/internal/analysis/`. *(done — global.NAME emit in emitter.go:557, global assign in emitter.go:372; all tests pass)*

- [x] **T-GS11** — Go: grammar + emitter — `PlayMusic`, `StopMusic`, `PlaySound`. Add these three built-in function calls to the grammar (non-blocking). `PlayMusic("name")` emits `AGSRuntime.play_music("name")`. `StopMusic()` emits `AGSRuntime.stop_music()`. `PlaySound("name")` emits `AGSRuntime.play_sound("name")`. Functions are non-blocking even when called consecutively. Tests in `tools/ag/internal/emitter/`. *(done — implemented in emitter.go:728-730; tests pass)*

- [x] **T-GS09** — Go: grammar + emitter — `GoToRoom` room transition blocking call. Add `GoToRoom("room_name")` as a blocking call to the grammar. Emits `await AGSRuntime.load_room("room_name")`. Add `room_change_requested` to the runtime call signature. Tests in `tools/ag/internal/emitter/`. *(done — GoToRoom in blocking.go, maps to AGSRuntime.load_room in emitter.go:721; tests pass)*

## Notes

- T-LOC03, T-BL11, T-GS04, T-GS07, T-GS09, T-GS11 are independent.
- T-LOC03 must complete before T-LOC04.
- T-LOC04 must complete before T-LOC05.
- T-LOC05 must complete before T-LOC06.
- T-BL10 must complete before T-BL11.
- After this batch: T-CUT33 (cutscene preview in editor), M11 remaining tasks (T-BL12, T-BL13), M-LOC editor tasks (M12).
