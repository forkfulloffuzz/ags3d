#include "ags_script_language.h"

#include "ags_runtime.h"
#include "ags_script.h"
#include "core/os/os.h"

AGSScriptLanguage *AGSScriptLanguage::singleton = nullptr;

AGSScriptLanguage::AGSScriptLanguage() {
	singleton = this;
}

AGSScriptLanguage *AGSScriptLanguage::get_singleton() {
	return singleton;
}

void AGSScriptLanguage::init() {
	_error_handler.errfunc = &AGSScriptLanguage::_on_script_error;
	_error_handler.userdata = nullptr;
	add_error_handler(&_error_handler);
}

void AGSScriptLanguage::finish() {
	remove_error_handler(&_error_handler);
}

// T34 — Intercept GDScript VM errors for generated .gd files and translate
// the reported location back to the originating AGS-spirit source line.
void AGSScriptLanguage::_on_script_error(void *p_ud, const char *p_func, const char *p_file,
		int p_line, const char *p_error, const char *p_errorexp,
		bool p_editor_notify, ErrorHandlerType p_type) {
	if (p_type != ERR_HANDLER_SCRIPT) {
		return;
	}
	AGSRuntime *rt = AGSRuntime::get_singleton();
	if (!rt) {
		return;
	}
	String gd_path = String::utf8(p_file);
	// Only translate errors from our generated files.
	if (!gd_path.contains(".engine/generated/") || !gd_path.ends_with(".gd")) {
		return;
	}
	Dictionary loc = rt->translate_script_error(gd_path, p_line);
	if (loc.is_empty()) {
		return;
	}
	String agscript_file = loc["file"];
	int agscript_line = int(loc["line"]);
	OS::get_singleton()->print("AGSScript error: %s:%d: %s\n",
			agscript_file.utf8().get_data(), agscript_line, p_error);
}

Ref<Script> AGSScriptLanguage::make_template(const String &p_template, const String &p_class_name, const String &p_base_class_name) const {
	Ref<AGSScript> script;
	script.instantiate();
	return script;
}