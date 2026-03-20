#include "ags_script_language.h"
#include "ags_script.h"

AGSScriptLanguage *AGSScriptLanguage::singleton = nullptr;

AGSScriptLanguage::AGSScriptLanguage() {
	singleton = this;
}

AGSScriptLanguage *AGSScriptLanguage::get_singleton() {
	return singleton;
}

Ref<Script> AGSScriptLanguage::make_template(const String &p_template, const String &p_class_name, const String &p_base_class_name) const {
	Ref<AGSScript> script;
	script.instantiate();
	return script;
}