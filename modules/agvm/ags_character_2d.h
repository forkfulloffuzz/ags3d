#pragma once

#include "ags_character_base.h"

// AGSCharacter2D is a character rendered as a 2D billboard sprite (Sprite3D).
// The sprite faces the camera each frame; direction selection and frame cycling
// are driven by ags_animation_player_2d.gd (T-GS29).
//
// Currently a stub — visual_mode property and Sprite3D setup are added in T-GS24/T-GS25.
class AGSCharacter2D : public AGSCharacterBase {
	GDCLASS(AGSCharacter2D, AGSCharacterBase);

protected:
	static void _bind_methods() {}
};
