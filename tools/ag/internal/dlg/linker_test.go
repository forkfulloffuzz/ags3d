package dlg_test

import (
	"testing"

	"github.com/ags3d/ag/internal/dlg"
)

func mustParseFiles(t *testing.T, srcs ...string) []*dlg.DialogueFile {
	t.Helper()
	var files []*dlg.DialogueFile
	for i, src := range srcs {
		filename := "test_" + string(rune('a'+i)) + ".agdlg"
		df, err := dlg.Parse(filename, src)
		if err != nil {
			t.Fatalf("Parse(%s) error: %v", filename, err)
		}
		files = append(files, df)
	}
	return files
}

func mustLink(t *testing.T, srcs ...string) *dlg.LinkedProject {
	t.Helper()
	files := mustParseFiles(t, srcs...)
	lp, err := dlg.Link(files)
	if err != nil {
		t.Fatalf("Link error: %v", err)
	}
	return lp
}

// --- Title index ---

func TestLink_BuildsNodeIndex(t *testing.T) {
	lp := mustLink(t, "title: node_a\n---\n===\n")
	if _, ok := lp.NodesByTitle["node_a"]; !ok {
		t.Error("node_a not in index")
	}
}

func TestLink_MultiFileIndex(t *testing.T) {
	lp := mustLink(t,
		"title: node_a\n---\n===",
		"title: node_b\n---\n===",
	)
	if len(lp.NodesByTitle) != 2 {
		t.Errorf("index size = %d, want 2", len(lp.NodesByTitle))
	}
}

func TestLink_DuplicateTitleIsHardError(t *testing.T) {
	files := mustParseFiles(t,
		"title: shared\n---\n===",
		"title: shared\n---\n===",
	)
	_, err := dlg.Link(files)
	if err == nil {
		t.Fatal("expected error for duplicate title, got nil")
	}
}

func TestLink_LinkedNodeHasFileRef(t *testing.T) {
	lp := mustLink(t, "title: n\n---\n===")
	ln := lp.NodesByTitle["n"]
	if ln.File == nil {
		t.Error("LinkedNode.File is nil")
	}
}

// --- Jump resolution ---

func TestLink_ResolvedJumpNoError(t *testing.T) {
	lp := mustLink(t, "title: a\n---\n<<jump b>>\n===\ntitle: b\n---\n===")
	for _, e := range lp.LinkErrors {
		if e.Code == "DLG-E002" {
			t.Errorf("unexpected DLG-E002: %v", e)
		}
	}
}

func TestLink_UnresolvedJumpRecordsError(t *testing.T) {
	lp := mustLink(t, "title: a\n---\n<<jump missing_target>>\n===")
	found := false
	for _, e := range lp.LinkErrors {
		if e.Code == "DLG-E002" && e.Msg != "" {
			found = true
		}
	}
	if !found {
		t.Error("expected DLG-E002 for unresolved jump, none found")
	}
}

func TestLink_JumpInsideOptionBody(t *testing.T) {
	// Jump inside an option body to a missing node.
	src := "title: a\n---\n-> Option.\n   <<jump nowhere>>\n   <<end>>\n==="
	lp := mustLink(t, src)
	found := false
	for _, e := range lp.LinkErrors {
		if e.Code == "DLG-E002" {
			found = true
		}
	}
	if !found {
		t.Error("expected DLG-E002 for jump inside option body, none found")
	}
}

func TestLink_PlaceholderJumpSkipped(t *testing.T) {
	// $character.node placeholder — should not produce DLG-E002 at link stage.
	lp := mustLink(t, "title: a\n---\n<<jump $guard.greeting>>\n===")
	for _, e := range lp.LinkErrors {
		if e.Code == "DLG-E002" {
			t.Errorf("placeholder jump produced unexpected DLG-E002: %v", e)
		}
	}
}

func TestLink_JumpAcrossFiles(t *testing.T) {
	// Jump from file A to node declared in file B — should resolve cleanly.
	lp := mustLink(t,
		"title: node_a\n---\n<<jump node_b>>\n===",
		"title: node_b\n---\n===",
	)
	for _, e := range lp.LinkErrors {
		if e.Code == "DLG-E002" {
			t.Errorf("cross-file jump produced unexpected DLG-E002: %v", e)
		}
	}
}

// --- Inherits / suppress ---

func TestLink_InheritsExistingNode(t *testing.T) {
	lp := mustLink(t,
		"title: global_farewell\ntags: [global]\n---\n===",
		"title: guard_greet\ninherits: global_farewell\n---\n===",
	)
	for _, e := range lp.LinkErrors {
		t.Errorf("unexpected link error: %v", e)
	}
}

func TestLink_InheritsMissingNode(t *testing.T) {
	lp := mustLink(t, "title: a\ninherits: nonexistent\n---\n===")
	found := false
	for _, e := range lp.LinkErrors {
		if e.Code == "DLG-E002" {
			found = true
		}
	}
	if !found {
		t.Error("expected DLG-E002 for missing inherits target, none found")
	}
}

func TestLink_SuppressGlobalNode(t *testing.T) {
	lp := mustLink(t,
		"title: global_farewell\ntags: [global]\n---\n===",
		"title: guard_greet\nsuppress: global_farewell\n---\n===",
	)
	for _, e := range lp.LinkErrors {
		t.Errorf("unexpected link error: %v", e)
	}
}

func TestLink_SuppressNonGlobalNodeIsError(t *testing.T) {
	lp := mustLink(t,
		"title: regular_node\n---\n===",
		"title: a\nsuppress: regular_node\n---\n===",
	)
	found := false
	for _, e := range lp.LinkErrors {
		if e.Code == "DLG-E006" {
			found = true
		}
	}
	if !found {
		t.Error("expected DLG-E006 for suppress of non-global node, none found")
	}
}

func TestLink_SuppressMissingNodeIsError(t *testing.T) {
	lp := mustLink(t, "title: a\nsuppress: ghost\n---\n===")
	found := false
	for _, e := range lp.LinkErrors {
		if e.Code == "DLG-E002" {
			found = true
		}
	}
	if !found {
		t.Error("expected DLG-E002 for suppress of missing node, none found")
	}
}

// --- Empty project ---

func TestLink_EmptyProject(t *testing.T) {
	lp, err := dlg.Link(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lp.NodesByTitle) != 0 {
		t.Errorf("expected empty index, got %d entries", len(lp.NodesByTitle))
	}
}
