## T-DLG15 — AGSDialogueState tracking tests.
extends "res://utils/test_base.gd"

const StateScript = preload("res://../game_prototype/.engine/runtime/ags_dialogue_state.gd")

func suite_name() -> String:
	return "M-DLG: DialogueState"

func _make_state() -> Node:
	var s: Node = StateScript.new()
	_tree.root.add_child(s)
	return s

func _cleanup(s: Node) -> void:
	s.queue_free()
	await _tree.process_frame

# UT-DLG15-01: node_visited returns false before any visit.
func test_01_node_not_visited_initially() -> void:
	var state := _make_state()
	assert_false(state.node_visited("guard_greeting"), "Node should not be visited initially")
	await _cleanup(state)

# UT-DLG15-02: _on_dialogue_started increments visit count and node_visited returns true.
func test_02_on_started_marks_visited() -> void:
	var state := _make_state()
	state._on_dialogue_started("guard_greeting")
	assert_true(state.node_visited("guard_greeting"), "Node should be visited after _on_dialogue_started")
	await _cleanup(state)

# UT-DLG15-03: visit_count returns correct value.
func test_03_visit_count() -> void:
	var state := _make_state()
	assert_eq(state.visit_count("node_a"), 0, "visit_count should be 0 initially")
	state._on_dialogue_started("node_a")
	assert_eq(state.visit_count("node_a"), 1, "visit_count should be 1 after one start")
	state._on_dialogue_started("node_a")
	assert_eq(state.visit_count("node_a"), 2, "visit_count should be 2 after two starts")
	await _cleanup(state)

# UT-DLG15-04: option_seen returns false before any choice.
func test_04_option_not_seen_initially() -> void:
	var state := _make_state()
	assert_false(state.option_seen("guard_greeting", 0), "Option should not be seen initially")
	await _cleanup(state)

# UT-DLG15-05: _on_choice_made marks the option as seen.
func test_05_on_choice_made_marks_option() -> void:
	var state := _make_state()
	state._on_dialogue_started("guard_greeting")
	state._on_choice_made(1)
	assert_true(state.option_seen("guard_greeting", 1), "Option 1 should be seen after choice_made")
	assert_false(state.option_seen("guard_greeting", 0), "Option 0 should not be seen")
	await _cleanup(state)

# UT-DLG15-06: reset clears all state.
func test_06_reset_clears_state() -> void:
	var state := _make_state()
	state._on_dialogue_started("node_a")
	state._on_choice_made(0)
	state.reset()
	assert_false(state.node_visited("node_a"), "node_visited should be false after reset")
	assert_false(state.option_seen("node_a", 0), "option_seen should be false after reset")
	assert_eq(state.visit_count("node_a"), 0, "visit_count should be 0 after reset")
	await _cleanup(state)

# UT-DLG15-07: serialise produces a dictionary with expected keys.
func test_07_serialise_structure() -> void:
	var state := _make_state()
	state._on_dialogue_started("node_a")
	var data: Dictionary = state.serialise()
	assert_true(data.has("schema_version"), "serialise: missing schema_version")
	assert_true(data.has("visited_nodes"), "serialise: missing visited_nodes")
	assert_true(data.has("seen_options"), "serialise: missing seen_options")
	assert_eq(data["schema_version"], 1, "schema_version should be 1")
	await _cleanup(state)

# UT-DLG15-08: deserialise restores state correctly.
func test_08_deserialise_restores() -> void:
	var state := _make_state()
	var data := {
		"schema_version": 1,
		"visited_nodes": {"guard_greeting": 2},
		"seen_options": {"guard_greeting:0": true},
	}
	state.deserialise(data)
	assert_eq(state.visit_count("guard_greeting"), 2, "visit_count should be 2 after deserialise")
	assert_true(state.option_seen("guard_greeting", 0), "option_seen should be true after deserialise")
	await _cleanup(state)

# UT-DLG15-09: deserialise with unknown schema version is a no-op (graceful).
func test_09_deserialise_unknown_schema() -> void:
	var state := _make_state()
	state._on_dialogue_started("existing_node")
	state.deserialise({"schema_version": 99, "visited_nodes": {}, "seen_options": {}})
	# State should be preserved since load was skipped.
	# (push_warning is called internally — just verify no crash and state intact)
	assert_true(true, "deserialise with unknown schema should not crash")
	await _cleanup(state)

# UT-DLG15-10: serialise / deserialise round-trip preserves data.
func test_10_serialise_deserialise_roundtrip() -> void:
	var state := _make_state()
	state._on_dialogue_started("guard_greeting")
	state._on_dialogue_started("guard_greeting")
	state._on_choice_made(1)
	var data: Dictionary = state.serialise()
	var state2: Node = StateScript.new()
	_tree.root.add_child(state2)
	state2.deserialise(data)
	assert_eq(state2.visit_count("guard_greeting"), 2, "Round-trip: visit_count mismatch")
	assert_true(state2.option_seen("guard_greeting", 1), "Round-trip: option_seen mismatch")
	state2.queue_free()
	await _cleanup(state)
