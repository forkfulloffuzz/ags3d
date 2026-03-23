#pragma once

#include "scene/3d/camera_3d.h"
#include "scene/3d/physics/area_3d.h"

class AGSHotspot : public Area3D {
	GDCLASS(AGSHotspot, Area3D);

	String hotspot_name;

protected:
	static void _bind_methods();
	void _notification(int p_what);
	void _input_event_call(Camera3D *p_camera, const Ref<InputEvent> &p_event,
			const Vector3 &p_pos, const Vector3 &p_normal, int p_shape) override;

public:
	void set_hotspot_name(const String &p_name);
	String get_hotspot_name() const;

	// Programmatically fire hotspot_clicked — used in headless tests.
	void simulate_click();
};
