#!/usr/bin/env bash
# Wrapper for the ag project build tool.
# Automatically rebuilds the ag binary if source files are newer than the binary.
#
# Standard commands (forwarded to ag binary):
#   .dev/ag.sh build                    # parse changed .agscript files, emit GDScript
#   .dev/ag.sh run                      # build + launch Godot editor
#   .dev/ag.sh validate                 # static analysis
#   .dev/ag.sh export --platform linux  # export for platform
#   .dev/ag.sh new mygame               # scaffold new project
#   .dev/ag.sh viz tokens  <file>       # print token stream
#   .dev/ag.sh viz ast     <file>       # print AST tree (text)
#   .dev/ag.sh viz ast-dot <file>       # print AST as Graphviz DOT
#   .dev/ag.sh viz blocking <file>      # print blocking call annotations
#   .dev/ag.sh viz emit    <file>       # print side-by-side AGS-spirit ↔ GDScript
#   .dev/ag.sh viz         <file>       # run all viz stages
#
# Graphic visualisation shortcuts (require graphviz):
#   .dev/ag.sh viz-svg FILE [out.svg]   # render AST as SVG  (default: <file>.svg)
#   .dev/ag.sh viz-png FILE [out.png]   # render AST as PNG  (default: <file>.png)
#   .dev/ag.sh viz-pdf FILE [out.pdf]   # render AST as PDF  (default: <file>.pdf)
#   .dev/ag.sh viz-open FILE            # render AST as SVG and open in browser/viewer

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
AG_MODULE="$REPO_ROOT/tools/ag"
AG_BIN="$REPO_ROOT/bin/ag"

if [[ $# -eq 0 ]]; then
  echo "Usage: .dev/ag.sh COMMAND [args]" >&2
  echo "  build | run | validate | export --platform NAME | new NAME" >&2
  echo "  viz tokens FILE | viz ast FILE | viz ast-dot FILE | viz blocking FILE | viz emit FILE | viz FILE" >&2
  echo "  viz-svg FILE [out] | viz-png FILE [out] | viz-pdf FILE [out] | viz-open FILE" >&2
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
  # Run in a subshell so the cd does not change the caller's working directory.
  (cd "$AG_MODULE" && go build -o "$AG_BIN" ./cmd/ag)
fi

# -------------------------------------------------------------------
# Graphic visualisation shortcuts — intercepted before exec
# -------------------------------------------------------------------

# require_dot checks that graphviz is installed.
require_dot() {
  if ! command -v dot &>/dev/null; then
    echo "✗ graphviz not found — install it first:" >&2
    echo "    Ubuntu/Debian:  sudo apt install graphviz" >&2
    echo "    macOS:          brew install graphviz" >&2
    echo "    Arch:           sudo pacman -S graphviz" >&2
    exit 1
  fi
}

# default_out FILE EXT — derive output path: strip .agscript, add EXT.
default_out() {
  local file="$1" ext="$2"
  echo "${file%.agscript}.$ext"
}

case "${1}" in

  viz-svg)
    require_dot
    file="${2:?usage: .dev/ag.sh viz-svg FILE [out.svg]}"
    out="${3:-$(default_out "$file" svg)}"
    "$AG_BIN" viz ast-dot "$file" | dot -Tsvg -o "$out"
    echo "→ $out"
    exit 0
    ;;

  viz-png)
    require_dot
    file="${2:?usage: .dev/ag.sh viz-png FILE [out.png]}"
    out="${3:-$(default_out "$file" png)}"
    "$AG_BIN" viz ast-dot "$file" | dot -Tpng -o "$out"
    echo "→ $out"
    exit 0
    ;;

  viz-pdf)
    require_dot
    file="${2:?usage: .dev/ag.sh viz-pdf FILE [out.pdf]}"
    out="${3:-$(default_out "$file" pdf)}"
    "$AG_BIN" viz ast-dot "$file" | dot -Tpdf -o "$out"
    echo "→ $out"
    exit 0
    ;;

  viz-open)
    require_dot
    file="${2:?usage: .dev/ag.sh viz-open FILE}"
    tmp="$(mktemp --suffix=.svg)"
    "$AG_BIN" viz ast-dot "$file" | dot -Tsvg -o "$tmp"
    # Try common viewers in order.
    if command -v xdg-open &>/dev/null; then
      xdg-open "$tmp"
    elif command -v open &>/dev/null; then   # macOS
      open "$tmp"
    elif command -v firefox &>/dev/null; then
      firefox "$tmp"
    else
      echo "→ $tmp  (no viewer found — open manually)" >&2
    fi
    exit 0
    ;;

esac

exec "$AG_BIN" "$@"
