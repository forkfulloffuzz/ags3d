# AGS3D — Current Development Tasks

This file tracks the active batch of tasks. Update status as work progresses.
When all tasks are done, ask Claude to pick the next 10.

## M-DLG — Cross-system integration (Batch 2)

- [x] **T-DLG10** — Go: `.agchar` `[dialogue]` block — parse `roots`, `inherits_globals`, `suppress_globals` fields; validate all listed node titles exist in the linked project *(depends on T-DLG03 — done)*
- [x] **T-DLG11** — Go: `.agroom` `[dialogue]` block — parse `on_enter` and `on_enter_repeat` fields (node titles); validate they exist in the linked project *(depends on T-DLG03 — done)*
- [x] **T-DLG12** — Go: `.agitem` `[dialogue]` block — parse `on_examine` and `on_use_failed` fields (node titles); validate they exist *(depends on T-DLG03 — done)*

## M-CUT — Go tooling (Batch 2)

- [x] **T-CUT08** — Go: sequencing warnings SEQ-W001..W006 in `cut/seqvalidator.go` (or new `cut/seqwarnings.go`). W001: bg step duration > 10s before sync. W002: foreground walk_to/run_to/camera move_to with no `timeout:`. W003: `<<sync>>` all with no active backgrounds at that point. W004: Blender frame tag (`.aganim` step) with no registered handler. W005: `on_fail:skip` on a step containing a critical state change (`<<action>>` / `<<set>>`). W006: `<<wait_for event:>>` where no command in the same file emits that event. Return `[]ValidationWarning`. Add tests for each code. *(depends on T-CUT07 — done)*

## M-LOC — Localisation pipeline (Batch 1)

- [x] **T-LOC01** — Design: `.agstrings` format spec — define the AGS3D native localisation format. A flat key=value file with optional metadata comments. Blocks (`[locale:en]`, `[locale:fr]`) contain `key = "value"` pairs. Header declares `base_locale` and `fallback_chain`. Write a `docs/localisation-milestone.md` milestone doc covering file format, parser API, diff semantics, and the full M-LOC task table. *(no blockers)*
- [x] **T-LOC02** — Go: `.agstrings` parser, writer, and diff engine in a new `tools/ag/internal/loc/` package. `Parse(filename, src)` → `StringsFile` (locale blocks, key→value map per locale). `Write(sf)` → canonical string. `Diff(base, updated)` → `[]DiffEntry` (added/changed/removed/stale). Tests covering round-trip, diff, and empty fallback. *(depends on T-LOC01)*

## M-CUT — Runtime (Batch 1: Event bus)

- [x] **T-CUT10** — C++: AGS3D synchronous event bus — `EventBus` class (registered as Engine singleton `AGSEventBus`). `emit(name: StringName, payload: Dictionary)` — calls all subscribers synchronously before returning. `subscribe(name: StringName, callable: Callable)` / `unsubscribe(name: StringName, callable: Callable)`. Namespaced event names (`event:{char}:{tag}`). GDScript bindings. Tests in `agstests/`. *(no blockers)*

## M-DLG — Runtime (Batch 1: Core engine)

- [x] **T-DLG14** — GDScript: runtime dialogue engine — load compiled graph from `.engine/generated/dialogue/` (JSON). `dialogue.Start(char, node_title)` and `dialogue.StartDefault(char)` as blocking coroutines (`await`). `dialogue.StartItem(item, trigger)` for item examine/use. Execute nodes: evaluate `visible_if` / `available_if` conditions against AGSRuntime state graph, present choices, follow jumps, fire `<<action>>` and `<<set>>` commands, handle `<<end>>`. Signals: `dialogue_started(node_title)`, `dialogue_ended(node_title)`, `line_spoken(char, text, emotion)`, `choices_presented(options[])`, `choice_made(index)`. *(depends on T-DLG07 — done)*
- [x] **T-DLG15** — GDScript: dialogue state tracking — `visited_nodes` (set), `seen_options` (map node→set of indices), `one_shot_consumed` (set), `node_visit_count` (map). Persisted in save graph under `dialogue_state`. `dialogue.NodeVisited(title)` → bool, `dialogue.OptionSeen(title, index)` → bool, `dialogue.VisitCount(title)` → int. Schema version field for save compatibility. *(depends on T-DLG14)*
- [x] **T-DLG16** — GDScript: dialogue presenter — `CanvasLayer`-based choice UI. Speaker name + text display. Option list rendering (visible, available, greyed-out states for `available_if` false options). Emotion tag signal dispatch (`portrait_requested(char, emotion)`). Auto-advance timer. Hold-to-advance input. Connects to dialogue engine signals. *(depends on T-DLG14)*

## Notes

- T-DLG10, T-DLG11, T-DLG12 are independent of each other and can be done in any order.
- T-CUT08 is independent and can be done alongside DLG tasks.
- T-LOC01 must precede T-LOC02; both precede the rest of M-LOC.
- T-CUT10 has no blockers and is the critical path for all M-CUT runtime (T-CUT11-T-CUT32).
- T-DLG14 is the critical path for T-DLG15, T-DLG16, T-DLG17, T-DLG18.
- T-DLG15 and T-DLG16 both depend on T-DLG14 but are independent of each other.
