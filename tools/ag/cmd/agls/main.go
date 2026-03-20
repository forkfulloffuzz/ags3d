// Command agls is the AGS-spirit Language Server.
//
// Communicates over stdin/stdout using JSON-RPC (Language Server Protocol).
// Launched automatically by VS Code and the Godot editor LSP client.
//
// Full implementation post-M2/M3 when the parser and symbol table are ready.
package main

import (
	"fmt"
	"os"

	"github.com/ags3d/ag/internal/lsp"
)

func main() {
	if err := lsp.Serve(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "agls:", err)
		os.Exit(1)
	}
}
