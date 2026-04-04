#include "ags_item.h"

#include "ags_runtime.h"
#include "ags_trace.h"
#include "core/object/class_db.h"

void AGSItem::_bind_methods() {
	ClassDB::bind_method(D_METHOD("set_item_name", "name"), &AGSItem::set_item_name);
	ClassDB::bind_method(D_METHOD("get_item_name"), &AGSItem::get_item_name);
	ADD_PROPERTY(PropertyInfo(Variant::STRING, "item_name"), "set_item_name", "get_item_name");

	ClassDB::bind_method(D_METHOD("set_display_name", "name"), &AGSItem::set_display_name);
	ClassDB::bind_method(D_METHOD("get_display_name"), &AGSItem::get_display_name);
	ADD_PROPERTY(PropertyInfo(Variant::STRING, "display_name"), "set_display_name", "get_display_name");

	ClassDB::bind_method(D_METHOD("set_description", "desc"), &AGSItem::set_description);
	ClassDB::bind_method(D_METHOD("get_description"), &AGSItem::get_description);
	ADD_PROPERTY(PropertyInfo(Variant::STRING, "description"), "set_description", "get_description");
}

void AGSItem::_notification(int p_what) {
	switch (p_what) {
		case NOTIFICATION_READY: {
			if (AGSRuntime::get_singleton()) {
				AGSRuntime::get_singleton()->register_item(this);
			}
		} break;
		case NOTIFICATION_EXIT_TREE:
		case NOTIFICATION_PREDELETE: {
			if (AGSRuntime::get_singleton()) {
				AGSRuntime::get_singleton()->unregister_item(this);
			}
		} break;
	}
}

void AGSItem::set_item_name(const String &p_name) { item_name = p_name; }
String AGSItem::get_item_name() const { return item_name; }

void AGSItem::set_display_name(const String &p_name) { display_name = p_name; }
String AGSItem::get_display_name() const { return display_name; }

void AGSItem::set_description(const String &p_desc) { description = p_desc; }
String AGSItem::get_description() const { return description; }
