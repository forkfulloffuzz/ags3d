#include "ags_point.h"

#include "ags_room.h"

void AGSPoint::_bind_methods() {
	ClassDB::bind_method(D_METHOD("set_point_name", "name"), &AGSPoint::set_point_name);
	ClassDB::bind_method(D_METHOD("get_point_name"), &AGSPoint::get_point_name);
	ADD_PROPERTY(PropertyInfo(Variant::STRING, "point_name"), "set_point_name", "get_point_name");
}

void AGSPoint::_notification(int p_what) {
	switch (p_what) {
		case NOTIFICATION_READY: {
			// Walk up the tree to find the nearest AGSRoom parent and register.
			Node *parent = get_parent();
			while (parent) {
				AGSRoom *room = Object::cast_to<AGSRoom>(parent);
				if (room) {
					room->register_point(this);
					break;
				}
				parent = parent->get_parent();
			}
		} break;

		case NOTIFICATION_EXIT_TREE: {
			Node *parent = get_parent();
			while (parent) {
				AGSRoom *room = Object::cast_to<AGSRoom>(parent);
				if (room) {
					room->unregister_point(this);
					break;
				}
				parent = parent->get_parent();
			}
		} break;
	}
}

void AGSPoint::set_point_name(const String &p_name) {
	point_name = p_name;
}

String AGSPoint::get_point_name() const {
	return point_name;
}
