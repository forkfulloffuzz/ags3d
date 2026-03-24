#pragma once

#include "scene/3d/navigation/navigation_agent_3d.h"
#include "scene/3d/physics/character_body_3d.h"

class AGSCharacter : public CharacterBody3D {
	GDCLASS(AGSCharacter, CharacterBody3D);

	String character_name;
	float move_speed = 4.0f;

	NavigationAgent3D *nav_agent = nullptr;
	bool navigating = false;

	void _on_velocity_computed(const Vector3 &p_safe_velocity);
	void _on_navigation_finished();

protected:
	static void _bind_methods();
	void _notification(int p_what);
	void _physics_process(double p_delta);

public:
	void set_character_name(const String &p_name);
	String get_character_name() const;

	void set_move_speed(float p_speed);
	float get_move_speed() const;

	// Navigate to a world-space position along the navmesh.
	void navigate_to(const Vector3 &p_target);
};
