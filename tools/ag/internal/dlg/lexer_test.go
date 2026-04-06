package dlg_test

import (
	"testing"

	"github.com/ags3d/ag/internal/dlg"
)

// tok is a compact (Kind, Value) pair for table-driven assertions.
type tok struct {
	kind  dlg.TokenKind
	value string
}

func assertTokens(t *testing.T, src string, want []tok) {
	t.Helper()
	got, err := dlg.Lex("test.agdlg", src)
	if err != nil {
		t.Fatalf("Lex error: %v", err)
	}
	// Drop trailing TokEOF from got before comparing.
	if len(got) > 0 && got[len(got)-1].Kind == dlg.TokEOF {
		got = got[:len(got)-1]
	}
	if len(got) != len(want) {
		t.Fatalf("token count: got %d, want %d\ngot:  %v\nwant: %v", len(got), len(want), got, want)
	}
	for i, w := range want {
		g := got[i]
		if g.Kind != w.kind || g.Value != w.value {
			t.Errorf("token[%d]: got (%s, %q), want (%s, %q)", i, g.Kind, g.Value, w.kind, w.value)
		}
	}
}

// --- Header section ---

func TestLex_HeaderKeyValue(t *testing.T) {
	src := "title: guard_greeting\n---\n==="
	assertTokens(t, src, []tok{
		{dlg.TokHeaderKey, "title"},
		{dlg.TokHeaderValue, "guard_greeting"},
		{dlg.TokSeparator, "---"},
		{dlg.TokNodeEnd, "==="},
	})
}

func TestLex_HeaderMultipleKeys(t *testing.T) {
	src := "title: intro\ncharacter: guard\n---\n==="
	assertTokens(t, src, []tok{
		{dlg.TokHeaderKey, "title"},
		{dlg.TokHeaderValue, "intro"},
		{dlg.TokHeaderKey, "character"},
		{dlg.TokHeaderValue, "guard"},
		{dlg.TokSeparator, "---"},
		{dlg.TokNodeEnd, "==="},
	})
}

func TestLex_HeaderTagList(t *testing.T) {
	src := "tags: [chapter:1, cinematic]\n---\n==="
	assertTokens(t, src, []tok{
		{dlg.TokHeaderKey, "tags"},
		{dlg.TokHeaderValue, "[chapter:1, cinematic]"},
		{dlg.TokTag, "chapter:1"},
		{dlg.TokTag, "cinematic"},
		{dlg.TokSeparator, "---"},
		{dlg.TokNodeEnd, "==="},
	})
}

func TestLex_HeaderComment(t *testing.T) {
	src := "// guard_dialogue.agdlg\ntitle: test\n---\n==="
	assertTokens(t, src, []tok{
		{dlg.TokComment, "guard_dialogue.agdlg"},
		{dlg.TokHeaderKey, "title"},
		{dlg.TokHeaderValue, "test"},
		{dlg.TokSeparator, "---"},
		{dlg.TokNodeEnd, "==="},
	})
}

func TestLex_MalformedHeaderLine(t *testing.T) {
	_, err := dlg.Lex("test.agdlg", "notaheader\n---\n===")
	if err == nil {
		t.Fatal("expected error for malformed header line, got nil")
	}
}

// --- Body: speaker lines ---

func TestLex_SpeakerLine(t *testing.T) {
	src := "title: t\n---\nGuard: You there. Stop.\n==="
	assertTokens(t, src, []tok{
		{dlg.TokHeaderKey, "title"},
		{dlg.TokHeaderValue, "t"},
		{dlg.TokSeparator, "---"},
		{dlg.TokSpeaker, "Guard"},
		{dlg.TokLine, "You there. Stop."},
		{dlg.TokNodeEnd, "==="},
	})
}

func TestLex_SpeakerLineWithCommand(t *testing.T) {
	src := "title: t\n---\nGuard: Hello. <<action flag.met = true>>\n==="
	assertTokens(t, src, []tok{
		{dlg.TokHeaderKey, "title"},
		{dlg.TokHeaderValue, "t"},
		{dlg.TokSeparator, "---"},
		{dlg.TokSpeaker, "Guard"},
		{dlg.TokLine, "Hello."},
		{dlg.TokCommand, "action flag.met = true"},
		{dlg.TokNodeEnd, "==="},
	})
}

// --- Body: narration lines ---

func TestLex_NarrationLine(t *testing.T) {
	src := "title: t\n---\nOnce upon a time.\n==="
	assertTokens(t, src, []tok{
		{dlg.TokHeaderKey, "title"},
		{dlg.TokHeaderValue, "t"},
		{dlg.TokSeparator, "---"},
		{dlg.TokLine, "Once upon a time."},
		{dlg.TokNodeEnd, "==="},
	})
}

// --- Body: option lines ---

func TestLex_OptionLine(t *testing.T) {
	src := "title: t\n---\n-> Just passing through.\n==="
	assertTokens(t, src, []tok{
		{dlg.TokHeaderKey, "title"},
		{dlg.TokHeaderValue, "t"},
		{dlg.TokSeparator, "---"},
		{dlg.TokOption, "Just passing through."},
		{dlg.TokNodeEnd, "==="},
	})
}

func TestLex_OptionWithCondition(t *testing.T) {
	src := "title: t\n---\n-> Pass through. <<visible_if not flag.suspicious>>\n==="
	assertTokens(t, src, []tok{
		{dlg.TokHeaderKey, "title"},
		{dlg.TokHeaderValue, "t"},
		{dlg.TokSeparator, "---"},
		{dlg.TokOption, "Pass through."},
		{dlg.TokCommand, "visible_if not flag.suspicious"},
		{dlg.TokNodeEnd, "==="},
	})
}

func TestLex_OptionDepth(t *testing.T) {
	src := "title: t\n---\n-> Top level option.\n   -> Nested option.\n==="
	got, err := dlg.Lex("test.agdlg", src)
	if err != nil {
		t.Fatalf("Lex error: %v", err)
	}
	var opts []dlg.Token
	for _, tok := range got {
		if tok.Kind == dlg.TokOption {
			opts = append(opts, tok)
		}
	}
	if len(opts) != 2 {
		t.Fatalf("expected 2 TokOption tokens, got %d", len(opts))
	}
	if opts[0].Depth != 0 {
		t.Errorf("top-level option depth: got %d, want 0", opts[0].Depth)
	}
	if opts[1].Depth != 1 {
		t.Errorf("nested option depth: got %d, want 1", opts[1].Depth)
	}
}

// --- Body: standalone commands ---

func TestLex_StandaloneCommand(t *testing.T) {
	src := "title: t\n---\n<<action flag.gate_open = true>>\n==="
	assertTokens(t, src, []tok{
		{dlg.TokHeaderKey, "title"},
		{dlg.TokHeaderValue, "t"},
		{dlg.TokSeparator, "---"},
		{dlg.TokCommand, "action flag.gate_open = true"},
		{dlg.TokNodeEnd, "==="},
	})
}

func TestLex_JumpCommand(t *testing.T) {
	src := "title: t\n---\n<<jump guard_checks_pass>>\n==="
	assertTokens(t, src, []tok{
		{dlg.TokHeaderKey, "title"},
		{dlg.TokHeaderValue, "t"},
		{dlg.TokSeparator, "---"},
		{dlg.TokCommand, "jump guard_checks_pass"},
		{dlg.TokNodeEnd, "==="},
	})
}

func TestLex_EndCommand(t *testing.T) {
	src := "title: t\n---\n<<end>>\n==="
	assertTokens(t, src, []tok{
		{dlg.TokHeaderKey, "title"},
		{dlg.TokHeaderValue, "t"},
		{dlg.TokSeparator, "---"},
		{dlg.TokCommand, "end"},
		{dlg.TokNodeEnd, "==="},
	})
}

// --- Body: loc key ---

func TestLex_LocKey(t *testing.T) {
	src := "title: t\n---\nGuard: Hi there. #loc:guard_hi_001\n==="
	assertTokens(t, src, []tok{
		{dlg.TokHeaderKey, "title"},
		{dlg.TokHeaderValue, "t"},
		{dlg.TokSeparator, "---"},
		{dlg.TokSpeaker, "Guard"},
		{dlg.TokLine, "Hi there."},
		{dlg.TokLocKey, "guard_hi_001"},
		{dlg.TokNodeEnd, "==="},
	})
}

func TestLex_LocKeyOnOptionLine(t *testing.T) {
	src := "title: t\n---\n-> Let me pass. #loc:opt_pass_001\n==="
	assertTokens(t, src, []tok{
		{dlg.TokHeaderKey, "title"},
		{dlg.TokHeaderValue, "t"},
		{dlg.TokSeparator, "---"},
		{dlg.TokOption, "Let me pass."},
		{dlg.TokLocKey, "opt_pass_001"},
		{dlg.TokNodeEnd, "==="},
	})
}

// --- Body: comment lines ---

func TestLex_BodyComment(t *testing.T) {
	src := "title: t\n---\n// narrator speaks\nOnce upon a time.\n==="
	assertTokens(t, src, []tok{
		{dlg.TokHeaderKey, "title"},
		{dlg.TokHeaderValue, "t"},
		{dlg.TokSeparator, "---"},
		{dlg.TokComment, "narrator speaks"},
		{dlg.TokLine, "Once upon a time."},
		{dlg.TokNodeEnd, "==="},
	})
}

// --- Multi-node file ---

func TestLex_TwoNodes(t *testing.T) {
	src := "title: node_a\n---\nGuard: Hello.\n===\ntitle: node_b\n---\nGuard: Goodbye.\n==="
	got, err := dlg.Lex("test.agdlg", src)
	if err != nil {
		t.Fatalf("Lex error: %v", err)
	}
	// Count node-end markers — should be 2.
	ends := 0
	for _, tok := range got {
		if tok.Kind == dlg.TokNodeEnd {
			ends++
		}
	}
	if ends != 2 {
		t.Errorf("expected 2 TokNodeEnd tokens, got %d", ends)
	}
	// Count speaker tokens — should be 2.
	speakers := 0
	for _, tok := range got {
		if tok.Kind == dlg.TokSpeaker {
			speakers++
		}
	}
	if speakers != 2 {
		t.Errorf("expected 2 TokSpeaker tokens, got %d", speakers)
	}
}

// --- Full example from spec ---

func TestLex_FullExampleNode(t *testing.T) {
	src := `title: guard_greeting
character: guard
tags: [chapter:1]
---
Guard: You there. Stop.
-> I'm just passing through. <<visible_if not flag.guard_suspicious>>
   Guard: Move along then.
   <<action flag.guard_spoken = true>>
   <<end>>
-> I have a pass. <<visible_if item.gate_pass in player.inventory>>
   Guard: Let me see that.
   <<jump guard_checks_pass>>
-> Never mind.
   <<end>>
===`

	got, err := dlg.Lex("guard_dialogue.agdlg", src)
	if err != nil {
		t.Fatalf("Lex error: %v", err)
	}

	// Spot-check key token kinds and values.
	checks := []struct {
		kind  dlg.TokenKind
		value string
	}{
		{dlg.TokHeaderKey, "title"},
		{dlg.TokHeaderValue, "guard_greeting"},
		{dlg.TokHeaderKey, "character"},
		{dlg.TokHeaderValue, "guard"},
		{dlg.TokSeparator, "---"},
		{dlg.TokSpeaker, "Guard"},
		{dlg.TokLine, "You there. Stop."},
		{dlg.TokOption, "I'm just passing through."},
		{dlg.TokCommand, "visible_if not flag.guard_suspicious"},
		{dlg.TokNodeEnd, "==="},
	}

	// Build a map from value to kind for quick lookup.
	byValue := make(map[string]dlg.TokenKind)
	for _, t := range got {
		byValue[t.Value] = t.Kind
	}

	for _, c := range checks {
		if k, ok := byValue[c.value]; !ok {
			t.Errorf("expected token with value %q not found", c.value)
		} else if k != c.kind {
			t.Errorf("token value %q: got kind %s, want %s", c.value, k, c.kind)
		}
	}
}

// --- Error cases ---

func TestLex_UnclosedCommand(t *testing.T) {
	src := "title: t\n---\n<<action flag = true\n==="
	_, err := dlg.Lex("test.agdlg", src)
	if err == nil {
		t.Fatal("expected error for unclosed <<, got nil")
	}
}

func TestLex_EmptyFile(t *testing.T) {
	got, err := dlg.Lex("empty.agdlg", "")
	if err != nil {
		t.Fatalf("unexpected error on empty file: %v", err)
	}
	if len(got) != 1 || got[0].Kind != dlg.TokEOF {
		t.Errorf("expected single TokEOF, got %v", got)
	}
}

func TestLex_BlankLinesIgnored(t *testing.T) {
	src := "title: t\n\n\n---\n\nGuard: Hi.\n\n==="
	assertTokens(t, src, []tok{
		{dlg.TokHeaderKey, "title"},
		{dlg.TokHeaderValue, "t"},
		{dlg.TokSeparator, "---"},
		{dlg.TokSpeaker, "Guard"},
		{dlg.TokLine, "Hi."},
		{dlg.TokNodeEnd, "==="},
	})
}

func TestLex_WindowsLineEndings(t *testing.T) {
	src := "title: t\r\n---\r\nGuard: Hi.\r\n===\r\n"
	assertTokens(t, src, []tok{
		{dlg.TokHeaderKey, "title"},
		{dlg.TokHeaderValue, "t"},
		{dlg.TokSeparator, "---"},
		{dlg.TokSpeaker, "Guard"},
		{dlg.TokLine, "Hi."},
		{dlg.TokNodeEnd, "==="},
	})
}

func TestLex_PosLineNumbers(t *testing.T) {
	src := "title: t\n---\nGuard: Hi.\n==="
	got, err := dlg.Lex("test.agdlg", src)
	if err != nil {
		t.Fatalf("Lex error: %v", err)
	}
	for _, tok := range got {
		if tok.Pos.File != "test.agdlg" {
			t.Errorf("token %s: expected File=test.agdlg, got %q", tok.Kind, tok.Pos.File)
		}
		if tok.Pos.Line < 1 {
			t.Errorf("token %s: expected Line >= 1, got %d", tok.Kind, tok.Pos.Line)
		}
	}
}
