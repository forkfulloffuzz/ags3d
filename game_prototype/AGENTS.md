# game_prototype — Agent Instructions

The Godot 4 project used to prototype and validate AGS3D runtime features.

## Structure

```
game_prototype/
  .engine/
    runtime/          GDScript AutoLoad singletons and base classes
    generated/        Output from `ag build` — do not edit manually
    cache/            ag build mtime cache
  game.agp            AGS3D project manifest
  rooms/              .agroom source files
  characters/         .agchar source files
  dialogue/           .agdlg dialogue scripts
  audio/              Audio assets
  inventory/          .agitem item definitions
  scripts/            Misc GDScript
```

## Source files → generated output

`ag build game_prototype` reads files in `rooms/`, `characters/`, etc. and writes:
- `rooms/<name>.tscn` — Godot scene with walkable surfaces, hotspots, cameras
- `characters/<name>.tscn` — Character scene with animation player metadata
- `scripts/<name>.gd` — GDScript generated from `.agscript` files

**Never edit files under `.engine/generated/` directly.** Edit the source file
and re-run `ag build`.

## Runtime AutoLoads

See [.engine/runtime/AGENTS.md](.engine/runtime/AGENTS.md) for the full list of
singletons, their APIs, and testability notes.

## Testing runtime changes

Changes to `.engine/runtime/` are exercised by the headless test suite:

```sh
.dev/test.sh
```

Tests live in `agstests/`. When adding a new runtime method or signal, add a
test in the relevant `agstests/<module>/` directory and register it in
`agstests/run_tests.gd`.
