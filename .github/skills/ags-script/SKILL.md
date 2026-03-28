---
name: ags-script
description: "Use when creating or editing .agscript files — the AGS-spirit scripting language for room logic, character interaction, hotspot events, and game flow. ag build transpiles .agscript to GDScript. The formal grammar is in docs/grammar.md."
argument-hint: "Room name or script task, e.g. 'write start room script' or 'add hotspot interaction'"
---

# AGS Script Files (.agscript)

## Role in the Pipeline

`.agscript` contains all room logic in AGS-spirit syntax. `ag build` transpiles it to GDScript in `.engine/generated/`. The generated `.gd` is attached as the script on the `AGSRoom` node.

```
rooms/start/
  start.agscript   ← author edits this
  .engine/generated/rooms/start/start.agscript.gd  ← generated, never edited
```

The formal grammar is in [`docs/grammar.md`](../../docs/grammar.md) — that document is authoritative for all syntax questions.

## Event Handler Reference

Event handlers are functions with reserved names that the engine calls automatically:

| Handler | When it fires |
|---------|---------------|
| `room_load()` | Once when the room scene loads, before `room_enter` |
| `room_enter()` | Each time the player enters the room |
| `hotspot_NAME_Interact()` | Player clicks/activates hotspot named NAME |
| `region_walked_into(region_name)` | Character enters a TriggerRegion |
| `region_walked_off(region_name)` | Character leaves a TriggerRegion |

## Built-in Functions

These are the primary built-in calls available in `.agscript`. Blocking calls (marked ⏳) emit `await` in the generated GDScript — the script suspends until the action completes.

| Function | Blocking | Description |
|----------|----------|-------------|
| `Character(name).walk_to(point)` | ⏳ yes | Walk character to named point |
| `Character(name).face_to(point)` | ⏳ yes | Rotate character to face named point |
| `Character(name).say(text)` | ⏳ yes | Character speaks a line |
| `Wait(frames)` | ⏳ yes | Pause script for N frames |
| `FadeIn()` / `FadeOut()` | ⏳ yes | Screen fade transitions |
| `ChangeRoom(room)` | ⏳ yes | Transition to another room |
| `Character(name).move_speed` | no | Get/set movement speed |

`Character(name)` resolves via `AGSRuntime.get_character(name)`. The emitter maps this call to the appropriate GDScript.

## Global State

`global.NAME` accesses engine-owned game state:

```
global.player    // the player character
global.room      // current room name
global.score     // integer game score
```

These map to `AGSRuntime` properties in the generated GDScript.

## Language Quick Reference

```agscript
// Variables
int count = 0;
string message = "hello";
bool flag = true;

// Control flow
if (count > 5) {
    count = 0;
} else {
    count = count + 1;
}

while (flag) {
    // ...
}

// Event handler
function room_enter() {
    Character("player").walk_to("door_left");   // blocking — emits await
    Character("player").face_to("window");       // blocking — emits await
}

// Hotspot handler
function hotspot_ancient_tome_Interact() {
    Character("player").say("An old book...");
}

// Cross-file sharing
namespace Utils {
    export function greet(string name) {
        Character(name).say("Hello!");
    }
}
// Called as: Utils.greet("player");
```

## Complete Minimal Example

```agscript
// rooms/start/start.agscript

function room_enter() {
    // Walk player to door, then face the window
    Character("player").walk_to("door_left");
    Character("player").face_to("window");
}

function hotspot_notice_board_Interact() {
    Character("player").say("A notice about the festival.");
}

function region_walked_into(string region_name) {
    if (region_name == "exit_zone") {
        ChangeRoom("town_square");
    }
}
```

## Conventions

- File name matches the room directory: `rooms/library/library.agscript`
- All `Point` names used in `walk_to`/`face_to` must exist in the room's `.agroom`
- All `Character` names must exist as `.agchar` files
- All `Hotspot` names referenced in handlers must exist in the room's `.agroom`
- `ag validate` catches broken references before build

## See Also

- [AGS-Spirit Grammar](../../docs/grammar.md) — full formal grammar, all token types, AST shapes, emit rules
- [ags-room skill](../ags-room/SKILL.md) — defining points, hotspots, and regions in `.agroom`
