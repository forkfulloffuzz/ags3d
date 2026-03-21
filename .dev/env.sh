#!/usr/bin/env bash
# AGS3D development environment setup.
# SOURCE this file — do not execute it directly.
#
# Usage:
#   source .dev/env.sh
#   . .dev/env.sh

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
  echo "Error: env.sh must be sourced, not executed." >&2
  echo "  Run:  source .dev/env.sh" >&2
  exit 1
fi

AGS3D_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

export AGS3D_ROOT
export GODOT="$AGS3D_ROOT/bin/godot.linuxbsd.editor.x86_64"

# Add repo bin/ to PATH (ag, agls, godot binary)
case ":$PATH:" in
  *":$AGS3D_ROOT/bin:"*) ;;
  *) export PATH="$AGS3D_ROOT/bin:$PATH" ;;
esac

echo "AGS3D_ROOT  = $AGS3D_ROOT"
echo "GODOT       = $GODOT"
echo "PATH        → bin/ prepended (ag, agls, godot available directly)"
