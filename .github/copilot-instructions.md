# AGS3D Project Guidelines

## What This Is

AGS3D is a purpose-built 3D adventure game engine — a Godot fork that encodes adventure game concepts (rooms, characters, inventory, dialogue, hotspots) as first-class engine primitives. Authors write in AGS-spirit scripting; the engine transpiles it to GDScript at build time.

## Repository Layout

```
modules/agvm/          # AGS3D C++ module — all AGS-specific engine code lives here
tools/ag/              # Go module: ag CLI + transpiler pipeline + agls language server
  cmd/ag/              # ag CLI entry point
  cmd/agls/            # language server entry point
  internal/project/    # game.agp parsing, directory scanner, build manifest
  internal/scanner/    # lexer / tokenizer (T07)
  internal/parser/     # AST, recursive descent parser, symbol table (T08-T11)
  internal/emitter/    # GDScript emitter, await transform, source maps (T13-T17)
  internal/analysis/   # static analysis shared by validate + LSP
  internal/lsp/        # LSP server and request handlers
game_prototype/        # Minimal test project used for end-to-end validation
  game.agp             # TOML project manifest
  characters/          # .agchar files
  rooms/               # .agroom and .agscript files
  .engine/             # Mixed: generated artifacts + authored runtime scripts
    generated/         # Transpiled GDScript output (gitignored, regenerated each build)
    cache/             # Build manifest (gitignored)
    runtime/           # Authored GDScript runtime behavior — version-controlled, never generated
      ags_character.gd # Navigation behavior for AGSCharacter (walk_to, face_to, _physics_process)
```

## Tech Stack

- **Engine base**: Godot 4 (C++), forked — minimize diff surface against upstream
- **Module**: `modules/agvm/` — all AGS3D code, registered via SCsub
- **Scripting input**: AGS-spirit (`.agscript`) — AGS-compatible syntax, new semantics
- **Scripting output**: GDScript (`.gd`) — generated build artifact, never hand-edited
- **CLI tool + transpiler + LSP**: Go — `tools/ag/` is a single Go module containing `ag` CLI, the full parser/emitter pipeline, and the `agls` language server
- **Build artifact dir**: `.engine/generated/` — deleted and regenerated on each build

## Key Architecture Rules

- The transpiler (parser → emitter) is the highest-risk component. The parser (T09) and await emission (T16) are the hardest tasks — invest in correctness first.
- Generated GDScript is a build artifact. Source maps (`.agmap`) translate runtime errors back to AGS-spirit file and line number. Authors must never see a GDScript path.
- Authors never reference 3D coordinates. All spatial references use named points (`point.door_left`).
- Logic geometry (WalkableSurface, BlockerVolume, TriggerRegion, HotspotSurface) is invisible at runtime — it exists only for pathfinding, collision, and hit testing.
- Blocking calls in AGS-spirit (`WalkTo`, `PlayAnimation`, `Wait`) must emit `await` in GDScript. The parser annotates blocking call sites; the emitter acts on those annotations.
- `.engine/generated/` and `.engine/cache/` are gitignored build artifacts — never commit them.
- `.engine/runtime/` is version-controlled authored GDScript — commit changes there like any source file.
- Navigation/movement behavior for AGSCharacter lives in `.engine/runtime/ags_character.gd`, not in C++.
- Room cameras: set `initial_camera` on AGSRoom to a camera name; AGSRuntime activates it before room scripts run.

## Dev Scripts

All developer workflow scripts live in `.dev/` (hidden, version-controlled).
When implementing a task that needs a helper script (one-off migration, code generator, data validator, etc.), write it here.

```sh
.dev/build.sh              # standard editor build (wraps scons)
.dev/build.sh debug        # debug build with extra assertions
.dev/build.sh release      # release template
.dev/build.sh clean        # wipe compiled objects
.dev/build.sh -- EXTRA=1   # passthrough arbitrary scons args
.dev/test.sh               # run all GDScript test suites headlessly
.dev/test.sh --verbose     # show raw Godot output
.dev/test.sh --filter M1   # filter output by pattern
```

Other scripts in `.dev/` are task-specific — read the file header for usage. When adding a new script, make it executable (`chmod +x`) and add a usage comment at the top.

## Build & Run (ag CLI)

```sh
ag build     # parse all changed .agscript files, emit GDScript to .engine/generated/
ag run       # build + launch Godot editor with the project
ag validate  # static analysis: broken references, unreachable options, unset flags
ag new my_game  # scaffold a new project from template

ag viz tokens <file>    # token stream table: line/col/kind/lexeme (VIZ-01, after T07)
ag viz ast <file>       # AST indented tree with node types and source positions (VIZ-02, after T09)
ag viz blocking <file>  # which call sites are annotated blocking=true (VIZ-03, after T11)
ag viz emit <file>      # side-by-side AGS-spirit ↔ GDScript with source-map links (VIZ-04, after T17)
ag viz <file>           # all four stages in sequence
```

```sh
.dev/build-ag.sh          # build ag + agls → bin/
.dev/build-ag.sh ag       # build ag only
.dev/build-ag.sh clean    # remove binaries
.dev/ag.sh build          # auto-rebuild ag if stale, then run ag build
.dev/ag.sh run            # auto-rebuild ag if stale, then run ag run
.dev/ag.sh new <name>     # scaffold new project
.dev/test-ag.sh           # go test ./... (transpiler unit tests)
.dev/test-ag.sh --verbose # show individual test names
.dev/test-ag.sh --filter scanner  # run tests matching pattern
.dev/test-ag.sh --cover   # generate coverage report → bin/ag-cover.out
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

AGS3D maintains three parallel test layers:

- **Godot C++ tests** in `tests/` — engine internals, do not touch
- **AGS3D GDScript tests** in `agstests/` — run headlessly, one suite per milestone (M1 module/script-language, M4 room nodes, M5 character nodes, M6 integration)
- **Go tests** in `tools/ag/` — `go test ./...` covers the transpiler (lexer, parser, emitter, source maps) and CLI logic (M2–M3 tasks T07–T17)

Test tasks are tracked as `TEST-INFRA-xx` and `TEST-Mx-xx` GitHub issues (separate from T01–T36). Every implementation task that completes a milestone should be followed by implementing its corresponding test suite. Use the `unit-testing` skill for all test work.

## AGS-Spirit Language

The grammar is formally specified in [`../docs/grammar.md`](../docs/grammar.md). That document is the authoritative source for token types (scanner), AST node shapes (parser), and emit rules (emitter). Any language change must start there.

Key points:
- C-like syntax: `function`, `if`/`else`, `while`, `for`, `switch`, typed variables (`int x = 5;`)
- Blocking calls (`WalkTo`, `Say`, `Think`, `Wait`, `WaitKey`, `FadeIn`, `FadeOut`, etc.) emit `await` in GDScript
- Spatial references always use named points (`point.NAME`), never raw coordinates
- Event handlers follow the naming convention: `room_Load`, `hotspot_NAME_Interact`, etc.
- `true`, `false`, `null`, `global`, `public` are keywords, not identifiers
- `global.player` / `global.room` / `global.score` — engine-owned game state namespace; never a magic variable
- Type system is **structural** (Go-style): functions accept anything with the right shape, no inheritance
- Visibility: functions are **file-scoped by default**; cross-file sharing requires a `namespace` block with `export function` — called as `X.Func()`; `export` outside a namespace is a transpiler error; the symbol table errors on duplicate exported names within the same namespace
- `GlobalExpr` AST node represents `global.NAME` accesses — emitter maps these to `AGSRuntime` properties

## See Also

- [AGS-Spirit Grammar](../docs/grammar.md) ← start here for any language / parser / emitter work
- [AGS3D Design Document](../docs/AGS3D_Design_Document.docx)
- [Prototype Task List](../docs/AGS3D_Prototype_Tasks.docx)
