# AGS3D Cutscene System — Milestone M-CUT

> **Status: Not started.**
> Design documents: `docs/AGS3D_Cutscene_System.docx`, `docs/AGS3D_Cutscene_Sequencing.docx`
> Depends on: M-DLG (inline cutscenes in dialogue, `<<line>>` command routing)

## Goal

A complete authored cutscene system for AGS3D. Authors write `.agcut` files for
complex sequences and `<<cutscene>>` inline blocks inside `.agdlg` nodes for
simple in-dialogue sequences. Both use the same command vocabulary. The system
uses **event/action-driven sequencing** — each step fires when the previous
signals completion, not on a timer. Parallel execution, conditional blocks,
skip policies, and state consistency guarantees are first-class.

At milestone end an author can:

1. Write a `.agcut` file, reference named room points and characters, and have
   it fully validated at `ag build` time.
2. Call `cutscene.Play("chapter1_opening")` from an AGS-spirit script; the
   sequence runs, handles skip input, and returns when complete.
3. Tag Blender animation frames with named events; those events fire into the
   AGS3D event bus mid-playback.
4. Saving is automatically blocked during any cutscene, with queued saves.

---

## File Format (.agcut)

```
// cutscenes/chapter1_opening.agcut
title: chapter1_opening
skip:         after_first_view
save_block:   true
tags:         [chapter:1, cinematic]
fallback:     halt
audio_scope:  stop

// Dialogue ducking — applied automatically to every <<line>> and <<dialogue>> call
duck_channels: room_music room_ambient
duck_level:    0.25
duck_fade:     0.3
duck_restore:  0.5
auto_duck:     true

sequence:
	<<fade_in duration:2.0>>
	<<camera set point.rooftop_wide fov:60>>
	<<music theme_main fade_in:3.0>>
	<<title_card "Chapter One: The Market District" duration:2.0>>
	<<camera move_to point.street_level duration:4.0 ease:out bg:cam_move>>
	<<character player spawn_at point.alley_entrance>>
	<<sync cam_move>>
	<<character player animation play:look_around>>
	<<line narrator "Three years. And it still felt foreign.">>
	<<action flag.chapter1_started = true>>
	<<action room.transition("market")>>
	<<fade_out duration:1.0>>
```

### Identifier naming rule

All author-assigned identifiers must match `^[a-z][a-z0-9_]*$` — lowercase letters,
digits, and underscores only; must start with a letter. This applies to:

- `title:` value
- `<<label name>>` and `<<skip_to name>>` arguments
- `bg:id` / `id:` step identifiers referenced by `<<sync>>`
- `<<cutscene file:name>>` nested reference
- `loc_group:` and `voice_session:` header values
- Named audio channels (other than the reserved `room_music` / `room_ambient`)

Reserved identifiers (`room_music`, `room_ambient`, `all`) are exempt. Tags
(`chapter:1`, `cinematic`) are metadata and are not subject to this rule.
Locale codes follow BCP 47 and are also exempt.

Violations are a hard build error (CUT-E013).

### Key format features

- `skip:` — one of `always`, `never`, `after_first_view`, `author_controlled`.
- `save_block:` — default `true`; `false` only for ambient/decorative sequences with no state changes.
- `bg:id` on any step — fires and continues immediately; step runs concurrently.
- `<<sync id1 id2>>` / `<<sync>>` — wait for named (or all) background steps.
- `<<parallel>>` / `<<end_parallel>>` — all enclosed steps start simultaneously.
- `<<if condition>>` / `<<else>>` / `<<end_if>>` — same AGS-spirit expression syntax.
- `<<label name>>` / `<<skip_to name>>` — author-defined skip targets.
- `<<on event:char:tag>>` / `<<end_on>>` — react to event bus events mid-sequence.
- Inline mode: `<<cutscene skip:always>>` ... `<<end_cutscene>>` inside `.agdlg`.
- `audio_scope:` — controls room audio at cutscene boundaries (see below).
- `duck_channels` / `duck_level` / `duck_fade` / `duck_restore` / `auto_duck` — dialogue ducking defaults (see below).

---

### Audio scope — room audio at cutscene boundaries

The cutscene is the score that joins all active elements. Room music and ambient audio
are started outside the cutscene by the room script, so the sequencer cannot track them
as normal channels. `audio_scope:` declares the relationship at the boundary:

| Value | Behaviour |
|-------|-----------|
| `keep` *(default)* | Room audio continues uninterrupted. Cutscene layers on top. No auto-restore on end. |
| `pause` | Room music + ambient are paused at cutscene start; resumed (with `duck_restore` fade) on any end path — normal `<<end>>`, room transition, skip, or fallback halt. |
| `stop` | Room music + ambient are stopped (with optional `fade_out:` on the header) at cutscene start. Not restored — the room's `on_enter` handler is expected to set new audio state. |

Authors can also reference live room audio directly in commands using the reserved channel
identifiers `room_music` and `room_ambient`:

```
// Manually crossfade room music into cutscene music
<<music stop channel:room_music fade_out:2.0 bg:fade_room>>
<<music cinematic_theme fade_in:2.0 bg:fade_cut>>
<<sync fade_room fade_cut>>

// Duck room ambient for a tense line, then restore it
<<ambient volume channel:room_ambient value:0.15 duration:0.3>>
<<line player "Did you hear that?">>
<<ambient volume channel:room_ambient value:1.0 duration:0.5>>
```

`channel:room_music` and `channel:room_ambient` are resolved at runtime to whatever the room's
audio slots currently hold. CUT-W009 (leaked audio) does **not** fire for these channels — the
room owns them.

---

### Dialogue ducking

The cutscene header can declare default ducking behaviour that applies automatically to every
`<<line>>` and `<<dialogue>>` command in the sequence. This is a score-level concern: the
`.agdlg` file has no knowledge of what audio is active, so ducking configuration belongs in
`.agcut` where the layers are composed.

**Header fields:**

| Field | Default | Description |
|-------|---------|-------------|
| `duck_channels` | *(none)* | Space-separated list of channels to duck. Accepts `room_music`, `room_ambient`, any named `bg:id` started earlier, or `all` for every active channel. Required when `auto_duck: true`. |
| `duck_level` | `0.25` | Target volume multiplier (0.0–1.0) during the line. |
| `duck_fade` | `0.3` | Ramp duration into duck (seconds). Completes before the line begins. |
| `duck_restore` | `0.5` | Ramp duration out of duck (seconds). Runs as background after line completes. |
| `auto_duck` | `true` if `duck_channels` set, else `false` | When `true`, every `<<line>>` and `<<dialogue>>` auto-ducks. When `false`, ducking only fires on explicit per-call `duck:` arguments. |

**Per-call override on `<<line>>` and `<<dialogue>>`:**

```
// Uses header defaults (auto_duck: true)
<<line narrator "The rain hammers the roof.">>

// Override channels and level for this line only
<<line player "I can barely hear you." duck:room_ambient duck_level:0.1>>

// Silence the room during this line — override fade timings too
<<line narrator "Listen." duck:room_music,room_ambient duck_level:0.0 duck_fade:0.1>>

// Suppress auto-duck — music stays full (intentional, story beat)
<<line narrator "The silence was deafening." duck:none>>

// Carry defaults into a full dialogue sequence
<<dialogue node:guard_confrontation duck:room_music,room_ambient duck_level:0.2 duck_restore:1.0>>
```

**`duck:all` shorthand** ducks every channel currently active (including mid-cutscene started
tracks). It is allowed but triggers CUT-W010 at build time because the set of active channels
depends on runtime state — the validator cannot verify the intent is correct.

**Runtime execution order (T-CUT32):**
1. Ramp listed channels from current volume → `duck_level × current_volume` over `duck_fade` s (foreground — line waits until ramp completes).
2. Fire `<<line>>` or `<<dialogue>>` to completion.
3. Ramp channels back to pre-duck volume over `duck_restore` s (background — next step starts immediately).

Duck ramps are internal operations and do not count as stopping a channel for CUT-W009 purposes.

---

## Architecture

```
.agcut files + .agdlg inline blocks
  │
  ▼  (ag build)
Go parser (header + command vocabulary + validation)
  │
  ├─ Validator: CUT-E/W (format) + SEQ-E/W (sequencing)
  └─ Emit: .engine/generated/cutscenes/ (JSON, one file per .agcut)

Blender add-on
  └─ Frame tag export → .aganim sidecar (per character)
       │
       ▼  (ag build)
       Go: inject AnimationPlayer method-call tracks into .tscn

.engine/generated/cutscenes/
  │
  ▼  (runtime — GDScript)
Event bus (C++ sync pub/sub)
  │
Sequencer
  ├─ Foreground step queue
  ├─ Background step set (with ids)
  ├─ Sync point waiting
  ├─ Timeout + fallback policies
  └─ Skip system (dry-run state-change pass)
  │
Command executors
  ├─ Camera, Character, Audio, Visual, Flow, Dialogue
  └─ Each emits typed completion signal on done
```

---

## Task Breakdown

### Phase 1 — Format & Parser

| Task | Description | Depends on |
|---|---|---|
| T-CUT01 | Go: `.agcut` file parser — header fields (`title`, `skip`, `save_block`, `tags`, `fallback`, `loc_group`, `voice_session`, `audio_scope`, `duck_channels`, `duck_level`, `duck_fade`, `duck_restore`, `auto_duck`), `sequence:` body start. Token types for command vocabulary: `COMMAND_OPEN` (`<<`), `COMMAND_CLOSE` (`>>`), `COMMAND_NAME`, `NAMED_PARAM`, `STRING_VALUE`, `IDENTIFIER`, `BLOCK_OPEN` (parallel/if/on), `BLOCK_CLOSE`. | — |
| T-CUT02 | Go: Full command vocabulary parser — all command types (camera, character, dialogue/text, audio, visual, flow/state), named parameters, `bg:id` / `id:` / `timeout:` / `on_fail:` modifiers. Parse `<<parallel>>` / `<<end_parallel>>`, `<<if>>` / `<<else>>` / `<<end_if>>`, `<<label>>` / `<<skip_to>>`, `<<on event:>>` / `<<end_on>>` blocks. | T-CUT01 |
| T-CUT03 | Go: Inline cutscene parser — extend `.agdlg` parser (T-DLG02) to recognise `<<cutscene skip:policy>>` ... `<<end_cutscene>>` blocks. Extract inline sequence as embedded `CutsceneSequence` node in the dialogue AST. | T-CUT01, T-DLG02 |
| T-CUT04 | Go: `game.agp` `[cutscenes]` block — `fallback_debug`, `fallback_release`, `fallback_qa`, `step_timeout_default`. `[input]` block — `dialogue_advance`, `cutscene_skip`, `dialogue_hold_advance` input action bindings. | — |

### Phase 2 — Validator

| Task | Description | Depends on |
|---|---|---|
| T-CUT05 | Go: Cutscene format validator — errors CUT-E001..E013 (see reference below): title uniqueness, named point existence, character existence, audio file existence, video file existence, skip_to label existence, `<<choice>>` in parallel, nested cutscene existence, circular nesting, animation name on character, room transition with dialogue after, `save_block:false` with state changes, identifier naming rule (`^[a-z][a-z0-9_]*$`). | T-CUT02, T-DLG07 |
| T-CUT06 | Go: Cutscene format warnings — CUT-W001..W010: cutscene never triggered, very long with `skip:never`, state change after room transition, parallel with very different durations, voice line with no audio file, cutscene has no `<<end>>` or room transition, label never used as skip target, `author_controlled` with no labels, audio started with no reachable stop (leaked audio), `duck:all` used (unverifiable channel set), `auto_duck:true` with no `duck_channels` declared. | T-CUT05 |
| T-CUT07 | Go: Sequencing validator — errors SEQ-E001..E007: sync references undeclared id, sync references foreground id, background step with no eventual sync, `on_fail:jump_to` references missing label, `wait_for` event never emitted in project, circular `wait_for`, duplicate step id. | T-CUT02 |
| T-CUT08 | Go: Sequencing warnings — SEQ-W001..W006: background step with long duration before sync, foreground step with no timeout on long-running operation, sync-all with already-completed backgrounds, Blender frame tag with no registered handler, `on_fail:skip` on step with critical state change, `wait_for` event may never fire in context. | T-CUT07 |
| T-CUT09 | Go: Emit validated cutscene data to `.engine/generated/cutscenes/` (JSON). Integrate into `ag build` pipeline. Integrate CUT/SEQ validators into `ag validate` report. | T-CUT06, T-CUT08 |

### Phase 3 — Event Bus

| Task | Description | Depends on |
|---|---|---|
| T-CUT10 | C++: AGS3D event bus — synchronous pub/sub (`EventBus::emit(name, payload)`, `subscribe(name, fn)`, `unsubscribe(name, fn)`). Namespaced event names (`event:{char}:{tag}`). All handlers fire before `emit()` returns. Registered as Engine singleton. | — |
| T-CUT11 | GDScript: Event bus AGS-spirit surface — `on_event(name)` room function hook (room receives all events while active); `cutscene.EmitEvent(name)`, `cutscene.WaitFor(event_name)` (blocking), `cutscene.OnEvent(name, handler)` (one-time); priority order: character handlers → room → cutscene → dialogue. | T-CUT10 |

### Phase 4 — Sequencer

| Task | Description | Depends on |
|---|---|---|
| T-CUT12 | GDScript: Core sequencer — step queue, active background step set. Execution loop: dequeue next step; if foreground, fire and set as current; if background, fire and add to background set with id, dequeue again; if sync point, wait for named ids. Step states: `pending`, `running`, `complete`, `failed`, `skipped`. Runs on main game loop — no threads. | T-CUT10 |
| T-CUT13 | GDScript: Sync points — `<<sync id1 id2>>` blocks sequencer until all named background ids are in `complete` state. `<<sync>>` (no args) waits for all active backgrounds. Sync over already-completed steps passes immediately. | T-CUT12 |
| T-CUT14 | GDScript: Timeout mechanism — per-step `timeout:` parameter (seconds). Global default from `game.agp` `step_timeout_default`. `timeout:none` for steps that must never time out (dialogue, video). On timeout: step enters `failed` state, fallback policy fires. | T-CUT12 |
| T-CUT15 | GDScript: Fallback policies — `skip_and_continue` (fire state changes, mark complete, continue), `halt` (stop cutscene, report error, return to game), `log_and_continue` (log + continue), `retry_once` (retry once, then escalate), `jump_to label` (fire state changes, jump to label). Per-step `on_fail:` → per-cutscene `fallback:` → global `game.agp` → build-type default. State changes always fire regardless of policy. | T-CUT14 |

### Phase 5 — Command Executors

| Task | Description | Depends on |
|---|---|---|
| T-CUT16 | GDScript: Camera commands — `<<camera set point fov? rotation?>>` (instant), `<<camera move_to point duration ease?>>` (tween), `<<camera look_at target duration? ease?>>`, `<<camera follow character offset? duration?>>`, `<<camera shake intensity duration falloff?>>`, `<<camera fov value duration? ease?>>`, `<<camera return duration? ease?>>`. Each emits `camera_complete` on finish. Delegates to `AGSCamera` C++ node. | T-CUT12 |
| T-CUT17 | GDScript: Character commands — `<<character X walk_to point speed? timeout?>>`, `<<character X run_to point>>`, `<<character X animation play:name loop? blend?>>`, `<<character X face_to target duration?>>`, `<<character X spawn_at point>>`, `<<character X hide>>`, `<<character X show>>`, `<<character X expression name>>`, `<<character X move_speed value>>`. Completion signals: `walk_complete(X)`, `run_complete(X)`, `animation_complete(X, name)`, `face_complete(X)`. Resolves character name via `AGSRuntime.get_character()`. | T-CUT12 |
| T-CUT18 | GDScript: Audio commands — `<<music name fade_in? volume? loop?>>`, `<<music stop fade_out?>>`, `<<music stop channel:room_music fade_out?>>`, `<<sound name volume? fade_in? position?>>` (optional spatial), `<<ambient name fade_in? volume?>>`, `<<ambient stop name fade_out?>>`, `<<ambient stop channel:room_ambient fade_out?>>`, `<<ambient volume channel:name value duration?>>`, `<<voice character file loc_key?>>`. Sequencer tracks every channel started by name at dispatch time for T-CUT31 leak cleanup. `channel:room_music` and `channel:room_ambient` resolve at runtime to the room's live audio slots; CUT-W009 does not fire for these. On skip: cutscene-owned audio fades out in 0.3s; room channels respect `audio_scope:` boundary. | T-CUT12 |
| T-CUT19 | GDScript: Visual commands — `<<fade_in duration? color?>>`, `<<fade_out duration? color?>>`, `<<flash color? duration?>>`, `<<vignette intensity duration?>>`, `<<letterbox enable duration?>>`, `<<overlay image fade_in? fade_out? duration?>>`, `<<video file skip:end_video?>>`. Each emits typed completion signal. `CanvasLayer`-based overlays. | T-CUT12 |
| T-CUT20 | GDScript: Flow and state commands — `<<wait seconds skip:skip_wait?>>`, `<<action expression>>` (evaluates AGS-spirit expression; always fires, cannot be skipped), `<<set variable=value>>`, `<<parallel>>` / `<<end_parallel>>` block executor (all enclosed steps fire simultaneously; block completes when longest completes unless `<<parallel_end_on_first>>`), `<<if>>` / `<<else>>` / `<<end_if>>` evaluator, `<<cutscene file:name>>` nested playback, `<<label name>>`, `<<skip_to name>>`, `<<end>>`, `<<on event:>>` / `<<end_on>>` reactive handler registration. | T-CUT12, T-CUT11 |
| T-CUT21 | GDScript: Dialogue commands — `<<line character "text" loc_key? skip:advance? duck:channels? duck_level? duck_fade? duck_restore?>>`, `<<line narrator "text" loc_key?>>`, `<<title_card "text" duration? fade_in? fade_out?>>`, `<<subtitle "text" duration?>>`, `<<choice options[] skip:skip_choices?>>` inline choice, `<<dialogue node:title duck:channels? duck_level? duck_restore?>>`, `<<dialogue file:name>>`. Per-call `duck:` args override header duck defaults; `duck:none` suppresses auto-duck for that call. Delegates to dialogue presenter (T-DLG16). `dialogue_complete` signal when player advances. | T-CUT12, T-CUT32, T-DLG16 |
| T-CUT31 | GDScript: Audio leak cleanup — the sequencer tracks every audio channel started by the cutscene (`music`, `ambient`, `sound`) at step dispatch time. On sequence completion (normal `<<end>>`, room transition, skip, or fallback halt) any channel that was started but never explicitly stopped by a matching `stop` command is stopped **immediately with no fade**. The hard cut is intentional: it makes leaked audio noticeable during development rather than silently persisting into gameplay. The validator issues CUT-W009 at build time for each reachable sequence path where a channel is started with no stop, so authors can fix leaks before shipping. Cleanup fires after the last `<<action>>`/`<<set>>` dry-run pass so state is always consistent before audio is silenced. `audio_scope: pause` additionally resumes room channels (with `duck_restore` fade) on the same completion path. | T-CUT18 |
| T-CUT32 | GDScript: Dialogue ducking — reads `duck_channels`, `duck_level`, `duck_fade`, `duck_restore`, `auto_duck` from the parsed cutscene header. When `auto_duck: true` (or per-call `duck:` is set), wraps every `<<line>>` and `<<dialogue>>` execution in a duck/restore cycle: (1) ramp listed channels to `duck_level × current_volume` over `duck_fade` s (foreground; line waits for ramp completion); (2) play line/dialogue to completion; (3) ramp channels back over `duck_restore` s (background; next step starts immediately). Per-call `duck:channels duck_level? duck_fade? duck_restore?` overrides header defaults for that step. `duck:none` suppresses the ramp entirely. `duck:all` ducks every currently active channel. `channel:room_music` / `channel:room_ambient` resolved via `AGSRuntime` audio slots. Duck ramps are internal and do not affect CUT-W009 leak tracking. | T-CUT18, T-CUT21 |

### Phase 6 — Skip System

| Task | Description | Depends on |
|---|---|---|
| T-CUT22 | GDScript: Skip input routing — single press (advance one line / trigger element skip), hold (rapid dialogue advance), double press or dedicated button (cutscene-level skip). Input actions from `game.agp` `[input]`. Engine-level input handling, not per-script. | T-CUT04 |
| T-CUT23 | GDScript: Skip system — 4 cutscene-level policies: `always` (skip at any time), `never` (no skip input processed), `after_first_view` (locked first play, unlocked after), `author_controlled` (only at explicit `<<label>>` positions). Per-element skip semantics: `walk_to` teleports character; `animation` cuts to end frame; `camera move_to` teleports camera; `wait` skips entirely; `action`/`set` cannot be skipped (always fires); `music`/`sound` fade out in 0.3s; `video` cuts to end; `choice` selects default option if `skip:skip_choices` set. | T-CUT22 |
| T-CUT24 | GDScript: State consistency guarantee — when skip is triggered to a label or cutscene end, perform a **dry-run pass** over all skipped steps: collect and immediately fire every `<<action>>` and `<<set>>` command between skip origin and destination before rendering the skip destination. State is always consistent on arrival. | T-CUT23 |

### Phase 7 — Save System Integration

| Task | Description | Depends on |
|---|---|---|
| T-CUT25 | GDScript: Automatic save blocking — any active cutscene blocks `AGSRuntime.save_game()`. Exception: `save_block: false` header (for ambient-only sequences with no state changes). Queued save: save attempt during a blocked cutscene is stored and fires automatically on cutscene completion. Brief player notification on queue. | T-CUT12 |
| T-CUT26 | GDScript: Cutscene state in save graph — `viewed_cutscenes` (set of titles seen to completion), `skipped_cutscenes` (set), `cutscene_view_count` (map title → int). Serialised with save graph. `after_first_view` policy checks `viewed_cutscenes` before enabling skip. Available to scripts: `cutscene.Viewed(name)`, `cutscene.Skipped(name)`, `cutscene.ViewCount(name)`. Usable in dialogue conditions: `<<visible_if cutscene.Viewed("intro")>>`. | T-CUT25 |

### Phase 8 — Blender Frame Tags

| Task | Description | Depends on |
|---|---|---|
| T-CUT27 | Python (Blender): Frame tag export — read `action.pose_markers` from NLA tracks on character armature during character export (T-BL10). Output `.aganim` sidecar file alongside `.glb`: `[frame_tags]` block mapping frame number → tag name string. Tag names must match AGS-spirit event name conventions. | T-BL10 |
| T-CUT28 | Go: `ag build` — read `.aganim` sidecar for each character `.glb`. Inject `AnimationPlayer` method-call tracks into character `.tscn` at the correct frames; emitted call: `_on_anim_event("event:{char_name}:{tag_name}")`. Runtime script forwards to `EventBus::emit()`. Validator checks all frame tags against registered handlers (SEQ-W004). | T-CUT27, T-CUT10 |

### Phase 9 — AGS-Spirit Bindings

| Task | Description | Depends on |
|---|---|---|
| T-CUT29 | Go: Grammar + emitter — `cutscene.Play("name")` (blocking, emits `await`), `cutscene.PlayAsync("name")` (non-blocking), `cutscene.Stop()`, `cutscene.Viewed("name")` → bool, `cutscene.Skipped("name")` → bool, `cutscene.ViewCount("name")` → int, `cutscene.IsPlaying()` → bool, `cutscene.SetSkipPolicy("name", policy)`, `cutscene.EmitEvent("name")`, `cutscene.WaitFor("event")` (blocking). | T-CUT12 |

### Phase 10 — Localisation

| Task | Description | Depends on |
|---|---|---|
| T-CUT30 | Go: Cutscene localisation — `<<line>>` and `<<title_card>>` commands participate in the loc key pipeline using cutscene title as namespace (`{cutscene_title}:{line_index}:{text_hash}`). `#loc:` manual pin supported. Lines appear in `ag export --locale` output. `voice_session` header groups all cutscene lines in `ag export --voicescript`. Title cards, subtitles, and inline choices also receive loc keys. | T-CUT09, T-DLG08 |

---

## Validator Error & Warning Reference

### Cutscene Format Errors (block build)

| Code | Check |
|---|---|
| CUT-E001 | Cutscene title not unique across project |
| CUT-E002 | Named point referenced does not exist in room |
| CUT-E003 | Character referenced not defined |
| CUT-E004 | Audio file referenced does not exist |
| CUT-E005 | Video file referenced does not exist |
| CUT-E006 | `skip_to` label target does not exist in sequence |
| CUT-E007 | `<<choice>>` inside `<<parallel>>` block |
| CUT-E008 | Nested cutscene reference does not exist |
| CUT-E009 | Circular cutscene nesting |
| CUT-E010 | Animation name not defined on character |
| CUT-E011 | Room transition inside inline cutscene with dialogue after |
| CUT-E012 | `save_block:false` on cutscene containing state changes (`<<action>>` / `<<set>>`) |
| CUT-E013 | Identifier does not match `^[a-z][a-z0-9_]*$` — applies to `title`, `<<label>>`, `bg:id`, `<<cutscene file:>>`, `loc_group`, `voice_session`, and named channel identifiers |

### Cutscene Format Warnings

| Code | Warning |
|---|---|
| CUT-W001 | Cutscene never triggered from any dialogue, script, or room file |
| CUT-W002 | Very long cutscene (> 60s estimated) with `skip:never` |
| CUT-W003 | State change after room transition (may not execute) |
| CUT-W004 | Parallel block with very different step durations (likely unintentional) |
| CUT-W005 | `<<line>>` command with no voice audio file for loc key |
| CUT-W006 | Cutscene has no `<<end>>` or room transition |
| CUT-W007 | `<<label>>` defined but never used as `<<skip_to>>` target |
| CUT-W008 | `author_controlled` skip with no `<<label>>` commands |
| CUT-W009 | Audio channel started (`<<music>>`, `<<ambient>>`, `<<sound>>`) with no reachable `stop` on at least one sequence path — channel will be hard-stopped at sequence end by the runtime |
| CUT-W010 | `duck:all` used on a `<<line>>` or `<<dialogue>>` call — the set of active channels is determined at runtime and cannot be validated at build time |
| CUT-W011 | `auto_duck: true` declared but `duck_channels` not set — no channels will be ducked |

### Sequencing Errors (block build)

| Code | Check |
|---|---|
| SEQ-E001 | `<<sync>>` references id that was never declared |
| SEQ-E002 | `<<sync>>` references a foreground step id (not a background) |
| SEQ-E003 | Background step with no eventual `<<sync>>` or sequence end (leaked step) |
| SEQ-E004 | `on_fail:jump_to` references label not in sequence |
| SEQ-E005 | `<<wait_for event:>>` — event never emitted anywhere in project |
| SEQ-E006 | Circular `wait_for` (step A waits for event only emitted after A completes) |
| SEQ-E007 | Duplicate step id in same sequence |

### Sequencing Warnings

| Code | Warning |
|---|---|
| SEQ-W001 | Background step with very long estimated duration before sync |
| SEQ-W002 | Foreground step with no `timeout:` on long-running operation (walk_to, etc.) |
| SEQ-W003 | `<<sync>>` (all) — some background steps may already be complete before sync |
| SEQ-W004 | Blender frame tag has no registered handler anywhere in project |
| SEQ-W005 | `on_fail:skip` on step containing a story-critical state change |
| SEQ-W006 | `<<wait_for event:>>` — event may never fire in this cutscene's room context |

---

## Script API Reference

```
// Blocking
cutscene.Play("chapter1_opening");
cutscene.WaitFor("event:char_b:char_b_hide");

// Non-blocking
cutscene.PlayAsync("ambient_rain_loop");

// Control
cutscene.Stop();
cutscene.SetSkipPolicy("guard_reveals_secret", "always");

// State queries
cutscene.Viewed("chapter1_opening")   // → bool
cutscene.Skipped("chapter1_opening")  // → bool
cutscene.ViewCount("intro")           // → int
cutscene.IsPlaying()                  // → bool

// Event bus
cutscene.EmitEvent("event:player:found_clue");
cutscene.OnEvent("event:char_b:char_b_hide", func(): ...);
```

---

## Completion Signal Reference

| Step type | Signal | Fires when |
|---|---|---|
| `<<character X walk_to>>` | `walk_complete(X)` | Character arrives at destination |
| `<<character X run_to>>` | `run_complete(X)` | Character arrives at destination |
| `<<character X animation>>` | `animation_complete(X, name)` | Animation reaches last frame |
| `<<character X face_to>>` | `face_complete(X)` | Rotation tween finishes |
| `<<camera move_to>>` | `camera_complete` | Camera reaches target |
| `<<camera shake>>` | `camera_shake_complete` | Shake duration elapsed |
| `<<line>>` | `dialogue_complete` | Player advances or auto-advance fires |
| `<<sound>>` | `sound_complete(name)` | Clip finishes (non-looped) |
| `<<fade_in / fade_out>>` | `fade_complete` | Fade reaches target opacity |
| `<<video>>` | `video_complete` | Playback ends or is skipped |
| `<<title_card>>` | `title_complete` | Duration elapsed or dismissed |
| `<<wait>>` | `wait_complete` | Duration elapsed |
| `<<action>>` | `action_complete` | Expression evaluated and applied |
| `<<anim_event>>` | `event:{tag_name}` | Blender frame tag reached |

---

## Out of Scope for This Milestone

- Visual timeline / sequence editor (M12 — Custom Editor)
- Video encoding or video import pipeline
- Real-time preview of camera moves in Blender (IDEA category)
- Dialogue tree editor that embeds cutscene blocks visually (M12)
- Voice recording session management tool (M12 or later)
