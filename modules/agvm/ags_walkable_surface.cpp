#include "ags_walkable_surface.h"

#include "core/config/engine.h"
#include "scene/resources/material.h"
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
	// Apply a semi-transparent green material to all MeshInstance3D children
	// so the walkable area is visible in the editor.
	Ref<StandardMaterial3D> mat;
	mat.instantiate();
	mat->set_albedo(Color(0.0f, 0.8f, 0.2f, 0.35f));
	mat->set_transparency(BaseMaterial3D::TRANSPARENCY_ALPHA);
	mat->set_shading_mode(BaseMaterial3D::SHADING_MODE_UNSHADED);
	mat->set_flag(BaseMaterial3D::FLAG_DISABLE_DEPTH_TEST, false);

	for (int i = 0; i < get_child_count(); i++) {
		MeshInstance3D *mi = Object::cast_to<MeshInstance3D>(get_child(i));
		if (mi) {
			mi->set_material_overlay(mat);
		}
	}
}

void AGSWalkableSurface::_setup_navmesh() {
	// Hide visual geometry at runtime — navmesh handles pathfinding.
	for (int i = 0; i < get_child_count(); i++) {
		MeshInstance3D *mi = Object::cast_to<MeshInstance3D>(get_child(i));
		if (mi) {
			mi->set_visible(false);
		}
	}

	// Add self to the group the NavigationMesh will source geometry from.
	add_to_group("ags_walkable");

	// Create a NavigationRegion3D to own the baked navmesh.
	nav_region = memnew(NavigationRegion3D);
	add_child(nav_region);

	Ref<NavigationMesh> nav_mesh;
	nav_mesh.instantiate();
	nav_mesh->set_parsed_geometry_type(NavigationMesh::PARSED_GEOMETRY_MESH_INSTANCES);
	nav_mesh->set_source_geometry_mode(NavigationMesh::SOURCE_GEOMETRY_GROUPS_WITH_CHILDREN);
	nav_mesh->set_source_group_name("ags_walkable");

	nav_region->set_navigation_mesh(nav_mesh);
	// Bake synchronously on scene load. Baking requires the node to be inside
	// the SceneTree (NavigationServer constraint), so skip it in unit-test contexts.
	if (is_inside_tree()) {
		nav_region->bake_navigation_mesh(false);
	}
}
