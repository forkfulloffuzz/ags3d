# AGS3D Editor Milestone — Design Document

## Goal

Build **AG Studio** — a specialized adventure-game authoring environment built
directly on top of the Godot editor. AG Studio replaces the generic Godot UI
with AGS3D-specific panels and workflows, so an author never sees Godot
internals. Because AGS3D is already a Godot fork, the editor is part of the
same binary.

At milestone end an author can:

1. Launch the AG Studio binary, create a project, and author rooms, characters,
   and scripts entirely through the custom UI.
2. Press **Play** and the game runs inside the same editor window.
3. Pass `--godot-editor` at launch to switch to the standard Godot editor for
   engine-level debugging.

"Minimal game" scope: one room, one character, room-enter navigation, basic
hotspot interaction. No speech, inventory, or GUI system (those are the next
milestone).

---

## Architecture

AG Studio is Godot running in editor mode, with the editor UI replaced by
AGS3D-specific panels.

```
godot (AGS3D fork)
├── Engine           unchanged — physics, rendering, navigation, C++ nodes
└── Editor
    ├── Standard Godot editor (hidden by default, exposed with --godot-editor)
    └── AG Studio layer (EditorPlugin + C++ editor tweaks)
        ├── Main screen: Room editor (3D viewport + AGS toolbar)
        ├── Main screen: Script editor (AGS-spirit, replaces GDScript editor)
        ├── Left dock: Project panel (rooms, characters, scripts tree)
        ├── Right dock: Inspector (AGS-aware property forms)
        ├── Bottom dock: Build log
        └── Toolbar: Play / Build buttons
```

**Implementation mechanisms (in order of preference):**

1. **GDScript `EditorPlugin`** — hides/shows Godot docks, adds custom main
   screens and docks, registers gizmos. The majority of AG Studio lives here.
2. **C++ editor modifications** — remove menu items, file associations, and
   dock entries that cannot be hidden via the plugin API.
3. **`--godot-editor` flag** — a command-line flag (handled in `main.cpp`)
   that skips the AG Studio plugin activation and boots the standard Godot
   editor as normal.

The Wails desktop app (`tools/agui`) remains a separate developer tool for
inspecting the transpile pipeline. It is **not** AG Studio.

---

## Godot Editor Mode vs AG Studio Mode

| Concern | AG Studio (default) | Godot editor (`--godot-editor`) |
|---|---|---|
| Main screens | Room editor, Script editor | 2D, 3D, Script, AssetLib |
| Scene dock | Hidden — replaced by Project panel | Visible |
| FileSystem dock | Hidden — replaced by Project panel | Visible |
| Inspector | AGS property forms | Standard Godot Inspector |
| Node creation | Via AG Studio room editor only | Via Scene dock |
| File visibility | `.agroom`, `.agchar`, `.agscript` only | All files |
| Use case | Game authoring | Engine debugging |

Switching between modes requires restarting the editor with or without the flag.
A menu item in AG Studio can restart with `--godot-editor` without closing
the project.

---

## Feature Specifications

### F1 — `--godot-editor` Flag

A command-line argument that disables the AG Studio EditorPlugin and runs the
standard Godot editor unmodified.

**Implementation:** In `editor/editor_node.cpp` (or the AG Studio plugin's
`_enter_tree()`), check `OS.get_cmdline_args()` for `--godot-editor`. When
present, skip all AG Studio customisation — do not hide any docks, do not
register custom main screens.

A **"Debug in Godot Editor"** menu entry in AG Studio's top menu calls:

```gdscript
OS.create_process(OS.get_executable_path(),
    ["--editor", "--godot-editor", "--path", ProjectSettings.globalize_path("res://")])
```

---

### F2 — Project Panel (left dock)

Replaces the Scene dock and FileSystem dock with a single AGS-aware tree.

```
▾ rooms/
    ▸ start          (start.agroom · start.agscript)
    ▸ garden
▾ characters/
      player.agchar
      npc.agchar
▾ scripts/
      global.agscript
```

- Single-click a room → opens it in the Room editor.
- Single-click a character → opens Character editor.
- Single-click a script → opens Script editor.
- Right-click → New Room / New Character / New Script / Delete.
- Refresh button (or auto-watch) picks up files added outside AG Studio.

**How to hide Godot's native docks:**

```gdscript
func _enter_tree() -> void:
    # Hide docks AG Studio replaces
    var ei := get_editor_interface()
    ei.get_file_system_dock().hide()
    # Scene dock has no direct accessor; hide by name
    _hide_dock_by_title("Scene")
    _hide_dock_by_title("Import")
    add_control_to_dock(DOCK_SLOT_LEFT_UL, _project_panel)
```

---

### F3 — Room Editor (main screen)

A custom main screen that opens when the author selects a room. It re-uses
Godot's existing **3D viewport** for scene display and adds an AGS-specific
toolbar and property sidebar on top.

**Main screen registration:**

```gdscript
func _has_main_screen() -> bool: return true
func _get_main_screen_name() -> String: return "Room"
func _get_main_screen_icon() -> Texture2D: return preload("icons/room.svg")
func _make_visible(visible: bool) -> void: _room_editor_panel.visible = visible
```

#### 3D viewport

The standard Godot 3D sub-viewport is embedded inside the Room editor panel.
The author sees the room from the `initial_camera` by default; the editor
camera can be orbited freely while authoring.

#### AGS toolbar (above viewport)

Icon buttons for each node type:

```
[ + Floor ] [ + Blocker ] [ + Point ] [ + Camera ] [ + Spawn ] [ + Hotspot ] [ + Region ]
[ Grid 0.5m ▾ ]   [ Snap ]   [ Gizmos ]
```

Clicking an add button inserts the corresponding node at the scene origin; the
author then drags it into position.

#### Gizmos

Custom `EditorNode3DGizmo` / `EditorNode3DGizmoPlugin` for each AGS node type:

| Node | Gizmo |
|---|---|
| `AGSWalkableSurface` | Green flat box with resize handles on XZ edges |
| `AGSBlockerVolume` | Red semi-transparent box with resize handles |
| `AGSPoint` | Named diamond + vertical line to floor |
| `AGSCamera` | Camera frustum + look-at arrow |
| `AGSSpawnPoint` | Character silhouette icon at position |
| `AGSHotspot` | Blue outlined box with resize handles |
| `AGSTriggerRegion` | Purple outlined box with resize handles |

Gizmos respond to drag-to-reposition and handle-to-resize. All edits go
through `EditorUndoRedoManager` so Ctrl+Z works.

#### Property sidebar (right)

When an AGS node is selected, the right dock shows an AGS-aware form instead
of the generic Godot Inspector:

- **AGSRoom**: room name, initial camera dropdown (lists Camera children).
- **AGSWalkableSurface**: size (XZ), offset Y.
- **AGSBlockerVolume**: size (XYZ), position.
- **AGSPoint**: point name, position.
- **AGSCamera**: camera name, position, look-at position.
- **AGSSpawnPoint**: character dropdown (lists known `.agchar` names), position.
- **AGSHotspot**: hotspot name, size, position.
- **AGSTriggerRegion**: region name, size, position.

**Implementation:** `EditorInspectorPlugin` that returns custom controls for
AGS node types, suppressing Godot's default property list for those types.

#### Saving

The Room editor works directly on the `.tscn` scene (generated once by
`ag build` the first time, then edited in-place by the gizmos). It also writes
back changes to the `.agroom` source file so the two stay in sync.

Sync direction:
- **Editor gizmo move/resize → update `.agroom`**: after each undoable action,
  serialise the scene state back to `.agroom` format.
- **`.agroom` external edit → regenerate `.tscn`**: file watcher triggers
  `ag build <file>` and reloads the scene.

---

### F4 — Character Editor (main screen or dock)

A simple property form that opens when the author selects a `.agchar` file.

| Field | Widget |
|---|---|
| Internal name | Read-only label |
| Display name | Text input |
| Mesh file | File picker (`.glb`, `.obj`) — optional; default capsule |

Save writes the `.agchar` file and triggers `ag build <file>`.

---

### F5 — Script Editor (main screen)

A custom main screen that replaces the Godot Script editor tab for `.agscript`
files. Godot's built-in script editor remains available for `.gd` files but
is not the primary authoring surface.

**Main screen:** similar to Godot's built-in CodeEdit-based editor, but:

- Syntax highlighting for AGS-spirit keywords, built-in function names,
  blocking call markers, event handler names.
- File tab bar listing open `.agscript` files.
- Error / warning gutters populated from `ag validate` output.
- **Ctrl+S** saves and triggers `ag build` on the file; errors appear in the
  gutter and Build Log immediately.

**Blocking call indicator:** calls that emit `await` in generated GDScript
(e.g. `WalkTo`, `Say`, `Wait`) are annotated with a small clock icon in the
gutter so the author knows the script will pause there.

**Implementation:** A `Control`-based panel using Godot's `CodeEdit` node,
registered as a custom main screen. AGS-spirit syntax rules are encoded as
`CodeHighlighter` keyword/region entries.

---

### F6 — Scene Generator (`ag build` → `.tscn`)

The Go CLI must generate `.tscn` files from `.agroom` and `.agchar` before the
editor can display them. This is the backend work that the editor depends on.

#### `.agroom` → `.tscn`

| `.agroom` block | Godot node | Key properties |
|---|---|---|
| `Room` (root) | `AGSRoom` | `room_name`, `initial_camera`, `script` |
| `Camera` | `AGSCamera` | `camera_name`, transform |
| `Point` | `AGSPoint` | `point_name`, transform |
| `WalkableSurface` | `AGSWalkableSurface` | `BoxMesh` + `CollisionShape3D` from size |
| `BlockerVolume` | `AGSBlockerVolume` | `CollisionShape3D` from size + position |
| `SpawnPoint` | `AGSSpawnPoint` | `spawn_character`, transform |
| `Hotspot` | `AGSHotspot` | `hotspot_name`, `CollisionShape3D` |
| `TriggerRegion` | `AGSTriggerRegion` | `region_name`, `CollisionShape3D` |

UIDs are deterministic (hashed from room name).

#### `.agchar` → `.tscn`

```
AGSCharacter (root)
  character_name = <name>
  script = res://.engine/runtime/ags_character.gd
  MeshInstance3D  (capsule or custom mesh)
  CollisionShape3D
```

#### `ag validate`

Run as a pre-build step. See cross-reference rules in F7.

---

### F7 — `ag validate`

| Rule | Error example |
|---|---|
| `SpawnPoint.character` resolves to a `.agchar` | `start.agroom:27: unknown character "hero"` |
| `initial_camera` matches a Camera block in same room | `start.agroom:2: camera "side" not defined in this room` |
| Point names in `.agscript` exist in the room's `.agroom` | `start.agscript:4: unknown point "exit" in room "start"` |
| Character names in `.agscript` exist as `.agchar` | `start.agscript:4: unknown character "npc1"` |
| `game.agp` paths resolve to real files | `game.agp:3: start_room not found` |

---

### F8 — Build Log Dock (bottom)

A bottom dock showing the output of the most recent `ag build` / `ag validate`
run.

- Each line is parsed; errors are shown in red with a clickable file+line link
  that jumps the Script editor to the offending line.
- A **Build** button (also in the toolbar) triggers a full project build.
- A **Clear** button clears the log.

**Implementation:** `EditorPlugin.add_control_to_bottom_panel()` with a
`RichTextLabel` for clickable links.

---

### F9 — Play Button

A **Play** button in the top toolbar (replacing or alongside Godot's play
buttons) that:

1. Triggers `ag build` on the whole project.
2. If build succeeds: runs the game via Godot's normal **Play** action
   (`EditorInterface.play_main_scene()` or the existing play shortcut).
3. If build fails: displays errors in the Build Log; does not run.

Because AG Studio is Godot, the game runs inside the same process — exactly
as pressing F5 in the standard editor. No subprocess is needed.

---

### F10 — Project Wizard

**New Project** action in the AG Studio menu:

1. Native folder picker for project root.
2. Enter project name and initial room name.
3. AG Studio writes the scaffold:
   - `game.agp`
   - `rooms/<name>/<name>.agroom` (default 10×10 floor, one camera)
   - `rooms/<name>/<name>.agscript` (empty `room_enter` stub)
   - `characters/player.agchar`
4. Calls `ag build` to generate the initial `.tscn`.
5. Opens the project in the Room editor.

**Implementation:** GDScript backend method that shells out to `ag` for
scaffold generation, or writes the files directly.

---

### F12 — Billboard Camera Warnings

**What it does:** Detects camera configurations in a room that would produce
visual artefacts for billboard-mode characters, and surfaces them as
editor warnings — in the room editor gizmo overlay and the Build Log dock.

Two warning classes:

#### W1 — Camera elevation too steep

**Condition:** A camera's elevation angle (the angle between the camera
position and the `look_at` point, measured from the XZ plane) exceeds **30°**
AND the room contains at least one character with `visual_mode = "billboard"`.

**Why it matters:** Billboard quads always face the camera horizontally. A
steep downward angle reveals the top edge of the sprite, which has no art.

**Editor display:**
- The camera gizmo in the Room editor shows a yellow warning icon.
- Hovering the icon shows: *"Camera elevation [N°] may clip billboard
  character sprites. Recommended: keep below 30°."*
- The Build Log lists it as `WARNING` (not an error — does not block Play).

#### W2 — Single-angle sprite, camera orbit not locked

**Condition:** A character uses `sprite_angles = 1` AND the room's active
camera does not have `sprite_locked = true`.

**Why it matters:** Single-direction art only looks correct from one angle. If
the camera can orbit the character freely (default), the sprite will be shown
from the wrong side.

**Editor display:**
- Warning on the camera node gizmo: *"Room has single-angle billboard
  characters but camera is not sprite_locked. Add sprite_locked = true to
  restrict orbiting."*
- Also emitted as a `WARNING` line in the Build Log, with a clickable link to
  the camera node in the Room editor.

#### W3 — Arc coverage too wide for 4-angle sprites

**Condition:** The camera's position spans a horizontal arc greater than **45°**
relative to the room origin AND the room has characters with
`sprite_angles = 4`.

**Why it matters:** The gap between 4-way directions is 90°. If gameplay
allows the camera to rotate more than 45° from a cardinal axis, characters
will snap direction visibly.

**Editor display:** Same pattern — yellow gizmo icon on the camera + Build Log
`WARNING`.

#### Implementation

- Warning logic runs as part of `ag validate` (a new `--warn` pass that does
  not abort the build).
- In the editor, the Room editor re-evaluates warnings whenever:
  - A camera node is moved or its `look_at` changes.
  - A character's `visual_mode` or `sprite_angles` changes.
- Gizmo warning overlays are drawn by the camera's `EditorNode3DGizmo`
  subclass (added in T-E10) using `add_unscaled_billboard()` for the icon.

---

### F11 — Prototype Migration

Delete `game_prototype/rooms/start/start.tscn` (hand-maintained) and verify
`ag build` regenerates it with functionally identical behaviour.

**Acceptance:** all M6 end-to-end tests pass against the generated scene.

---

## Minimal Game Walkthrough

An author with no Godot or coding experience:

1. Launches the AG Studio binary.
2. **New Project** → picks a folder, enters name.
3. In the **Room editor**: resizes the walkable floor, adds a blocker, places
   two named points (`door`, `window`), positions the camera.
4. In the **Character editor**: sets display name.
5. In the **Script editor**: writes:
   ```
   function room_enter() {
       player.WalkTo("door");
       player.FaceTo("window");
   }
   ```
6. Adds a Hotspot in the Room editor, adds a `hotspot_interact` handler in
   the script.
7. Presses **Play** — the game runs in the same window.
8. Closes the game. Edits the script. Presses **Play** again.

---

## Task Breakdown

| Task | Description | Depends on |
|---|---|---|
| T-E01 | Go: `.agroom` parser → `RoomData` struct | — |
| T-E02 | Go: `RoomData` → `.tscn` serialiser | T-E01 |
| T-E03 | Go: `.agchar` parser + `CharData` → `.tscn` | — |
| T-E04 | Go: wire scene gen into `ag build` pipeline | T-E01, T-E02, T-E03 |
| T-E05 | Go: `ag validate` cross-reference checks | T-E01, T-E03 |
| T-E06 | C++/GDScript: `--godot-editor` flag (F1) | — |
| T-E07 | GDScript: `EditorPlugin` skeleton — hide Godot docks, register screens | T-E06 |
| T-E08 | GDScript: Project panel dock (F2) | T-E07 |
| T-E09 | GDScript: Room editor main screen + 3D viewport embed (F3) | T-E07, T-E04 |
| T-E10 | GDScript: AGS gizmo plugins for all node types (F3) | T-E09 |
| T-E11 | GDScript: AGS Inspector plugin — property forms (F3) | T-E09 |
| T-E12 | GDScript: `.agroom` ↔ `.tscn` sync (edit in editor → write agroom) | T-E10, T-E11 |
| T-E13 | GDScript: Character editor main screen (F4) | T-E07 |
| T-E14 | GDScript: Script editor main screen — CodeEdit + AGS highlighting (F5) | T-E07 |
| T-E15 | GDScript: Build Log dock (F8) | T-E07 |
| T-E16 | GDScript: Play button wired to build + `play_main_scene()` (F9) | T-E04, T-E15 |
| T-E17 | GDScript: Project wizard (F10) | T-E08 |
| T-E18 | Integration: prototype migration (F11) | T-E04 |
| T-E19 | GDScript: billboard camera warnings — elevation/arc/lock checks in `ag validate` + gizmo overlays (F12) | T-E10, T-GS24 |
| T-E20 | GDScript: Character editor — type selector (3D/2D) with type-specific property sections (F4 extension) | T-E13, T-GS27 |
| T-E21 | GDScript: 3D Animation viewer — embedded `SubViewport`, clip selector, transport controls, frame scrubber | T-E20, T-GS28 |
| T-E22 | GDScript: 2D Animation viewer — sprite sheet grid, direction thumbnails, animated preview cell | T-E20, T-GS29 |

Critical path to first playable build: T-E01 → T-E02 → T-E03 → T-E04 →
T-E07 → T-E09 → T-E16.

---

## Out of Scope

- Speech / inventory / GUI systems (next milestone).
- LSP integration in the script editor.
- Undo/redo beyond what `EditorUndoRedoManager` gives for free.
- Multi-room room-transition authoring.
- Export / packaging for distribution.
- The `tools/agui` Wails app is unaffected — it remains a developer-facing
  transpile visualiser, not the game authoring editor.
