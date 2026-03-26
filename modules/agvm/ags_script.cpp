#include "ags_script.h"

#include "ags_script_language.h"

#include "core/config/project_settings.h"
#include "core/io/file_access.h"
#include "core/io/resource_loader.h"
#include "core/os/os.h"

ScriptLanguage *AGSScript::get_language() const {
	return AGSScriptLanguage::get_singleton();
}

void AGSScript::set_inner_script(const Ref<Script> &p_script) {
	_inner_script = p_script;
}

bool AGSScript::can_instantiate() const {
	return _inner_script.is_valid() && _inner_script->can_instantiate();
}

StringName AGSScript::get_instance_base_type() const {
	if (_inner_script.is_valid()) {
		return _inner_script->get_instance_base_type();
	}
	return StringName();
}

ScriptInstance *AGSScript::instance_create(Object *p_this) {
	if (_inner_script.is_valid()) {
		return _inner_script->instance_create(p_this);
	}
	return nullptr;
}

bool AGSScript::instance_has(const Object *p_this) const {
	if (_inner_script.is_valid()) {
		return _inner_script->instance_has(p_this);
	}
	return false;
}

bool AGSScript::is_valid() const {
	if (_inner_script.is_valid()) {
		return _inner_script->is_valid();
	}
	return true;
}

// ---- ResourceFormatLoaderAGSScript ----

Ref<Resource> ResourceFormatLoaderAGSScript::load(const String &p_path, const String &p_original_path,
		Error *r_error, bool p_use_sub_threads, float *r_progress, CacheMode p_cache_mode) {
	if (r_error) {
		*r_error = ERR_FILE_CANT_OPEN;
	}

	Ref<AGSScript> script;
	script.instantiate();

	// Load .agscript source text for editor display.
	Error err;
	Ref<FileAccess> f = FileAccess::open(p_path, FileAccess::READ, &err);
	if (f.is_valid()) {
		script->set_source_code(f->get_as_text());
	}

	// Transpile and wire to generated GDScript.
	// Only runs when a game.agp file is present (skipped in test/non-project environments).
	if (ProjectSettings::get_singleton() != nullptr) {
		String project_root = ProjectSettings::get_singleton()->globalize_path("res://");
		if (!project_root.ends_with("/")) {
			project_root += "/";
		}
		if (FileAccess::exists(project_root + "game.agp")) {
			String global_src = ProjectSettings::get_singleton()->globalize_path(p_path);
			String rel = global_src.substr(project_root.length());
			String gd_global = project_root + ".engine/generated/" + rel + ".gd";
			String gd_res = "res://.engine/generated/" + rel + ".gd";

			// Transpile if .gd output is absent or older than the .agscript source.
			uint64_t src_mtime = FileAccess::get_modified_time(global_src);
			uint64_t gd_mtime = FileAccess::exists(gd_global) ? FileAccess::get_modified_time(gd_global) : 0;

			if (src_mtime > gd_mtime) {
				// Run: cd {project_root} && {ag_binary} build
				// ag binary lives alongside the Godot executable.
				String ag = OS::get_singleton()->get_executable_path().get_base_dir().path_join("ag");
				String cmd = "cd " + project_root.trim_suffix("/") + " && " + ag + " build";
				List<String> sh_args;
				sh_args.push_back("-c");
				sh_args.push_back(cmd);
				String output;
				int exit_code = 0;
				OS::get_singleton()->execute("/bin/sh", sh_args, &output, &exit_code, true);
				if (exit_code != 0) {
					OS::get_singleton()->print("AGSScript: transpilation failed for '%s':\n%s\n",
							p_path.utf8().get_data(), output.utf8().get_data());
				}
			}

			// Back this AGSScript with the generated GDScript for runtime execution.
			if (FileAccess::exists(gd_global)) {
				Ref<Resource> inner = ResourceLoader::load(gd_res, "GDScript");
				if (inner.is_valid()) {
					script->set_inner_script(inner);
				}
			}
		}
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
