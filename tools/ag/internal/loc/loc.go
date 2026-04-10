// Package loc provides types and functions for the AGS3D native localisation
// format (.agstrings).
//
// A .agstrings file has a [meta] block and a [strings] block:
//
//	[meta]
//	base_locale    = en
//	locale         = fr
//	locale_name    = French
//	rtl            = false
//	fallback_chain = en
//
//	[strings]
//	guard_greeting:line0:a1b2c3d4 = "Vous. Arrêtez."
//	guard_greeting:line1:e5f6a7b8 = ""
//
// Lines beginning with // are comments; the special markers // [stale] and
// // [orphan] are written by the diff engine.
package loc

import (
	"fmt"
	"strings"
)

// --------------------------------------------------------------------------
// Data types
// --------------------------------------------------------------------------

// Meta holds the [meta] block fields.
type Meta struct {
	BaseLocale    string   // source language code (e.g. "en")
	Locale        string   // target language code (e.g. "fr")
	LocaleName    string   // human-readable name (optional)
	RTL           bool     // right-to-left layout
	FallbackChain []string // ordered fallback locales
}

// Entry is one line in the [strings] block.
type Entry struct {
	Key    string
	Value  string // empty string = untranslated
	Stale  bool   // source text changed since last export
	Orphan bool   // key no longer present in source
	// T-LOC11: metadata fields parsed from comment lines above the entry.
	Type  string // e.g. "spoken", "choice", "narration"
	Char  string // character name who speaks this string
	Scene string // dialogue node or cutscene title
	Ctx   string // author context / translator note
}

// StringsFile is the fully parsed representation of one .agstrings file.
type StringsFile struct {
	Meta    Meta
	Entries []Entry
	// index for O(1) lookup
	index map[string]int // key → index in Entries
}

// Get returns the translation for key, or "" if not present or untranslated.
func (sf *StringsFile) Get(key string) string {
	if sf.index == nil {
		return ""
	}
	if i, ok := sf.index[key]; ok {
		return sf.Entries[i].Value
	}
	return ""
}

// Index returns the internal key→index map for O(1) lookup.
func (sf *StringsFile) Index() map[string]int {
	return sf.index
}

// --------------------------------------------------------------------------
// DiffEntry
// --------------------------------------------------------------------------

// DiffKind classifies a change in the diff output.
type DiffKind int

const (
	DiffAdded     DiffKind = iota // key exists in updated but not in base
	DiffChanged                   // key exists in both but source text hash differs
	DiffRemoved                   // key exists in base but not in updated
	DiffUnchanged                 // key unchanged
)

// DiffEntry describes one key's diff status.
type DiffEntry struct {
	Key      string
	Kind     DiffKind
	OldValue string // previous translation (for Changed/Removed)
	NewValue string // new value (for Added/Changed)
}

// --------------------------------------------------------------------------
// Parse
// --------------------------------------------------------------------------

// Parse parses src as a .agstrings file.
// filename is used only in error messages.
func Parse(filename, src string) (*StringsFile, error) {
	sf := &StringsFile{index: make(map[string]int)}
	lines := strings.Split(src, "\n")
	section := ""
	lineNum := 0
	var pendingMeta map[string]string

	for _, raw := range lines {
		lineNum++
		line := strings.TrimSpace(raw)

		// Blank lines
		if line == "" {
			continue
		}

		// Comments — check for stale/orphan markers on string entries
		if strings.HasPrefix(line, "//") {
			body := strings.TrimSpace(strings.TrimPrefix(line, "//"))
			if strings.HasPrefix(body, "[stale]") || strings.HasPrefix(body, "[orphan]") {
				rest := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(body, "[stale]"), "[orphan]"))
				if e, ok := parseKV(rest); ok {
					stale := strings.HasPrefix(body, "[stale]")
					orphan := strings.HasPrefix(body, "[orphan]")
					entry := Entry{Key: e.key, Value: e.value, Stale: stale, Orphan: orphan}
					applyMeta(&entry, pendingMeta)
					sf.index[entry.Key] = len(sf.Entries)
					sf.Entries = append(sf.Entries, entry)
					pendingMeta = nil
				}
			} else if section == "strings" {
				// T-LOC11: metadata comment lines (// type:, // char:, // scene:, // ctx:).
				if pendingMeta == nil {
					pendingMeta = make(map[string]string)
				}
				if strings.HasPrefix(body, "type:") {
					pendingMeta["type"] = strings.TrimSpace(body[len("type:"):])
				} else if strings.HasPrefix(body, "char:") {
					pendingMeta["char"] = strings.TrimSpace(body[len("char:"):])
				} else if strings.HasPrefix(body, "scene:") {
					pendingMeta["scene"] = strings.TrimSpace(body[len("scene:"):])
				} else if strings.HasPrefix(body, "ctx:") {
					pendingMeta["ctx"] = strings.TrimSpace(body[len("ctx:"):])
				}
			}
			continue
		}

		// Section headers
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = line[1 : len(line)-1]
			continue
		}

		switch section {
		case "meta":
			kv, ok := parseKV(line)
			if !ok {
				return nil, fmt.Errorf("%s:%d: malformed meta line: %q", filename, lineNum, line)
			}
			switch kv.key {
			case "base_locale":
				sf.Meta.BaseLocale = kv.value
			case "locale":
				sf.Meta.Locale = kv.value
			case "locale_name":
				sf.Meta.LocaleName = kv.value
			case "rtl":
				sf.Meta.RTL = kv.value == "true"
			case "fallback_chain":
				sf.Meta.FallbackChain = strings.Fields(kv.value)
			default:
				return nil, fmt.Errorf("%s:%d: unknown meta field %q", filename, lineNum, kv.key)
			}
		case "strings":
			kv, ok := parseKV(line)
			if !ok {
				return nil, fmt.Errorf("%s:%d: malformed strings line: %q", filename, lineNum, line)
			}
			if _, dup := sf.index[kv.key]; dup {
				return nil, fmt.Errorf("%s:%d: duplicate key %q", filename, lineNum, kv.key)
			}
			entry := Entry{Key: kv.key, Value: kv.value}
			applyMeta(&entry, pendingMeta)
			sf.index[entry.Key] = len(sf.Entries)
			sf.Entries = append(sf.Entries, entry)
			pendingMeta = nil
		default:
			if section != "" {
				return nil, fmt.Errorf("%s:%d: unknown section [%s]", filename, lineNum, section)
			}
			return nil, fmt.Errorf("%s:%d: content outside section: %q", filename, lineNum, line)
		}
	}

	if sf.Meta.Locale == "" {
		return nil, fmt.Errorf("%s: [meta] block missing required field 'locale'", filename)
	}
	if sf.Meta.BaseLocale == "" {
		return nil, fmt.Errorf("%s: [meta] block missing required field 'base_locale'", filename)
	}
	return sf, nil
}

// applyMeta copies metadata from the pending map into an entry.
func applyMeta(entry *Entry, meta map[string]string) {
	if meta == nil {
		return
	}
	if v, ok := meta["type"]; ok {
		entry.Type = v
	}
	if v, ok := meta["char"]; ok {
		entry.Char = v
	}
	if v, ok := meta["scene"]; ok {
		entry.Scene = v
	}
	if v, ok := meta["ctx"]; ok {
		entry.Ctx = v
	}
}

// --------------------------------------------------------------------------
// Write
// --------------------------------------------------------------------------

// Write serialises a StringsFile to its canonical text representation.
func Write(sf *StringsFile) string {
	var sb strings.Builder

	sb.WriteString("[meta]\n")
	fmt.Fprintf(&sb, "base_locale    = %s\n", sf.Meta.BaseLocale)
	fmt.Fprintf(&sb, "locale         = %s\n", sf.Meta.Locale)
	if sf.Meta.LocaleName != "" {
		fmt.Fprintf(&sb, "locale_name    = %s\n", sf.Meta.LocaleName)
	}
	if sf.Meta.RTL {
		sb.WriteString("rtl            = true\n")
	}
	if len(sf.Meta.FallbackChain) > 0 {
		fmt.Fprintf(&sb, "fallback_chain = %s\n", strings.Join(sf.Meta.FallbackChain, " "))
	}

	sb.WriteString("\n[strings]\n")
	for _, e := range sf.Entries {
		writeEntry(&sb, e)
	}

	return sb.String()
}

// writeEntry writes one entry with optional metadata comments.
func writeEntry(sb *strings.Builder, e Entry) {
	if e.Type != "" || e.Char != "" || e.Scene != "" || e.Ctx != "" {
		if e.Type != "" {
			fmt.Fprintf(sb, "// type: %s\n", e.Type)
		}
		if e.Char != "" {
			fmt.Fprintf(sb, "// char: %s\n", e.Char)
		}
		if e.Scene != "" {
			fmt.Fprintf(sb, "// scene: %s\n", e.Scene)
		}
		if e.Ctx != "" {
			fmt.Fprintf(sb, "// ctx: %s\n", e.Ctx)
		}
	}
	if e.Stale {
		fmt.Fprintf(sb, "// [stale] %s = %s\n", e.Key, quoteString(e.Value))
	} else if e.Orphan {
		fmt.Fprintf(sb, "// [orphan] %s = %s\n", e.Key, quoteString(e.Value))
	} else {
		fmt.Fprintf(sb, "%s = %s\n", e.Key, quoteString(e.Value))
	}
}

// --------------------------------------------------------------------------
// Diff
// --------------------------------------------------------------------------

// Diff computes the difference between a base StringsFile (existing translation)
// and an updated key set (new source export). Returns one DiffEntry per key.
//
// updated maps key → source text hash (the hash component of the loc key itself).
// Since loc keys already encode a hash, pass the key strings directly; Diff
// treats any key present in updated-but-not-base as Added, any present in
// base-but-not-updated as Removed, and matching keys as Unchanged.
func Diff(base *StringsFile, updatedKeys []string) []DiffEntry {
	updatedSet := make(map[string]bool, len(updatedKeys))
	for _, k := range updatedKeys {
		updatedSet[k] = true
	}

	var out []DiffEntry

	// Keys in base
	baseSet := make(map[string]bool)
	for _, e := range base.Entries {
		if e.Orphan {
			continue // already marked orphan
		}
		baseSet[e.Key] = true
		if updatedSet[e.Key] {
			out = append(out, DiffEntry{Key: e.Key, Kind: DiffUnchanged, OldValue: e.Value, NewValue: e.Value})
		} else {
			out = append(out, DiffEntry{Key: e.Key, Kind: DiffRemoved, OldValue: e.Value})
		}
	}

	// Keys in updated but not base
	for _, k := range updatedKeys {
		if !baseSet[k] {
			out = append(out, DiffEntry{Key: k, Kind: DiffAdded})
		}
	}

	return out
}

// Apply merges a diff result into a StringsFile, returning a new StringsFile
// with stale/orphan markers applied and new entries appended.
func Apply(base *StringsFile, diff []DiffEntry) *StringsFile {
	out := &StringsFile{
		Meta:  base.Meta,
		index: make(map[string]int),
	}

	// Build a map of base entries by key for quick lookup.
	baseByKey := make(map[string]Entry)
	for _, e := range base.Entries {
		baseByKey[e.Key] = e
	}

	// Apply diff entries.
	seen := make(map[string]bool)
	for _, d := range diff {
		switch d.Kind {
		case DiffUnchanged:
			// Preserve existing entry as-is.
			e := baseByKey[d.Key]
			e.Stale = false
			e.Orphan = false
			out.index[e.Key] = len(out.Entries)
			out.Entries = append(out.Entries, e)
		case DiffAdded:
			// New key — add with empty value.
			e := Entry{Key: d.Key, Value: ""}
			out.index[e.Key] = len(out.Entries)
			out.Entries = append(out.Entries, e)
		case DiffRemoved:
			// Mark existing entry as orphan.
			e := baseByKey[d.Key]
			e.Orphan = true
			out.index[e.Key] = len(out.Entries)
			out.Entries = append(out.Entries, e)
		}
		seen[d.Key] = true
	}

	// Preserve existing orphan entries not in diff.
	for _, e := range base.Entries {
		if e.Orphan && !seen[e.Key] {
			out.index[e.Key] = len(out.Entries)
			out.Entries = append(out.Entries, e)
		}
	}

	return out
}

// --------------------------------------------------------------------------
// Helpers
// --------------------------------------------------------------------------

type kvPair struct{ key, value string }

// parseKV parses a line of the form: key = "value" or key = value.
// Returns false if the line does not match.
func parseKV(line string) (kvPair, bool) {
	idx := strings.IndexByte(line, '=')
	if idx < 0 {
		return kvPair{}, false
	}
	key := strings.TrimSpace(line[:idx])
	val := strings.TrimSpace(line[idx+1:])
	if key == "" {
		return kvPair{}, false
	}
	// Strip quotes from value if present.
	if len(val) >= 2 && val[0] == '"' && val[len(val)-1] == '"' {
		val = unquoteString(val[1 : len(val)-1])
	}
	return kvPair{key: key, value: val}, true
}

// quoteString wraps s in double quotes with minimal escaping.
func quoteString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\t", `\t`)
	return `"` + s + `"`
}

// unquoteString decodes escape sequences inside a quoted string body
// (caller must strip the surrounding quotes before calling).
func unquoteString(s string) string {
	var b strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\\' && i+1 < len(s) {
			switch s[i+1] {
			case '"':
				b.WriteByte('"')
			case '\\':
				b.WriteByte('\\')
			case 'n':
				b.WriteByte('\n')
			case 't':
				b.WriteByte('\t')
			default:
				b.WriteByte('\\')
				b.WriteByte(s[i+1])
			}
			i += 2
		} else {
			b.WriteByte(s[i])
			i++
		}
	}
	return b.String()
}
