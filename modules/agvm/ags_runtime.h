#pragma once

#include "core/object/object.h"
#include "core/templates/hash_map.h"

class AGSCharacter;
class AGSRoom;

// AGSRuntime is the central singleton that tracks all live AGSRoom and
// AGSCharacter nodes by name. Scripts query it to resolve names to nodes.
// Registered as an Engine singleton so GDScript can access it as AGSRuntime.
class AGSRuntime : public Object {
	GDCLASS(AGSRuntime, Object);

	static AGSRuntime *singleton;

	HashMap<StringName, AGSCharacter *> characters;
	HashMap<StringName, AGSRoom *> rooms;

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

	AGSRuntime();
	~AGSRuntime();
};
