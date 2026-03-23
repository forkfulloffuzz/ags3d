#include "register_types.h"

#include "ags_room.h"
#include "ags_script.h"
#include "ags_script_language.h"

#include "core/io/resource_loader.h"
#include "core/io/resource_saver.h"
#include "core/object/class_db.h"
#include "core/os/os.h"

static AGSScriptLanguage *ags_language = nullptr;
static Ref<ResourceFormatLoaderAGSScript> ags_loader;
static Ref<ResourceFormatSaverAGSScript> ags_saver;

void initialize_agvm_module(ModuleInitializationLevel p_level) {
	if (p_level != MODULE_INITIALIZATION_LEVEL_SCENE) {
		return;
	}
	OS::get_singleton()->print("AGS3D: agvm module loaded.\n");

	GDREGISTER_CLASS(AGSScript);
	GDREGISTER_CLASS(AGSRoom);

	ags_loader.instantiate();
	ResourceLoader::add_resource_format_loader(ags_loader);

	ags_saver.instantiate();
	ResourceSaver::add_resource_format_saver(ags_saver);

	ags_language = memnew(AGSScriptLanguage);
	ScriptServer::register_language(ags_language);
}

void uninitialize_agvm_module(ModuleInitializationLevel p_level) {
	if (p_level != MODULE_INITIALIZATION_LEVEL_SCENE) {
		return;
	}
	if (ags_language) {
		ScriptServer::unregister_language(ags_language);
		memdelete(ags_language);
		ags_language = nullptr;
	}
	ResourceLoader::remove_resource_format_loader(ags_loader);
	ags_loader.unref();
	ResourceSaver::remove_resource_format_saver(ags_saver);
	ags_saver.unref();
}