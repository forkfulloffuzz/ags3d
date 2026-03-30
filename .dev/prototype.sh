#!/usr/bin/env bash
# Open the AGS3D prototype project in the Godot editor.
#
# Usage:
#   .dev/prototype.sh                # open in AG Studio editor (plugin active)
#   .dev/prototype.sh --godot-editor # open in standard Godot editor (plugin bypassed)
#   .dev/prototype.sh --play         # run the game with display (no editor UI)
#   .dev/prototype.sh --run          # run the project headless (no display, for CI)

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
  --play)
    echo "→ Playing prototype project…"
    exec "$GODOT" --path "$PROTOTYPE_DIR"
    ;;
  --run)
    echo "→ Running prototype project (headless)…"
    exec "$GODOT" --path "$PROTOTYPE_DIR" --headless
    ;;
  --godot-editor)
    echo "→ Opening prototype in standard Godot editor (AG Studio plugin bypassed)…"
    exec "$GODOT" --editor --path "$PROTOTYPE_DIR" --godot-editor
    ;;
  "")
    echo "→ Opening prototype project in Godot editor…"
    exec "$GODOT" --editor --path "$PROTOTYPE_DIR"
    ;;
  *)
    echo "Usage: .dev/prototype.sh [--godot-editor|--play|--run]" >&2
    exit 1
    ;;
esac
