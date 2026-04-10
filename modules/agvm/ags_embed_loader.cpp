// ResourceFormatLoaderAGSEmbed — serves embedded GDScripts from C++ string constants.
//
// This replaces loading .gd files from disk with loading from embedded strings
// in the agvm C++ module. Used by T-FINAL to embed .engine/runtime/*.gd into
// the Godot binary so they don't need to be distributed separately.
//
// To add a new embedded script: add it to embedded_scripts.go and run the
// generator to produce embedded_scripts_strings.h.
#include "ags_embed_loader.h"

#include "core/io/resource_loader.h"
#include "core/io/resource_saver.h"
#include "modules/gdscript/gdscript.h"

bool ResourceFormatLoaderAGSEmbed::handles_type(const String &p_type) const {
	return p_type == "GDScript";
}

bool ResourceFormatLoaderAGSEmbed::recognize_path(const String &p_path, const String &p_for_type) const {
	if (p_for_type == "GDScript" || p_for_type == "Script") {
		return _get_embedded_source(p_path).length() > 0;
	}
	return false;
}

String ResourceFormatLoaderAGSEmbed::_get_embedded_source(const String &p_path) const {
	// Try each registered embedded script.
	for (const EmbeddedScript &es : _scripts) {
		if (p_path == es.path) {
			return es.source;
		}
	}
	return String();
}

Ref<Resource> ResourceFormatLoaderAGSEmbed::load(const String &p_path, const String &p_for_type,
		bool p_no_cache, bool p_use_sub_threads, float *r_progress, CacheMode p_cache_mode) {
	String src = _get_embedded_source(p_path);
	if (src.is_empty()) {
		return Ref<Resource>();
	}

	Ref<GDScript> script;
	script.instantiate();
	script->set_source_code(src);

	// Store the path so Godot knows how to identify this script.
	script->set_path(p_path);

	return script;
}

void ResourceFormatLoaderAGSEmbed::get_recognized_extensions(List<String> *p_extensions) const {
	for (const EmbeddedScript &es : _scripts) {
		String ext = es.path.get_extension();
		if (!ext.is_empty()) {
			p_extensions->push_back(ext);
		}
	}
}

String ResourceFormatLoaderAGSEmbed::get_resource_type(const String &p_path) const {
	if (!_get_embedded_source(p_path).is_empty()) {
		return "GDScript";
	}
	return "";
}

void ResourceFormatLoaderAGSEmbed::get_dependencies(const String &p_path,
		HashMap<String, HashSet<String>> &p_dependencies, int p_depth) {
	// No dependencies — scripts are self-contained.
}

void ResourceFormatLoaderAGSEmbed::register_script(const String &p_path, const String &p_source) {
	EmbeddedScript es;
	es.path = p_path;
	es.source = p_source;
	_scripts.push_back(es);
}

Vector<String> ResourceFormatLoaderAGSEmbed::get_embedded_paths() const {
	Vector<String> paths;
	for (const EmbeddedScript &es : _scripts) {
		paths.push_back(es.path);
	}
	return paths;
}

ResourceFormatLoaderAGSEmbed::ResourceFormatLoaderAGSEmbed() {}
ResourceFormatLoaderAGSEmbed::~ResourceFormatLoaderAGSEmbed() {}
