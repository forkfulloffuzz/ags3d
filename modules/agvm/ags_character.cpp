#include "ags_character.h"

#include "ags_runtime.h"
#include "core/object/class_db.h"
#include "scene/3d/mesh_instance_3d.h"
#include "scene/resources/3d/primitive_meshes.h"

void AGSCharacter::_bind_methods() {
	ClassDB::bind_method(D_METHOD("set_character_name", "name"), &AGSCharacter::set_character_name);
	ClassDB::bind_method(D_METHOD("get_character_name"), &AGSCharacter::get_character_name);
	ADD_PROPERTY(PropertyInfo(Variant::STRING, "character_name"), "set_character_name", "get_character_name");

	ClassDB::bind_method(D_METHOD("set_move_speed", "speed"), &AGSCharacter::set_move_speed);
	ClassDB::bind_method(D_METHOD("get_move_speed"), &AGSCharacter::get_move_speed);
	ADD_PROPERTY(PropertyInfo(Variant::FLOAT, "move_speed"), "set_move_speed", "get_move_speed");

	ClassDB::bind_method(D_METHOD("set_say_text", "text"), &AGSCharacter::set_say_text);
	ClassDB::bind_method(D_METHOD("get_say_text"), &AGSCharacter::get_say_text);
	ADD_PROPERTY(PropertyInfo(Variant::STRING, "say_text"), "set_say_text", "get_say_text");

	ADD_SIGNAL(MethodInfo("walk_completed"));
	ADD_SIGNAL(MethodInfo("face_completed"));
	ADD_SIGNAL(MethodInfo("say_completed"));
}

void AGSCharacter::_notification(int p_what) {
	switch (p_what) {
		case NOTIFICATION_ENTER_TREE: {
			// Placeholder capsule mesh — visible in editor and at runtime.
			// Only created on first entry so editor-placed children are respected.
			if (get_child_count() == 0) {
				MeshInstance3D *mesh = memnew(MeshInstance3D);
				Ref<CapsuleMesh> capsule;
				capsule.instantiate();
				mesh->set_mesh(capsule);
				add_child(mesh);
			}
		} break;

		case NOTIFICATION_READY: {
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

		case NOTIFICATION_PREDELETE: {
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

void AGSCharacter::set_move_speed(float p_speed) {
	move_speed = p_speed;
}

float AGSCharacter::get_move_speed() const {
	return move_speed;
}

void AGSCharacter::set_say_text(const String &p_text) {
	say_text = p_text;
}

String AGSCharacter::get_say_text() const {
	return say_text;
}
