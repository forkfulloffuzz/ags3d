# AGS3D — Current Development Tasks

This file tracks the active batch of tasks. Update status as work progresses.
When all tasks are done, ask Claude to pick the next 10.

## M10 / M11 / M-LOC — Phase 8

Dependencies: T-GS02 (C++ AGSItem) unblocks T-GS14 → T-GS15; T-BL10 unblocks T-BL12/T-BL13; T-LOC10 → T-LOC11 → T-LOC12 → T-LOC13.

- [x] **T-E05** — Go: `ag validate` cross-reference checks. Ensure `ag validate` checks: start_room/start_character files exist, initial_camera matches Camera block in same room, SpawnPoint.character matches .agchar, WalkTo/FaceTo point names resolve, inventory item references resolve. Exit non-zero on errors. Tests in `tools/ag/internal/validate/`. *(depends T-E01, T-E03 — both done; checks 1-6 already done, check 7 (character receiver refs in .agscript method calls) added)*

- [x] **T-LOC10** — Design: author context annotations in `.agdlg` and `.agcut`. Design `#ctx:` comment syntax to attach translator notes to individual strings. Document in `docs/localisation-milestone.md`. *(independent; design complete — syntax: trailing `#ctx:` on string lines, stored in LocEntry.Ctx, exported as PO `#. ctx:` comment and CSV `context` column)*

- [x] **T-LOC11** — Design: string taxonomy and source metadata in `.agstrings`. Design `type:` (spoken/choice/narration/ui/subtitle), `char:`, `scene:` metadata fields per entry. Document in `docs/localisation-milestone.md`. *(depends T-LOC10; design complete — metadata as comment lines above key, type/char/scene/ctx fields, backward-compatible, CSV header extended)*

- [x] **T-LOC12** — Go: `ag loc` subcommand — advanced search, filter, sort, and group-by over `.agstrings` files. Commands: `ag loc find <project> --locale fr --pattern "guard_*"`, `ag loc filter <project> --locale fr --untranslated`. Tests in `tools/ag/internal/loc/`. *(depends T-LOC11; find and filter subcommands added with --pattern/--locale/--group-by and --untranslated/--char/--node/--type flags)*

- [ ] **T-LOC13** — Go: `ag loc report` — condition-based report generation. `ag loc report <project> --locale fr --by-character` shows all strings for one character; `--untranslated` shows only empty translations. Tests in `tools/ag/internal/loc/`. *(depends T-LOC12)*

- [ ] **T-GS15** — Go: grammar + emitter — `SetStatusText`, `SetActiveVerb`, `GetActiveVerb`. Emit `AGSRuntime.set_status_text("...")`, `AGSRuntime.set_active_verb("...")`, `AGSRuntime.get_active_verb()`. These are non-blocking. Tests in `tools/ag/internal/emitter/`. *(depends T-GS14 — waits on T-GS02 C++ AGSItem)*

- [ ] **T-GS19** — Go: grammar + emitter — `SetPlayerControl`, `FadeIn`, `FadeOut`, `Wait`. `SetPlayerControl(false)` emits `AGSRuntime.set_player_control(false)`. `FadeIn()`/`FadeOut()` emit `AGSCutscene.fade_in()`/`fade_out()`. `Wait(seconds)` emits `await get_tree().create_timer(seconds).timeout`. Tests in `tools/ag/internal/emitter/`. *(depends T-GS18 — GDScript cutscene runtime; Editor milestone needed first)*

- [ ] **T-BL12** — Go: `.agchar` animation clip wiring in generated `.tscn`. Parse `mesh` and `animations` fields from `.agchar`; emit `anim_idle`, `anim_walk`, `anim_talk` properties on the character root node in the `.tscn`. Tests in `tools/ag/internal/scene/`. *(depends T-BL10 — character export operator)*

- [ ] **T-BL13** — GDScript: `ags_character.gd` drives AnimationPlayer on state transitions. When velocity > 0 play `walk` clip; when `say()` called play `talk` clip; idle when stationary. Uses `anim_walk`/`anim_idle`/`anim_talk` from `.tscn` properties. Tests in `agstests/`. *(depends T-BL12)*

- [ ] **T-GS14** — GDScript: GUI runtime — `InventoryBar`, `VerbBar`, `StatusLine`. `InventoryBar` auto-populates from character inventory and refreshes on add/remove. `VerbBar` buttons call `AGSRuntime.set_active_verb()`. `StatusLine` displays text via `AGSRuntime.set_status_text()`. Tests in `agstests/`. *(depends T-GS02 — C++ AGSItem node)*

## Notes

- T-LOC10 is independent (design only).
- T-LOC11 depends on T-LOC10.
- T-LOC12 depends on T-LOC11.
- T-LOC13 depends on T-LOC12.
- T-E05 is independent (already mostly implemented).
- T-GS14/T-GS15 depend on T-GS02 (C++ AGSItem) — high priority for C++ implementor.
- T-GS19 depends on T-GS18 (GDScript cutscene runtime) — Editor milestone dependency.
- T-BL12 depends on T-BL10 (character export operator).
- T-BL13 depends on T-BL12.
- After this batch: M12 (AG Studio custom editor UI), remaining M11 Blender tasks.
