#pragma once

#include "scene/3d/node_3d.h"

class AGSRoom : public Node3D {
	GDCLASS(AGSRoom, Node3D);

	String room_name;

protected:
	static void _bind_methods();

public:
	void set_room_name(const String &p_name);
	String get_room_name() const;
};
