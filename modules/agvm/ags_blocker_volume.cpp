#include "ags_blocker_volume.h"

#include "core/config/engine.h"
#include "scene/resources/material.h"

void AGSBlockerVolume::_bind_methods() {}

void AGSBlockerVolume::_notification(int p_what) {
	if (p_what == NOTIFICATION_READY) {
		if (Engine::get_singleton()->is_editor_hint()) {
			_apply_editor_overlay();
		} else {
			// Hide visual geometry at runtime; collision shape remains active
			// so it carves itself out of any NavigationMesh bake.
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
	Ref<StandardMaterial3D> mat;
	mat.instantiate();
	mat->set_albedo(Color(0.9f, 0.1f, 0.1f, 0.35f));
	mat->set_transparency(BaseMaterial3D::TRANSPARENCY_ALPHA);
	mat->set_shading_mode(BaseMaterial3D::SHADING_MODE_UNSHADED);

	for (int i = 0; i < get_child_count(); i++) {
		MeshInstance3D *mi = Object::cast_to<MeshInstance3D>(get_child(i));
		if (mi) {
			mi->set_material_overlay(mat);
		}
	}
}
