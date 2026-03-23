#!/usr/bin/env bash
# Run every test suite across the AGS3D project.
#
# Suites:
#   1. tools/ag/     — Go tests (transpiler, CLI, LSP)
#   2. tools/agui/   — Go tests (AG Studio backend)
#   3. agstests/     — GDScript tests run headlessly via Godot
#                      (skipped with a warning if Godot binary is absent)
#
# Usage:
#   .dev/test-all.sh               # run everything
#   .dev/test-all.sh --verbose     # pass -v to go test; show raw Godot output
#   .dev/test-all.sh --filter FOO  # filter test names / Godot output
#   .dev/test-all.sh --no-godot    # skip GDScript suite even if Godot is present
#
# Exit code: 0 = all suites passed, 1 = any failure.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

VERBOSE=0
FILTER=""
NO_GODOT=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --verbose|-v)   VERBOSE=1 ;;
    --filter|-f)    FILTER="$2"; shift ;;
    --no-godot)     NO_GODOT=1 ;;
    *)
      echo "Usage: .dev/test-all.sh [--verbose] [--filter PATTERN] [--no-godot]" >&2
      exit 1
      ;;
  esac
  shift
done

PASS=()
FAIL=()

# ── helpers ────────────────────────────────────────────────────────────────────

section() { echo ""; echo "══════════════════════════════════════"; echo "  $*"; echo "══════════════════════════════════════"; }
ok()      { PASS+=("$*"); echo "✓ $*"; }
fail()    { FAIL+=("$*"); echo "✗ $*"; }

# ── 1. tools/ag Go tests ───────────────────────────────────────────────────────

section "Go — tools/ag (transpiler / CLI / LSP)"

if ! command -v go &>/dev/null; then
  fail "tools/ag — Go not found, skipping"
else
  GO_ARGS=("./...")
  [[ $VERBOSE -eq 1 ]] && GO_ARGS=("-v" "${GO_ARGS[@]}")
  [[ -n "$FILTER" ]]   && GO_ARGS=("-run" "$FILTER" "${GO_ARGS[@]}")

  if (cd "$REPO_ROOT/tools/ag" && go test "${GO_ARGS[@]}"); then
    ok "tools/ag"
  else
    fail "tools/ag"
  fi
fi

# ── 2. tools/agui Go tests ─────────────────────────────────────────────────────

section "Go — tools/agui (AG Studio backend)"

if ! command -v go &>/dev/null; then
  fail "tools/agui — Go not found, skipping"
else
  GO_ARGS=("./...")
  [[ $VERBOSE -eq 1 ]] && GO_ARGS=("-v" "${GO_ARGS[@]}")
  [[ -n "$FILTER" ]]   && GO_ARGS=("-run" "$FILTER" "${GO_ARGS[@]}")

  if (cd "$REPO_ROOT/tools/agui" && go test "${GO_ARGS[@]}"); then
    ok "tools/agui"
  else
    fail "tools/agui"
  fi
fi

# ── 3. GDScript tests (Godot headless) ────────────────────────────────────────

section "GDScript — agstests (Godot headless)"

GODOT="$REPO_ROOT/bin/godot.linuxbsd.editor.x86_64"

if [[ $NO_GODOT -eq 1 ]]; then
  echo "  (skipped via --no-godot)"
elif [[ ! -x "$GODOT" ]]; then
  echo "  ⚠ Godot binary not found at bin/godot.linuxbsd.editor.x86_64"
  echo "    Run .dev/build.sh to build it, or pass --no-godot to skip."
else
  rm -rf "$REPO_ROOT/agstests/.godot"

  NOISE="Godot Engine|godotengine\.org|nvidia|Gtk|Adwaita|Thread|libpulse|libvulkan|libVk|^Xlib|^$"
  TMPOUT="$(mktemp)"

  if [[ $VERBOSE -eq 1 ]]; then
    "$GODOT" --headless --path "$REPO_ROOT/agstests" --script run_tests.gd
    GDX=$?
  else
    "$GODOT" --headless --path "$REPO_ROOT/agstests" --script run_tests.gd >"$TMPOUT" 2>&1
    GDX=$?
    if [[ -n "$FILTER" ]]; then
      grep -v -E "$NOISE" "$TMPOUT" | grep -i "$FILTER" || true
    else
      grep -v -E "$NOISE" "$TMPOUT" || true
    fi
  fi

  rm -f "$TMPOUT"

  if [[ $GDX -eq 0 ]]; then
    ok "agstests (GDScript)"
  else
    fail "agstests (GDScript)"
  fi
fi

# ── summary ───────────────────────────────────────────────────────────────────

echo ""
echo "══════════════════════════════════════"
for s in "${PASS[@]+"${PASS[@]}"}"; do echo "  ✓ $s"; done
for s in "${FAIL[@]+"${FAIL[@]}"}"; do echo "  ✗ $s"; done
echo "══════════════════════════════════════"

if [[ ${#FAIL[@]} -gt 0 ]]; then
  echo "  FAILED (${#FAIL[@]} suite(s))"
  exit 1
else
  echo "  ALL SUITES PASSED (${#PASS[@]})"
  exit 0
fi
