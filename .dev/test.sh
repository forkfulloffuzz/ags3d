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
  STATUS=0
  "$GODOT" --headless --path "$TEST_PROJECT" --script "$SCRIPT" || STATUS=$?
  exit $STATUS
fi

# Suppress Godot engine noise; only show AGS3D test output.
# Output is streamed live (--line-buffered) so progress is visible while running.
NOISE_PATTERN="Godot Engine|godotengine\.org|nvidia|Gtk|Adwaita|Thread|libpulse|libvulkan|libVk|^Xlib|^$"
# Engine noise: warning/error boilerplate lines (at:, GDScript backtrace, bracketed entries,
# tab-indented continuation lines) and known expected warnings from specific tests.
NOISE_PATTERN+="|^\s+at:|GDScript backtrace|^\s+\[|^\t"
NOISE_PATTERN+="|AGSSpawnPoint.*not found in AGSRuntime"
NOISE_PATTERN+="|Source geometry parsing.*navigation mesh|visual meshes store geometry|For runtime.*baking navigation"
# M6/E2E: NavigationServer3D fires "!found" when the nav map isn't fully initialised
#     during the first physics frame of an async E2E test.  Tests still pass.
NOISE_PATTERN+="|Condition.*!found.*is true"
# M6: ScriptWiring — room.room_enter accessed as signal while GDScript method of same name
#     shadows it; the connect() call fails but the signal still fires via AGSRoom's C++ emit.
NOISE_PATTERN+="|Nonexistent function 'connect' in base 'Callable'"
# AGSRuntime trace output (enabled by default in all builds; verbose mode shows it)
NOISE_PATTERN+="|\[AGS/"
# M10: Cutscene — CanvasLayer/ColorRect have no viewport in headless mode; the overlay
#     is visual-only so this is harmless. Also suppress the is_inside_tree guard that
#     fires when Godot checks layout for a node freed before its deferred _ready runs.
NOISE_PATTERN+="|Parameter.*get_viewport.*is null|Condition.*is_inside_tree.*Returning"
# M10: Globals — test_21 intentionally queries a non-existent global to verify null return.
NOISE_PATTERN+="|AGSRuntime.get_global: unknown global"
# Godot resource-leak warnings at exit — harmless cleanup noise
NOISE_PATTERN+="|RID allocations.*were leaked|ObjectDB instances were leaked|resources still in use"

# Pipe output through the noise filter in real-time.
# set -e is suspended for the pipe so grep's exit code doesn't abort the script.
set +e
if [[ -n "$FILTER" ]]; then
  "$GODOT" --headless --path "$TEST_PROJECT" --script "$SCRIPT" 2>&1 | \
    grep --line-buffered -v -E "$NOISE_PATTERN" | \
    grep --line-buffered -i "$FILTER"
else
  "$GODOT" --headless --path "$TEST_PROJECT" --script "$SCRIPT" 2>&1 | \
    grep --line-buffered -v -E "$NOISE_PATTERN"
fi
STATUS=${PIPESTATUS[0]}
set -e
exit $STATUS
