# SPDX-License-Identifier: GPL-3.0-or-later
#
# AGS3D Frame Tag Exporter (T-CUT27)
#
# Reads Blender Action pose markers from every NLA track on an armature and
# writes a .aganim sidecar JSON file alongside the .glb export. The sidecar
# maps each animation clip name → list of (frame, tag_name) pairs.
#
# .aganim JSON format:
#   {
#     "character": "<char_name>",
#     "clips": [
#       {
#         "name": "<action_name>",
#         "frame_tags": [
#           {"name": "<tag_name>", "frame": <int>},
#           ...
#         ]
#       },
#       ...
#     ]
#   }
#
# Frame numbers are 1-based integers matching Blender's pose marker frames.
# frame_tags is sorted by frame ascending. Clips with no pose markers are
# included with an empty frame_tags list so ag build can confirm they exist.
#
# Usage (called from the character export operator):
#
#   from . import ags_frame_tags
#   ags_frame_tags.export_aganim(armature_obj, output_path)
#
# Headless test:
#   python -m pytest tools/blender_addon/tests/test_frame_tags.py

from __future__ import annotations

import json
import os
from typing import Any


# ------------------------------------------------------------------ #
# Core extraction — works on Blender objects (or duck-typed mocks)    #
# ------------------------------------------------------------------ #

def collect_clip_frame_tags(armature_obj: Any) -> list[dict]:
    """
    Return a list of clip dicts for every NLA track on *armature_obj* that
    has at least one strip.  Each dict has the form::

        {
            "name": "<track_name>",
            "frame_tags": [{"name": "<marker_name>", "frame": <int>}, ...]
        }

    Pose markers are taken from the Action bound to the *first* strip on
    each NLA track.  If the action has no pose markers the clip is included
    with an empty ``frame_tags`` list.

    If *armature_obj* has no ``animation_data`` or no NLA tracks this
    returns an empty list.
    """
    clips: list[dict] = []

    anim_data = getattr(armature_obj, "animation_data", None)
    if anim_data is None:
        return clips

    nla_tracks = getattr(anim_data, "nla_tracks", [])
    for track in nla_tracks:
        strips = list(getattr(track, "strips", []))
        if not strips:
            continue  # Track has no strips — skip.

        # Use the action from the first strip.
        action = getattr(strips[0], "action", None)
        frame_tags: list[dict] = []
        if action is not None:
            pose_markers = getattr(action, "pose_markers", [])
            for marker in pose_markers:
                name = getattr(marker, "name", "")
                frame = int(getattr(marker, "frame", 0))
                if name:
                    frame_tags.append({"name": name, "frame": frame})
            # Sort by frame ascending.
            frame_tags.sort(key=lambda m: m["frame"])

        clips.append({
            "name": track.name,
            "frame_tags": frame_tags,
        })

    return clips


def build_aganim_data(char_name: str, armature_obj: Any) -> dict:
    """
    Build the full .aganim dictionary for *armature_obj* with *char_name*.
    """
    return {
        "character": char_name,
        "clips": collect_clip_frame_tags(armature_obj),
    }


# ------------------------------------------------------------------ #
# File I/O                                                             #
# ------------------------------------------------------------------ #

def export_aganim(char_name: str, armature_obj: Any, output_path: str) -> None:
    """
    Export a .aganim sidecar for *armature_obj* to *output_path*.

    Creates parent directories as needed.  Silently overwrites existing
    files (the export operator always regenerates the sidecar).

    :param char_name:    Character name (written into the JSON ``character``
                         field); typically the Blender object name.
    :param armature_obj: The Blender Armature object (or a duck-typed mock
                         with the same interface).
    :param output_path:  Absolute or relative path for the .aganim file.
                         Should be ``<glb_path_without_ext>.aganim``.
    """
    data = build_aganim_data(char_name, armature_obj)
    out_dir = os.path.dirname(output_path)
    if out_dir:
        os.makedirs(out_dir, exist_ok=True)
    with open(output_path, "w", encoding="utf-8") as fp:
        json.dump(data, fp, indent=2)
        fp.write("\n")  # trailing newline


def aganim_path_for_glb(glb_path: str) -> str:
    """
    Return the .aganim sidecar path for a given .glb path.

    ``characters/player/player.glb`` → ``characters/player/player.aganim``
    """
    base, _ = os.path.splitext(glb_path)
    return base + ".aganim"
