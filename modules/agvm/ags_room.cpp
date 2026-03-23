#include "ags_room.h"

void AGSRoom::_bind_methods() {
	ClassDB::bind_method(D_METHOD("set_room_name", "name"), &AGSRoom::set_room_name);
	ClassDB::bind_method(D_METHOD("get_room_name"), &AGSRoom::get_room_name);
	ADD_PROPERTY(PropertyInfo(Variant::STRING, "room_name"), "set_room_name", "get_room_name");
}

void AGSRoom::set_room_name(const String &p_name) {
	room_name = p_name;
}

String AGSRoom::get_room_name() const {
	return room_name;
}
