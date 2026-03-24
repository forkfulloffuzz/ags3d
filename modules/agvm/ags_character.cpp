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

	ClassDB::bind_method(D_METHOD("navigate_to", "target"), &AGSCharacter::navigate_to);

	ClassDB::bind_method(D_METHOD("_on_velocity_computed", "safe_velocity"), &AGSCharacter::_on_velocity_computed);
	ClassDB::bind_method(D_METHOD("_on_navigation_finished"), &AGSCharacter::_on_navigation_finished);
}

void AGSCharacter::_notification(int p_what) {
	switch (p_what) {
		case NOTIFICATION_READY: {
			// Placeholder capsule mesh — visible in editor and at runtime.
			if (get_child_count() == 0) {
				MeshInstance3D *mesh = memnew(MeshInstance3D);
				Ref<CapsuleMesh> capsule;
				capsule.instantiate();
				mesh->set_mesh(capsule);
				add_child(mesh);
			}

			// Create and wire up the NavigationAgent3D.
			nav_agent = memnew(NavigationAgent3D);
			add_child(nav_agent);
			nav_agent->connect("velocity_computed", Callable(this, "_on_velocity_computed"));
			nav_agent->connect("navigation_finished", Callable(this, "_on_navigation_finished"));

			set_physics_process(false);

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

void AGSCharacter::_physics_process(double p_delta) {
	if (!nav_agent || nav_agent->is_navigation_finished()) {
		navigating = false;
		set_physics_process(false);
		return;
	}

	Vector3 next_pos = nav_agent->get_next_path_position();
	Vector3 direction = (next_pos - get_global_position()).normalized();
	Vector3 desired_velocity = direction * move_speed;

	// Submit to avoidance; movement is applied in _on_velocity_computed.
	nav_agent->set_velocity(desired_velocity);
}

void AGSCharacter::_on_velocity_computed(const Vector3 &p_safe_velocity) {
	set_velocity(p_safe_velocity);
	move_and_slide();
}

void AGSCharacter::_on_navigation_finished() {
	navigating = false;
	set_physics_process(false);
	set_velocity(Vector3());
}

void AGSCharacter::navigate_to(const Vector3 &p_target) {
	ERR_FAIL_NULL_MSG(nav_agent, "AGSCharacter: NavigationAgent3D not initialised — call navigate_to() after _ready().");
	nav_agent->set_target_position(p_target);
	navigating = true;
	set_physics_process(true);
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
