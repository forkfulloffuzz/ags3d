## T-CUT31 — Audio leak cleanup tests.
##
## Verifies that cutscene-owned audio channels are faded out and stopped when
## a sequence ends normally or is skipped, and that room-scoped channels are
## left untouched.
extends "res://utils/test_base.gd"

const SeqCmds := preload("res://../game_prototype/.engine/runtime/ags_sequencer_commands.gd")

func suite_name() -> String:
	return "M-CUT: AudioCleanup (T-CUT31)"


# ── Helpers ────────────────────────────────────────────────────────────────────

func _make_seq() -> Node:
	var s: Node = SeqCmds.new()
	_tree.root.add_child(s)
	return s

func _cleanup_nodes(nodes: Array) -> void:
	for n: Node in nodes:
		if is_instance_valid(n):
			n.queue_free()
	await _tree.process_frame


# ── UT-CUT31-01: audio_cleanup_complete signal exists ────────────────────────

func test_01_audio_cleanup_complete_signal_exists() -> void:
	var seq := _make_seq()
	assert_true(seq.has_signal("audio_cleanup_complete"),
		"audio_cleanup_complete signal should exist on AGSSequencerCommands")
	await _cleanup_nodes([seq])


# ── UT-CUT31-02: cleanup emits audio_cleanup_complete on empty sequence ───────

func test_02_cleanup_emits_on_empty_sequence() -> void:
	var seq := _make_seq()
	var cleanup_fired := [false]
	seq.audio_cleanup_complete.connect(func() -> void: cleanup_fired[0] = true, CONNECT_ONE_SHOT)

	var completed := [false]
	seq.sequence_complete.connect(func() -> void: completed[0] = true, CONNECT_ONE_SHOT)
	await seq.run([])

	assert_true(completed[0], "sequence should complete")
	assert_true(cleanup_fired[0], "audio_cleanup_complete should fire even with no channels")
	await _cleanup_nodes([seq])


# ── UT-CUT31-03: cutscene-owned music channel removed after sequence ──────────

func test_03_music_channel_cleared_after_sequence() -> void:
	var seq := _make_seq()
	# Inject a music channel as if a <<music>> step ran.
	seq._audio_channels["music"] = {"type": "music", "name": "theme", "scope": "cutscene"}

	var cleanup_fired := [false]
	seq.audio_cleanup_complete.connect(func() -> void: cleanup_fired[0] = true, CONNECT_ONE_SHOT)

	await seq.run([])

	assert_true(cleanup_fired[0], "audio_cleanup_complete should fire")
	assert_false(seq._audio_channels.has("music"),
		"music channel should be removed after cleanup")
	await _cleanup_nodes([seq])


# ── UT-CUT31-04: room-scoped music channel survives sequence end ──────────────

func test_04_room_scoped_music_not_cleaned_up() -> void:
	var seq := _make_seq()
	seq._audio_channels["music"] = {"type": "music", "name": "room_theme", "scope": "room"}

	var cleanup_fired := [false]
	seq.audio_cleanup_complete.connect(func() -> void: cleanup_fired[0] = true, CONNECT_ONE_SHOT)

	await seq.run([])

	assert_true(cleanup_fired[0], "audio_cleanup_complete should fire")
	assert_true(seq._audio_channels.has("music"),
		"room-scoped music channel should NOT be removed by cutscene cleanup")
	await _cleanup_nodes([seq])


# ── UT-CUT31-05: ambient player stopped and removed after sequence ────────────

func test_05_ambient_channel_stopped_after_sequence() -> void:
	var seq := _make_seq()
	# Inject a playing ambient player.
	var player := AudioStreamPlayer.new()
	seq.add_child(player)
	seq._ambient_players["ambient_wind"] = player
	seq._audio_channels["ambient_wind"] = {"type": "ambient", "name": "wind", "scope": "cutscene"}

	var cleanup_fired := [false]
	seq.audio_cleanup_complete.connect(func() -> void: cleanup_fired[0] = true, CONNECT_ONE_SHOT)

	await seq.run([])

	assert_true(cleanup_fired[0], "audio_cleanup_complete should fire")
	assert_false(seq._audio_channels.has("ambient_wind"),
		"ambient_wind channel should be removed after cleanup")
	assert_false(seq._ambient_players.has("ambient_wind"),
		"ambient_wind player should be removed from _ambient_players")
	await _cleanup_nodes([seq])


# ── UT-CUT31-06: room-scoped ambient player survives sequence end ─────────────

func test_06_room_scoped_ambient_not_cleaned_up() -> void:
	var seq := _make_seq()
	var player := AudioStreamPlayer.new()
	seq.add_child(player)
	seq._ambient_players["ambient_rain"] = player
	seq._audio_channels["ambient_rain"] = {"type": "ambient", "name": "rain", "scope": "room"}

	var cleanup_fired := [false]
	seq.audio_cleanup_complete.connect(func() -> void: cleanup_fired[0] = true, CONNECT_ONE_SHOT)

	await seq.run([])

	assert_true(cleanup_fired[0], "audio_cleanup_complete should fire")
	assert_true(seq._audio_channels.has("ambient_rain"),
		"room-scoped ambient channel should NOT be removed by cutscene cleanup")
	assert_true(seq._ambient_players.has("ambient_rain"),
		"room-scoped ambient player should remain in _ambient_players")
	await _cleanup_nodes([seq])


# ── UT-CUT31-07: multiple cutscene channels all cleaned up ───────────────────

func test_07_multiple_channels_all_cleaned_up() -> void:
	var seq := _make_seq()
	var player1 := AudioStreamPlayer.new()
	var player2 := AudioStreamPlayer.new()
	seq.add_child(player1)
	seq.add_child(player2)
	seq._ambient_players["ambient_forest"] = player1
	seq._ambient_players["ambient_crowd"] = player2
	seq._audio_channels["music"] = {"type": "music", "name": "battle", "scope": "cutscene"}
	seq._audio_channels["ambient_forest"] = {"type": "ambient", "name": "forest", "scope": "cutscene"}
	seq._audio_channels["ambient_crowd"] = {"type": "ambient", "name": "crowd", "scope": "cutscene"}

	await seq.run([])

	assert_false(seq._audio_channels.has("music"), "music channel cleared")
	assert_false(seq._audio_channels.has("ambient_forest"), "ambient_forest channel cleared")
	assert_false(seq._audio_channels.has("ambient_crowd"), "ambient_crowd channel cleared")
	await _cleanup_nodes([seq])


# ── UT-CUT31-08: cleanup fires after skip ────────────────────────────────────

func test_08_cleanup_fires_after_skip() -> void:
	var seq := _make_seq()
	seq.skip_policy = "always"
	seq._audio_channels["music"] = {"type": "music", "name": "intro", "scope": "cutscene"}

	var cleanup_fired := [false]
	seq.audio_cleanup_complete.connect(func() -> void: cleanup_fired[0] = true, CONNECT_ONE_SHOT)

	# Trigger skip immediately after a wait starts.
	_tree.create_timer(0.05).timeout.connect(func() -> void: seq.request_skip(), CONNECT_ONE_SHOT)

	await seq.run([{"type": "wait", "duration": 10.0}])

	assert_true(cleanup_fired[0], "audio_cleanup_complete should fire after skip")
	assert_false(seq._audio_channels.has("music"), "music channel should be cleared after skip")
	await _cleanup_nodes([seq])


# ── UT-CUT31-09: mixed scopes — only cutscene channels cleaned up ─────────────

func test_09_mixed_scopes_only_cutscene_cleaned() -> void:
	var seq := _make_seq()
	seq._audio_channels["music"] = {"type": "music", "name": "cut_music", "scope": "cutscene"}
	seq._audio_channels["ambient_rain"] = {"type": "ambient", "name": "rain", "scope": "room"}

	await seq.run([])

	assert_false(seq._audio_channels.has("music"),
		"cutscene-scoped music should be cleaned up")
	assert_true(seq._audio_channels.has("ambient_rain"),
		"room-scoped ambient should survive")
	await _cleanup_nodes([seq])


# ── UT-CUT31-10: cleanup_complete fires exactly once per run ─────────────────

func test_10_cleanup_fires_exactly_once() -> void:
	var seq := _make_seq()
	var count := [0]
	var cb := func() -> void: count[0] += 1
	seq.audio_cleanup_complete.connect(cb)

	await seq.run([])
	await seq.run([])

	assert_eq(count[0], 2, "audio_cleanup_complete should fire once per run() call")
	seq.audio_cleanup_complete.disconnect(cb)
	await _cleanup_nodes([seq])
