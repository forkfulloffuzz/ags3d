package cut_test

import (
	"testing"

	"github.com/ags3d/ag/internal/cut"
)

func mustParseCF(t *testing.T, src string) *cut.CutsceneFile {
	t.Helper()
	cf, err := cut.Parse("test.agcut", src)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	return cf
}

func validate(t *testing.T, src string) []cut.ValidationError {
	t.Helper()
	cf := mustParseCF(t, src)
	return cut.ValidateCutscene(cf, map[string]bool{cf.Title: true}, nil)
}

func hasErrCode(errs []cut.ValidationError, code string) bool {
	for _, e := range errs {
		if e.Code == code {
			return true
		}
	}
	return false
}

// --- CUT-E001: title uniqueness / missing ---

func TestValidate_MissingTitleIsE001(t *testing.T) {
	cf, err := cut.Parse("test.agcut", "sequence:\n<<end>>\n")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	errs := cut.ValidateCutscene(cf, nil, nil)
	if !hasErrCode(errs, "CUT-E001") {
		t.Error("expected CUT-E001 for missing title, none found")
	}
}

func TestValidateProject_DuplicateTitleIsE001(t *testing.T) {
	a := mustParseCF(t, "title: shared\nsequence:\n<<end>>\n")
	b := mustParseCF(t, "title: shared\nsequence:\n<<end>>\n")
	b.Path = "b.agcut"
	errs := cut.ValidateProjectCutscenes([]*cut.CutsceneFile{a, b}, nil)
	if !hasErrCode(errs, "CUT-E001") {
		t.Error("expected CUT-E001 for duplicate title, none found")
	}
}

func TestValidateProject_UniqueTitlesNoError(t *testing.T) {
	a := mustParseCF(t, "title: cut_a\nsequence:\n<<end>>\n")
	b := mustParseCF(t, "title: cut_b\nsequence:\n<<end>>\n")
	b.Path = "b.agcut"
	errs := cut.ValidateProjectCutscenes([]*cut.CutsceneFile{a, b}, nil)
	for _, e := range errs {
		if e.Code == "CUT-E001" {
			t.Errorf("unexpected CUT-E001: %v", e)
		}
	}
}

// --- CUT-E006: skip_to label exists ---

func TestValidate_SkipToExistingLabelNoError(t *testing.T) {
	errs := validate(t, "title: t\nsequence:\n<<label end_scene>>\n<<skip_to end_scene>>\n<<end>>\n")
	for _, e := range errs {
		if e.Code == "CUT-E006" {
			t.Errorf("unexpected CUT-E006: %v", e)
		}
	}
}

func TestValidate_SkipToMissingLabelIsE006(t *testing.T) {
	errs := validate(t, "title: t\nsequence:\n<<skip_to nowhere>>\n<<end>>\n")
	if !hasErrCode(errs, "CUT-E006") {
		t.Error("expected CUT-E006 for missing label, none found")
	}
}

// --- CUT-E007: choice inside parallel ---

func TestValidate_ChoiceOutsideParallelNoError(t *testing.T) {
	errs := validate(t, "title: t\nsequence:\n<<choice>>\n<<end>>\n")
	for _, e := range errs {
		if e.Code == "CUT-E007" {
			t.Errorf("unexpected CUT-E007: %v", e)
		}
	}
}

func TestValidate_ChoiceInsideParallelIsE007(t *testing.T) {
	errs := validate(t, "title: t\nsequence:\n<<parallel>>\n<<choice>>\n<<end_parallel>>\n<<end>>\n")
	if !hasErrCode(errs, "CUT-E007") {
		t.Error("expected CUT-E007 for choice inside parallel, none found")
	}
}

// --- CUT-E008: nested cutscene exists ---

func TestValidate_NestedCutsceneExistsNoError(t *testing.T) {
	cf := mustParseCF(t, "title: t\nsequence:\n<<cutscene other>>\n<<end>>\n")
	allTitles := map[string]bool{"t": true, "other": true}
	errs := cut.ValidateCutscene(cf, allTitles, nil)
	for _, e := range errs {
		if e.Code == "CUT-E008" {
			t.Errorf("unexpected CUT-E008: %v", e)
		}
	}
}

func TestValidate_NestedCutsceneMissingIsE008(t *testing.T) {
	cf := mustParseCF(t, "title: t\nsequence:\n<<cutscene ghost>>\n<<end>>\n")
	allTitles := map[string]bool{"t": true}
	errs := cut.ValidateCutscene(cf, allTitles, nil)
	if !hasErrCode(errs, "CUT-E008") {
		t.Error("expected CUT-E008 for missing nested cutscene, none found")
	}
}

// --- CUT-E009: circular nesting (self-reference) ---

func TestValidate_SelfReferenceIsE009(t *testing.T) {
	cf := mustParseCF(t, "title: t\nsequence:\n<<cutscene t>>\n<<end>>\n")
	allTitles := map[string]bool{"t": true}
	errs := cut.ValidateCutscene(cf, allTitles, nil)
	if !hasErrCode(errs, "CUT-E009") {
		t.Error("expected CUT-E009 for self-reference, none found")
	}
}

// --- CUT-E012: save_block:false with state changes ---

func TestValidate_SaveBlockFalseWithActionIsE012(t *testing.T) {
	errs := validate(t, "title: t\nsave_block: false\nsequence:\n<<action flag.x = true>>\n<<end>>\n")
	if !hasErrCode(errs, "CUT-E012") {
		t.Error("expected CUT-E012 for save_block:false with <<action>>, none found")
	}
}

func TestValidate_SaveBlockFalseWithSetIsE012(t *testing.T) {
	errs := validate(t, "title: t\nsave_block: false\nsequence:\n<<set counter = 1>>\n<<end>>\n")
	if !hasErrCode(errs, "CUT-E012") {
		t.Error("expected CUT-E012 for save_block:false with <<set>>, none found")
	}
}

func TestValidate_SaveBlockTrueWithActionNoError(t *testing.T) {
	errs := validate(t, "title: t\nsave_block: true\nsequence:\n<<action flag.x = true>>\n<<end>>\n")
	for _, e := range errs {
		if e.Code == "CUT-E012" {
			t.Errorf("unexpected CUT-E012: %v", e)
		}
	}
}

func TestValidate_SaveBlockFalseNoStateChangeNoError(t *testing.T) {
	errs := validate(t, "title: t\nsave_block: false\nsequence:\n<<fade_in duration:2.0>>\n<<fade_out duration:1.0>>\n")
	for _, e := range errs {
		if e.Code == "CUT-E012" {
			t.Errorf("unexpected CUT-E012: %v", e)
		}
	}
}

// --- Cross-system checks ---

func TestValidate_CharacterExistsNoError(t *testing.T) {
	cf := mustParseCF(t, "title: t\nsequence:\n<<character player walk_to point.entrance>>\n<<end>>\n")
	idx := &cut.ProjectIndex{
		Characters: map[string]bool{"player": true},
	}
	errs := cut.ValidateCutscene(cf, map[string]bool{"t": true}, idx)
	for _, e := range errs {
		if e.Code == "CUT-E003" {
			t.Errorf("unexpected CUT-E003: %v", e)
		}
	}
}

func TestValidate_CharacterMissingIsE003(t *testing.T) {
	cf := mustParseCF(t, "title: t\nsequence:\n<<character ghost walk_to point.x>>\n<<end>>\n")
	idx := &cut.ProjectIndex{
		Characters: map[string]bool{"player": true},
	}
	errs := cut.ValidateCutscene(cf, map[string]bool{"t": true}, idx)
	if !hasErrCode(errs, "CUT-E003") {
		t.Error("expected CUT-E003 for unknown character, none found")
	}
}

func TestValidate_AudioMissingIsE004(t *testing.T) {
	cf := mustParseCF(t, "title: t\nsequence:\n<<music ghost_track>>\n<<end>>\n")
	idx := &cut.ProjectIndex{
		AudioFiles: map[string]bool{"theme_main": true},
	}
	errs := cut.ValidateCutscene(cf, map[string]bool{"t": true}, idx)
	if !hasErrCode(errs, "CUT-E004") {
		t.Error("expected CUT-E004 for missing audio, none found")
	}
}

func TestValidate_VideoMissingIsE005(t *testing.T) {
	cf := mustParseCF(t, "title: t\nsequence:\n<<video ghost_clip>>\n<<end>>\n")
	idx := &cut.ProjectIndex{
		VideoFiles: map[string]bool{"intro_video": true},
	}
	errs := cut.ValidateCutscene(cf, map[string]bool{"t": true}, idx)
	if !hasErrCode(errs, "CUT-E005") {
		t.Error("expected CUT-E005 for missing video, none found")
	}
}

func TestValidate_PointMissingIsE002(t *testing.T) {
	cf := mustParseCF(t, "title: t\nsequence:\n<<camera move_to point.ghost_point>>\n<<end>>\n")
	idx := &cut.ProjectIndex{
		NamedPoints: map[string]map[string]bool{
			"start_room": {"entrance": true},
		},
	}
	errs := cut.ValidateCutscene(cf, map[string]bool{"t": true}, idx)
	if !hasErrCode(errs, "CUT-E002") {
		t.Error("expected CUT-E002 for missing named point, none found")
	}
}

func TestValidate_PointExistsNoError(t *testing.T) {
	cf := mustParseCF(t, "title: t\nsequence:\n<<camera move_to point.entrance>>\n<<end>>\n")
	idx := &cut.ProjectIndex{
		NamedPoints: map[string]map[string]bool{
			"start_room": {"entrance": true},
		},
	}
	errs := cut.ValidateCutscene(cf, map[string]bool{"t": true}, idx)
	for _, e := range errs {
		if e.Code == "CUT-E002" {
			t.Errorf("unexpected CUT-E002: %v", e)
		}
	}
}

// --- CUT-E013: identifier naming rule ---

func TestValidate_TitleUppercaseIsE013(t *testing.T) {
	cf := mustParseCF(t, "title: Chapter1Opening\nsequence:\n<<end>>\n")
	errs := cut.ValidateCutscene(cf, nil, nil)
	if !hasErrCode(errs, "CUT-E013") {
		t.Error("expected CUT-E013 for uppercase title, none found")
	}
}

func TestValidate_TitleStartsWithDigitIsE013(t *testing.T) {
	cf := mustParseCF(t, "title: 1st_scene\nsequence:\n<<end>>\n")
	errs := cut.ValidateCutscene(cf, nil, nil)
	if !hasErrCode(errs, "CUT-E013") {
		t.Error("expected CUT-E013 for digit-leading title, none found")
	}
}

func TestValidate_TitleValidNoE013(t *testing.T) {
	cf := mustParseCF(t, "title: chapter1_opening\nsequence:\n<<end>>\n")
	errs := cut.ValidateCutscene(cf, nil, nil)
	for _, e := range errs {
		if e.Code == "CUT-E013" {
			t.Errorf("unexpected CUT-E013: %v", e)
		}
	}
}

func TestValidate_LabelInvalidIsE013(t *testing.T) {
	errs := validate(t, "title: t\nsequence:\n<<label Act-Two>>\n<<end>>\n")
	if !hasErrCode(errs, "CUT-E013") {
		t.Error("expected CUT-E013 for hyphen in label, none found")
	}
}

func TestValidate_LabelValidNoE013(t *testing.T) {
	errs := validate(t, "title: t\nsequence:\n<<label act_two>>\n<<skip_to act_two>>\n<<end>>\n")
	for _, e := range errs {
		if e.Code == "CUT-E013" {
			t.Errorf("unexpected CUT-E013: %v", e)
		}
	}
}

func TestValidate_BgIDInvalidIsE013(t *testing.T) {
	errs := validate(t, "title: t\nsequence:\n<<fade_in duration:1.0 bg:Cam.Move>>\n<<end>>\n")
	if !hasErrCode(errs, "CUT-E013") {
		t.Error("expected CUT-E013 for invalid bg: id, none found")
	}
}

func TestValidate_BgIDValidNoE013(t *testing.T) {
	errs := validate(t, "title: t\nsequence:\n<<fade_in duration:1.0 bg:cam_move>>\n<<end>>\n")
	for _, e := range errs {
		if e.Code == "CUT-E013" {
			t.Errorf("unexpected CUT-E013: %v", e)
		}
	}
}

func TestValidate_CutsceneRefInvalidIsE013(t *testing.T) {
	cf := mustParseCF(t, "title: t\nsequence:\n<<cutscene BadRef>>\n<<end>>\n")
	errs := cut.ValidateCutscene(cf, nil, nil)
	if !hasErrCode(errs, "CUT-E013") {
		t.Error("expected CUT-E013 for uppercase cutscene ref, none found")
	}
}

func TestValidate_CutsceneFileParamInvalidIsE013(t *testing.T) {
	cf := mustParseCF(t, "title: t\nsequence:\n<<cutscene file:BadRef>>\n<<end>>\n")
	errs := cut.ValidateCutscene(cf, nil, nil)
	if !hasErrCode(errs, "CUT-E013") {
		t.Error("expected CUT-E013 for uppercase file: cutscene ref, none found")
	}
}

func TestValidate_LocGroupInvalidIsE013(t *testing.T) {
	cf := mustParseCF(t, "title: t\nloc_group: Chapter1\nsequence:\n<<end>>\n")
	errs := cut.ValidateCutscene(cf, nil, nil)
	if !hasErrCode(errs, "CUT-E013") {
		t.Error("expected CUT-E013 for uppercase loc_group, none found")
	}
}

func TestValidate_VoiceSessionInvalidIsE013(t *testing.T) {
	cf := mustParseCF(t, "title: t\nvoice_session: Session-A\nsequence:\n<<end>>\n")
	errs := cut.ValidateCutscene(cf, nil, nil)
	if !hasErrCode(errs, "CUT-E013") {
		t.Error("expected CUT-E013 for hyphen in voice_session, none found")
	}
}

func TestValidate_ReservedIdentNoE013(t *testing.T) {
	// "room_music" and "room_ambient" satisfy the regex anyway, but exercise the exempt path.
	errs := validate(t, "title: t\nsequence:\n<<ambient stop channel:room_music>>\n<<end>>\n")
	for _, e := range errs {
		if e.Code == "CUT-E013" {
			t.Errorf("unexpected CUT-E013 for reserved ident: %v", e)
		}
	}
}

// --- CUT-E015: <<cutscene file:name>> named-param form ---

func TestValidate_CutsceneFileParamExistsNoError(t *testing.T) {
	cf := mustParseCF(t, "title: t\nsequence:\n<<cutscene file:other>>\n<<end>>\n")
	allTitles := map[string]bool{"t": true, "other": true}
	errs := cut.ValidateCutscene(cf, allTitles, nil)
	for _, e := range errs {
		if e.Code == "CUT-E008" {
			t.Errorf("unexpected CUT-E008: %v", e)
		}
	}
}

func TestValidate_CutsceneFileParamMissingIsE008(t *testing.T) {
	cf := mustParseCF(t, "title: t\nsequence:\n<<cutscene file:does_not_exist>>\n<<end>>\n")
	allTitles := map[string]bool{"t": true}
	errs := cut.ValidateCutscene(cf, allTitles, nil)
	if !hasErrCode(errs, "CUT-E008") {
		t.Error("expected CUT-E008 for missing file: cutscene ref, none found")
	}
}

func TestValidate_CutsceneFileParamSelfIsE009(t *testing.T) {
	cf := mustParseCF(t, "title: t\nsequence:\n<<cutscene file:t>>\n<<end>>\n")
	allTitles := map[string]bool{"t": true}
	errs := cut.ValidateCutscene(cf, allTitles, nil)
	if !hasErrCode(errs, "CUT-E009") {
		t.Error("expected CUT-E009 for self-referencing file: form, none found")
	}
}

// --- Clean cutscene produces no errors ---

func TestValidate_CleanCutsceneNoErrors(t *testing.T) {
	src := `title: chapter1_opening
skip: after_first_view
save_block: true
tags: [chapter:1, cinematic]
fallback: halt
sequence:
<<fade_in duration:2.0>>
<<label start>>
<<music theme_main fade_in:3.0>>
<<line narrator "Three years.">>
<<action flag.started = true>>
<<fade_out duration:1.0>>
<<end>>
`
	cf := mustParseCF(t, src)
	errs := cut.ValidateCutscene(cf, map[string]bool{"chapter1_opening": true}, nil)
	if len(errs) != 0 {
		t.Errorf("expected 0 errors for valid cutscene, got: %v", errs)
	}
}
