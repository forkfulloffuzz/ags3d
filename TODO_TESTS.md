# AGS3D — Manual Tests

One test section per TODO task. Update this file immediately after marking a task complete in `TODO.md`.

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
