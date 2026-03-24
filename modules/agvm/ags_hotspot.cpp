#include "ags_hotspot.h"

#include "ags_room.h"
#include "core/input/input_event.h"
#include "core/object/class_db.h"

void AGSHotspot::_bind_methods() {
	ClassDB::bind_method(D_METHOD("set_hotspot_name", "name"), &AGSHotspot::set_hotspot_name);
	ClassDB::bind_method(D_METHOD("get_hotspot_name"), &AGSHotspot::get_hotspot_name);
	ADD_PROPERTY(PropertyInfo(Variant::STRING, "hotspot_name"), "set_hotspot_name", "get_hotspot_name");

	ADD_SIGNAL(MethodInfo("hotspot_clicked", PropertyInfo(Variant::STRING, "hotspot_name")));

	ClassDB::bind_method(D_METHOD("simulate_click"), &AGSHotspot::simulate_click);
}

void AGSHotspot::_notification(int p_what) {
	switch (p_what) {
		case NOTIFICATION_READY: {
			// Enable mouse/touch picking on this area.
			set_ray_pickable(true);

			// Self-register with the nearest AGSRoom ancestor.
			Node *parent = get_parent();
			while (parent) {
				AGSRoom *room = Object::cast_to<AGSRoom>(parent);
				if (room) {
					room->register_hotspot(this);
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
					room->unregister_hotspot(this);
					break;
				}
				parent = parent->get_parent();
			}
		} break;
	}
}

void AGSHotspot::_input_event_call(Camera3D *p_camera, const Ref<InputEvent> &p_event,
		const Vector3 &p_pos, const Vector3 &p_normal, int p_shape) {
	const InputEventMouseButton *mb = Object::cast_to<InputEventMouseButton>(*p_event);
	if (mb && mb->is_pressed() && mb->get_button_index() == MouseButton::LEFT) {
		emit_signal("hotspot_clicked", hotspot_name);

		// Also notify the parent AGSRoom so room scripts get the event.
		Node *parent = get_parent();
		while (parent) {
			AGSRoom *room = Object::cast_to<AGSRoom>(parent);
			if (room) {
				room->emit_signal("hotspot_clicked", hotspot_name);
				break;
			}
			parent = parent->get_parent();
		}
	}
}

void AGSHotspot::simulate_click() {
	emit_signal("hotspot_clicked", hotspot_name);

	Node *parent = get_parent();
	while (parent) {
		AGSRoom *room = Object::cast_to<AGSRoom>(parent);
		if (room) {
			room->emit_signal("hotspot_clicked", hotspot_name);
			break;
		}
		parent = parent->get_parent();
	}
}

void AGSHotspot::set_hotspot_name(const String &p_name) {
	hotspot_name = p_name;
}

String AGSHotspot::get_hotspot_name() const {
	return hotspot_name;
}
