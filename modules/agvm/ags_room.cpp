#include "ags_room.h"

#include "ags_camera.h"
#include "core/object/callable_mp.h"
#include "scene/3d/node_3d.h"
#include "ags_hotspot.h"
#include "ags_point.h"
#include "ags_room_item.h"
#include "ags_runtime.h"
#include "ags_trace.h"
#include "ags_trigger_region.h"
#include "core/object/class_db.h"

void AGSRoom::_bind_methods() {
	ClassDB::bind_method(D_METHOD("set_room_name", "name"), &AGSRoom::set_room_name);
	ClassDB::bind_method(D_METHOD("get_room_name"), &AGSRoom::get_room_name);
	ADD_PROPERTY(PropertyInfo(Variant::STRING, "room_name"), "set_room_name", "get_room_name");

	ClassDB::bind_method(D_METHOD("set_initial_camera", "name"), &AGSRoom::set_initial_camera);
	ClassDB::bind_method(D_METHOD("get_initial_camera"), &AGSRoom::get_initial_camera);
	ADD_PROPERTY(PropertyInfo(Variant::STRING, "initial_camera"), "set_initial_camera", "get_initial_camera");

	ClassDB::bind_method(D_METHOD("get_point", "name"), &AGSRoom::get_point);

	ADD_SIGNAL(MethodInfo("hotspot_clicked", PropertyInfo(Variant::STRING, "hotspot_name")));
	ADD_SIGNAL(MethodInfo("item_clicked", PropertyInfo(Variant::STRING, "item_name")));
	ADD_SIGNAL(MethodInfo("room_enter"));
}

void AGSRoom::_notification(int p_what) {
	switch (p_what) {
		case NOTIFICATION_READY: {
			if (AGSRuntime::get_singleton()) {
				AGSRuntime::get_singleton()->register_room(this);
			}

			// Activate the initial camera before any room script runs.
			// Children's NOTIFICATION_READY fires before the parent's, so the
			// AGSCamera is guaranteed to be registered by the time we get here.
			if (!initial_camera.is_empty() && AGSRuntime::get_singleton()) {
				AGSCamera *cam = AGSRuntime::get_singleton()->get_camera(initial_camera);
				if (cam) {
					cam->make_current();
					AGS_TRACE("AGSRoom", "_notification", vformat("activated camera '%s'", initial_camera))
				} else {
					WARN_PRINT(vformat("AGSRoom: initial_camera '%s' not found.", initial_camera));
				}
			}

			// T33 — bind AGS-spirit event handlers from the attached script.
			// Script handlers are called directly to avoid a signal/method name
			// collision (e.g. signal "room_enter" vs. GDScript func "room_enter").
			// The signals are still emitted for external listeners.

			if (has_method("hotspot_interact")) {
				connect("hotspot_clicked", Callable(this, "hotspot_interact"));
			}
			if (has_method("item_interact")) {
				connect("item_clicked", Callable(this, "item_interact"));
			}

			// Connect each child AGSTriggerRegion's signals to the room's handlers.
			// region_entered/region_exited fire with a Node3D body; the bridge methods
			// receive (body, region_name) and forward only region_name to the GDScript handler.
			for (int i = 0; i < get_child_count(); i++) {
				AGSTriggerRegion *region = Object::cast_to<AGSTriggerRegion>(get_child(i));
				if (!region) {
					continue;
				}
				if (has_method("region_walked_into")) {
					region->connect("region_entered",
							callable_mp(this, &AGSRoom::_on_region_body_entered).bind(region->get_region_name()));
				}
				if (has_method("region_walked_off")) {
					region->connect("region_exited",
							callable_mp(this, &AGSRoom::_on_region_body_exited).bind(region->get_region_name()));
				}
			}

			// Call room_load first (fires before room_enter in AGS semantics).
			AGS_TRACE("AGSRoom", "_notification", vformat("room_name=%s, calling room_load", room_name))
			if (has_method("room_load")) {
				call("room_load");
			}
			// Call room_enter directly — do not connect via signal to avoid the
			// "Method not found" error Godot raises when the signal name and the
			// GDScript method name are identical and the callable is dispatched
			// through the signal machinery.
			AGS_TRACE("AGSRoom", "_notification", "calling room_enter")
			emit_signal("room_enter"); // for external listeners
			if (has_method("room_enter")) {
				call("room_enter");
			}
			AGS_TRACE("AGSRoom", "_notification", "room_enter done")
		} break;
		case NOTIFICATION_EXIT_TREE: {
			if (AGSRuntime::get_singleton()) {
				AGSRuntime::get_singleton()->unregister_room(this);
			}
		} break;
		case NOTIFICATION_PREDELETE: {
			if (AGSRuntime::get_singleton()) {
				AGSRuntime::get_singleton()->unregister_room(this);
			}
		} break;
	}
}

void AGSRoom::set_room_name(const String &p_name) {
	room_name = p_name;
}

String AGSRoom::get_room_name() const {
	return room_name;
}

void AGSRoom::set_initial_camera(const String &p_name) {
	initial_camera = p_name;
}

String AGSRoom::get_initial_camera() const {
	return initial_camera;
}

void AGSRoom::register_point(AGSPoint *p_point) {
	ERR_FAIL_NULL(p_point);
	points[p_point->get_point_name()] = p_point;
}

void AGSRoom::unregister_point(AGSPoint *p_point) {
	ERR_FAIL_NULL(p_point);
	points.erase(p_point->get_point_name());
}

Vector3 AGSRoom::get_point(const String &p_name) const {
	const AGSPoint *const *found = points.getptr(StringName(p_name));
	if (!found) {
		return Vector3();
	}
	const AGSPoint *pt = *found;
	return pt->is_inside_tree() ? pt->get_global_position() : pt->get_position();
}

void AGSRoom::_on_region_body_entered(Node3D *p_body, const String &p_region_name) {
	if (has_method("region_walked_into")) {
		call("region_walked_into", p_region_name);
	}
}

void AGSRoom::_on_region_body_exited(Node3D *p_body, const String &p_region_name) {
	if (has_method("region_walked_off")) {
		call("region_walked_off", p_region_name);
	}
}

void AGSRoom::register_region(AGSTriggerRegion *p_region) {
	ERR_FAIL_NULL(p_region);
	regions[p_region->get_region_name()] = p_region;
}

void AGSRoom::unregister_region(AGSTriggerRegion *p_region) {
	ERR_FAIL_NULL(p_region);
	regions.erase(p_region->get_region_name());
}

void AGSRoom::register_hotspot(AGSHotspot *p_hotspot) {
	ERR_FAIL_NULL(p_hotspot);
	hotspots[p_hotspot->get_hotspot_name()] = p_hotspot;
}

void AGSRoom::unregister_hotspot(AGSHotspot *p_hotspot) {
	ERR_FAIL_NULL(p_hotspot);
	hotspots.erase(p_hotspot->get_hotspot_name());
}

void AGSRoom::register_room_item(AGSRoomItem *p_item) {
	ERR_FAIL_NULL(p_item);
	room_items[p_item->get_item_name()] = p_item;
}

void AGSRoom::unregister_room_item(AGSRoomItem *p_item) {
	ERR_FAIL_NULL(p_item);
	room_items.erase(p_item->get_item_name());
}
