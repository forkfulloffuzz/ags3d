#include "ags_walkable_surface.h"

#include "core/config/engine.h"
#include "scene/resources/navigation_mesh.h"

void AGSWalkableSurface::_bind_methods() {}

void AGSWalkableSurface::_notification(int p_what) {
	if (p_what == NOTIFICATION_READY) {
		if (Engine::get_singleton()->is_editor_hint()) {
			_apply_editor_overlay();
		} else {
			_setup_navmesh();
		}
	}
}

void AGSWalkableSurface::_apply_editor_overlay() {
	// Mesh children remain visible — the standard Godot editor has no gizmo
	// plugin to replace them. M12 (Custom Editor) will add gizmo overlays.
}

void AGSWalkableSurface::_setup_navmesh() {
	// Visual mesh children remain visible — they show the walkable floor.
	// Authors can hide them manually in the scene if they add real geometry.

	// Add self to the group the NavigationMesh will source geometry from.
	add_to_group("ags_walkable");

	// Create a NavigationRegion3D to own the baked navmesh.
	nav_region = memnew(NavigationRegion3D);
	add_child(nav_region);

	Ref<NavigationMesh> nav_mesh;
	nav_mesh.instantiate();
	nav_mesh->set_parsed_geometry_type(NavigationMesh::PARSED_GEOMETRY_STATIC_COLLIDERS);
	nav_mesh->set_source_geometry_mode(NavigationMesh::SOURCE_GEOMETRY_GROUPS_WITH_CHILDREN);
	nav_mesh->set_source_group_name("ags_walkable");

	nav_region->set_navigation_mesh(nav_mesh);
	// Bake synchronously on scene load. Baking requires the node to be inside
	// the SceneTree (NavigationServer constraint), so skip it in unit-test contexts.
	if (is_inside_tree()) {
		nav_region->bake_navigation_mesh(false);
	}
}
