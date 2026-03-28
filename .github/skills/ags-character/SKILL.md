---
name: ags-character
description: "Use when creating or editing .agchar files — the source-of-truth for character definitions. .agchar defines display name, movement speed, and future visual/collision properties. ag build generates the AGSCharacter node configuration from this."
argument-hint: "Character name or task, e.g. 'create player character' or 'add npc guard'"
---

# AGS Character Files (.agchar)

## Role in the Pipeline

`.agchar` is the **source-of-truth** for a character. `ag build` reads it to validate character references in `.agroom` and `.agscript` files, and will generate character node configuration in the scene.

```
characters/
  player.agchar    ← author edits this
```

Characters are referenced by name in:
- `game.agp` — `start_character = "characters/player.agchar"`
- `.agroom` — `SpawnPoint { character = "player" }`
- `.agscript` — `Character("player").walk_to("door_left")`

## Format Reference

```
Character "NAME" {
    // Required
    display_name = "DISPLAY STRING"

    // Optional
    move_speed   = FLOAT        // units per second, default 4.0
}
```

## Field Reference

| Field          | Required | Default | Description |
|----------------|----------|---------|-------------|
| `display_name` | yes      | —       | Human-readable name shown in-game (dialogue captions, inventory, etc.) |
| `move_speed`   | no       | 4.0     | Navigation speed in world units per second; passed to AGSCharacter.move_speed |

## Runtime Mapping

The character name (first token after `Character`) is the key used everywhere:
- `AGSRuntime.get_character("player")` returns the AGSCharacter node
- `AGSCharacter.character_name` is set to this name at scene load
- `AGSCharacter.move_speed` is set from the `.agchar` value

Navigation behavior (walk_to, face_to) is implemented in `.engine/runtime/ags_character.gd` which is attached to every AGSCharacter node.

## Complete Example

```
Character "player" {
    display_name = "Alex"
    move_speed   = 4.0
}
```

```
Character "guard" {
    display_name = "Palace Guard"
    move_speed   = 3.0
}
```

## Conventions

- File name matches character name: `characters/player.agchar` → character `"player"`
- `display_name` may contain spaces and punctuation; character name (the key) is lowercase, no spaces
- `move_speed` typical range: 2.0 (slow NPC) to 6.0 (fast action sequence)
- One `.agchar` file per character — never define two characters in one file
