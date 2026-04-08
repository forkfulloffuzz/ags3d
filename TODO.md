# AGS3D — Current Development Tasks

This file tracks the active batch of tasks. Update status as work progresses.
When all tasks are done, ask Claude to pick the next 10.

## M-CUT / M-DLG — Runtime (Phase 6: Audio cleanup, save, ducking, Blender, localisation, DLG validators)

Dependencies: T-CUT25 → T-CUT26; T-DLG05 → T-DLG06 → T-DLG19; T-CUT27 → T-CUT28.
T-CUT31 and T-CUT32 are independent. T-CUT30 is independent (depends on completed T-CUT09/T-DLG08).

- [x] **T-CUT31** — GDScript: Audio leak cleanup on cutscene end/skip. The sequencer already tracks every audio channel started by `<<music>>`, `<<sound>>`, `<<ambient>>`, and `<<voice>>` commands (stored in `_active_audio_channels`). When the sequence ends normally or is skipped, iterate that list and stop or fade out each channel that was started by this cutscene (scope: `cutscene_owned`). Channels marked `audio_scope:room` during the command are excluded from cleanup. Emit `audio_cleanup_complete` signal after all fades finish. Tests in `agstests/m_cut/` (add to `test_audio_commands.gd` or a new `test_audio_cleanup.gd`). *(independent)*

- [x] **T-CUT25** — GDScript: Save blocking during cutscenes. While a cutscene sequence is running (`AGSSequencer.is_playing() == true`), calls to `AGSSaveLoad.save_game()` must be queued or rejected — the game must not be saved mid-cutscene. Implement a `save_blocked` property on `AGSSaveLoad` (or a blocking mechanism in `AGSSequencer`) so that `save_game()` returns `false` (or emits `save_blocked` signal) when called while a cutscene is active. When the cutscene finishes, any queued save is automatically triggered. Tests in `agstests/m_cut/test_save_blocking.gd`. *(independent — T-CUT26 depends on this)*

- [ ] **T-CUT26** — GDScript: Cutscene state in save graph. Extend `AGSSaveLoad` (or the save data structure) to record, for each named cutscene: `viewed` (bool), `view_count` (int), `skipped` (bool). On `load_game()` restore these values into `AGSSequencer` so `cutscene.Viewed("name")`, `cutscene.ViewCount("name")`, `cutscene.Skipped("name")` return correct values after a load. Tests in `agstests/m_cut/test_cutscene_save_state.gd`. *(depends T-CUT25)*

- [ ] **T-CUT32** — GDScript: Dialogue ducking during cutscenes. The `<<line>>` and `<<dialogue>>` commands already accept a `duck:channels` parameter (stored but not yet acted on). Implement the ducking mechanism: when a `<<line>>` with `duck:` starts, lower volume on the specified audio channel names by `duck_level` (default −12 dB) over `duck_fade` seconds (default 0.2s); when the line ends, restore volume over `duck_restore` seconds (default 0.3s). Delegate to `AGSAudio` for the volume tween. If `AGSAudio` does not yet expose a `duck_channel(name, level, fade)` method, add it. Tests in `agstests/m_cut/test_dialogue_ducking.gd`. *(independent)*

- [ ] **T-DLG05** — Go: Cross-system dialogue validator. In `tools/ag/internal/dlg/`, add a `Validator` that cross-checks a compiled `.agdlg` graph against the full project: (1) every `<<character X>>` line references a character defined in a `.agchar` file; (2) every `goto:` / `jump:` target node exists in the same file; (3) every `loc_key:` references a key present in at least one `.agloc` locale file; (4) every `item:` flag references an item defined in a `.agitem` file. Return structured `ValidationError{File, Line, Code, Message}` results. Tests in `tools/ag/internal/dlg/` using table-driven Go tests. *(independent — T-DLG06 and T-DLG19 depend on this)*

- [ ] **T-DLG06** — Go: Dialogue static analysis warnings. Extend the T-DLG05 validator with lint-style warnings (not errors): (1) unreachable nodes (no incoming `goto:` and not the root node); (2) nodes with no outgoing choice and no `end`/`return` — implicit dead ends; (3) duplicate `loc_key:` values within the same file; (4) choice branches with identical display text. Return `ValidationWarning{File, Line, Code, Message}` alongside errors. Tests in `tools/ag/internal/dlg/`. *(depends T-DLG05)*

- [ ] **T-DLG19** — Go: Integrate dialogue validator into `ag validate`. Wire the T-DLG05/T-DLG06 validator into `tools/ag/cmd/validate.go` so `ag validate <project>` runs the dialogue cross-system check in addition to existing room/character validations. Print errors as `[ERROR] file:line: message` and warnings as `[WARN]  file:line: message`. Exit non-zero if any errors (not warnings). Tests in `tools/ag/` via `TestCLI_Validate_*`. *(depends T-DLG05, T-DLG06)*

- [ ] **T-CUT27** — Python/Blender: Frame tag export from Blender actions. In `tools/blender/ags_export.py` (or a new `ags_frame_tags.py`), read Blender Action pose markers and export them as `frame_tags` in the `.aganim` JSON format: `{"name": "hit", "frame": 12}`. Export runs as part of the existing character export pipeline. One `.aganim` file is produced per Blender Action (mapped to an AGS animation name). Tests: add a `test_frame_tags.py` Blender-headless unit test in `tools/blender/tests/` that verifies correct JSON output for a mock action with two markers. *(independent — T-CUT28 depends on this)*

- [ ] **T-CUT28** — Go: `ag build` injects `.aganim` frame tags into generated scenes. When `ag build` processes a character, read all `.aganim` files in the character's asset directory and inject their `frame_tags` as `AGSAnimationPlayer` metadata in the generated `.tscn`. `AGSAnimationPlayer.get_frame_tag(anim_name, frame)` should return the tag name (or `""`) for a given frame. Tests in `tools/ag/` verifying that `.tscn` output contains correct metadata when `.aganim` files are present. *(depends T-CUT27)*

- [ ] **T-CUT30** — Go: Cutscene localisation. In `tools/ag/internal/cut/`, add a localisation pass that extracts every string literal from `<<line>>`, `<<title_card>>`, `<<subtitle>>`, and `<<choice>>` commands that has a `loc_key:` attribute, and writes them to an `.agloc` template file (same format as dialogue localisation, T-DLG08 — done). Also validate at `ag build` time that every `loc_key:` used in a cutscene script is present in the project's active locale file. Tests in `tools/ag/internal/cut/`. *(independent — T-CUT09 and T-DLG08 are done)*

## Notes

- T-CUT31, T-CUT32, T-DLG05, T-CUT27, T-CUT30 are all independent — may be worked in parallel.
- T-CUT25 must complete before T-CUT26.
- T-DLG05 must complete before T-DLG06 and T-DLG19.
- T-DLG06 must complete before T-DLG19.
- T-CUT27 must complete before T-CUT28.
- After this batch: T-CUT33 (cutscene preview in editor), M-LOC remaining tasks, M11 Blender remaining tasks.
