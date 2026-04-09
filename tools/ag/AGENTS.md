# tools/ag — Go CLI Agent Instructions

The `ag` CLI is the AGS3D build and validation pipeline. It parses adventure-game
source files (`.agroom`, `.agchar`, `.agcut`, `.agdlg`, `.agscript`) and produces
Godot `.tscn` scenes and GDScript.

## Module

`github.com/ags3d/ag` (see `go.mod`)

## Commands

| Command | Purpose |
|---------|---------|
| `ag build <project>` | Parse source files, emit `.tscn` + GDScript into `.engine/generated/` |
| `ag validate <project>` | Cross-system validation — characters, rooms, dialogue, cutscenes, localisation |
| `ag viz <stage> <file>` | Visualise a pipeline stage (tokens, ast, sym, emit, …) |
| `ag ls <project>` | List all discovered source files |

## Package layout

```
cmd/ag/           CLI entry point (main.go)
api/              Public facade — external tools use this, not internal/
internal/
  aganim/         .aganim frame-tag sidecar (parse, GDScript literal)
  analysis/       Semantic analysis passes
  char/           .agchar parser
  cut/            .agcut cutscene parser, localisation pass, sequence validator
  dlg/            .agdlg dialogue parser, cross-validator, warnings, emitter
  emitter/        AGS-spirit → GDScript emitter
  gui/            .aggui GUI descriptor parser
  item/           .agitem parser
  loc/            .agstrings / .agloc locale file parser
  lsp/            Language Server Protocol (experimental)
  parser/         AGS-spirit language parser
  project/        Manifest, file scanning, change detection, mtime cache
  room/           .agroom parser
  scene/          .tscn scene generator (room, character, GUI)
  validate/       ValidateFiles() orchestrator + per-type validators
  viz/            AST / symbol / emit visualisers
testdata/         Fixture project directories for integration tests
```

## Running tests

```sh
cd tools/ag && go test ./...
# or from project root:
.dev/test-ag.sh
```

## Testdata fixtures

**Every new parser, validator, or pipeline feature must add fixture files to `testdata/`.**

### When to add fixtures

| When you add... | Add these fixtures... |
|---|---|
| A new `.agroom` block type | `testdata/rooms/valid/NN_name.agroom` + `testdata/rooms/invalid/err_NN_desc.agroom` |
| A new `.agdlg` feature | `testdata/dialogues/valid/NN_name.agdlg` + `testdata/dialogues/invalid/err_NN_desc.agdlg` |
| A new `.agcut` command or header field | `testdata/cutscenes/valid/NN_name.agcut` |
| A new `.agitem` field | `testdata/items/valid/NN_name.agitem` |
| A new `.agscript` expression type | `testdata/scripts/valid/NN_name.agscript` |
| A new `.agstrings` feature | `testdata/locale/valid/NN_name.agstrings` + `testdata/locale/invalid/err_NN_desc.agstrings` |

### Naming convention

```
testdata/<type>/valid/NN_descriptive_name.ext
testdata/<type>/invalid/err_NN_error_description.ext
```

- `NN` = two-digit sequential number starting at `01`
- `name` = short lowercase identifier for the feature
- `desc` = what makes the file invalid
- Invalid files must start with `// EXPECT_ERROR: <description>` comment

### Rules

- **Never delete** existing fixtures — other tests may depend on them.
- **Never renumber** existing fixtures — adds noise to git history.
- Add new fixtures at the end of the sequence.
- For a new file type, create the `testdata/<type>/valid/` and `testdata/<type>/invalid/` directories.
- After adding fixtures, update `testdata/README.md` with a table entry.
- File-based fixture tests use `filepath.Glob` and live in the package's `_test.go`.

### Example: adding a locale fixture

```go
// testdata/locale/valid/06_new_feature.agstrings
[meta]
base_locale = en
locale = fr

[strings]
new_key = "value"
```

```go
// internal/loc/loc_test.go — already wired via TestParse_ValidFixtures
```

### game_prototype as integration target

`game_prototype/` is the integration test project. When adding a new feature that affects dialogue, cutscenes, rooms, or characters, add realistic fixture files there too so the feature can be exercised end-to-end via `ag validate game_prototype`.

## Adding a validator

1. Add the validator function to the relevant `internal/<pkg>/` package.
2. Call it from `internal/validate/validate.go` inside `ValidateFiles()`.
3. Return `[]Issue` — each issue has `File`, `Line`, `Severity` (`"error"`/`"warning"`), `Code`, `Message`.
4. Wire errors/warnings into `cmd/ag/main.go` print loop if a new exit code is needed.

## Adding a new `ag` command or viz stage

1. Add the command handler in `cmd/ag/main.go`.
2. Expose it via `api/api.go` if external tools need it.
3. **Update `.dev/ag.sh`** with the new command/stage — this is mandatory.

## Validate pipeline entry point

`validate.ValidateFiles(root string, files []project.SourceFile) ([]Issue, error)`

Orchestrates all subsystem checks. Wire new validators here rather than in `main.go`.
