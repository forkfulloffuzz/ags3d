#include "register_types.h"

#include "ags_event_bus.h"
#include "ags_blocker_volume.h"
#include "ags_camera.h"
#include "ags_camera_zone.h"
#include "ags_character_base.h"
#include "ags_character_3d.h"
#include "ags_character_2d.h"
#include "ags_spawn_point.h"
#include "ags_item.h"
#include "ags_room_item.h"
#include "ags_hotspot.h"
#include "ags_point.h"
#include "ags_room.h"
#include "ags_runtime.h"
#include "ags_script.h"
#include "ags_trigger_region.h"
#include "ags_walkable_surface.h"
#include "ags_script_language.h"

#include "core/config/engine.h"
#include "core/io/resource_loader.h"
#include "core/io/resource_saver.h"
#include "core/object/class_db.h"
#include "core/os/os.h"

static AGSScriptLanguage *ags_language = nullptr;
static AGSRuntime *ags_runtime = nullptr;
static AGSEventBus *ags_event_bus = nullptr;
static Ref<ResourceFormatLoaderAGSScript> ags_loader;
static Ref<ResourceFormatSaverAGSScript> ags_saver;

void initialize_agvm_module(ModuleInitializationLevel p_level) {
	if (p_level != MODULE_INITIALIZATION_LEVEL_SCENE) {
		return;
	}
	OS::get_singleton()->print("AGS3D: agvm module loaded.\n");

	GDREGISTER_CLASS(AGSScript);
	GDREGISTER_CLASS(AGSRuntime);
	GDREGISTER_CLASS(AGSEventBus);
	GDREGISTER_CLASS(AGSCamera);
	GDREGISTER_CLASS(AGSCameraZone);
	GDREGISTER_CLASS(AGSCharacterBase);
	GDREGISTER_CLASS(AGSCharacter3D);
	GDREGISTER_CLASS(AGSCharacter2D);
	GDREGISTER_CLASS(AGSRoom);
	GDREGISTER_CLASS(AGSPoint);
	GDREGISTER_CLASS(AGSItem);
	GDREGISTER_CLASS(AGSRoomItem);
	GDREGISTER_CLASS(AGSHotspot);
	GDREGISTER_CLASS(AGSTriggerRegion);
	GDREGISTER_CLASS(AGSWalkableSurface);
	GDREGISTER_CLASS(AGSBlockerVolume);
	GDREGISTER_CLASS(AGSSpawnPoint);

	ags_loader.instantiate();
	ResourceLoader::add_resource_format_loader(ags_loader);

	ags_saver.instantiate();
	ResourceSaver::add_resource_format_saver(ags_saver);

	ags_runtime = memnew(AGSRuntime);
	Engine::get_singleton()->add_singleton(Engine::Singleton("AGSRuntime", ags_runtime));

	ags_event_bus = memnew(AGSEventBus);
	Engine::get_singleton()->add_singleton(Engine::Singleton("AGSEventBus", ags_event_bus));

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

	if (ags_runtime) {
		// Remove from Engine's singleton list before freeing — otherwise Engine's
		// destructor walks the list after the object is freed, corrupting the heap.
		Engine::get_singleton()->remove_singleton("AGSRuntime");
		memdelete(ags_runtime);
		ags_runtime = nullptr;
	}

	if (ags_event_bus) {
		Engine::get_singleton()->remove_singleton("AGSEventBus");
		memdelete(ags_event_bus);
		ags_event_bus = nullptr;
	}
}