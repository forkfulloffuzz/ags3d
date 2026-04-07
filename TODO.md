# AGS3D — Current Development Tasks

This file tracks the active batch of tasks. Update status as work progresses.
When all tasks are done, ask Claude to pick the next 10.

## M-CUT — Runtime (Phase 4: Fallback policies)

- [x] **T-CUT15** — GDScript: Fallback policies — per-step `on_fail:` param and per-cutscene `fallback:` header field. Five policies: `skip_and_continue` (fire state changes, mark complete, continue), `halt` (stop cutscene, report error, return to game), `log_and_continue` (log + continue), `retry_once` (retry once, then escalate to halt), `jump_to label` (fire state changes, seek to label, continue from there). Resolution order: per-step `on_fail:` → per-cutscene `fallback:` → global default `halt`. State changes (`<<action>>`, `<<set>>`) always fire regardless of policy. Integrate with the `failed` step state already emitted by the timeout mechanism. Tests in `agstests/`. *(depends on T-CUT14 — done)*

## M-LOC — Validation

- [x] **T-LOC04** — Go: `ag validate` localisation pass — three codes: DLG-LOC-E001 (key present but empty value and build is `--release`), DLG-LOC-W001 (key is stale — `// [stale]` prefix in .agstrings), DLG-LOC-W002 (key is orphaned — `// [orphan]` prefix). Scans all `.agstrings` files under `locale/`. Integrates into `ValidateProject` / `ValidateFiles`. Tests in `tools/ag/`. *(depends on T-LOC02 — done)*

## M-CUT — Format & Parser (Phase 1)

- [x] **T-CUT01** — Go: `.agcut` file parser — header fields (`title`, `skip`, `save_block`, `tags`, `fallback`, `loc_group`, `voice_session`, `audio_scope`, `duck_channels`, `duck_level`, `duck_fade`, `duck_restore`, `auto_duck`), `sequence:` body start. Token types: `COMMAND_OPEN` (`<<`), `COMMAND_CLOSE` (`>>`), `COMMAND_NAME`, `NAMED_PARAM`, `STRING_VALUE`, `IDENTIFIER`, `BLOCK_OPEN` (parallel/if/on), `BLOCK_CLOSE`. Lives in `tools/ag/internal/cut/` package. Tests for each token type and header field. *(no dependencies)*

- [x] **T-CUT04** — Go: `game.agp` `[cutscenes]` and `[input]` blocks — `[cutscenes]`: `fallback_debug`, `fallback_release`, `fallback_qa`, `step_timeout_default`. `[input]`: `dialogue_advance`, `cutscene_skip`, `dialogue_hold_advance` input action bindings. Parse into the existing `Manifest` struct in `tools/ag/internal/project/`. Tests for each field. *(no dependencies)*

- [x] **T-CUT02** — Go: Full command vocabulary parser — all command types in the `sequence:` body: camera, character, dialogue/text, audio, visual, flow/state commands; named params; `bg:id` / `id:` / `timeout:` / `on_fail:` modifiers; `<<parallel>>` / `<<end_parallel>>`, `<<if>>` / `<<else>>` / `<<end_if>>`, `<<label>>` / `<<skip_to>>`, `<<on event:>>` / `<<end_on>>` blocks. Produces `CutsceneSequence` AST in `cut` package. Tests covering each command type. *(depends on T-CUT01)*

- [x] **T-CUT03** — Go: Inline cutscene parser — extend `.agdlg` parser (`tools/ag/internal/dlg/`) to recognise `<<cutscene skip:policy>>` ... `<<end_cutscene>>` blocks. Extract the inline sequence as an embedded `CutsceneSequence` node in the dialogue AST. Tests for basic inline block, skip policy field, nested commands. *(depends on T-CUT01)*

## M-CUT — Validator (Phase 2)

- [x] **T-CUT07** — Go: Sequencing validator — errors SEQ-E001..SEQ-E007: sync references undeclared id, sync references foreground id, background step with no eventual sync, `on_fail:jump_to` references missing label, `wait_for` event never emitted in project, circular `wait_for`, duplicate step id. Lives in `tools/ag/internal/cut/` as `ValidateSequencing(seq)`. Tests for each error code. *(depends on T-CUT02)*

- [x] **T-CUT05** — Go: Cutscene format validator — errors CUT-E001..CUT-E013: title uniqueness, named point existence (from project symbol table), character existence, audio file existence, video file existence, `skip_to` label existence, `<<choice>>` in parallel, nested cutscene existence, circular nesting, animation name on character, room transition with dialogue after, `save_block:false` with state changes, identifier naming rule (`^[a-z][a-z0-9_]*$`). `ValidateFormat(lp, sym ProjectSymbolTable)` function. Tests for each error code. *(depends on T-CUT02)*

- [x] **T-CUT06** — Go: Cutscene format warnings — CUT-W001..CUT-W011: cutscene never triggered, very long with `skip:never`, state change after room transition, parallel with very different durations, voice line with no audio file, cutscene has no `<<end>>` or room transition, label never used as skip target, `author_controlled` with no labels, audio started with no reachable stop (CUT-W009), `duck:all` used (CUT-W010), `auto_duck:true` with no `duck_channels` (CUT-W011). `WarnFormat(lp)` function. Tests for each warning code. *(depends on T-CUT05)*

- [ ] **T-CUT09** — Go: Emit validated cutscene data to `.engine/generated/cutscenes/` (JSON, one file per `.agcut`). Integrate `.agcut` parsing + CUT/SEQ validation into `ag build` pipeline. Integrate CUT-E/W and SEQ-E/W codes into `ag validate` report alongside DLG codes. *(depends on T-CUT06, T-CUT07 — both from this batch; T-CUT08 — done)*

## Notes

- T-CUT15 and T-LOC04 are independent of each other and of the parser batch.
- T-CUT01 must come before T-CUT02 and T-CUT03 (both independent of each other).
- T-CUT04 is independent of T-CUT01/T-CUT02.
- T-CUT05, T-CUT07 both depend on T-CUT02 and are independent of each other.
- T-CUT06 depends on T-CUT05; T-CUT09 depends on both T-CUT06 and T-CUT07.
