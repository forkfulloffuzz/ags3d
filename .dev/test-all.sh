#!/usr/bin/env bash
# Run every test suite across the AGS3D project.
#
# Output: one ✓/✗ line per package/suite, grouped by tool.
# GDScript output streams live so hanging tests are visible immediately.
#
# Usage:
#   .dev/test-all.sh               # run everything
#   .dev/test-all.sh --verbose     # pass -v to go test; show raw Godot output
#   .dev/test-all.sh --filter FOO  # filter test names / Godot output
#   .dev/test-all.sh --no-godot    # skip GDScript suite even if Godot is present
#   .dev/test-all.sh --timeout N   # per-Godot-run timeout in seconds (default 300)
#
# Exit code: 0 = all suites passed, 1 = any failure.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

VERBOSE=0
FILTER=""
NO_GODOT=0
GODOT_TIMEOUT=300

while [[ $# -gt 0 ]]; do
  case "$1" in
    --verbose|-v)   VERBOSE=1 ;;
    --filter|-f)    FILTER="$2"; shift ;;
    --no-godot)     NO_GODOT=1 ;;
    --timeout|-t)   GODOT_TIMEOUT="$2"; shift ;;
    *)
      echo "Usage: .dev/test-all.sh [--verbose] [--filter PATTERN] [--no-godot] [--timeout N]" >&2
      exit 1
      ;;
  esac
  shift
done

PASS=()
FAIL=()

# ── helpers ────────────────────────────────────────────────────────────────────

section() { echo ""; echo "══════════════════════════════════════"; echo "  $*"; echo "══════════════════════════════════════"; }
ok()      { PASS+=("$*"); printf "  \033[32m✓\033[0m %s\n" "$*"; }
fail()    { FAIL+=("$*"); printf "  \033[31m✗\033[0m %s\n" "$*"; }
skip()    { printf "  \033[33m-\033[0m %s\n" "$*"; }

# Run a Godot headless process, streaming output live through the noise filter.
# Captures the exit code even though output is piped. Applies $GODOT_TIMEOUT.
# Usage: run_godot <godot_bin> <project_path> <script> <noise_pattern> <label>
run_godot() {
  local godot="$1" project="$2" script="$3" noise="$4" label="$5"
  local exitfile
  exitfile="$(mktemp)"

  if [[ $VERBOSE -eq 1 ]]; then
    # Raw output — no filtering.
    {
      timeout --kill-after=10 "$GODOT_TIMEOUT" \
        "$godot" --headless --path "$project" --script "$script"
      echo "$?" > "$exitfile"
    }
  else
    # Stream live, filter noise; FILTER is applied as an extra grep if set.
    {
      timeout --kill-after=10 "$GODOT_TIMEOUT" \
        "$godot" --headless --path "$project" --script "$script"
      echo "$?" > "$exitfile"
    } 2>&1 | {
      if [[ -n "$FILTER" ]]; then
        grep -v -E "$noise" | grep -i "$FILTER" || true
      else
        grep -v -E "$noise" || true
      fi
    }
  fi

  local gx
  gx=$(cat "$exitfile" 2>/dev/null || echo 1)
  rm -f "$exitfile"

  if [[ "$gx" -eq 0 ]]; then
    ok "$label"
  elif [[ "$gx" -eq 124 || "$gx" -eq 137 ]]; then
    fail "$label (timed out after ${GODOT_TIMEOUT}s)"
  else
    fail "$label"
  fi
}

# Run one Go package; label is the display name shown in results.
# Usage: run_go_pkg <dir> <import_path> <label>
run_go_pkg() {
  local dir="$1" pkg="$2" label="$3"
  local args=("$pkg")
  [[ $VERBOSE -eq 1 ]] && args=("-v" "${args[@]}")
  [[ -n "$FILTER" ]]   && args=("-run" "$FILTER" "${args[@]}")

  local tmpout
  tmpout="$(mktemp)"
  local exit_code=0
  (cd "$dir" && go test "${args[@]}") >"$tmpout" 2>&1 || exit_code=$?

  if [[ $exit_code -eq 0 ]]; then
    ok "$label"
    [[ $VERBOSE -eq 1 ]] && cat "$tmpout"
  else
    fail "$label"
    cat "$tmpout"
  fi
  rm -f "$tmpout"
}

# ── 1. tools/ag ────────────────────────────────────────────────────────────────

section "Go — tools/ag"

if ! command -v go &>/dev/null; then
  fail "tools/ag — Go not found"
else
  AG="$REPO_ROOT/tools/ag"

  # Discover packages that have tests; strip the module prefix for the label.
  while IFS= read -r pkg; do
    label="${pkg#github.com/ags3d/ag/}"
    run_go_pkg "$AG" "$pkg" "tools/ag/$label"
  done < <(cd "$AG" && go list -f '{{if or .TestGoFiles .XTestGoFiles}}{{.ImportPath}}{{end}}' ./... 2>/dev/null)
fi

# ── 2. tools/agui ──────────────────────────────────────────────────────────────

section "Go — tools/agui"

if ! command -v go &>/dev/null; then
  fail "tools/agui — Go not found"
else
  AGUI="$REPO_ROOT/tools/agui"

  while IFS= read -r pkg; do
    label="${pkg#agui/}"
    run_go_pkg "$AGUI" "$pkg" "tools/agui/$label"
  done < <(cd "$AGUI" && go list -f '{{if or .TestGoFiles .XTestGoFiles}}{{.ImportPath}}{{end}}' ./... 2>/dev/null)
fi

# ── 3. AG Studio plugin tests (Godot headless, game_prototype project) ────────

section "GDScript — AG Studio plugin"

GODOT="$REPO_ROOT/bin/godot.linuxbsd.editor.x86_64"

if [[ $NO_GODOT -eq 1 ]]; then
  skip "plugin tests — skipped via --no-godot"
elif [[ ! -x "$GODOT" ]]; then
  skip "plugin tests — Godot binary not found"
else
  PLUGIN_NOISE="Godot Engine|godotengine\.org|nvidia|Gtk|Adwaita|Thread|libpulse|libvulkan|libVk|^Xlib|^$"
  PLUGIN_NOISE+="|^\s+at:|GDScript backtrace|^\s+\[|^\t|\[AGS/"
  PLUGIN_NOISE+="|AGSSpawnPoint.*not found in AGSRuntime"
  PLUGIN_NOISE+="|Source geometry parsing.*navigation mesh|visual meshes store geometry|For runtime.*baking navigation"
  PLUGIN_NOISE+="|RID allocations.*were leaked|ObjectDB instances were leaked|resources still in use"
  PLUGIN_NOISE+="|EditorPlugin|Invalid get index|Cannot call|method.*EditorInterface"
  PLUGIN_NOISE+="|RoomSync: root is not AGSRoom|RoomSync: scene has no file path"
  PLUGIN_NOISE+="|Parent path.*Player.*has vanished"

  rm -rf "$REPO_ROOT/game_prototype/.godot"
  run_godot "$GODOT" "$REPO_ROOT/game_prototype" "test_plugin.gd" \
    "$PLUGIN_NOISE" "GDScript/plugin (game_prototype)"
fi

# ── 4. GDScript tests (Godot headless) ────────────────────────────────────────

section "GDScript — agstests"

GODOT="$REPO_ROOT/bin/godot.linuxbsd.editor.x86_64"

if [[ $NO_GODOT -eq 1 ]]; then
  skip "agstests — skipped via --no-godot"
elif [[ ! -x "$GODOT" ]]; then
  skip "agstests — Godot binary not found"
else
  rm -rf "$REPO_ROOT/agstests/.godot"

  NOISE="Godot Engine|godotengine\.org|nvidia|Gtk|Adwaita|Thread|libpulse|libvulkan|libVk|^Xlib|^$"
  NOISE+="|^\s+at:|GDScript backtrace|^\s+\[|^\t"
  NOISE+="|AGSSpawnPoint.*not found in AGSRuntime"
  NOISE+="|Source geometry parsing.*navigation mesh|visual meshes store geometry|For runtime.*baking navigation"
  NOISE+="|Nonexistent function '(walk_to|face_to)' in base 'AGSCharacter'"
  NOISE+="|Nonexistent function 'connect' in base 'Callable'"
  NOISE+="|\[AGS/"
  NOISE+="|RID allocations.*were leaked|ObjectDB instances were leaked|resources still in use"

  run_godot "$GODOT" "$REPO_ROOT/agstests" "run_tests.gd" \
    "$NOISE" "GDScript/agstests"
fi

# ── summary ───────────────────────────────────────────────────────────────────

echo ""
echo "══════════════════════════════════════"
for s in "${PASS[@]+"${PASS[@]}"}"; do printf "  \033[32m✓\033[0m %s\n" "$s"; done
for s in "${FAIL[@]+"${FAIL[@]}"}"; do printf "  \033[31m✗\033[0m %s\n" "$s"; done
echo "══════════════════════════════════════"

if [[ ${#FAIL[@]} -gt 0 ]]; then
  echo "  FAILED (${#FAIL[@]} suite(s))"
  exit 1
else
  echo "  ALL PASSED (${#PASS[@]})"
  exit 0
fi
