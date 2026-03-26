#!/usr/bin/env bash
# Open the AGS3D prototype project in the Godot editor.
#
# Usage:
#   .dev/prototype.sh          # open in editor
#   .dev/prototype.sh --run    # run the project (headless, no editor UI)

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PROTOTYPE_DIR="$REPO_ROOT/game_prototype"
GODOT="$REPO_ROOT/bin/godot.linuxbsd.editor.x86_64"

if [[ ! -x "$GODOT" ]]; then
  echo "✗ Godot binary not found at $GODOT" >&2
  echo "  Run .dev/build.sh first." >&2
  exit 1
fi

case "${1:-}" in
  --run)
    echo "→ Running prototype project…"
    exec "$GODOT" --path "$PROTOTYPE_DIR" --headless
    ;;
  "")
    echo "→ Opening prototype project in Godot editor…"
    exec "$GODOT" --editor --path "$PROTOTYPE_DIR"
    ;;
  *)
    echo "Usage: .dev/prototype.sh [--run]" >&2
    exit 1
    ;;
esac
