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
