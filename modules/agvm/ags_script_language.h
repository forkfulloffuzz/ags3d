#pragma once

#include "core/error/error_macros.h"
#include "core/object/script_language.h"

class AGSScriptLanguage : public ScriptLanguage {
	ErrorHandlerList _error_handler;

	static void _on_script_error(void *p_ud, const char *p_func, const char *p_file,
			int p_line, const char *p_error, const char *p_errorexp,
			bool p_editor_notify, ErrorHandlerType p_type);

public:
	static AGSScriptLanguage *singleton;

	static AGSScriptLanguage *get_singleton();

	AGSScriptLanguage();

	virtual String get_name() const override { return "AGSScript"; }
	virtual String get_type() const override { return "AGSScript"; }
	virtual String get_extension() const override { return "agscript"; }

	virtual void init() override;
	virtual void finish() override;

	virtual Vector<String> get_reserved_words() const override { return Vector<String>(); }
	virtual bool is_control_flow_keyword(const String &p_string) const override { return false; }
	virtual Vector<String> get_comment_delimiters() const override { return Vector<String>(); }
	virtual Vector<String> get_doc_comment_delimiters() const override { return Vector<String>(); }
	virtual Vector<String> get_string_delimiters() const override { return Vector<String>(); }

	virtual bool validate(const String &p_script, const String &p_path = "",
			List<String> *r_functions = nullptr,
			List<ScriptLanguage::ScriptError> *r_errors = nullptr,
			List<ScriptLanguage::Warning> *r_warnings = nullptr,
			HashSet<int> *r_safe_lines = nullptr) const override { return true; }

	virtual Ref<Script> make_template(const String &p_template, const String &p_class_name, const String &p_base_class_name) const override;

	virtual bool supports_builtin_mode() const override { return false; }
	virtual int find_function(const String &p_function, const String &p_code) const override { return -1; }
	virtual String make_function(const String &p_class, const String &p_name,
			const PackedStringArray &p_args) const override { return ""; }
	virtual void auto_indent_code(String &p_code, int p_from_line, int p_to_line) const override {}
	virtual void add_global_constant(const StringName &p_variable, const Variant &p_value) override {}

	virtual String debug_get_error() const override { return ""; }
	virtual int debug_get_stack_level_count() const override { return 0; }
	virtual int debug_get_stack_level_line(int p_level) const override { return 0; }
	virtual String debug_get_stack_level_function(int p_level) const override { return ""; }
	virtual String debug_get_stack_level_source(int p_level) const override { return ""; }
	virtual void debug_get_stack_level_locals(int p_level, List<String> *p_locals, List<Variant> *p_values, int p_max_subitems = -1, int p_max_depth = -1) override {}
	virtual void debug_get_stack_level_members(int p_level, List<String> *p_members, List<Variant> *p_values, int p_max_subitems = -1, int p_max_depth = -1) override {}
	virtual void debug_get_globals(List<String> *p_globals, List<Variant> *p_values, int p_max_subitems = -1, int p_max_depth = -1) override {}
	virtual String debug_parse_stack_level_expression(int p_level, const String &p_expression, int p_max_subitems = -1, int p_max_depth = -1) override { return ""; }

	virtual void reload_all_scripts() override {}
	virtual void reload_scripts(const Array &p_scripts, bool p_soft_reload) override {}
	virtual void reload_tool_script(const Ref<Script> &p_script, bool p_soft_reload) override {}

	virtual void get_recognized_extensions(List<String> *p_extensions) const override {
		p_extensions->push_back("agscript");
	}
	virtual void get_public_functions(List<MethodInfo> *p_functions) const override {}
	virtual void get_public_constants(List<Pair<String, Variant>> *p_constants) const override {}
	virtual void get_public_annotations(List<MethodInfo> *p_annotations) const override {}

	virtual void profiling_start() override {}
	virtual void profiling_stop() override {}
	virtual void profiling_set_save_native_calls(bool p_enable) override {}
	virtual int profiling_get_accumulated_data(ProfilingInfo *p_info_arr, int p_info_max) override { return 0; }
	virtual int profiling_get_frame_data(ProfilingInfo *p_info_arr, int p_info_max) override { return 0; }
};
