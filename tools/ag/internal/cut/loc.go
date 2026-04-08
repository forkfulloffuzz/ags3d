// T-CUT30 — Cutscene localisation pass.
//
// Extracts localizable strings from .agcut files and writes .agstrings
// template files. Also validates that every loc_key: used in a cutscene
// sequence is present in the project's active locale file.
package cut

import (
	"fmt"
	"hash/fnv"
	"strings"
)

// LocEntry is one localizable string extracted from a cutscene sequence.
type LocEntry struct {
	// LocKey is the stable identifier: "<cutscene_title>:<seq_idx>:<hash8>".
	// If the source command supplies an explicit loc_key: argument that value
	// is used instead.
	LocKey  string
	Source  string // original source text (content of the string literal)
	CmdName string // "line" | "title_card" | "subtitle" | "choice"
	Pos     Pos
}

// CollectLocEntries walks a CutsceneFile's Sequence and returns all
// localizable strings, in source order.
//
// Localizable commands: <<line>>, <<title_card>>, <<subtitle>>, <<choice>>.
// Each command's quoted string argument is extracted; an explicit loc_key:
// parameter overrides the auto-generated key.
func CollectLocEntries(cf *CutsceneFile) []LocEntry {
	var entries []LocEntry
	var seqIdx int
	for _, cmd := range cf.Sequence {
		switch cmd.Name {
		case "line", "title_card", "subtitle", "choice":
			text, locKey := extractLocArgs(cmd.Args)
			if text == "" {
				continue
			}
			if locKey == "" {
				locKey = cutLocKey(cf.Title, seqIdx, text)
			}
			entries = append(entries, LocEntry{
				LocKey:  locKey,
				Source:  text,
				CmdName: cmd.Name,
				Pos:     cmd.Pos,
			})
			seqIdx++
		}
	}
	return entries
}

// WriteAgstringsTemplate returns the content of a .agstrings template file
// for the given entries. The file uses the native AGS3D localisation format.
//
// The output is a base-locale template (untranslated) suitable for
// submission to translators. Keys are in the form expected by the runtime.
func WriteAgstringsTemplate(title string, entries []LocEntry) string {
	var sb strings.Builder
	sb.WriteString("[meta]\n")
	fmt.Fprintf(&sb, "base_locale    = en\n")
	fmt.Fprintf(&sb, "locale         = en\n")
	fmt.Fprintf(&sb, "// Generated from cutscene: %s\n", title)
	sb.WriteString("\n[strings]\n")
	for _, e := range entries {
		fmt.Fprintf(&sb, "%s = %q\n", e.LocKey, e.Source)
	}
	return sb.String()
}

// ValidateLocKeys checks that every explicit loc_key: used in cf's sequence
// is present in the provided locale map (key → translated value).
// Returns one error string per missing key.
//
// Only commands with an explicit loc_key: argument are validated; auto-generated
// keys are always considered valid (they are generated deterministically).
func ValidateLocKeys(cf *CutsceneFile, localeMap map[string]string) []string {
	var errs []string
	for _, cmd := range cf.Sequence {
		switch cmd.Name {
		case "line", "title_card", "subtitle", "choice":
			_, locKey := extractLocArgs(cmd.Args)
			if locKey == "" {
				continue // auto-generated keys are not validated here
			}
			if _, ok := localeMap[locKey]; !ok {
				errs = append(errs, fmt.Sprintf("%s: cutscene %q: loc_key %q not found in locale file",
					cmd.Pos, cf.Title, locKey))
			}
		}
	}
	return errs
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// extractLocArgs returns the (text, loc_key) pair from a command Args string.
// text is the content of the first double-quoted string literal found.
// loc_key is the value of the "loc_key:" named parameter, or "" if absent.
func extractLocArgs(args string) (text, locKey string) {
	// Extract quoted string — first "..." in args.
	text = extractFirstQuotedString(args)

	// Extract loc_key: value.
	const needle = "loc_key:"
	if idx := strings.Index(args, needle); idx >= 0 {
		rest := strings.TrimSpace(args[idx+len(needle):])
		// Value is either a quoted string or an unquoted identifier.
		if len(rest) > 0 && rest[0] == '"' {
			locKey, _, _ = extractQuotedArgWithRest(rest)
		} else {
			end := strings.IndexAny(rest, " \t\r\n")
			if end < 0 {
				end = len(rest)
			}
			locKey = rest[:end]
		}
	}
	return text, locKey
}

// extractFirstQuotedString returns the content of the first "..." in s.
func extractFirstQuotedString(s string) string {
	start := strings.Index(s, `"`)
	if start < 0 {
		return ""
	}
	end := strings.Index(s[start+1:], `"`)
	if end < 0 {
		return ""
	}
	raw := s[start+1 : start+1+end]
	// Unescape simple escape sequences.
	raw = strings.ReplaceAll(raw, `\"`, `"`)
	raw = strings.ReplaceAll(raw, `\\`, `\`)
	return raw
}

// extractQuotedArgWithRest returns (value, rest, ok) for the first "..." at
// the start of s.
func extractQuotedArgWithRest(s string) (string, string, bool) {
	s = strings.TrimLeft(s, " \t")
	if len(s) == 0 || s[0] != '"' {
		return "", s, false
	}
	end := strings.Index(s[1:], `"`)
	if end < 0 {
		return "", s, false
	}
	return s[1 : end+1], s[end+2:], true
}

// cutLocKey generates a stable loc key for a cutscene string.
// Format: "<title>:<seqIdx>:<hash8>".
func cutLocKey(title string, seqIdx int, text string) string {
	h := fnv.New32a()
	_, _ = h.Write([]byte(text))
	return fmt.Sprintf("%s:%d:%08x", title, seqIdx, h.Sum32())
}
