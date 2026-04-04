#pragma once

#include "scene/main/node.h"

// AGSItem represents a named inventory item definition.
// It holds data about an item (name, display name, description) but has no
// scene presence — items are not placed in rooms directly. AGSRoomItem (T-GS03)
// is the in-room counterpart.
//
// AGSItem nodes are registered with AGSRuntime at NOTIFICATION_READY so scripts
// can call AGSRuntime.get_item("rusty_key") to get the item data.
class AGSItem : public Node {
	GDCLASS(AGSItem, Node);

	String item_name;
	String display_name;
	String description;

protected:
	static void _bind_methods();
	void _notification(int p_what);

public:
	void set_item_name(const String &p_name);
	String get_item_name() const;

	void set_display_name(const String &p_name);
	String get_display_name() const;

	void set_description(const String &p_desc);
	String get_description() const;
};
