# SPDX-License-Identifier: GPL-3.0-or-later
#
# AGS3D Blender Add-on
#
# Provides tools for tagging Blender objects as AGS3D gameplay elements
# (WalkableSurface, BlockerVolume, Points, Cameras, SpawnPoints, etc.) and
# exporting them to .agroom / .glb files for use in the AGS3D engine.
#
# Installation (Blender 4.2+):
#   Preferences → Extensions → Install from Disk → select the ags3d/ folder
#   (or a .zip of it).

bl_info = {
    "name": "AGS3D",
    "author": "AGS3D contributors",
    "version": (0, 1, 0),
    "blender": (4, 2, 0),
    "location": "Properties > Object > AGS3D | View3D > Sidebar > AGS3D",
    "description": "Adventure game development tools for the AGS3D engine",
    "category": "Game Engine",
}

from . import panels, overlay, operators, char_operators  # noqa: E402

_modules = [panels, overlay, operators, char_operators]


def register() -> None:
    for mod in _modules:
        mod.register()


def unregister() -> None:
    for mod in reversed(_modules):
        mod.unregister()
