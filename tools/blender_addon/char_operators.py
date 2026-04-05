# SPDX-License-Identifier: GPL-3.0-or-later
#
# AGS3D Character Export Operator (T-BL10)
#
# File → Export → AGS3D Character (.glb)
#
# Exports the selected armature + its mesh children + all NLA actions as a
# single GLTF 2.0 binary (.glb) file. The output is suitable for use as the
# `mesh` path in a .agchar file; animation clip names in the .glb match the
# NLA track names and should be referenced in the `animations { }` block.
#
# Selection rules:
#   - If an Armature object is active/selected, export it and all its mesh
#     children (objects whose Armature modifier targets this armature).
#   - If no armature is selected, export all selected mesh objects (for
#     non-rigged billboard sprite characters).
#
# NLA actions:
#   All NLA tracks on the active armature that have at least one strip are
#   included as separate animation clips in the GLTF file. Track names
#   (e.g. "Idle", "Walk", "Talk") become the animation names in the .glb.
#
# Output path:
#   The file selector pre-fills with characters/<active-object-name>/<name>.glb
#   relative to the blend file location. Authors can override it.

from __future__ import annotations

import os
import bpy


# ------------------------------------------------------------------ #
# Helpers                                                              #
# ------------------------------------------------------------------ #

def _find_armature(context: bpy.types.Context) -> bpy.types.Object | None:
    """Return the active armature, or None."""
    obj = context.active_object
    if obj and obj.type == "ARMATURE":
        return obj
    # Fallback: any selected armature.
    for o in context.selected_objects:
        if o.type == "ARMATURE":
            return o
    return None


def _mesh_children(armature: bpy.types.Object) -> list[bpy.types.Object]:
    """Return mesh objects parented to or weighted to `armature`."""
    result: list[bpy.types.Object] = []
    for obj in bpy.data.objects:
        if obj.type != "MESH":
            continue
        # Direct parent.
        if obj.parent == armature:
            result.append(obj)
            continue
        # Armature modifier targeting this armature.
        for mod in obj.modifiers:
            if mod.type == "ARMATURE" and mod.object == armature:
                result.append(obj)
                break
    return result


def _nla_track_names(armature: bpy.types.Object) -> list[str]:
    """Return names of NLA tracks that have at least one strip."""
    if not armature.animation_data:
        return []
    tracks = []
    for track in armature.animation_data.nla_tracks:
        if len(track.strips) > 0:
            tracks.append(track.name)
    return tracks


def _default_output_path(context: bpy.types.Context, char_name: str) -> str:
    """Suggest characters/<name>/<name>.glb relative to blend file."""
    blend = bpy.data.filepath
    if blend:
        root = os.path.dirname(blend)
        return os.path.join(root, "characters", char_name, char_name + ".glb")
    return os.path.join(os.path.expanduser("~"), char_name + ".glb")


# ------------------------------------------------------------------ #
# Export operator                                                      #
# ------------------------------------------------------------------ #

class AGS3D_OT_ExportCharacter(bpy.types.Operator):
    """Export selected armature + meshes + NLA actions as AGS3D character .glb"""

    bl_idname = "ags3d.export_character"
    bl_label = "AGS3D Character (.glb)"
    bl_options = {"REGISTER"}

    filepath: bpy.props.StringProperty(
        name="File Path",
        subtype="FILE_PATH",
    )  # type: ignore[assignment]

    filter_glob: bpy.props.StringProperty(
        default="*.glb",
        options={"HIDDEN"},
    )  # type: ignore[assignment]

    export_animations: bpy.props.BoolProperty(
        name="Export Animations",
        description="Include all NLA track actions as GLTF animation clips",
        default=True,
    )  # type: ignore[assignment]

    apply_modifiers: bpy.props.BoolProperty(
        name="Apply Modifiers",
        description="Apply mesh modifiers before export",
        default=True,
    )  # type: ignore[assignment]

    def invoke(self, context: bpy.types.Context, event: bpy.types.Event) -> set:
        armature = _find_armature(context)
        char_name = armature.name if armature else (
            context.active_object.name if context.active_object else "character"
        )
        self.filepath = _default_output_path(context, char_name)
        context.window_manager.fileselect_add(self)
        return {"RUNNING_MODAL"}

    def execute(self, context: bpy.types.Context) -> set:
        armature = _find_armature(context)

        # Build the export object set.
        export_objects: list[bpy.types.Object] = []
        if armature:
            export_objects.append(armature)
            export_objects.extend(_mesh_children(armature))
            track_names = _nla_track_names(armature)
        else:
            # No armature — export all selected meshes (billboard sprites).
            export_objects = [o for o in context.selected_objects if o.type == "MESH"]
            track_names = []

        if not export_objects:
            self.report({"WARNING"}, "AGS3D: nothing to export — select an armature or mesh")
            return {"CANCELLED"}

        # Ensure output directory exists.
        out_dir = os.path.dirname(self.filepath)
        if out_dir:
            os.makedirs(out_dir, exist_ok=True)

        # Temporarily select only the export objects.
        prev_selected = list(context.selected_objects)
        prev_active = context.active_object
        bpy.ops.object.select_all(action="DESELECT")
        for obj in export_objects:
            obj.select_set(True)
        if export_objects:
            context.view_layer.objects.active = export_objects[0]

        # Call Blender's built-in GLTF exporter.
        result = bpy.ops.export_scene.gltf(
            filepath=self.filepath,
            export_format="GLB",
            use_selection=True,
            export_apply=self.apply_modifiers,
            export_animations=self.export_animations and bool(armature),
            export_nla_strips=self.export_animations and bool(armature),
            export_nla_strips_merged_animation_name="",
            export_yup=True,
        )

        # Restore selection.
        bpy.ops.object.select_all(action="DESELECT")
        for obj in prev_selected:
            obj.select_set(True)
        if prev_active:
            context.view_layer.objects.active = prev_active

        if "FINISHED" not in result:
            self.report({"ERROR"}, "AGS3D: GLTF export failed")
            return {"CANCELLED"}

        # Report which animation clips were included.
        if track_names:
            self.report(
                {"INFO"},
                f"AGS3D: exported {armature.name!r} with animations: {', '.join(track_names)}"
            )
        else:
            self.report({"INFO"}, f"AGS3D: exported {len(export_objects)} object(s) to {self.filepath}")

        return {"FINISHED"}


# ------------------------------------------------------------------ #
# File → Export menu entry                                             #
# ------------------------------------------------------------------ #

def _menu_export(self: bpy.types.Menu, context: bpy.types.Context) -> None:
    self.layout.operator(AGS3D_OT_ExportCharacter.bl_idname, text="AGS3D Character (.glb)")


# ------------------------------------------------------------------ #
# Registration                                                         #
# ------------------------------------------------------------------ #

_classes = [AGS3D_OT_ExportCharacter]


def register() -> None:
    for cls in _classes:
        bpy.utils.register_class(cls)
    bpy.types.TOPBAR_MT_file_export.append(_menu_export)


def unregister() -> None:
    bpy.types.TOPBAR_MT_file_export.remove(_menu_export)
    for cls in reversed(_classes):
        bpy.utils.unregister_class(cls)
