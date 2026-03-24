#include "ags_character.h"

#include "ags_runtime.h"
#include "core/object/class_db.h"
#include "scene/3d/mesh_instance_3d.h"
#include "scene/resources/3d/primitive_meshes.h"

void AGSCharacter::_bind_methods() {
	ClassDB::bind_method(D_METHOD("set_character_name", "name"), &AGSCharacter::set_character_name);
	ClassDB::bind_method(D_METHOD("get_character_name"), &AGSCharacter::get_character_name);
	ADD_PROPERTY(PropertyInfo(Variant::STRING, "character_name"), "set_character_name", "get_character_name");
}

void AGSCharacter::_notification(int p_what) {
	switch (p_what) {
		case NOTIFICATION_READY: {
			// Add a placeholder capsule so the character is visible in editor and at runtime.
			if (get_child_count() == 0) {
				MeshInstance3D *mesh = memnew(MeshInstance3D);
				Ref<CapsuleMesh> capsule;
				capsule.instantiate();
				mesh->set_mesh(capsule);
				add_child(mesh);
			}

			// Register with AGSRuntime so scripts can resolve this character by name.
			if (AGSRuntime::get_singleton()) {
				AGSRuntime::get_singleton()->register_character(this);
			}
		} break;

		case NOTIFICATION_EXIT_TREE: {
			if (AGSRuntime::get_singleton()) {
				AGSRuntime::get_singleton()->unregister_character(this);
			}
		} break;
	}
}

void AGSCharacter::set_character_name(const String &p_name) {
	character_name = p_name;
}

String AGSCharacter::get_character_name() const {
	return character_name;
}
