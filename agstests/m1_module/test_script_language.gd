## UT-M1-01..05 — AGSScriptLanguage registration tests.
## Tests language registration via ResourceLoader (the GDScript-visible surface).
extends "res://utils/test_base.gd"

const FIXTURE_PATH := "res://m1_module/fixtures/empty.agscript"

func suite_name() -> String:
	return "M1: ScriptLanguage"

# UT-M1-01: Engine boots headlessly without crash.
# Reaching this line proves the engine started successfully.
func test_01_engine_boots_without_crash() -> void:
	assert_true(true, "Engine did not boot")

# UT-M1-02: ResourceLoader recognises .agscript as a loadable resource.
func test_02_agscript_extension_recognised() -> void:
	var recognised := ResourceLoader.exists(FIXTURE_PATH)
	assert_true(recognised, "ResourceLoader does not recognise .agscript as a loadable resource")

# UT-M1-03: ResourceLoader can load an .agscript file without error.
func test_03_agscript_loads_without_error() -> void:
	var res: Resource = ResourceLoader.load(FIXTURE_PATH)
	assert_not_null(res, "ResourceLoader.load() returned null for .agscript")

# UT-M1-04: Loaded resource is a Script subclass.
func test_04_loaded_resource_is_script() -> void:
	var res: Resource = ResourceLoader.load(FIXTURE_PATH)
	assert_not_null(res, "ResourceLoader.load() returned null")
	assert_true(res is Script, "Loaded .agscript resource is not a Script")

# UT-M1-05: Loaded resource class name is AGSScript (confirms correct type registered).
func test_05_get_language_returns_ags_language() -> void:
	var res: Resource = ResourceLoader.load(FIXTURE_PATH)
	assert_not_null(res, "ResourceLoader.load() returned null")
	# get_language() is not bound to GDScript from C++, so verify via class name.
	assert_eq(res.get_class(), "AGSScript", "Loaded resource class is not AGSScript")
