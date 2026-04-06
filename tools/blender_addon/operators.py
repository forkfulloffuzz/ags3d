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
    props: dict = {"name": bname, "_type": btype}

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
# Coordinate conversion: Blender → Godot (delegates to coords.py)    #
# ------------------------------------------------------------------ #

def _blender_to_godot(bx: float, by: float, bz: float) -> tuple[float, float, float]:
    """Convert Blender world-space position to Godot world-space position."""
    from . import coords as _coords
    from mathutils import Vector as _V
    return _coords.blender_to_godot_pos(_V((bx, by, bz)))


def _bbox_size_godot(obj: bpy.types.Object) -> tuple[float, float, float]:
    """Return the object's world-space bounding box size in Godot coordinates."""
    from . import coords as _coords
    return _coords.bbox_size_godot(obj)


def _bbox_center_godot(obj: bpy.types.Object) -> tuple[float, float, float]:
    """Return the object's world-space bounding box centre in Godot coordinates."""
    from . import coords as _coords
    return _coords.bbox_center_godot(obj)


# ------------------------------------------------------------------ #
# .agroom writer                                                       #
# ------------------------------------------------------------------ #

def _fmt(v: float) -> str:
    """Format a float for .agroom output (trim trailing zeros)."""
    s = f"{v:.4f}".rstrip("0").rstrip(".")
    return s if s else "0"


def _append_existing_blocks(
    lines: list[str],
    existing_list: list[dict],
    bl_names: set[str],
) -> None:
    """Append blocks from *existing_list* whose name is NOT in *bl_names*.

    Used for merge mode (T-BL07): preserves blocks that exist in the .agroom
    file but have no corresponding tagged object in the Blender scene.
    """
    for blk in existing_list:
        name = blk.get("name", "")
        if name in bl_names:
            continue  # covered by Blender; already written
        # Re-serialise the parsed block verbatim.
        btype = blk.get("_type", "")
        lines.append(f'    {btype} "{name}" {{')
        if "position" in blk:
            p = blk["position"]
            lines.append(f'        position = ({_fmt(p[0])}, {_fmt(p[1])}, {_fmt(p[2])})')
        if "look_at" in blk:
            la = blk["look_at"]
            lines.append(f'        look_at  = ({_fmt(la[0])}, {_fmt(la[1])}, {_fmt(la[2])})')
        if "size" in blk:
            s = blk["size"]
            if btype == "WalkableSurface":
                lines.append(f'        size   = ({_fmt(s[0])}, {_fmt(s[2])})')
            else:
                lines.append(f'        size     = ({_fmt(s[0])}, {_fmt(s[1])}, {_fmt(s[2])})')
        if "offset" in blk:
            o = blk["offset"]
            lines.append(f'        offset = ({_fmt(o[0])}, {_fmt(o[1])}, {_fmt(o[2])})')
        if "character" in blk:
            lines.append(f'        character = "{blk["character"]}"')
        lines.append("    }")
        lines.append("")


def _write_agroom(
    room_name: str,
    objects: list[bpy.types.Object],
    existing: "_RoomData | None" = None,
) -> str:
    """Serialise AGS-tagged Blender objects to .agroom text.

    If *existing* is supplied (merge mode, T-BL07), blocks present in the
    existing file but absent from the Blender scene are preserved verbatim.
    Geometry-derived fields (position, size, look_at) for blocks that DO
    appear in Blender are always overwritten.
    """
    from .panels import _get_ags_type

    lines: list[str] = [f'Room "{room_name}" {{']

    # Collect by type.
    cameras: list[bpy.types.Object] = []
    points: list[bpy.types.Object] = []
    walkable: list[bpy.types.Object] = []
    blockers: list[bpy.types.Object] = []
    spawns: list[bpy.types.Object] = []
    hotspots: list[bpy.types.Object] = []
    triggers: list[bpy.types.Object] = []

    for obj in objects:
        t = _get_ags_type(obj)
        if t == "CAMERA":
            cameras.append(obj)
        elif t == "POINT":
            points.append(obj)
        elif t == "WALKABLE":
            walkable.append(obj)
        elif t == "BLOCKER":
            blockers.append(obj)
        elif t == "SPAWN":
            spawns.append(obj)
        elif t == "HOTSPOT":
            hotspots.append(obj)
        elif t == "TRIGGER":
            triggers.append(obj)

    # initial_camera = first camera's name.
    if cameras:
        first_cam_name = cameras[0].get("AGS_name", cameras[0].name)
        lines.append(f'    initial_camera = "{first_cam_name}"')
        lines.append("")

    # Cameras
    cam_names: set[str] = set()
    for obj in cameras:
        name = obj.get("AGS_name", obj.name)
        cam_names.add(name)
        bx, by, bz = obj.matrix_world.translation
        gx, gy, gz = _blender_to_godot(bx, by, bz)
        # look_at: only written when user explicitly set an AGS_look_at target
        # Empty in Blender. When absent, ag build falls back to auto-look-at
        # (floor centre). The camera's position/rotation in Blender is the
        # author's responsibility; we never auto-compute from the forward vector.
        look_at_name = obj.get("AGS_look_at", "")
        target = bpy.data.objects.get(look_at_name) if look_at_name else None
        lines.append(f'    Camera "{name}" {{')
        lines.append(f'        position = ({_fmt(gx)}, {_fmt(gy)}, {_fmt(gz)})')
        if target is not None:
            tx, ty, tz = target.matrix_world.translation
            lax, lay, laz = _blender_to_godot(tx, ty, tz)
            lines.append(f'        look_at  = ({_fmt(lax)}, {_fmt(lay)}, {_fmt(laz)})')
        lines.append("    }")
        lines.append("")
    if existing:
        _append_existing_blocks(lines, existing.cameras, cam_names)

    # Points
    pt_names: set[str] = set()
    for obj in points:
        name = obj.get("AGS_name", obj.name)
        pt_names.add(name)
        bx, by, bz = obj.matrix_world.translation
        gx, gy, gz = _blender_to_godot(bx, by, bz)
        lines.append(f'    Point "{name}" {{')
        lines.append(f'        position = ({_fmt(gx)}, {_fmt(gy)}, {_fmt(gz)})')
        lines.append("    }")
        lines.append("")
    if existing:
        _append_existing_blocks(lines, existing.points, pt_names)

    # WalkableSurface
    wk_names: set[str] = set()
    for obj in walkable:
        name = obj.get("AGS_name", obj.name)
        wk_names.add(name)
        gsx, _gsy, gsz = _bbox_size_godot(obj)  # Godot XZ plane for WalkableSurface
        gcx, gcy, gcz = _bbox_center_godot(obj)
        lines.append(f'    WalkableSurface "{name}" {{')
        lines.append(f'        size   = ({_fmt(gsx)}, {_fmt(gsz)})')  # XZ plane (no Y)
        lines.append(f'        offset = ({_fmt(gcx)}, {_fmt(gcy)}, {_fmt(gcz)})')
        lines.append("    }")
        lines.append("")
    if existing:
        _append_existing_blocks(lines, existing.walkable, wk_names)

    # BlockerVolume
    bl_names: set[str] = set()
    for obj in blockers:
        name = obj.get("AGS_name", obj.name)
        bl_names.add(name)
        gsx, gsy, gsz = _bbox_size_godot(obj)
        gcx, gcy, gcz = _bbox_center_godot(obj)
        lines.append(f'    BlockerVolume "{name}" {{')
        lines.append(f'        size     = ({_fmt(gsx)}, {_fmt(gsy)}, {_fmt(gsz)})')
        lines.append(f'        position = ({_fmt(gcx)}, {_fmt(gcy)}, {_fmt(gcz)})')
        lines.append("    }")
        lines.append("")
    if existing:
        _append_existing_blocks(lines, existing.blockers, bl_names)

    # SpawnPoint
    sp_names: set[str] = set()
    for obj in spawns:
        name = obj.get("AGS_name", obj.name)
        sp_names.add(name)
        bx, by, bz = obj.matrix_world.translation
        gx, gy, gz = _blender_to_godot(bx, by, bz)
        char = obj.get("AGS_character", "")
        lines.append(f'    SpawnPoint "{name}" {{')
        if char:
            lines.append(f'        character = "{char}"')
        lines.append(f'        position  = ({_fmt(gx)}, {_fmt(gy)}, {_fmt(gz)})')
        lines.append("    }")
        lines.append("")
    if existing:
        _append_existing_blocks(lines, existing.spawns, sp_names)

    # Hotspot
    hs_names: set[str] = set()
    for obj in hotspots:
        name = obj.get("AGS_name", obj.name)
        hs_names.add(name)
        gsx, gsy, gsz = _bbox_size_godot(obj)
        gcx, gcy, gcz = _bbox_center_godot(obj)
        lines.append(f'    Hotspot "{name}" {{')
        lines.append(f'        size     = ({_fmt(gsx)}, {_fmt(gsy)}, {_fmt(gsz)})')
        lines.append(f'        position = ({_fmt(gcx)}, {_fmt(gcy)}, {_fmt(gcz)})')
        lines.append("    }")
        lines.append("")
    if existing:
        _append_existing_blocks(lines, existing.hotspots, hs_names)

    # TriggerRegion
    tr_names: set[str] = set()
    for obj in triggers:
        name = obj.get("AGS_name", obj.name)
        tr_names.add(name)
        gsx, gsy, gsz = _bbox_size_godot(obj)
        gcx, gcy, gcz = _bbox_center_godot(obj)
        lines.append(f'    TriggerRegion "{name}" {{')
        lines.append(f'        size     = ({_fmt(gsx)}, {_fmt(gsy)}, {_fmt(gsz)})')
        lines.append(f'        position = ({_fmt(gcx)}, {_fmt(gcy)}, {_fmt(gcz)})')
        lines.append("    }")
        lines.append("")
    if existing:
        _append_existing_blocks(lines, existing.triggers, tr_names)

    lines.append("}")
    return "\n".join(lines) + "\n"


# ------------------------------------------------------------------ #
# NavMesh baking (T-BL09)                                              #
# ------------------------------------------------------------------ #

_NAVMESH_OBJ_NAME = "AGS_NavMesh"
_NAVMESH_COLLECTION = "AGS_NavMesh"


def _top_quad_verts(obj: bpy.types.Object) -> list:
    """Return the 4 world-space corners of the top XZ face of *obj*'s bounding box."""
    from mathutils import Vector as _V
    corners = [obj.matrix_world @ _V(v) for v in obj.bound_box]
    xs = [c.x for c in corners]
    ys = [c.y for c in corners]
    zs = [c.z for c in corners]
    x_min, x_max = min(xs), max(xs)
    y_top = max(ys)
    z_min, z_max = min(zs), max(zs)
    return [
        _V((x_min, y_top, z_min)),
        _V((x_max, y_top, z_min)),
        _V((x_max, y_top, z_max)),
        _V((x_min, y_top, z_max)),
    ]


def _count_navmesh_islands(bm: bmesh.types.BMesh) -> int:
    """Count disconnected face islands in *bm* via flood-fill over shared edges."""
    visited: set = set()
    islands = 0
    for start_face in bm.faces:
        if start_face in visited:
            continue
        islands += 1
        stack = [start_face]
        while stack:
            face = stack.pop()
            if face in visited:
                continue
            visited.add(face)
            for edge in face.edges:
                for linked in edge.link_faces:
                    if linked not in visited:
                        stack.append(linked)
    return islands


def _bake_navmesh(scene: bpy.types.Scene) -> "tuple[bpy.types.Object | None, int]":
    """Bake a navigation mesh from all WalkableSurface objects in *scene*.

    Creates (or replaces) a single Blender mesh object named "AGS_NavMesh"
    containing one quad per WalkableSurface.  The object is tagged with
    AGS_type = "NAVMESH" so Godot's importer can identify it.

    Returns (nav_obj, island_count).  nav_obj is None if no WalkableSurface
    objects exist.  island_count > 1 means the navmesh has disconnected
    islands (a warning, not an error — WalkTo() navigates to the closest
    reachable point on the current island).
    """
    from .panels import _get_ags_type

    walkable = [
        obj for obj in scene.objects
        if _get_ags_type(obj) == "WALKABLE" and obj.type == "MESH"
    ]
    if not walkable:
        return None, 0

    # Remove stale AGS_NavMesh object/mesh if present.
    old_obj = bpy.data.objects.get(_NAVMESH_OBJ_NAME)
    if old_obj is not None:
        old_mesh = old_obj.data
        bpy.data.objects.remove(old_obj, do_unlink=True)
        if old_mesh and old_mesh.users == 0:
            bpy.data.meshes.remove(old_mesh)

    # Build combined mesh from top quads of all WalkableSurface objects.
    mesh = bpy.data.meshes.new(_NAVMESH_OBJ_NAME)
    bm = bmesh.new()
    for ws_obj in walkable:
        verts = _top_quad_verts(ws_obj)
        bm_verts = [bm.verts.new(v) for v in verts]
        bm.faces.new(bm_verts)

    island_count = _count_navmesh_islands(bm)
    bm.to_mesh(mesh)
    bm.free()

    # Create object and tag it.
    nav_obj = bpy.data.objects.new(_NAVMESH_OBJ_NAME, mesh)
    nav_obj["AGS_type"] = "NAVMESH"
    nav_obj["AGS_name"] = _NAVMESH_OBJ_NAME

    # Link to a dedicated collection (create if needed).
    col = bpy.data.collections.get(_NAVMESH_COLLECTION)
    if col is None:
        col = bpy.data.collections.new(_NAVMESH_COLLECTION)
        scene.collection.children.link(col)
    col.objects.link(nav_obj)

    return nav_obj, island_count


class AGS3D_OT_BakeNavMesh(bpy.types.Operator):
    """Bake a navigation mesh from all WalkableSurface objects in the scene.

    The result is stored as a mesh object named AGS_NavMesh in the AGS_NavMesh
    collection.  It is included automatically when exporting a room (.glb).
    """

    bl_idname = "ags3d.bake_navmesh"
    bl_label = "Bake NavMesh"
    bl_options = {"REGISTER", "UNDO"}

    def execute(self, context: bpy.types.Context) -> set:
        nav_obj, islands = _bake_navmesh(context.scene)
        if nav_obj is None:
            self.report({"WARNING"}, "AGS3D: no WalkableSurface objects found — nothing baked")
            return {"CANCELLED"}
        if islands > 1:
            self.report(
                {"WARNING"},
                f"AGS3D: NavMesh has {islands} disconnected islands. "
                "WalkTo() will navigate to the closest reachable point — "
                "characters cannot cross gaps unless custom movement is used.",
            )
        else:
            self.report({"INFO"}, f"AGS3D: baked {_NAVMESH_OBJ_NAME} from WalkableSurface objects")
        return {"FINISHED"}


# ------------------------------------------------------------------ #
# Export operator (T-BL04)                                             #
# ------------------------------------------------------------------ #

# Types excluded from visual GLTF export (pure gameplay descriptors).
# NAVMESH is exported separately as part of the visual mesh (carries geo).
_GAMEPLAY_ONLY = {"WALKABLE", "BLOCKER", "HOTSPOT", "TRIGGER", "SPAWN"}


class AGS3D_OT_ExportRoom(bpy.types.Operator):
    """Export AGS3D room — gameplay objects → .agroom, visual mesh → .glb"""

    bl_idname = "ags3d.export_room"
    bl_label = "AGS3D Room (.agroom + .glb)"
    bl_options = {"REGISTER"}

    filepath: bpy.props.StringProperty(
        name=".agroom File Path",
        subtype="FILE_PATH",
    )  # type: ignore[assignment]

    filter_glob: bpy.props.StringProperty(
        default="*.agroom",
        options={"HIDDEN"},
    )  # type: ignore[assignment]

    room_name: bpy.props.StringProperty(
        name="Room Name",
        description="Written to the Room block; defaults to blend file stem",
        default="",
    )  # type: ignore[assignment]

    export_visual: bpy.props.BoolProperty(
        name="Export Visual Mesh (.glb)",
        description="Call the GLTF exporter for non-gameplay objects",
        default=True,
    )  # type: ignore[assignment]

    export_gameplay: bpy.props.BoolProperty(
        name="Export Gameplay Data (.agroom)",
        description="Write .agroom from AGS-tagged objects",
        default=True,
    )  # type: ignore[assignment]

    merge_mode: bpy.props.BoolProperty(
        name="Merge (preserve non-geometry fields)",
        description=(
            "Read the existing .agroom before exporting and preserve any blocks "
            "that have no corresponding tagged Blender object (non-geometry fields "
            "edited manually are kept; geometry-derived fields are overwritten)"
        ),
        default=True,
    )  # type: ignore[assignment]

    def invoke(self, context: bpy.types.Context, event: bpy.types.Event) -> set:
        # Pre-fill filepath from blend file name.
        blend = bpy.data.filepath
        if blend:
            import os as _os
            stem = _os.path.splitext(_os.path.basename(blend))[0]
            self.filepath = _os.path.join(_os.path.dirname(blend), stem + ".agroom")
            if not self.room_name:
                self.room_name = stem
        context.window_manager.fileselect_add(self)
        return {"RUNNING_MODAL"}

    def execute(self, context: bpy.types.Context) -> set:
        import os as _os

        agroom_path = self.filepath
        if not agroom_path.endswith(".agroom"):
            agroom_path += ".agroom"
        glb_path = agroom_path.replace(".agroom", ".glb")

        room_name = self.room_name or _os.path.splitext(_os.path.basename(agroom_path))[0]

        all_objects = list(context.scene.objects)

        # --- write .agroom ---
        if self.export_gameplay:
            # Merge mode: read and parse the existing file (if any) so that
            # blocks absent from Blender are preserved verbatim (T-BL07).
            existing: "_RoomData | None" = None
            if self.merge_mode and _os.path.isfile(agroom_path):
                try:
                    with open(agroom_path, "r", encoding="utf-8") as fh:
                        existing = _parse_agroom(fh.read())
                except (OSError, ValueError):
                    existing = None  # parse failure → full overwrite

            agroom_text = _write_agroom(room_name, all_objects, existing)
            try:
                with open(agroom_path, "w", encoding="utf-8") as fh:
                    fh.write(agroom_text)
            except OSError as exc:
                self.report({"ERROR"}, f"Cannot write .agroom: {exc}")
                return {"CANCELLED"}

        # --- export .glb (visual objects + baked navmesh) ---
        if self.export_visual:
            from .panels import _get_ags_type

            # Auto-bake navmesh from WalkableSurface objects before export.
            nav_obj, nav_islands = _bake_navmesh(context.scene)
            if nav_islands > 1:
                self.report(
                    {"WARNING"},
                    f"AGS3D: NavMesh has {nav_islands} disconnected islands. "
                    "WalkTo() navigates to the closest reachable point — "
                    "characters cannot cross gaps unless custom movement is used.",
                )

            # Select only visual objects (non-gameplay-only, non-camera, non-none-type)
            # PLUS the freshly baked NavMesh (AGS_type = "NAVMESH").
            all_objects_now = list(context.scene.objects)
            visual_objects = [
                obj for obj in all_objects_now
                if (_get_ags_type(obj) not in _GAMEPLAY_ONLY
                    and _get_ags_type(obj) != "CAMERA"
                    and obj.type in {"MESH", "LIGHT", "EMPTY"}
                    and _get_ags_type(obj) != "POINT"
                    and _get_ags_type(obj) != "SPAWN")
                or _get_ags_type(obj) == "NAVMESH"
            ]

            prev_selected = list(context.selected_objects)
            prev_active = context.active_object
            bpy.ops.object.select_all(action="DESELECT")
            for obj in visual_objects:
                obj.select_set(True)
            if visual_objects:
                context.view_layer.objects.active = visual_objects[0]

            result = bpy.ops.export_scene.gltf(
                filepath=glb_path,
                export_format="GLB",
                use_selection=True,
                export_apply=True,
                export_animations=False,
                export_yup=True,
                export_extras=True,  # include AGS_type / AGS_name as GLTF extras
            )

            bpy.ops.object.select_all(action="DESELECT")
            for obj in prev_selected:
                obj.select_set(True)
            if prev_active:
                context.view_layer.objects.active = prev_active

            if "FINISHED" not in result:
                self.report({"WARNING"}, "AGS3D: GLTF export failed or produced no output")

        parts = []
        if self.export_gameplay:
            parts.append(f".agroom → {_os.path.basename(agroom_path)}")
        if self.export_visual:
            parts.append(f".glb → {_os.path.basename(glb_path)}")
        self.report({"INFO"}, f"AGS3D: exported {room_name!r}: {', '.join(parts)}")
        return {"FINISHED"}


# ------------------------------------------------------------------ #
# File → Import / Export menu entries                                  #
# ------------------------------------------------------------------ #

def _menu_import(self: bpy.types.Menu, context: bpy.types.Context) -> None:
    self.layout.operator(AGS3D_OT_ImportRoom.bl_idname, text="AGS3D Room (.agroom)")


def _menu_export(self: bpy.types.Menu, context: bpy.types.Context) -> None:
    self.layout.operator(AGS3D_OT_ExportRoom.bl_idname, text="AGS3D Room (.agroom + .glb)")


# ------------------------------------------------------------------ #
# Registration                                                         #
# ------------------------------------------------------------------ #

_classes = [AGS3D_OT_ImportRoom, AGS3D_OT_ExportRoom, AGS3D_OT_BakeNavMesh]


def register() -> None:
    for cls in _classes:
        bpy.utils.register_class(cls)
    bpy.types.TOPBAR_MT_file_import.append(_menu_import)
    bpy.types.TOPBAR_MT_file_export.append(_menu_export)


def unregister() -> None:
    bpy.types.TOPBAR_MT_file_export.remove(_menu_export)
    bpy.types.TOPBAR_MT_file_import.remove(_menu_import)
    for cls in reversed(_classes):
        bpy.utils.unregister_class(cls)
