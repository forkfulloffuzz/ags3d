#include "ags_camera.h"

#include "ags_runtime.h"
#include "core/object/class_db.h"

void AGSCamera::_bind_methods() {
	ClassDB::bind_method(D_METHOD("set_camera_name", "name"), &AGSCamera::set_camera_name);
	ClassDB::bind_method(D_METHOD("get_camera_name"), &AGSCamera::get_camera_name);
	ADD_PROPERTY(PropertyInfo(Variant::STRING, "camera_name"), "set_camera_name", "get_camera_name");
}

void AGSCamera::_notification(int p_what) {
	switch (p_what) {
		case NOTIFICATION_READY: {
			if (AGSRuntime::get_singleton()) {
				AGSRuntime::get_singleton()->register_camera(this);
			}
		} break;
		case NOTIFICATION_EXIT_TREE: {
			if (AGSRuntime::get_singleton()) {
				AGSRuntime::get_singleton()->unregister_camera(this);
			}
		} break;
		case NOTIFICATION_PREDELETE: {
			if (AGSRuntime::get_singleton()) {
				AGSRuntime::get_singleton()->unregister_camera(this);
			}
		} break;
	}
}

void AGSCamera::set_camera_name(const String &p_name) {
	camera_name = p_name;
}

String AGSCamera::get_camera_name() const {
	return camera_name;
}
