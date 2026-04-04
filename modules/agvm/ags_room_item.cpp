#include "ags_room_item.h"

#include "ags_room.h"
#include "core/input/input_event.h"
#include "core/object/class_db.h"
#include "scene/main/node.h"

void AGSRoomItem::_bind_methods() {
	ClassDB::bind_method(D_METHOD("set_item_name", "name"), &AGSRoomItem::set_item_name);
	ClassDB::bind_method(D_METHOD("get_item_name"), &AGSRoomItem::get_item_name);
	ADD_PROPERTY(PropertyInfo(Variant::STRING, "item_name"), "set_item_name", "get_item_name");

	ADD_SIGNAL(MethodInfo("item_clicked", PropertyInfo(Variant::STRING, "item_name")));

	ClassDB::bind_method(D_METHOD("simulate_click"), &AGSRoomItem::simulate_click);
}

void AGSRoomItem::_notification(int p_what) {
	switch (p_what) {
		case NOTIFICATION_READY: {
			set_ray_pickable(true);
			// Register with the nearest AGSRoom ancestor.
			Node *parent = get_parent();
			while (parent) {
				AGSRoom *room = Object::cast_to<AGSRoom>(parent);
				if (room) {
					room->register_room_item(this);
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
					room->unregister_room_item(this);
					break;
				}
				parent = parent->get_parent();
			}
		} break;
	}
}

void AGSRoomItem::_input_event_call(Camera3D *p_camera, const Ref<InputEvent> &p_event,
		const Vector3 &p_pos, const Vector3 &p_normal, int p_shape) {
	const InputEventMouseButton *mb = Object::cast_to<InputEventMouseButton>(*p_event);
	if (mb && mb->is_pressed() && mb->get_button_index() == MouseButton::LEFT) {
		emit_signal("item_clicked", item_name);

		// Notify the parent AGSRoom so room scripts get item_interact().
		Node *parent = get_parent();
		while (parent) {
			AGSRoom *room = Object::cast_to<AGSRoom>(parent);
			if (room) {
				room->emit_signal("item_clicked", item_name);
				break;
			}
			parent = parent->get_parent();
		}
	}
}

void AGSRoomItem::simulate_click() {
	emit_signal("item_clicked", item_name);

	Node *parent = get_parent();
	while (parent) {
		AGSRoom *room = Object::cast_to<AGSRoom>(parent);
		if (room) {
			room->emit_signal("item_clicked", item_name);
			break;
		}
		parent = parent->get_parent();
	}
}

void AGSRoomItem::set_item_name(const String &p_name) { item_name = p_name; }
String AGSRoomItem::get_item_name() const { return item_name; }
