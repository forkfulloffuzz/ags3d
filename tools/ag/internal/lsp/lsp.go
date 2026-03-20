// Package lsp implements the AGS-spirit Language Server Protocol server.
// It exposes the parser and analysis packages over JSON-RPC (stdio transport)
// to VS Code, the Godot editor, and any LSP-compatible editor.
// Stub — full implementation post-M2/M3.
package lsp

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Serve runs the LSP server reading from r and writing to w (typically
// os.Stdin / os.Stdout). Blocks until the connection is closed.
// TODO: implement full JSON-RPC dispatch and LSP request handlers.
func Serve(r io.Reader, w io.Writer) error {
	fmt.Fprintln(os.Stderr, "agls: language server not yet implemented")
	// Drain stdin so the client does not hang.
	_, _ = io.Copy(io.Discard, r)
	_ = json.NewEncoder(w) // suppress unused import
	return nil
}
