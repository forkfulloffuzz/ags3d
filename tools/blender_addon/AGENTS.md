# tools/blender_addon — Blender Addon Agent Instructions

The AGS3D Blender addon exports character assets for use in Godot projects.

## What it produces

| Output file | Format | Purpose |
|-------------|--------|---------|
| `<char>.glb` | GLTF binary | Mesh + skeleton + animations for Godot |
| `<char>.aganim` | JSON | Frame tags from Blender Action pose markers |

## Key files

| File | Notes |
|------|-------|
| `__init__.py` | Addon registration. Imports `bpy` — **cannot be imported outside Blender** |
| `char_operators.py` | Export operator. After GLTF export, calls `ags_frame_tags.export_aganim()` |
| `ags_frame_tags.py` | Frame tag extraction from NLA tracks. **No `bpy` dependency** — fully testable headlessly |
| `panels.py` | Blender sidebar panels |
| `operators.py` | Other operators (room export, etc.) |
| `coords.py` | Coordinate conversion helpers |

## .aganim format

```json
{
  "character": "hero",
  "clips": [
    {
      "name": "Walk",
      "frame_tags": [
        { "frame": 4, "name": "footstep_left" },
        { "frame": 10, "name": "footstep_right" }
      ]
    }
  ]
}
```

The `ag build` pipeline reads `.aganim` sidecars and injects `metadata/anim_frame_tags`
into the generated character `.tscn`. The runtime reads it via `AGSAnimationPlayerBase.get_frame_tag()`.

## Running tests

```sh
python3 tools/blender_addon/tests/test_frame_tags.py -v
```

**Do not** run `python3 -m unittest discover` from this directory — `__init__.py`
imports `bpy` and fails outside Blender.

## Headless test pattern

Tests use duck-typed mock objects in place of `bpy` types:

```python
import sys, os
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
import ags_frame_tags  # imports the module directly, bypasses __init__.py
```

Mock classes (`MockAction`, `MockStrip`, `MockNLATrack`, `MockObject`) replicate
the minimal interface used by `ags_frame_tags.py`.

## Adding tests

Add test files to `tests/`. Follow the direct-import pattern above. Use
`unittest.TestCase`. Run with `python3 tests/<file>.py -v`.
