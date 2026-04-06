#pragma once

#include "core/object/object.h"
#include "core/string/string_name.h"
#include "core/templates/hash_map.h"
#include "core/templates/vector.h"
#include "core/variant/callable.h"
#include "core/variant/dictionary.h"

// AGSEventBus is a synchronous publish/subscribe event bus for the AGS3D
// cutscene system. It is registered as an Engine singleton ("AGSEventBus").
//
// All subscribers fire synchronously before emit() returns, in subscription
// order. Event names are StringNames; the recommended convention is
// "event:{character_name}:{tag_name}" (e.g. "event:player:land").
//
// GDScript usage:
//
//   AGSEventBus.subscribe("event:player:land", _on_player_land)
//   AGSEventBus.emit("event:player:land", {})
//   AGSEventBus.unsubscribe("event:player:land", _on_player_land)
class AGSEventBus : public Object {
	GDCLASS(AGSEventBus, Object);

	static AGSEventBus *singleton;

	// Subscribers per event name. Each entry is a list of callables; order is
	// preserved (FIFO). Callables are compared by identity for unsubscribe.
	HashMap<StringName, Vector<Callable>> subscribers;

protected:
	static void _bind_methods();

public:
	static AGSEventBus *get_singleton();

	AGSEventBus();
	~AGSEventBus();

	// emit fires all subscribers for p_name synchronously, passing p_payload
	// as the single Dictionary argument. Subscribers added during emit are not
	// called for the current dispatch (snapshot semantics).
	void emit_event(const StringName &p_name, const Dictionary &p_payload);

	// subscribe registers p_callable to receive events with p_name.
	// Adding the same callable twice results in it being called twice per emit.
	void subscribe(const StringName &p_name, const Callable &p_callable);

	// unsubscribe removes the first matching p_callable for p_name.
	// Does nothing if the callable is not registered.
	void unsubscribe(const StringName &p_name, const Callable &p_callable);

	// subscriber_count returns the number of subscribers for p_name.
	// Useful for testing and debugging.
	int subscriber_count(const StringName &p_name) const;

	// clear removes all subscribers for all events.
	void clear();
};
