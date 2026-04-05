# AGS3D — Current Development Tasks

This file tracks the active batch of tasks. Update status as work progresses.
When all tasks are done, ask Claude to pick the next 10.

## M11 — Blender Integration (Batch 2)

- [x] **T-BL11** — Go: `ag build` — detect `.glb` in room directory; embed as sub-scene node in generated `.tscn`
- [x] **T-BL12** — Go: `.agchar` — parse `mesh` + `animations` fields; embed character `.glb` in `.tscn`; wire `AnimationPlayer` clips *(depends on T-BL10)*
- [x] **T-BL13** — GDScript: `ags_character.gd` — drive `AnimationPlayer` clips: idle/walk/talk from character state *(depends on T-BL12)*
- [x] **T-BL09** — Python: NavMesh baking — auto-bake from `WalkableSurface` objects; tag result as `AGS_NavMesh`; include in GLTF export

## M10 — Game Systems (Finish)

- [x] **T-GS14** — GDScript: GUI runtime — `InventoryBar`, `VerbBar`, `StatusLine` AutoLoad nodes driven by generated `.agui` scene *(depends on T-GS13)*
- [x] **T-GS15** — Go: grammar + emitter — `SetStatusText`, `SetActiveVerb`, `GetActiveVerb` *(depends on T-GS14)*

## M-DLG — Dialogue System (Batch 1)

- [ ] **T-DLG01** — Go: `.agdlg` lexer — token types: `HEADER_KEY`, `HEADER_VALUE`, `SEPARATOR`, `NODE_END`, `SPEAKER`, `LINE`, `OPTION`, `COMMAND`, `COMMENT`, `TAG`, `LOC_KEY`
- [ ] **T-DLG02** — Go: `.agdlg` parser — stages 1–3 (scan, lex, parse); produces `DialogueFile` → `DialogueNode[]` AST *(depends on T-DLG01)*
- [ ] **T-DLG03** — Go: link stage — resolve all `<<jump>>` targets, `$character` placeholders, global option inheritance across all files *(depends on T-DLG02)*
- [ ] **T-DLG13** — Go: `game.agp` `[locales]` + `[localisation]` blocks — locale declarations (`name`, `rtl`), `base_locale`, `fallback_chain`

## Notes

- T-BL14/T-BL15 are deferred to M12 (Custom Editor). T-BL16 remains a stub.
- T-DLG13 has no blockers — can proceed in parallel with T-DLG01–03.
- T-GS14 + T-GS15 complete M10 entirely.
