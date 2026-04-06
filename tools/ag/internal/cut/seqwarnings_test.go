package cut_test

import (
	"testing"

	"github.com/ags3d/ag/internal/cut"
)

func seqWarn(t *testing.T, src string) []cut.ValidationWarning {
	t.Helper()
	cf := mustParse(t, src)
	return cut.WarnSequence(cf)
}

func hasSeqWarnCode(warns []cut.ValidationWarning, code string) bool {
	for _, w := range warns {
		if w.Code == code {
			return true
		}
	}
	return false
}

// --- SEQ-W001: bg step with duration > 10 s ---

func TestSeqW001_LongBgStep(t *testing.T) {
	src := "title: t\nsequence:\n<<fade_in duration:15.0 bg:long>>\n<<sync long>>\n<<end>>\n"
	warns := seqWarn(t, src)
	if !hasSeqWarnCode(warns, "SEQ-W001") {
		t.Error("expected SEQ-W001 for bg step with duration 15s")
	}
}

func TestSeqW001_ShortBgNoWarn(t *testing.T) {
	src := "title: t\nsequence:\n<<fade_in duration:2.0 bg:short>>\n<<sync short>>\n<<end>>\n"
	warns := seqWarn(t, src)
	if hasSeqWarnCode(warns, "SEQ-W001") {
		t.Error("SEQ-W001 should not fire for bg step with duration 2s")
	}
}

func TestSeqW001_BgNoDurationNoWarn(t *testing.T) {
	src := "title: t\nsequence:\n<<fade_in bg:nodur>>\n<<sync nodur>>\n<<end>>\n"
	warns := seqWarn(t, src)
	if hasSeqWarnCode(warns, "SEQ-W001") {
		t.Error("SEQ-W001 should not fire when no duration param")
	}
}

// --- SEQ-W002: long-running foreground step with no timeout ---

func TestSeqW002_WalkToNoTimeout(t *testing.T) {
	src := "title: t\nsequence:\n<<walk_to player point.door>>\n<<end>>\n"
	warns := seqWarn(t, src)
	if !hasSeqWarnCode(warns, "SEQ-W002") {
		t.Error("expected SEQ-W002 for walk_to with no timeout")
	}
}

func TestSeqW002_WalkToWithTimeout(t *testing.T) {
	src := "title: t\nsequence:\n<<walk_to player point.door timeout:5.0>>\n<<end>>\n"
	warns := seqWarn(t, src)
	if hasSeqWarnCode(warns, "SEQ-W002") {
		t.Error("SEQ-W002 should not fire when timeout is set")
	}
}

func TestSeqW002_RunToNoTimeout(t *testing.T) {
	src := "title: t\nsequence:\n<<run_to player point.door>>\n<<end>>\n"
	warns := seqWarn(t, src)
	if !hasSeqWarnCode(warns, "SEQ-W002") {
		t.Error("expected SEQ-W002 for run_to with no timeout")
	}
}

func TestSeqW002_CameraMoveToNoTimeout(t *testing.T) {
	src := "title: t\nsequence:\n<<camera move_to point.wide duration:3.0>>\n<<end>>\n"
	warns := seqWarn(t, src)
	if !hasSeqWarnCode(warns, "SEQ-W002") {
		t.Error("expected SEQ-W002 for camera move_to with no timeout")
	}
}

func TestSeqW002_CameraSetNoWarn(t *testing.T) {
	src := "title: t\nsequence:\n<<camera set point.wide>>\n<<end>>\n"
	warns := seqWarn(t, src)
	if hasSeqWarnCode(warns, "SEQ-W002") {
		t.Error("SEQ-W002 should not fire for camera set (not long-running)")
	}
}

// --- SEQ-W003: <<sync>> all with no pending backgrounds ---

func TestSeqW003_SyncAllNoPending(t *testing.T) {
	src := "title: t\nsequence:\n<<sync>>\n<<end>>\n"
	warns := seqWarn(t, src)
	if !hasSeqWarnCode(warns, "SEQ-W003") {
		t.Error("expected SEQ-W003 for <<sync>> with no pending bg steps")
	}
}

func TestSeqW003_SyncAllWithPending(t *testing.T) {
	src := "title: t\nsequence:\n<<fade_in bg:f>>\n<<sync>>\n<<end>>\n"
	warns := seqWarn(t, src)
	if hasSeqWarnCode(warns, "SEQ-W003") {
		t.Error("SEQ-W003 should not fire when bg steps are pending")
	}
}

func TestSeqW003_SyncAllAfterPreviousSync(t *testing.T) {
	src := "title: t\nsequence:\n<<fade_in bg:f>>\n<<sync f>>\n<<sync>>\n<<end>>\n"
	warns := seqWarn(t, src)
	if !hasSeqWarnCode(warns, "SEQ-W003") {
		t.Error("expected SEQ-W003 for second <<sync>> all after pending steps were already cleared")
	}
}

// --- SEQ-W005: on_fail:skip on state-change step ---

func TestSeqW005_ActionOnFailSkip(t *testing.T) {
	src := "title: t\nsequence:\n<<action flag.done = true on_fail:skip>>\n<<end>>\n"
	warns := seqWarn(t, src)
	if !hasSeqWarnCode(warns, "SEQ-W005") {
		t.Error("expected SEQ-W005 for <<action>> with on_fail:skip")
	}
}

func TestSeqW005_SetOnFailSkip(t *testing.T) {
	src := "title: t\nsequence:\n<<set x=1 on_fail:skip>>\n<<end>>\n"
	warns := seqWarn(t, src)
	if !hasSeqWarnCode(warns, "SEQ-W005") {
		t.Error("expected SEQ-W005 for <<set>> with on_fail:skip")
	}
}

func TestSeqW005_ActionOnFailContinueNoWarn(t *testing.T) {
	src := "title: t\nsequence:\n<<action flag.done = true on_fail:log_and_continue>>\n<<end>>\n"
	warns := seqWarn(t, src)
	if hasSeqWarnCode(warns, "SEQ-W005") {
		t.Error("SEQ-W005 should not fire for on_fail:log_and_continue")
	}
}

// --- SEQ-W006: <<wait_for event:>> not emitted in same file ---

func TestSeqW006_WaitForUnknownEvent(t *testing.T) {
	src := "title: t\nsequence:\n<<wait_for event:player_arrived>>\n<<end>>\n"
	warns := seqWarn(t, src)
	if !hasSeqWarnCode(warns, "SEQ-W006") {
		t.Error("expected SEQ-W006 for wait_for event not emitted in file")
	}
}

func TestSeqW006_WaitForEmittedEvent(t *testing.T) {
	src := "title: t\nsequence:\n<<emit event:player_arrived>>\n<<wait_for event:player_arrived>>\n<<end>>\n"
	warns := seqWarn(t, src)
	if hasSeqWarnCode(warns, "SEQ-W006") {
		t.Error("SEQ-W006 should not fire when event is emitted in same sequence")
	}
}
