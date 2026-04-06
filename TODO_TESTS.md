# AGS3D — Manual Tests

One test section per TODO task. Update this file immediately after marking a task complete in `TODO.md`.

---

## M10 — Game Systems (Batch 2)

### T-GS25 — Billboard `.agchar` properties + `ag build` Sprite3D scene

**Setup:** Have `ag` CLI built (`go build ./cmd/ag`). Create a test `.agchar` file.

- [ ] `Character "guard" { visual_mode = "billboard" }` — `ag build` outputs a `.tscn` with root node `type="AGSCharacter2D"` and `visual_mode = "billboard"`.
- [ ] `Character "guard" { type = "2d" }` — same result as above (existing syntax still works).
- [ ] `Character "guard" { visual_mode = "billboard" sprite_sheet = "assets/sprites/guard.png" sprite_angles = 8 frames_per_angle = 6 }` — generated `.tscn` contains `[ext_resource type="Texture2D" path="res://assets/sprites/guard.png"]`, `hframes = 6`, `vframes = 8`, and a `Sprite3D` child.
- [ ] `Character "hero" { visual_mode = "mesh" }` — `ag build` outputs root node `type="AGSCharacter3D"` and `visual_mode = "mesh"`.
- [ ] `Character "hero" {}` (no visual_mode) — same as above (default is 3D/mesh).
- [ ] `Character "x" { visual_mode = "puppet" }` — `ag validate` / parse reports error `visual_mode must be "mesh" or "billboard"`.
- [ ] Open the generated billboard `.tscn` in Godot. The `AGSCharacter2D` node should appear in the scene tree; `visual_mode` property shows `billboard` in the Inspector.

### T-BL12 — `.agchar` animation clip wiring in generated `.tscn`

**Setup:** Have `ag` CLI built. Create a test `.agchar` with a `mesh` + `animations` block.

- [ ] `Character "player" { mesh = "characters/player/player.glb" animations = { idle = "Idle" walk = "Walk" talk = "Talk" } }` — `ag build` generates a `.tscn` containing `anim_idle = "Idle"`, `anim_walk = "Walk"`, `anim_talk = "Talk"` as properties on the root `AGSCharacter3D` node.
- [ ] Animation properties are emitted in sorted order (idle < talk < walk alphabetically) — regenerating the same `.agchar` twice produces identical `.tscn` output.
- [ ] A `.agchar` with no `animations` block generates a `.tscn` with no `anim_` properties.
- [ ] Open the generated `.tscn` in Godot — the `AGSCharacter3D` root node Inspector shows `anim_idle`, `anim_walk`, `anim_talk` fields with the correct clip name strings.

### T-BL13 — `ags_character.gd` drives AnimationPlayer on state transitions

**Setup:** Generate a character `.tscn` from a `.agchar` that has `mesh` + `animations` (T-BL12). Open in Godot with a playable room.

- [ ] Run the game. Place the character in a room. Trigger `walk_to()` — the character should play the `walk` animation clip immediately.
- [ ] When the character reaches the destination, it should switch back to the `idle` clip.
- [ ] Trigger `say()` — the character should play the `talk` clip while the dialogue line is displayed, then return to `idle` after it finishes.
- [ ] A character with no `animations` block in `.agchar` (no `anim_*` properties in `.tscn`) should navigate and speak without errors — no AnimationPlayer driven, no warnings.
- [ ] A character with `anim_walk = "Walk"` but the `.glb` AnimationPlayer missing a "Walk" clip should log a warning (`ags_character: clip 'Walk' not found`) but not crash.

### T-BL09 — NavMesh baking from WalkableSurface objects

**Setup:** Blender with the AGS3D add-on installed. Create a room `.blend` with at least two box objects tagged `AGS_type = "WALKABLE"`.

- [ ] Open the `.blend`. Run **Object → AGS3D → Bake NavMesh** (or press F3 and search "Bake NavMesh"). An `AGS_NavMesh` object should appear in a new `AGS_NavMesh` collection.
- [ ] The `AGS_NavMesh` mesh contains one flat quad per WalkableSurface, positioned at the top Y of each box's bounding box.
- [ ] Running "Bake NavMesh" again replaces the existing `AGS_NavMesh` (no duplicate objects).
- [ ] With no WalkableSurface objects in the scene, running "Bake NavMesh" reports a warning and does not create any object.
- [ ] Run **File → Export → AGS3D Room (.agroom + .glb)**. The resulting `.glb` contains a node named `AGS_NavMesh` with extras `{"AGS_type": "NAVMESH", "AGS_name": "AGS_NavMesh"}` (inspect with a GLTF viewer or `gltf-validator`).
- [ ] After exporting, the `AGS_NavMesh` object persists in the Blender scene for inspection.

### T-GS14 — GUI runtime (`InventoryBar`, `VerbBar`, `StatusLine`)

**Setup:** Build `ag build` for a room + .agui file. Add the generated CanvasLayer `.tscn` as an AutoLoad in the Godot project. Run the game.

- [ ] With a `.agui` that has a `StatusLine`, calling `AGSRuntime.set_status_text("Look at the door")` from any script displays "Look at the door" on the Label widget in the scene.
- [ ] `AGSRuntime.set_status_text("")` clears the status line.
- [ ] With a `.agui` that has a `VerbBar` with verbs ["Look", "Use", "Pick up"], three buttons appear at the defined anchor position. Clicking "Look" calls `AGSRuntime.set_active_verb("Look")`.
- [ ] After `set_active_verb("Look")`, the "Look" button shows as pressed/highlighted; clicking another verb unpresses the previous one.
- [ ] With a `.agui` that has an `InventoryBar`, picking up an item (via `AddInventory`) causes a new button with the item name to appear in the GridContainer.
- [ ] Losing an item (via `LoseInventory`) removes its button from the InventoryBar.
- [ ] The CanvasLayer persists across room transitions (since it is an AutoLoad).

### T-GS15 — Emitter: `SetStatusText`, `SetActiveVerb`, `GetActiveVerb`

**Setup:** Build `ag` CLI. Write a `.agscript` file using these calls.

- [ ] `SetStatusText("Look at the chest")` compiles to `AGSRuntime.set_status_text("Look at the chest")` with no `await`.
- [ ] `SetActiveVerb("Look")` compiles to `AGSRuntime.set_active_verb("Look")` with no `await`.
- [ ] `var v = GetActiveVerb()` compiles to `var v = AGSRuntime.get_active_verb()` with no `await`.
- [ ] Using these calls inside a longer script with blocking calls (Say, WalkTo) does not incorrectly mark them as blocking.

### T-BL01 — Blender add-on scaffold

**Setup:** Blender 4.2+ installed.

- [ ] In Blender: Preferences → Extensions → Install from Disk → select `tools/blender_addon/` folder (or a .zip of it). Add-on installs without errors.
- [ ] After installation, the add-on appears in the Extensions list as **AGS3D** version 0.1.0.
- [ ] Enable the add-on → no errors in the Blender console (`register()` called silently).
- [ ] Disable the add-on → no errors in the Blender console (`unregister()` called silently).
- [ ] Re-enable the add-on after a Blender restart — it stays enabled across restarts.

### T-GS13 — `.agui` parser + GUI scene generator

**Setup:** `ag` CLI built. Create a `main_hud.agui` file and run `ag build`.

- [ ] `GUI "main_hud" { layer = 10 }` → generates `main_hud.tscn` containing a `CanvasLayer` node named `MainHud` with `layer = 10` and `metadata/ags_gui_name = "main_hud"`.
- [ ] `InventoryBar "inv_bar" { position = (0, 0, bottom) item_size = (64, 64) columns = 6 }` → `main_hud.tscn` contains a `GridContainer` child with `anchors_preset = 12`, `columns = 6`, `metadata/ags_widget = "InventoryBar"`, `metadata/item_size = Vector2i(64, 64)`.
- [ ] `VerbBar "verbs" { position = (0, 0, bottom_right) verbs = ["Look", "Use", "Pick up"] }` → `HBoxContainer` child with three `Button` children; `text = "Look"`, `text = "Use"`, `text = "Pick up"`; each button has `metadata/ags_verb`.
- [ ] `StatusLine "status" { position = (0, 0, top) font = "assets/fonts/main.ttf" }` → `Label` child with `anchors_preset = 10` and font ext_resource.
- [ ] `StatusLine "status" { position = (0, 0, top) }` (no font) → `Label` child with no font ext_resource.
- [ ] Invalid syntax (e.g. `GUI "hud" { bad_prop = 1 }`) → `ag build` prints a parse error and exits non-zero.
- [ ] Open the generated `.tscn` in Godot. The `CanvasLayer` node and its children appear in the scene tree.

---

### T-GS29 — AGSAnimationPlayer2D

**Setup:** An `AGSCharacter2D` node with a `Sprite3D` child (hframes=18: 6 idle + 6 walk + 6 talk per row, vframes=4 for 4 directions), an `AGSBillboardController` child (`sprite_angles=4`, `frames_per_angle=6`), and an `AGSAnimationPlayer2D` child (`frames_per_state=6`, `fps=8`).

- [ ] `set_state("idle")` → Sprite3D cycles through frames 0–5 of the current direction row.
- [ ] `set_state("walk")` → Sprite3D cycles through frames 6–11 of the current direction row.
- [ ] `set_state("talk")` → Sprite3D cycles through frames 12–17 of the current direction row.
- [ ] Calling `set_state("walk")` twice in a row does not reset the frame counter (frame continues from where it was).
- [ ] `stop()` freezes the Sprite3D on the current frame (no further cycling).
- [ ] After `stop()`, calling `set_state("walk")` resumes cycling from frame 0 of walk state.
- [ ] Direction changes (from `AGSBillboardController`) correctly shift the row while keeping the state column offset intact.

---

### T-GS26 — Billboard direction selection runtime

**Setup:** A Godot project with an `AGSCharacter2D` node, a `Sprite3D` child with `vframes = 8` (8 directions), and an `AGSBillboardController` child (`sprite_angles = 8`, `frames_per_angle = 1`).

- [ ] Character moving away from camera → Sprite3D shows row 0 (N, back sprite).
- [ ] Character moving toward camera → row 4 (S, front sprite).
- [ ] Character moving right (from camera's perspective) → row 2 (E).
- [ ] Character moving left → row 6 (W).
- [ ] `sprite_locked = true` → row stays 0 regardless of movement direction.
- [ ] `sprite_angles = 4` with `frames_per_angle = 6`: moving character cycles through 6 frames in the correct row; row changes on direction change.
- [ ] `fps = 8.0` with `frames_per_angle = 6`: visible frame advance at ~8 fps during movement.
- [ ] Character standing still → row does not change (last direction held).

---

### T-BL04 — Room export operator

**Setup:** Blender 4.2+ with AGS3D add-on. A scene with visual mesh objects and AGS-tagged gameplay objects (Camera, WalkableSurface, BlockerVolume, etc.).

- [ ] `File → Export → AGS3D Room (.agroom + .glb)` appears in the Export menu.
- [ ] File selector opens pre-filled with the blend file stem as `.agroom` path.
- [ ] After export: `.agroom` file is created containing `Room "name" { ... }` with all AGS-tagged objects serialised.
- [ ] A `.glb` is created beside the `.agroom` containing only visual (non-gameplay-tagged) objects.
- [ ] WalkableSurface objects are NOT in the `.glb` (gameplay-only types excluded).
- [ ] Camera objects are NOT in the `.glb`.
- [ ] Visual mesh objects (untagged or `VisualMesh` type) ARE in the `.glb`.
- [ ] Camera position and computed look_at appear in the `.agroom` `Camera` block.
- [ ] SpawnPoint with `AGS_character = "player"` → `character = "player"` in SpawnPoint block.
- [ ] WalkableSurface size derived from object's bounding box × scale.
- [ ] `Export Gameplay Data` unchecked → only `.glb` written, no `.agroom`.
- [ ] `Export Visual Mesh` unchecked → only `.agroom` written, no `.glb`.
- [ ] Generated `.agroom` can be parsed by `ag validate` without errors.

---

### T-BL05 — Coordinate system conversion + bounding-box extraction

**Setup:** Blender 4.2+ with AGS3D add-on. Create a scene with:
- A cube at Blender location `(1, 2, 3)` tagged as `BLOCKER`
- A second cube scaled to 2×4×6 (Blender XYZ) tagged as `WALKABLE`
- A third cube rotated 45° on Z tagged as `HOTSPOT`

- [ ] Export → AGS3D Room. In the `.agroom`, the BlockerVolume `position` should be `(1.0, 3.0, -2.0)` (Godot: x=1, y=bl.z=3, z=-bl.y=-2).
- [ ] The BlockerVolume `size` should match the cube's world-space bbox in Godot axes (not just the raw Blender scale).
- [ ] WalkableSurface `size` tuple has 2 components (XZ plane); the values match Godot X and Z (no Y component).
- [ ] A cube at the origin with Blender scale (2, 4, 6) has Godot size `(2.0, 6.0, 4.0)` (x=bl.x=2, y=bl.z=6, z=bl.y=4).
- [ ] A rotated HOTSPOT: its `size` and `position` are computed from the **world-space** bounding box of all 8 corners — not from the local bbox — so rotation is properly accounted for (a 45°-rotated cube will have a larger world bbox than an aligned cube of the same size).
- [ ] A TRIGGER at `(0, 0, 5)` in Blender → `position = (0.0, 0.0, -5.0)` in `.agroom`.
- [ ] `position` for volume types (BlockerVolume, Hotspot, TriggerRegion) is the **bounding box centre**, not the object origin — move the object origin off-centre and confirm the exported position is the bbox midpoint.

---

### T-BL06 — Camera look_at: eyedropper-picked Empty

**Setup:** Blender 4.2+ with AGS3D add-on. Create a scene with a Camera object tagged as `CAMERA` and a separate Empty object.

- [ ] Select the Camera object; in the AGS3D panel (Object Properties or N-panel), the **Look-at** field appears for CAMERA type objects.
- [ ] Click the eyedropper / type the Empty's name in the Look-at field → `AGS_look_at` custom property is set to the Empty's name.
- [ ] Export → AGS3D Room. In the `.agroom`, the Camera `look_at` vector should match the Empty's world position (converted to Godot coords), NOT the auto-computed forward vector.
- [ ] Move the Empty to a different location, re-export → look_at updates to the new Empty position.
- [ ] Clear the Look-at field (empty string) → `look_at` line is **absent** from the Camera block; `ag build` falls back to auto-look-at (floor centre).
- [ ] Delete the referenced Empty from the scene, then export → `look_at` line is absent (no crash).
- [ ] A Camera with no `AGS_look_at` property → no `look_at` in the exported block; `ag build` auto-computes orientation from the floor centre.
- [ ] Camera position is always written; orientation in Blender (rotation, FOV) is the author's concern and is NOT re-derived by the exporter.

---

### T-BL07 — Export merge mode

**Setup:** Blender 4.2+ with AGS3D add-on. A Blender scene with two Camera objects, one WalkableSurface, and one Point tagged and named. An existing `.agroom` on disk that contains those same names **plus** an extra `SpawnPoint "old_start"` block not present as a Blender object.

- [ ] `File → Export → AGS3D Room`. The export dialog shows a **Merge** checkbox, enabled by default.
- [ ] Export with **Merge = on**: the written `.agroom` contains the blocks from Blender (with updated positions/sizes) **and** the `SpawnPoint "old_start"` block from the existing file (preserved verbatim).
- [ ] Export with **Merge = off**: the written `.agroom` contains only Blender-derived blocks; `SpawnPoint "old_start"` is gone.
- [ ] Add a new Camera to Blender and re-export with Merge = on. The new camera appears in the file; the existing cameras have their positions updated; the old_start spawn is still present.
- [ ] Delete a Camera from Blender and re-export with Merge = on. That camera block is absent from the file (Blender is authoritative for Blender objects); old_start spawn is still present.
- [ ] If the existing `.agroom` is malformed/unreadable, export falls back to full overwrite (no crash).
- [ ] If no existing `.agroom` exists (first export), Merge = on behaves identically to Merge = off.
- [ ] Generated `.agroom` passes `ag validate` without errors.

---

### T-BL10 — Character export operator

**Setup:** Blender 4.2+ with AGS3D add-on. A scene with an `Armature` object parented to a `Mesh` object. NLA editor has tracks named `Idle`, `Walk`, `Talk` with at least one strip each.

- [ ] `File → Export → AGS3D Character (.glb)` appears in the Export menu.
- [ ] File selector opens, pre-filled with `characters/<armature-name>/<name>.glb` relative to the blend file.
- [ ] Export completes; a valid `.glb` file is created at the chosen path.
- [ ] The `.glb` contains animation clips named `Idle`, `Walk`, `Talk` (matching NLA track names).
- [ ] Opening the `.glb` in Godot shows the armature + mesh + three animation clips in the AnimationPlayer.
- [ ] With no armature selected but a mesh selected: exports the mesh as a static `.glb` (no animation clips). No crash.
- [ ] With nothing selected: operator reports a warning "nothing to export" and returns cancelled.
- [ ] `Export Animations = false` toggle: exported `.glb` has no animation clips.
- [ ] The `.agchar` `mesh` + `animations` block can reference the exported file and clip names:
  ```
  Character "hero" {
      mesh = "characters/hero/hero.glb"
      animations = { idle = "Idle"  walk = "Walk"  talk = "Talk" }
  }
  ```

---

### T-BL08 — Import operator

**Setup:** Blender 4.2+ with AGS3D add-on. Have a sample `.agroom` file with Camera, WalkableSurface, BlockerVolume, Point, SpawnPoint, Hotspot, and TriggerRegion blocks.

- [ ] `File → Import → AGS3D Room (.agroom)` appears in the Import menu.
- [ ] Select the `.agroom` file → import completes without errors; info bar shows "AGS3D: imported N gameplay objects from 'room_name'".
- [ ] An **AGS_Gameplay** collection appears in the Outliner containing the imported objects.
- [ ] Cameras → `ARROWS` empties named `AGS_Cam_<name>`; custom property `AGS_type = "CAMERA"`, `AGS_name = "<name>"`.
- [ ] Points → `SINGLE_ARROW` empties; `AGS_type = "POINT"`.
- [ ] WalkableSurface → thin wire box with `AGS_type = "WALKABLE"`; scale matches `size` from `.agroom`.
- [ ] BlockerVolume → wire box with `AGS_type = "BLOCKER"`.
- [ ] SpawnPoint → `CIRCLE` empty; `AGS_type = "SPAWN"`, `AGS_character` set if specified.
- [ ] Hotspot / TriggerRegion → wire boxes with correct types.
- [ ] Running the import again on the same file clears the old objects and re-creates them (no duplicates).
- [ ] Imported positions are converted from Godot coords (Z negated): an object at `position = (1, 0, 2)` in `.agroom` lands at Blender location `(1, 0, -2)`.
- [ ] Non-.agroom files or malformed files report an error and do not crash Blender.

---

### T-BL03 — Viewport overlay

**Setup:** Blender 4.2+ with AGS3D add-on enabled. A scene with several mesh objects tagged as different AGS types.

- [ ] In the 3D viewport header, click **Overlays** → dropdown contains an **AGS3D** toggle. It is enabled by default.
- [ ] With overlay enabled: objects tagged as `WalkableSurface` show a **green** wireframe bounding box; `BlockerVolume` → **red**; `Hotspot` → **blue**; `TriggerRegion` → **purple**; `SpawnPoint` → **cyan**; `Camera` → **yellow**; `NavMesh` → **teal**; `Point` → **white**.
- [ ] Each tagged object shows a text label at its origin with format `"TypeName: object_name"` (or `AGS_name` if set).
- [ ] Objects with type `None` show no overlay.
- [ ] Disable the AGS3D overlay toggle → all colored wireframes and labels disappear immediately.
- [ ] Re-enable → overlays reappear.
- [ ] Disable the add-on → no errors in console; overlay is removed cleanly.

---

### T-BL02 — AGS3D object type panel

**Setup:** Blender 4.2+ with AGS3D add-on enabled (T-BL01). Open any scene.

- [ ] Select any mesh object → Object Properties → **AGS3D** section appears.
- [ ] The **Type** dropdown defaults to **None** on a new object.
- [ ] Select **WalkableSurface** → a **Name** text field appears below the dropdown.
- [ ] Enter `"floor"` in the Name field → the object gains a custom property `AGS_name = "floor"` (visible in Object Properties → Custom Properties).
- [ ] Select the object → Object Properties → Custom Properties shows `AGS_type = "WALKABLE"`.
- [ ] Switch to **SpawnPoint** type → a **Character** field appears in addition to Name.
- [ ] Enter `"player"` in Character → `AGS_character = "player"` custom property is set.
- [ ] Select **None** type → name and character fields disappear.
- [ ] Open View3D → Sidebar (N key) → **AGS3D** tab → same panel appears and stays in sync with Object Properties panel.
- [ ] Save the .blend file, reopen it → `AGS_type` and `AGS_name` custom properties are preserved on the object.
- [ ] Select a light object → the panel still renders without crashing (type stays None by default).

---

## M9 — AG Studio Editor (Batch 2)

### #94 T-E11 — AGS Inspector plugin

**Setup:** Launch AG Studio (`prototype.sh`). Open a room scene (double-click a `.agroom` in the project panel). Select an AGS node in the 3D viewport or scene tree. The Inspector should show **only** the AGS-specific form — no Godot category bars, no section groups (Transform, Collision, etc.).

- [ ] Select an `AGSWalkableSurface` node → Inspector shows label **"AGSWalkableSurface"**, then **Size (X / Z)** (two spinboxes: x, z), **Offset Y** (one spinbox), **Position** (three spinboxes: x, y, z). Nothing else.
- [ ] Select an `AGSBlockerVolume` node → Inspector shows **"AGSBlockerVolume"**, **Size** (three spinboxes: x, y, z), **Position** (x, y, z). Nothing else.
- [ ] Select an `AGSHotspot` node → Inspector shows **"AGSHotspot"**, **Hotspot name** (text field), **Size** (x, y, z), **Position** (x, y, z). Nothing else.
- [ ] Select an `AGSTriggerRegion` node → Inspector shows **"AGSTriggerRegion"**, **Region name** (text field), **Size** (x, y, z), **Position** (x, y, z). Nothing else.
- [ ] Select an `AGSPoint` node → Inspector shows **"AGSPoint"**, **Point name** (text field), **Position** (x, y, z). Nothing else.
- [ ] Select an `AGSSpawnPoint` node → Inspector shows **"AGSSpawnPoint"**, **Character** (dropdown populated from `.agchar` files in the project), **Position** (x, y, z). Nothing else.
- [ ] Select an `AGSCamera` node → Inspector shows **"AGSCamera"**, **Camera name** (text field), **Position** (x, y, z), a read-only hint "Look-at: edit via gizmo in 3D viewport". Nothing else.
- [ ] Select an `AGSRoom` node → Inspector shows **"AGSRoom"**, **Room name** (text field), **Initial camera** (dropdown of AGSCamera child names). Nothing else.
- [ ] Select an `AGSCharacter` node → Inspector shows **"AGSCharacter"**, **Character name** (text field), **Move speed** (spinbox), **Say text** (text field). Nothing else.
- [ ] Select a non-AGS node (e.g. `MeshInstance3D`) → Inspector shows the normal Godot inspector (not the AGS form).
- [ ] Edit a field (e.g. change Position X on an `AGSPoint`) → value persists when you deselect and reselect the node.
- [ ] No Godot section headers ("Transform", "Collision", "Axis Lock", "Visibility", etc.) are visible for any AGS node.

---

### #95 T-E15 — Build Log dock

**Setup:** Launch AG Studio. Ensure AG Studio menu → Build Log bottom panel is visible.

- [ ] AG Studio menu → Build → Build Log panel becomes visible and a build starts
- [ ] Build output lines appear in the RichTextLabel as the build runs
- [ ] A successful build shows a green "Build succeeded" summary line
- [ ] Introduce a syntax error in a `.agscript`, re-run build → error line appears in red
- [ ] Clicking an error link in the Build Log opens the relevant `.agscript` file (or shows the path) — error is navigable
- [ ] The Build Log clears between runs (no stale output from previous build)

---

### #96 T-E16 — Play button

**Setup:** Launch AG Studio. Locate the `▶ Play` button in the top toolbar.

- [ ] `▶ Play` button is visible in the top toolbar
- [ ] Clicking `▶ Play` opens the Build Log panel and starts a build
- [ ] The button is disabled (greyed out) while the build runs
- [ ] After a successful build, the game scene launches (Godot play mode starts)
- [ ] After a failed build, the game does NOT launch; button re-enables
- [ ] Button re-enables after build completes regardless of success/failure

---

### #97 T-E13 — Character editor main screen

**Setup:** Launch AG Studio. Double-click a `.agchar` file in the project panel.

- [ ] The character editor main screen opens (replaces any other custom panel)
- [ ] Character name, sprite sheet path, animation rows, move speed fields are populated from the `.agchar` file
- [ ] Editing a field and pressing Save writes the changes back to the `.agchar` file
- [ ] Re-opening the same `.agchar` shows the saved values
- [ ] Opening a second `.agchar` replaces the first in the editor (no stacking)
- [ ] Double-clicking a non-char file (e.g. `.agscript`) switches away from the char editor

---

### #98 T-E14 — Script editor main screen

**Setup:** Launch AG Studio. Double-click a `.agscript` file in the project panel.

- [ ] The script editor main screen opens with the file loaded in a CodeEdit tab
- [ ] AGS keywords (`room_Enter`, `room_Leave`, `WalkTo`, `FaceTo`, etc.) are syntax-highlighted in a distinct colour
- [ ] String literals (`"text"`) are highlighted
- [ ] Comments (`//`) are highlighted
- [ ] Blocking calls (`say`, `WalkTo`, etc.) show a clock icon in the gutter
- [ ] Opening a second `.agscript` opens a new tab rather than replacing the first
- [ ] Ctrl+S saves the current file; modified indicator (asterisk) clears after save
- [ ] Closing a tab removes it; switching tabs restores the correct file

---

### #99 T-E12 — `.agroom` ↔ `.tscn` sync

**Setup:** Launch AG Studio with `--godot-editor` flag (`prototype.sh --godot-editor`). Open a room `.tscn`.

- [ ] Select an `AGSHotspot` gizmo handle and drag it → box resizes in the 3D viewport
- [ ] Press Ctrl+S to save the scene → no error in Output
- [ ] Open the corresponding `.agroom` file in a text editor → the hotspot's `size` values match the dragged position
- [ ] Repeat with `AGSBlocker`, `AGSWalkableSurface`, `AGSPoint` — all write back correctly
- [ ] Saving a non-room `.tscn` does NOT trigger any `.agroom` write

---

### #100 T-E17 — Project wizard

**Setup:** Launch AG Studio. Open AG Studio menu → New Project…

- [ ] "New Project…" dialog opens with fields: project folder, project name, starting room name
- [ ] Clicking OK with valid inputs writes scaffold files:
  - `game.agp`
  - `rooms/<room_name>/<room_name>.agroom`
  - `rooms/<room_name>/<room_name>.agscript`
  - `characters/player.agchar`
- [ ] `ag build` runs automatically after scaffold is written
- [ ] Generated `.tscn` appears in the project panel after build
- [ ] Clicking OK with an empty project name shows a validation error and does not proceed
- [ ] Clicking Cancel dismisses without writing any files

---

### #101 T-E18 — Prototype migration (ag build regenerates .tscn)

**Setup:** From repo root, run `rm game_prototype/rooms/start/start.tscn` then `.dev/ag.sh build`.

- [ ] `ag build` completes without errors
- [ ] `game_prototype/rooms/start/start.tscn` is recreated
- [ ] The regenerated `.tscn` is byte-identical (or functionally equivalent) to the deleted hand-authored file
- [ ] Repeat for `library.tscn` if present
- [ ] Launch the prototype via `prototype.sh` — the game runs and both rooms load correctly

---

### #111 T-GS10 — `AGSRuntime.load_room()` + `room_change_requested`

**Setup:** Launch the prototype game (not editor). Add a trigger or script call that calls `AGSRuntime.load_room("library")` after a few seconds.

- [ ] `AGSRuntime.load_room("library")` emits `room_change_requested` (verify via print in `ags_room_manager.gd`)
- [ ] The room manager receives the signal and loads `rooms/library/library.tscn`
- [ ] The old room scene is removed from the tree before the new one is added (no double-room overlap)
- [ ] The player character (AutoLoad) persists across the room change — it is not freed
- [ ] Calling `load_room` with an unknown room name prints an error and does not crash
- [ ] `AGSRuntime.load_room("start")` switches back to the start room

---

### #103 T-GS01 — `AGSCharacter` `say_completed` + `say()` / `think()`

**Setup:** In a room script, call `player.say("Hello world")` inside `room_Enter`. Build and play the prototype.

- [ ] `say("Hello world")` sets `player.say_text = "Hello world"` (visible in remote inspector during play)
- [ ] After ~2 seconds `say_text` clears back to `""`
- [ ] `say_completed` signal fires after the duration (add a print in the room script to verify)
- [ ] `await player.say("Hello")` blocks the room script until say finishes, then the next line runs
- [ ] `player.think("Hmm…")` behaves identically to `say()` (same timing, same signal)
- [ ] Calling `say()` with a custom `duration` argument (e.g. `say("Long line", 4.0)`) waits the correct amount
- [ ] `say_text` property is visible in the Godot Inspector on an `AGSCharacter` node (requires Godot rebuild with new C++ changes)
- [ ] `say_completed` signal appears in the Node panel's Signals tab (requires Godot rebuild)

---

### T-GS04 — `ag build` `.agitem` parser + `ag validate` inventory checks

**Setup:** Create a small project with at least one `.agitem` file and a room script.

#### ag build
- [ ] Create `items/rusty_key.agitem` with `Item "rusty_key" { display_name = "Rusty Key" description = "An old key." sprite = "assets/key.png" }` — `ag build --force` reports `rusty_key.agitem (item — data only, no scene generated)` with no errors
- [ ] Create an `.agitem` with a syntax error (e.g. `Item "bad" {`) — `ag build` prints the parse error and exits with a non-zero code
- [ ] After a successful build, running `ag build` again (no changes) reports "nothing to do"

#### ag validate — check 6
- [ ] With `rusty_key.agitem` present, a script containing `AddInventory("rusty_key")` produces **no** issues
- [ ] With `rusty_key.agitem` present, a script containing `AddInventory("magic_wand")` produces an error: `"magic_wand"` is not defined
- [ ] `LoseInventory("unknown")` in a script produces the same kind of error
- [ ] `HasInventory("unknown")` inside an `if` condition produces the same kind of error
- [ ] A global script (not paired with any `.agroom`) with `AddInventory("no_such")` still produces the error (not silently skipped)
- [ ] The error message includes the line number of the offending call

---

## M10 — Game Systems (Batch 1)

### T-GS05 — Go: grammar + emitter — `Say`, `Think`, `AddInventory`, `LoseInventory`, `HasInventory`

**Setup:** Write an `.agscript` file and run `ag build` or manually inspect emitted `.gd` in `.engine/generated/`.

- [ ] `global.player.Say("Hello!")` in a room script emits `await AGSRuntime.get_character("player").say("Hello!")` — note `await` present
- [ ] `global.player.Think("Hmm...")` emits `await AGSRuntime.get_character("player").think("Hmm...")` — note `await` present
- [ ] `global.player.AddInventory("rusty_key")` emits `AGSRuntime.get_character("player").add_inventory("rusty_key")` — no `await`
- [ ] `global.player.LoseInventory("rusty_key")` emits `AGSRuntime.get_character("player").lose_inventory("rusty_key")` — no `await`
- [ ] `if (global.player.HasInventory("rusty_key"))` emits `AGSRuntime.get_character("player").has_inventory("rusty_key")` inside an `if` — no `await`
- [ ] `cGuard.Say("Halt!")` (identifier receiver) emits `await AGSRuntime.get_character("c_guard").say("Halt!")` (snake_case applied to receiver name)
- [ ] A function that calls `Say` is itself treated as blocking — a call to it from another function emits `await that_func()`

### T-GS06 — Go: grammar + emitter — `HideRoomItem`, `ShowRoomItem`, `item_interact` handler

**Setup:** Write an `.agscript` in a room directory and run `ag build`; inspect the emitted `.gd`.

- [ ] `HideRoomItem("old_chest")` in a room script emits `AGSRuntime.hide_room_item("old_chest")` — no `await`
- [ ] `ShowRoomItem("old_chest")` emits `AGSRuntime.show_room_item("old_chest")` — no `await`
- [ ] A function named `item_interact` with a `string name` parameter emits `func item_interact(name: String):`
- [ ] Inside `item_interact`, `HideRoomItem("old_chest")` emits the runtime call correctly
- [ ] Inside `item_interact`, `global.player.AddInventory("rusty_key")` emits the character runtime call correctly
- [ ] In-game: place an `AGSRoomItem` in a room; clicking it calls `item_interact` on the room script with the item name

---

### T-GS18 — Cutscene support: `SetPlayerControl`, `FadeIn`, `FadeOut`, `Wait`

**Setup:** Add an `AGSCutscene` AutoLoad pointing to `.engine/runtime/ags_cutscene.gd`. Write a room script that triggers a cutscene sequence.

#### Player control
- [ ] `AGSRuntime.set_player_control(false)` in the console (Remote inspector during play) → clicking the floor does not move the player
- [ ] `AGSRuntime.set_player_control(true)` → clicking the floor moves the player again
- [ ] `AGSRuntime.player_control_changed` signal fires on each toggle (verify with a print connection)

#### Fade
- [ ] `await AGSCutscene.fade_out()` — screen fades to black over 0.5 s, execution resumes after
- [ ] `await AGSCutscene.fade_in()` — screen fades from black to clear over 0.5 s
- [ ] `fade_out(2.0)` / `fade_in(2.0)` — custom duration is respected
- [ ] The fade overlay (CanvasLayer layer=100) renders on top of all 3D content

#### Wait
- [ ] `await AGSCutscene.wait(1.0)` — pauses execution for 1 second, then resumes
- [ ] `await AGSCutscene.wait(0.0)` — returns on the next frame (no hang)

#### Hide / show room items
- [ ] `AGSRuntime.hide_room_item("chest")` sets the `AGSRoomItem` named "chest" invisible
- [ ] `AGSRuntime.show_room_item("chest")` makes it visible again
- [ ] Calling `hide_room_item` for a name not in the scene prints a warning and does not crash

---

### T-GS11 — Go: grammar + emitter — `PlayMusic`, `StopMusic`, `PlaySound`

**Setup:** Write an `.agscript` file and run `ag build`; inspect the emitted `.gd`.

- [ ] `PlayMusic("theme_main")` in a room script emits `AGSRuntime.play_music("theme_main")` — no `await`
- [ ] `StopMusic()` emits `AGSRuntime.stop_music()` — no `await`
- [ ] `PlaySound("door_creak")` emits `AGSRuntime.play_sound("door_creak")` — no `await`
- [ ] A function that only calls `PlayMusic` is **not** treated as blocking — calling it from another function does **not** emit `await`

---

### T-GS12 — GDScript: `AGSRuntime` audio manager

**Setup:** Add `AGSAudio` AutoLoad pointing to `.engine/runtime/ags_audio.gd`. Place audio files in `audio/music/` and `audio/sfx/`. Write a room script that calls `PlayMusic` and `PlaySound`.

- [ ] `AGSRuntime.play_music("theme_main")` — `audio/music/theme_main.ogg` begins playing (looping)
- [ ] `AGSRuntime.play_music("track_b")` while music is playing — previous track stops, new track starts
- [ ] `AGSRuntime.stop_music()` — music stops immediately
- [ ] `AGSRuntime.play_sound("door_creak")` — `audio/sfx/door_creak.ogg` plays once and stops on its own
- [ ] Calling `play_sound` 8 times rapidly (>pool size limit) does not crash — oldest playing slot is reused
- [ ] Calling `play_music` with a name that has no file in `audio/music/` prints a warning and does not crash
- [ ] Calling `play_sound` with a name that has no file in `audio/sfx/` prints a warning and does not crash
- [ ] `play_music_requested`, `stop_music_requested`, `play_sound_requested` signals appear in the Node panel's Signals tab on `AGSRuntime` (requires Godot rebuild)

---

### T-GS16 — GDScript: `AGSRuntime` save_game / load_game

**Setup:** Add a room script that saves and loads. Use Godot remote inspector to verify state.

- [ ] `AGSRuntime.game_saved(1)` returns `false` before any save has been made to slot 1
- [ ] `AGSRuntime.save_game(1)` creates `user://save_1.json`; `game_saved(1)` returns `true`
- [ ] `AGSRuntime.load_game(1)` restores global variables to their saved values
- [ ] After `load_game`, character inventory matches what was saved (inspect via remote debugger)
- [ ] After `load_game`, hidden room items remain hidden; visible items remain visible
- [ ] `load_game_requested` signal fires with a dictionary containing `"globals"`, `"characters"`, `"room_items"`, `"room"`, `"music"` keys
- [ ] Saving to slot 1 and slot 2 independently — loading slot 1 does not affect slot 2 state
- [ ] `load_game` with a non-existent slot prints a warning and does not crash
- [ ] `save_game` / `load_game` appear as signals in the Node panel (requires Godot rebuild)

---

### T-DLG01 — Go: `.agdlg` lexer

All tests are automated (`go test ./internal/dlg/...`). No manual test required.

---

### T-DLG13 — Go: `game.agp` `[locales]` + `[localisation]` blocks

All tests are automated (`go test ./internal/project/...`). Manual verification:

- [ ] Add a `[locale.en]` / `[locale.fr]` / `[locale.ar]` block to a real `game.agp`; run `ag build` — build does not crash and no error is printed
- [ ] Add `[localisation]\nbase_locale = "zz"` (undeclared locale) to `game.agp`; run `ag validate` — error message references `base_locale "zz"` not declared
- [ ] Add `[locale.INVALID]` to `game.agp`; run `ag validate` — error message references invalid locale code

---

### T-DLG02 — Go: `.agdlg` parser

All tests are automated (`go test ./internal/dlg/...`). No manual test required.

---

### T-DLG03 — Go: `.agdlg` link stage

All tests are automated (`go test ./internal/dlg/...`). No manual test required.

---

### T-DLG04 — Go: structural dialogue validator

All tests are automated (`go test ./internal/dlg/...`). No manual test required.

---

### T-DLG07 — Go: dialogue emit stage + `ag build` integration

Unit tests are automated (`go test ./internal/dlg/...`). Manual verification:

- [ ] Create `dialogue/guard.agdlg` with two nodes (see spec example); run `ag build` — `.engine/generated/dialogue/guard.json` is created
- [ ] Open the JSON; verify `nodes[0].title`, `character`, `body[*].type` match source
- [ ] Verify auto loc keys are present on speaker lines and options (format `"nodeTitle:0:xxxxxxxx"`)
- [ ] Run `ag build` again without changing the file — output prints "nothing to do" (mtime cache hit)
- [ ] Run `ag build --force` — file is re-emitted
- [ ] Add `<<jump missing_target>>` to a node; run `ag build` — build fails with DLG-E002 error message
- [ ] Add duplicate `title:` in two files; run `ag build` — build fails with DLG-E001 error message

---

### T-CUT01 — Go: `.agcut` file parser

All tests are automated (`go test ./internal/cut/...`). No manual test required.

---

### T-CUT02 — Go: full command vocabulary parser

All tests are automated (`go test ./internal/cut/...`). No manual test required.
