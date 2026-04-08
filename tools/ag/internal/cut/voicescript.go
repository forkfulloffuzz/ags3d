package cut

import (
	"encoding/json"
	"fmt"
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
		locKey := cmd.Params["loc_key"]
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
