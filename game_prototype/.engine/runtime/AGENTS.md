# .engine/runtime — GDScript AutoLoad Runtime Agent Instructions

GDScript singletons and base classes that form the AGS3D runtime. These files
are deployed into Godot projects as AutoLoad nodes.

## AutoLoad singletons

| File | Node name | Purpose |
|------|-----------|---------|
| `ags_sequencer.gd` | AGSSequencer | Cutscene playback, skip system, per-title state tracking |
| `ags_sequencer_commands.gd` | — (extends AGSSequencer) | Dispatches `<<cmd arg>>` syntax in cutscene scripts |
| `ags_save_load.gd` | AGSSaveLoad | Save/load with cutscene state persistence; blocks saves mid-cutscene |
| `ags_audio.gd` | AGSAudio | Music / sound / ambient playback, bus ducking |
| `ags_dialogue.gd` | AGSDialogue | Dialogue engine (runs compiled `.agdlg` graphs) |
| `ags_dialogue_state.gd` | AGSDialogueState | Persistent dialogue variable store |
| `ags_localisation.gd` | AGSLocalisation | String lookup from `.agstrings` locale files |
| `ags_room_manager.gd` | AGSRoomManager | Room transitions and spawn point resolution |
| `ags_gui.gd` | AGSGUI | HUD / inventory / menu management |

## Base classes (not AutoLoads)

| File | Base class | Purpose |
|------|-----------|---------|
| `ags_animation_player_base.gd` | AGSAnimationPlayerBase | Shared animation API; `get_frame_tag(anim, frame)` |
| `ags_animation_player_3d.gd` | AGSAnimationPlayer3D | 3D character animation |
| `ags_animation_player_2d.gd` | AGSAnimationPlayer2D | 2D sprite-sheet character animation |
| `ags_character.gd` | AGSCharacter | Character walk / turn / say / inventory API |
| `ags_cutscene.gd` | AGSCutscene | Cutscene script loader |
| `ags_dialogue_ui.gd` | AGSDialogueUI | Dialogue choice/line presentation base |

## Testability rules

AutoLoad singletons cannot be retrieved via `/root/<Name>` in headless tests.
Use the **meta-injection pattern** to substitute test doubles:

```gdscript
# Singleton under test:
func _get_sequencer() -> Node:
    if has_meta("_seq_override"):
        return get_meta("_seq_override")
    return get_node("/root/AGSSequencer")

# In test setup:
save_load.set_meta("_seq_override", local_sequencer)
```

Every singleton that depends on another singleton must expose a `_get_X()`
helper that checks for a meta override first.

## Signal conventions

- Signals are emitted **after** the state change is complete, not before.
- Signals used across singletons are connected in `_ready()`, not constructors.
- One-shot connections use `CONNECT_ONE_SHOT`.

## Save / load cutscene state

`ags_save_load.gd` uses a read-modify-write pattern to persist cutscene state:

1. `runtime.save_game(slot)` writes base JSON (C++ side).
2. `_inject_cutscene_state(slot)` reads that file, adds `"cutscenes"` key, writes back.
3. On `load_game(slot)`, reads the file post-load to extract and restore cutscene state
   into `AGSSequencer` via `restore_cutscene_state(data)`.
