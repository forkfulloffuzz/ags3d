# AGS3D Custom Editor Milestone (M12) — Design Document

## Goal

Build the full **AG Studio** custom editor — the specialized adventure-game
authoring UI that replaces the standard Godot editor UI with AGS3D-specific
panels and workflows. This milestone is intentionally deferred until all engine
features work and are usable through the normal Godot editor, so the UI is
built on a stable, proven engine foundation.

At milestone end:
- Authors never see raw Godot UI — everything is AGS-specific panels.
- All engine features built in M9/M10/M11 are exposed through AG Studio.
- The prototype addon in `game_prototype/addons/ag_studio/` is replaced with
  the polished production version.

> **Note on prototype work:** The `game_prototype/addons/` directory contains
> early prototype implementations of many of these features (character editor,
> script editor, build log, play button, project wizard). That code is
> exploratory. The tasks below are the production implementations — some will
> reuse prototype code, others will be rewrites.

---

## Architecture (unchanged from M9 design)

```
godot (AGS3D fork)
└── Editor
    ├── Standard Godot editor  (hidden by default; accessible via --godot-editor)
    └── AG Studio layer  (EditorPlugin + C++ tweaks)
        ├── Main screen: Room editor  (3D viewport + AGS toolbar + gizmos)
        ├── Main screen: Script editor  (CodeEdit + AGS highlighting + gutters)
        ├── Main screen: Character editor  (property form + animation preview)
        ├── Left dock: Project panel  (rooms / characters / scripts tree)
        ├── Right dock: Inspector  (AGS-aware property forms per node type)
        └── Bottom dock: Build log  (ag build / ag validate output)
```

---

## Task Breakdown

Tasks are ordered by dependency. Foundation first, feature editors after.

### Foundation

| Task | Description | Depends on |
|------|-------------|------------|
| T-CE01 | C++: `--godot-editor` command-line flag — skip AG Studio plugin, boot standard Godot editor unmodified | — |
| T-CE02 | GDScript: `EditorPlugin` skeleton — hide Godot docks (Scene, FileSystem, Import), register custom main screens, hook `_enter_tree` / `_exit_tree` cleanly | T-CE01 |
| T-CE03 | GDScript: Project panel dock — rooms/characters/scripts tree; single-click opens the relevant editor main screen; right-click for New/Delete | T-CE02 |

### Room Editor

| Task | Description | Depends on |
|------|-------------|------------|
| T-CE04 | GDScript: Room editor main screen — embed Godot's 3D sub-viewport; AGS toolbar with Add buttons for each node type (Floor, Blocker, Point, Camera, Spawn, Hotspot, Region) | T-CE02 |
| T-CE05 | GDScript: AGS gizmo plugins — `EditorNode3DGizmoPlugin` for every AGS node type (WalkableSurface green box, BlockerVolume red box, Point diamond, Camera frustum, SpawnPoint icon, Hotspot blue box, TriggerRegion purple box); resize handles on volume types | T-CE04 |
| T-CE06 | GDScript: `.agroom` ↔ `.tscn` sync — gizmo drag/resize → write back to `.agroom`; external `.agroom` edit → re-run `ag build` + reload scene | T-CE05 |
| T-CE07 | GDScript: Room editor — `RoomItem` placement gizmo + toolbar "Add Item" button; shows item sprite icon at position | T-CE05 |
| T-CE08 | GDScript: Room editor — "Re-import from Blender" button; detects `.glb` newer than `.tscn`, re-runs `ag build`, shows conflict banner when `.agroom` was edited in AG Studio after last Blender export | T-CE06 |
| T-CE09 | GDScript: Room editor — billboard camera warnings as gizmo overlays; elevation angle warning (>30°), arc width warning for 4-angle sprites, sprite_locked indicator | T-CE05 |

### Inspector

| Task | Description | Depends on |
|------|-------------|------------|
| T-CE10 | GDScript: `EditorInspectorPlugin` — AGS-aware property forms for all AGS node types (AGSRoom, AGSWalkableSurface, AGSBlockerVolume, AGSPoint, AGSCamera, AGSSpawnPoint, AGSHotspot, AGSTriggerRegion, AGSRoomItem) suppressing the generic Godot inspector for those types | T-CE02 |

### Character Editor

| Task | Description | Depends on |
|------|-------------|------------|
| T-CE11 | GDScript: Character editor main screen — property form for `.agchar`; type selector (3D / 2D / puppet) with type-specific sections; speech section (colour, font); save → `ag build` | T-CE02 |
| T-CE12 | GDScript: Character editor — 3D animation viewer: embedded `SubViewport`, clip selector dropdown, transport controls (play/stop/loop), frame scrubber | T-CE11 |
| T-CE13 | GDScript: Character editor — 2D animation viewer: sprite sheet grid, per-direction thumbnail strip, animated preview cell | T-CE11 |
| T-CE14 | GDScript: Character editor — Animations section listing clip names from `.glb` with preview button; feeds into T-CE12/T-CE13 | T-CE12, T-CE13 |

### Script Editor

| Task | Description | Depends on |
|------|-------------|------------|
| T-CE15 | GDScript: Script editor main screen — full production implementation: tab bar for open files, error/warning gutters from `ag validate` output, Ctrl+S saves + triggers build, blocking-call clock icon in gutter | T-CE02 |

### Build Log & Play

| Task | Description | Depends on |
|------|-------------|------------|
| T-CE16 | GDScript: Build Log dock — bottom panel with `RichTextLabel`; clickable error links jump Script editor to offending line; Clear button | T-CE02 |
| T-CE17 | GDScript: Play button — top toolbar button; runs `ag build` on whole project, plays main scene if build succeeds, shows errors in Build Log if it fails | T-CE16 |

### Project Wizard

| Task | Description | Depends on |
|------|-------------|------------|
| T-CE18 | GDScript: Project wizard — AG Studio menu; folder picker; name + start room inputs; writes `game.agp` + scaffold files; runs `ag build`; opens project in new editor window | T-CE02 |

### Other Panels

| Task | Description | Depends on |
|------|-------------|------------|
| T-CE19 | GDScript: Item editor panel — list + create/delete items; name, display name, description, sprite picker; save → `items/<name>.agitem` | T-CE03 |
| T-CE20 | GDScript: GUI layout editor — canvas for `.agui` files; drag widgets (InventoryBar, VerbBar, StatusLine) onto screen-sized canvas; save → `ag build` | T-CE03 |
| T-CE21 | GDScript: Global variables editor — section in project settings panel; add/remove/edit globals from `game.agp [globals]`; writes back on save | T-CE03 |

---

## Pending UI Stubs

These are engine features already shipped that need AG Studio UI. Add new stubs
here whenever an engine/runtime/tooling task is completed.

| Stub | Engine feature | What the UI should do |
|------|---------------|----------------------|
| T-CE-UI01 | WalkableSurface (T-E02) | Toolbar "Add Floor" button; green box gizmo with XZ resize handles; Inspector shows Size, Offset Y |
| T-CE-UI02 | BlockerVolume (T-E02) | Toolbar "Add Blocker" button; red box gizmo with XYZ resize handles |
| T-CE-UI03 | TriggerRegion (T-GS10) | Toolbar "Add Region" button; purple box gizmo with XYZ resize handles; Inspector shows region_name |
| T-CE-UI04 | SpawnPoint (T-E02) | Toolbar "Add Spawn" button; character name dropdown in Inspector (lists known `.agchar` names) |
| T-CE-UI05 | Hotspot (T-E02) | Toolbar "Add Hotspot" button; blue box gizmo with XYZ resize handles |
| T-CE-UI06 | AGSCamera (T-E02) | Toolbar "Add Camera" button; frustum gizmo with look-at handle; Inspector shows name, position, look_at |
| T-CE-UI07 | Room transitions / `load_room` (T-GS10) | SpawnPoint Inspector shows "used as target by: GoToRoom(…)" (read-only informational) |
| T-CE-UI08 | `say()` / `think()` (T-GS01, T-GS05) | Character editor Speech section: speech_colour picker, speech_font file picker; script editor autocomplete for Say/Think on character receivers |
| T-CE-UI18 | `AddInventory`/`LoseInventory`/`HasInventory` emitter (T-GS05) | Script editor autocomplete for inventory methods on character receivers; no separate UI panel needed |
| T-CE-UI09 | `ag validate` cross-reference checks (T-E05) | Build Log shows validate warnings/errors with clickable links; gutters in Script editor |
| T-CE-UI10 | RoomItem node + HideRoomItem/ShowRoomItem (T-GS03, T-GS06) | Room editor "Add Item" button; sprite icon at position; Inspector shows item reference + visible toggle; script editor autocomplete for HideRoomItem/ShowRoomItem |
| T-CE-UI11 | Inventory system (T-GS02/T-GS04, when done) | Item editor panel (T-CE19) |
| T-CE-UI12 | GUI system / .agui (T-GS13, when done) | GUI layout editor (T-CE20) |
| T-CE-UI13 | Global variables (T-GS07/T-GS08, when done) | Global variables editor (T-CE21) |
| T-CE-UI14 | Billboard characters / Sprite3D (T-GS24-T-GS26, when done) | Character editor 2D section: sprite_angles selector, frame_size inputs, sprite sheet file picker |
| T-CE-UI15 | Character type split (T-GS27, when done) | Character editor type selector (3D / 2D / puppet) with type-specific property sections |
| T-CE-UI16 | 3D animation clips (T-BL12/T-BL13, when done) | 3D animation viewer (T-CE12, T-CE14) |
| T-CE-UI17 | Blender .glb visual mesh in rooms (T-BL11, when done) | Room editor shows visual mesh from .glb alongside gizmos; "Re-import from Blender" button (T-CE08) |
| T-CE-UI18 | Cutscene support (T-GS18/T-GS19) | Cutscene panel in script editor: fade-in/out preview button, player-control toggle indicator; `Wait`, `FadeIn`, `FadeOut`, `SetPlayerControl` shown as blocking call annotations |
| T-CE-UI19 | Audio system (T-GS11/T-GS12) | Project panel "Audio" section: lists files in `audio/music/` and `audio/sfx/`; drag-to-assign into script editor; `PlayMusic`/`PlaySound` calls show the file icon inline |
| T-CE-UI20 | Save / Load (T-GS16/T-GS17) | Save slot manager panel: list of used slots, slot names/timestamps, delete button; `SaveGame`/`LoadGame` autocomplete in script editor |
| T-CE-UI21 | Billboard char scene gen (T-GS25) | Character editor 2D tab: visual_mode selector (mesh/billboard); sprite_sheet file picker; sprite_angles, frame_size, frames_per_angle fields; `ag build` preview shows Sprite3D tree |
| T-CE-UI22 | Blender add-on scaffold (T-BL01) | No AG Studio UI needed for scaffold; once panels exist (T-BL02+), AG Studio "Open in Blender" button launches Blender with the room's .blend file |
| T-CE-UI23 | AGS3D object type panel (T-BL02) | No AG Studio UI; the panel lives in Blender. AG Studio docs/tooltip should mention tagging workflow |
| T-CE-UI24 | Billboard direction runtime (T-GS26) | Character editor 2D tab: live preview of direction/frame selection based on simulated camera angle; sprite_locked toggle |
| T-CE-UI25 | AGSAnimationPlayer2D (T-GS29) | Character editor 2D animation section: frames_per_state spinbox, fps spinbox; state preview buttons (Idle/Walk/Talk) |
| T-CE-UI26 | NavMesh baking (T-BL09) | Room editor "Bake NavMesh" button (or auto on export); NavMesh overlay toggleable in 3D viewport; generated AGS_NavMesh mesh highlighted in a dedicated colour |
| T-CE-UI26 | GUI scene generator (T-GS13) | GUI editor: .agui file list in Project panel; layout canvas showing InventoryBar/VerbBar/StatusLine at their anchors; property inspector for each widget |
| T-CE-UI27 | .agroom import operator (T-BL08) | No AG Studio UI needed; round-trip is Blender→AG Studio; AG Studio "Re-import from Blender" button (T-CE08) uses the export operator (T-BL04), not the import |
| T-CE-UI28 | Character export operator (T-BL10) | Character editor "Export to Blender" button; shows NLA track → clip name mapping; links to .agchar animations block |
| T-CE-UI29 | Room export operator (T-BL04) | Room editor "Export to Blender" button; opens Blender with AGS_Gameplay collection pre-tagged from the .agroom |
| T-CE-UI30 | Dialogue node graph (T-DLG01/02) | Visual dialogue node editor: shows nodes as cards, options as edges, jump targets as arrows |
| T-CE-UI31 | Locale manager panel (T-DLG13) | Project settings panel listing declared locales; add/remove/edit name+RTL flag; base_locale picker; fallback_chain reorder list |
| T-CE-UI32 | Dialogue build log (T-DLG07) | Build log panel showing .agdlg → .json output, DLG-E/W error codes inline |
| T-CE-UI33 | Cutscene format validator output (T-CUT05) | Build log shows CUT-E/W error codes with source file and line |
| T-CE-UI34 | Cutscene input bindings panel (T-CUT04) | Project settings panel for [input] bindings: dialogue_advance, cutscene_skip, dialogue_hold_advance |

---

## Out of Scope

- LSP / language server integration in the script editor.
- Multi-room transition graph view.
- Export / distribution packaging.
- Undo/redo beyond what `EditorUndoRedoManager` gives for free.
