#!/usr/bin/env bash
# Run the AGS3D GDScript test suite headlessly.
#
# Usage:
#   .dev/test.sh               # run all suites
#   .dev/test.sh --verbose     # show Godot engine output too
#   .dev/test.sh --filter M1   # only print results matching a pattern
#
# Exit code: 0 = all pass, 1 = any failure.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

GODOT="$REPO_ROOT/bin/godot.linuxbsd.editor.x86_64"
TEST_PROJECT="$REPO_ROOT/agstests"
SCRIPT="run_tests.gd"

VERBOSE=0
FILTER=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --verbose|-v) VERBOSE=1 ;;
    --filter|-f)  FILTER="$2"; shift ;;
    *)
      echo "Usage: .dev/test.sh [--verbose] [--filter PATTERN]" >&2
      exit 1
      ;;
  esac
  shift
done

if [[ ! -x "$GODOT" ]]; then
  echo "✗ Godot binary not found at: $GODOT" >&2
  echo "  Run .dev/build.sh first." >&2
  exit 1
fi

# Clear Godot's bytecode cache so edited scripts are always picked up fresh.
rm -rf "$TEST_PROJECT/.godot"

echo ""

if [[ $VERBOSE -eq 1 ]]; then
  "$GODOT" --headless --path "$TEST_PROJECT" --script "$SCRIPT"
  exit $?
fi

# Suppress Godot engine noise; only show AGS3D test output.
TMPOUT="$(mktemp)"
"$GODOT" --headless --path "$TEST_PROJECT" --script "$SCRIPT" >"$TMPOUT" 2>&1
STATUS=$?

NOISE_PATTERN="Godot Engine|godotengine\.org|nvidia|Gtk|Adwaita|Thread|libpulse|libvulkan|libVk|^Xlib|^$"
# Engine noise: warning/error boilerplate lines (at:, GDScript backtrace, bracketed entries,
# tab-indented continuation lines) and known expected warnings from specific tests.
NOISE_PATTERN+="|^\s+at:|GDScript backtrace|^\s+\[|^\t"
NOISE_PATTERN+="|AGSSpawnPoint.*not found in AGSRuntime"
NOISE_PATTERN+="|Source geometry parsing.*navigation mesh|visual meshes store geometry|For runtime.*baking navigation"

if [[ -n "$FILTER" ]]; then
  grep -v -E "$NOISE_PATTERN" "$TMPOUT" | grep -i "$FILTER" || true
else
  grep -v -E "$NOISE_PATTERN" "$TMPOUT" || true
fi

rm -f "$TMPOUT"
exit $STATUS
