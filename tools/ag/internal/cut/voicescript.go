package cut

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// VoiceLine is one <<line character "text">> entry collected for a voicescript.
type VoiceLine struct {
	LocKey    string // loc_key: param or auto-generated from title+index
	Session   string // voice_session from the cutscene header (may be empty)
	Cutscene  string // cf.Title
	Char      string // first positional of <<line>>
	Text      string // quoted string arg of <<line>>
	Emotion   string // emotion: param if present
	Preceding string // summary of the command immediately before this line
}

// CollectVoiceLines walks a slice of CutsceneFiles and returns all <<line>>
// commands as VoiceLine entries, in source order.
func CollectVoiceLines(files []*CutsceneFile) []VoiceLine {
	var lines []VoiceLine
	for _, cf := range files {
		lines = append(lines, collectFromFile(cf)...)
	}
	return lines
}

func collectFromFile(cf *CutsceneFile) []VoiceLine {
	var out []VoiceLine
	lineIdx := 0
	for i, rc := range cf.Sequence {
		if rc.Name != "line" {
			continue
		}
		cmd := ParseCommand(rc)
		if len(cmd.Positional) == 0 {
			continue
		}
		char := cmd.Positional[0]

		// Text is the first string value in the raw args.
		text := extractStringArg(rc.Args)

		// Optional params.
		emotion := cmd.Params["emotion"]
		locKey := cmd.Params["#loc_key"]
		if locKey == "" {
			locKey = fmt.Sprintf("%s:line%d", cf.Title, lineIdx)
		}
		lineIdx++

		// Preceding command context.
		preceding := ""
		if i > 0 {
			prev := cf.Sequence[i-1]
			preceding = precedingContext(prev)
		}

		out = append(out, VoiceLine{
			LocKey:    locKey,
			Session:   cf.VoiceSession,
			Cutscene:  cf.Title,
			Char:      char,
			Text:      text,
			Emotion:   emotion,
			Preceding: preceding,
		})
	}
	return out
}

// extractStringArg finds the first "quoted string" in a raw args string.
func extractStringArg(args string) string {
	start := strings.Index(args, `"`)
	if start < 0 {
		return ""
	}
	end := strings.Index(args[start+1:], `"`)
	if end < 0 {
		return ""
	}
	return args[start+1 : start+1+end]
}

// precedingContext returns a short description of the given command for context.
func precedingContext(rc *RawCommand) string {
	if rc.Name == "line" {
		cmd := ParseCommand(rc)
		char := ""
		if len(cmd.Positional) > 0 {
			char = cmd.Positional[0]
		}
		text := extractStringArg(rc.Args)
		if text != "" {
			return fmt.Sprintf("[line:%s] %q", char, text)
		}
		return fmt.Sprintf("[line:%s]", char)
	}
	if rc.Args != "" {
		return fmt.Sprintf("[%s %s]", rc.Name, rc.Args)
	}
	return fmt.Sprintf("[%s]", rc.Name)
}

// RenderVoicescripts produces one Markdown string per (session, character) group.
// The outer map key is "session/character" (or "_default/character" for no session).
// translations maps loc_key → translated text (may be nil).
// charFilter limits output to one character (empty = all).
func RenderVoicescripts(lines []VoiceLine, translations map[string]string, charFilter string) map[string]string {
	// Group lines by session → character.
	type groupKey struct{ session, char string }
	groups := make(map[groupKey][]VoiceLine)
	order := []groupKey{}
	seen := make(map[groupKey]bool)

	for _, vl := range lines {
		if charFilter != "" && vl.Char != charFilter {
			continue
		}
		k := groupKey{session: vl.Session, char: vl.Char}
		if !seen[k] {
			order = append(order, k)
			seen[k] = true
		}
		groups[k] = append(groups[k], vl)
	}

	result := make(map[string]string)
	for _, k := range order {
		session := k.session
		if session == "" {
			session = "_default"
		}
		outKey := session + "/" + k.char
		result[outKey] = renderGroup(k.char, session, groups[k], translations)
	}
	return result
}

func renderGroup(char, session string, lines []VoiceLine, translations map[string]string) string {
	var sb strings.Builder
	displaySession := session
	if displaySession == "_default" {
		displaySession = "(no session)"
	}
	fmt.Fprintf(&sb, "# Voice Session: %s\n## Character: %s\n\n", displaySession, char)

	for _, vl := range lines {
		fmt.Fprintf(&sb, "---\n\n")
		fmt.Fprintf(&sb, "**[%s]**  \n", vl.Cutscene)
		fmt.Fprintf(&sb, "*Loc key: `%s`*  \n", vl.LocKey)
		if vl.Emotion != "" {
			fmt.Fprintf(&sb, "*Emotion: %s*  \n", vl.Emotion)
		}
		if vl.Preceding != "" {
			fmt.Fprintf(&sb, "*Preceding: %s*  \n", vl.Preceding)
		}
		fmt.Fprintf(&sb, "\n> %s\n", vl.Text)
		if translations != nil {
			if tr := translations[vl.LocKey]; tr != "" {
				fmt.Fprintf(&sb, "\n*Translation: %s*\n", tr)
			}
		}
		fmt.Fprintln(&sb)
	}
	return sb.String()
}

type VoiceSessionEntry struct {
	Name      string   `json:"name"`
	Character string   `json:"character"`
	Lines     []string `json:"lines"`
}

type VoiceSessionsJSON struct {
	Sessions []VoiceSessionEntry `json:"sessions"`
}

func ExportVoiceSessionsJSON(files []*CutsceneFile) (string, error) {
	lines := CollectVoiceLines(files)

	type groupKey struct {
		session string
		char    string
	}
	groups := make(map[groupKey][]string)

	for _, vl := range lines {
		if vl.Session == "" {
			continue
		}
		k := groupKey{session: vl.Session, char: vl.Char}
		lineID := vl.Char + "/" + vl.LocKey
		groups[k] = append(groups[k], lineID)
	}

	var sessions []VoiceSessionEntry
	for k, lineIDs := range groups {
		sessions = append(sessions, VoiceSessionEntry{
			Name:      k.session,
			Character: k.char,
			Lines:     lineIDs,
		})
	}

	data := VoiceSessionsJSON{Sessions: sessions}
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal voice sessions: %w", err)
	}
	return string(jsonBytes), nil
}

// ---------------------------------------------------------------------------
// T-LOC16 — Voice coverage tracking
// ---------------------------------------------------------------------------

type VoiceCoverageEntry struct {
	LocKey     string `json:"loc_key"`
	File       string `json:"file"`
	DurationMs int    `json:"duration_ms,omitempty"`
	Hash       string `json:"hash,omitempty"`
}

type VoiceCoverageFile struct {
	Version int                  `json:"version"`
	Locale  string               `json:"locale"`
	Entries []VoiceCoverageEntry `json:"entries"`
}

type VoiceCoverageReport struct {
	Covered []VoiceCoverageEntry // lines with recorded audio
	Missing []VoiceLine          // lines with no recorded audio
	Stale   []StaleVoiceEntry    // lines recorded but source text changed
}

type StaleVoiceEntry struct {
	VoiceLine
	OldHash string `json:"old_hash"`
}

func LoadVoiceCoverage(path string) (*VoiceCoverageFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read voice coverage file: %w", err)
	}
	var cf VoiceCoverageFile
	if err := json.Unmarshal(data, &cf); err != nil {
		return nil, fmt.Errorf("parse voice_coverage.json: %w", err)
	}
	return &cf, nil
}

func ScanVoiceDirectory(root, locale string) ([]VoiceCoverageEntry, error) {
	dir := filepath.Join(root, "audio", "voice", locale)
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return nil, nil
	}

	var entries []VoiceCoverageEntry
	err = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		ext := strings.ToLower(filepath.Ext(path))
		if !isAudioExt(ext) {
			return nil
		}
		rel, _ := filepath.Rel(root, path)

		hash, _ := fileHash(path)
		size := fileSize(path)
		duration := estimateDuration(size, ext)

		locKey := inferLocKeyFromPath(rel)

		entries = append(entries, VoiceCoverageEntry{
			LocKey:     locKey,
			File:       rel,
			DurationMs: duration,
			Hash:       hash,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return entries, nil
}

func inferLocKeyFromPath(rel string) string {
	rel = filepath.ToSlash(rel)
	parts := strings.Split(rel, "/")
	if len(parts) < 2 {
		return ""
	}
	filename := parts[len(parts)-1]
	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filename, ext)
	parts = parts[:len(parts)-1]
	char := parts[len(parts)-1]
	return fmt.Sprintf("%s/%s", char, base)
}

func BuildVoiceCoverageReport(cutsceneFiles []*CutsceneFile, audioEntries []VoiceCoverageEntry, localeMap map[string]string) *VoiceCoverageReport {
	lines := CollectVoiceLines(cutsceneFiles)
	lineMap := make(map[string]VoiceLine, len(lines))
	for _, l := range lines {
		lineMap[l.LocKey] = l
	}

	coveredMap := make(map[string]VoiceCoverageEntry, len(audioEntries))
	for _, ae := range audioEntries {
		coveredMap[ae.LocKey] = ae
	}

	report := &VoiceCoverageReport{}
	for _, ae := range audioEntries {
		if line, ok := lineMap[ae.LocKey]; ok {
			if localeMap != nil && localeMap[ae.LocKey] != "" && localeMap[ae.LocKey] != line.Text {
				report.Stale = append(report.Stale, StaleVoiceEntry{VoiceLine: line, OldHash: ae.Hash})
			} else {
				report.Covered = append(report.Covered, ae)
			}
		}
	}

	for _, line := range lines {
		if _, ok := coveredMap[line.LocKey]; !ok {
			report.Missing = append(report.Missing, line)
		}
	}

	return report
}

func WriteVoiceCoverageJSON(report *VoiceCoverageReport, w *strings.Builder, locale string) {
	fmt.Fprintf(w, "# AGS3D Voice Coverage Report — %s\n\n", locale)
	fmt.Fprintf(w, "## Coverage Summary\n\n")
	total := len(report.Covered) + len(report.Missing) + len(report.Stale)
	fmt.Fprintf(w, "- **Recorded:** %d / %d (%.0f%%)\n", len(report.Covered), total, float64(len(report.Covered))/float64(total)*100)
	fmt.Fprintf(w, "- **Missing:** %d\n", len(report.Missing))
	fmt.Fprintf(w, "- **Stale:** %d\n\n", len(report.Stale))

	if len(report.Stale) > 0 {
		fmt.Fprintf(w, "## Stale Recordings (source text changed)\n\n")
		for _, s := range report.Stale {
			fmt.Fprintf(w, "- **%s** [%s] — %q → %q\n", s.LocKey, s.Cutscene, s.Text, "")
			fmt.Fprintf(w, "  Old hash: `%s`\n\n", s.OldHash)
		}
	}

	if len(report.Missing) > 0 {
		fmt.Fprintf(w, "## Missing Recordings\n\n")
		grouped := groupMissingByCharacter(report.Missing)
		for char, lines := range grouped {
			fmt.Fprintf(w, "### %s (%d lines)\n\n", char, len(lines))
			for _, line := range lines {
				fmt.Fprintf(w, "- **%s** [%s] — %q\n", line.LocKey, line.Cutscene, line.Text)
				if line.Emotion != "" {
					fmt.Fprintf(w, "  emotion: %s\n", line.Emotion)
				}
			}
			fmt.Fprintf(w, "\n")
		}
	}

	if len(report.Covered) > 0 {
		fmt.Fprintf(w, "## Recorded Lines\n\n")
		for _, c := range report.Covered {
			fmt.Fprintf(w, "- **%s** — %s (%.1fs)\n", c.LocKey, c.File, float64(c.DurationMs)/1000.0)
		}
		fmt.Fprintf(w, "\n")
	}
}

func groupMissingByCharacter(lines []VoiceLine) map[string][]VoiceLine {
	m := make(map[string][]VoiceLine)
	for _, l := range lines {
		char := l.Char
		if char == "" {
			char = "(no character)"
		}
		m[char] = append(m[char], l)
	}
	return m
}

func isAudioExt(ext string) bool {
	return ext == ".ogg" || ext == ".opus" || ext == ".mp3" || ext == ".wav" || ext == ".flac"
}

func fileHash(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(data)
	return fmt.Sprintf("sha256:%x", h[:8]), nil
}

func fileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}

func estimateDuration(size int64, ext string) int {
	var bps int
	switch ext {
	case ".ogg", ".opus":
		bps = 16000
	case ".mp3":
		bps = 32000
	case ".wav":
		bps = 176400
	case ".flac":
		bps = 88200
	default:
		bps = 16000
	}
	if bps == 0 {
		return 0
	}
	ms := (size * 8 * 1000) / int64(bps)
	return int(ms)
}
