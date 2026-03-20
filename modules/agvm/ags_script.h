#pragma once

#include "core/io/resource_loader.h"
#include "core/io/resource_saver.h"
#include "core/object/script_language.h"

class AGSScript : public Script {
	GDCLASS(AGSScript, Script);

protected:
	static void _bind_methods() {}

public:
	// Script pure virtuals
	virtual bool can_instantiate() const override { return false; }
	virtual Ref<Script> get_base_script() const override { return Ref<Script>(); }
	virtual StringName get_global_name() const override { return StringName(); }
	virtual bool inherits_script(const Ref<Script> &p_script) const override { return false; }
	virtual StringName get_instance_base_type() const override { return StringName(); }
	virtual ScriptInstance *instance_create(Object *p_this) override { return nullptr; }
	virtual bool instance_has(const Object *p_this) const override { return false; }

	virtual bool has_source_code() const override { return true; }
	virtual String get_source_code() const override { return source_code; }
	virtual void set_source_code(const String &p_code) override { source_code = p_code; }
	virtual Error reload(bool p_keep_state = false) override { return OK; }

	virtual StringName get_doc_class_name() const override { return StringName(); }
	virtual Vector<DocData::ClassDoc> get_documentation() const override { return Vector<DocData::ClassDoc>(); }
	virtual String get_class_icon_path() const override { return ""; }

	virtual bool has_method(const StringName &p_method) const override { return false; }
	virtual MethodInfo get_method_info(const StringName &p_method) const override { return MethodInfo(); }

	virtual bool is_tool() const override { return false; }
	virtual bool is_valid() const override { return true; }
	virtual bool is_abstract() const override { return false; }

	virtual ScriptLanguage *get_language() const override;

	virtual bool has_script_signal(const StringName &p_signal) const override { return false; }
	virtual void get_script_signal_list(List<MethodInfo> *r_signals) const override {}
	virtual bool get_property_default_value(const StringName &p_property, Variant &r_value) const override { return false; }
	virtual void get_script_method_list(List<MethodInfo> *p_list) const override {}
	virtual void get_script_property_list(List<PropertyInfo> *p_list) const override {}

	virtual const Variant get_rpc_config() const override { return Variant(); }

private:
	String source_code;
};

class ResourceFormatLoaderAGSScript : public ResourceFormatLoader {
public:
	virtual Ref<Resource> load(const String &p_path, const String &p_original_path = "",
			Error *r_error = nullptr, bool p_use_sub_threads = false,
			float *r_progress = nullptr, CacheMode p_cache_mode = CACHE_MODE_REUSE) override;
	virtual void get_recognized_extensions(List<String> *p_extensions) const override;
	virtual bool handles_type(const String &p_type) const override;
	virtual String get_resource_type(const String &p_path) const override;
};

class ResourceFormatSaverAGSScript : public ResourceFormatSaver {
public:
	virtual Error save(const Ref<Resource> &p_resource, const String &p_path, uint32_t p_flags = 0) override;
	virtual void get_recognized_extensions(const Ref<Resource> &p_resource, List<String> *p_extensions) const override;
	virtual bool recognize(const Ref<Resource> &p_resource) const override;
};
