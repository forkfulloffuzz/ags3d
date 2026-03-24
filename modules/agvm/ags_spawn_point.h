#pragma once

#include "scene/3d/node_3d.h"

class AGSSpawnPoint : public Node3D {
	GDCLASS(AGSSpawnPoint, Node3D);

	String spawn_character;

protected:
	static void _bind_methods();
	void _notification(int p_what);

public:
	void set_spawn_character(const String &p_name);
	String get_spawn_character() const;
};
