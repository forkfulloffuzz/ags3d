#pragma once

#include "scene/3d/mesh_instance_3d.h"
#include "scene/3d/physics/static_body_3d.h"

class AGSBlockerVolume : public StaticBody3D {
	GDCLASS(AGSBlockerVolume, StaticBody3D);

protected:
	static void _bind_methods();
	void _notification(int p_what);

private:
	void _apply_editor_overlay();
};
