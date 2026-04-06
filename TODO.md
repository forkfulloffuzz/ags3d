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
- [ ] **T-CUT04** — Go: `game.agp` `[cutscenes]` + `[input]` blocks — cutscene registry, skip-input binding *(depends on T-CUT01)*
- [ ] **T-CUT05** — Go: cutscene format validator (errors CUT-E001..E012) — unknown commands, missing labels, invalid skip values *(depends on T-CUT02)*

## Notes

- T-DLG01 and T-DLG13 have no blockers — start with both.
- M-CUT formally depends on M-DLG for inline cutscene integration, but the parser
  foundation (T-CUT01–05) is independent and can proceed in parallel.
- T-DLG05, T-DLG06 (cross-system + static analysis validators) deferred to batch 2.
- T-CUT03 (inline cutscene parser in .agdlg) deferred until T-DLG02 + T-CUT02 are done.
