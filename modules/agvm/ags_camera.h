#pragma once

#include "scene/3d/camera_3d.h"

class AGSCamera : public Camera3D {
	GDCLASS(AGSCamera, Camera3D);

	String camera_name;

protected:
	static void _bind_methods();
	void _notification(int p_what);

public:
	void set_camera_name(const String &p_name);
	String get_camera_name() const;
};
