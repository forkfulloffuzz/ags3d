#include "ags_script.h"

#include "ags_script_language.h"

#include "core/io/file_access.h"

ScriptLanguage *AGSScript::get_language() const {
	return AGSScriptLanguage::get_singleton();
}

// ---- ResourceFormatLoaderAGSScript ----

Ref<Resource> ResourceFormatLoaderAGSScript::load(const String &p_path, const String &p_original_path,
		Error *r_error, bool p_use_sub_threads, float *r_progress, CacheMode p_cache_mode) {
	if (r_error) {
		*r_error = ERR_FILE_CANT_OPEN;
	}

	Ref<AGSScript> script;
	script.instantiate();

	Error err;
	Ref<FileAccess> f = FileAccess::open(p_path, FileAccess::READ, &err);
	if (f.is_valid()) {
		String source = f->get_as_text();
		script->set_source_code(source);
	}

	if (r_error) {
		*r_error = OK;
	}
	return script;
}

void ResourceFormatLoaderAGSScript::get_recognized_extensions(List<String> *p_extensions) const {
	p_extensions->push_back("agscript");
}

bool ResourceFormatLoaderAGSScript::handles_type(const String &p_type) const {
	return p_type == "AGSScript" || p_type == "Script";
}

String ResourceFormatLoaderAGSScript::get_resource_type(const String &p_path) const {
	if (p_path.get_extension().to_lower() == "agscript") {
		return "AGSScript";
	}
	return "";
}

// ---- ResourceFormatSaverAGSScript ----

Error ResourceFormatSaverAGSScript::save(const Ref<Resource> &p_resource, const String &p_path, uint32_t p_flags) {
	Ref<AGSScript> script = p_resource;
	ERR_FAIL_COND_V(script.is_null(), ERR_INVALID_PARAMETER);

	Error err;
	Ref<FileAccess> f = FileAccess::open(p_path, FileAccess::WRITE, &err);
	ERR_FAIL_COND_V_MSG(f.is_null(), err, "Cannot save AGSScript file: " + p_path);

	f->store_string(script->get_source_code());
	return OK;
}

void ResourceFormatSaverAGSScript::get_recognized_extensions(const Ref<Resource> &p_resource, List<String> *p_extensions) const {
	if (Object::cast_to<AGSScript>(*p_resource)) {
		p_extensions->push_back("agscript");
	}
}

bool ResourceFormatSaverAGSScript::recognize(const Ref<Resource> &p_resource) const {
	return Object::cast_to<AGSScript>(*p_resource) != nullptr;
}
