# ag — AGS3D CLI usage

## Commands

### `ag build [--force]`
Parse changed `.agscript`, `.agroom`, `.agchar`, `.agitem`, and `.agdlg` source
files and emit GDScript into the project's `.engine/generated/` directory.
`--force` reprocesses all files regardless of modification time.

### `ag run`
Runs `ag build` then launches the Godot editor.

### `ag validate`
Runs static analysis across the project. Checks include:

- `game.agp`: `start_room` and `start_character` file references exist
- `.agroom`: `initial_camera` names a Camera block defined in the same room
- `.agroom`: each `SpawnPoint.character` matches a known `.agchar`
- `.agscript`: `WalkTo`/`FaceTo` point-name args exist in the paired room
- `.agscript`: `AddInventory`/`LoseInventory`/`HasInventory` item names resolve
- `.agdlg`: structural errors DLG-E001..E025 and warnings DLG-W001..W012

Must be run from inside a project directory (one containing `game.agp`).

#### Pipe mode

`ag validate` also accepts a newline-separated list of file paths on stdin.
This lets you target specific files or use `find` for recursive discovery
without needing `game.agp` in scope:

```sh
# Validate only dialogue files (all analysed together — cross-file checks work):
find . -name "*.agdlg" | ag validate

# Validate everything under a subdirectory:
find rooms/library -name "*.ag*" | ag validate

# Validate the whole project tree from any working directory:
find /path/to/project -name "*.ag*" | ag validate
```

Blank lines and lines starting with `#` in stdin are ignored.
In pipe mode the `game.agp` manifest checks (start_room, start_character) are
skipped because no manifest is loaded.

### `ag export --platform NAME`
Runs `ag build` then invokes the Godot headless export pipeline.
`NAME` is one of: `windows`, `mac`, `linux`, `web`, `ios`, `android`.

### `ag new NAME`
Scaffolds a new AGS3D project named `NAME` in the current directory.

### `ag export --locale LANG [--format FORMAT] [--diff] OUTFILE`
Exports localisation strings from compiled dialogue JSON.

- `--locale LANG` — language code (e.g. `en`, `fr`, `de`)
- `--format FORMAT` — `po` (default), `csv`, or `agstrings`
- `--diff` — emit only untranslated and stale strings

### `ag viz` — visualisation stages

```sh
ag viz tokens <file>   # print token stream (VIZ-01)
ag viz ast <file>      # print AST tree (VIZ-02)
ag viz blocking <file> # print blocking-call annotations (VIZ-03)
ag viz emit <file>     # side-by-side AGS-spirit ↔ GDScript (VIZ-04)
ag viz <file>          # run all four stages
```
