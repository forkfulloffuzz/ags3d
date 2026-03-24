#pragma once

#include "scene/3d/physics/character_body_3d.h"

class AGSCharacter : public CharacterBody3D {
	GDCLASS(AGSCharacter, CharacterBody3D);

	String character_name;

protected:
	static void _bind_methods();
	void _notification(int p_what);

public:
	void set_character_name(const String &p_name);
	String get_character_name() const;
};
