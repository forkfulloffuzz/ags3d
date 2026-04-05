# SPDX-License-Identifier: GPL-3.0-or-later
#
# AGS3D Coordinate System Conversion + Bounding Box Extraction (T-BL05)
#
# Blender uses a Y-up, right-handed coordinate system.
# Godot uses a Y-up, left-handed coordinate system.
#
# The GLTF 2.0 spec uses right-handed Y-up, and Godot's GLTF importer
# applies a Z→-Z flip automatically when importing .glb files.  This
# means the standard Blender GLTF exporter with `export_yup=True` produces
# files that Godot imports with the correct orientation.
#
# For .agroom positions and sizes (written directly, not via GLTF), the
# conversion is:
#
#   godot_x =  blender_x
#   godot_y =  blender_z      (Blender Z-up → Godot Y-up)
#   godot_z = -blender_y      (right-hand → left-hand flip)
#
# This matches the axis convention used by the standard GLTF→Godot import
# pipeline, so positions written to .agroom will align with the visual mesh.
#
# Bounding box extraction:
#   Use the full world-space bounding box (all 8 corners transformed to world
#   space) to correctly handle rotated and non-uniformly scaled objects.
#   The resulting size is always axis-aligned in world space.

from __future__ import annotations

import bpy
from mathutils import Vector


# ------------------------------------------------------------------ #
# Coordinate conversion                                                #
# ------------------------------------------------------------------ #

def blender_to_godot_pos(loc: Vector) -> tuple[float, float, float]:
    """Convert a Blender world-space position to Godot world-space position.

    Blender: X-right, Y-forward, Z-up (right-handed)
    Godot:   X-right, Y-up,      Z-back (left-handed, Y-up)

    Mapping:  godot.x =  blender.x
              godot.y =  blender.z
              godot.z = -blender.y
    """
    return loc.x, loc.z, -loc.y


def blender_to_godot_size(sx: float, sy: float, sz: float) -> tuple[float, float, float]:
    """Convert a world-space axis-aligned bounding-box size from Blender to Godot.

    Size components (unsigned extents) follow the same axis permutation as
    positions — but since size is always positive, no sign flip is needed.

    godot.size.x = blender.size.x  (X-right shared)
    godot.size.y = blender.size.z  (Blender Z-up → Godot Y-up)
    godot.size.z = blender.size.y  (Blender Y-forward → Godot Z-back, unsigned)
    """
    return sx, sz, sy


# ------------------------------------------------------------------ #
# World-space bounding box                                             #
# ------------------------------------------------------------------ #

def world_bbox(obj: bpy.types.Object) -> tuple[
    tuple[float, float, float],
    tuple[float, float, float],
]:
    """Return (min_xyz, max_xyz) of the object's bounding box in Blender world space.

    Transforms all 8 local bbox corners by the object's world matrix to
    correctly handle rotations and non-uniform scales.
    """
    corners = [obj.matrix_world @ Vector(c) for c in obj.bound_box]
    xs = [c.x for c in corners]
    ys = [c.y for c in corners]
    zs = [c.z for c in corners]
    return (min(xs), min(ys), min(zs)), (max(xs), max(ys), max(zs))


def bbox_size_godot(obj: bpy.types.Object) -> tuple[float, float, float]:
    """Return the object's axis-aligned bounding box size in Godot coordinates.

    Uses the full world-space bbox (via world_bbox) and converts the
    Blender (sx, sy, sz) extents to Godot coordinates via blender_to_godot_size.
    """
    (xmin, ymin, zmin), (xmax, ymax, zmax) = world_bbox(obj)
    sx_bl = xmax - xmin
    sy_bl = ymax - ymin
    sz_bl = zmax - zmin
    return blender_to_godot_size(sx_bl, sy_bl, sz_bl)


def bbox_center_godot(obj: bpy.types.Object) -> tuple[float, float, float]:
    """Return the world-space bounding box center in Godot coordinates."""
    (xmin, ymin, zmin), (xmax, ymax, zmax) = world_bbox(obj)
    cx = (xmin + xmax) / 2.0
    cy = (ymin + ymax) / 2.0
    cz = (zmin + zmax) / 2.0
    return blender_to_godot_pos(Vector((cx, cy, cz)))
