#!/usr/bin/env bash
# Build the ag CLI, agls language server, and agui desktop app.
#
# Usage:
#   .dev/build-ag.sh           # build ag + agls → bin/
#   .dev/build-ag.sh ag        # build ag only
#   .dev/build-ag.sh agls      # build agls only
#   .dev/build-ag.sh agui      # build AG Studio desktop app → bin/agui
#   .dev/build-ag.sh clean     # remove compiled binaries

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
AG_MODULE="$REPO_ROOT/tools/ag"
BIN_DIR="$REPO_ROOT/bin"

TARGET="${1:-all}"

require_go() {
  if ! command -v go &>/dev/null; then
    echo "✗ Go not found. Install from https://go.dev/dl/ or: sudo apt install golang-go" >&2
    exit 1
  fi
}

mkdir -p "$BIN_DIR"

case "$TARGET" in
  clean)
    echo "→ Removing bin/ag, bin/agls, bin/agui…"
    rm -f "$BIN_DIR/ag" "$BIN_DIR/agls" "$BIN_DIR/agui"
    echo "✓ Done"
    exit 0
    ;;
  ag)
    require_go
    echo "→ Building ag…"
    cd "$AG_MODULE"
    go build -o "$BIN_DIR/ag" ./cmd/ag
    echo "✓ bin/ag"
    ;;
  agls)
    require_go
    echo "→ Building agls…"
    cd "$AG_MODULE"
    go build -o "$BIN_DIR/agls" ./cmd/agls
    echo "✓ bin/agls"
    ;;
  agui)
    require_go
    if ! command -v wails &>/dev/null && ! command -v ~/go/bin/wails &>/dev/null; then
      echo "✗ wails not found. Install: go install github.com/wailsapp/wails/v2/cmd/wails@latest" >&2
      exit 1
    fi
    WAILS="${HOME}/go/bin/wails"
    [[ -x "$(command -v wails)" ]] && WAILS="wails"
    echo "→ Building AG Studio (agui)…"
    cd "$REPO_ROOT/tools/agui"
    "$WAILS" build -tags webkit2_41
    mv "build/bin/agui" "$BIN_DIR/agui"
    echo "✓ bin/agui"
    ;;
  all)
    require_go
    echo "→ Building ag + agls…"
    cd "$AG_MODULE"
    go build -o "$BIN_DIR/ag"   ./cmd/ag
    go build -o "$BIN_DIR/agls" ./cmd/agls
    echo "✓ bin/ag  bin/agls"
    ;;
  *)
    echo "Usage: .dev/build-ag.sh [ag|agls|agui|all|clean]" >&2
    exit 1
    ;;
esac
