#include "ags_runtime.h"

#include "ags_character.h"
#include "ags_room.h"
#include "core/object/class_db.h"

AGSRuntime *AGSRuntime::singleton = nullptr;

AGSRuntime::AGSRuntime() {
	ERR_FAIL_COND(singleton != nullptr);
	singleton = this;
}

AGSRuntime::~AGSRuntime() {
	singleton = nullptr;
}

AGSRuntime *AGSRuntime::get_singleton() {
	return singleton;
}

void AGSRuntime::_bind_methods() {
	ClassDB::bind_method(D_METHOD("get_character", "name"), &AGSRuntime::get_character);
	ClassDB::bind_method(D_METHOD("get_room", "name"), &AGSRuntime::get_room);
}

void AGSRuntime::register_character(AGSCharacter *p_character) {
	ERR_FAIL_NULL(p_character);
	characters[p_character->get_character_name()] = p_character;
}

void AGSRuntime::unregister_character(AGSCharacter *p_character) {
	ERR_FAIL_NULL(p_character);
	characters.erase(p_character->get_character_name());
}

AGSCharacter *AGSRuntime::get_character(const String &p_name) const {
	const AGSCharacter *const *found = characters.getptr(StringName(p_name));
	if (!found) {
		return nullptr;
	}
	return const_cast<AGSCharacter *>(*found);
}

void AGSRuntime::register_room(AGSRoom *p_room) {
	ERR_FAIL_NULL(p_room);
	rooms[p_room->get_room_name()] = p_room;
}

void AGSRuntime::unregister_room(AGSRoom *p_room) {
	ERR_FAIL_NULL(p_room);
	rooms.erase(p_room->get_room_name());
}

AGSRoom *AGSRuntime::get_room(const String &p_name) const {
	const AGSRoom *const *found = rooms.getptr(StringName(p_name));
	if (!found) {
		return nullptr;
	}
	return const_cast<AGSRoom *>(*found);
}
