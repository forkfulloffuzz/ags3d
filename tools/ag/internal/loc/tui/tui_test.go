package tui

import (
	"testing"

	"github.com/ags3d/ag/internal/loc"
)

func TestNextFilter(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"all", "untranslated"},
		{"untranslated", "translated"},
		{"translated", "stale"},
		{"stale", "orphan"},
		{"orphan", "all"},
		{"unknown", "all"},
	}
	for _, tt := range tests {
		got := nextFilter(tt.input)
		if got != tt.want {
			t.Errorf("nextFilter(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestWrapText(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		width  int
		expect []string
	}{
		{
			name:   "short text fits",
			s:      "Hello world",
			width:  20,
			expect: []string{"Hello world"},
		},
		{
			name:   "long text wraps",
			s:      "The quick brown fox jumps over the lazy dog",
			width:  15,
			expect: []string{"The quick brown", "fox jumps over", "the lazy dog"},
		},
		{
			name:   "zero width",
			s:      "Hello",
			width:  0,
			expect: []string{"Hello"},
		},
		{
			name:   "single word longer than width",
			s:      "superlongword",
			width:  8,
			expect: []string{"superlongword"},
		},
		{
			name:   "paragraph break",
			s:      "Hello\nWorld",
			width:  20,
			expect: []string{"Hello", "World"},
		},
		{
			name:   "empty string",
			s:      "",
			width:  10,
			expect: []string{""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapText(tt.s, tt.width)
			if len(got) != len(tt.expect) {
				t.Errorf("wrapText(%q, %d) = %v, want %v", tt.s, tt.width, got, tt.expect)
				return
			}
			for i := range got {
				if got[i] != tt.expect[i] {
					t.Errorf("wrapText line %d: got %q, want %q", i, got[i], tt.expect[i])
				}
			}
		})
	}
}

func TestPadRight(t *testing.T) {
	tests := []struct {
		s    string
		w    int
		want string
	}{
		{"hello", 10, "hello     "},
		{"hello", 5, "hello"},
		{"hello", 3, "hel"},
		{"", 5, "     "},
		{"hello", 0, "hello"},
	}
	for _, tt := range tests {
		got := padRight(tt.s, tt.w)
		if got != tt.want {
			t.Errorf("padRight(%q, %d) = %q, want %q", tt.s, tt.w, got, tt.want)
		}
	}
}

func TestBuildEntryViews(t *testing.T) {
	src := `[meta]
base_locale = en
locale = fr

[strings]
k1 = "Bonjour"
k2 = ""
// [stale] k3 = "Au revoir"
`
	sf, err := loc.Parse("test.agstrings", src)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	localeSrc := []loc.LocaleEntryFull{
		{LocKey: "k1", Source: "Hello", Character: "Guard", NodeTitle: "greet", LineType: "spoken"},
		{LocKey: "k2", Source: "Goodbye", Character: "", NodeTitle: "farewell", LineType: "narration"},
		{LocKey: "k3", Source: "Bye", Character: "Guard", NodeTitle: "greet", LineType: "spoken"},
	}
	views := buildEntryViews(sf, localeSrc, false)

	if len(views) != 3 {
		t.Fatalf("len = %d, want 3", len(views))
	}

	if views[0].Source != "Hello" {
		t.Errorf("views[0].Source = %q, want Hello", views[0].Source)
	}
	if views[0].Translated != "Bonjour" {
		t.Errorf("views[0].Translated = %q, want Bonjour", views[0].Translated)
	}
	if views[1].Translated != "" {
		t.Errorf("views[1].Translated = %q, want empty", views[1].Translated)
	}
	if !views[2].Stale {
		t.Error("views[2] should be stale")
	}
	if views[2].Source != "Bye" {
		t.Errorf("views[2].Source = %q, want Bye", views[2].Source)
	}
}

func TestBuildEntryViews_SourceLocale(t *testing.T) {
	src := `[meta]
base_locale = en
locale = en

[strings]
k1 = "Hello"
k2 = ""
`
	sf, err := loc.Parse("test.agstrings", src)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	localeSrc := []loc.LocaleEntryFull{
		{LocKey: "k1", Source: "Hello", Character: "Guard", NodeTitle: "greet", LineType: "spoken"},
		{LocKey: "k2", Source: "Goodbye", Character: "", NodeTitle: "farewell", LineType: "narration"},
	}
	views := buildEntryViews(sf, localeSrc, true)

	if len(views) != 2 {
		t.Fatalf("len = %d, want 2", len(views))
	}

	if views[0].Translated != "" {
		t.Errorf("views[0].Translated = %q, want empty for source locale", views[0].Translated)
	}
	if views[0].Source != "Hello" {
		t.Errorf("views[0].Source = %q, want Hello", views[0].Source)
	}
	if views[1].Translated != "" {
		t.Errorf("views[1].Translated = %q, want empty for source locale", views[1].Translated)
	}
}
