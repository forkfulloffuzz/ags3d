#!/usr/bin/env bash
# Build the AGS3D Godot engine fork.
#
# Usage:
#   .dev/build.sh              # standard editor build
#   .dev/build.sh debug        # debug build (slower, more assertions)
#   .dev/build.sh release      # release template (stripped, no editor)
#   .dev/build.sh clean        # remove all compiled objects
#   .dev/build.sh -- ARGS      # pass extra scons args directly

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

MODE="${1:-}"
JOBS="${JOBS:-$(nproc)}"

case "$MODE" in
  clean)
    echo "→ Cleaning build artefacts…"
    scons --clean platform=linuxbsd
    exit 0
    ;;
  debug)
    TARGET=editor
    DEV_BUILD=yes
    echo "→ Building (debug, ${JOBS} jobs)…"
    ;;
  release)
    TARGET=template_release
    DEV_BUILD=no
    echo "→ Building release template (${JOBS} jobs)…"
    ;;
  --*)
    # Passthrough: .dev/build.sh -- extra=args …
    shift
    echo "→ Building with extra args: $*"
    scons platform=linuxbsd -j"$JOBS" "$@"
    exit 0
    ;;
  "")
    TARGET=editor
    DEV_BUILD=no
    echo "→ Building editor (${JOBS} jobs)…"
    ;;
  *)
    echo "Usage: .dev/build.sh [debug|release|clean|-- ARGS]" >&2
    exit 1
    ;;
esac

scons platform=linuxbsd target="${TARGET}" dev_build="${DEV_BUILD}" -j"$JOBS"

echo ""
echo "✓ Build complete: bin/godot.linuxbsd.editor.x86_64"
