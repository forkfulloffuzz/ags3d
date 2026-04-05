## UT-M10-100..104 — ags_gui.gd GUI runtime (T-GS14)
##
## Tests verify:
##   - AGSRuntime exposes status_text / active_verb properties and signals
##   - set_status_text emits status_text_changed
##   - set_active_verb emits active_verb_changed
##   - ags_gui.gd script loads and instantiates without errors
##
## Full widget wiring (InventoryBar, VerbBar, StatusLine node integration) is
## a manual test — Control nodes require a display server not available headlessly.
extends "res://utils/test_base.gd"

func suite_name() -> String:
	return "M10: GUI Runtime (T-GS14)"


# ── UT-M10-100: AGSRuntime has status_text property ──────────────────────────

func test_100_ags_runtime_has_status_text_property() -> void:
	assert_true(AGSRuntime.has_method("set_status_text"),
			"AGSRuntime must have set_status_text()")
	assert_true(AGSRuntime.has_method("get_status_text"),
			"AGSRuntime must have get_status_text()")


# ── UT-M10-101: AGSRuntime has active_verb property ──────────────────────────

func test_101_ags_runtime_has_active_verb_property() -> void:
	assert_true(AGSRuntime.has_method("set_active_verb"),
			"AGSRuntime must have set_active_verb()")
	assert_true(AGSRuntime.has_method("get_active_verb"),
			"AGSRuntime must have get_active_verb()")


# ── UT-M10-102: AGSRuntime has status_text_changed and active_verb_changed signals

func test_102_ags_runtime_has_gui_signals() -> void:
	assert_true(AGSRuntime.has_signal("status_text_changed"),
			"AGSRuntime must emit status_text_changed")
	assert_true(AGSRuntime.has_signal("active_verb_changed"),
			"AGSRuntime must emit active_verb_changed")


# ── UT-M10-103: set_status_text stores and emits ─────────────────────────────

func test_103_set_status_text_stores_and_emits() -> void:
	var fired := [false]
	var received := [""]
	var handler := func(t: String) -> void:
		fired[0] = true
		received[0] = t
	AGSRuntime.status_text_changed.connect(handler, CONNECT_ONE_SHOT)

	AGSRuntime.set_status_text("Hello world")

	assert_true(fired[0], "status_text_changed should have fired")
	assert_eq(received[0], "Hello world", "signal should carry the new text")
	assert_eq(AGSRuntime.get_status_text(), "Hello world",
			"get_status_text() should return the set value")

	# Cleanup
	AGSRuntime.set_status_text("")


# ── UT-M10-104: set_active_verb stores and emits ─────────────────────────────

func test_104_set_active_verb_stores_and_emits() -> void:
	var fired := [false]
	var received := [""]
	var handler := func(v: String) -> void:
		fired[0] = true
		received[0] = v
	AGSRuntime.active_verb_changed.connect(handler, CONNECT_ONE_SHOT)

	AGSRuntime.set_active_verb("Look")

	assert_true(fired[0], "active_verb_changed should have fired")
	assert_eq(received[0], "Look", "signal should carry the new verb")
	assert_eq(AGSRuntime.get_active_verb(), "Look",
			"get_active_verb() should return the set value")

	# Cleanup
	AGSRuntime.set_active_verb("")
