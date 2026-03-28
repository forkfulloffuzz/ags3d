#include "ags_spawn_point.h"

#include "ags_character.h"
#include "ags_runtime.h"
#include "ags_trace.h"
#include "core/object/class_db.h"

void AGSSpawnPoint::_bind_methods() {
	ClassDB::bind_method(D_METHOD("set_spawn_character", "name"), &AGSSpawnPoint::set_spawn_character);
	ClassDB::bind_method(D_METHOD("get_spawn_character"), &AGSSpawnPoint::get_spawn_character);
	ADD_PROPERTY(PropertyInfo(Variant::STRING, "spawn_character"), "set_spawn_character", "get_spawn_character");
}

void AGSSpawnPoint::_notification(int p_what) {
	if (p_what != NOTIFICATION_READY) {
		return;
	}
	if (spawn_character.is_empty()) {
		return;
	}
	AGSRuntime *runtime = AGSRuntime::get_singleton();
	if (!runtime) {
		return;
	}
	AGSCharacter *character = runtime->get_character(spawn_character);
	if (!character) {
		WARN_PRINT(vformat("AGSSpawnPoint: character '%s' not found in AGSRuntime.", spawn_character));
		return;
	}
	// Use global position when in the scene tree; fall back to local position
	// in headless test contexts where global transform is unavailable.
	AGS_TRACE("AGSSpawnPoint", "_notification", vformat("spawning '%s' at %s", spawn_character, get_global_position()))
	if (is_inside_tree()) {
		character->set_global_position(get_global_position());
	} else {
		character->set_position(get_position());
	}
	AGS_TRACE("AGSSpawnPoint", "_notification", "set_global_position done")
}

void AGSSpawnPoint::set_spawn_character(const String &p_name) {
	spawn_character = p_name;
}

String AGSSpawnPoint::get_spawn_character() const {
	return spawn_character;
}
