## T-CUT18 — AGSSequencerCommands audio command tests.
##
## Audio playback is not tested (no AudioServer in headless — files not present).
## Tests verify: channel tracking, command dispatch (no crash), fade paths,
## ambient player creation/cleanup, and voice dispatch.
extends "res://utils/test_base.gd"

const SeqCmds := preload("res://../game_prototype/.engine/runtime/ags_sequencer_commands.gd")

func suite_name() -> String:
	return "M-CUT: AudioCommands (T-CUT18)"


# ── Helpers ────────────────────────────────────────────────────────────────────

func _make_seq() -> Node:
	var s: Node = SeqCmds.new()
	_tree.root.add_child(s)
	return s

func _run_step(seq: Node, step: Dictionary) -> Array:
	var completed := [false]
	var failed := [false]
	seq.sequence_complete.connect(func() -> void: completed[0] = true, CONNECT_ONE_SHOT)
	seq.sequence_failed.connect(func(_r: String) -> void: failed[0] = true, CONNECT_ONE_SHOT)
	await seq.run([step])
	return [completed[0], failed[0]]

func _cleanup_nodes(nodes: Array) -> void:
	for n: Node in nodes:
		if is_instance_valid(n):
			n.queue_free()
	await _tree.process_frame


# ── UT-CUT18-01: music — no crash, channel tracked ───────────────────────────

func test_01_music_play_tracks_channel() -> void:
	note("No audio file present headless — play_music is called but may warn. " +
		"Channel tracking is what we verify.")
	var seq := _make_seq()

	var step := {"type": "audio", "command": "music", "name": "test_track"}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "music step should complete")
	assert_true(seq._audio_channels.has("music"), "music channel should be tracked")
	assert_eq(seq._audio_channels["music"]["name"], "test_track", "channel name should match")
	await _cleanup_nodes([seq])


# ── UT-CUT18-02: music_stop — removes channel tracking ───────────────────────

func test_02_music_stop_removes_channel() -> void:
	var seq := _make_seq()
	# Pre-populate channel tracking as if music was started.
	seq._audio_channels["music"] = {"type": "music", "name": "test_track"}

	var step := {"type": "audio", "command": "music_stop"}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "music_stop step should complete")
	assert_false(seq._audio_channels.has("music"), "music channel should be removed after stop")
	await _cleanup_nodes([seq])


# ── UT-CUT18-03: sound — no crash, not tracked (fire-and-forget) ──────────────

func test_03_sound_no_crash_not_tracked() -> void:
	note("play_sound called; no audio file present headless — may warn about missing file.")
	var seq := _make_seq()

	var step := {"type": "audio", "command": "sound", "name": "explosion"}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "sound step should complete")
	assert_false(seq._audio_channels.has("sound"), "sound should NOT be in channel tracking (fire-and-forget)")
	await _cleanup_nodes([seq])


# ── UT-CUT18-04: ambient — creates player and tracks channel ──────────────────

func test_04_ambient_creates_player_and_tracks_channel() -> void:
	note("No ambient audio file — ambient player is created but stream is null.")
	var seq := _make_seq()

	var step := {"type": "audio", "command": "ambient", "name": "forest"}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "ambient step should complete")
	assert_true(seq._audio_channels.has("ambient_forest"), "ambient_forest channel should be tracked")
	assert_true(seq._ambient_players.has("ambient_forest"), "ambient_forest player should exist")
	await _cleanup_nodes([seq])


# ── UT-CUT18-05: ambient_stop — removes channel and stops player ──────────────

func test_05_ambient_stop_removes_channel() -> void:
	var seq := _make_seq()
	# Create an ambient player manually and track it.
	var player := AudioStreamPlayer.new()
	seq.add_child(player)
	seq._ambient_players["ambient_wind"] = player
	seq._audio_channels["ambient_wind"] = {"type": "ambient", "name": "wind"}

	var step := {"type": "audio", "command": "ambient_stop", "name": "wind"}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "ambient_stop step should complete")
	assert_false(seq._audio_channels.has("ambient_wind"), "ambient_wind channel should be removed")
	assert_false(player.playing, "ambient player should be stopped")
	await _cleanup_nodes([seq])


# ── UT-CUT18-06: ambient_volume — adjusts player volume ──────────────────────

func test_06_ambient_volume_sets_db() -> void:
	var seq := _make_seq()
	var player := AudioStreamPlayer.new()
	player.volume_db = 0.0
	seq.add_child(player)
	seq._ambient_players["ambient_rain"] = player

	var step := {"type": "audio", "command": "ambient_volume", "channel": "rain", "value": -12.0}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "ambient_volume step should complete")
	assert_eq(player.volume_db, -12.0, "ambient player volume_db should be set to -12.0")
	await _cleanup_nodes([seq])


# ── UT-CUT18-07: ambient_volume with tween duration completes ─────────────────

func test_07_ambient_volume_tween_completes() -> void:
	var seq := _make_seq()
	var player := AudioStreamPlayer.new()
	player.volume_db = 0.0
	seq.add_child(player)
	seq._ambient_players["ambient_crowd"] = player

	var step := {"type": "audio", "command": "ambient_volume", "channel": "crowd", "value": -20.0, "duration": 0.05}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "ambient_volume tween step should complete")
	assert_eq(player.volume_db, -20.0, "volume should reach -20.0 after tween")
	await _cleanup_nodes([seq])


# ── UT-CUT18-08: voice — no crash, dispatches play_sound ─────────────────────

func test_08_voice_dispatches_without_crash() -> void:
	note("Voice dispatch calls AGSRuntime.play_sound — no file present headless, may warn.")
	var seq := _make_seq()

	var step := {"type": "audio", "command": "voice", "character": "guard", "file": "intro_01"}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "voice step should complete")
	await _cleanup_nodes([seq])


# ── UT-CUT18-09: music missing 'name' — no crash, returns true ───────────────

func test_09_music_missing_name_no_crash() -> void:
	var seq := _make_seq()

	var step := {"type": "audio", "command": "music"}  # no 'name'
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "music with missing name should complete gracefully")
	await _cleanup_nodes([seq])


# ── UT-CUT18-10: unknown audio command — no crash ────────────────────────────

func test_10_unknown_command_noop() -> void:
	var seq := _make_seq()

	var step := {"type": "audio", "command": "play_jingle_bells"}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "unknown audio command should complete as no-op")
	await _cleanup_nodes([seq])


# ── UT-CUT18-11: channel tracking accumulates across multiple audio steps ─────

func test_11_multiple_channels_tracked() -> void:
	var seq := _make_seq()

	var steps: Array = [
		{"type": "audio", "command": "music", "name": "theme"},
		{"type": "audio", "command": "ambient", "name": "wind"},
		{"type": "audio", "command": "ambient", "name": "rain"},
	]
	var completed := [false]
	seq.sequence_complete.connect(func() -> void: completed[0] = true, CONNECT_ONE_SHOT)
	await seq.run(steps)

	assert_true(seq._audio_channels.has("music"), "music channel should be tracked")
	assert_true(seq._audio_channels.has("ambient_wind"), "ambient_wind channel should be tracked")
	assert_true(seq._audio_channels.has("ambient_rain"), "ambient_rain channel should be tracked")
	assert_eq(seq._audio_channels.size(), 3, "three channels should be tracked")
	await _cleanup_nodes([seq])


# ── UT-CUT18-12: music_stop with fade_out — completes without crash ───────────

func test_12_music_stop_fade_out_no_crash() -> void:
	note("fade_out path calls _fade_music_out; AGSAudio may not expose get_music_player. " +
		"Falls back to timer-based wait. Should complete cleanly.")
	var seq := _make_seq()
	seq._audio_channels["music"] = {"type": "music", "name": "theme"}

	var step := {"type": "audio", "command": "music_stop", "fade_out": 0.05}
	var result: Array = await _run_step(seq, step)

	assert_true(result[0], "music_stop with fade_out should complete")
	assert_false(seq._audio_channels.has("music"), "music channel removed after stop")
	await _cleanup_nodes([seq])
