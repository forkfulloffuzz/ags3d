#include "ags_blocker_volume.h"

#include "core/config/engine.h"
#include "scene/3d/mesh_instance_3d.h"
#include "scene/3d/physics/collision_shape_3d.h"
#include "scene/resources/3d/box_shape_3d.h"
#include "scene/resources/3d/primitive_meshes.h"
#include "scene/resources/material.h"

void AGSBlockerVolume::_bind_methods() {}

void AGSBlockerVolume::_notification(int p_what) {
	if (p_what == NOTIFICATION_READY) {
		if (Engine::get_singleton()->is_editor_hint()) {
			_apply_editor_overlay();
		} else {
			// Hide any scene-authored mesh children at runtime; collision shape
			// stays active so it carves itself out of any NavigationMesh bake.
			for (int i = 0; i < get_child_count(); i++) {
				MeshInstance3D *mi = Object::cast_to<MeshInstance3D>(get_child(i));
				if (mi) {
					mi->set_visible(false);
				}
			}
		}
	}
}

void AGSBlockerVolume::_apply_editor_overlay() {
	for (int i = 0; i < get_child_count(); i++) {
		CollisionShape3D *cs = Object::cast_to<CollisionShape3D>(get_child(i));
		if (!cs) {
			continue;
		}
		Ref<Shape3D> shape = cs->get_shape();
		if (shape.is_null()) {
			continue;
		}
		BoxShape3D *box = Object::cast_to<BoxShape3D>(*shape);
		if (!box) {
			continue;
		}

		// --- Solid fill ---
		Ref<BoxMesh> fill_mesh;
		fill_mesh.instantiate();
		fill_mesh->set_size(box->get_size());

		Ref<StandardMaterial3D> fill_mat;
		fill_mat.instantiate();
		fill_mat->set_albedo(Color(0.9f, 0.15f, 0.15f, 0.45f));
		fill_mat->set_transparency(BaseMaterial3D::TRANSPARENCY_ALPHA);
		fill_mat->set_shading_mode(BaseMaterial3D::SHADING_MODE_UNSHADED);
		fill_mat->set_cull_mode(BaseMaterial3D::CULL_DISABLED);

		MeshInstance3D *fill_mi = memnew(MeshInstance3D);
		fill_mi->set_mesh(fill_mesh);
		fill_mi->set_surface_override_material(0, fill_mat);
		fill_mi->set_transform(cs->get_transform());
		add_child(fill_mi, false, Node::INTERNAL_MODE_BACK);

		// --- Wireframe outline (always visible, shows exact blocker bounds) ---
		// get_debug_mesh() returns a PRIMITIVE_LINES mesh — same as CollisionShape3D's gizmo.
		Ref<ArrayMesh> wire_mesh = shape->get_debug_mesh();

		Ref<StandardMaterial3D> wire_mat;
		wire_mat.instantiate();
		wire_mat->set_albedo(Color(1.0f, 0.25f, 0.25f, 1.0f));
		wire_mat->set_shading_mode(BaseMaterial3D::SHADING_MODE_UNSHADED);

		MeshInstance3D *wire_mi = memnew(MeshInstance3D);
		wire_mi->set_mesh(wire_mesh);
		wire_mi->set_surface_override_material(0, wire_mat);
		wire_mi->set_transform(cs->get_transform());
		add_child(wire_mi, false, Node::INTERNAL_MODE_BACK);
	}
}
