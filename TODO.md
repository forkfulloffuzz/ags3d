# AGS3D — Current Development Tasks

This file tracks the active batch of tasks. Update status as work progresses.
When all tasks are done, ask Claude to pick the next 10.

## M-DLG — Dialogue System (Batch 1: Parser pipeline)

- [x] **T-DLG01** — Go: `.agdlg` lexer — token types: `HEADER_KEY`, `HEADER_VALUE`, `SEPARATOR`, `NODE_END`, `SPEAKER`, `LINE`, `OPTION`, `COMMAND`, `COMMENT`, `TAG`, `LOC_KEY`
- [x] **T-DLG13** — Go: `game.agp` `[locales]` + `[localisation]` blocks — locale declarations (`name`, `rtl`), `base_locale`, `fallback_chain` *(no blockers — can run in parallel with T-DLG01)*
- [x] **T-DLG02** — Go: `.agdlg` parser — stages 1–3 (scan, lex, parse); produces `DialogueFile` → `DialogueNode[]` AST *(depends on T-DLG01)*
- [x] **T-DLG03** — Go: link stage — resolve all `<<jump>>` targets, `$character` placeholders, global option inheritance across all files *(depends on T-DLG02)*
- [x] **T-DLG04** — Go: structural dialogue validator (errors DLG-E001..E0xx) — duplicate titles, missing jump targets, malformed headers *(depends on T-DLG02, T-DLG03)*
- [x] **T-DLG07** — Go: dialogue emit stage + `ag build` integration — write compiled dialogue to `.engine/generated/dialogue/`; wire into `ag build` pipeline *(depends on T-DLG03, T-DLG04)*

## M-CUT — Cutscene System (Batch 1: Parser foundation)

- [x] **T-CUT01** — Go: `.agcut` file parser — header block (`title`, `skip`, `save_block`, `tags`, `fallback`, `sequence:`); token types for all command forms
- [x] **T-CUT02** — Go: full command vocabulary parser — all `<<command>>` forms (character, camera, audio, visual, flow control, sync, parallel) *(depends on T-CUT01)*
- [x] **T-CUT04** — Go: `game.agp` `[cutscenes]` + `[input]` blocks — cutscene registry, skip-input binding *(depends on T-CUT01)*
- [x] **T-CUT05** — Go: cutscene format validator (errors CUT-E001..E012) — unknown commands, missing labels, invalid skip values *(depends on T-CUT02)*

## Testing Infrastructure

- [x] **T-FIXT01** — Go: testdata fixture test harness — a new `tools/ag/internal/fixtures/` package with a single `TestFixtures` test that walks `tools/ag/testdata/`, dispatches to the correct parser by file extension (`.agscript` → agscript parser, `.agdlg` → `dlg.Parse`, `.agcut` → `cut.Parse`, `.agroom` → room parser, `.agitem` → item parser — skip with `t.Skip` if no parser for that type yet), and asserts: files under `valid/` parse with zero errors, files under `invalid/` produce at least one error. Test naming follows `fixtures/<category>/<valid|invalid>/<filename>`. Update `test-all.sh` to call out this suite as a named section alongside the existing `go test ./...` run.

## M-CUT / M-DLG — Missing fields + identifier rules

- [x] **T-CUT13** — Go: add `audio_scope` (`keep`|`pause`|`stop`, default `keep`), `duck_channels` (space-separated string), `duck_level` (float64, default 0.25), `duck_fade` (float64, default 0.3), `duck_restore` (float64, default 0.5), `auto_duck` (bool, default false) fields to `CutsceneFile` struct in `cut.go`. Parse them in `parseHeaderLine`. Add tests in `cut_test.go` covering each field round-trip and the `auto_duck`/`duck_level` defaults.

- [x] **T-CUT14** — Go: implement CUT-E013 identifier naming rule (`^[a-z][a-z0-9_]*$`) in `cut/validator.go`. Check: `cf.Title`, every `<<label name>>` arg, every `bg:id` and `id:` step identifier, `<<cutscene file:name>>` ref values, `cf.LocGroup`, `cf.VoiceSession`. Reserved words (`room_music`, `room_ambient`, `all`) are exempt. Add unit tests in `validator_test.go` covering each checked position plus a clean-passing case.

- [x] **T-DLG20** — Go: implement DLG-E011 identifier naming rule (`^[a-z][a-z0-9_]*$`) in `dlg/validator.go`. Check: node `title`, every `<<jump target>>` argument, `character:` header value, `loc_id:` header value. Exempt tags (`chapter:1`, `global`) and locale codes. Add unit tests in `dlg/validator_test.go`.

- [x] **T-CUT15** — Go: add the missing test for the `<<cutscene file:name>>` named-param form of the nested cutscene reference. `validator_test.go` currently only tests the positional form (`<<cutscene ghost>>`). Add both: `<<cutscene file:does_not_exist>>` → CUT-E008, `<<cutscene file:self>>` → CUT-E009, and `<<cutscene file:existing>>` → no error.

## M-CUT — Validator (phase 2)

- [x] **T-CUT06** — Go: cutscene format warnings — implement CUT-W001..W011 in `cut/validator.go` (or a new `cut/warnings.go`). W001: cutscene title never referenced from any other file in `allTitles`. W006: sequence has no `<<end>>` and no `room.transition`. W007: `<<label>>` declared but never used as a `<<skip_to>>` target. W008: `author_controlled` skip with no `<<label>>`. W009: audio channel started (`<<music>>`, `<<ambient>>`, `<<sound>>`) with no reachable `stop` on the flat command list. W010: `duck:all` present in any command args. W011: `auto_duck:true` with empty `duck_channels`. Return as `[]ValidationWarning` (separate type from `ValidationError`). Add tests for each warning code.

- [ ] **T-CUT07** — Go: sequencing validator — implement SEQ-E001..E007 in `cut/seqvalidator.go`. Walk the flat `cf.Sequence`: collect all `bg:id` step identifiers; on `<<sync ids…>>` verify each id was declared as `bg:` (SEQ-E001) and is not a foreground `id:` (SEQ-E002); after the full walk, any `bg:id` with no covering `<<sync>>` or sequence `<<end>>` is SEQ-E003; `on_fail:jump_to:label` references checked against label set (SEQ-E004); duplicate step ids (SEQ-E007). Return `[]ValidationError`. Add tests for each code.

## M-CUT — Emit (phase 2)

- [ ] **T-CUT09** — Go: emit validated cutscene data to `.engine/generated/cutscenes/` (JSON, one file per `.agcut`). JSON schema mirrors the dialogue emit format: top-level object with `title`, `skip`, `save_block`, `tags`, `fallback`, `audio_scope`, `duck_*` fields, and a `sequence` array of command objects (`name`, `args`, `params`, `condition`, `expr`, `text`, `bg_id`, nested `body`/`else` for block commands). Wire into `ag build` pipeline (scan `.agcut` files, parse, validate all errors+warnings, emit changed files to output dir). Integrate CUT and SEQ validator results into `ag validate` report.

## M-DLG — Export pipeline

- [ ] **T-DLG08** — Go: `ag export --locale <lang>` — PO format export. Walk the linked dialogue project; for each line/option/narration with a loc key, emit a `msgid` / `msgstr` pair with translator context comment (character name, node title, line type: spoken/choice/narration). `--diff` flag emits only strings where `msgstr` is empty or whose source text hash has changed since last export. Also support `--format csv`. Output file: `locale/<lang>.po` (or `.csv`). Integrate into `ag` CLI as a subcommand.

- [ ] **T-DLG09** — Go: `ag export --voicescript` — per-character voice actor scripts. Group all `<<line character …>>` commands by `voice_session` header, then by character name. Each entry: loc key, speaker, text, the preceding player/NPC line for timing context, emotion tag if present. Markdown or plain-text output. `--character <name>` and `--locale <lang>` filters. Output: `voicescripts/<session>/<character>.md`.

## Notes

- T-FIXT01 has no blockers — start here.
- T-CUT13 and T-CUT14 are independent of each other and can be done in either order.
- T-DLG20 mirrors T-CUT14 and can be done in parallel.
- T-CUT06 and T-CUT07 are independent; both feed T-CUT09 (emit depends on validators being complete).
- T-DLG08 and T-DLG09 depend only on T-DLG07 (done) and are unblocked.
- T-CUT03 (inline cutscene parser in .agdlg) still deferred until T-DLG02 + T-CUT02 are stable.
