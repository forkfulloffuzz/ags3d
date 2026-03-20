#!/usr/bin/env bash
# Run the Go unit tests for tools/ag/ (transpiler, CLI, LSP).
#
# Usage:
#   .dev/test-ag.sh                   # run all packages
#   .dev/test-ag.sh --verbose         # show individual test names (go test -v)
#   .dev/test-ag.sh --filter scanner  # run tests matching pattern (go test -run)
#   .dev/test-ag.sh --cover           # generate coverage report
#
# Exit code: 0 = all pass, 1 = any failure.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
AG_MODULE="$REPO_ROOT/tools/ag"

if ! command -v go &>/dev/null; then
  echo "✗ Go not found. Install from https://go.dev/dl/ or: sudo apt install golang-go" >&2
  exit 1
fi

VERBOSE=0
FILTER=""
COVER=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --verbose|-v) VERBOSE=1 ;;
    --filter|-f)  FILTER="$2"; shift ;;
    --cover|-c)   COVER=1 ;;
    *)
      echo "Usage: .dev/test-ag.sh [--verbose] [--filter PATTERN] [--cover]" >&2
      exit 1
      ;;
  esac
  shift
done

cd "$AG_MODULE"

ARGS=("./...")
[[ $VERBOSE -eq 1 ]] && ARGS=("-v" "${ARGS[@]}")
[[ -n "$FILTER" ]]   && ARGS=("-run" "$FILTER" "${ARGS[@]}")
[[ $COVER -eq 1 ]]   && ARGS=("-coverprofile=$REPO_ROOT/bin/ag-cover.out" "${ARGS[@]}")

echo "→ go test ${ARGS[*]}"
go test "${ARGS[@]}"

if [[ $COVER -eq 1 ]]; then
  echo ""
  go tool cover -func="$REPO_ROOT/bin/ag-cover.out" | tail -1
  echo "  Full report: go tool cover -html=$REPO_ROOT/bin/ag-cover.out"
fi
