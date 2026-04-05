#include "ags_runtime.h"

#include "ags_camera.h"
#include "ags_character_base.h"
#include "ags_item.h"
#include "ags_room.h"
#include "ags_room_item.h"
#include "ags_trace.h"
#include "core/io/file_access.h"
#include "core/io/json.h"
#include "core/object/class_db.h"

AGSRuntime *AGSRuntime::singleton = nullptr;
bool AGSRuntime::_trace_enabled = true;

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

bool AGSRuntime::is_trace_enabled() {
	return _trace_enabled;
}

void AGSRuntime::set_trace_enabled(bool p_enabled) {
	_trace_enabled = p_enabled;
}

bool AGSRuntime::get_trace_enabled() const {
	return _trace_enabled;
}

void AGSRuntime::_bind_methods() {
	ClassDB::bind_method(D_METHOD("get_camera", "name"), &AGSRuntime::get_camera);
	ClassDB::bind_method(D_METHOD("set_camera", "name"), &AGSRuntime::set_camera);
	ClassDB::bind_method(D_METHOD("get_character", "name"), &AGSRuntime::get_character);
	ClassDB::bind_method(D_METHOD("get_room", "name"), &AGSRuntime::get_room);
	ClassDB::bind_method(D_METHOD("get_point", "room_name", "point_name"), &AGSRuntime::get_point);
	ClassDB::bind_method(D_METHOD("register_source_map", "gd_path", "agmap"), &AGSRuntime::register_source_map);
	ClassDB::bind_method(D_METHOD("translate_script_error", "gd_path", "gd_line"), &AGSRuntime::translate_script_error);

	ClassDB::bind_method(D_METHOD("load_room", "room_name"), &AGSRuntime::load_room);

	ClassDB::bind_method(D_METHOD("get_item", "name"), &AGSRuntime::get_item);

	ClassDB::bind_method(D_METHOD("get_global", "name"), &AGSRuntime::get_global);
	ClassDB::bind_method(D_METHOD("set_global", "name", "value"), &AGSRuntime::set_global);
	ClassDB::bind_method(D_METHOD("init_globals", "defaults"), &AGSRuntime::init_globals);

	ClassDB::bind_method(D_METHOD("hide_room_item", "name"), &AGSRuntime::hide_room_item);
	ClassDB::bind_method(D_METHOD("show_room_item", "name"), &AGSRuntime::show_room_item);

	ClassDB::bind_method(D_METHOD("set_player_control", "enabled"), &AGSRuntime::set_player_control);
	ClassDB::bind_method(D_METHOD("is_player_control_enabled"), &AGSRuntime::is_player_control_enabled);
	ADD_PROPERTY(PropertyInfo(Variant::BOOL, "player_control_enabled"), "set_player_control", "is_player_control_enabled");

	ClassDB::bind_method(D_METHOD("set_trace_enabled", "enabled"), &AGSRuntime::set_trace_enabled);
	ClassDB::bind_method(D_METHOD("get_trace_enabled"), &AGSRuntime::get_trace_enabled);
	ADD_PROPERTY(PropertyInfo(Variant::BOOL, "trace_enabled"), "set_trace_enabled", "get_trace_enabled");

	ClassDB::bind_method(D_METHOD("save_game", "slot"), &AGSRuntime::save_game);
	ClassDB::bind_method(D_METHOD("load_game", "slot"), &AGSRuntime::load_game);
	ClassDB::bind_method(D_METHOD("game_saved", "slot"), &AGSRuntime::game_saved);
	ClassDB::bind_method(D_METHOD("get_current_room"), &AGSRuntime::get_current_room);
	ClassDB::bind_method(D_METHOD("get_current_music"), &AGSRuntime::get_current_music);

	ClassDB::bind_method(D_METHOD("play_music", "name"), &AGSRuntime::play_music);
	ClassDB::bind_method(D_METHOD("stop_music"), &AGSRuntime::stop_music);
	ClassDB::bind_method(D_METHOD("play_sound", "name"), &AGSRuntime::play_sound);

	ADD_SIGNAL(MethodInfo("room_change_requested", PropertyInfo(Variant::STRING, "room_name")));
	ADD_SIGNAL(MethodInfo("player_control_changed", PropertyInfo(Variant::BOOL, "enabled")));
	ADD_SIGNAL(MethodInfo("play_music_requested", PropertyInfo(Variant::STRING, "name")));
	ADD_SIGNAL(MethodInfo("stop_music_requested"));
	ADD_SIGNAL(MethodInfo("play_sound_requested", PropertyInfo(Variant::STRING, "name")));
	ADD_SIGNAL(MethodInfo("load_game_requested", PropertyInfo(Variant::DICTIONARY, "data")));
}

// --------------------------------------------------------------------------
// Global variable store
// --------------------------------------------------------------------------

Variant AGSRuntime::get_global(const String &p_name) const {
	AGS_TRACE("AGSRuntime", "get_global", vformat("name=%s", p_name))
	if (_globals.has(p_name)) {
		return _globals[p_name];
	}
	WARN_PRINT(vformat("AGSRuntime.get_global: unknown global '%s'", p_name));
	return Variant();
}

void AGSRuntime::set_global(const String &p_name, const Variant &p_value) {
	AGS_TRACE("AGSRuntime", "set_global", vformat("name=%s value=%s", p_name, p_value))
	_globals[p_name] = p_value;
}

void AGSRuntime::init_globals(const Dictionary &p_defaults) {
	// Merge defaults into _globals without overwriting values already set
	// (allows save/load to restore values before init is called again).
	Array keys = p_defaults.keys();
	for (int i = 0; i < keys.size(); i++) {
		String key = keys[i];
		if (!_globals.has(key)) {
			_globals[key] = p_defaults[key];
		}
	}
}

// --------------------------------------------------------------------------
// Item registry
// --------------------------------------------------------------------------

void AGSRuntime::register_item(AGSItem *p_item) {
	ERR_FAIL_NULL(p_item);
	AGS_TRACE("AGSRuntime", "register_item", vformat("name=%s", p_item->get_item_name()))
	items[p_item->get_item_name()] = p_item;
}

void AGSRuntime::unregister_item(AGSItem *p_item) {
	ERR_FAIL_NULL(p_item);
	items.erase(p_item->get_item_name());
}

AGSItem *AGSRuntime::get_item(const String &p_name) const {
	AGS_TRACE("AGSRuntime", "get_item", vformat("name=%s", p_name))
	const AGSItem *const *found = items.getptr(StringName(p_name));
	if (!found) {
		return nullptr;
	}
	return const_cast<AGSItem *>(*found);
}

// --------------------------------------------------------------------------

void AGSRuntime::register_camera(AGSCamera *p_camera) {
	ERR_FAIL_NULL(p_camera);
	cameras[p_camera->get_camera_name()] = p_camera;
}

void AGSRuntime::unregister_camera(AGSCamera *p_camera) {
	ERR_FAIL_NULL(p_camera);
	cameras.erase(p_camera->get_camera_name());
}

AGSCamera *AGSRuntime::get_camera(const String &p_name) const {
	const AGSCamera *const *found = cameras.getptr(StringName(p_name));
	if (!found) {
		return nullptr;
	}
	return const_cast<AGSCamera *>(*found);
}

void AGSRuntime::set_camera(const String &p_name) {
	AGSCamera *cam = get_camera(p_name);
	if (!cam) {
		WARN_PRINT(vformat("AGSRuntime::set_camera: camera '%s' not found.", p_name));
		return;
	}
	AGS_TRACE("AGSRuntime", "set_camera", vformat("activating '%s'", p_name))
	cam->make_current();
}

void AGSRuntime::register_character(AGSCharacterBase *p_character) {
	ERR_FAIL_NULL(p_character);
	characters[p_character->get_character_name()] = p_character;
}

void AGSRuntime::unregister_character(AGSCharacterBase *p_character) {
	ERR_FAIL_NULL(p_character);
	characters.erase(p_character->get_character_name());
}

AGSCharacterBase *AGSRuntime::get_character(const String &p_name) const {
	const AGSCharacterBase *const *found = characters.getptr(StringName(p_name));
	AGS_TRACE("AGSRuntime", "get_character", vformat("name=%s → %s", p_name, found ? "found" : "null"))
	if (!found) {
		return nullptr;
	}
	return const_cast<AGSCharacterBase *>(*found);
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

// --------------------------------------------------------------------------
// RoomItem global registry
// --------------------------------------------------------------------------

void AGSRuntime::register_room_item(AGSRoomItem *p_item) {
	ERR_FAIL_NULL(p_item);
	AGS_TRACE("AGSRuntime", "register_room_item", vformat("name=%s", p_item->get_item_name()))
	room_items[p_item->get_item_name()] = p_item;
}

void AGSRuntime::unregister_room_item(AGSRoomItem *p_item) {
	ERR_FAIL_NULL(p_item);
	room_items.erase(p_item->get_item_name());
}

void AGSRuntime::hide_room_item(const String &p_name) {
	AGS_TRACE("AGSRuntime", "hide_room_item", vformat("name=%s", p_name))
	const AGSRoomItem *const *found = room_items.getptr(StringName(p_name));
	if (!found) {
		WARN_PRINT(vformat("AGSRuntime.hide_room_item: room item '%s' not found.", p_name));
		return;
	}
	const_cast<AGSRoomItem *>(*found)->set_visible(false);
}

void AGSRuntime::show_room_item(const String &p_name) {
	AGS_TRACE("AGSRuntime", "show_room_item", vformat("name=%s", p_name))
	const AGSRoomItem *const *found = room_items.getptr(StringName(p_name));
	if (!found) {
		WARN_PRINT(vformat("AGSRuntime.show_room_item: room item '%s' not found.", p_name));
		return;
	}
	const_cast<AGSRoomItem *>(*found)->set_visible(true);
}

// --------------------------------------------------------------------------
// Player control
// --------------------------------------------------------------------------

void AGSRuntime::set_player_control(bool p_enabled) {
	AGS_TRACE("AGSRuntime", "set_player_control", vformat("enabled=%s", p_enabled ? "true" : "false"))
	_player_control_enabled = p_enabled;
	emit_signal("player_control_changed", p_enabled);
}

bool AGSRuntime::is_player_control_enabled() const {
	return _player_control_enabled;
}

// --------------------------------------------------------------------------

void AGSRuntime::load_room(const String &p_room_name) {
	AGS_TRACE("AGSRuntime", "load_room", vformat("room_name=%s", p_room_name))
	_current_room = p_room_name;
	emit_signal("room_change_requested", p_room_name);
}

// --------------------------------------------------------------------------
// Audio — emit signals; AGSAudio AutoLoad handles actual playback
// --------------------------------------------------------------------------

void AGSRuntime::play_music(const String &p_name) {
	AGS_TRACE("AGSRuntime", "play_music", vformat("name=%s", p_name))
	_current_music = p_name;
	emit_signal("play_music_requested", p_name);
}

void AGSRuntime::stop_music() {
	AGS_TRACE("AGSRuntime", "stop_music", "")
	_current_music = "";
	emit_signal("stop_music_requested");
}

void AGSRuntime::play_sound(const String &p_name) {
	AGS_TRACE("AGSRuntime", "play_sound", vformat("name=%s", p_name))
	emit_signal("play_sound_requested", p_name);
}

// --------------------------------------------------------------------------
// Save / Load
// --------------------------------------------------------------------------

static String _save_path(int p_slot) {
	return vformat("user://save_%d.json", p_slot);
}

void AGSRuntime::save_game(int p_slot) {
	AGS_TRACE("AGSRuntime", "save_game", vformat("slot=%d", p_slot))

	Dictionary data;
	data["room"] = _current_room;
	data["music"] = _current_music;
	data["globals"] = _globals;

	// Character inventories — call GDScript get_inventory() on each character.
	Dictionary char_data;
	for (const KeyValue<StringName, AGSCharacterBase *> &kv : characters) {
		Variant inv = kv.value->call("get_inventory");
		char_data[String(kv.key)] = inv;
	}
	data["characters"] = char_data;

	// Room item visibility.
	Dictionary item_vis;
	for (const KeyValue<StringName, AGSRoomItem *> &kv : room_items) {
		item_vis[String(kv.key)] = kv.value->is_visible();
	}
	data["room_items"] = item_vis;

	String json = JSON::stringify(data, "\t");
	Error err;
	Ref<FileAccess> fa = FileAccess::open(_save_path(p_slot), FileAccess::WRITE, &err);
	if (err != OK || fa.is_null()) {
		WARN_PRINT(vformat("AGSRuntime.save_game: could not write '%s'", _save_path(p_slot)));
		return;
	}
	fa->store_string(json);
}

void AGSRuntime::load_game(int p_slot) {
	AGS_TRACE("AGSRuntime", "load_game", vformat("slot=%d", p_slot))
	String path = _save_path(p_slot);
	Error err;
	Ref<FileAccess> fa = FileAccess::open(path, FileAccess::READ, &err);
	if (err != OK || fa.is_null()) {
		WARN_PRINT(vformat("AGSRuntime.load_game: save file '%s' not found.", path));
		return;
	}
	String json_text = fa->get_as_text();
	Variant parsed = JSON::parse_string(json_text);
	if (parsed.get_type() != Variant::DICTIONARY) {
		WARN_PRINT(vformat("AGSRuntime.load_game: could not parse '%s'.", path));
		return;
	}

	Dictionary data = parsed;

	// Restore globals immediately.
	if (data.has("globals")) {
		Dictionary saved_globals = data["globals"];
		Array keys = saved_globals.keys();
		for (int i = 0; i < keys.size(); i++) {
			_globals[keys[i]] = saved_globals[keys[i]];
		}
	}

	// Track restored room and music.
	if (data.has("room")) {
		_current_room = String(data["room"]);
	}
	if (data.has("music")) {
		_current_music = String(data["music"]);
	}

	// Restore character inventories immediately (characters may be in memory).
	if (data.has("characters")) {
		Dictionary char_data = data["characters"];
		Array char_keys = char_data.keys();
		for (int i = 0; i < char_keys.size(); i++) {
			String char_name = char_keys[i];
			const AGSCharacterBase *const *found = characters.getptr(StringName(char_name));
			if (found) {
				(*const_cast<AGSCharacterBase **>(found))->call("set_inventory", char_data[char_name]);
			}
		}
	}

	// Restore room item visibility immediately (items may be in memory).
	if (data.has("room_items")) {
		Dictionary item_vis = data["room_items"];
		Array item_keys = item_vis.keys();
		for (int i = 0; i < item_keys.size(); i++) {
			String item_name = item_keys[i];
			const AGSRoomItem *const *found = room_items.getptr(StringName(item_name));
			if (found) {
				const_cast<AGSRoomItem *>(*found)->set_visible(bool(item_vis[item_name]));
			}
		}
	}

	// Emit full data for the game scene manager to handle deferred restoration
	// (e.g. loading the saved room, replaying music, applying state after room loads).
	emit_signal("load_game_requested", data);
}

bool AGSRuntime::game_saved(int p_slot) const {
	return FileAccess::exists(_save_path(p_slot));
}

Vector3 AGSRuntime::get_point(const String &p_room_name, const String &p_point_name) const {
	AGSRoom *room = get_room(p_room_name);
	ERR_FAIL_NULL_V_MSG(room, Vector3(), vformat("AGSRuntime: room '%s' not found.", p_room_name));
	Vector3 pos = room->get_point(p_point_name);
	AGS_TRACE("AGSRuntime", "get_point", vformat("room=%s, point=%s → %s", p_room_name, p_point_name, pos))
	return pos;
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
