#pragma once

#include "scene/3d/physics/character_body_3d.h"

// AGSCharacterBase is the shared base for all AGS character types.
// It holds the runtime properties (character_name, move_speed, say_text),
// the shared signals (walk_completed, face_completed, say_completed), and
// the AGSRuntime registration/unregistration logic.
//
// Concrete subclasses:
//   AGSCharacter3D — 3D mesh + CharacterBody3D navigation (classic polygonal character)
//   AGSCharacter2D — 2D billboard sprite (Sprite3D facing the camera)
class AGSCharacterBase : public CharacterBody3D {
	GDCLASS(AGSCharacterBase, CharacterBody3D);

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
