#pragma once

#include "scene/3d/node_3d.h"

class AGSPoint : public Node3D {
	GDCLASS(AGSPoint, Node3D);

	String point_name;

protected:
	static void _bind_methods();
	void _notification(int p_what);

public:
	void set_point_name(const String &p_name);
	String get_point_name() const;
};
