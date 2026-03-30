#pragma once

#include "scene/3d/physics/character_body_3d.h"

class AGSCharacter : public CharacterBody3D {
	GDCLASS(AGSCharacter, CharacterBody3D);

	String character_name;
	float move_speed = 4.0f;
	String say_text;

protected:
	static void _bind_methods();
	void _notification(int p_what);

public:
	void set_character_name(const String &p_name);
	String get_character_name() const;

	void set_move_speed(float p_speed);
	float get_move_speed() const;

	void set_say_text(const String &p_text);
	String get_say_text() const;
};
