# SPDX-License-Identifier: GPL-3.0-or-later
#
# AGS3D Object Type Panel
#
# Registers:
#   - OBJECT_PT_ags3d  (Object Properties → AGS3D section)
#   - VIEW3D_PT_ags3d  (3D Viewport → N-panel → AGS3D tab)
#
# Both panels share the same draw function (_draw_ags3d_panel).
# The type dropdown and name field write to custom properties
# AGS_type and AGS_name on the active object.

import bpy

# ------------------------------------------------------------------ #
# Constants                                                            #
# ------------------------------------------------------------------ #

AGS_TYPE_ITEMS = [
    ("NONE",           "None",             "Object is exported as visual geometry only"),
    ("WALKABLE",       "WalkableSurface",  "Walkable floor area for character navigation"),
    ("BLOCKER",        "BlockerVolume",    "Invisible collision volume blocking movement"),
    ("POINT",          "Point",            "Named world-space point (door, item anchor, etc.)"),
    ("CAMERA",         "Camera",           "Named camera position with look_at target"),
    ("HOTSPOT",        "Hotspot",          "Player-interactable area"),
    ("TRIGGER",        "TriggerRegion",    "Region that fires script events on enter/exit"),
    ("SPAWN",          "SpawnPoint",       "Character spawn location"),
    ("NAVMESH",        "NavMesh",          "Baked navigation mesh for character pathfinding"),
    ("VISUAL",         "VisualMesh",       "Explicitly marked visual geometry (excluded from gameplay export)"),
    ("LIGHT",          "Light",            "Blender light exported via GLTF"),
]

# Types that require a Name field.
_NAMED_TYPES = {"WALKABLE", "BLOCKER", "POINT", "CAMERA", "HOTSPOT", "TRIGGER", "SPAWN"}

# ------------------------------------------------------------------ #
# Helpers                                                              #
# ------------------------------------------------------------------ #

def _get_ags_type(obj: bpy.types.Object) -> str:
    return str(obj.get("AGS_type", "NONE"))


def _set_ags_type(obj: bpy.types.Object, value: str) -> None:
    obj["AGS_type"] = value


def _get_ags_name(obj: bpy.types.Object) -> str:
    return str(obj.get("AGS_name", ""))


def _set_ags_name(obj: bpy.types.Object, value: str) -> None:
    obj["AGS_name"] = value


def _get_ags_character(obj: bpy.types.Object) -> str:
    return str(obj.get("AGS_character", ""))


def _set_ags_character(obj: bpy.types.Object, value: str) -> None:
    obj["AGS_character"] = value


# ------------------------------------------------------------------ #
# Operators                                                            #
# ------------------------------------------------------------------ #

class AGS3D_OT_SetType(bpy.types.Operator):
    """Set the AGS3D type of the active object"""

    bl_idname = "ags3d.set_type"
    bl_label = "Set AGS3D Type"
    bl_options = {"REGISTER", "UNDO"}

    ags_type: bpy.props.EnumProperty(
        name="AGS3D Type",
        items=AGS_TYPE_ITEMS,
        default="NONE",
    )  # type: ignore[assignment]

    def execute(self, context: bpy.types.Context) -> set:
        obj = context.active_object
        if obj is None:
            self.report({"WARNING"}, "No active object")
            return {"CANCELLED"}
        _set_ags_type(obj, self.ags_type)
        return {"FINISHED"}


# ------------------------------------------------------------------ #
# Shared panel draw function                                           #
# ------------------------------------------------------------------ #

def _draw_ags3d_panel(layout: bpy.types.UILayout, obj: bpy.types.Object) -> None:
    """Draw the AGS3D type selector and type-specific fields."""

    current_type = _get_ags_type(obj)

    # --- Type dropdown ------------------------------------------------
    row = layout.row()
    row.label(text="Type:")
    # Use a menu button that calls AGS3D_OT_SetType for each option.
    # A simple approach: use a UILayout.prop on a custom property via
    # a dynamic EnumProperty on the object.  We use operator buttons
    # grouped in a menu because custom-property enum dropdowns are
    # cumbersome in Blender 4.x.  The clean approach is a proper
    # PropertyGroup; we use that here.
    op_row = layout.row(align=True)
    op_row.operator_menu_enum("ags3d.set_type", "ags_type", text=_type_label(current_type))

    if current_type == "NONE":
        return

    layout.separator()

    # --- Name field (most types) -------------------------------------
    if current_type in _NAMED_TYPES:
        row = layout.row()
        row.label(text="Name:")
        name_row = layout.row()
        # We cannot bind a StringProperty to a custom dict key directly,
        # so we use a prop_search fallback or a plain string operator.
        # For now render the current value + a button to edit it.
        name_row.prop(obj, '["AGS_name"]', text="")

    # --- SpawnPoint extra field: character name ----------------------
    if current_type == "SPAWN":
        layout.separator()
        row = layout.row()
        row.label(text="Character:")
        layout.prop(obj, '["AGS_character"]', text="")

    # --- Camera extra field: look_at target --------------------------
    if current_type == "CAMERA":
        layout.separator()
        row = layout.row()
        row.label(text="Look-at:")
        layout.prop_search(
            obj,
            '["AGS_look_at"]',
            bpy.context.scene,
            "objects",
            text="",
        )


def _type_label(ags_type: str) -> str:
    for id_, label, _ in AGS_TYPE_ITEMS:
        if id_ == ags_type:
            return label
    return "None"


# ------------------------------------------------------------------ #
# Panel: Object Properties                                             #
# ------------------------------------------------------------------ #

class OBJECT_PT_ags3d(bpy.types.Panel):
    bl_label = "AGS3D"
    bl_idname = "OBJECT_PT_ags3d"
    bl_space_type = "PROPERTIES"
    bl_region_type = "WINDOW"
    bl_context = "object"

    @classmethod
    def poll(cls, context: bpy.types.Context) -> bool:
        return context.active_object is not None

    def draw(self, context: bpy.types.Context) -> None:
        _draw_ags3d_panel(self.layout, context.active_object)


# ------------------------------------------------------------------ #
# Panel: N-panel (3D Viewport sidebar)                                 #
# ------------------------------------------------------------------ #

class VIEW3D_PT_ags3d(bpy.types.Panel):
    bl_label = "AGS3D"
    bl_idname = "VIEW3D_PT_ags3d"
    bl_space_type = "VIEW_3D"
    bl_region_type = "UI"
    bl_category = "AGS3D"

    @classmethod
    def poll(cls, context: bpy.types.Context) -> bool:
        return context.active_object is not None

    def draw(self, context: bpy.types.Context) -> None:
        _draw_ags3d_panel(self.layout, context.active_object)


# ------------------------------------------------------------------ #
# Registration                                                         #
# ------------------------------------------------------------------ #

_classes = [
    AGS3D_OT_SetType,
    OBJECT_PT_ags3d,
    VIEW3D_PT_ags3d,
]


def register() -> None:
    for cls in _classes:
        bpy.utils.register_class(cls)


def unregister() -> None:
    for cls in reversed(_classes):
        bpy.utils.unregister_class(cls)
