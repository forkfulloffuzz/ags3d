# AGS3D Character System — Architecture Document

## Overview

Every character in AGS3D has two independent concerns:

1. **Navigation & gameplay** — shared by all characters: walking on a navmesh,
   facing directions, saying dialogue, carrying inventory. Authored identically
   in AGS-spirit regardless of visual type.
2. **Visual & animation** — type-specific: a 3D mesh driven by a skeleton, or
   a 2D sprite sheet displayed as a billboard quad in 3D space.

The system is designed so that the AGS-spirit API and the Godot signal
interface are identical for both types. Only the internal rendering and
animation machinery differs.

---

## Character Types

| Type | Visual | Animation | Blender workflow |
|---|---|---|---|
| **3D** | 3D mesh (`.glb`) — skeleton, skin weights | `AnimationPlayer` with skeletal clips | Full Blender rig + NLA export |
| **2D** | `Sprite3D` billboard in 3D world — sprite sheet | Frame selection by movement direction | Sprite sheet only; no Blender rig |

Both types navigate in 3D space using `NavigationAgent3D`. The difference is
purely visual.

`.agchar` selects the type via the `type` field:

```
Character "player" {
    display_name = "Player"
    type         = "3d"        # "3d" (default) | "2d"

    # 3D-specific
    mesh           = "characters/player/player.glb"
    animations = {
        idle = "Idle"
        walk = "Walk"
        talk = "Talk"
    }
}

Character "npc" {
    display_name = "Guard"
    type         = "2d"

    # 2D-specific
    sprite_sheet     = "assets/sprites/guard.png"
    sprite_angles    = 8
    frame_size       = (64, 128)
    frames_per_angle = 6
}
```

---

## C++ Node Hierarchy

```
AGSCharacterBase   (C++ — extends CharacterBody3D)
│  Properties shared by all characters:
│    character_name: String
│    move_speed: float
│    character_type: String   ("3d" | "2d")
│  Signals shared by all characters:
│    walk_completed
│    face_completed
│    say_completed
│    animation_event(label: String)
│
├── AGSCharacter3D  (C++ — extends AGSCharacterBase)
│     (no extra C++ properties — visual handled entirely by GDScript runtime)
│
└── AGSCharacter2D  (C++ — extends AGSCharacterBase)
      (no extra C++ properties — visual handled entirely by GDScript runtime)
```

Both subclasses are thin. All navigation behaviour, animation dispatch,
and the AGS API (`walk_to`, `face_to`, `say`, etc.) live in the GDScript
runtime scripts (`ags_character_3d.gd` / `ags_character_2d.gd`) that extend
the respective C++ base via `set_script()`.

**Why split at C++ level at all?** `AGSRuntime` can use `character_type` to
determine which runtime script to attach, `ag build` can generate the correct
`.tscn` node type, and the AG Studio editor can display type-appropriate UI
without inspecting internal GDScript properties.

---

## GDScript Runtime Scripts

Both runtime scripts live in `.engine/runtime/`.

### `ags_character_3d.gd`

```gdscript
extends AGSCharacter3D

var _nav_agent: NavigationAgent3D
var _anim_player: AGSAnimationPlayer3D

func _ready() -> void:
    _nav_agent = _setup_nav_agent()
    _anim_player = AGSAnimationPlayer3D.new(get_node("AnimationPlayer"))

# Shared AGS API
func walk_to(point_name: String) -> void: ...
func face_to(point_name: String) -> void: ...
func say(text: String) -> void: ...

# Delegates all animation state to the typed player
func _set_anim_state(state: String) -> void:
    _anim_player.set_state(state)
```

### `ags_character_2d.gd`

```gdscript
extends AGSCharacter2D

var _nav_agent: NavigationAgent3D
var _anim_player: AGSAnimationPlayer2D

func _ready() -> void:
    _nav_agent = _setup_nav_agent()
    var sprite := get_node("Sprite3D") as Sprite3D
    _anim_player = AGSAnimationPlayer2D.new(sprite, sprite_angles, frames_per_angle)

# Identical AGS API — walk_to, face_to, say — same as 3D
func walk_to(point_name: String) -> void: ...
func face_to(point_name: String) -> void: ...
func say(text: String) -> void: ...

# Delegates direction + frame selection to the typed player
func _physics_process(delta: float) -> void:
    _anim_player.update(velocity, _camera_forward())
```

---

## Animation Player Hierarchy

```
AGSAnimationPlayerBase   (GDScript base class)
│  Common API:
│    play_clip(name: String) — play a named animation clip
│    stop()
│    set_state(state: String)  — "idle" | "walk" | "talk"
│    on_anim_event(label: String)  — dispatch frame-tag events (T-BL16)
│  Abstract methods (must override):
│    _do_play(name: String)
│    _do_stop()
│    _do_set_state(state: String)
│
├── AGSAnimationPlayer3D  (GDScript)
│     Wraps a Godot AnimationPlayer node.
│     Maps logical state → animation resource names from .agchar animations block.
│     Wires method call tracks for frame-tag events (when T-BL16 is implemented).
│
└── AGSAnimationPlayer2D  (GDScript)
      Wraps a Sprite3D node.
      Holds sprite_angles, frames_per_angle config from .agchar.
      update(velocity, camera_forward): quantize direction → select row.
      Drives frame index via an internal Timer for idle/walk cycles.
      set_state("talk"): freezes direction logic, plays talk cycle frames.
```

### Why a base class?

- The rest of the character runtime (`walk_to`, `say`, state transitions) calls
  `_anim_player.set_state("walk")` without knowing which type it is.
- Frame-tag event dispatch (`on_anim_event`) is implemented once in the base
  and both players call it when an event fires.
- Future types (e.g. a pre-rendered video character, or a puppeted marionette)
  can add a third subclass without touching the character or AGS API.

---

## AG Studio — Character Editor

The character editor (T-E13) shows type-specific sections controlled by the
`type` field in `.agchar`.

```
┌─ Character Editor ─────────────────────────────────┐
│  Name:  player  (read-only)                        │
│  Display name: [ Player          ]                 │
│  Type: ( 3D )  ( 2D )                              │
├────────────────────────────────────────────────────┤
│  ── 3D Properties ─────────────────── (type = 3d) │
│  Mesh file:  [ characters/player/player.glb  📁 ] │
│                                                    │
│  Animations                                        │
│  idle  [ Idle  ]  ▶ preview                       │
│  walk  [ Walk  ]  ▶ preview                       │
│  talk  [ Talk  ]  ▶ preview                       │
├────────────────────────────────────────────────────┤
│  ── 2D Properties ─────────────────── (type = 2d) │
│  Sprite sheet: [ assets/sprites/guard.png    📁 ] │
│  Angles:  ( 1 )  ( 4 )  ( 8 )                    │
│  Frame size:  W [ 64 ]  H [ 128 ]                 │
│  Frames/angle:  [ 6 ]                             │
└────────────────────────────────────────────────────┘
```

Switching type clears the type-specific fields (after a confirmation prompt)
and rewrites the `.agchar` file.

---

## AG Studio — Animation Viewer

A sub-panel within the character editor showing a live preview of the
selected animation. The viewer is type-specific.

### 3D Animation Viewer

An embedded Godot `SubViewport` showing the character's 3D mesh, lit with a
simple three-point light rig:

- Clip selector dropdown (lists clips from the `animations` block).
- **Play / Pause / Stop** transport controls.
- Frame scrubber slider.
- Playback speed multiplier.
- The character mesh plays the selected `AnimationPlayer` clip in the viewport.

### 2D Animation Viewer

A sprite-sheet inspector panel:

- **Sheet grid**: displays the sprite sheet PNG with a grid overlay derived
  from `frame_size`. Hover over a cell to see frame index, row (direction),
  column (frame number).
- **Direction preview**: a row of thumbnails showing each direction's idle
  frame (frame 0 of each row), labelled N / NE / E / etc.
- **Animated preview**: a single large cell that cycles through frames of the
  selected direction at configurable FPS. Direction selected by clicking a
  thumbnail or a compass widget.
- **Export validation**: highlights cells that fall outside the sheet bounds
  (misconfigured `frame_size` or `frames_per_angle`).

---

## AGS-Spirit API (identical for both types)

From the script author's perspective, there is no difference:

```
// Works for both 3D and 2D characters
player.WalkTo("door_left");       // blocking
player.FaceTo("window");          // blocking
player.Say("Hello there.");       // blocking
player.Think("I wonder...");      // blocking
player.AddInventory("rusty_key"); // non-blocking
player.HasInventory("rusty_key"); // returns bool
player.GoToRoom("garden", "gate_spawn");  // blocking
```

The runtime script for each type implements the same method signatures.
`AGSRuntime.get_character(name)` returns `AGSCharacterBase`, so scripts never
need to know the type.

---

## `ag build` Behaviour

`ag build` reads `type` from `.agchar` and generates the appropriate `.tscn`:

**3D character `.tscn`:**
```
AGSCharacter3D (root)
  character_name = "player"
  script = res://.engine/runtime/ags_character_3d.gd
  MeshInstance3D  (or sub-scene from .glb if mesh field set)
  CollisionShape3D
  AnimationPlayer  (present only if mesh/.glb provided)
```

**2D character `.tscn`:**
```
AGSCharacter2D (root)
  character_name = "npc"
  script = res://.engine/runtime/ags_character_2d.gd
  Sprite3D
    billboard = BILLBOARD_ENABLED
    texture = <sprite_sheet path>
    hframes = <sheet_width / frame_width>
    vframes = <sprite_angles>
  CollisionShape3D  (CapsuleShape3D — always, regardless of visual type)
```

---

## Task Summary

These tasks are distributed across existing milestones:

| Task | Milestone | Description |
|---|---|---|
| T-GS27 | M10 | C++: split `AGSCharacter` → `AGSCharacterBase`, `AGSCharacter3D`, `AGSCharacter2D` |
| T-GS28 | M10 | GDScript: `AGSAnimationPlayerBase` + `AGSAnimationPlayer3D` |
| T-GS29 | M10 | GDScript: `AGSAnimationPlayer2D` (billboard direction + frame cycling) |
| T-GS30 | M10 | Go: `ag build` — generate `AGSCharacter3D` or `AGSCharacter2D` `.tscn` based on `type` field |
| T-E20 | M9 | GDScript: Character editor — type selector (3D/2D) with type-specific property sections |
| T-E21 | M9 | GDScript: 3D Animation viewer — embedded SubViewport, transport controls, frame scrubber |
| T-E22 | M9 | GDScript: 2D Animation viewer — sprite sheet grid, direction thumbnails, animated preview |
