#include "ags_character_base.h"

#include "ags_runtime.h"
#include "core/object/class_db.h"

void AGSCharacterBase::_bind_methods() {
	ClassDB::bind_method(D_METHOD("set_character_name", "name"), &AGSCharacterBase::set_character_name);
	ClassDB::bind_method(D_METHOD("get_character_name"), &AGSCharacterBase::get_character_name);
	ADD_PROPERTY(PropertyInfo(Variant::STRING, "character_name"), "set_character_name", "get_character_name");

	ClassDB::bind_method(D_METHOD("set_move_speed", "speed"), &AGSCharacterBase::set_move_speed);
	ClassDB::bind_method(D_METHOD("get_move_speed"), &AGSCharacterBase::get_move_speed);
	ADD_PROPERTY(PropertyInfo(Variant::FLOAT, "move_speed"), "set_move_speed", "get_move_speed");

	ClassDB::bind_method(D_METHOD("set_say_text", "text"), &AGSCharacterBase::set_say_text);
	ClassDB::bind_method(D_METHOD("get_say_text"), &AGSCharacterBase::get_say_text);
	ADD_PROPERTY(PropertyInfo(Variant::STRING, "say_text"), "set_say_text", "get_say_text");

	ADD_SIGNAL(MethodInfo("walk_completed"));
	ADD_SIGNAL(MethodInfo("face_completed"));
	ADD_SIGNAL(MethodInfo("say_completed"));
}

void AGSCharacterBase::_notification(int p_what) {
	switch (p_what) {
		case NOTIFICATION_READY: {
			if (AGSRuntime::get_singleton()) {
				AGSRuntime::get_singleton()->register_character(this);
			}
		} break;

		case NOTIFICATION_EXIT_TREE: {
			if (AGSRuntime::get_singleton()) {
				AGSRuntime::get_singleton()->unregister_character(this);
			}
		} break;

		case NOTIFICATION_PREDELETE: {
			if (AGSRuntime::get_singleton()) {
				AGSRuntime::get_singleton()->unregister_character(this);
			}
		} break;
	}
}

void AGSCharacterBase::set_character_name(const String &p_name) {
	character_name = p_name;
}

String AGSCharacterBase::get_character_name() const {
	return character_name;
}

void AGSCharacterBase::set_move_speed(float p_speed) {
	move_speed = p_speed;
}

float AGSCharacterBase::get_move_speed() const {
	return move_speed;
}

void AGSCharacterBase::set_say_text(const String &p_text) {
	say_text = p_text;
}

String AGSCharacterBase::get_say_text() const {
	return say_text;
}
