# AGS3D — Current Development Tasks

This file tracks the active batch of tasks. Update status as work progresses.
When all tasks are done, ask Claude to pick the next 10.

## Phase 13 — M10 Game Systems Complete (done)

All M10 Game Systems tasks are implemented:
- **T-GS01** — C++: `AGSCharacterBase` with `say_completed` signal; GDScript `say()`/`think()`
- **T-GS02** — C++: `AGSItem` node + `AGSRuntime.get_item()`
- **T-GS03** — C++: `AGSRoomItem` with `item_clicked` signal
- **T-GS04** — Go: `.agitem` parser + `ag validate` item checks
- **T-GS05** — Go: `Say`/`Think`/`AddInventory`/`LoseInventory`/`HasInventory` emitter
- **T-GS06** — Go: `HideRoomItem`/`ShowRoomItem`/`item_interact` emitter
- **T-GS07** — Go: `global.NAME` read/write grammar + emitter + `game.agp` globals section
- **T-GS08** — C++: `AGSRuntime` `get_global`/`set_global`/`init_globals`
- **T-GS09** — Go: `GoToRoom` grammar + emitter
- **T-GS10** — GDScript: `AGSRuntime` AutoLoad + `load_room()`
- **T-GS11** — Go: `PlayMusic`/`StopMusic`/`PlaySound` emitter
- **T-GS12** — GDScript: `AGSAudio` AutoLoad (music player + sfx pool)
- **T-GS13** — Go: `.agui` parser + `ag build` GUI scene generator
- **T-GS14** — GDScript: `ags_gui.gd` (InventoryBar, VerbBar, StatusLine)
- **T-GS15** — Go: `SetStatusText`/`SetActiveVerb`/`GetActiveVerb` emitter
- **T-GS16** — C++: `AGSRuntime` save/load with global variables
- **T-GS17** — Go: `SaveGame`/`LoadGame`/`GameSaved` emitter
- **T-GS18** — GDScript: cutscene runtime (`fade_in`/`fade_out`/`wait`)
- **T-GS19** — Go: `SetPlayerControl`/`FadeIn`/`FadeOut`/`Wait` emitter
- **T-GS24** — C++: `AGSCharacter.visual_mode` property
- **T-GS25** — Go: billboard `.agchar` properties + `Sprite3D` scene generation
- **T-GS26** — GDScript: billboard direction runtime
- **T-GS27** — C++: `AGSCharacterBase` / `AGSCharacter3D` / `AGSCharacter2D` split
- **T-GS28** — GDScript: `AGSAnimationPlayer3D`
- **T-GS29** — GDScript: `AGSAnimationPlayer2D`
- **T-GS30** — Go: generate `AGSCharacter3D` vs `AGSCharacter2D` `.tscn`

## Phase 14 — M12 AG Studio Complete (done)

All M12 AG Studio tasks are implemented (see `docs/custom-editor-milestone.md`):
- Foundation panels, Room editor, Character editor, Script editor, Build Log,
  Play button, Project Wizard, Item editor, GUI Layout editor, Global variables editor
- All AG Studio gizmos (WalkableSurface, BlockerVolume, Hotspot, TriggerRegion,
  Point, SpawnPoint, Camera, Item)

## Deferred / Blocked

- **T40** — C++/GDScript: disable AGSRuntime trace in production builds (done: _trace_enabled = false by default; auto-enabled in editor)
- **T-FINAL** — C++: embed `.engine/runtime/` GDScripts into the C++ module at build time (done: ResourceFormatLoaderAGSEmbed + embedded_scripts generator)

### Requires Godot C++ fork compilation
The C++ module at `modules/agvm/` is fully written and ready. It needs compilation
into the Godot binary. All C++ node types (AGSCharacterBase, AGSCharacter3D,
AGSCharacter2D, AGSItem, AGSRoomItem, AGSRuntime, etc.) are implemented, registered,
and the 19 runtime GDScripts are embedded as C++ string constants.
