#pragma once

#include "core/object/object.h"
#include "core/templates/hash_map.h"
#include "core/templates/vector.h"

class AGSCamera;
class AGSCharacter;
class AGSItem;
class AGSRoom;

// SourceMapEntry maps one GDScript line back to the originating AGS-spirit
// source file and line, as recorded in the .agmap sidecar produced by ag build.
struct SourceMapEntry {
	int gd_line = 0;
	String agscript_file;
	int agscript_line = 0;
};

// AGSRuntime is the central singleton that tracks all live AGSRoom and
// AGSCharacter nodes by name. Scripts query it to resolve names to nodes.
// Registered as an Engine singleton so GDScript can access it as AGSRuntime.
class AGSRuntime : public Object {
	GDCLASS(AGSRuntime, Object);

	static AGSRuntime *singleton;

	// Trace flag — true by default so every runtime function logs its calls.
	// Set false in production builds (see GitHub issue for build-flag task).
	static bool _trace_enabled;

	HashMap<StringName, AGSCamera *> cameras;
	HashMap<StringName, AGSCharacter *> characters;
	HashMap<StringName, AGSItem *> items;
	HashMap<StringName, AGSRoom *> rooms;

	// Source maps keyed by the res:// path of the generated .gd file.
	HashMap<String, Vector<SourceMapEntry>> _source_maps;

	// User-defined global variables from game.agp [globals], plus engine-owned
	// globals (player, room, camera) registered at runtime.
	Dictionary _globals;

protected:
	static void _bind_methods();

public:
	static AGSRuntime *get_singleton();

	// Tracing — when enabled every runtime call emits a [AGS/Type::func] log line.
	// is_trace_enabled() is a static so ags_trace.h can call it without a null check.
	static bool is_trace_enabled();
	void set_trace_enabled(bool p_enabled);
	bool get_trace_enabled() const;

	// Camera registry — called by AGSCamera on NOTIFICATION_READY / EXIT_TREE.
	void register_camera(AGSCamera *p_camera);
	void unregister_camera(AGSCamera *p_camera);
	AGSCamera *get_camera(const String &p_name) const;

	// Activate a named camera; deactivates the previously active one.
	void set_camera(const String &p_name);

	// Read-only access to the camera map — used by AGSCameraZone to find the current camera.
	const HashMap<StringName, AGSCamera *> &get_cameras() const { return cameras; }

	// Character registry — called by AGSCharacter on NOTIFICATION_READY / EXIT_TREE.
	void register_character(AGSCharacter *p_character);
	void unregister_character(AGSCharacter *p_character);
	AGSCharacter *get_character(const String &p_name) const;

	// Item registry — called by AGSItem on NOTIFICATION_READY / EXIT_TREE.
	void register_item(AGSItem *p_item);
	void unregister_item(AGSItem *p_item);
	AGSItem *get_item(const String &p_name) const;

	// Room registry — called by AGSRoom on NOTIFICATION_READY / EXIT_TREE.
	void register_room(AGSRoom *p_room);
	void unregister_room(AGSRoom *p_room);
	AGSRoom *get_room(const String &p_name) const;

	// Cross-room point lookup — delegates to AGSRoom::get_point().
	Vector3 get_point(const String &p_room_name, const String &p_point_name) const;

	// Room transition — emits room_change_requested(room_name) for the game
	// scene manager to handle (load the new room scene and free the old one).
	void load_room(const String &p_room_name);

	// Global variable store — user-defined vars from game.agp [globals] and
	// engine-owned globals (player, room, camera).
	// Scripts access these via AGSRuntime.get_global("name") / set_global("name", value).
	Variant get_global(const String &p_name) const;
	void set_global(const String &p_name, const Variant &p_value);

	// Initialise globals from a Dictionary of { name: default_value_string }.
	// Called once at startup by the game's autoload script after loading game.agp.
	void init_globals(const Dictionary &p_defaults);

	// Source map registry — called by AGSScript loader after transpilation.
	// p_agmap is the parsed JSON: [[gd_line, "rel/path.agscript", agscript_line], ...].
	void register_source_map(const String &p_gd_path, const Array &p_agmap);

	// Translate a GDScript error location to its AGS-spirit source location.
	// Returns {"file": String, "line": int} or an empty Dictionary if not found.
	Dictionary translate_script_error(const String &p_gd_path, int p_gd_line) const;

	AGSRuntime();
	~AGSRuntime();
};
