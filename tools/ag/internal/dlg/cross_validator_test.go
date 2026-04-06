package dlg_test

import (
	"testing"

	"github.com/ags3d/ag/internal/dlg"
)

// mustCrossValidate parses & links src files, then runs ValidateCrossSystem
// with sym. Returns all cross-validation errors.
func mustCrossValidate(t *testing.T, sym dlg.ProjectSymbolTable, srcs ...string) []dlg.CrossValidationError {
	t.Helper()
	files := mustParseFiles(t, srcs...)
	lp, err := dlg.Link(files)
	if err != nil {
		t.Fatalf("Link error: %v", err)
	}
	return dlg.ValidateCrossSystem(lp, sym)
}

func hasCrossCode(errs []dlg.CrossValidationError, code string) bool {
	for _, e := range errs {
		if e.Code == code {
			return true
		}
	}
	return false
}

// ─── DLG-E020: inventory item not defined ───────────────────────────────────

func TestCross_E020_ItemMissing(t *testing.T) {
	sym := dlg.ProjectSymbolTable{
		ItemNames: map[string]bool{"gate_pass": true},
	}
	src := "title: n\n---\n-> Take it <<visible_if item.rusty_key in player.inventory>>\n   <<end>>\n===\n"
	errs := mustCrossValidate(t, sym, src)
	if !hasCrossCode(errs, "DLG-E020") {
		t.Error("expected DLG-E020 for unknown item 'rusty_key', none found")
	}
}

func TestCross_E020_ItemPresent_NoError(t *testing.T) {
	sym := dlg.ProjectSymbolTable{
		ItemNames: map[string]bool{"rusty_key": true},
	}
	src := "title: n\n---\n-> Take it <<visible_if item.rusty_key in player.inventory>>\n   <<end>>\n===\n"
	errs := mustCrossValidate(t, sym, src)
	if hasCrossCode(errs, "DLG-E020") {
		t.Error("unexpected DLG-E020 for known item 'rusty_key'")
	}
}

func TestCross_E020_EmptySymTable_NoError(t *testing.T) {
	// Empty symbol table = project not scanned — skip cross-system checks.
	sym := dlg.ProjectSymbolTable{}
	src := "title: n\n---\n-> Go <<visible_if item.ghost_item in player.inventory>>\n   <<end>>\n===\n"
	errs := mustCrossValidate(t, sym, src)
	if hasCrossCode(errs, "DLG-E020") {
		t.Error("should skip DLG-E020 when ItemNames is nil/empty")
	}
}

// ─── DLG-E021: room not defined ─────────────────────────────────────────────

func TestCross_E021_RoomMissing(t *testing.T) {
	sym := dlg.ProjectSymbolTable{
		RoomNames: map[string]bool{"market": true},
	}
	src := "title: n\n---\n<<action room.transition(\"inner_courtyard\")>>\n<<end>>\n===\n"
	errs := mustCrossValidate(t, sym, src)
	if !hasCrossCode(errs, "DLG-E021") {
		t.Error("expected DLG-E021 for unknown room 'inner_courtyard'")
	}
}

func TestCross_E021_RoomPresent_NoError(t *testing.T) {
	sym := dlg.ProjectSymbolTable{
		RoomNames: map[string]bool{"inner_courtyard": true},
	}
	src := "title: n\n---\n<<action room.transition(\"inner_courtyard\")>>\n<<end>>\n===\n"
	errs := mustCrossValidate(t, sym, src)
	if hasCrossCode(errs, "DLG-E021") {
		t.Error("unexpected DLG-E021 for known room 'inner_courtyard'")
	}
}

// ─── DLG-E022: character not defined ────────────────────────────────────────

func TestCross_E022_CharMissing(t *testing.T) {
	sym := dlg.ProjectSymbolTable{
		CharacterNames: map[string]bool{"guard": true},
	}
	src := "title: n\n---\n-> Ask <<visible_if char.elara.present>>\n   <<end>>\n===\n"
	errs := mustCrossValidate(t, sym, src)
	if !hasCrossCode(errs, "DLG-E022") {
		t.Error("expected DLG-E022 for unknown char 'elara'")
	}
}

func TestCross_E022_CharPresent_NoError(t *testing.T) {
	sym := dlg.ProjectSymbolTable{
		CharacterNames: map[string]bool{"elara": true},
	}
	src := "title: n\n---\n-> Ask <<visible_if char.elara.present>>\n   <<end>>\n===\n"
	errs := mustCrossValidate(t, sym, src)
	if hasCrossCode(errs, "DLG-E022") {
		t.Error("unexpected DLG-E022 for known char 'elara'")
	}
}

// ─── DLG-E023: flag never set ───────────────────────────────────────────────

func TestCross_E023_FlagNeverSet(t *testing.T) {
	sym := dlg.ProjectSymbolTable{
		FlagsEverSet: map[string]bool{"gate_open": true},
	}
	src := "title: n\n---\n-> Go <<visible_if flag.guard_spoken>>\n   <<end>>\n===\n"
	errs := mustCrossValidate(t, sym, src)
	if !hasCrossCode(errs, "DLG-E023") {
		t.Error("expected DLG-E023 for flag 'guard_spoken' never set")
	}
}

func TestCross_E023_FlagIsSet_NoError(t *testing.T) {
	sym := dlg.ProjectSymbolTable{
		FlagsEverSet: map[string]bool{"guard_spoken": true},
	}
	src := "title: n\n---\n-> Go <<visible_if flag.guard_spoken>>\n   <<end>>\n===\n"
	errs := mustCrossValidate(t, sym, src)
	if hasCrossCode(errs, "DLG-E023") {
		t.Error("unexpected DLG-E023 for flag 'guard_spoken' that is set")
	}
}

// ─── DLG-E024: named point not in room ──────────────────────────────────────

func TestCross_E024_PointMissingFromRoom(t *testing.T) {
	sym := dlg.ProjectSymbolTable{
		RoomPoints: map[string]map[string]bool{
			"market": {"stall_left": true},
		},
	}
	src := "title: n\n---\n<<action char.guard.walk_to(room.point(\"market\", \"fountain\"))>>\n<<end>>\n===\n"
	errs := mustCrossValidate(t, sym, src)
	if !hasCrossCode(errs, "DLG-E024") {
		t.Error("expected DLG-E024 for point 'fountain' not in room 'market'")
	}
}

func TestCross_E024_PointPresent_NoError(t *testing.T) {
	sym := dlg.ProjectSymbolTable{
		RoomPoints: map[string]map[string]bool{
			"market": {"stall_left": true, "fountain": true},
		},
	}
	src := "title: n\n---\n<<action char.guard.walk_to(room.point(\"market\", \"fountain\"))>>\n<<end>>\n===\n"
	errs := mustCrossValidate(t, sym, src)
	if hasCrossCode(errs, "DLG-E024") {
		t.Error("unexpected DLG-E024 for known point 'fountain' in room 'market'")
	}
}

// ─── DLG-E025: knowledge flag never granted ─────────────────────────────────

func TestCross_E025_KnowledgeNeverGranted(t *testing.T) {
	sym := dlg.ProjectSymbolTable{
		KnowledgeFlags: map[string]bool{"guard_secret": true},
	}
	src := "title: n\n---\n-> Reveal <<visible_if knowledge.hidden_passage>>\n   <<end>>\n===\n"
	errs := mustCrossValidate(t, sym, src)
	if !hasCrossCode(errs, "DLG-E025") {
		t.Error("expected DLG-E025 for knowledge 'hidden_passage' never granted")
	}
}

func TestCross_E025_KnowledgeGranted_NoError(t *testing.T) {
	sym := dlg.ProjectSymbolTable{
		KnowledgeFlags: map[string]bool{"hidden_passage": true},
	}
	src := "title: n\n---\n-> Reveal <<visible_if knowledge.hidden_passage>>\n   <<end>>\n===\n"
	errs := mustCrossValidate(t, sym, src)
	if hasCrossCode(errs, "DLG-E025") {
		t.Error("unexpected DLG-E025 for knowledge 'hidden_passage' that is granted")
	}
}

// ─── Clean project — no errors ──────────────────────────────────────────────

func TestCross_CleanProject_NoErrors(t *testing.T) {
	sym := dlg.ProjectSymbolTable{
		CharacterNames: map[string]bool{"guard": true, "player": true},
		RoomNames:      map[string]bool{"market": true, "courtyard": true},
		ItemNames:      map[string]bool{"gate_pass": true},
		FlagsEverSet:   map[string]bool{"guard_spoken": true, "gate_open": true},
		KnowledgeFlags: map[string]bool{},
	}
	src := "title: greet\ncharacter: guard\n---\nGuard: Halt.\n" +
		"-> I have a pass <<visible_if item.gate_pass in player.inventory>>\n" +
		"   Guard: Very well.\n" +
		"   <<action flag.guard_spoken = true>>\n" +
		"   <<action room.transition(\"courtyard\")>>\n" +
		"===\n"
	errs := mustCrossValidate(t, sym, src)
	if len(errs) > 0 {
		t.Errorf("expected no errors for clean project, got: %v", errs)
	}
}
