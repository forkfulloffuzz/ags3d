package cut_test

import (
	"strings"
	"testing"

	"github.com/ags3d/ag/internal/cut"
)

func mustParseVS(t *testing.T, src string) *cut.CutsceneFile {
	t.Helper()
	cf, err := cut.Parse("test.agcut", src)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	return cf
}

// --- CollectVoiceLines ---

func TestCollect_LineCommandExtracted(t *testing.T) {
	cf := mustParseVS(t, "title: intro\nsequence:\n<<line narrator \"Three years.\">>\n<<end>>\n")
	lines := cut.CollectVoiceLines([]*cut.CutsceneFile{cf})
	if len(lines) != 1 {
		t.Fatalf("lines len = %d, want 1", len(lines))
	}
	l := lines[0]
	if l.Char != "narrator" {
		t.Errorf("Char = %q, want narrator", l.Char)
	}
	if l.Text != "Three years." {
		t.Errorf("Text = %q, want 'Three years.'", l.Text)
	}
	if l.Cutscene != "intro" {
		t.Errorf("Cutscene = %q, want intro", l.Cutscene)
	}
}

func TestCollect_SessionFromHeader(t *testing.T) {
	cf := mustParseVS(t, "title: t\nvoice_session: vs_chapter1\nsequence:\n<<line player \"Hello.\">>\n<<end>>\n")
	lines := cut.CollectVoiceLines([]*cut.CutsceneFile{cf})
	if lines[0].Session != "vs_chapter1" {
		t.Errorf("Session = %q, want vs_chapter1", lines[0].Session)
	}
}

func TestCollect_EmotionParam(t *testing.T) {
	cf := mustParseVS(t, "title: t\nsequence:\n<<line player \"Hello.\" emotion:sad>>\n<<end>>\n")
	lines := cut.CollectVoiceLines([]*cut.CutsceneFile{cf})
	if lines[0].Emotion != "sad" {
		t.Errorf("Emotion = %q, want sad", lines[0].Emotion)
	}
}

func TestCollect_LocKeyParam(t *testing.T) {
	cf := mustParseVS(t, "title: t\nsequence:\n<<line player \"Hello.\" loc_key:player_hello>>\n<<end>>\n")
	lines := cut.CollectVoiceLines([]*cut.CutsceneFile{cf})
	if lines[0].LocKey != "player_hello" {
		t.Errorf("LocKey = %q, want player_hello", lines[0].LocKey)
	}
}

func TestCollect_AutoLocKeyGenerated(t *testing.T) {
	cf := mustParseVS(t, "title: intro\nsequence:\n<<line narrator \"Hello.\">>\n<<end>>\n")
	lines := cut.CollectVoiceLines([]*cut.CutsceneFile{cf})
	if lines[0].LocKey == "" {
		t.Error("LocKey should not be empty when no loc_key: param")
	}
	if !strings.HasPrefix(lines[0].LocKey, "intro:line") {
		t.Errorf("LocKey = %q, want prefix intro:line", lines[0].LocKey)
	}
}

func TestCollect_PrecedingContext(t *testing.T) {
	cf := mustParseVS(t, "title: t\nsequence:\n<<fade_in duration:1.0>>\n<<line narrator \"Hello.\">>\n<<end>>\n")
	lines := cut.CollectVoiceLines([]*cut.CutsceneFile{cf})
	if !strings.Contains(lines[0].Preceding, "fade_in") {
		t.Errorf("Preceding = %q, want to contain fade_in", lines[0].Preceding)
	}
}

func TestCollect_NonLineCommandsSkipped(t *testing.T) {
	cf := mustParseVS(t, "title: t\nsequence:\n<<fade_in duration:1.0>>\n<<music theme>>\n<<end>>\n")
	lines := cut.CollectVoiceLines([]*cut.CutsceneFile{cf})
	if len(lines) != 0 {
		t.Errorf("expected 0 voice lines, got %d", len(lines))
	}
}

func TestCollect_MultipleFiles(t *testing.T) {
	a := mustParseVS(t, "title: a\nsequence:\n<<line narrator \"A.\">>\n<<end>>\n")
	b := mustParseVS(t, "title: b\nsequence:\n<<line player \"B.\">>\n<<end>>\n")
	lines := cut.CollectVoiceLines([]*cut.CutsceneFile{a, b})
	if len(lines) != 2 {
		t.Errorf("lines len = %d, want 2", len(lines))
	}
}

// --- RenderVoicescripts ---

func TestRender_GroupsBySessionAndChar(t *testing.T) {
	a := mustParseVS(t, "title: a\nvoice_session: s1\nsequence:\n<<line player \"Hello.\">>\n<<end>>\n")
	b := mustParseVS(t, "title: b\nvoice_session: s1\nsequence:\n<<line guard \"Stop.\">>\n<<end>>\n")
	lines := cut.CollectVoiceLines([]*cut.CutsceneFile{a, b})
	groups := cut.RenderVoicescripts(lines, nil, "")
	if _, ok := groups["s1/player"]; !ok {
		t.Error("expected group s1/player")
	}
	if _, ok := groups["s1/guard"]; !ok {
		t.Error("expected group s1/guard")
	}
}

func TestRender_DefaultSessionKey(t *testing.T) {
	cf := mustParseVS(t, "title: t\nsequence:\n<<line narrator \"Hi.\">>\n<<end>>\n")
	lines := cut.CollectVoiceLines([]*cut.CutsceneFile{cf})
	groups := cut.RenderVoicescripts(lines, nil, "")
	if _, ok := groups["_default/narrator"]; !ok {
		t.Errorf("expected group _default/narrator, got keys: %v", groupKeys(groups))
	}
}

func TestRender_CharFilterApplied(t *testing.T) {
	cf := mustParseVS(t, "title: t\nsequence:\n<<line player \"Hi.\">>\n<<line guard \"Stop.\">>\n<<end>>\n")
	lines := cut.CollectVoiceLines([]*cut.CutsceneFile{cf})
	groups := cut.RenderVoicescripts(lines, nil, "player")
	if _, ok := groups["_default/guard"]; ok {
		t.Error("guard should be filtered out")
	}
	if _, ok := groups["_default/player"]; !ok {
		t.Error("player should be present")
	}
}

func TestRender_ContentContainsText(t *testing.T) {
	cf := mustParseVS(t, "title: t\nsequence:\n<<line narrator \"Three years.\">>\n<<end>>\n")
	lines := cut.CollectVoiceLines([]*cut.CutsceneFile{cf})
	groups := cut.RenderVoicescripts(lines, nil, "")
	md := groups["_default/narrator"]
	if !strings.Contains(md, "Three years.") {
		t.Errorf("Markdown missing source text: %s", md)
	}
}

func TestRender_TranslationIncluded(t *testing.T) {
	cf := mustParseVS(t, "title: t\nsequence:\n<<line narrator \"Hello.\" loc_key:nar_hello>>\n<<end>>\n")
	lines := cut.CollectVoiceLines([]*cut.CutsceneFile{cf})
	tr := map[string]string{"nar_hello": "Bonjour."}
	groups := cut.RenderVoicescripts(lines, tr, "")
	md := groups["_default/narrator"]
	if !strings.Contains(md, "Bonjour.") {
		t.Errorf("Markdown missing translation: %s", md)
	}
}

func TestRender_EmotionIncluded(t *testing.T) {
	cf := mustParseVS(t, "title: t\nsequence:\n<<line player \"Stop.\" emotion:angry>>\n<<end>>\n")
	lines := cut.CollectVoiceLines([]*cut.CutsceneFile{cf})
	groups := cut.RenderVoicescripts(lines, nil, "")
	md := groups["_default/player"]
	if !strings.Contains(md, "angry") {
		t.Errorf("Markdown missing emotion: %s", md)
	}
}

func groupKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// --- ExportVoiceSessionsJSON ---

func TestExportVoiceSessionsJSON_Basic(t *testing.T) {
	cf := mustParseVS(t, "title: intro\nvoice_session: act1\nsequence:\n<<line guard \"Stop!\">>\n<<end>>\n")
	jsonStr, err := cut.ExportVoiceSessionsJSON([]*cut.CutsceneFile{cf})
	if err != nil {
		t.Fatalf("ExportVoiceSessionsJSON: %v", err)
	}
	if !strings.Contains(jsonStr, `"name": "act1"`) {
		t.Errorf("JSON missing session name: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"character": "guard"`) {
		t.Errorf("JSON missing character: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"lines"`) {
		t.Errorf("JSON missing lines array: %s", jsonStr)
	}
}

func TestExportVoiceSessionsJSON_NoSession(t *testing.T) {
	cf := mustParseVS(t, "title: intro\nsequence:\n<<line guard \"Stop!\">>\n<<end>>\n")
	jsonStr, err := cut.ExportVoiceSessionsJSON([]*cut.CutsceneFile{cf})
	if err != nil {
		t.Fatalf("ExportVoiceSessionsJSON: %v", err)
	}
	if !strings.Contains(jsonStr, `"sessions"`) {
		t.Errorf("Expected sessions field, got: %s", jsonStr)
	}
}

func TestExportVoiceSessionsJSON_MultipleLines(t *testing.T) {
	cf := mustParseVS(t, "title: intro\nvoice_session: act1\nsequence:\n<<line guard \"Stop!\">>\n<<line guard \"Who goes there?\">>\n<<end>>\n")
	jsonStr, err := cut.ExportVoiceSessionsJSON([]*cut.CutsceneFile{cf})
	if err != nil {
		t.Fatalf("ExportVoiceSessionsJSON: %v", err)
	}
	if !strings.Contains(jsonStr, `"name": "act1"`) {
		t.Errorf("JSON missing session name: %s", jsonStr)
	}
}
