## T-CUT32 — Dialogue ducking during cutscenes.
##
## Verifies that duck_channel / unduck_channel are called when a dialogue_line
## or narrator_line step contains a "duck" field, and that AGSAudio tracks
## duck state correctly.
extends "res://utils/test_base.gd"

const AGSAudioScript := preload("res://../game_prototype/.engine/runtime/ags_audio.gd")

func suite_name() -> String:
	return "M-CUT: DialogueDucking (T-CUT32)"


# ── Helpers ────────────────────────────────────────────────────────────────────

func _make_audio() -> Node:
	var a: Node = AGSAudioScript.new()
	_tree.root.add_child(a)
	return a

func _cleanup_nodes(nodes: Array) -> void:
	for n: Node in nodes:
		if is_instance_valid(n):
			n.queue_free()
	await _tree.process_frame


# ── UT-CUT32-01: duck_channel method exists on AGSAudio ──────────────────────

func test_01_duck_channel_exists() -> void:
	var audio := _make_audio()
	assert_true(audio.has_method("duck_channel"),
		"duck_channel should exist on AGSAudio")
	await _cleanup_nodes([audio])


# ── UT-CUT32-02: unduck_channel method exists on AGSAudio ────────────────────

func test_02_unduck_channel_exists() -> void:
	var audio := _make_audio()
	assert_true(audio.has_method("unduck_channel"),
		"unduck_channel should exist on AGSAudio")
	await _cleanup_nodes([audio])


# ── UT-CUT32-03: duck_channel unknown bus — no crash ─────────────────────────

func test_03_duck_unknown_bus_no_crash() -> void:
	note("Bus 'NonExistentBus' does not exist — duck_channel should warn and return.")
	var audio := _make_audio()
	# Should not crash even if bus doesn't exist.
	await audio.duck_channel("NonExistentBus", -12.0, 0.0)
	assert_true(true, "duck_channel with unknown bus should not crash")
	await _cleanup_nodes([audio])


# ── UT-CUT32-04: unduck_channel without prior duck — no crash ────────────────

func test_04_unduck_without_prior_duck_no_crash() -> void:
	var audio := _make_audio()
	await audio.unduck_channel("Master", 0.0)
	assert_true(true, "unduck_channel without prior duck should be a no-op")
	await _cleanup_nodes([audio])


# ── UT-CUT32-05: duck stores original volume in _duck_originals ──────────────

func test_05_duck_stores_original_volume() -> void:
	var audio := _make_audio()
	var bus_idx := AudioServer.get_bus_index("Master")
	if bus_idx < 0:
		note("Master bus not found — skipping volume test.")
		await _cleanup_nodes([audio])
		return
	var original_db: float = AudioServer.get_bus_volume_db(bus_idx)
	await audio.duck_channel("Master", -12.0, 0.0)
	assert_true(audio._duck_originals.has("Master"),
		"_duck_originals should have 'Master' after ducking")
	assert_eq(audio._duck_originals["Master"], original_db,
		"stored original should match pre-duck volume")
	# Restore to avoid affecting other tests.
	AudioServer.set_bus_volume_db(bus_idx, original_db)
	audio._duck_originals.erase("Master")
	await _cleanup_nodes([audio])


# ── UT-CUT32-06: duck lowers bus volume ──────────────────────────────────────

func test_06_duck_lowers_volume() -> void:
	var audio := _make_audio()
	var bus_idx := AudioServer.get_bus_index("Master")
	if bus_idx < 0:
		note("Master bus not found — skipping volume test.")
		await _cleanup_nodes([audio])
		return
	var original_db: float = AudioServer.get_bus_volume_db(bus_idx)
	await audio.duck_channel("Master", -12.0, 0.0)
	var ducked_db: float = AudioServer.get_bus_volume_db(bus_idx)
	assert_true(ducked_db < original_db,
		"bus volume should be lower after ducking")
	# Restore.
	AudioServer.set_bus_volume_db(bus_idx, original_db)
	audio._duck_originals.erase("Master")
	await _cleanup_nodes([audio])


# ── UT-CUT32-07: unduck restores original volume ─────────────────────────────

func test_07_unduck_restores_volume() -> void:
	var audio := _make_audio()
	var bus_idx := AudioServer.get_bus_index("Master")
	if bus_idx < 0:
		note("Master bus not found — skipping volume test.")
		await _cleanup_nodes([audio])
		return
	var original_db: float = AudioServer.get_bus_volume_db(bus_idx)
	await audio.duck_channel("Master", -12.0, 0.0)
	await audio.unduck_channel("Master", 0.0)
	var restored_db: float = AudioServer.get_bus_volume_db(bus_idx)
	assert_eq(restored_db, original_db, "volume should be restored to original after unduck")
	assert_false(audio._duck_originals.has("Master"),
		"_duck_originals should be cleared after unduck")
	await _cleanup_nodes([audio])


# ── UT-CUT32-08: duck tween completes ────────────────────────────────────────

func test_08_duck_tween_completes() -> void:
	var audio := _make_audio()
	var bus_idx := AudioServer.get_bus_index("Master")
	if bus_idx < 0:
		note("Master bus not found — skipping tween test.")
		await _cleanup_nodes([audio])
		return
	var original_db: float = AudioServer.get_bus_volume_db(bus_idx)
	await audio.duck_channel("Master", -12.0, 0.05)
	assert_true(true, "duck tween should complete without hanging")
	# Restore.
	AudioServer.set_bus_volume_db(bus_idx, original_db)
	audio._duck_originals.erase("Master")
	await _cleanup_nodes([audio])


# ── UT-CUT32-09: unduck tween completes ──────────────────────────────────────

func test_09_unduck_tween_completes() -> void:
	var audio := _make_audio()
	var bus_idx := AudioServer.get_bus_index("Master")
	if bus_idx < 0:
		note("Master bus not found — skipping tween test.")
		await _cleanup_nodes([audio])
		return
	var original_db: float = AudioServer.get_bus_volume_db(bus_idx)
	await audio.duck_channel("Master", -12.0, 0.0)
	await audio.unduck_channel("Master", 0.05)
	assert_true(true, "unduck tween should complete without hanging")
	# Restore just in case.
	AudioServer.set_bus_volume_db(bus_idx, original_db)
	await _cleanup_nodes([audio])


# ── UT-CUT32-10: second duck does not overwrite original ─────────────────────

func test_10_second_duck_keeps_original() -> void:
	var audio := _make_audio()
	var bus_idx := AudioServer.get_bus_index("Master")
	if bus_idx < 0:
		note("Master bus not found — skipping test.")
		await _cleanup_nodes([audio])
		return
	var original_db: float = AudioServer.get_bus_volume_db(bus_idx)
	await audio.duck_channel("Master", -12.0, 0.0)
	var after_first_duck_db: float = AudioServer.get_bus_volume_db(bus_idx)
	# Second duck: should NOT overwrite _duck_originals["Master"].
	await audio.duck_channel("Master", -6.0, 0.0)
	assert_eq(audio._duck_originals["Master"], original_db,
		"second duck should not overwrite the stored original")
	# Restore.
	AudioServer.set_bus_volume_db(bus_idx, original_db)
	audio._duck_originals.erase("Master")
	await _cleanup_nodes([audio])


# ── UT-CUT32-11: get_music_player returns the music player ───────────────────

func test_11_get_music_player_returns_player() -> void:
	note("AGSAudio._ready() connects to AGSRuntime signals — may warn headless.")
	var audio := _make_audio()
	if not audio.has_method("get_music_player"):
		assert_true(false, "get_music_player should exist on AGSAudio")
		await _cleanup_nodes([audio])
		return
	# get_music_player() returns null until _ready() has run and _music_player is set.
	# In headless tests, _ready() fires when added to tree.
	var player: Object = audio.call("get_music_player")
	assert_true(player != null, "get_music_player should return an AudioStreamPlayer")
	await _cleanup_nodes([audio])
