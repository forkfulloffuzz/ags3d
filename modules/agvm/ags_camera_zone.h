#pragma once

#include "scene/3d/physics/area_3d.h"

class AGSCharacter;

class AGSCameraZone : public Area3D {
	GDCLASS(AGSCameraZone, Area3D);

	String target_camera;
	bool revert_on_exit = false;
	String _previous_camera; // name of camera active before entry, for revert

	void _on_body_entered(Node3D *p_body);
	void _on_body_exited(Node3D *p_body);

protected:
	static void _bind_methods();
	void _notification(int p_what);

public:
	void set_target_camera(const String &p_name);
	String get_target_camera() const;

	void set_revert_on_exit(bool p_revert);
	bool get_revert_on_exit() const;
};
