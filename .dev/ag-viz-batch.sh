#!/usr/bin/env bash
# Batch visualisation — runs ag viz on every .agscript file matching a glob
# pattern and writes output to .output/ mirroring the source directory tree.
#
# Usage:
#   .dev/ag-viz-batch.sh STAGE GLOB [options]
#
# Stages and output formats:
#   tokens       each file → .output/<path>.tokens.txt
#   ast          each file → .output/<path>.ast.txt
#   ast-dot      each file → .output/<path>.dot
#   ast-svg      each file → .output/<path>.svg        (requires graphviz)
#   ast-png      each file → .output/<path>.png        (requires graphviz)
#   ast-pdf      each file → .output/<path>.pdf        (requires graphviz)
#   symbols      each file → .output/<path>.symbols.txt
#   symbols-dot  each file → .output/<path>.sym.dot
#   symbols-svg  each file → .output/<path>.sym.svg    (requires graphviz)
#   symbols-png  each file → .output/<path>.sym.png    (requires graphviz)
#   symbols-pdf  each file → .output/<path>.sym.pdf    (requires graphviz)
#
# Options:
#   --outdir DIR   write output under DIR instead of .output  (default: .output)
#   --jobs N       run N files in parallel                    (default: 1)
#   --quiet        suppress per-file progress lines
#
# Examples:
#   .dev/ag-viz-batch.sh tokens  "rooms/**/*.agscript"
#   .dev/ag-viz-batch.sh ast     "rooms/**/*.agscript"
#   .dev/ag-viz-batch.sh ast-dot "rooms/**/*.agscript"
#   .dev/ag-viz-batch.sh ast-svg "**/*.agscript" --outdir /tmp/ast-graphs
#   .dev/ag-viz-batch.sh ast-png "**/*.agscript" --jobs 4

set -euo pipefail
shopt -s globstar nullglob   # ** expansion + silence empty globs

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
AG_SH="$REPO_ROOT/.dev/ag.sh"

# -------------------------------------------------------------------
# Argument parsing
# -------------------------------------------------------------------

STAGE="${1:-}"
PATTERN="${2:-}"

if [[ -z "$STAGE" || -z "$PATTERN" ]]; then
  echo "Usage: .dev/ag-viz-batch.sh STAGE GLOB [--outdir DIR] [--jobs N] [--quiet]" >&2
  echo "Stages: tokens | ast | ast-dot | ast-svg | ast-png | ast-pdf" >&2
  exit 1
fi
shift 2

OUT_DIR=".output"
JOBS=1
QUIET=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --outdir) OUT_DIR="$2"; shift 2 ;;
    --jobs)   JOBS="$2";    shift 2 ;;
    --quiet)  QUIET=1;      shift ;;
    *) echo "Unknown option: $1" >&2; exit 1 ;;
  esac
done

# -------------------------------------------------------------------
# Stage → (ag subcommand, output extension, needs graphviz)
# -------------------------------------------------------------------

AG_STAGE=""
OUT_EXT=""
NEEDS_DOT=0

case "$STAGE" in
  tokens)      AG_STAGE="tokens";      OUT_EXT="tokens.txt"; NEEDS_DOT=0 ;;
  ast)         AG_STAGE="ast";         OUT_EXT="ast.txt";    NEEDS_DOT=0 ;;
  ast-dot)     AG_STAGE="ast-dot";     OUT_EXT="dot";        NEEDS_DOT=0 ;;
  ast-svg)     AG_STAGE="ast-dot";     OUT_EXT="svg";        NEEDS_DOT=1 ;;
  ast-png)     AG_STAGE="ast-dot";     OUT_EXT="png";        NEEDS_DOT=1 ;;
  ast-pdf)     AG_STAGE="ast-dot";     OUT_EXT="pdf";        NEEDS_DOT=1 ;;
  symbols)     AG_STAGE="symbols";     OUT_EXT="symbols.txt";NEEDS_DOT=0 ;;
  symbols-dot) AG_STAGE="symbols-dot"; OUT_EXT="sym.dot";    NEEDS_DOT=0 ;;
  symbols-svg) AG_STAGE="symbols-dot"; OUT_EXT="sym.svg";    NEEDS_DOT=1 ;;
  symbols-png) AG_STAGE="symbols-dot"; OUT_EXT="sym.png";    NEEDS_DOT=1 ;;
  symbols-pdf) AG_STAGE="symbols-dot"; OUT_EXT="sym.pdf";    NEEDS_DOT=1 ;;
  *)
    echo "Unknown stage: $STAGE" >&2
    echo "Valid stages: tokens | ast | ast-dot | ast-svg | ast-png | ast-pdf" >&2
    echo "              symbols | symbols-dot | symbols-svg | symbols-png | symbols-pdf" >&2
    exit 1
    ;;
esac

if [[ $NEEDS_DOT -eq 1 ]]; then
  if ! command -v dot &>/dev/null; then
    echo "✗ graphviz not found — install it to render $STAGE output:" >&2
    echo "    Ubuntu/Debian:  sudo apt install graphviz" >&2
    echo "    macOS:          brew install graphviz" >&2
    echo "    Arch:           sudo pacman -S graphviz" >&2
    exit 1
  fi
  DOT_FMT="${OUT_EXT}"   # svg / png / pdf
fi

# -------------------------------------------------------------------
# Expand glob
# -------------------------------------------------------------------

# eval the pattern so ** and ? expand from the caller's working directory.
mapfile -d '' FILES < <(eval "printf '%s\0' $PATTERN" 2>/dev/null || true)

if [[ ${#FILES[@]} -eq 0 ]]; then
  echo "ag-viz-batch: no files matched: $PATTERN" >&2
  exit 1
fi

# -------------------------------------------------------------------
# Process one file — called directly or via xargs -P
# -------------------------------------------------------------------

process_file() {
  local src="$1"
  local ag_stage="$2"
  local out_ext="$3"
  local out_dir="$4"
  local needs_dot="$5"
  local dot_fmt="${6:-}"
  local quiet="$7"

  # Strip leading ./ if present.
  src="${src#./}"

  # Derive output path: <out_dir>/<src-without-agscript-ext>.<out_ext>
  local base="${src%.agscript}"
  local out_file="$out_dir/$base.$out_ext"
  local out_parent
  out_parent="$(dirname "$out_file")"
  mkdir -p "$out_parent"

  if [[ $needs_dot -eq 1 ]]; then
    "$AG_SH" viz ast-dot "$src" | dot -T"$dot_fmt" -o "$out_file"
  else
    "$AG_SH" viz "$ag_stage" "$src" > "$out_file"
  fi

  if [[ $quiet -eq 0 ]]; then
    echo "  $src → $out_file"
  fi
}
export -f process_file

# -------------------------------------------------------------------
# Run
# -------------------------------------------------------------------

[[ $QUIET -eq 0 ]] && echo "ag-viz-batch: stage=$STAGE  files=${#FILES[@]}  outdir=$OUT_DIR"

if [[ $JOBS -gt 1 ]]; then
  printf '%s\0' "${FILES[@]}" \
    | xargs -0 -P "$JOBS" -I{} bash -c \
        'process_file "$@"' _ \
        {} "$AG_STAGE" "$OUT_EXT" "$OUT_DIR" "$NEEDS_DOT" "$DOT_FMT" "$QUIET"
else
  for src in "${FILES[@]}"; do
    process_file "$src" "$AG_STAGE" "$OUT_EXT" "$OUT_DIR" "$NEEDS_DOT" "${DOT_FMT:-}" "$QUIET"
  done
fi

[[ $QUIET -eq 0 ]] && echo "ag-viz-batch: done → $OUT_DIR"
