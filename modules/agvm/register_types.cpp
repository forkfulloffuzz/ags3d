#include "register_types.h"
#include "core/os/os.h"

void initialize_agvm_module(ModuleInitializationLevel p_level) {
    if (p_level != MODULE_INITIALIZATION_LEVEL_SCENE) {
        return;
    }
    OS::get_singleton()->print("AGS3D: agvm module loaded.\n");
}

void uninitialize_agvm_module(ModuleInitializationLevel p_level) {
    if (p_level != MODULE_INITIALIZATION_LEVEL_SCENE) {
        return;
    }
}
