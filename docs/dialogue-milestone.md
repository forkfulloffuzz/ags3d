# AGS3D Dialogue System — Milestone M-DLG

> **Status: Not started.**
> Design document: `docs/AGS3D_Dialogue_System.docx`

## Goal

A complete authored dialogue system for AGS3D. Authors write `.agdlg` files in a
Yarn Spinner-inspired plain-text format. The build system compiles them to a
binary runtime format. At runtime, dialogue plays as a blocking coroutine with
a choice UI. The system integrates with characters, rooms, items, and the save
graph. A full localisation pipeline (PO/CSV export, voice script export, RTL
support) is included.

At milestone end an author can:

1. Write `.agdlg` files and have them validated against the full project symbol
   table (characters, rooms, items, flags) at `ag build` time.
2. Call `dialogue.Start(guard, "guard_greeting")` from an AGS-spirit script;
   the conversation plays as a blocking coroutine.
3. Export strings for translation and voice recording with `ag export --locale`
   and `ag export --voicescript`.

---

## File Format (.agdlg)

```
// guard_dialogue.agdlg
title: guard_greeting
character: guard
tags: [chapter:1]
---
Guard: You there. Stop.
-> I'm just passing through. <<visible_if not flag.guard_suspicious>>
   Guard: Move along then.
   <<action flag.guard_spoken = true>>
   <<end>>
-> I have a pass. <<visible_if item.gate_pass in player.inventory>>
   Guard: Let me see that.
   <<jump guard_checks_pass>>
-> Never mind.
   <<end>>
===

title: guard_checks_pass
character: guard
---
Guard: Hmm. This looks legitimate.
Guard: You can go through.
<<action flag.gate_open = true>>
<<action room.transition("inner_courtyard")>>
===
```

### Key format features

- Node titles unique across the **entire project** (not just per file).
- Conditions use AGS-spirit expression syntax — same parser as scripts.
- `[global]` tagged nodes inherited by all characters; suppressible per character.
- `#loc:` pins stable localisation keys. Parser auto-assigns if absent.
- `inherits:` and `suppress:` rules for per-character global option overrides.

---

## Architecture

```
.agdlg files
  │
  ▼  (ag build)
Go parser (Scan → Lex → Parse → Link → Validate → Emit)
  │
  ├─ Validator: structural + cross-system errors, static analysis warnings
  ├─ Emit: .engine/generated/dialogue/ (binary, one file per .agdlg)
  └─ ag export: PO / CSV locale files, voice actor scripts
  
.engine/generated/dialogue/
  │
  ▼  (runtime)
GDScript runtime
  ├─ dialogue.Start() — blocking coroutine
  ├─ Condition evaluator (state graph)
  ├─ Choice UI presenter (CanvasLayer)
  └─ State persistence (in save graph)
```

---

## Task Breakdown

### Phase 1 — Format & Parser

| Task | Description | Depends on |
|---|---|---|
| T-DLG01 | Go: `.agdlg` lexer — token types: `HEADER_KEY`, `HEADER_VALUE`, `SEPARATOR` (`---`), `NODE_END` (`===`), `SPEAKER`, `LINE`, `OPTION`, `COMMAND`, `COMMENT`, `TAG`, `LOC_KEY`. Handles UTF-8 source. | — |
| T-DLG02 | Go: `.agdlg` parser — Stage 1 (project scan), Stage 2 (lex), Stage 3 (parse). Produces `DialogueFile` → `DialogueNode[]` AST with `DialogueLine`, `DialogueOption`, `DialogueCommand`, `DialogueCondition` node types. Parses all header fields (`title`, `character`, `tags`, `inherits`, `suppress`, `loc_id`). | T-DLG01 |
| T-DLG03 | Go: Link stage — Stage 4. Resolve all `<<jump target>>` targets, `$character.node_name` placeholders, and `[global]` option inheritance/suppression across all parsed files. Build project-wide dialogue graph. | T-DLG02 |

### Phase 2 — Validator

| Task | Description | Depends on |
|---|---|---|
| T-DLG04 | Go: Structural validator — errors DLG-E001..E010: duplicate node titles, jump target missing, node with no reachable end, character not defined, global placeholder unresolvable, suppress of non-global, malformed command expression, option with no content/jump, circular jump, loc_id collision. Blocks build. | T-DLG03 |
| T-DLG05 | Go: Cross-system validator — errors DLG-E020..E025: inventory item referenced does not exist, room referenced does not exist, character property not defined, flag never set anywhere, named point not in room, knowledge flag never granted. Requires full project symbol table from ag validate pass. | T-DLG04 |
| T-DLG06 | Go: Static analysis warnings — DLG-W001..W012: orphaned node, dead end option, condition always false/true, one-shot with no state change, global never suppressed, modified line missing loc annotation, character missing portrait, node only reachable via always-false condition, deep nesting, empty node, duplicate manual loc key. Reachability graph traversal from root nodes. Warning suppression via `// @suppress DLG-Wxxx` annotation. | T-DLG05 |

### Phase 3 — Build Pipeline & Export

| Task | Description | Depends on |
|---|---|---|
| T-DLG07 | Go: Emit stage — Stage 6. Compile validated dialogue graph to `.engine/generated/dialogue/` (JSON format per `.agdlg` source). Integrate into `ag build` pipeline. Assign automatic loc keys (`{node_title}:{line_index}:{text_hash}`) to all unannotated lines and options. | T-DLG06 |
| T-DLG08 | Go: `ag export --locale <lang>` — PO format export with stable loc keys, translator context comments (character, chapter, context type: spoken/choice/narration). `--diff` flag emits only untranslated and stale strings. CSV format option. | T-DLG07 |
| T-DLG09 | Go: `ag export --voicescript` — per-character actor scripts grouping all lines with scene context, emotion tag annotations, and preceding player line for timing reference. Grouped by `voice_session` header field. `--character <name>` and `--locale <lang>` filters. | T-DLG07 |

### Phase 4 — Cross-System Integration

| Task | Description | Depends on |
|---|---|---|
| T-DLG10 | Go: `.agchar` `[dialogue]` block — parse `roots`, `inherits_globals`, `suppress_globals` fields. `roots` lists entry point node titles for this character. Validate that all listed titles exist. | T-DLG03 |
| T-DLG11 | Go: `.agroom` `[dialogue]` block — parse `on_enter` and `on_enter_repeat` fields (node titles auto-triggered on room enter). Validate node titles exist. Integrate with room script wiring. | T-DLG03 |
| T-DLG12 | Go: `.agitem` `[dialogue]` block — parse `on_examine` and `on_use_failed` fields. Validate node titles. Wire to `item_interact` handler dispatch. | T-DLG03 |
| T-DLG13 | Go: `game.agp` `[locales]` block — locale declarations with `name` and `rtl` flag. `[localisation]` block — `base_locale`, `fallback_chain`. Validate locale codes. | — |

### Phase 5 — Runtime

| Task | Description | Depends on |
|---|---|---|
| T-DLG14 | GDScript: Runtime dialogue engine — load compiled graph from `.engine/generated/dialogue/`. Execute nodes: evaluate `visible_if` / `available_if` conditions, present choices, follow jumps. `dialogue.Start(char, node_title)` and `dialogue.StartDefault(char)` as blocking coroutines. `dialogue.StartItem(item, trigger)` for item examine/use. | T-DLG07 |
| T-DLG15 | GDScript: Dialogue state tracking — `visited_nodes` (set), `seen_options` (map), `one_shot_consumed` (set), `node_visit_count` (map). Serialised into save graph. `dialogue.NodeVisited()`, `dialogue.OptionSeen()` query API. Supports schema versioning for save compatibility. | T-DLG14 |
| T-DLG16 | GDScript: Dialogue presenter — `CanvasLayer`-based choice UI. Speaker name + text display. Option list rendering (visible, available, greyed-out states). Emotion tag dispatch (emit signal for portrait system). Auto-advance timer. Hold-to-advance input handling. | T-DLG14 |
| T-DLG17 | GDScript: Localisation runtime — `Game.SetLocale(code)` switches active locale without restart; loads translation table (PO or compiled binary); fallback chain on missing strings; RTL layout flag propagated to dialogue presenter. In-progress dialogue restarts current node on locale switch. | T-DLG13, T-DLG16 |

### Phase 6 — AGS-Spirit Bindings

| Task | Description | Depends on |
|---|---|---|
| T-DLG18 | Go: Grammar + emitter — `dialogue.Start(char, "node")`, `dialogue.StartDefault(char)`, `dialogue.NodeVisited("node")`, `dialogue.OptionSeen("node", index)`, `dialogue.StartItem(item, "trigger")`, `Game.SetLocale("code")`. All Start variants emit `await` (blocking). Query variants emit direct calls. | T-DLG14 |

### Phase 7 — Validation Integration

| Task | Description | Depends on |
|---|---|---|
| T-DLG19 | Go: Integrate dialogue validator into `ag validate` project-wide pass. Dialogue errors and warnings appear in the same report as room, character, and item errors. Provide full cross-system symbol table to dialogue validator. | T-DLG05, T-DLG06 |

---

## Validator Error & Warning Reference

### Structural Errors (block build)

| Code | Check |
|---|---|
| DLG-E001 | Duplicate node title within project |
| DLG-E002 | Jump target does not exist |
| DLG-E003 | Node has no reachable end |
| DLG-E004 | Character reference not defined |
| DLG-E005 | Global option `$character` placeholder unresolvable |
| DLG-E006 | Suppress target is not a global option |
| DLG-E007 | Malformed command expression (includes string concatenation in display context) |
| DLG-E008 | Option with no content and no jump |
| DLG-E009 | Circular jump with no exit |
| DLG-E010 | `loc_id` collision between nodes |

### Cross-System Errors (block build)

| Code | Check |
|---|---|
| DLG-E020 | Inventory item referenced in condition does not exist |
| DLG-E021 | Room referenced in action does not exist |
| DLG-E022 | Character property referenced does not exist |
| DLG-E023 | Flag referenced in condition never set anywhere in project |
| DLG-E024 | Named point referenced in dialogue action not in room |
| DLG-E025 | Knowledge flag referenced but never granted anywhere |

### Warnings (reported, suppressible)

| Code | Warning |
|---|---|
| DLG-W001 | Orphaned node (unreachable from any root) |
| DLG-W002 | Dead end option branch (only `<<end>>`, no state changes) |
| DLG-W003 | Condition always false |
| DLG-W004 | Condition always true |
| DLG-W005 | One-shot option with no state change |
| DLG-W006 | Global option never suppressed for any character |
| DLG-W007 | Line text changed since last localisation export |
| DLG-W008 | Character speaks with no portrait defined |
| DLG-W009 | Node reachable only via always-false condition |
| DLG-W010 | Deep nesting (> 4 option levels) |
| DLG-W011 | Empty node |
| DLG-W012 | Same `#loc:` key manually assigned to multiple lines |

---

## Script API Reference

```
// Blocking — script pauses until conversation ends
dialogue.Start(guard, "guard_greeting");
dialogue.StartDefault(guard);
dialogue.StartItem(item.rusty_key, "on_examine");

// State queries
dialogue.NodeVisited("guard_greeting")     // → bool
dialogue.OptionSeen("guard_greeting", 0)   // → bool

// Localisation
Game.SetLocale("fr");
```

---

## Out of Scope for This Milestone

- Visual node graph editor (M12 — Custom Editor, T-CE-DLG01+)
- Localisation view panel in editor (M12)
- Live incremental validation in editor (M12)
- Branching dialogue trees with complex state machines beyond the linear/jump model
- Voice asset management in editor (M12)
- Emotion tag animation system (depends on character portrait system, later milestone)
