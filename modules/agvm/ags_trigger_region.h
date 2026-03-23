#pragma once

#include "scene/3d/physics/area_3d.h"

class AGSTriggerRegion : public Area3D {
	GDCLASS(AGSTriggerRegion, Area3D);

	String region_name;

	void _on_body_entered(Node3D *p_body);
	void _on_body_exited(Node3D *p_body);

protected:
	static void _bind_methods();
	void _notification(int p_what);

public:
	void set_region_name(const String &p_name);
	String get_region_name() const;
};
