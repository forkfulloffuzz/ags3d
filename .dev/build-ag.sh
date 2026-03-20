#!/usr/bin/env bash
# Build the ag CLI and agls language server Go binaries.
#
# Usage:
#   .dev/build-ag.sh           # build ag + agls → bin/
#   .dev/build-ag.sh ag        # build ag only
#   .dev/build-ag.sh agls      # build agls only
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
    echo "→ Removing bin/ag and bin/agls…"
    rm -f "$BIN_DIR/ag" "$BIN_DIR/agls"
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
  all)
    require_go
    echo "→ Building ag + agls…"
    cd "$AG_MODULE"
    go build -o "$BIN_DIR/ag"   ./cmd/ag
    go build -o "$BIN_DIR/agls" ./cmd/agls
    echo "✓ bin/ag  bin/agls"
    ;;
  *)
    echo "Usage: .dev/build-ag.sh [ag|agls|all|clean]" >&2
    exit 1
    ;;
esac
