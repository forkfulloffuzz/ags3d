package item_test

import (
	"strings"
	"testing"

	"github.com/ags3d/ag/internal/item"
)

func mustParse(t *testing.T, src string) *item.ItemData {
	t.Helper()
	it, err := item.ParseItem("test.agitem", src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return it
}

func mustFail(t *testing.T, src, wantSubstr string) {
	t.Helper()
	_, err := item.ParseItem("test.agitem", src)
	if err == nil {
		t.Fatalf("expected error containing %q, got nil", wantSubstr)
	}
	if !strings.Contains(err.Error(), wantSubstr) {
		t.Fatalf("expected error containing %q, got: %v", wantSubstr, err)
	}
}

// --------------------------------------------------------------------------
// Happy path
// --------------------------------------------------------------------------

func TestMinimalItem(t *testing.T) {
	it := mustParse(t, `Item "rusty_key" {}`)
	if it.Name != "rusty_key" {
		t.Errorf("Name = %q", it.Name)
	}
}

func TestDisplayName(t *testing.T) {
	it := mustParse(t, `Item "coin" { display_name = "Gold Coin" }`)
	if it.DisplayName != "Gold Coin" {
		t.Errorf("DisplayName = %q", it.DisplayName)
	}
}

func TestDescription(t *testing.T) {
	it := mustParse(t, `Item "coin" { description = "A shiny gold coin." }`)
	if it.Description != "A shiny gold coin." {
		t.Errorf("Description = %q", it.Description)
	}
}

func TestSprite(t *testing.T) {
	it := mustParse(t, `Item "key" { sprite = "assets/items/key.png" }`)
	if it.Sprite != "assets/items/key.png" {
		t.Errorf("Sprite = %q", it.Sprite)
	}
}

func TestFullItem(t *testing.T) {
	src := `Item "rusty_key" {
		display_name = "Rusty Key"
		description  = "An old iron key."
		sprite       = "assets/items/rusty_key.png"
	}`
	it := mustParse(t, src)
	if it.Name != "rusty_key" {
		t.Errorf("Name = %q", it.Name)
	}
	if it.DisplayName != "Rusty Key" {
		t.Errorf("DisplayName = %q", it.DisplayName)
	}
	if it.Description != "An old iron key." {
		t.Errorf("Description = %q", it.Description)
	}
	if it.Sprite != "assets/items/rusty_key.png" {
		t.Errorf("Sprite = %q", it.Sprite)
	}
}

func TestComments(t *testing.T) {
	src := `// top comment
Item "key" {
	// field comment
	display_name = "Key" // trailing
}`
	it := mustParse(t, src)
	if it.DisplayName != "Key" {
		t.Errorf("DisplayName = %q", it.DisplayName)
	}
}

// --------------------------------------------------------------------------
// Error cases
// --------------------------------------------------------------------------

func TestErrNotItem(t *testing.T) {
	mustFail(t, `Room "r" {}`, `expected "Item"`)
}

func TestErrMissingName(t *testing.T) {
	mustFail(t, `Item {}`, `expected item name string`)
}

func TestErrMissingBrace(t *testing.T) {
	mustFail(t, `Item "key"`, `expected "{"`)
}

func TestErrUnterminatedBlock(t *testing.T) {
	mustFail(t, `Item "key" { display_name = "Key"`, `unexpected end of file`)
}

func TestErrUnterminatedString(t *testing.T) {
	mustFail(t, `Item "key" { display_name = "Key }`, `unterminated string`)
}
