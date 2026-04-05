#include "ags_character_3d.h"

#include "scene/3d/mesh_instance_3d.h"
#include "scene/resources/3d/primitive_meshes.h"

void AGSCharacter3D::_notification(int p_what) {
	// Propagate all notifications to the base class first.
	AGSCharacterBase::_notification(p_what);

	switch (p_what) {
		case NOTIFICATION_ENTER_TREE: {
			// Add a placeholder capsule mesh so the node is visible before a real
			// mesh asset is assigned. Only create on first entry so existing children
			// placed in the editor are not duplicated.
			if (get_child_count() == 0) {
				MeshInstance3D *mesh = memnew(MeshInstance3D);
				Ref<CapsuleMesh> capsule;
				capsule.instantiate();
				mesh->set_mesh(capsule);
				add_child(mesh);
			}
		} break;
	}
}
