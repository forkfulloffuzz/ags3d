#!/usr/bin/env bash
# Wrapper for the ag project build tool.
# Automatically rebuilds the ag binary if source files are newer than the binary.
#
# Usage (mirrors ag directly):
#   .dev/ag.sh build                    # parse changed .agscript files, emit GDScript
#   .dev/ag.sh run                      # build + launch Godot editor
#   .dev/ag.sh validate                 # static analysis
#   .dev/ag.sh export --platform linux  # export for platform
#   .dev/ag.sh new mygame               # scaffold new project

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
AG_MODULE="$REPO_ROOT/tools/ag"
AG_BIN="$REPO_ROOT/bin/ag"

if [[ $# -eq 0 ]]; then
  echo "Usage: .dev/ag.sh COMMAND [args]" >&2
  echo "  build | run | validate | export --platform NAME | new NAME" >&2
  exit 1
fi

# Rebuild ag binary if it is missing or any source file is newer.
needs_build() {
  [[ ! -x "$AG_BIN" ]] && return 0
  while IFS= read -r -d '' src; do
    [[ "$src" -nt "$AG_BIN" ]] && return 0
  done < <(find "$AG_MODULE" -name '*.go' -print0)
  return 1
}

if needs_build; then
  if ! command -v go &>/dev/null; then
    echo "✗ Go not found — cannot build ag. Install from https://go.dev/dl/" >&2
    exit 1
  fi
  echo "→ rebuilding ag…"
  mkdir -p "$REPO_ROOT/bin"
  cd "$AG_MODULE"
  go build -o "$AG_BIN" ./cmd/ag
fi

exec "$AG_BIN" "$@"
