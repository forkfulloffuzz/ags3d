#pragma once

#include "scene/3d/mesh_instance_3d.h"
#include "scene/3d/navigation/navigation_region_3d.h"
#include "scene/3d/physics/static_body_3d.h"

class AGSWalkableSurface : public StaticBody3D {
	GDCLASS(AGSWalkableSurface, StaticBody3D);

	NavigationRegion3D *nav_region = nullptr;

	void _apply_editor_overlay();
	void _setup_navmesh();

protected:
	static void _bind_methods();
	void _notification(int p_what);
};
