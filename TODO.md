# AGS3D — Current Development Tasks

This file tracks the active batch of tasks. Update status as work progresses.
When all tasks are done, ask Claude to pick the next 10.

## M-CUT — Runtime (Phase 3: Event bus surface)

- [x] **T-CUT11** — GDScript: Event bus AGS-spirit surface — `on_event(name)` room function hook (room receives all events while active); `cutscene.EmitEvent(name)`, `cutscene.WaitFor(event_name)` (blocking coroutine), `cutscene.OnEvent(name, handler)` (one-time); priority order: character handlers → room → cutscene → dialogue. Thin GDScript wrapper over AGSEventBus C++ singleton. Tests in `agstests/`. *(depends on T-CUT10 — done)*

## M-DLG — Validator & Runtime (Batch 3)

- [x] **T-DLG05** — Go: Cross-system dialogue validator — errors DLG-E020..E025: inventory item referenced does not exist, room referenced does not exist, character property not defined, flag never set anywhere in project, named point not in room, knowledge flag never granted. Requires full project symbol table from `ag validate` pass. Tests for each error code. *(depends on T-DLG04 — done)*
- [x] **T-DLG17** — GDScript: Localisation runtime — `Game.SetLocale(code)` switches active locale without restart; loads translation table from `.engine/generated/locale/`; fallback chain on missing strings; RTL layout flag propagated to dialogue presenter. In-progress dialogue restarts current node on locale switch. AutoLoad `AGSLocalisation`. Tests in `agstests/`. *(depends on T-DLG13 — done, T-DLG16 — done)*
- [x] **T-DLG18** — Go: AGS-spirit grammar + emitter for dialogue API — `dialogue.Start(char, "node")`, `dialogue.StartDefault(char)`, `dialogue.NodeVisited("node")` → bool, `dialogue.OptionSeen("node", index)` → bool, `dialogue.StartItem(item, "trigger")`, `Game.SetLocale("code")`. All Start variants emit `await` (blocking). Query variants emit direct calls. *(depends on T-DLG14 — done)*

## M-LOC — Localisation pipeline (Batch 2)

- [x] **T-LOC03** — Go: `ag export --locale <lang>` — PO format export from compiled dialogue JSON. Stable loc keys, translator context comments (character, chapter, context type: spoken/choice/narration). `--diff` flag emits only untranslated and stale strings. CSV format option with `--format csv`. Integrates with `.agstrings` diff engine (T-LOC02). *(depends on T-LOC02 — done)*

## M-CUT — Runtime (Phase 4: Core sequencer)

- [ ] **T-CUT12** — GDScript: Core sequencer — `AGSSequencer` AutoLoad. Step queue and active background step set. Execution loop: dequeue next step; foreground step fires and blocks; background step fires with id and is added to background set; sync point waits for named ids. Step states: `pending`, `running`, `complete`, `failed`, `skipped`. Signals: `step_started(id)`, `step_complete(id)`, `step_failed(id)`, `sequence_complete`. Runs on main game loop. Tests in `agstests/`. *(depends on T-CUT10 — done, T-CUT11)*

## M-DLG — Validator (Batch 3 continued)

- [ ] **T-DLG06** — Go: Static analysis warnings — DLG-W001..W012: orphaned node, dead end option, condition always false/true, one-shot with no state change, global never suppressed, modified line missing loc annotation, character missing portrait, node only reachable via always-false condition, deep nesting (> 4 levels), empty node, duplicate manual loc key. Reachability graph traversal from root nodes. Warning suppression via `// @suppress DLG-Wxxx`. *(depends on T-DLG05)*

## M-CUT — Runtime (Phase 4: Sequencer features)

- [ ] **T-CUT13** — GDScript: Sync points — `<<sync id1 id2>>` blocks sequencer until all named background step ids reach `complete` state. `<<sync>>` (no args) waits for all active backgrounds. Sync over already-completed steps passes immediately. Tests for: named sync, all-sync, sync-over-complete, mixed complete/pending. *(depends on T-CUT12)*
- [ ] **T-CUT14** — GDScript: Timeout mechanism — per-step `timeout:` seconds parameter. Global default from `game.agp` `step_timeout_default`. `timeout:none` for steps that must never time out (dialogue, video). On timeout: step enters `failed` state, fallback policy fires. Tests for: timeout fires, timeout:none respected, default applied. *(depends on T-CUT12)*

## M-DLG — Validation Integration

- [ ] **T-DLG19** — Go: Integrate dialogue validator into `ag validate` project-wide pass. Dialogue errors (DLG-E001..E025) and warnings (DLG-W001..W012) appear in the same structured report as room, character, and item errors. Provide full cross-system symbol table to cross-system validator (T-DLG05). *(depends on T-DLG05, T-DLG06)*

## Notes

- T-CUT11 is the critical path for all M-CUT sequencer runtime (T-CUT12 onward).
- T-CUT12 must be done before T-CUT13, T-CUT14 (both independent of each other).
- T-DLG05 must precede T-DLG06 and T-DLG19.
- T-DLG17, T-DLG18 are independent of each other and of the validator chain.
- T-LOC03 is independent of all M-DLG and M-CUT tasks in this batch.
