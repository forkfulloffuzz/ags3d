#pragma once

#include "ags_character_base.h"

// AGSCharacter3D is a character rendered as a 3D mesh (polygonal model).
// On first enter-tree it creates a placeholder capsule mesh so the character
// is visible in the editor before a real mesh asset is assigned.
// Navigation and animation are driven by the ags_character.gd runtime script.
class AGSCharacter3D : public AGSCharacterBase {
	GDCLASS(AGSCharacter3D, AGSCharacterBase);

protected:
	static void _bind_methods() {}
	void _notification(int p_what);
};
