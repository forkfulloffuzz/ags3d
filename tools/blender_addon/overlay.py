# SPDX-License-Identifier: GPL-3.0-or-later
#
# AGS3D Viewport Overlay
#
# Draws colored wireframe outlines for AGS-tagged objects in the Blender 3D
# viewport. Each AGS type has a distinct color matching the AG Studio gizmo
# palette. A toggle in View → Overlays → AGS3D enables / disables the overlay.
#
# Implementation notes:
#   - Uses gpu + blf for immediate-mode drawing via SpaceView3D.draw_handler_add.
#   - Draws bounding-box wireframes for volume types; dot + label for Points.
#   - The handler is added once per viewport and cleaned up on unregister.

from __future__ import annotations

import bpy
import gpu
import blf
from gpu_extras.batch import batch_for_shader
from mathutils import Vector

from .panels import _get_ags_type, AGS_TYPE_ITEMS

# ------------------------------------------------------------------ #
# Type colors (RGBA, linear)                                           #
# ------------------------------------------------------------------ #

_TYPE_COLORS: dict[str, tuple[float, float, float, float]] = {
    "WALKABLE": (0.2, 0.9, 0.2, 0.8),   # green
    "BLOCKER":  (0.9, 0.2, 0.2, 0.8),   # red
    "POINT":    (1.0, 1.0, 1.0, 0.9),   # white
    "CAMERA":   (1.0, 0.9, 0.1, 0.8),   # yellow
    "HOTSPOT":  (0.2, 0.4, 1.0, 0.8),   # blue
    "TRIGGER":  (0.7, 0.2, 0.9, 0.8),   # purple
    "SPAWN":    (0.1, 0.9, 0.9, 0.8),   # cyan
    "NAVMESH":  (0.1, 0.7, 0.6, 0.5),   # teal (semi-transparent)
}

# ------------------------------------------------------------------ #
# Overlay toggle property                                              #
# ------------------------------------------------------------------ #

class AGS3DOverlayProperties(bpy.types.PropertyGroup):
    show_overlay: bpy.props.BoolProperty(
        name="AGS3D",
        description="Show AGS3D gameplay object overlays in the viewport",
        default=True,
    )  # type: ignore[assignment]


# ------------------------------------------------------------------ #
# Overlay panel (View → Overlays popover)                             #
# ------------------------------------------------------------------ #

class VIEW3D_PT_ags3d_overlay(bpy.types.Panel):
    bl_label = ""
    bl_idname = "VIEW3D_PT_ags3d_overlay"
    bl_space_type = "VIEW_3D"
    bl_region_type = "HEADER"
    bl_parent_id = "VIEW3D_PT_overlay"

    @classmethod
    def poll(cls, context: bpy.types.Context) -> bool:
        return True

    def draw(self, context: bpy.types.Context) -> None:
        layout = self.layout
        overlay = context.scene.ags3d_overlay
        layout.prop(overlay, "show_overlay")


# ------------------------------------------------------------------ #
# Draw handler                                                         #
# ------------------------------------------------------------------ #

_SHADER: gpu.types.GPUShader | None = None


def _get_shader() -> gpu.types.GPUShader:
    global _SHADER
    if _SHADER is None:
        _SHADER = gpu.shader.from_builtin("UNIFORM_COLOR")
    return _SHADER


def _bbox_lines(obj: bpy.types.Object) -> list[tuple[float, float, float]]:
    """Return world-space line pairs for the 12 edges of the object's bounding box."""
    bb = obj.bound_box  # 8 corners in local space
    corners = [obj.matrix_world @ Vector(c) for c in bb]
    # Edge indices for a bounding box (pairs of corner indices).
    edges = [
        (0, 1), (1, 2), (2, 3), (3, 0),  # bottom face
        (4, 5), (5, 6), (6, 7), (7, 4),  # top face
        (0, 4), (1, 5), (2, 6), (3, 7),  # verticals
    ]
    verts: list[tuple[float, float, float]] = []
    for a, b in edges:
        verts.append(corners[a].to_tuple())
        verts.append(corners[b].to_tuple())
    return verts


def _draw_callback() -> None:
    context = bpy.context
    if not hasattr(context.scene, "ags3d_overlay"):
        return
    if not context.scene.ags3d_overlay.show_overlay:
        return

    shader = _get_shader()
    gpu.state.blend_set("ALPHA")
    gpu.state.line_width_set(2.0)

    for obj in context.scene.objects:
        ags_type = _get_ags_type(obj)
        if ags_type == "NONE":
            continue
        color = _TYPE_COLORS.get(ags_type)
        if color is None:
            continue

        # --- wireframe bounding box ---
        verts = _bbox_lines(obj)
        batch = batch_for_shader(shader, "LINES", {"pos": verts})
        shader.bind()
        shader.uniform_float("color", color)
        batch.draw(shader)

    # --- name labels ---
    font_id = 0
    blf.size(font_id, 12)
    for obj in context.scene.objects:
        ags_type = _get_ags_type(obj)
        if ags_type == "NONE":
            continue
        color = _TYPE_COLORS.get(ags_type)
        if color is None:
            continue

        # Project object origin to 2D screen coords.
        region = context.region
        rv3d = context.region_data
        if region is None or rv3d is None:
            continue
        co_3d = obj.matrix_world.translation
        co_2d = bpy.ops.view3d.snap_cursor_to_active.__doc__  # placeholder
        # Use bpy_extras for projection.
        from bpy_extras.view3d_utils import location_3d_to_region_2d
        co_2d = location_3d_to_region_2d(region, rv3d, co_3d)
        if co_2d is None:
            continue

        label = obj.get("AGS_name", "") or obj.name
        ags_label = _type_display(ags_type)
        blf.color(font_id, *color)
        blf.position(font_id, co_2d.x + 8, co_2d.y + 4, 0.0)
        blf.draw(font_id, f"{ags_label}: {label}")

    gpu.state.blend_set("NONE")
    gpu.state.line_width_set(1.0)


def _type_display(ags_type: str) -> str:
    for id_, label, _ in AGS_TYPE_ITEMS:
        if id_ == ags_type:
            return label
    return ags_type


# ------------------------------------------------------------------ #
# Handler registration                                                 #
# ------------------------------------------------------------------ #

_draw_handle = None


def _add_handler() -> None:
    global _draw_handle
    if _draw_handle is None:
        _draw_handle = bpy.types.SpaceView3D.draw_handler_add(
            _draw_callback, (), "WINDOW", "POST_VIEW"
        )


def _remove_handler() -> None:
    global _draw_handle
    if _draw_handle is not None:
        bpy.types.SpaceView3D.draw_handler_remove(_draw_handle, "WINDOW")
        _draw_handle = None


# ------------------------------------------------------------------ #
# Registration                                                         #
# ------------------------------------------------------------------ #

_classes = [
    AGS3DOverlayProperties,
    VIEW3D_PT_ags3d_overlay,
]


def register() -> None:
    for cls in _classes:
        bpy.utils.register_class(cls)
    bpy.types.Scene.ags3d_overlay = bpy.props.PointerProperty(
        type=AGS3DOverlayProperties
    )
    _add_handler()


def unregister() -> None:
    _remove_handler()
    if hasattr(bpy.types.Scene, "ags3d_overlay"):
        del bpy.types.Scene.ags3d_overlay
    for cls in reversed(_classes):
        bpy.utils.unregister_class(cls)
