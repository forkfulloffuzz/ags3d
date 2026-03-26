#pragma once

#include "core/object/object.h"
#include "core/templates/hash_map.h"
#include "core/templates/vector.h"

class AGSCharacter;
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

	HashMap<StringName, AGSCharacter *> characters;
	HashMap<StringName, AGSRoom *> rooms;

	// Source maps keyed by the res:// path of the generated .gd file.
	HashMap<String, Vector<SourceMapEntry>> _source_maps;

protected:
	static void _bind_methods();

public:
	static AGSRuntime *get_singleton();

	// Character registry — called by AGSCharacter on NOTIFICATION_READY / EXIT_TREE.
	void register_character(AGSCharacter *p_character);
	void unregister_character(AGSCharacter *p_character);
	AGSCharacter *get_character(const String &p_name) const;

	// Room registry — called by AGSRoom on NOTIFICATION_READY / EXIT_TREE.
	void register_room(AGSRoom *p_room);
	void unregister_room(AGSRoom *p_room);
	AGSRoom *get_room(const String &p_name) const;

	// Cross-room point lookup — delegates to AGSRoom::get_point().
	Vector3 get_point(const String &p_room_name, const String &p_point_name) const;

	// Source map registry — called by AGSScript loader after transpilation.
	// p_agmap is the parsed JSON: [[gd_line, "rel/path.agscript", agscript_line], ...].
	void register_source_map(const String &p_gd_path, const Array &p_agmap);

	// Translate a GDScript error location to its AGS-spirit source location.
	// Returns {"file": String, "line": int} or an empty Dictionary if not found.
	Dictionary translate_script_error(const String &p_gd_path, int p_gd_line) const;

	AGSRuntime();
	~AGSRuntime();
};
