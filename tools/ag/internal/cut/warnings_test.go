package cut_test

import (
	"testing"

	"github.com/ags3d/ag/internal/cut"
)

func warn(t *testing.T, src string) []cut.ValidationWarning {
	t.Helper()
	cf := mustParse(t, src)
	return cut.WarnCutscene(cf, nil, nil)
}

func hasWarnCode(warns []cut.ValidationWarning, code string) bool {
	for _, w := range warns {
		if w.Code == code {
			return true
		}
	}
	return false
}

// --- CUT-W006: no <<end>> and no room.transition ---

func TestWarn_NoEndNoTransitionIsW006(t *testing.T) {
	ws := warn(t, "title: t\nsequence:\n<<fade_in duration:1.0>>\n")
	if !hasWarnCode(ws, "CUT-W006") {
		t.Error("expected CUT-W006 for sequence with no end, none found")
	}
}

func TestWarn_HasEndNoW006(t *testing.T) {
	ws := warn(t, "title: t\nsequence:\n<<fade_in duration:1.0>>\n<<end>>\n")
	for _, w := range ws {
		if w.Code == "CUT-W006" {
			t.Errorf("unexpected CUT-W006: %v", w)
		}
	}
}

func TestWarn_HasRoomTransitionNoW006(t *testing.T) {
	ws := warn(t, "title: t\nsequence:\n<<action room.transition(\"market\")>>\n")
	for _, w := range ws {
		if w.Code == "CUT-W006" {
			t.Errorf("unexpected CUT-W006: %v", w)
		}
	}
}

// --- CUT-W007: label declared but never used as skip_to target ---

func TestWarn_UnusedLabelIsW007(t *testing.T) {
	ws := warn(t, "title: t\nsequence:\n<<label unused_point>>\n<<end>>\n")
	if !hasWarnCode(ws, "CUT-W007") {
		t.Error("expected CUT-W007 for unused label, none found")
	}
}

func TestWarn_UsedLabelNoW007(t *testing.T) {
	ws := warn(t, "title: t\nsequence:\n<<label act_two>>\n<<skip_to act_two>>\n<<end>>\n")
	for _, w := range ws {
		if w.Code == "CUT-W007" {
			t.Errorf("unexpected CUT-W007: %v", w)
		}
	}
}

// --- CUT-W008: author_controlled with no labels ---

func TestWarn_AuthorControlledNoLabelsIsW008(t *testing.T) {
	ws := warn(t, "title: t\nskip: author_controlled\nsequence:\n<<end>>\n")
	if !hasWarnCode(ws, "CUT-W008") {
		t.Error("expected CUT-W008 for author_controlled with no labels, none found")
	}
}

func TestWarn_AuthorControlledWithLabelNoW008(t *testing.T) {
	ws := warn(t, "title: t\nskip: author_controlled\nsequence:\n<<label skip_here>>\n<<skip_to skip_here>>\n<<end>>\n")
	for _, w := range ws {
		if w.Code == "CUT-W008" {
			t.Errorf("unexpected CUT-W008: %v", w)
		}
	}
}

func TestWarn_OtherSkipPolicyNoW008(t *testing.T) {
	ws := warn(t, "title: t\nskip: never\nsequence:\n<<end>>\n")
	for _, w := range ws {
		if w.Code == "CUT-W008" {
			t.Errorf("unexpected CUT-W008: %v", w)
		}
	}
}

// --- CUT-W009: audio channel started but never stopped ---

func TestWarn_MusicNotStoppedIsW009(t *testing.T) {
	ws := warn(t, "title: t\nsequence:\n<<music theme_main fade_in:3.0>>\n<<end>>\n")
	if !hasWarnCode(ws, "CUT-W009") {
		t.Error("expected CUT-W009 for music not stopped, none found")
	}
}

func TestWarn_MusicStoppedNoW009(t *testing.T) {
	ws := warn(t, "title: t\nsequence:\n<<music theme_main fade_in:3.0>>\n<<music theme_main stop>>\n<<end>>\n")
	for _, w := range ws {
		if w.Code == "CUT-W009" {
			t.Errorf("unexpected CUT-W009: %v", w)
		}
	}
}

func TestWarn_AmbientStoppedByChannelParamNoW009(t *testing.T) {
	ws := warn(t, "title: t\nsequence:\n<<ambient market_day volume:0.5>>\n<<ambient stop channel:market_day>>\n<<end>>\n")
	for _, w := range ws {
		if w.Code == "CUT-W009" {
			t.Errorf("unexpected CUT-W009: %v", w)
		}
	}
}

func TestWarn_SoundNotStoppedIsW009(t *testing.T) {
	ws := warn(t, "title: t\nsequence:\n<<sound rain_sfx loop:true>>\n<<end>>\n")
	if !hasWarnCode(ws, "CUT-W009") {
		t.Error("expected CUT-W009 for sound not stopped, none found")
	}
}

// --- CUT-W010: duck:all in command args ---

func TestWarn_DuckAllIsW010(t *testing.T) {
	ws := warn(t, "title: t\nsequence:\n<<line narrator \"Hello.\" duck:all>>\n<<end>>\n")
	if !hasWarnCode(ws, "CUT-W010") {
		t.Error("expected CUT-W010 for duck:all, none found")
	}
}

func TestWarn_DuckSpecificChannelNoW010(t *testing.T) {
	ws := warn(t, "title: t\nsequence:\n<<line narrator \"Hello.\" duck:room_music>>\n<<end>>\n")
	for _, w := range ws {
		if w.Code == "CUT-W010" {
			t.Errorf("unexpected CUT-W010: %v", w)
		}
	}
}

// --- CUT-W011: auto_duck:true with no duck_channels ---

func TestWarn_AutoDuckNoDuckChannelsIsW011(t *testing.T) {
	ws := warn(t, "title: t\nauto_duck: true\nsequence:\n<<end>>\n")
	if !hasWarnCode(ws, "CUT-W011") {
		t.Error("expected CUT-W011 for auto_duck with no duck_channels, none found")
	}
}

func TestWarn_AutoDuckWithChannelsNoW011(t *testing.T) {
	ws := warn(t, "title: t\nauto_duck: true\nduck_channels: room_music\nsequence:\n<<end>>\n")
	for _, w := range ws {
		if w.Code == "CUT-W011" {
			t.Errorf("unexpected CUT-W011: %v", w)
		}
	}
}

func TestWarn_AutoDuckFalseNoW011(t *testing.T) {
	ws := warn(t, "title: t\nsequence:\n<<end>>\n")
	for _, w := range ws {
		if w.Code == "CUT-W011" {
			t.Errorf("unexpected CUT-W011: %v", w)
		}
	}
}

// --- CUT-W001: cutscene never referenced (project-level) ---

func TestWarnProject_UnreferencedCutsceneIsW001(t *testing.T) {
	a := mustParse(t, "title: intro\nsequence:\n<<end>>\n")
	b := mustParse(t, "title: unused\nsequence:\n<<end>>\n")
	// Only 'a' references nothing; both are unreferenced in this small project.
	warns := cut.WarnProjectCutscenes([]*cut.CutsceneFile{a, b}, nil)
	// Both should get W001 since neither is called.
	w001count := 0
	for _, w := range warns {
		if w.Code == "CUT-W001" {
			w001count++
		}
	}
	if w001count != 2 {
		t.Errorf("expected 2 CUT-W001 (both unreferenced), got %d", w001count)
	}
}

func TestWarnProject_ReferencedCutsceneNoW001(t *testing.T) {
	// 'caller' triggers 'intro' — intro should not get W001.
	caller, err := cut.Parse("caller.agcut", "title: caller\nsequence:\n<<cutscene intro>>\n<<end>>\n")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	intro, err := cut.Parse("intro.agcut", "title: intro\nsequence:\n<<end>>\n")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	warns := cut.WarnProjectCutscenes([]*cut.CutsceneFile{caller, intro}, nil)
	for _, w := range warns {
		if w.Code == "CUT-W001" && w.Pos.File == "intro.agcut" {
			t.Errorf("unexpected CUT-W001 for referenced cutscene 'intro': %v", w)
		}
	}
}

// --- ValidationWarning.Error() format ---

func TestValidationWarning_ErrorFormat(t *testing.T) {
	w := cut.ValidationWarning{
		Pos:  cut.Pos{File: "test.agcut", Line: 5, Col: 1},
		Code: "CUT-W006",
		Msg:  "test warning",
	}
	got := w.Error()
	want := "test.agcut:5:1: CUT-W006: test warning"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}
