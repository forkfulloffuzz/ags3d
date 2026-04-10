// Command ag-loc is a standalone translation editor for AGS3D .agstrings files.
//
// It provides an interactive TUI for translators to fill in translations
// without needing to use AG Studio or edit .agstrings files by hand.
//
// Usage:
//
//	ag-loc <project> [--locale LANG]
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ags3d/ag/internal/loc/tui"
)

func main() {
	fs := flag.NewFlagSet("ag-loc", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, `Usage: ag-loc <project> [--locale LANG]

ag-loc is a standalone translation editor for AGS3D projects.

Arguments:
  project         Path to the AGS3D project root (directory containing game.agp)

Flags:
  --locale LANG  Locale code to edit (e.g. fr, es, he). Defaults to the project's
                 default_author_locale or "en".

Navigation:
  ↑/↓ or j/k    Navigate string list
  Enter          Edit the selected translation
  Tab            Cycle filter: all → untranslated → translated → stale → orphan
  l              Switch locale
  ctrl+s         Save changes
  q              Quit (prompts if unsaved changes)`)
	}
	locale := fs.String("locale", "", "locale code to edit")
	if err := fs.Parse(os.Args[1:]); err != nil {
		if err == flag.ErrHelp {
			return
		}
		os.Exit(1)
	}

	if fs.NArg() < 1 {
		fs.Usage()
		os.Exit(1)
	}

	root := fs.Arg(0)
	if err := tui.RunTUIMain(root, *locale); err != nil {
		fmt.Fprintf(os.Stderr, "ag-loc: %v\n", err)
		os.Exit(1)
	}
}
