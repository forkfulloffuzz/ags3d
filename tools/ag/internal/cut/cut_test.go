package cut_test

import (
	"testing"

	"github.com/ags3d/ag/internal/cut"
)

func mustParse(t *testing.T, src string) *cut.CutsceneFile {
	t.Helper()
	cf, err := cut.Parse("test.agcut", src)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	return cf
}

// --- Header parsing ---

func TestParse_TitleField(t *testing.T) {
	cf := mustParse(t, "title: chapter1_opening\nsequence:\n<<end>>\n")
	if cf.Title != "chapter1_opening" {
		t.Errorf("Title = %q, want chapter1_opening", cf.Title)
	}
}

func TestParse_SkipField(t *testing.T) {
	cf := mustParse(t, "title: t\nskip: after_first_view\nsequence:\n<<end>>\n")
	if cf.Skip != "after_first_view" {
		t.Errorf("Skip = %q, want after_first_view", cf.Skip)
	}
}

func TestParse_SaveBlockDefault(t *testing.T) {
	cf := mustParse(t, "title: t\nsequence:\n<<end>>\n")
	if !cf.SaveBlock {
		t.Error("SaveBlock default should be true")
	}
}

func TestParse_SaveBlockFalse(t *testing.T) {
	cf := mustParse(t, "title: t\nsave_block: false\nsequence:\n<<end>>\n")
	if cf.SaveBlock {
		t.Error("SaveBlock should be false when set to false")
	}
}

func TestParse_SaveBlockTrue(t *testing.T) {
	cf := mustParse(t, "title: t\nsave_block: true\nsequence:\n<<end>>\n")
	if !cf.SaveBlock {
		t.Error("SaveBlock should be true when set to true")
	}
}

func TestParse_TagsField(t *testing.T) {
	cf := mustParse(t, "title: t\ntags: [chapter:1, cinematic]\nsequence:\n<<end>>\n")
	if len(cf.Tags) != 2 {
		t.Fatalf("Tags len = %d, want 2", len(cf.Tags))
	}
	if cf.Tags[0] != "chapter:1" || cf.Tags[1] != "cinematic" {
		t.Errorf("Tags = %v", cf.Tags)
	}
}

func TestParse_FallbackField(t *testing.T) {
	cf := mustParse(t, "title: t\nfallback: halt\nsequence:\n<<end>>\n")
	if cf.Fallback != "halt" {
		t.Errorf("Fallback = %q, want halt", cf.Fallback)
	}
}

func TestParse_LocGroupField(t *testing.T) {
	cf := mustParse(t, "title: t\nloc_group: chapter1\nsequence:\n<<end>>\n")
	if cf.LocGroup != "chapter1" {
		t.Errorf("LocGroup = %q, want chapter1", cf.LocGroup)
	}
}

func TestParse_VoiceSessionField(t *testing.T) {
	cf := mustParse(t, "title: t\nvoice_session: session_a\nsequence:\n<<end>>\n")
	if cf.VoiceSession != "session_a" {
		t.Errorf("VoiceSession = %q, want session_a", cf.VoiceSession)
	}
}

func TestParse_LanguageField(t *testing.T) {
	cf := mustParse(t, "title: t\nlanguage: fr\nsequence:\n<<end>>\n")
	if cf.Language != "fr" {
		t.Errorf("Language = %q, want fr", cf.Language)
	}
}

func TestParse_LanguageField_DefaultsToEmpty(t *testing.T) {
	cf := mustParse(t, "title: t\nsequence:\n<<end>>\n")
	if cf.Language != "" {
		t.Errorf("Language = %q, want empty string", cf.Language)
	}
}

func TestParse_AudioScopeField(t *testing.T) {
	cf := mustParse(t, "title: t\naudio_scope: pause\nsequence:\n<<end>>\n")
	if cf.AudioScope != "pause" {
		t.Errorf("AudioScope = %q, want pause", cf.AudioScope)
	}
}

func TestParse_AudioScopeDefault(t *testing.T) {
	cf := mustParse(t, "title: t\nsequence:\n<<end>>\n")
	if cf.AudioScope != "keep" {
		t.Errorf("AudioScope default = %q, want keep", cf.AudioScope)
	}
}

func TestParse_DuckChannelsField(t *testing.T) {
	cf := mustParse(t, "title: t\nduck_channels: room_music room_ambient\nsequence:\n<<end>>\n")
	if cf.DuckChannels != "room_music room_ambient" {
		t.Errorf("DuckChannels = %q", cf.DuckChannels)
	}
}

func TestParse_DuckLevelField(t *testing.T) {
	cf := mustParse(t, "title: t\nduck_level: 0.15\nsequence:\n<<end>>\n")
	if cf.DuckLevel != 0.15 {
		t.Errorf("DuckLevel = %v, want 0.15", cf.DuckLevel)
	}
}

func TestParse_DuckLevelDefault(t *testing.T) {
	cf := mustParse(t, "title: t\nsequence:\n<<end>>\n")
	if cf.DuckLevel != 0.25 {
		t.Errorf("DuckLevel default = %v, want 0.25", cf.DuckLevel)
	}
}

func TestParse_DuckFadeField(t *testing.T) {
	cf := mustParse(t, "title: t\nduck_fade: 0.5\nsequence:\n<<end>>\n")
	if cf.DuckFade != 0.5 {
		t.Errorf("DuckFade = %v, want 0.5", cf.DuckFade)
	}
}

func TestParse_DuckFadeDefault(t *testing.T) {
	cf := mustParse(t, "title: t\nsequence:\n<<end>>\n")
	if cf.DuckFade != 0.3 {
		t.Errorf("DuckFade default = %v, want 0.3", cf.DuckFade)
	}
}

func TestParse_DuckRestoreField(t *testing.T) {
	cf := mustParse(t, "title: t\nduck_restore: 1.0\nsequence:\n<<end>>\n")
	if cf.DuckRestore != 1.0 {
		t.Errorf("DuckRestore = %v, want 1.0", cf.DuckRestore)
	}
}

func TestParse_DuckRestoreDefault(t *testing.T) {
	cf := mustParse(t, "title: t\nsequence:\n<<end>>\n")
	if cf.DuckRestore != 0.5 {
		t.Errorf("DuckRestore default = %v, want 0.5", cf.DuckRestore)
	}
}

func TestParse_AutoDuckTrue(t *testing.T) {
	cf := mustParse(t, "title: t\nauto_duck: true\nsequence:\n<<end>>\n")
	if !cf.AutoDuck {
		t.Error("AutoDuck should be true")
	}
}

func TestParse_AutoDuckDefault(t *testing.T) {
	cf := mustParse(t, "title: t\nsequence:\n<<end>>\n")
	if cf.AutoDuck {
		t.Error("AutoDuck default should be false")
	}
}

func TestValidAudioScopes(t *testing.T) {
	for _, v := range []string{"keep", "pause", "stop"} {
		if !cut.ValidAudioScopes[v] {
			t.Errorf("expected %q in ValidAudioScopes", v)
		}
	}
	if cut.ValidAudioScopes["mute"] {
		t.Error("expected 'mute' not in ValidAudioScopes")
	}
}

func TestParse_UnknownHeaderKeyIgnored(t *testing.T) {
	cf := mustParse(t, "title: t\nfuture_field: value\nsequence:\n<<end>>\n")
	if cf.Title != "t" {
		t.Errorf("Title = %q after unknown key", cf.Title)
	}
}

func TestParse_MalformedHeaderReturnsError(t *testing.T) {
	_, err := cut.Parse("test.agcut", "notaheader\nsequence:\n<<end>>\n")
	if err == nil {
		t.Fatal("expected error for malformed header, got nil")
	}
}

func TestParse_BlankLinesIgnored(t *testing.T) {
	cf := mustParse(t, "title: t\n\nsequence:\n\n<<end>>\n")
	if len(cf.Sequence) != 1 {
		t.Errorf("Sequence len = %d, want 1", len(cf.Sequence))
	}
}

func TestParse_CommentsIgnored(t *testing.T) {
	cf := mustParse(t, "// opening\ntitle: t\nsequence:\n// step comment\n<<end>>\n")
	if len(cf.Sequence) != 1 {
		t.Errorf("Sequence len = %d, want 1 (comment should be ignored)", len(cf.Sequence))
	}
}

func TestParse_WindowsLineEndings(t *testing.T) {
	cf := mustParse(t, "title: t\r\nsequence:\r\n<<end>>\r\n")
	if cf.Title != "t" {
		t.Errorf("Title = %q with CRLF", cf.Title)
	}
}

// --- Sequence body: command parsing ---

func TestParse_SequenceStartProducesNoCommand(t *testing.T) {
	cf := mustParse(t, "title: t\nsequence:\n<<end>>\n")
	if len(cf.Sequence) != 1 || cf.Sequence[0].Name != "end" {
		t.Errorf("Sequence = %v, want [end]", cf.Sequence)
	}
}

func TestParse_CommandNameExtracted(t *testing.T) {
	cf := mustParse(t, "title: t\nsequence:\n<<fade_in duration:2.0>>\n")
	if len(cf.Sequence) != 1 {
		t.Fatalf("Sequence len = %d, want 1", len(cf.Sequence))
	}
	if cf.Sequence[0].Name != "fade_in" {
		t.Errorf("Name = %q, want fade_in", cf.Sequence[0].Name)
	}
}

func TestParse_CommandArgsPreserved(t *testing.T) {
	cf := mustParse(t, "title: t\nsequence:\n<<fade_in duration:2.0>>\n")
	if cf.Sequence[0].Args != "duration:2.0" {
		t.Errorf("Args = %q, want duration:2.0", cf.Sequence[0].Args)
	}
}

func TestParse_CommandNoArgs(t *testing.T) {
	cf := mustParse(t, "title: t\nsequence:\n<<end>>\n")
	if cf.Sequence[0].Args != "" {
		t.Errorf("Args = %q, want empty", cf.Sequence[0].Args)
	}
}

func TestParse_CommandMultipleArgs(t *testing.T) {
	cf := mustParse(t, "title: t\nsequence:\n<<camera move_to point.street_level duration:4.0 ease:out bg:cam_move>>\n")
	cmd := cf.Sequence[0]
	if cmd.Name != "camera" {
		t.Errorf("Name = %q, want camera", cmd.Name)
	}
	if cmd.Args != "move_to point.street_level duration:4.0 ease:out bg:cam_move" {
		t.Errorf("Args = %q", cmd.Args)
	}
}

func TestParse_CommandStringArg(t *testing.T) {
	cf := mustParse(t, `title: t`+"\nsequence:\n"+`<<title_card "Chapter One: The Market District" duration:2.0>>`+"\n")
	cmd := cf.Sequence[0]
	if cmd.Name != "title_card" {
		t.Errorf("Name = %q, want title_card", cmd.Name)
	}
	if cmd.Args != `"Chapter One: The Market District" duration:2.0` {
		t.Errorf("Args = %q", cmd.Args)
	}
}

func TestParse_CommandLineNarrator(t *testing.T) {
	cf := mustParse(t, "title: t\nsequence:\n<<line narrator \"Three years.\" >>\n")
	cmd := cf.Sequence[0]
	if cmd.Name != "line" {
		t.Errorf("Name = %q, want line", cmd.Name)
	}
}

func TestParse_ActionCommandRawExpr(t *testing.T) {
	cf := mustParse(t, "title: t\nsequence:\n<<action flag.chapter1_started = true>>\n")
	cmd := cf.Sequence[0]
	if cmd.Name != "action" {
		t.Errorf("Name = %q, want action", cmd.Name)
	}
	if cmd.Args != "flag.chapter1_started = true" {
		t.Errorf("Args = %q, want 'flag.chapter1_started = true'", cmd.Args)
	}
}

// --- Block commands ---

func TestParse_ParallelIsBlockOpen(t *testing.T) {
	cf := mustParse(t, "title: t\nsequence:\n<<parallel>>\n<<end_parallel>>\n")
	if !cf.Sequence[0].IsBlockOpen {
		t.Error("<<parallel>> should be IsBlockOpen=true")
	}
	if cf.Sequence[0].IsBlockClose {
		t.Error("<<parallel>> should not be IsBlockClose")
	}
}

func TestParse_EndParallelIsBlockClose(t *testing.T) {
	cf := mustParse(t, "title: t\nsequence:\n<<parallel>>\n<<end_parallel>>\n")
	if !cf.Sequence[1].IsBlockClose {
		t.Error("<<end_parallel>> should be IsBlockClose=true")
	}
	if cf.Sequence[1].IsBlockOpen {
		t.Error("<<end_parallel>> should not be IsBlockOpen")
	}
}

func TestParse_IfIsBlockOpen(t *testing.T) {
	cf := mustParse(t, "title: t\nsequence:\n<<if flag.x>>\n<<end_if>>\n")
	if !cf.Sequence[0].IsBlockOpen {
		t.Error("<<if>> should be IsBlockOpen=true")
	}
}

func TestParse_EndIfIsBlockClose(t *testing.T) {
	cf := mustParse(t, "title: t\nsequence:\n<<if flag.x>>\n<<end_if>>\n")
	if !cf.Sequence[1].IsBlockClose {
		t.Error("<<end_if>> should be IsBlockClose=true")
	}
}

func TestParse_OnIsBlockOpen(t *testing.T) {
	cf := mustParse(t, "title: t\nsequence:\n<<on event:player:found_clue>>\n<<end_on>>\n")
	if !cf.Sequence[0].IsBlockOpen {
		t.Error("<<on>> should be IsBlockOpen=true")
	}
}

func TestParse_EndOnIsBlockClose(t *testing.T) {
	cf := mustParse(t, "title: t\nsequence:\n<<on event:player:found_clue>>\n<<end_on>>\n")
	if !cf.Sequence[1].IsBlockClose {
		t.Error("<<end_on>> should be IsBlockClose=true")
	}
}

func TestParse_RegularCommandNotBlock(t *testing.T) {
	cf := mustParse(t, "title: t\nsequence:\n<<fade_in duration:2.0>>\n")
	cmd := cf.Sequence[0]
	if cmd.IsBlockOpen || cmd.IsBlockClose {
		t.Error("<<fade_in>> should not be block")
	}
}

// --- Position info ---

func TestParse_PosFileAndLine(t *testing.T) {
	cf := mustParse(t, "title: t\nsequence:\n<<fade_in>>\n")
	if cf.Sequence[0].Pos.File != "test.agcut" {
		t.Errorf("Pos.File = %q, want test.agcut", cf.Sequence[0].Pos.File)
	}
	if cf.Sequence[0].Pos.Line != 3 {
		t.Errorf("Pos.Line = %d, want 3", cf.Sequence[0].Pos.Line)
	}
}

// --- Error cases ---

func TestParse_EmptyCommandIsError(t *testing.T) {
	_, err := cut.Parse("test.agcut", "title: t\nsequence:\n<<>>\n")
	if err == nil {
		t.Fatal("expected error for <<>>, got nil")
	}
}

func TestParse_NonCommandBodyLineIsError(t *testing.T) {
	_, err := cut.Parse("test.agcut", "title: t\nsequence:\nnot_a_command\n")
	if err == nil {
		t.Fatal("expected error for non-command body line, got nil")
	}
}

// --- Full spec example ---

func TestParse_FullSpecExample(t *testing.T) {
	src := `// cutscenes/chapter1_opening.agcut
title: chapter1_opening
skip: after_first_view
save_block: true
tags: [chapter:1, cinematic]
fallback: halt
sequence:
<<fade_in duration:2.0>>
<<camera set point.rooftop_wide fov:60>>
<<music theme_main fade_in:3.0>>
<<title_card "Chapter One: The Market District" duration:2.0>>
<<camera move_to point.street_level duration:4.0 ease:out bg:cam_move>>
<<character player spawn_at point.alley_entrance>>
<<sync cam_move>>
<<character player animation play:look_around>>
<<line narrator "Three years. And it still felt foreign.">>
<<action flag.chapter1_started = true>>
<<action room.transition("market")>>
<<fade_out duration:1.0>>
`
	cf := mustParse(t, src)

	if cf.Title != "chapter1_opening" {
		t.Errorf("Title = %q", cf.Title)
	}
	if cf.Skip != "after_first_view" {
		t.Errorf("Skip = %q", cf.Skip)
	}
	if !cf.SaveBlock {
		t.Error("SaveBlock should be true")
	}
	if cf.Fallback != "halt" {
		t.Errorf("Fallback = %q", cf.Fallback)
	}
	if len(cf.Tags) != 2 {
		t.Errorf("Tags = %v", cf.Tags)
	}
	if len(cf.Sequence) != 12 {
		t.Errorf("Sequence len = %d, want 12", len(cf.Sequence))
	}
	if cf.Sequence[0].Name != "fade_in" {
		t.Errorf("Sequence[0].Name = %q, want fade_in", cf.Sequence[0].Name)
	}
	if cf.Sequence[len(cf.Sequence)-1].Name != "fade_out" {
		t.Errorf("last command = %q, want fade_out", cf.Sequence[len(cf.Sequence)-1].Name)
	}
}

// --- Token kind strings ---

func TestTokenKindStrings(t *testing.T) {
	cases := map[cut.TokenKind]string{
		cut.TokEOF:           "EOF",
		cut.TokComment:       "COMMENT",
		cut.TokHeaderKey:     "HEADER_KEY",
		cut.TokHeaderValue:   "HEADER_VALUE",
		cut.TokTag:           "TAG",
		cut.TokSequenceStart: "SEQUENCE_START",
		cut.TokCommandName:   "COMMAND_NAME",
		cut.TokNamedParam:    "NAMED_PARAM",
		cut.TokStringValue:   "STRING_VALUE",
		cut.TokIdentifier:    "IDENTIFIER",
		cut.TokBlockOpen:     "BLOCK_OPEN",
		cut.TokBlockClose:    "BLOCK_CLOSE",
	}
	for kind, want := range cases {
		if got := kind.String(); got != want {
			t.Errorf("TokenKind(%d).String() = %q, want %q", int(kind), got, want)
		}
	}
}

// --- ValidSkipPolicies / ValidFallbacks ---

func TestValidSkipPolicies(t *testing.T) {
	for _, v := range []string{"always", "never", "after_first_view", "author_controlled"} {
		if !cut.ValidSkipPolicies[v] {
			t.Errorf("expected %q in ValidSkipPolicies", v)
		}
	}
	if cut.ValidSkipPolicies["invalid"] {
		t.Error("expected 'invalid' not in ValidSkipPolicies")
	}
}

func TestValidFallbacks(t *testing.T) {
	for _, v := range []string{"halt", "skip_and_continue", "log_and_continue", "retry_once"} {
		if !cut.ValidFallbacks[v] {
			t.Errorf("expected %q in ValidFallbacks", v)
		}
	}
}
