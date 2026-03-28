#pragma once

#include "core/templates/hash_map.h"
#include "scene/3d/node_3d.h"

class AGSHotspot;
class AGSPoint;
class AGSTriggerRegion;

class AGSRoom : public Node3D {
	GDCLASS(AGSRoom, Node3D);

	String room_name;
	String initial_camera;
	HashMap<StringName, AGSPoint *> points;
	HashMap<StringName, AGSTriggerRegion *> regions;
	HashMap<StringName, AGSHotspot *> hotspots;

protected:
	static void _bind_methods();
	void _notification(int p_what);

public:
	void set_room_name(const String &p_name);
	String get_room_name() const;

	void set_initial_camera(const String &p_name);
	String get_initial_camera() const;

	void register_point(AGSPoint *p_point);
	void unregister_point(AGSPoint *p_point);
	Vector3 get_point(const String &p_name) const;

	void register_region(AGSTriggerRegion *p_region);
	void unregister_region(AGSTriggerRegion *p_region);

	void register_hotspot(AGSHotspot *p_hotspot);
	void unregister_hotspot(AGSHotspot *p_hotspot);

private:
	// Bridge: drops the Node3D body from region_entered/region_exited and passes
	// the region name (bound via Callable::bind) to the GDScript event handler.
	void _on_region_body_entered(Node3D *p_body, const String &p_region_name);
	void _on_region_body_exited(Node3D *p_body, const String &p_region_name);
};
