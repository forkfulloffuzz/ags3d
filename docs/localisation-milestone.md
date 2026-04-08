# AGS3D Localisation — Milestone M-LOC

> **Status: In progress.**
> Depends on: M-DLG (T-DLG08 done), M-CUT (T-CUT09 done)

## Goal

A complete localisation pipeline for AGS3D. Authors export translatable strings
from dialogue and cutscene files into the AGS3D native `.agstrings` format.
Translators work in `.agstrings` files (or external PO/CSV via a bridge).
At runtime, the GDScript localisation engine loads the correct locale file and
serves translated strings to the dialogue engine and UI.

At milestone end an author can:

1. Run `ag export --locale fr` to produce `locale/fr.agstrings` from all
   `.agdlg` and `.agcut` source files.
2. Hand the file to a translator, who fills in translations in-place.
3. Run `ag validate` to catch missing, stale, or orphaned strings.
4. Run the game with `Game.SetLocale("fr")` and see all dialogue and UI in French.

---

## File Format — `.agstrings`

`.agstrings` is a plain-text, human-readable key/value format. It is designed
to be diff-friendly and easy to edit in a text editor.

### Structure

```
// locale/fr.agstrings

[meta]
base_locale    = en
locale         = fr
locale_name    = French
rtl            = false
fallback_chain = en

[strings]
guard_greeting:line0:a1b2c3d4 = "Vous. Arrêtez."
guard_greeting:line1:e5f6a7b8 = "Montrez-moi votre laissez-passer."
key_examine:line0:c9d0e1f2   = "Une vieille clé en fer."

// untranslated — exported from source, awaiting translation
// [stale] guard_farewell:line0:99aabbcc = "Au revoir."
```

### Rules

- `[meta]` block is required and must appear first.
- `[strings]` block contains the translatable key/value pairs.
- Keys are stable loc keys produced by `ag export` — format `{node}:{index}:{hash8}`.
- Values are quoted strings. Escape sequences: `\"`, `\\`, `\n`, `\t`.
- Lines beginning with `//` are comments. A `// [stale]` prefix marks a key
  whose source text has changed since last export (hash mismatch).
- Blank lines are allowed anywhere.
- Keys must be unique within a file.

### Meta fields

| Field | Required | Description |
|-------|----------|-------------|
| `base_locale` | yes | The source language (usually `en`) |
| `locale` | yes | BCP 47 code for this file's target language |
| `locale_name` | no | Human-readable locale name |
| `rtl` | no | `true` if the locale is right-to-left (default `false`) |
| `fallback_chain` | no | Space-separated ordered list of locales to try on missing key |

### Diff semantics

`ag export --locale fr` writes a new `.agstrings` file or updates an existing one:

- **New key**: added at the end of `[strings]` with an empty value `""`.
- **Existing key, source text unchanged**: value preserved as-is.
- **Existing key, source text changed**: existing line commented out with
  `// [stale]` prefix; new key appended with empty value.
- **Key no longer in source**: existing line commented out with `// [orphan]` prefix.
- `--diff` flag: only outputs keys with empty values (untranslated) — useful for
  translator handoff.

---

## Architecture

```
.agdlg / .agcut source files
  │
  ▼  ag export --locale <code>
CollectLocEntries (dlg) + CollectVoiceLines (cut)
  │
  ▼
loc.Diff(existing, updated)  ← reads locale/<code>.agstrings
  │
  ▼
loc.Write(StringsFile)       → locale/<code>.agstrings

locale/<code>.agstrings
  │
  ▼  ag build
loc.Parse()  →  compiled string table
  │
  ▼  runtime
GDScript LocalisationRuntime.load_locale("fr")
  └─  serves translated strings to dialogue engine + UI
```

---

## Task Breakdown

### Phase 1 — Format & Pipeline

| Task | Description | Depends on |
|---|---|---|
| T-LOC01 | Design: `.agstrings` format spec (this document) | — |
| T-LOC02 | Go: `.agstrings` parser, writer, and diff engine in `tools/ag/internal/loc/` package | T-LOC01 |
| T-LOC03 | Go: `ag export --locale` integration with `.agstrings` — replaces/extends the existing PO/CSV export to also write `.agstrings`; reads existing file for diff | T-LOC02, T-DLG08 (done) |
| T-LOC04 | Go: `ag validate` localisation pass — DLG-LOC-E001 (missing translation), DLG-LOC-W001 (stale), DLG-LOC-W002 (orphan) | T-LOC02 |

### Phase 2 — Advanced Export

| Task | Description | Depends on |
|---|---|---|
| T-LOC05 | Standalone: `ag-loc` translation tool — interactive CLI for filling in `.agstrings` values; shows source text and context; marks complete on save | T-LOC02 |
| T-LOC06 | Go: PO/CSV import bridge — read a translated PO or CSV and write translations back into the corresponding `.agstrings` file | T-LOC02 |
| T-LOC07 | GDScript: runtime string loading from `.agstrings` + locale switching — `Game.SetLocale(code)` loads the compiled string table, propagates to all consumers, supports fallback chain | T-LOC02 |
| T-LOC08 | AG Editor: localisation panel integrated into AG Studio (M12 deferred) | T-LOC07 |
| T-LOC09 | Go: voice script export grouped by `voice_session` + character — extends `ag export --voicescript` to pull from both `.agdlg` and `.agcut` sources | T-DLG09 (done) |

### Phase 3 — String Taxonomy

| Task | Description | Depends on |
|---|---|---|
| T-LOC10 | Design: author context annotations in `.agdlg` and `.agcut` — `#ctx:` comment syntax to attach translator notes to individual strings | T-LOC02 |
| T-LOC11 | Design: string taxonomy and source metadata in `.agstrings` — `type:` (spoken/choice/narration/ui/subtitle), `char:`, `scene:` metadata fields per entry | T-LOC10 |
| T-LOC12 | Go: `ag loc` subcommand — advanced search, filter, sort, and group-by over `.agstrings` files | T-LOC11 |
| T-LOC13 | Go: `ag loc report` — condition-based report generation (e.g. all untranslated lines for character X) | T-LOC12 |

### Phase 4 — Editor & Coverage

| Task | Description | Depends on |
|---|---|---|
| T-LOC14 | AG Editor: advanced string browser with grouping, search, and filter (M12) | T-LOC11 |
| T-LOC15 | AG Editor: localisation reports panel and condition-based export (M12) | T-LOC13 |
| T-LOC16 | Pipeline + Editor: voice file connection and recording coverage tracking (M12) | T-LOC09 |

---

## Validator Error & Warning Reference

| Code | Type | Check |
|---|---|---|
| DLG-LOC-E001 | Error | Locale file present but key has no translation and `ag build` is in release mode |
| DLG-LOC-W001 | Warning | Key is stale (source text changed since last export) |
| DLG-LOC-W002 | Warning | Key is orphaned (no longer present in source) |
| DLG-LOC-W003 | Warning | Locale file missing for a locale declared in `game.agp [locales]` |

---

## T-LOC10 — Author Context Annotations (`#ctx:`)

### Problem

Translators often need more than the raw string to produce an accurate translation.
A line like `"That's not a problem."` is ambiguous — it could be dismissive,
reassuring, or sarcastic. Without scene context, a translator may pick the wrong
tone, resulting in an awkward or incorrect localisation.

### Solution

A `#ctx:` inline comment suffix that authors attach to any localizable string.
The annotation is ignored by the game engine but is harvested during `ag export`
and written into `.agstrings` files (or PO/CSV) as a translator note.

### Syntax

`#ctx:` is a trailing inline comment on the same line as the string it describes.
It is separated from the string by whitespace, and from any `#loc:` annotation by
whitespace. The `#ctx:` value extends to the end of the line.

```
// .agdlg
title: guard_greeting
---
Guard: Halt!  #ctx: guard is suspicious, not yet hostile
<<end>>
===

// .agcut
<<line guard "Stop right there!" #ctx:guard is threatening - firm, authoritative tone>>
```

**Rules:**
- `#ctx:` must appear on the same line as the string it annotates.
- Multiple `#ctx:` on the same line are concatenated (first wins for display).
- A `#ctx:` value containing spaces should be quoted (`"..."`). Unquoted
  `#ctx:` values are terminated at the first whitespace.
- `#ctx:` is stripped from the actual translatable string.
- `#ctx:` is optional on every string. Strings without it have no translator note.

### Supported File Types

| File type | Placement |
|-----------|-----------|
| `.agdlg` | On the same line as the speaker line, narration, or choice text |
| `.agcut` | On the same line as `<<line>>`, `<<title_card>>`, `<<subtitle>>`, `<<choice>>` commands |

### `.agstrings` Export Format

When `ag export` runs, each `#ctx:` value is written into the PO file as a
`#.` comment (translator note) above the `msgctxt`/`msgid`/`msgstr` block, and
into CSV as a separate `context` column:

```po
#. Context: guard is suspicious, not yet hostile
#. Type: spoken
#. Node: guard_greeting
msgctxt "guard_greeting:line0:aabb1122"
msgid "Halt!"
msgstr ""

```

```csv
loc_key,node,character,type,source_text,translation,context
guard_greeting:line0:aabb1122,guard_greeting,Guard,spoken,"Halt!","",guard is suspicious not yet hostile
```

### Parser Requirements

**`.agdlg` parser changes:**
- `SpeakerLine`, `NarrationLine`, `OptionBranch` structs gain a `Ctx string` field.
- During lexing/parsing, trailing `#ctx:` content on the text line is extracted,
  the `#ctx:` portion is stripped from the text, and the remainder is stored in `Ctx`.
- If both `#loc:` and `#ctx:` appear on the same line, both are extracted:
  `#loc:mykey #ctx:some context` → `LocKey = "mykey"`, `Ctx = "some context"`.

**`.agcut` parser changes:**
- `RawCommand.Args` (the raw string of arguments) is post-processed to extract
  `#ctx:` values. The `collectLocArgs()` helper in `cut/loc.go` is extended to
  also return a `context` string.
- The `LocEntry` struct in `cut/loc.go` gains a `Ctx string` field.

### `LocEntry` Struct Changes

```go
// dlg/export.go
type LocEntry struct {
    LocKey    string
    NodeTitle string
    Character string
    LineType  string // "spoken" | "choice" | "narration"
    Source    string
    Ctx       string // T-LOC10: author context / translator note
}

// cut/loc.go
type LocEntry struct {
    LocKey  string
    Source  string
    CmdName string
    Pos     Pos
    Ctx     string // T-LOC10: author context / translator note
}
```

### PO Export Changes

`dlg.ExportPO()` writes `# ctx:` comments (already uses `#.` prefix for other
metadata). `cut.WriteAgstringsTemplate()` stores context in a `// ctx:` comment
prefix on each key line.

### Implementation Plan

1. **Parser** (`dlg/` + `cut/`): Add `Ctx` field to `LocEntry`. Extract `#ctx:`
   in `collectStmtEntries()` (dlg) and `CollectLocEntries()` (cut).
2. **Export** (`dlg/export.go`): Include `Ctx` in PO comment output and CSV
   `context` column.
3. **Export** (`cut/loc.go`): Include `Ctx` in `WriteAgstringsTemplate()` output.
4. **Validate** (optional future): A `DLG-LOC-W004` warning when a string
   exceeds N characters without a `#ctx:` (configurable threshold, off by default).

---

## T-LOC11 — String Taxonomy and Source Metadata

### Goal

Extend `.agstrings` entries with structured metadata fields (`type:`, `char:`, `scene:`)
so translators and tooling can filter, group, and search strings without needing to
parse the source files. This builds on the `#ctx:` context annotations from T-LOC10.

### Proposed `.agstrings` Format

Metadata is written as `// key: value` comments directly above the key line.
All metadata is optional and absent fields are omitted (not written as empty comments).

```
// locale/fr.agstrings

[meta]
base_locale    = en
locale         = fr
locale_name    = French
rtl            = false
fallback_chain = en

[strings]

// type: spoken
// char: Guard
// scene: guard_greeting
// ctx: guard is suspicious, not yet hostile
guard_greeting:line0:a1b2c3d4 = "Vous. Arrêtez."

// type: choice
// char:
// scene: guard_greeting
guard_greeting:line1:e5f6a7b8 = "Show me your pass."

// type: ui
// char:
// scene: hud
hud_health_label:line0:b1c2d3e4 = "Health:"
```

### Metadata Fields

| Field | Source | Description |
|-------|--------|-------------|
| `type` | `LocEntry.LineType` | String category: `spoken`, `choice`, `narration`, `ui`, `subtitle`, `title_card` |
| `char` | `LocEntry.Character` | Character name who speaks this string (empty for narration/UI) |
| `scene` | `LocEntry.NodeTitle` | Dialogue node or cutscene title this string originates from |
| `ctx` | `LocEntry.Ctx` | Author context note from `#ctx:` annotation (T-LOC10) |

### Design Decisions

**Metadata as comments, not structure.** The key-value format `key = "value"` is
preserved. Metadata is comment lines directly above the key line. This keeps the
format diff-friendly (changing a translation only changes one line) and avoids
requiring a format migration for existing `.agstrings` files.

**Backward compatibility.** Entries without metadata (from older exports) are
still valid. `ag export` re-generates metadata on export, so metadata is
"best-effort" — it may be missing on entries that haven't been re-exported.

**`scene:` vs `nodeTitle`.** The field is named `scene` to be neutral across
dialogue (node title) and cutscene (cutscene title) sources.

**`type` values are an enumeration.** Unknown type values are silently ignored
by the parser and treated as absent. Types:

| Type | Source | Notes |
|------|--------|-------|
| `spoken` | `SpeakerLine` | Character dialogue |
| `choice` | `OptionBranch` | Player dialogue choice |
| `narration` | `NarrationLine` | Author narration / description |
| `ui` | GUI strings, HUD labels | From `.aggui` and built-in UI |
| `subtitle` | `<<subtitle>>` in `.agcut` | |
| `title_card` | `<<title_card>>` in `.agcut` | |

### Parser Requirements

**`LocEntry` struct changes (dlg + cut):**

```go
type LocEntry struct {
    LocKey    string
    NodeTitle string // used as "scene" in export
    Character string
    LineType  string
    Source    string
    Ctx       string // T-LOC10

    // T-LOC11: formal taxonomy fields (populated during collection)
    Scene string // always = NodeTitle (alias for clarity in .agstrings)
    Type  string // always = LineType (alias for clarity in .agstrings)
}
```

**`agstrings` format changes:**
- `loc.Parse()` is unchanged — it only reads `key = value` pairs.
- `loc.Write()` (used by `ag export`) writes metadata comments above each entry.
- Existing `.agstrings` files without metadata are still valid input.
- During `ag export --locale fr`, existing translation values are preserved;
  metadata comments are regenerated from source.

**CSV export changes:**
- Header gains additional columns: `scene,context`.
- Example: `loc_key,node,character,type,source_text,translation,scene,context`

**PO export changes:**
- Already emits `#. Type:` and `#. Node:` as comments — these are the `type`
  and `scene` fields. The `#. Character:` comment already exists for `char`.

### `agstrings` Writer Changes

`loc.Write()` is extended to accept a slice of enriched `LocEntry` values:

```go
// WriteStringsFile writes a .agstrings file with optional per-entry metadata.
// Each entry may have Ctx, Character, LineType, NodeTitle set.
func WriteStringsFile(entries []LocEntry, w io.Writer) error
```

When any metadata field is non-empty on an entry, the writer emits comment lines
before the key line. When all metadata fields are empty (legacy entry), it emits
the bare `key = "value"` line for backward compatibility.

### Diff Semantics (unchanged)

`ag export --locale fr` still:
- Adds new keys with empty value.
- Preserves existing translations.
- Marks stale keys with `// [stale]`.
- Marks orphaned keys with `// [orphan]`.
- **Now also regenerates metadata comments** above each entry (non-diff-affecting
  comment-only changes; translators' workflow is unchanged).

### Future Work (out of scope for T-LOC11)

- Per-entry metadata in the Go `LocEntry` struct (T-LOC11 struct changes above)
  are implemented as part of T-LOC11. Runtime GDScript loading of metadata
  fields is a separate task (T-LOC14 in M12).
- Filtering/grouping by metadata fields: T-LOC12 (`ag loc` subcommand).
- Condition-based reports: T-LOC13 (`ag loc report`).

---

## Out of Scope

- Machine translation integration
- Translation memory / fuzzy matching (future tooling idea)
- In-game string editing at runtime
- Pluralisation rules (CLDR) — first milestone uses simple key replacement only
