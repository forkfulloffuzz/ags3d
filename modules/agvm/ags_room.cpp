#include "ags_room.h"

#include "ags_point.h"
#include "ags_trigger_region.h"

void AGSRoom::_bind_methods() {
	ClassDB::bind_method(D_METHOD("set_room_name", "name"), &AGSRoom::set_room_name);
	ClassDB::bind_method(D_METHOD("get_room_name"), &AGSRoom::get_room_name);
	ADD_PROPERTY(PropertyInfo(Variant::STRING, "room_name"), "set_room_name", "get_room_name");

	ClassDB::bind_method(D_METHOD("get_point", "name"), &AGSRoom::get_point);
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
	ERR_FAIL_COND_V_MSG(!found, Vector3(), vformat("AGSRoom: no point named '%s'.", p_name));
	return (*found)->get_global_position();
}

void AGSRoom::register_region(AGSTriggerRegion *p_region) {
	ERR_FAIL_NULL(p_region);
	regions[p_region->get_region_name()] = p_region;
}

void AGSRoom::unregister_region(AGSTriggerRegion *p_region) {
	ERR_FAIL_NULL(p_region);
	regions.erase(p_region->get_region_name());
}
