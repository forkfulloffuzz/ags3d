#include "ags_room.h"

#include "ags_hotspot.h"
#include "ags_point.h"
#include "ags_runtime.h"
#include "ags_trigger_region.h"
#include "core/object/class_db.h"

void AGSRoom::_bind_methods() {
	ClassDB::bind_method(D_METHOD("set_room_name", "name"), &AGSRoom::set_room_name);
	ClassDB::bind_method(D_METHOD("get_room_name"), &AGSRoom::get_room_name);
	ADD_PROPERTY(PropertyInfo(Variant::STRING, "room_name"), "set_room_name", "get_room_name");

	ClassDB::bind_method(D_METHOD("get_point", "name"), &AGSRoom::get_point);

	ADD_SIGNAL(MethodInfo("hotspot_clicked", PropertyInfo(Variant::STRING, "hotspot_name")));
	ADD_SIGNAL(MethodInfo("room_enter"));
}

void AGSRoom::_notification(int p_what) {
	switch (p_what) {
		case NOTIFICATION_READY: {
			if (AGSRuntime::get_singleton()) {
				AGSRuntime::get_singleton()->register_room(this);
			}

			// T33 — bind AGS-spirit event handlers from the attached script.
			// Script handlers are called directly to avoid a signal/method name
			// collision (e.g. signal "room_enter" vs. GDScript func "room_enter").
			// The signals are still emitted for external listeners.

			if (has_method("hotspot_interact")) {
				connect("hotspot_clicked", Callable(this, "hotspot_interact"));
			}

			// Connect each child AGSTriggerRegion's signals to the room's handlers.
			for (int i = 0; i < get_child_count(); i++) {
				AGSTriggerRegion *region = Object::cast_to<AGSTriggerRegion>(get_child(i));
				if (!region) {
					continue;
				}
				if (has_method("region_walked_into")) {
					region->connect("region_entered", Callable(this, "region_walked_into"));
				}
				if (has_method("region_walked_off")) {
					region->connect("region_exited", Callable(this, "region_walked_off"));
				}
			}

			// Call room_load first (fires before room_enter in AGS semantics).
			if (has_method("room_load")) {
				call("room_load");
			}
			// Call room_enter directly — do not connect via signal to avoid the
			// "Method not found" error Godot raises when the signal name and the
			// GDScript method name are identical and the callable is dispatched
			// through the signal machinery.
			emit_signal("room_enter"); // for external listeners
			if (has_method("room_enter")) {
				call("room_enter");
			}
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
