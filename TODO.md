# AGS3D — Current Development Tasks

This file tracks the active batch of tasks. Update status as work progresses.
When all tasks are done, ask Claude to pick the next 10.

## M-CUT — Runtime (Phase 5: Command execution)

All tasks in this batch depend on T-CUT12 (core sequencer — done).

- [ ] **T-CUT17** — GDScript: Character commands — `<<character X walk_to point speed? timeout?>>`, `<<character X run_to point>>`, `<<character X animation play:name loop? blend?>>`, `<<character X face_to target duration?>>`, `<<character X spawn_at point>>`, `<<character X hide>>`, `<<character X show>>`, `<<character X expression name>>`, `<<character X move_speed value>>`. Completion signals: `walk_complete(X)`, `run_complete(X)`, `animation_complete(X, name)`, `face_complete(X)`. Resolves character name via `AGSRuntime.get_character()`. Tests in `agstests/`. *(depends on T-CUT12 — done)*

- [ ] **T-CUT16** — GDScript: Camera commands — `<<camera set point fov? rotation?>>` (instant), `<<camera move_to point duration ease?>>` (tween), `<<camera look_at target duration? ease?>>`, `<<camera follow character offset? duration?>>`, `<<camera shake intensity duration falloff?>>`, `<<camera fov value duration? ease?>>`, `<<camera return duration? ease?>>`. Each emits `camera_complete` on finish. Delegates to `AGSCamera` C++ node. Tests in `agstests/`. *(depends on T-CUT12 — done)*

- [ ] **T-CUT18** — GDScript: Audio commands — `<<music name fade_in? volume? loop?>>`, `<<music stop fade_out?>>`, `<<music stop channel:room_music fade_out?>>`, `<<sound name volume? fade_in? position?>>` (optional spatial), `<<ambient name fade_in? volume?>>`, `<<ambient stop name fade_out?>>`, `<<ambient stop channel:room_ambient fade_out?>>`, `<<ambient volume channel:name value duration?>>`, `<<voice character file loc_key?>>`. Sequencer tracks every channel started for T-CUT31 leak cleanup. On skip: cutscene-owned audio fades out in 0.3s; room channels respect `audio_scope:` boundary. Tests in `agstests/`. *(depends on T-CUT12 — done)*

- [ ] **T-CUT19** — GDScript: Visual commands — `<<fade_in duration? color?>>`, `<<fade_out duration? color?>>`, `<<flash color? duration?>>`, `<<vignette intensity duration?>>`, `<<letterbox enable duration?>>`, `<<overlay image fade_in? fade_out? duration?>>`, `<<video file skip:end_video?>>`. Each emits typed completion signal. `CanvasLayer`-based overlays. Tests in `agstests/`. *(depends on T-CUT12 — done)*

- [ ] **T-CUT20** — GDScript: Flow and state commands — `<<wait seconds skip:skip_wait?>>`, `<<action expression>>` (evaluates AGS-spirit expression; always fires, cannot be skipped), `<<set variable=value>>`, `<<parallel>>` / `<<end_parallel>>` block executor (all enclosed steps fire simultaneously; completes when longest completes unless `<<parallel_end_on_first>>`), `<<if>>` / `<<else>>` / `<<end_if>>` evaluator, `<<cutscene file:name>>` nested playback, `<<label name>>`, `<<skip_to name>>`, `<<end>>`, `<<on event:>>` / `<<end_on>>` reactive handler registration. Tests in `agstests/`. *(depends on T-CUT12, T-CUT11 — both done)*

- [ ] **T-CUT21** — GDScript: Dialogue commands — `<<line character "text" loc_key? skip:advance?>>`, `<<line narrator "text" loc_key?>>`, `<<title_card "text" duration? fade_in? fade_out?>>`, `<<subtitle "text" duration?>>`, `<<choice options[] skip:skip_choices?>>` inline choice, `<<dialogue node:title>>`, `<<dialogue file:name>>`. Delegates to dialogue presenter (T-DLG16 — done). `dialogue_complete` signal when player advances. Duck support (`duck:channels duck_level? duck_fade? duck_restore?`) is implemented when T-CUT32 is done but the command handler runs without it. Tests in `agstests/`. *(depends on T-CUT12 — done; T-DLG16 — done)*

- [ ] **T-CUT29** — Go: Grammar + emitter — `cutscene.Play("name")` (blocking, emits `await`), `cutscene.PlayAsync("name")` (non-blocking), `cutscene.Stop()`, `cutscene.Viewed("name")` → bool, `cutscene.Skipped("name")` → bool, `cutscene.ViewCount("name")` → int, `cutscene.IsPlaying()` → bool, `cutscene.SetSkipPolicy("name", policy)`, `cutscene.EmitEvent("name")`, `cutscene.WaitFor("event")` (blocking). Tests in `tools/ag/`. *(depends on T-CUT12 — done)*

- [ ] **T-CUT22** — GDScript: Skip input routing — single press (advance one line / trigger element skip), hold (rapid dialogue advance), double press or dedicated button (cutscene-level skip). Input actions from `game.agp` `[input]` (T-CUT04 — done). Engine-level input handling, not per-script. Tests in `agstests/`. *(depends on T-CUT04 — done)*

- [ ] **T-CUT23** — GDScript: Skip system — 4 cutscene-level policies: `always`, `never`, `after_first_view`, `author_controlled` (only at `<<label>>` positions). Per-element skip semantics: `walk_to` teleports; `animation` cuts to end frame; `camera move_to` teleports; `wait` skips; `action`/`set` always fires; `music`/`sound` fade out in 0.3s; `video` cuts to end; `choice` selects default if `skip:skip_choices`. Tests in `agstests/`. *(depends on T-CUT22)*

- [ ] **T-CUT24** — GDScript: State consistency guarantee — when skip triggers to a label or cutscene end, perform a dry-run pass over all skipped steps: collect and immediately fire every `<<action>>` and `<<set>>` command between skip origin and destination before rendering the skip destination. State is always consistent on arrival. Tests in `agstests/`. *(depends on T-CUT23)*

## Notes

- T-CUT16–T-CUT21 are independent command type implementations and may be worked in parallel.
- T-CUT29 (Go grammar/emitter) is independent of the GDScript command tasks.
- T-CUT22 → T-CUT23 → T-CUT24 must be done in order.
- After this batch: T-CUT25 (save blocking), T-CUT26 (cutscene state in save), T-CUT27/T-CUT28 (Blender frame tags), T-CUT30 (cutscene localisation), T-CUT31 (audio leak cleanup), T-CUT32 (dialogue ducking), then M-LOC remaining tasks.
