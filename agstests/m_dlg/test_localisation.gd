## T-DLG17 — AGSLocalisation runtime tests.
extends "res://utils/test_base.gd"

const LocScript = preload("res://../game_prototype/.engine/runtime/ags_localisation.gd")

func suite_name() -> String:
	return "M-DLG: Localisation"

func _make_loc() -> Node:
	var s: Node = LocScript.new()
	_tree.root.add_child(s)
	return s

func _cleanup(s: Node) -> void:
	s.queue_free()
	await _tree.process_frame

# Helper: inject a string table directly (no file I/O needed in tests).
func _inject_table(loc: Node, code: String, table: Dictionary) -> void:
	loc._tables[code] = table
	loc._tables[code]["__rtl__"] = table.get("__rtl__", false)

# UT-DLG17-01: get_string returns fallback when key not found.
func test_01_get_returns_fallback() -> void:
	var loc := _make_loc()
	var result := loc.get_string("missing:key", "fallback text")
	assert_eq(result, "fallback text", "Should return fallback when key absent")
	await _cleanup(loc)

# UT-DLG17-02: get_string returns key itself when no fallback provided and key missing.
func test_02_get_returns_key_when_no_fallback() -> void:
	var loc := _make_loc()
	var result := loc.get_string("some:key:abc")
	assert_eq(result, "some:key:abc", "Should return loc_key when no fallback and key missing")
	await _cleanup(loc)

# UT-DLG17-03: set_locale changes active_locale.
func test_03_set_locale_changes_active() -> void:
	var loc := _make_loc()
	_inject_table(loc, "fr", {"hello:0:abc": "Bonjour"})
	loc.set_locale("fr")
	assert_eq(loc.active_locale(), "fr", "active_locale should be 'fr' after set_locale")
	await _cleanup(loc)

# UT-DLG17-04: get_string returns translated string in active locale.
func test_04_get_returns_translated_string() -> void:
	var loc := _make_loc()
	_inject_table(loc, "fr", {"hello:0:abc": "Bonjour"})
	loc.set_locale("fr")
	var result := loc.get_string("hello:0:abc", "Hello")
	assert_eq(result, "Bonjour", "Should return French translation")
	await _cleanup(loc)

# UT-DLG17-05: fallback chain is tried when key missing from active locale.
func test_05_fallback_chain_used() -> void:
	var loc := _make_loc()
	_inject_table(loc, "es", {})  # empty Spanish table — key absent
	_inject_table(loc, "fr", {"greeting:0:abc": "Bonjour"})
	_inject_table(loc, "en", {"greeting:0:abc": "Hello"})
	loc.fallback_chain = ["fr"]
	loc.base_locale = "en"
	loc.set_locale("es")
	var result := loc.get_string("greeting:0:abc", "fallback")
	assert_eq(result, "Bonjour", "Fallback chain should supply French translation")
	await _cleanup(loc)

# UT-DLG17-06: base_locale is tried last before returning fallback_text.
func test_06_base_locale_last_resort() -> void:
	var loc := _make_loc()
	_inject_table(loc, "en", {"base_key:0:abc": "Base text"})
	_inject_table(loc, "de", {})
	loc.base_locale = "en"
	loc.set_locale("de")
	var result := loc.get_string("base_key:0:abc", "fallback")
	assert_eq(result, "Base text", "Base locale should supply string when active locale missing key")
	await _cleanup(loc)

# UT-DLG17-07: set_locale emits locale_changed signal.
func test_07_locale_changed_emitted() -> void:
	var loc := _make_loc()
	_inject_table(loc, "fr", {})
	var received: String = ""
	loc.locale_changed.connect(func(code: String) -> void: received = code)
	loc.set_locale("fr")
	assert_eq(received, "fr", "locale_changed should emit with new locale code")
	await _cleanup(loc)

# UT-DLG17-08: set_locale is a no-op if locale unchanged.
func test_08_set_same_locale_noop() -> void:
	var loc := _make_loc()
	loc.base_locale = "en"
	_inject_table(loc, "en", {})
	var count: int = 0
	loc.locale_changed.connect(func(_c: String) -> void: count += 1)
	loc.set_locale("en")  # same as active — should be no-op
	assert_eq(count, 0, "locale_changed should not emit when locale unchanged")
	await _cleanup(loc)

# UT-DLG17-09: is_rtl returns false by default.
func test_09_rtl_false_by_default() -> void:
	var loc := _make_loc()
	assert_false(loc.is_rtl(), "is_rtl should be false by default")
	await _cleanup(loc)

# UT-DLG17-10: is_rtl returns true for locale with __rtl__ = true.
func test_10_rtl_true_for_rtl_locale() -> void:
	var loc := _make_loc()
	_inject_table(loc, "ar", {"__rtl__": true})
	loc.set_locale("ar")
	assert_true(loc.is_rtl(), "is_rtl should be true for Arabic locale")
	await _cleanup(loc)

# UT-DLG17-11: get() alias works identically to get_string().
func test_11_get_alias() -> void:
	var loc := _make_loc()
	_inject_table(loc, "en", {"k:0:x": "Value"})
	loc.base_locale = "en"
	var result := loc.get("k:0:x", "fallback")
	assert_eq(result, "Value", "get() alias should work like get_string()")
	await _cleanup(loc)
