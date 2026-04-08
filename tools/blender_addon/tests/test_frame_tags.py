# SPDX-License-Identifier: GPL-3.0-or-later
#
# T-CUT27 — Frame tag export tests (headless, no bpy required).
#
# Uses duck-typed mock objects replicating the Blender armature/NLA/action/
# pose_marker interface so tests run without a Blender installation.
#
# Run with:
#   python3 -m unittest tools/blender_addon/tests/test_frame_tags.py -v
#   # or from repo root with pytest:
#   python3 -m pytest tools/blender_addon/tests/test_frame_tags.py -v

from __future__ import annotations

import json
import os
import sys
import tempfile
import unittest

# Make the parent package importable without Blender.
_PARENT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
if _PARENT not in sys.path:
    sys.path.insert(0, _PARENT)

from ags_frame_tags import (  # noqa: E402
    aganim_path_for_glb,
    build_aganim_data,
    collect_clip_frame_tags,
    export_aganim,
)


# ------------------------------------------------------------------ #
# Mock helpers                                                         #
# ------------------------------------------------------------------ #

class _Marker:
    def __init__(self, name: str, frame: int) -> None:
        self.name = name
        self.frame = frame


class _Action:
    def __init__(self, pose_markers: list[_Marker]) -> None:
        self.pose_markers = pose_markers


class _Strip:
    def __init__(self, action: "_Action | None") -> None:
        self.action = action


class _Track:
    def __init__(self, name: str, strips: "list[_Strip]") -> None:
        self.name = name
        self.strips = strips


class _AnimData:
    def __init__(self, nla_tracks: "list[_Track]") -> None:
        self.nla_tracks = nla_tracks


class _Armature:
    def __init__(self, name: str, nla_tracks: "list[_Track]") -> None:
        self.name = name
        self.animation_data = _AnimData(nla_tracks)


# ------------------------------------------------------------------ #
# Tests                                                                #
# ------------------------------------------------------------------ #

class TestCollectClipFrameTags(unittest.TestCase):

    def test_no_animation_data_returns_empty(self):
        class _Bare:
            animation_data = None
        self.assertEqual(collect_clip_frame_tags(_Bare()), [])

    def test_empty_nla_tracks(self):
        arm = _Armature("hero", [])
        self.assertEqual(collect_clip_frame_tags(arm), [])

    def test_track_with_no_strips_skipped(self):
        arm = _Armature("hero", [_Track("Idle", [])])
        self.assertEqual(collect_clip_frame_tags(arm), [])

    def test_track_with_no_pose_markers(self):
        action = _Action(pose_markers=[])
        track = _Track("Walk", [_Strip(action)])
        arm = _Armature("player", [track])
        clips = collect_clip_frame_tags(arm)
        self.assertEqual(len(clips), 1)
        self.assertEqual(clips[0]["name"], "Walk")
        self.assertEqual(clips[0]["frame_tags"], [])

    def test_single_marker_exported(self):
        action = _Action(pose_markers=[_Marker("footstep_left", 12)])
        track = _Track("Walk", [_Strip(action)])
        arm = _Armature("player", [track])
        clips = collect_clip_frame_tags(arm)
        self.assertEqual(clips[0]["frame_tags"], [{"name": "footstep_left", "frame": 12}])

    def test_multiple_markers_sorted_by_frame(self):
        markers = [_Marker("impact", 30), _Marker("anticipation", 5), _Marker("follow_through", 20)]
        action = _Action(pose_markers=markers)
        track = _Track("Attack", [_Strip(action)])
        arm = _Armature("guard", [track])
        clips = collect_clip_frame_tags(arm)
        tags = clips[0]["frame_tags"]
        frames = [t["frame"] for t in tags]
        self.assertEqual(frames, sorted(frames), "frame_tags should be sorted by frame ascending")
        self.assertEqual(tags[0], {"name": "anticipation", "frame": 5})
        self.assertEqual(tags[1], {"name": "follow_through", "frame": 20})
        self.assertEqual(tags[2], {"name": "impact", "frame": 30})

    def test_multiple_tracks_all_exported(self):
        arm = _Armature("player", [
            _Track("Walk", [_Strip(_Action([_Marker("step", 10)]))]),
            _Track("Idle", [_Strip(_Action([]))]),
        ])
        clips = collect_clip_frame_tags(arm)
        names = [c["name"] for c in clips]
        self.assertIn("Walk", names)
        self.assertIn("Idle", names)
        self.assertEqual(len(clips), 2)

    def test_strip_with_no_action_gives_empty_tags(self):
        track = _Track("Interact", [_Strip(action=None)])
        arm = _Armature("npc", [track])
        clips = collect_clip_frame_tags(arm)
        self.assertEqual(clips[0]["frame_tags"], [])

    def test_marker_with_empty_name_excluded(self):
        markers = [_Marker("", 5), _Marker("swing", 10)]
        action = _Action(pose_markers=markers)
        track = _Track("Attack", [_Strip(action)])
        arm = _Armature("warrior", [track])
        clips = collect_clip_frame_tags(arm)
        names = [t["name"] for t in clips[0]["frame_tags"]]
        self.assertNotIn("", names)
        self.assertIn("swing", names)


class TestBuildAganimData(unittest.TestCase):

    def test_structure(self):
        action = _Action(pose_markers=[_Marker("hit", 15)])
        track = _Track("Attack", [_Strip(action)])
        arm = _Armature("knight", [track])
        data = build_aganim_data("knight", arm)
        self.assertEqual(data["character"], "knight")
        self.assertEqual(len(data["clips"]), 1)
        self.assertEqual(data["clips"][0]["name"], "Attack")
        self.assertEqual(data["clips"][0]["frame_tags"], [{"name": "hit", "frame": 15}])


class TestExportAganim(unittest.TestCase):

    def test_writes_valid_json(self):
        action = _Action(pose_markers=[_Marker("footstep", 8)])
        track = _Track("Run", [_Strip(action)])
        arm = _Armature("runner", [track])
        with tempfile.TemporaryDirectory() as tmp:
            out_path = os.path.join(tmp, "runner.aganim")
            export_aganim("runner", arm, out_path)
            self.assertTrue(os.path.isfile(out_path))
            with open(out_path, encoding="utf-8") as fp:
                data = json.load(fp)
        self.assertEqual(data["character"], "runner")
        self.assertEqual(data["clips"][0]["name"], "Run")
        self.assertEqual(data["clips"][0]["frame_tags"][0], {"name": "footstep", "frame": 8})

    def test_creates_parent_directories(self):
        arm = _Armature("hero", [])
        with tempfile.TemporaryDirectory() as tmp:
            nested = os.path.join(tmp, "a", "b", "hero.aganim")
            export_aganim("hero", arm, nested)
            self.assertTrue(os.path.isfile(nested))


class TestAganimPathForGlb(unittest.TestCase):

    def test_replaces_extension(self):
        result = aganim_path_for_glb("characters/player/player.glb")
        self.assertEqual(result, "characters/player/player.aganim")

    def test_no_extension(self):
        self.assertEqual(aganim_path_for_glb("output"), "output.aganim")


if __name__ == "__main__":
    unittest.main()
