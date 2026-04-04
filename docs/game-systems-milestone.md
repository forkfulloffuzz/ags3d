# AGS3D Game Systems Milestone — Design Document

## Goal

Complete the missing adventure-game systems needed to author a full classic
adventure game. By the end of this milestone, an author can create a game with
dialogue, inventory, items, a GUI layer, audio, global state, and save/load —
all authored in AGS-spirit and surfaced through the AG Studio editor.

This milestone depends on the Editor milestone being complete: AG Studio
(the custom Godot-based editor), the scene generator, and the minimal game
loop must already work.

---

## Systems Overview

| System | What it enables |
|---|---|
| **Speech** | Characters speak and think; text displayed on screen; voice audio optional |
| **Inventory** | Pick up, drop, combine items; persistent per-character inventory |
| **Items** | Named objects that exist in rooms and can be interacted with |
| **GUI / HUD** | On-screen panels: inventory bar, verb bar, status text, custom overlays |
| **Global state** | Named integer/boolean/string variables shared across rooms |
| **Room transitions** | Walk to a trigger region → load a new room |
| **Audio** | Play music and sound effects from AGS-spirit |
| **Save / Load** | Persist game state to a named slot and restore it |
| **Cutscenes** | Non-interactive sequences using existing blocking call infrastructure |
| **2D Billboard Characters** | Sprite-sheet characters rendered as camera-facing quads in the 3D world; direction selection from movement angle |

---

## System Specifications

### S1 — Speech System

**What it does:** Lets a character `Say` or `Think` dialogue lines. Text is
displayed on screen near the character; the engine waits until the line
finishes before the script continues.

#### AGS-spirit additions

```
character.Say("Some dialogue line.");       // blocking
character.Think("Internal monologue...");   // blocking, different visual style
character.SayAt(x, y, "Positioned text.");  // blocking, fixed position
```

#### Runtime behaviour

- Emit `await character.say("text")` in generated GDScript (blocking call).
- `say()` in `ags_character.gd`:
  1. Creates a `Label3D` (or a 2D `Label` in a `CanvasLayer`) positioned above
     the character.
  2. Waits for a duration derived from word count (or audio clip length if voice
     is provided).
  3. Removes the label; emits `say_completed` signal.
- Text rendering is configurable per character: font, colour, style.

#### C++ additions

- `AGSCharacter`: add `say_completed` signal.
- `AGSRuntime`: add `display_message(text: String)` for narrator lines.

#### `.agchar` additions

```
Character "player" {
    display_name  = "Player"
    speech_colour = "#FFFFFF"
    speech_font   = "assets/fonts/main.ttf"   # optional
}
```

#### AG Studio

- Script editor recognises `Say` / `Think` as blocking calls (already in
  grammar; emitter support is already in `grammar.md`).
- Character editor gains a Speech section with colour picker and font picker.

---

### S2 — Inventory System

**What it does:** Gives each character a named inventory. Scripts can add,
remove, and query items.

#### AGS-spirit additions

```
character.AddInventory("rusty_key");
character.LoseInventory("rusty_key");
character.HasInventory("rusty_key")    // returns bool
character.InventoryCount()             // returns int
```

#### Runtime

- `AGSCharacter` holds an `Array[StringName]` of item names.
- `add_inventory(name)` / `lose_inventory(name)` / `has_inventory(name)` are
  GDScript methods on the character runtime script.
- These are **non-blocking** — no `await` emitted.

#### Item definitions (`.agitem`)

```
Item "rusty_key" {
    display_name = "Rusty Key"
    description  = "An old iron key."
    sprite       = "assets/items/rusty_key.png"
}
```

Items live in `items/rusty_key.agitem`. `ag build` generates no scene for
items — they are data only, loaded by `AGSRuntime` at startup.

#### `AGSRuntime` additions

- `get_item(name: String) -> AGSItem` — returns item data node.
- `AGSItem` C++ node: `item_name`, `display_name`, `description`, sprite path.

#### AG Studio

- New **Items** section in the project panel (list + create/delete).
- Item editor: name, display name, description, sprite picker.

---

### S3 — Items in Rooms (AGSRoomItem)

**What it does:** A named item that exists at a position in a room. The player
can click on it to interact; scripts can show/hide it.

#### `.agroom` addition

```
RoomItem "old_chest" {
    item       = "rusty_key"    # references items/rusty_key.agitem
    position   = (2.0, 0.0, 1.5)
    visible    = true
}
```

#### C++ node: `AGSRoomItem`

- Properties: `item_name`, `visible`.
- Emits `item_clicked(item_name: String)` signal when clicked.
- `AGSRoom` connects `item_clicked` to the script handler `item_interact(name)`.

#### AGS-spirit additions

```
function item_interact(name) {
    if name == "old_chest" {
        player.AddInventory("rusty_key");
        HideRoomItem("old_chest");
    }
}

HideRoomItem("old_chest");     // non-blocking built-in
ShowRoomItem("old_chest");
```

#### AG Studio

- Room editor canvas: `RoomItem` shown as a sprite icon; click to place; right
  panel shows item reference + position.

---

### S4 — GUI / HUD System

**What it does:** An in-game on-screen layer for displaying game state and
accepting player input. Authored in a new `.agui` file format; rendered as a
Godot `CanvasLayer`.

**MVP scope:** inventory bar showing held items, optional verb bar (Look, Use,
Pick Up), and a status text area.

#### `.agui` file

```
GUI "main_hud" {
    layer = 10

    InventoryBar "inv_bar" {
        position   = (0, 0, bottom)    // anchor: bottom-left
        item_size  = (48, 48)
        columns    = 8
    }

    VerbBar "verbs" {
        position = (0, 0, bottom_right)
        verbs    = ["Look", "Use", "Pick up", "Talk to"]
    }

    StatusLine "status" {
        position = (0, 0, top)
        font     = "assets/fonts/main.ttf"
    }
}
```

#### Runtime

- `ag build` generates a `CanvasLayer` scene from `.agui`.
- `AGSRuntime` holds a reference to the active GUI and exposes:
  ```
  AGSRuntime.show_gui("main_hud")
  AGSRuntime.hide_gui("main_hud")
  AGSRuntime.set_status_text("...")
  ```
- `InventoryBar` auto-populates from the player character's inventory and
  refreshes when items are added/removed.
- Clicking a verb updates `AGSRuntime.active_verb`; scripts can read it.

#### AGS-spirit additions

```
SetStatusText("Look at what?");
SetActiveVerb("Use");
GetActiveVerb()      // returns string
```

#### AG Studio

- New **GUIs** section in the project panel.
- A layout editor for `.agui` files: drag widgets onto a screen-sized canvas.

---

### S5 — Global State

**What it does:** Named variables shared across all rooms and scripts.

#### `.agp` addition

```
[globals]
score        = 0
door_unlocked = false
player_name  = ""
```

#### AGS-spirit additions

```
global.score += 10;
global.door_unlocked = true;
if global.door_unlocked { ... }
```

#### Runtime

- `AGSRuntime` holds a `Dictionary` of global variables, initialised from
  `game.agp` at startup.
- `get_global(name)` / `set_global(name, value)` exposed to GDScript and
  (via binding) to generated scripts.
- Globals persist across room transitions and are included in save data.

#### Grammar change

The `global.NAME` expression already appears in `grammar.md` as a read-only
built-in. It must become read/write for user-defined globals.

---

### S6 — Room Transitions

**What it does:** Allows scripts to change the active room and optionally
place the player character at a named spawn point in the new room.

#### AGS-spirit additions

```
player.GoToRoom("garden", "garden_entrance");   // blocking
```

#### Runtime

- `go_to_room(room_name, spawn_name)` in `ags_character.gd`:
  1. Emits a `room_change_requested` signal on `AGSRuntime`.
  2. `AGSRuntime` unloads the current room scene, loads the target room scene,
     and places the character at the named spawn.
- `AGSRoom` emits `room_enter` after the new room is ready.

#### `AGSRuntime` additions

- `load_room(room_name: String)` — replaces the current scene.
- `room_change_requested(room_name, spawn_name)` signal.

#### AG Studio

- Room editor: `SpawnPoint` property panel shows "target of which GoToRoom
  calls?" (informational, no blocking).

---

### S7 — Audio System

**What it does:** Play music tracks and sound effects from AGS-spirit scripts.

#### AGS-spirit additions

```
PlayMusic("theme_main");       // non-blocking, loops
StopMusic();
PlaySound("door_creak");       // non-blocking, one-shot
```

#### Asset format

Audio files live in `audio/music/` and `audio/sfx/`. No special manifest file;
`ag build` scans those directories and makes the names available.

#### Runtime

- `AGSRuntime` manages an `AudioStreamPlayer` for music (one active at a time)
  and a pool for sound effects.
- `play_music(name)`, `stop_music()`, `play_sound(name)` are GDScript methods
  on `AGSRuntime`.

#### AG Studio

- No dedicated audio editor for MVP. Authors drop audio files into
  `audio/music/` or `audio/sfx/` and reference them by filename stem.

---

### S8 — Save / Load

**What it does:** Persist and restore complete game state.

#### State captured

- Current room name.
- Character positions and inventories.
- All global variables.
- All room item visible/hidden states.
- Active music track name.

#### AGS-spirit additions

```
SaveGame(1);           // save to slot 1
LoadGame(1);           // restore from slot 1
GameSaved(1)           // returns bool — does slot 1 exist?
```

#### Runtime

- `AGSRuntime.save_game(slot: int)` serialises state to `user://save_<slot>.json`.
- `AGSRuntime.load_game(slot: int)` restores state.
- Save/load are blocking calls that happen between frames (no async needed for MVP).

---

### S9 — Cutscenes

**What it does:** Non-interactive sequences composed of blocking AGS-spirit
calls. No new language feature needed — cutscenes are just room scripts or
global script functions that chain blocking calls.

What *is* needed:

- `SetPlayerControl(false)` / `SetPlayerControl(true)` to disable click-to-walk
  during a cutscene.
- `FadeIn()` / `FadeOut()` screen fade transitions (blocking).
- `Wait(seconds)` built-in (already in grammar, needs runtime implementation).

#### AGS-spirit additions

```
SetPlayerControl(false);
FadeOut();
player.WalkTo("door");
camera.MoveTo("main");
FadeIn();
SetPlayerControl(true);
```

#### Runtime

- `AGSRuntime` tracks `player_control_enabled: bool`.
- `fade_in()` / `fade_out()` animate a black `ColorRect` in a `CanvasLayer`.
- `wait(seconds: float)` is `await get_tree().create_timer(seconds).timeout`.

---

### S10 — 2D Billboard Characters

**What it does:** Lets a character use a 2D sprite sheet rendered as a
camera-facing quad (`Sprite3D`) in the 3D world, instead of a 3D mesh.
This reproduces the classic look of games like Broken Sword or early
Sierra 3D titles.

#### Visual modes

| Mode | Description |
|---|---|
| `mesh` | Default — 3D geometry (capsule or custom `.glb`) |
| `billboard` | 2D sprite sheet, quad always faces the camera |

#### Sprite angle sets

The sprite sheet is divided into angle sets. The runtime selects the correct
set based on the character's movement direction relative to the camera's
forward vector.

| `sprite_angles` | Directions | Horizontal snap |
|---|---|---|
| `1` | Single fixed frame — no direction logic | n/a (always same art) |
| `4` | N / E / S / W | 90° |
| `8` | N / NE / E / SE / S / SW / W / NW | 45° |

The angle sets are laid out vertically in the sprite sheet (top = first angle;
order: same as table top-to-bottom). Frames within each angle are laid out
horizontally (idle cycle left-to-right).

#### `.agchar` additions

```
Character "player" {
    display_name     = "Player"
    visual_mode      = "billboard"
    sprite_sheet     = "assets/sprites/player.png"
    sprite_angles    = 4          # 1 | 4 | 8
    frame_size       = (64, 128)  # pixels (width, height) per frame
    frames_per_angle = 8          # animation frames per direction
}
```

#### Runtime

- `AGSCharacter` C++: add `visual_mode` property (`"mesh"` | `"billboard"`).
- `ag build` generates the character `.tscn` with a `Sprite3D` root (instead
  of `MeshInstance3D`) when `visual_mode = "billboard"`:
  - `billboard = BaseMaterial3D.BILLBOARD_ENABLED`
  - `texture` set to the sprite sheet asset
  - `hframes` / `vframes` derived from `frame_size` + sheet dimensions
- `ags_character.gd` runtime:
  - Computes the angle between the character's velocity vector and the camera's
    forward vector projected onto XZ.
  - Quantizes to the nearest valid direction (4- or 8-way snap).
  - Sets `Sprite3D.frame` to the correct row offset + current animation frame.
  - For `sprite_angles = 1`: no direction logic; always row 0.

#### Camera interaction

Billboard sprites are designed to be viewed from a roughly horizontal camera.
Two constraints arise:

1. **Elevation constraint** — a camera pitched steeply downward shows the top
   of the billboard quad; there is no "top view" art. Authors should keep
   camera elevation below ~30° from horizontal when using billboard characters.

2. **Arc constraint** — for `sprite_angles = 1` (single-direction art), the
   camera must not orbit around the character; it must stay within the facing
   arc of the sprite. The `.agroom` Camera block accepts `sprite_locked = true`
   which disables in-game camera orbiting for that camera.

The AG Studio editor enforces these constraints as **warnings** (see F12 in
the Editor milestone). The game will still run when violated — the visual
artefact is the author's responsibility to resolve.

#### `.agroom` Camera addition

```
Camera "main" {
    position     = (4.79, 3.0, 5.60)
    look_at      = (0.0, 0.0, 0.0)
    sprite_locked = true    # optional; disables orbit for single-angle sprites
}
```

---

## Task Breakdown

| Task | Description | Depends on |
|---|---|---|
| T-GS01 | C++: `AGSCharacter` — `say_completed` signal; `say()`, `think()` in runtime script | Editor milestone |
| T-GS02 | C++: `AGSItem` node + `AGSRuntime.get_item()` | Editor milestone |
| T-GS03 | C++: `AGSRoomItem` node — `item_clicked` signal, room wiring | T-GS02 |
| T-GS04 | Go: `ag build` — `.agitem` parser + inventory/item validation in `ag validate` | T-GS02 |
| T-GS05 | Go: grammar + emitter — `Say`, `Think`, `AddInventory`, `LoseInventory`, `HasInventory` | T-GS01, T-GS02 |
| T-GS06 | Go: grammar + emitter — `HideRoomItem`, `ShowRoomItem`, `item_interact` handler | T-GS03 |
| T-GS07 | Go: grammar + emitter — `global.NAME` read/write, `game.agp` globals section | Editor milestone |
| T-GS08 | C++: `AGSRuntime` — global variable store | T-GS07 |
| T-GS09 | Go: grammar + emitter — `GoToRoom`, room transition blocking call | Editor milestone |
| T-GS10 | GDScript: `AGSRuntime` — `load_room()`, `room_change_requested` signal | T-GS09 |
| T-GS11 | Go: grammar + emitter — `PlayMusic`, `StopMusic`, `PlaySound` | Editor milestone |
| T-GS12 | GDScript: `AGSRuntime` — audio manager (music player + sfx pool) | T-GS11 |
| T-GS13 | Go: `.agui` parser + `ag build` GUI scene generator | Editor milestone |
| T-GS14 | GDScript: GUI runtime — `InventoryBar`, `VerbBar`, `StatusLine` | T-GS02, T-GS13 |
| T-GS15 | Go: grammar + emitter — `SetStatusText`, `SetActiveVerb`, `GetActiveVerb` | T-GS14 |
| T-GS16 | GDScript: `AGSRuntime` — `save_game()`, `load_game()` | T-GS07, T-GS08 |
| T-GS17 | Go: grammar + emitter — `SaveGame`, `LoadGame`, `GameSaved` | T-GS16 |
| T-GS18 | GDScript/C++: cutscene support — `SetPlayerControl`, `FadeIn`, `FadeOut`, `Wait` | Editor milestone |
| T-GS19 | Go: grammar + emitter — `SetPlayerControl`, `FadeIn`, `FadeOut`, `Wait` | T-GS18 |
| ~~T-GS20~~ | ➡️ M12 T-CE19 | ~~GDScript: AG Studio — item editor panel~~ | moved to Custom Editor milestone |
| ~~T-GS21~~ | ➡️ M12 T-CE07 | ~~GDScript: AG Studio — room editor RoomItem gizmo~~ | moved to Custom Editor milestone |
| ~~T-GS22~~ | ➡️ M12 T-CE20 | ~~GDScript: AG Studio — GUI layout editor~~ | moved to Custom Editor milestone |
| ~~T-GS23~~ | ➡️ M12 T-CE21 | ~~GDScript: AG Studio — global variables editor~~ | moved to Custom Editor milestone |
| T-GS24 | C++: `AGSCharacter` — `visual_mode` property; `ag build` generates `Sprite3D` scene for billboard mode | Editor milestone |
| T-GS25 | Go: `.agchar` billboard properties (`visual_mode`, `sprite_sheet`, `sprite_angles`, `frame_size`, `frames_per_angle`); scene generator outputs `Sprite3D` node | T-GS24 |
| T-GS26 | GDScript: billboard direction selection runtime — angle quantization (4/8-way), frame cycling, `sprite_locked` camera support | T-GS24 |
| T-GS27 | C++: split `AGSCharacter` → `AGSCharacterBase`, `AGSCharacter3D`, `AGSCharacter2D`; preserve existing signal/property interface | Editor milestone |
| T-GS28 | GDScript: `AGSAnimationPlayerBase` (common API: `play_clip`, `stop`, `set_state`, `on_anim_event`) + `AGSAnimationPlayer3D` wrapping `AnimationPlayer` | T-GS27 |
| T-GS29 | GDScript: `AGSAnimationPlayer2D` — billboard direction + frame cycling; implements `AGSAnimationPlayerBase` API | T-GS27, T-GS24 |
| T-GS30 | Go: `ag build` — generate `AGSCharacter3D` or `AGSCharacter2D` `.tscn` based on `type` field in `.agchar` | T-GS27 |

---

## Out of Scope

- Voice acting / lip sync (deferred to a later audio milestone).
- Dialogue trees with branching choices (deferred; requires a new `.agdialog`
  format).
- Pathfinding across multiple rooms (characters stay within their current room).
- Visual inventory combining (drag-and-drop item-on-item puzzles) — basic
  `HasInventory` checks in scripts suffice for MVP.
- Localization / multiple language support.
- Export / distribution packaging.
