# SPDX-License-Identifier: GPL-3.0-or-later
#
# AGS3D Import / Export Operators
#
# T-BL08: Import operator — reads an .agroom file and creates Blender empties
#         or box meshes for each gameplay element in an AGS_Gameplay collection.
#
# The .agroom format mirrors the Go parser (tools/ag/internal/room/room.go):
#   Room "name" {
#       initial_camera = "main"
#       Camera "main" { position = (x, y, z)  look_at = (x, y, z) }
#       Point "door" { position = (x, y, z) }
#       WalkableSurface "floor" { size = (w, d)  offset = (x, y, z) }
#       BlockerVolume "wall" { size = (x, y, z)  position = (x, y, z) }
#       SpawnPoint "start" { character = "player"  position = (x, y, z) }
#       Hotspot "shelf" { size = (x, y, z)  position = (x, y, z) }
#       TriggerRegion "door_zone" { size = (x, y, z)  position = (x, y, z) }
#   }
#
# Coordinate conversion: .agroom uses Godot Y-up left-handed coords;
# Blender uses Y-up right-handed. The minimal conversion needed to place
# objects visually correctly is to negate the Z axis:
#   blender_x = godot_x
#   blender_y = godot_y
#   blender_z = -godot_z
#
# Full round-trip coordinate conversion (T-BL05) will refine this further.

from __future__ import annotations

import re
import bpy
import bmesh
from mathutils import Vector

# ------------------------------------------------------------------ #
# .agroom mini-parser                                                  #
# ------------------------------------------------------------------ #

_VEC3_RE = re.compile(
    r"\(\s*([+-]?\d+(?:\.\d*)?)\s*,\s*([+-]?\d+(?:\.\d*)?)\s*,\s*([+-]?\d+(?:\.\d*)?)\s*\)"
)
_VEC2_RE = re.compile(
    r"\(\s*([+-]?\d+(?:\.\d*)?)\s*,\s*([+-]?\d+(?:\.\d*)?)\s*\)"
)
_STR_RE = re.compile(r'"([^"]*)"')


def _vec3(text: str) -> tuple[float, float, float] | None:
    m = _VEC3_RE.search(text)
    if m:
        return float(m.group(1)), float(m.group(2)), float(m.group(3))
    return None


def _vec2(text: str) -> tuple[float, float] | None:
    m = _VEC2_RE.search(text)
    if m:
        return float(m.group(1)), float(m.group(2))
    return None


def _str_val(text: str) -> str | None:
    m = _STR_RE.search(text)
    return m.group(1) if m else None


class _RoomData:
    """Minimal parsed representation of a .agroom file."""

    def __init__(self) -> None:
        self.name: str = ""
        self.cameras: list[dict] = []
        self.points: list[dict] = []
        self.walkable: list[dict] = []
        self.blockers: list[dict] = []
        self.spawns: list[dict] = []
        self.hotspots: list[dict] = []
        self.triggers: list[dict] = []


def _parse_agroom(text: str) -> _RoomData:
    """Parse .agroom text into a _RoomData.  Raises ValueError on bad syntax."""
    rd = _RoomData()

    # Tokenise into nested blocks by scanning char-by-char.
    lines = text.splitlines()
    # Strip // and # comments from each line.
    clean: list[str] = []
    for ln in lines:
        ln = re.sub(r"//.*", "", ln)
        ln = re.sub(r"#.*", "", ln)
        clean.append(ln)
    body = "\n".join(clean)

    # Extract top-level Room "name" { ... }
    m = re.search(r'\bRoom\s+"([^"]+)"\s*\{', body)
    if not m:
        raise ValueError(".agroom: missing 'Room \"name\" {' block")
    rd.name = m.group(1)
    # Find matching closing brace.
    room_body = _extract_block(body, m.end() - 1)

    # Walk block contents.
    pos = 0
    while pos < len(room_body):
        # Skip whitespace.
        ws = re.match(r"\s+", room_body[pos:])
        if ws:
            pos += ws.end()
            continue

        # initial_camera = "name"
        m2 = re.match(r'initial_camera\s*=\s*"([^"]*)"', room_body[pos:])
        if m2:
            pos += m2.end()
            continue

        # Named block: Type "name" { ... }
        m3 = re.match(r'(\w+)\s+"([^"]+)"\s*\{', room_body[pos:])
        if m3:
            btype = m3.group(1)
            bname = m3.group(2)
            block_start = pos + m3.end() - 1
            block_body = _extract_block(room_body, block_start)
            _parse_block(rd, btype, bname, block_body)
            pos = block_start + len(block_body) + 2  # skip body + braces
            continue

        # Skip anything else (unknown keys, etc.)
        pos += 1

    return rd


def _extract_block(text: str, brace_pos: int) -> str:
    """Extract the content between matching braces starting at brace_pos."""
    assert text[brace_pos] == "{", f"expected '{{' at {brace_pos}, got {text[brace_pos]!r}"
    depth = 0
    for i in range(brace_pos, len(text)):
        if text[i] == "{":
            depth += 1
        elif text[i] == "}":
            depth -= 1
            if depth == 0:
                return text[brace_pos + 1:i]
    raise ValueError("unmatched '{' in .agroom")


def _parse_block(rd: _RoomData, btype: str, bname: str, body: str) -> None:
    props: dict = {"name": bname}

    for line in body.splitlines():
        line = line.strip()
        if not line:
            continue

        if "position" in line:
            v = _vec3(line)
            if v:
                props["position"] = v
        elif "look_at" in line:
            v = _vec3(line)
            if v:
                props["look_at"] = v
        elif "size" in line:
            # Try vec3 first, then vec2.
            v3 = _vec3(line)
            if v3:
                props["size"] = v3
            else:
                v2 = _vec2(line)
                if v2:
                    props["size"] = (v2[0], 0.1, v2[1])  # WalkableSurface: thin slab
        elif "offset" in line:
            v = _vec3(line)
            if v:
                props["offset"] = v
        elif "character" in line:
            s = _str_val(line)
            if s:
                props["character"] = s

    dispatch = {
        "Camera":         rd.cameras,
        "Point":          rd.points,
        "WalkableSurface": rd.walkable,
        "BlockerVolume":  rd.blockers,
        "SpawnPoint":     rd.spawns,
        "Hotspot":        rd.hotspots,
        "TriggerRegion":  rd.triggers,
    }
    lst = dispatch.get(btype)
    if lst is not None:
        lst.append(props)


# ------------------------------------------------------------------ #
# Coordinate conversion: Godot → Blender                              #
# ------------------------------------------------------------------ #

def _godot_to_blender(gx: float, gy: float, gz: float) -> Vector:
    """Convert Godot (Y-up, left-handed) to Blender (Y-up, right-handed)."""
    return Vector((gx, gy, -gz))


# ------------------------------------------------------------------ #
# Blender object creation helpers                                      #
# ------------------------------------------------------------------ #

_AGS_COLLECTION = "AGS_Gameplay"

# AGS type → custom property value (matches panels.py AGS_TYPE_ITEMS ids).
_BLOCK_TYPE_TAG = {
    "cameras":  "CAMERA",
    "points":   "POINT",
    "walkable": "WALKABLE",
    "blockers": "BLOCKER",
    "spawns":   "SPAWN",
    "hotspots": "HOTSPOT",
    "triggers": "TRIGGER",
}


def _ensure_collection(scene: bpy.types.Scene) -> bpy.types.Collection:
    col = bpy.data.collections.get(_AGS_COLLECTION)
    if col is None:
        col = bpy.data.collections.new(_AGS_COLLECTION)
        scene.collection.children.link(col)
    return col


def _clear_collection(col: bpy.types.Collection) -> None:
    for obj in list(col.objects):
        bpy.data.objects.remove(obj, do_unlink=True)


def _make_empty(name: str, pos: Vector, display: str = "PLAIN_AXES") -> bpy.types.Object:
    obj = bpy.data.objects.new(name, None)
    obj.empty_display_type = display
    obj.location = pos
    return obj


def _make_box(name: str, pos: Vector, size: tuple[float, float, float]) -> bpy.types.Object:
    mesh = bpy.data.meshes.new(name + "_mesh")
    bm = bmesh.new()
    bmesh.ops.create_cube(bm, size=1.0)
    bm.to_mesh(mesh)
    bm.free()
    obj = bpy.data.objects.new(name, mesh)
    obj.location = pos
    obj.scale = (size[0], size[1], size[2])
    obj.display_type = "WIRE"
    return obj


def _tag(obj: bpy.types.Object, ags_type: str, ags_name: str, **extra: str) -> None:
    obj["AGS_type"] = ags_type
    obj["AGS_name"] = ags_name
    for k, v in extra.items():
        obj[k] = v


# ------------------------------------------------------------------ #
# Import operator                                                      #
# ------------------------------------------------------------------ #

class AGS3D_OT_ImportRoom(bpy.types.Operator):
    """Import an AGS3D .agroom file — creates gameplay objects in AGS_Gameplay collection"""

    bl_idname = "ags3d.import_room"
    bl_label = "AGS3D Room (.agroom)"
    bl_options = {"REGISTER", "UNDO"}

    filepath: bpy.props.StringProperty(
        name="File Path",
        subtype="FILE_PATH",
    )  # type: ignore[assignment]

    filter_glob: bpy.props.StringProperty(
        default="*.agroom",
        options={"HIDDEN"},
    )  # type: ignore[assignment]

    def invoke(self, context: bpy.types.Context, event: bpy.types.Event) -> set:
        context.window_manager.fileselect_add(self)
        return {"RUNNING_MODAL"}

    def execute(self, context: bpy.types.Context) -> set:
        try:
            text = open(self.filepath, encoding="utf-8").read()
        except OSError as exc:
            self.report({"ERROR"}, f"Cannot read file: {exc}")
            return {"CANCELLED"}

        try:
            rd = _parse_agroom(text)
        except ValueError as exc:
            self.report({"ERROR"}, f"Parse error: {exc}")
            return {"CANCELLED"}

        col = _ensure_collection(context.scene)
        _clear_collection(col)

        # Cameras → plain-axes empties with a look_at arrow (second empty).
        for cam in rd.cameras:
            pos = _godot_to_blender(*cam.get("position", (0, 0, 0)))
            obj = _make_empty(f"AGS_Cam_{cam['name']}", pos, "ARROWS")
            _tag(obj, "CAMERA", cam["name"])
            if "look_at" in cam:
                la = cam["look_at"]
                obj["AGS_look_at_pos"] = _godot_to_blender(*la).to_tuple()
            col.objects.link(obj)

        # Points → single_arrow empties.
        for pt in rd.points:
            pos = _godot_to_blender(*pt.get("position", (0, 0, 0)))
            obj = _make_empty(f"AGS_Point_{pt['name']}", pos, "SINGLE_ARROW")
            _tag(obj, "POINT", pt["name"])
            col.objects.link(obj)

        # WalkableSurface → thin wire box.
        for ws in rd.walkable:
            size = ws.get("size", (4.0, 0.1, 4.0))
            offset = ws.get("offset", (0.0, 0.0, 0.0))
            pos = _godot_to_blender(*offset)
            obj = _make_box(f"AGS_Walk_{ws['name']}", pos, size)
            _tag(obj, "WALKABLE", ws["name"])
            col.objects.link(obj)

        # BlockerVolume → wire box.
        for bv in rd.blockers:
            size = bv.get("size", (1.0, 1.0, 1.0))
            pos = _godot_to_blender(*bv.get("position", (0, 0, 0)))
            obj = _make_box(f"AGS_Blocker_{bv['name']}", pos, size)
            _tag(obj, "BLOCKER", bv["name"])
            col.objects.link(obj)

        # SpawnPoint → circle empty.
        for sp in rd.spawns:
            pos = _godot_to_blender(*sp.get("position", (0, 0, 0)))
            obj = _make_empty(f"AGS_Spawn_{sp['name']}", pos, "CIRCLE")
            _tag(obj, "SPAWN", sp["name"])
            if "character" in sp:
                obj["AGS_character"] = sp["character"]
            col.objects.link(obj)

        # Hotspot → wire box.
        for hs in rd.hotspots:
            size = hs.get("size", (1.0, 1.0, 1.0))
            pos = _godot_to_blender(*hs.get("position", (0, 0, 0)))
            obj = _make_box(f"AGS_Hotspot_{hs['name']}", pos, size)
            _tag(obj, "HOTSPOT", hs["name"])
            col.objects.link(obj)

        # TriggerRegion → wire box.
        for tr in rd.triggers:
            size = tr.get("size", (1.0, 1.0, 1.0))
            pos = _godot_to_blender(*tr.get("position", (0, 0, 0)))
            obj = _make_box(f"AGS_Trigger_{tr['name']}", pos, size)
            _tag(obj, "TRIGGER", tr["name"])
            col.objects.link(obj)

        total = (len(rd.cameras) + len(rd.points) + len(rd.walkable) +
                 len(rd.blockers) + len(rd.spawns) + len(rd.hotspots) +
                 len(rd.triggers))
        self.report({"INFO"}, f"AGS3D: imported {total} gameplay objects from '{rd.name}'")
        return {"FINISHED"}


# ------------------------------------------------------------------ #
# File → Import menu entry                                             #
# ------------------------------------------------------------------ #

def _menu_import(self: bpy.types.Menu, context: bpy.types.Context) -> None:
    self.layout.operator(AGS3D_OT_ImportRoom.bl_idname, text="AGS3D Room (.agroom)")


# ------------------------------------------------------------------ #
# Registration                                                         #
# ------------------------------------------------------------------ #

_classes = [AGS3D_OT_ImportRoom]


def register() -> None:
    for cls in _classes:
        bpy.utils.register_class(cls)
    bpy.types.TOPBAR_MT_file_import.append(_menu_import)


def unregister() -> None:
    bpy.types.TOPBAR_MT_file_import.remove(_menu_import)
    for cls in reversed(_classes):
        bpy.utils.unregister_class(cls)
