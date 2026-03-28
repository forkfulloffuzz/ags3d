# AGS3D *(working title)*

> **Note:** AGS3D is a working title. The project will be renamed before any public release.

A purpose-built 3D adventure game engine — a Godot 4 fork that encodes classic adventure game concepts (rooms, characters, hotspots, dialogue, inventory) as first-class engine primitives. Authors write in a clean scripting language; the engine handles the 3D navigation, camera systems, and event wiring.

The goal is to make 3D adventure games as approachable as classic 2D point-and-click engines were — no knowledge of 3D coordinates, node trees, or GDScript required.

---

## What It Is

AGS3D is built on three layers:

**1. Engine (Godot fork)**
Custom C++ node types registered directly into Godot's ClassDB: `AGSRoom`, `AGSCharacter`, `AGSCamera`, `AGSWalkableSurface`, `AGSBlockerVolume`, `AGSTriggerRegion`, `AGSHotspot`, `AGSPoint`, `AGSSpawnPoint`. These appear in the editor's Add Node dialog and carry adventure-game semantics natively.

**2. Scripting (AGS-spirit)**
Authors write in `.agscript` files — a C-like scripting language inspired by the original Adventure Game Studio. The `ag` CLI transpiles these to GDScript at build time. Authors never see GDScript.

```agscript
function room_enter() {
    Character("player").walk_to("door_left");
    Character("player").face_to("window");
}

function hotspot_notice_board_Interact() {
    Character("player").say("A notice about the festival.");
}
```

**3. Project format (AGS files)**
All game content is defined in plain-text AGS source files. Godot scenes and scripts are generated — authors never edit `.tscn` or `.gd` files directly.

| File | Purpose |
|------|---------|
| `game.agp` | TOML project manifest — start room, start character, settings |
| `rooms/NAME/NAME.agroom` | Room config — cameras, walkable surfaces, hotspots, spawn points |
| `rooms/NAME/NAME.agscript` | Room logic — event handlers, character commands |
| `characters/NAME.agchar` | Character definition — display name, movement speed |

---

## Current Status

The engine prototype is functional end-to-end:

- A character spawns in a room, navigates around a blocker volume to a named point, then turns to face a second point
- All authored in `.agscript` — no GDScript written by hand
- Navigation uses Godot's `NavigationAgent3D` with a runtime-baked navmesh
- The `AGSCamera` system activates the room's initial camera automatically on load

### Completed Milestones

| Milestone | Focus | Status |
|-----------|-------|--------|
| M1 — Godot Fork Setup | Fork, module skeleton, `ag` CLI stub | ✅ Done |
| M2 — AGS-Spirit Parser | Lexer, AST, recursive-descent parser, symbol table | ✅ Done |
| M3 — GDScript Emitter | Emit GDScript, `await` blocking calls, source maps | ✅ Done |
| M4 — Room Node | AGSRoom, WalkableSurface, BlockerVolume, TriggerRegion, Hotspot, AGSPoint | ✅ Done |
| M5 — Character Node | AGSCharacter, NavigationAgent3D, WalkTo, FaceTo, SpawnPoint | ✅ Done |
| M6 — End-to-End Wiring | Script language integration, AGSRuntime singleton, event binding | ✅ Done |
| M7 — Camera System | AGSCamera node, `initial_camera` room config, AGSRuntime camera registry | 🔄 In progress |
| Tooling — AG Studio | Web-based developer UI for pipeline visualization and project management | ✅ Done |

---

## Architecture

```
Author writes .agroom / .agscript / .agchar
        │
        ▼
  ag build (Go CLI)
  ├── .agscript ──► .engine/generated/*.gd   (GDScript, attached to AGSRoom)
  └── .agroom   ──► rooms/NAME/NAME.tscn     (Godot scene, Phase 1 interim)
        │
        ▼
  Godot editor opens project
  ├── AGSRoom node (C++) reads initial_camera → activates AGSCamera
  ├── AGSSpawnPoint places character at spawn position
  ├── AGSWalkableSurface bakes NavigationMesh at runtime
  └── AGSRuntime singleton indexes all rooms, characters, cameras by name
        │
        ▼
  Room script (generated GDScript) runs:
  └── await character.walk_to("point") → NavigationAgent3D pathfinds → walk_completed signal
```

### Key design rules

- **Authors never see Godot internals.** No `.tscn`, no `.gd`, no 3D coordinates — only AGS source files.
- **All Godot files are generated.** Scenes from `.agroom`, scripts from `.agscript`. The generated files are build artifacts.
- **Blocking calls are first-class.** `walk_to`, `face_to`, `say`, `wait` all pause the script coroutine and resume on completion. The parser identifies blocking calls; the emitter wraps them in `await`.
- **Spatial references use names, never coordinates.** `walk_to("door_left")` not `walk_to(3.1, 0.18, 3.4)`.
- **Runtime behavior in GDScript, node types in C++.** C++ defines the node types and properties; navigation/movement logic lives in `.engine/runtime/ags_character.gd` for fast iteration without recompiling.

---

## Planned Work

### M7 — Camera System (in progress)
- **T38** — `AGSRuntime.set_camera()` global script function
- **T39** — `AGSCameraZone` trigger volume for automatic camera switching on character entry

### M8 — Basic MVP Editor
The goal of M8 is to make Godot itself the authoring surface — `.agroom` files auto-import as live scenes, and the editor hides all internal Godot details from the author.

- **T-M8-01** — `EditorImportPlugin` for `.agroom` — auto-generate scene on save, no manual `ag build` step needed for scene changes
- **T-M8-02** — C++ `.agroom` parser producing a `RoomData` struct consumed by the importer
- **T-M8-03** — `.agchar` importer — character config feeds into generated AGSCharacter nodes
- **T-M8-04** — Hide internal files from the Godot FileSystem dock (no `.gd`, `.tscn`, `.engine/` visible to authors)
- **T-M8-05** — Migrate prototype off hand-maintained `.tscn` once the importer works
- **T-M8-06** — `ag validate` cross-reference checks: catch broken point/camera/character references between `.agroom` and `.agscript` before build

### Future — Runtime Script Embedding (#75)
Runtime GDScripts (currently in `.engine/runtime/`) will be embedded into the C++ module as string constants at build time. In production builds, no `.gd` files exist on disk. A `agvm_runtime_scripts=files` SCons flag keeps the file-based approach available for engine developers during iteration.

### Future — Disable trace in production (#74)
`AGSRuntime` trace logging (currently always enabled) will be controlled by a build flag for production builds.

---

## Repository Layout

```
modules/agvm/          # AGS3D C++ module — node types, AGSRuntime, script language
tools/ag/              # Go CLI: ag build, ag run, ag validate, transpiler pipeline
  cmd/ag/              # CLI entry point
  internal/scanner/    # AGS-spirit lexer
  internal/parser/     # AST + recursive-descent parser + symbol table
  internal/emitter/    # GDScript emitter with await transform and source maps
  internal/project/    # game.agp parsing and build manifest
  internal/lsp/        # Language server (agls)
tools/agui/            # AG Studio — web UI for pipeline visualization
game_prototype/        # End-to-end test project
  game.agp
  characters/
  rooms/start/
    start.agroom       ← room config (source of truth)
    start.agscript     ← room logic (source of truth)
    start.tscn         ← hand-maintained interim; replaced by importer in M8
  .engine/
    generated/         # ag build output — gitignored
    runtime/           # authored GDScript runtime — version controlled
      ags_character.gd # navigation behavior for AGSCharacter
agstests/              # AGS3D GDScript test suites (headless)
docs/                  # Design documents and grammar spec
.dev/                  # Developer workflow scripts
.github/
  agents/              # Specialist agent definitions for AI-assisted development
  skills/              # SKILL files for AGS file type authoring
    ags-room/          # .agroom format reference
    ags-script/        # .agscript authoring guide
    ags-character/     # .agchar format reference
```

---

## Getting Started (Developer)

### Prerequisites

- Godot 4 build dependencies (see [Godot docs](https://docs.godotengine.org/en/stable/contributing/development/compiling/))
- Go 1.21+
- SCons

### Build the engine

```sh
.dev/build.sh          # standard editor build
.dev/build.sh debug    # debug build with extra assertions
```

### Build and run the prototype

```sh
.dev/ag.sh run         # build ag CLI if needed, transpile scripts, launch Godot editor
```

### Run tests

```sh
.dev/test.sh           # AGS3D GDScript test suites (headless)
.dev/test-ag.sh        # Go transpiler unit tests
```

### ag CLI

```sh
ag build               # transpile changed .agscript files to GDScript
ag build --trace       # include debug trace prints in emitted GDScript
ag build --force       # rebuild all files regardless of mtime
ag run                 # build + launch Godot editor
ag validate            # static analysis: broken references, unset flags
ag viz emit FILE       # side-by-side AGS-spirit ↔ GDScript view
ag viz ast FILE        # AST tree
ag viz tokens FILE     # token stream
```

---

## Design Philosophy

Adventure games have a long history of purpose-built engines — AGS, SCUMM, Wintermute — because general-purpose engines impose too much of their own mental model on authors. A game author should think in rooms, characters, and interactions, not scene trees and physics layers.

AGS3D takes the same approach for 3D: Godot is the runtime and editor foundation, but authors never directly encounter it. The scripting language, file formats, and eventual editor surface are all designed around adventure game concepts.

The engine is currently a research prototype. The name AGS3D is a working title chosen to communicate intent — it will change before any public release.

---

## License

Based on [Godot Engine](https://github.com/godotengine/godot), © 2014-present Juan Linietsky, Ariel Manzur, and contributors. Godot is MIT licensed.

AGS3D additions: see LICENSE for details.
