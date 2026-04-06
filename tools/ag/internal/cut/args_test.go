package cut_test

import (
	"testing"

	"github.com/ags3d/ag/internal/cut"
)

// ---- TokenizeArgs ----

func TestTokenizeArgs_Empty(t *testing.T) {
	if toks := cut.TokenizeArgs(""); len(toks) != 0 {
		t.Errorf("expected 0 tokens, got %d", len(toks))
	}
}

func TestTokenizeArgs_NamedParam(t *testing.T) {
	toks := cut.TokenizeArgs("duration:2.0")
	if len(toks) != 1 || toks[0].Kind != cut.TokNamedParam {
		t.Fatalf("expected 1 NAMED_PARAM, got %v", toks)
	}
	if toks[0].Value != "duration:2.0" {
		t.Errorf("Value = %q, want duration:2.0", toks[0].Value)
	}
}

func TestTokenizeArgs_Identifier(t *testing.T) {
	toks := cut.TokenizeArgs("narrator")
	if len(toks) != 1 || toks[0].Kind != cut.TokIdentifier {
		t.Fatalf("expected 1 IDENTIFIER, got %v", toks)
	}
	if toks[0].Value != "narrator" {
		t.Errorf("Value = %q", toks[0].Value)
	}
}

func TestTokenizeArgs_DottedPathIsIdentifier(t *testing.T) {
	toks := cut.TokenizeArgs("point.street_level")
	if len(toks) != 1 || toks[0].Kind != cut.TokIdentifier {
		t.Fatalf("expected 1 IDENTIFIER for dotted path, got %v", toks)
	}
}

func TestTokenizeArgs_StringValue(t *testing.T) {
	toks := cut.TokenizeArgs(`"Chapter One: The Market District"`)
	if len(toks) != 1 || toks[0].Kind != cut.TokStringValue {
		t.Fatalf("expected 1 STRING_VALUE, got %v", toks)
	}
	if toks[0].Value != "Chapter One: The Market District" {
		t.Errorf("Value = %q", toks[0].Value)
	}
}

func TestTokenizeArgs_Mixed(t *testing.T) {
	toks := cut.TokenizeArgs(`move_to point.street_level duration:4.0 ease:out bg:cam_move`)
	if len(toks) != 5 {
		t.Fatalf("expected 5 tokens, got %d: %v", len(toks), toks)
	}
	if toks[0].Kind != cut.TokIdentifier {
		t.Errorf("toks[0] kind = %s, want IDENTIFIER", toks[0].Kind)
	}
	if toks[1].Kind != cut.TokIdentifier {
		t.Errorf("toks[1] kind = %s, want IDENTIFIER", toks[1].Kind)
	}
	if toks[2].Kind != cut.TokNamedParam || toks[3].Kind != cut.TokNamedParam || toks[4].Kind != cut.TokNamedParam {
		t.Errorf("expected NAMED_PARAM for toks[2..4]")
	}
}

func TestTokenizeArgs_StringThenNamedParam(t *testing.T) {
	toks := cut.TokenizeArgs(`"Chapter One" duration:2.0`)
	if len(toks) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(toks))
	}
	if toks[0].Kind != cut.TokStringValue {
		t.Errorf("toks[0] = %s, want STRING_VALUE", toks[0].Kind)
	}
	if toks[1].Kind != cut.TokNamedParam {
		t.Errorf("toks[1] = %s, want NAMED_PARAM", toks[1].Kind)
	}
}

// ---- NamedParamValue ----

func TestNamedParamValue_Basic(t *testing.T) {
	tok := cut.ArgToken{Kind: cut.TokNamedParam, Value: "duration:2.0"}
	k, v := cut.NamedParamValue(tok)
	if k != "duration" || v != "2.0" {
		t.Errorf("got key=%q value=%q", k, v)
	}
}

func TestNamedParamValue_DottedValue(t *testing.T) {
	tok := cut.ArgToken{Kind: cut.TokNamedParam, Value: "play:look_around"}
	k, v := cut.NamedParamValue(tok)
	if k != "play" || v != "look_around" {
		t.Errorf("got key=%q value=%q", k, v)
	}
}

// ---- ParseCommand ----

func makeRC(name, args string) *cut.RawCommand {
	return &cut.RawCommand{
		Name:         name,
		Args:         args,
		IsBlockOpen:  cut.IsBlockOpenName(name),
		IsBlockClose: cut.IsBlockCloseName(name),
	}
}

func TestParseCommand_FadeIn(t *testing.T) {
	cmd := cut.ParseCommand(makeRC("fade_in", "duration:2.0"))
	if cmd.Name != "fade_in" {
		t.Errorf("Name = %q", cmd.Name)
	}
	if cmd.Params["duration"] != "2.0" {
		t.Errorf("Params[duration] = %q, want 2.0", cmd.Params["duration"])
	}
}

func TestParseCommand_Camera(t *testing.T) {
	cmd := cut.ParseCommand(makeRC("camera", "move_to point.street_level duration:4.0 ease:out bg:cam_move"))
	if len(cmd.Positional) != 2 {
		t.Fatalf("Positional = %v, want [move_to, point.street_level]", cmd.Positional)
	}
	if cmd.Positional[0] != "move_to" || cmd.Positional[1] != "point.street_level" {
		t.Errorf("Positional = %v", cmd.Positional)
	}
	if cmd.Params["duration"] != "4.0" {
		t.Errorf("Params[duration] = %q", cmd.Params["duration"])
	}
	if cmd.Params["ease"] != "out" {
		t.Errorf("Params[ease] = %q", cmd.Params["ease"])
	}
	if cmd.Params["bg"] != "cam_move" {
		t.Errorf("Params[bg] = %q", cmd.Params["bg"])
	}
}

func TestParseCommand_LineWithText(t *testing.T) {
	cmd := cut.ParseCommand(makeRC("line", `narrator "Three years. And it still felt foreign."`))
	if cmd.Text != "Three years. And it still felt foreign." {
		t.Errorf("Text = %q", cmd.Text)
	}
	if len(cmd.Positional) != 1 || cmd.Positional[0] != "narrator" {
		t.Errorf("Positional = %v, want [narrator]", cmd.Positional)
	}
}

func TestParseCommand_TitleCard(t *testing.T) {
	cmd := cut.ParseCommand(makeRC("title_card", `"Chapter One: The Market District" duration:2.0`))
	if cmd.Text != "Chapter One: The Market District" {
		t.Errorf("Text = %q", cmd.Text)
	}
	if cmd.Params["duration"] != "2.0" {
		t.Errorf("Params[duration] = %q", cmd.Params["duration"])
	}
}

func TestParseCommand_ActionIsRawExpr(t *testing.T) {
	cmd := cut.ParseCommand(makeRC("action", "flag.chapter1_started = true"))
	if cmd.Expr != "flag.chapter1_started = true" {
		t.Errorf("Expr = %q", cmd.Expr)
	}
	if len(cmd.Params) != 0 {
		t.Errorf("expected no parsed params for action, got %v", cmd.Params)
	}
}

func TestParseCommand_IfIsCondition(t *testing.T) {
	cmd := cut.ParseCommand(makeRC("if", "flag.x and not flag.y"))
	if cmd.Condition != "flag.x and not flag.y" {
		t.Errorf("Condition = %q", cmd.Condition)
	}
}

func TestParseCommand_SetIsRawExpr(t *testing.T) {
	cmd := cut.ParseCommand(makeRC("set", "counter = 3"))
	if cmd.Expr != "counter = 3" {
		t.Errorf("Expr = %q", cmd.Expr)
	}
}

func TestParseCommand_SyncNoArgs(t *testing.T) {
	cmd := cut.ParseCommand(makeRC("sync", ""))
	if len(cmd.Positional) != 0 {
		t.Errorf("expected empty Positional for <<sync>>, got %v", cmd.Positional)
	}
}

func TestParseCommand_SyncWithID(t *testing.T) {
	cmd := cut.ParseCommand(makeRC("sync", "cam_move"))
	if len(cmd.Positional) != 1 || cmd.Positional[0] != "cam_move" {
		t.Errorf("Positional = %v, want [cam_move]", cmd.Positional)
	}
}

func TestParseCommand_CharacterAnimation(t *testing.T) {
	cmd := cut.ParseCommand(makeRC("character", "player animation play:look_around"))
	if len(cmd.Positional) < 2 {
		t.Fatalf("Positional = %v", cmd.Positional)
	}
	if cmd.Positional[0] != "player" || cmd.Positional[1] != "animation" {
		t.Errorf("Positional = %v", cmd.Positional)
	}
	if cmd.Params["play"] != "look_around" {
		t.Errorf("Params[play] = %q", cmd.Params["play"])
	}
}

func TestParseCommand_BlockOpen(t *testing.T) {
	cmd := cut.ParseCommand(makeRC("parallel", ""))
	if !cmd.IsBlockOpen {
		t.Error("parallel should be IsBlockOpen")
	}
}

func TestParseCommand_BlockClose(t *testing.T) {
	cmd := cut.ParseCommand(makeRC("end_parallel", ""))
	if !cmd.IsBlockClose {
		t.Error("end_parallel should be IsBlockClose")
	}
}

// ---- ParseSequenceTree ----

func TestParseSequenceTree_FlatSequence(t *testing.T) {
	cmds := cut.ParseSequence([]*cut.RawCommand{
		makeRC("fade_in", "duration:2.0"),
		makeRC("fade_out", "duration:1.0"),
	})
	seq := cut.ParseSequenceTree(cmds)
	if len(seq.Steps) != 2 {
		t.Errorf("expected 2 steps, got %d", len(seq.Steps))
	}
}

func TestParseSequenceTree_ParallelBlock(t *testing.T) {
	cmds := cut.ParseSequence([]*cut.RawCommand{
		makeRC("parallel", ""),
		makeRC("music", "theme_main"),
		makeRC("camera", "shake intensity:0.5"),
		makeRC("end_parallel", ""),
	})
	seq := cut.ParseSequenceTree(cmds)
	if len(seq.Steps) != 1 {
		t.Fatalf("expected 1 top-level step, got %d", len(seq.Steps))
	}
	if seq.Steps[0].Body == nil {
		t.Fatal("parallel step should have a Body")
	}
	if len(seq.Steps[0].Body.Steps) != 2 {
		t.Errorf("parallel body has %d steps, want 2", len(seq.Steps[0].Body.Steps))
	}
}

func TestParseSequenceTree_IfElseBlock(t *testing.T) {
	cmds := cut.ParseSequence([]*cut.RawCommand{
		makeRC("if", "flag.x"),
		makeRC("fade_in", ""),
		makeRC("else", ""),
		makeRC("fade_out", ""),
		makeRC("end_if", ""),
	})
	seq := cut.ParseSequenceTree(cmds)
	if len(seq.Steps) != 1 {
		t.Fatalf("expected 1 top-level step, got %d", len(seq.Steps))
	}
	step := seq.Steps[0]
	if step.Body == nil || len(step.Body.Steps) != 1 {
		t.Errorf("if body: %v", step.Body)
	}
	if step.Else == nil || len(step.Else.Steps) != 1 {
		t.Errorf("else body: %v", step.Else)
	}
}

func TestParseSequenceTree_NestedBlocks(t *testing.T) {
	cmds := cut.ParseSequence([]*cut.RawCommand{
		makeRC("parallel", ""),
		makeRC("if", "flag.x"),
		makeRC("music", "theme"),
		makeRC("end_if", ""),
		makeRC("end_parallel", ""),
	})
	seq := cut.ParseSequenceTree(cmds)
	if len(seq.Steps) != 1 {
		t.Fatalf("expected 1 top-level step, got %d", len(seq.Steps))
	}
	parallel := seq.Steps[0]
	if parallel.Body == nil || len(parallel.Body.Steps) != 1 {
		t.Fatalf("parallel body should have 1 step (the if block)")
	}
	ifStep := parallel.Body.Steps[0]
	if ifStep.Body == nil || len(ifStep.Body.Steps) != 1 {
		t.Errorf("if body should have 1 step")
	}
}
