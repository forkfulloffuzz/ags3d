#include "ags_event_bus.h"

#include "core/config/engine.h"
#include "core/object/class_db.h"

AGSEventBus *AGSEventBus::singleton = nullptr;

AGSEventBus::AGSEventBus() {
	singleton = this;
}

AGSEventBus::~AGSEventBus() {
	if (singleton == this) {
		singleton = nullptr;
	}
}

AGSEventBus *AGSEventBus::get_singleton() {
	return singleton;
}

void AGSEventBus::emit_event(const StringName &p_name, const Dictionary &p_payload) {
	if (!subscribers.has(p_name)) {
		return;
	}
	// Snapshot the list so that subscribe/unsubscribe during dispatch does not
	// affect the current iteration.
	Vector<Callable> snap = subscribers[p_name];
	for (int i = 0; i < snap.size(); i++) {
		const Callable &c = snap[i];
		if (c.is_valid()) {
			c.call(p_payload);
		}
	}
}

void AGSEventBus::subscribe(const StringName &p_name, const Callable &p_callable) {
	subscribers[p_name].push_back(p_callable);
}

void AGSEventBus::unsubscribe(const StringName &p_name, const Callable &p_callable) {
	if (!subscribers.has(p_name)) {
		return;
	}
	Vector<Callable> &list = subscribers[p_name];
	for (int i = 0; i < list.size(); i++) {
		if (list[i] == p_callable) {
			list.remove_at(i);
			return;
		}
	}
}

int AGSEventBus::subscriber_count(const StringName &p_name) const {
	if (!subscribers.has(p_name)) {
		return 0;
	}
	return subscribers[p_name].size();
}

void AGSEventBus::clear() {
	subscribers.clear();
}

void AGSEventBus::_bind_methods() {
	ClassDB::bind_method(D_METHOD("emit_event", "name", "payload"), &AGSEventBus::emit_event);
	ClassDB::bind_method(D_METHOD("subscribe", "name", "callable"), &AGSEventBus::subscribe);
	ClassDB::bind_method(D_METHOD("unsubscribe", "name", "callable"), &AGSEventBus::unsubscribe);
	ClassDB::bind_method(D_METHOD("subscriber_count", "name"), &AGSEventBus::subscriber_count);
	ClassDB::bind_method(D_METHOD("clear"), &AGSEventBus::clear);
}
