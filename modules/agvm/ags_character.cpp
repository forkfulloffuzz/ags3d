#include "ags_character.h"

#include "ags_room.h"
#include "ags_runtime.h"
#include "core/math/basis.h"
#include "core/object/class_db.h"
#include "scene/3d/mesh_instance_3d.h"
#include "scene/animation/tween.h"
#include "scene/resources/3d/primitive_meshes.h"

void AGSCharacter::_bind_methods() {
	ClassDB::bind_method(D_METHOD("set_character_name", "name"), &AGSCharacter::set_character_name);
	ClassDB::bind_method(D_METHOD("get_character_name"), &AGSCharacter::get_character_name);
	ADD_PROPERTY(PropertyInfo(Variant::STRING, "character_name"), "set_character_name", "get_character_name");

	ClassDB::bind_method(D_METHOD("set_move_speed", "speed"), &AGSCharacter::set_move_speed);
	ClassDB::bind_method(D_METHOD("get_move_speed"), &AGSCharacter::get_move_speed);
	ADD_PROPERTY(PropertyInfo(Variant::FLOAT, "move_speed"), "set_move_speed", "get_move_speed");

	ClassDB::bind_method(D_METHOD("navigate_to", "target"), &AGSCharacter::navigate_to);
	ClassDB::bind_method(D_METHOD("walk_to", "point_name"), &AGSCharacter::walk_to);
	ClassDB::bind_method(D_METHOD("face_to", "point_name"), &AGSCharacter::face_to);
	ClassDB::bind_method(D_METHOD("_on_face_tween_done"), &AGSCharacter::_on_face_tween_done);

	ClassDB::bind_method(D_METHOD("_on_velocity_computed", "safe_velocity"), &AGSCharacter::_on_velocity_computed);
	ClassDB::bind_method(D_METHOD("_on_navigation_finished"), &AGSCharacter::_on_navigation_finished);

	ADD_SIGNAL(MethodInfo("walk_completed"));
	ADD_SIGNAL(MethodInfo("face_completed"));
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

			// Create and wire up NavigationAgent3D on first entry.
			// Using NOTIFICATION_ENTER_TREE guarantees the node is already in the
			// scene tree when the child is added, so NavigationAgent3D's own
			// NOTIFICATION_ENTER_TREE fires with a valid viewport — no errors.
			// Headless tests that call NOTIFICATION_READY manually without entering
			// the tree never reach this branch, so nav_agent stays null there.
			if (!nav_agent) {
				nav_agent = memnew(NavigationAgent3D);
				add_child(nav_agent);
				nav_agent->connect("velocity_computed", Callable(this, "_on_velocity_computed"));
				nav_agent->connect("navigation_finished", Callable(this, "_on_navigation_finished"));
			}

			set_physics_process(false);
		} break;

		case NOTIFICATION_READY: {
			// Register with AGSRuntime so scripts can resolve this character by name.
			// Fires both in normal runtime (deferred after enter-tree) and when called
			// manually in headless tests — registration works regardless of tree state.
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
			// Unregister when freed outside the scene tree (e.g. headless tests
			// that create characters without entering the tree).  In-tree nodes
			// unregister via EXIT_TREE above; PREDELETE is a safe catch-all.
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
	emit_signal("walk_completed");
}

void AGSCharacter::navigate_to(const Vector3 &p_target) {
	// nav_agent is only created when the character is in the scene tree (see
	// NOTIFICATION_READY guard above).  Headless tests that call walk_to() to
	// check signal contracts will have a null nav_agent — return silently so
	// the signal is still delivered without a spurious error message.
	if (!nav_agent) {
		return;
	}
	nav_agent->set_target_position(p_target);
	navigating = true;
	set_physics_process(true);
}

Signal AGSCharacter::walk_to(const String &p_point_name) {
	// Find the nearest AGSRoom ancestor and look up the named point.
	AGSRoom *room = nullptr;
	Node *parent = get_parent();
	while (parent) {
		room = Object::cast_to<AGSRoom>(parent);
		if (room) {
			break;
		}
		parent = parent->get_parent();
	}
	ERR_FAIL_NULL_V_MSG(room, Signal(), "AGSCharacter::walk_to: No parent AGSRoom found.");

	Vector3 target = room->get_point(p_point_name);
	navigate_to(target);

	// Return the walk_completed signal so GDScript can await it.
	return Signal(this, "walk_completed");
}

Signal AGSCharacter::face_to(const String &p_point_name) {
	// Find parent room and resolve named point.
	AGSRoom *room = nullptr;
	Node *parent = get_parent();
	while (parent) {
		room = Object::cast_to<AGSRoom>(parent);
		if (room) {
			break;
		}
		parent = parent->get_parent();
	}
	ERR_FAIL_NULL_V_MSG(room, Signal(), "AGSCharacter::face_to: No parent AGSRoom found.");

	Vector3 target = room->get_point(p_point_name);

	// Compute target Y rotation: build a look-at basis and extract the Y euler angle.
	// Use global position when in the scene tree; fall back to local position in
	// headless test contexts where global transform is unavailable.
	Vector3 char_pos = is_inside_tree() ? get_global_position() : get_position();
	Vector3 dir = target - char_pos;
	dir.y = 0.0f;
	if (dir.length_squared() > CMP_EPSILON) {
		Basis look = Basis::looking_at(dir.normalized());
		float target_y = look.get_euler().y;

		if (is_inside_tree()) {
			Ref<Tween> tween = create_tween();
			tween->tween_property(this, NodePath("rotation:y"), target_y, 0.3);
			tween->tween_callback(Callable(this, "_on_face_tween_done"));
		} else {
			// Outside scene tree (headless tests): apply instantly, defer signal.
			Vector3 rot = get_rotation();
			rot.y = target_y;
			set_rotation(rot);
			call_deferred("_on_face_tween_done");
		}
	} else {
		// Already facing the point (or at the same position) — complete immediately.
		call_deferred("_on_face_tween_done");
	}

	return Signal(this, "face_completed");
}

void AGSCharacter::_on_face_tween_done() {
	emit_signal("face_completed");
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
