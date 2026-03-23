#pragma once

#include "core/templates/hash_map.h"
#include "scene/3d/node_3d.h"

class AGSPoint;

class AGSRoom : public Node3D {
	GDCLASS(AGSRoom, Node3D);

	String room_name;
	HashMap<StringName, AGSPoint *> points;

protected:
	static void _bind_methods();

public:
	void set_room_name(const String &p_name);
	String get_room_name() const;

	void register_point(AGSPoint *p_point);
	void unregister_point(AGSPoint *p_point);
	Vector3 get_point(const String &p_name) const;
};
