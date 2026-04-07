package dlg_test

import (
	"testing"

	"github.com/ags3d/ag/internal/dlg"
)

func mustParse(t *testing.T, src string) *dlg.DialogueFile {
	t.Helper()
	df, err := dlg.Parse("test.agdlg", src)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	return df
}

// --- Header parsing ---

func TestParse_TitleRequired(t *testing.T) {
	_, err := dlg.Parse("test.agdlg", "character: guard\n---\n===")
	if err == nil {
		t.Fatal("expected error for node without title, got nil")
	}
}

func TestParse_NodeTitle(t *testing.T) {
	df := mustParse(t, "title: guard_greeting\n---\n===")
	if len(df.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(df.Nodes))
	}
	if df.Nodes[0].Title != "guard_greeting" {
		t.Errorf("Title = %q, want guard_greeting", df.Nodes[0].Title)
	}
}

func TestParse_HeaderFields(t *testing.T) {
	src := "title: intro\ncharacter: guard\ntags: [chapter:1, cinematic]\ninherits: global_options\nsuppress: farewell\nloc_id: guard_intro\n---\n==="
	df := mustParse(t, src)
	n := df.Nodes[0]

	if n.Character != "guard" {
		t.Errorf("Character = %q, want guard", n.Character)
	}
	if len(n.Tags) != 2 || n.Tags[0] != "chapter:1" || n.Tags[1] != "cinematic" {
		t.Errorf("Tags = %v, want [chapter:1 cinematic]", n.Tags)
	}
	if len(n.Inherits) != 1 || n.Inherits[0] != "global_options" {
		t.Errorf("Inherits = %v, want [global_options]", n.Inherits)
	}
	if len(n.Suppress) != 1 || n.Suppress[0] != "farewell" {
		t.Errorf("Suppress = %v, want [farewell]", n.Suppress)
	}
	if n.LocID != "guard_intro" {
		t.Errorf("LocID = %q, want guard_intro", n.LocID)
	}
}

// --- Body: speaker lines ---

func TestParse_SpeakerLine(t *testing.T) {
	df := mustParse(t, "title: t\n---\nGuard: You there. Stop.\n===")
	body := df.Nodes[0].Body
	if len(body) != 1 {
		t.Fatalf("body len = %d, want 1", len(body))
	}
	sl, ok := body[0].(*dlg.SpeakerLine)
	if !ok {
		t.Fatalf("body[0] type = %T, want *SpeakerLine", body[0])
	}
	if sl.Speaker != "Guard" {
		t.Errorf("Speaker = %q, want Guard", sl.Speaker)
	}
	if sl.Text != "You there. Stop." {
		t.Errorf("Text = %q, want 'You there. Stop.'", sl.Text)
	}
}

func TestParse_SpeakerLineWithInlineCommand(t *testing.T) {
	df := mustParse(t, "title: t\n---\nGuard: Hello. <<action flag.met = true>>\n===")
	sl := df.Nodes[0].Body[0].(*dlg.SpeakerLine)
	if len(sl.Commands) != 1 {
		t.Fatalf("Commands len = %d, want 1", len(sl.Commands))
	}
	if sl.Commands[0].Raw != "action flag.met = true" {
		t.Errorf("Command.Raw = %q, want 'action flag.met = true'", sl.Commands[0].Raw)
	}
}

func TestParse_SpeakerLineWithLocKey(t *testing.T) {
	df := mustParse(t, "title: t\n---\nGuard: Hi. #loc:guard_hi_001\n===")
	sl := df.Nodes[0].Body[0].(*dlg.SpeakerLine)
	if sl.LocKey != "guard_hi_001" {
		t.Errorf("LocKey = %q, want guard_hi_001", sl.LocKey)
	}
}

// --- Body: narration lines ---

func TestParse_NarrationLine(t *testing.T) {
	df := mustParse(t, "title: t\n---\nOnce upon a time.\n===")
	nl, ok := df.Nodes[0].Body[0].(*dlg.NarrationLine)
	if !ok {
		t.Fatalf("body[0] type = %T, want *NarrationLine", df.Nodes[0].Body[0])
	}
	if nl.Text != "Once upon a time." {
		t.Errorf("Text = %q", nl.Text)
	}
}

// --- Body: standalone commands ---

func TestParse_StandaloneCommand(t *testing.T) {
	df := mustParse(t, "title: t\n---\n<<action flag.gate_open = true>>\n===")
	cs, ok := df.Nodes[0].Body[0].(*dlg.CommandStatement)
	if !ok {
		t.Fatalf("body[0] type = %T, want *CommandStatement", df.Nodes[0].Body[0])
	}
	if cs.Raw != "action flag.gate_open = true" {
		t.Errorf("Raw = %q", cs.Raw)
	}
}

func TestParse_EndCommand(t *testing.T) {
	df := mustParse(t, "title: t\n---\n<<end>>\n===")
	cs := df.Nodes[0].Body[0].(*dlg.CommandStatement)
	if cs.Raw != "end" {
		t.Errorf("Raw = %q, want end", cs.Raw)
	}
}

func TestParse_JumpCommand(t *testing.T) {
	df := mustParse(t, "title: t\n---\n<<jump other_node>>\n===")
	cs := df.Nodes[0].Body[0].(*dlg.CommandStatement)
	if cs.Raw != "jump other_node" {
		t.Errorf("Raw = %q, want 'jump other_node'", cs.Raw)
	}
}

// --- Body: options ---

func TestParse_OptionLine(t *testing.T) {
	df := mustParse(t, "title: t\n---\n-> Just passing through.\n===")
	ob, ok := df.Nodes[0].Body[0].(*dlg.OptionBranch)
	if !ok {
		t.Fatalf("body[0] type = %T, want *OptionBranch", df.Nodes[0].Body[0])
	}
	if ob.Text != "Just passing through." {
		t.Errorf("Text = %q", ob.Text)
	}
	if ob.Depth != 0 {
		t.Errorf("Depth = %d, want 0", ob.Depth)
	}
}

func TestParse_OptionWithCondition(t *testing.T) {
	df := mustParse(t, "title: t\n---\n-> Pass. <<visible_if not flag.x>>\n===")
	ob := df.Nodes[0].Body[0].(*dlg.OptionBranch)
	if len(ob.Commands) != 1 || ob.Commands[0].Raw != "visible_if not flag.x" {
		t.Errorf("Commands = %v", ob.Commands)
	}
}

func TestParse_OptionWithBody(t *testing.T) {
	src := "title: t\n---\n-> Option A.\n   Guard: Response.\n   <<end>>\n==="
	df := mustParse(t, src)
	ob := df.Nodes[0].Body[0].(*dlg.OptionBranch)
	if len(ob.Body) != 2 {
		t.Fatalf("option body len = %d, want 2", len(ob.Body))
	}
	_, ok1 := ob.Body[0].(*dlg.SpeakerLine)
	_, ok2 := ob.Body[1].(*dlg.CommandStatement)
	if !ok1 || !ok2 {
		t.Errorf("option body types: got %T, %T", ob.Body[0], ob.Body[1])
	}
}

func TestParse_MultipleTopLevelOptions(t *testing.T) {
	src := "title: t\n---\n-> Option A.\n   <<end>>\n-> Option B.\n   <<end>>\n==="
	df := mustParse(t, src)
	body := df.Nodes[0].Body
	if len(body) != 2 {
		t.Fatalf("body len = %d, want 2", len(body))
	}
	a := body[0].(*dlg.OptionBranch)
	b := body[1].(*dlg.OptionBranch)
	if a.Text != "Option A." || b.Text != "Option B." {
		t.Errorf("options: %q, %q", a.Text, b.Text)
	}
}

// --- Multi-node file ---

func TestParse_TwoNodes(t *testing.T) {
	src := "title: node_a\n---\nGuard: Hello.\n===\ntitle: node_b\n---\nGuard: Goodbye.\n==="
	df := mustParse(t, src)
	if len(df.Nodes) != 2 {
		t.Fatalf("nodes = %d, want 2", len(df.Nodes))
	}
	if df.Nodes[0].Title != "node_a" || df.Nodes[1].Title != "node_b" {
		t.Errorf("titles: %q, %q", df.Nodes[0].Title, df.Nodes[1].Title)
	}
}

// --- Full spec example ---

func TestParse_FullExample(t *testing.T) {
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
===

title: guard_checks_pass
character: guard
---
Guard: Hmm. This looks legitimate.
Guard: You can go through.
<<action flag.gate_open = true>>
<<action room.transition("inner_courtyard")>>
===`

	df := mustParse(t, src)
	if len(df.Nodes) != 2 {
		t.Fatalf("nodes = %d, want 2", len(df.Nodes))
	}

	n0 := df.Nodes[0]
	if n0.Title != "guard_greeting" {
		t.Errorf("node0 title = %q", n0.Title)
	}
	if n0.Character != "guard" {
		t.Errorf("node0 character = %q", n0.Character)
	}

	// First body item should be a SpeakerLine.
	if _, ok := n0.Body[0].(*dlg.SpeakerLine); !ok {
		t.Errorf("body[0] type = %T, want *SpeakerLine", n0.Body[0])
	}

	// Next three body items should be OptionBranches.
	for i := 1; i <= 3; i++ {
		if _, ok := n0.Body[i].(*dlg.OptionBranch); !ok {
			t.Errorf("body[%d] type = %T, want *OptionBranch", i, n0.Body[i])
		}
	}

	// Second node.
	n1 := df.Nodes[1]
	if n1.Title != "guard_checks_pass" {
		t.Errorf("node1 title = %q", n1.Title)
	}
	// Should have 4 statements: 2 SpeakerLines + 2 CommandStatements.
	if len(n1.Body) != 4 {
		t.Errorf("node1 body len = %d, want 4", len(n1.Body))
	}
}

// --- Source positions ---

func TestParse_NodePosSet(t *testing.T) {
	df := mustParse(t, "title: t\n---\n===")
	if df.Nodes[0].Pos.File != "test.agdlg" {
		t.Errorf("Pos.File = %q, want test.agdlg", df.Nodes[0].Pos.File)
	}
	if df.Nodes[0].Pos.Line < 1 {
		t.Errorf("Pos.Line = %d, want >= 1", df.Nodes[0].Pos.Line)
	}
}

// --- Error cases ---

func TestParse_MissingSeparator(t *testing.T) {
	_, err := dlg.Parse("test.agdlg", "title: t\n===")
	if err == nil {
		t.Fatal("expected error for missing ---, got nil")
	}
}

func TestParse_EmptyFile(t *testing.T) {
	df := mustParse(t, "")
	if len(df.Nodes) != 0 {
		t.Errorf("expected 0 nodes, got %d", len(df.Nodes))
	}
}

// --- Inline cutscene blocks (T-CUT03) ---

func TestParse_InlineCutscene_Basic(t *testing.T) {
	src := "title: t\n---\n<<cutscene>>\n<<wait duration:1.0>>\n<<end_cutscene>>\n==="
	df := mustParse(t, src)
	body := df.Nodes[0].Body
	if len(body) != 1 {
		t.Fatalf("body len = %d, want 1", len(body))
	}
	block, ok := body[0].(*dlg.InlineCutsceneBlock)
	if !ok {
		t.Fatalf("body[0] type = %T, want *dlg.InlineCutsceneBlock", body[0])
	}
	if block.SkipPolicy != "" {
		t.Errorf("SkipPolicy = %q, want empty", block.SkipPolicy)
	}
	if block.Sequence == nil {
		t.Fatal("Sequence is nil")
	}
	if len(block.Sequence.Steps) != 1 {
		t.Errorf("Sequence.Steps len = %d, want 1", len(block.Sequence.Steps))
	}
	if block.Sequence.Steps[0].Cmd.Name != "wait" {
		t.Errorf("step name = %q, want wait", block.Sequence.Steps[0].Cmd.Name)
	}
}

func TestParse_InlineCutscene_SkipPolicy(t *testing.T) {
	src := "title: t\n---\n<<cutscene skip:after_first_view>>\n<<end_cutscene>>\n==="
	df := mustParse(t, src)
	block, ok := df.Nodes[0].Body[0].(*dlg.InlineCutsceneBlock)
	if !ok {
		t.Fatalf("expected *dlg.InlineCutsceneBlock, got %T", df.Nodes[0].Body[0])
	}
	if block.SkipPolicy != "after_first_view" {
		t.Errorf("SkipPolicy = %q, want after_first_view", block.SkipPolicy)
	}
}

func TestParse_InlineCutscene_NestedCommands(t *testing.T) {
	src := "title: t\n---\n<<cutscene>>\n<<fade_in duration:2.0>>\n<<wait duration:0.5>>\n<<end>>\n<<end_cutscene>>\n==="
	df := mustParse(t, src)
	block, ok := df.Nodes[0].Body[0].(*dlg.InlineCutsceneBlock)
	if !ok {
		t.Fatalf("expected *dlg.InlineCutsceneBlock, got %T", df.Nodes[0].Body[0])
	}
	if len(block.Sequence.Steps) != 3 {
		t.Errorf("Sequence.Steps len = %d, want 3", len(block.Sequence.Steps))
	}
	names := []string{"fade_in", "wait", "end"}
	for i, want := range names {
		if block.Sequence.Steps[i].Cmd.Name != want {
			t.Errorf("step[%d] name = %q, want %q", i, block.Sequence.Steps[i].Cmd.Name, want)
		}
	}
}

func TestParse_InlineCutscene_UnclosedError(t *testing.T) {
	_, err := dlg.Parse("test.agdlg", "title: t\n---\n<<cutscene>>\n<<wait duration:1.0>>\n===")
	if err == nil {
		t.Fatal("expected error for unclosed <<cutscene>> block, got nil")
	}
}

func TestParse_InlineCutscene_SurroundedByOtherStmts(t *testing.T) {
	src := "title: t\n---\nBefore text.\n<<cutscene skip:never>>\n<<wait duration:0.1>>\n<<end_cutscene>>\nAfter text.\n==="
	df := mustParse(t, src)
	body := df.Nodes[0].Body
	if len(body) != 3 {
		t.Fatalf("body len = %d, want 3", len(body))
	}
	if _, ok := body[0].(*dlg.NarrationLine); !ok {
		t.Errorf("body[0] = %T, want *dlg.NarrationLine", body[0])
	}
	if _, ok := body[1].(*dlg.InlineCutsceneBlock); !ok {
		t.Errorf("body[1] = %T, want *dlg.InlineCutsceneBlock", body[1])
	}
	if _, ok := body[2].(*dlg.NarrationLine); !ok {
		t.Errorf("body[2] = %T, want *dlg.NarrationLine", body[2])
	}
}
