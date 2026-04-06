package cut_test

import (
	"testing"

	"github.com/ags3d/ag/internal/cut"
)

func seq(t *testing.T, src string) []cut.ValidationError {
	t.Helper()
	cf := mustParse(t, src)
	return cut.ValidateSequence(cf)
}

func hasSeqCode(errs []cut.ValidationError, code string) bool {
	return hasErrCode(errs, code) // reuse helper from validator_test.go
}

// --- SEQ-E001: <<sync>> references undeclared id ---

func TestSeq_SyncUndeclaredIDIsE001(t *testing.T) {
	errs := seq(t, "title: t\nsequence:\n<<fade_in duration:1.0 bg:fade_bg>>\n<<sync ghost_step>>\n<<end>>\n")
	if !hasSeqCode(errs, "SEQ-E001") {
		t.Error("expected SEQ-E001 for sync referencing undeclared id, none found")
	}
}

func TestSeq_SyncDeclaredBGNoE001(t *testing.T) {
	errs := seq(t, "title: t\nsequence:\n<<fade_in duration:1.0 bg:fade_bg>>\n<<sync fade_bg>>\n<<end>>\n")
	for _, e := range errs {
		if e.Code == "SEQ-E001" {
			t.Errorf("unexpected SEQ-E001: %v", e)
		}
	}
}

// --- SEQ-E002: <<sync>> references foreground id: ---

func TestSeq_SyncForegroundIDIsE002(t *testing.T) {
	errs := seq(t, "title: t\nsequence:\n<<fade_in duration:1.0 id:main_step>>\n<<sync main_step>>\n<<end>>\n")
	if !hasSeqCode(errs, "SEQ-E002") {
		t.Error("expected SEQ-E002 for sync referencing foreground id, none found")
	}
}

func TestSeq_SyncBGNotFGNoE002(t *testing.T) {
	errs := seq(t, "title: t\nsequence:\n<<music theme_main fade_in:2.0 bg:music_bg>>\n<<sync music_bg>>\n<<end>>\n")
	for _, e := range errs {
		if e.Code == "SEQ-E002" {
			t.Errorf("unexpected SEQ-E002: %v", e)
		}
	}
}

// --- SEQ-E003: bg step never synced ---

func TestSeq_BGStepNeverSyncedIsE003(t *testing.T) {
	errs := seq(t, "title: t\nsequence:\n<<music theme_main fade_in:3.0 bg:music_bg>>\n<<fade_in duration:1.0>>\n<<end>>\n")
	if !hasSeqCode(errs, "SEQ-E003") {
		t.Error("expected SEQ-E003 for bg step with no sync, none found")
	}
}

func TestSeq_BGStepSyncedNoE003(t *testing.T) {
	errs := seq(t, "title: t\nsequence:\n<<music theme_main fade_in:3.0 bg:music_bg>>\n<<sync music_bg>>\n<<end>>\n")
	for _, e := range errs {
		if e.Code == "SEQ-E003" {
			t.Errorf("unexpected SEQ-E003: %v", e)
		}
	}
}

func TestSeq_SyncAllCoversBGNoE003(t *testing.T) {
	// <<sync>> with no args covers all bg steps.
	errs := seq(t, "title: t\nsequence:\n<<music theme_main fade_in:3.0 bg:music_bg>>\n<<sync>>\n<<end>>\n")
	for _, e := range errs {
		if e.Code == "SEQ-E003" {
			t.Errorf("unexpected SEQ-E003 after <<sync>> all: %v", e)
		}
	}
}

func TestSeq_NoBGStepsNoE003(t *testing.T) {
	errs := seq(t, "title: t\nsequence:\n<<fade_in duration:1.0>>\n<<end>>\n")
	for _, e := range errs {
		if e.Code == "SEQ-E003" {
			t.Errorf("unexpected SEQ-E003 with no bg steps: %v", e)
		}
	}
}

// --- SEQ-E004: on_fail:jump_to references missing label ---

func TestSeq_OnFailMissingLabelIsE004(t *testing.T) {
	errs := seq(t, "title: t\nsequence:\n<<fade_in duration:1.0>>\n<<character player walk_to point.x timeout:5.0 on_fail:jump_to:fallback>>\n<<end>>\n")
	if !hasSeqCode(errs, "SEQ-E004") {
		t.Error("expected SEQ-E004 for on_fail jump_to missing label, none found")
	}
}

func TestSeq_OnFailExistingLabelNoE004(t *testing.T) {
	errs := seq(t, "title: t\nsequence:\n<<character player walk_to point.x timeout:5.0 on_fail:jump_to:fallback>>\n<<label fallback>>\n<<end>>\n")
	for _, e := range errs {
		if e.Code == "SEQ-E004" {
			t.Errorf("unexpected SEQ-E004: %v", e)
		}
	}
}

// --- SEQ-E007: duplicate step id ---

func TestSeq_DuplicateBGIDIsE007(t *testing.T) {
	errs := seq(t, "title: t\nsequence:\n<<fade_in duration:1.0 bg:step_a>>\n<<music theme_main fade_in:2.0 bg:step_a>>\n<<sync step_a>>\n<<end>>\n")
	if !hasSeqCode(errs, "SEQ-E007") {
		t.Error("expected SEQ-E007 for duplicate bg id, none found")
	}
}

func TestSeq_DuplicateFGIDIsE007(t *testing.T) {
	errs := seq(t, "title: t\nsequence:\n<<fade_in duration:1.0 id:step_a>>\n<<fade_out duration:1.0 id:step_a>>\n<<end>>\n")
	if !hasSeqCode(errs, "SEQ-E007") {
		t.Error("expected SEQ-E007 for duplicate fg id, none found")
	}
}

func TestSeq_UniqueIDsNoE007(t *testing.T) {
	errs := seq(t, "title: t\nsequence:\n<<fade_in duration:1.0 bg:step_a>>\n<<music theme_main fade_in:2.0 bg:step_b>>\n<<sync step_a>>\n<<sync step_b>>\n<<end>>\n")
	for _, e := range errs {
		if e.Code == "SEQ-E007" {
			t.Errorf("unexpected SEQ-E007: %v", e)
		}
	}
}

// --- Clean sequence: no errors ---

func TestSeq_CleanSequenceNoErrors(t *testing.T) {
	src := `title: t
sequence:
<<fade_in duration:2.0 bg:fade_bg>>
<<music theme_main fade_in:3.0 bg:music_bg>>
<<label act_start>>
<<character player walk_to point.entrance timeout:10.0 on_fail:jump_to:act_start>>
<<sync fade_bg>>
<<sync music_bg>>
<<fade_out duration:1.0>>
<<end>>
`
	errs := seq(t, src)
	if len(errs) != 0 {
		t.Errorf("expected 0 SEQ errors for valid sequence, got: %v", errs)
	}
}
