# AGS3D Project Guidelines

## What This Is

AGS3D is a purpose-built 3D adventure game engine — a Godot fork that encodes adventure game concepts (rooms, characters, inventory, dialogue, hotspots) as first-class engine primitives. Authors write in AGS-spirit scripting; the engine transpiles it to GDScript at build time.

## Repository Layout

```
modules/agvm/          # AGS3D C++ module — all AGS-specific engine code lives here
tools/ag/              # ag CLI tool (build, run, export, validate)
game_prototype/        # Minimal test project used for end-to-end validation
  game.agp             # TOML project manifest
  characters/          # .agchar files
  rooms/               # .agroom and .agscript files
  .engine/             # Build artifact — never manually edited
    generated/         # Transpiled GDScript output
    cache/
```

## Tech Stack

- **Engine base**: Godot 4 (C++), forked — minimize diff surface against upstream
- **Module**: `modules/agvm/` — all AGS3D code, registered via SCsub
- **Scripting input**: AGS-spirit (`.agscript`) — AGS-compatible syntax, new semantics
- **Scripting output**: GDScript (`.gd`) — generated build artifact, never hand-edited
- **CLI tool**: `ag` — invokes parser + emitter pipeline
- **Build artifact dir**: `.engine/generated/` — deleted and regenerated on each build

## Key Architecture Rules

- The transpiler (parser → emitter) is the highest-risk component. The parser (T09) and await emission (T16) are the hardest tasks — invest in correctness first.
- Generated GDScript is a build artifact. Source maps (`.agmap`) translate runtime errors back to AGS-spirit file and line number. Authors must never see a GDScript path.
- Authors never reference 3D coordinates. All spatial references use named points (`point.door_left`).
- Logic geometry (WalkableSurface, BlockerVolume, TriggerRegion, HotspotSurface) is invisible at runtime — it exists only for pathfinding, collision, and hit testing.
- Blocking calls in AGS-spirit (`WalkTo`, `PlayAnimation`, `Wait`) must emit `await` in GDScript. The parser annotates blocking call sites; the emitter acts on those annotations.
- `.engine/` is gitignored. Never commit generated GDScript.

## Build & Run

```sh
ag build     # parse all changed .agscript files, emit GDScript to .engine/generated/
ag run       # build + launch Godot editor with the project
ag validate  # static analysis: broken references, unreachable options, unset flags
ag new my_game  # scaffold a new project from template
```

## Prototype Success Criterion

One `.agscript` file transpiles to GDScript, a room with basic logic geometry loads in Godot, a character walks to a named point referenced by name in script, and the result runs in Godot without any manual GDScript editing.

## Prototype Milestone Map

| Milestone              | Tasks   | Focus                                |
| ---------------------- | ------- | ------------------------------------ |
| M1 — Godot Fork Setup  | T01–T05 | Fork, module skeleton, ag CLI stub   |
| M2 — AGS-Spirit Parser | T06–T12 | Lexer, AST, parser, symbol table     |
| M3 — GDScript Emitter  | T13–T18 | Emit GDScript, await, source maps    |
| M4 — Room Node         | T19–T24 | AGSRoom, logic geometry nodes        |
| M5 — Character Node    | T25–T29 | AGSCharacter, navigation, WalkTo     |
| M6 — End-to-End Wiring | T30–T36 | Script language integration, runtime |

## High-Risk Tasks

T09 (parser), T16 (await emission), T27 (blocking WalkTo), T30 (script language wiring), T32 (built-in name mapping), T36 (end-to-end). Address these first within each milestone.

## Testing

AGS3D maintains two parallel test layers:

- **Godot C++ tests** in `tests/` — engine internals, do not touch
- **AGS3D GDScript tests** in `agstests/` — run headlessly, one suite per milestone

Test tasks are tracked as `TEST-INFRA-xx` and `TEST-Mx-xx` GitHub issues (separate from T01–T36). Every implementation task that completes a milestone should be followed by implementing its corresponding test suite. Use the `unit-testing` skill for all test work.

## See Also

- [AGS3D Design Document](docs/AGS3D_Design_Document.docx)
- [Prototype Task List](docs/AGS3D_Prototype_Tasks.docx)
