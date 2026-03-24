#include "ags_trigger_region.h"

#include "ags_room.h"
#include "core/object/class_db.h"

void AGSTriggerRegion::_bind_methods() {
	ClassDB::bind_method(D_METHOD("set_region_name", "name"), &AGSTriggerRegion::set_region_name);
	ClassDB::bind_method(D_METHOD("get_region_name"), &AGSTriggerRegion::get_region_name);
	ADD_PROPERTY(PropertyInfo(Variant::STRING, "region_name"), "set_region_name", "get_region_name");

	ClassDB::bind_method(D_METHOD("_on_body_entered", "body"), &AGSTriggerRegion::_on_body_entered);
	ClassDB::bind_method(D_METHOD("_on_body_exited", "body"), &AGSTriggerRegion::_on_body_exited);

	ADD_SIGNAL(MethodInfo("region_entered", PropertyInfo(Variant::OBJECT, "character", PROPERTY_HINT_NODE_TYPE, "Node3D")));
	ADD_SIGNAL(MethodInfo("region_exited", PropertyInfo(Variant::OBJECT, "character", PROPERTY_HINT_NODE_TYPE, "Node3D")));
}

void AGSTriggerRegion::_notification(int p_what) {
	switch (p_what) {
		case NOTIFICATION_READY: {
			connect("body_entered", Callable(this, "_on_body_entered"));
			connect("body_exited", Callable(this, "_on_body_exited"));

			// Register with parent AGSRoom.
			Node *parent = get_parent();
			while (parent) {
				AGSRoom *room = Object::cast_to<AGSRoom>(parent);
				if (room) {
					room->register_region(this);
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
					room->unregister_region(this);
					break;
				}
				parent = parent->get_parent();
			}
		} break;
	}
}

void AGSTriggerRegion::_on_body_entered(Node3D *p_body) {
	emit_signal("region_entered", p_body);
}

void AGSTriggerRegion::_on_body_exited(Node3D *p_body) {
	emit_signal("region_exited", p_body);
}

void AGSTriggerRegion::set_region_name(const String &p_name) {
	region_name = p_name;
}

String AGSTriggerRegion::get_region_name() const {
	return region_name;
}
