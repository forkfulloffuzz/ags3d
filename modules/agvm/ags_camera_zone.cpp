#include "ags_camera_zone.h"

#include "ags_camera.h"
#include "ags_character.h"
#include "ags_runtime.h"
#include "ags_trace.h"
#include "core/object/class_db.h"

void AGSCameraZone::_bind_methods() {
	ClassDB::bind_method(D_METHOD("set_target_camera", "name"), &AGSCameraZone::set_target_camera);
	ClassDB::bind_method(D_METHOD("get_target_camera"), &AGSCameraZone::get_target_camera);
	ADD_PROPERTY(PropertyInfo(Variant::STRING, "target_camera"), "set_target_camera", "get_target_camera");

	ClassDB::bind_method(D_METHOD("set_revert_on_exit", "revert"), &AGSCameraZone::set_revert_on_exit);
	ClassDB::bind_method(D_METHOD("get_revert_on_exit"), &AGSCameraZone::get_revert_on_exit);
	ADD_PROPERTY(PropertyInfo(Variant::BOOL, "revert_on_exit"), "set_revert_on_exit", "get_revert_on_exit");

	ClassDB::bind_method(D_METHOD("_on_body_entered", "body"), &AGSCameraZone::_on_body_entered);
	ClassDB::bind_method(D_METHOD("_on_body_exited", "body"), &AGSCameraZone::_on_body_exited);
}

void AGSCameraZone::_notification(int p_what) {
	if (p_what == NOTIFICATION_READY) {
		connect("body_entered", Callable(this, "_on_body_entered"));
		connect("body_exited", Callable(this, "_on_body_exited"));
	}
}

void AGSCameraZone::_on_body_entered(Node3D *p_body) {
	if (!Object::cast_to<AGSCharacter>(p_body)) {
		return;
	}
	if (target_camera.is_empty()) {
		return;
	}
	AGSRuntime *runtime = AGSRuntime::get_singleton();
	if (!runtime) {
		return;
	}
	// Remember the current camera for revert.
	if (revert_on_exit) {
		_previous_camera = "";
		for (const KeyValue<StringName, AGSCamera *> &kv : runtime->get_cameras()) {
			if (kv.value->is_current()) {
				_previous_camera = kv.value->get_camera_name();
				break;
			}
		}
	}
	AGS_TRACE("AGSCameraZone", "_on_body_entered", vformat("switching to '%s'", target_camera))
	runtime->set_camera(target_camera);
}

void AGSCameraZone::_on_body_exited(Node3D *p_body) {
	if (!revert_on_exit) {
		return;
	}
	if (!Object::cast_to<AGSCharacter>(p_body)) {
		return;
	}
	if (_previous_camera.is_empty()) {
		return;
	}
	AGSRuntime *runtime = AGSRuntime::get_singleton();
	if (!runtime) {
		return;
	}
	AGS_TRACE("AGSCameraZone", "_on_body_exited", vformat("reverting to '%s'", _previous_camera))
	runtime->set_camera(_previous_camera);
}

void AGSCameraZone::set_target_camera(const String &p_name) {
	target_camera = p_name;
}

String AGSCameraZone::get_target_camera() const {
	return target_camera;
}

void AGSCameraZone::set_revert_on_exit(bool p_revert) {
	revert_on_exit = p_revert;
}

bool AGSCameraZone::get_revert_on_exit() const {
	return revert_on_exit;
}
