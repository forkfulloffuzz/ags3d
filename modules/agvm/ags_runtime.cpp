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
	ClassDB::bind_method(D_METHOD("get_point", "room_name", "point_name"), &AGSRuntime::get_point);
	ClassDB::bind_method(D_METHOD("register_source_map", "gd_path", "agmap"), &AGSRuntime::register_source_map);
	ClassDB::bind_method(D_METHOD("translate_script_error", "gd_path", "gd_line"), &AGSRuntime::translate_script_error);
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

Vector3 AGSRuntime::get_point(const String &p_room_name, const String &p_point_name) const {
	AGSRoom *room = get_room(p_room_name);
	ERR_FAIL_NULL_V_MSG(room, Vector3(), vformat("AGSRuntime: room '%s' not found.", p_room_name));
	return room->get_point(p_point_name);
}

void AGSRuntime::register_source_map(const String &p_gd_path, const Array &p_agmap) {
	Vector<SourceMapEntry> entries;
	for (int i = 0; i < p_agmap.size(); i++) {
		Array row = p_agmap[i];
		if (row.size() < 3) {
			continue;
		}
		SourceMapEntry e;
		e.gd_line = int(row[0]);
		e.agscript_file = String(row[1]);
		e.agscript_line = int(row[2]);
		entries.push_back(e);
	}
	_source_maps[p_gd_path] = entries;
}

Dictionary AGSRuntime::translate_script_error(const String &p_gd_path, int p_gd_line) const {
	const Vector<SourceMapEntry> *entries = _source_maps.getptr(p_gd_path);
	if (!entries) {
		return Dictionary();
	}

	// Walk the sorted entries and keep the last one whose gd_line <= p_gd_line.
	const SourceMapEntry *best = nullptr;
	for (int i = 0; i < entries->size(); i++) {
		const SourceMapEntry &e = (*entries)[i];
		if (e.gd_line <= p_gd_line) {
			best = &e;
		} else {
			break;
		}
	}
	if (!best) {
		return Dictionary();
	}

	Dictionary d;
	d["file"] = best->agscript_file;
	d["line"] = best->agscript_line;
	return d;
}
