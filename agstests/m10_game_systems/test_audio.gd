## UT-M10-60..65 — Audio: PlayMusic, StopMusic, PlaySound signals (T-GS12)
##
## AGSRuntime is an engine singleton — never call AGSRuntime.new().
## Tests verify that play_music / stop_music / play_sound emit the correct
## signals with the correct arguments. Actual AudioStreamPlayer playback is
## not tested headlessly.
extends "res://utils/test_base.gd"

func suite_name() -> String:
	return "M10: Audio"


# ── Signal emission tests ─────────────────────────────────────────────────────

# UT-M10-60: play_music emits play_music_requested with the music name.
func test_60_play_music_emits_signal() -> void:
	var fired := [false]
	var got_name := [""]
	AGSRuntime.play_music_requested.connect(func(name: String) -> void:
		fired[0] = true
		got_name[0] = name
	, CONNECT_ONE_SHOT)
	AGSRuntime.play_music("theme_main")
	assert_true(fired[0], "play_music_requested should fire")
	assert_eq(got_name[0], "theme_main", "play_music_requested should carry the name")


# UT-M10-61: play_music with a different name carries that name.
func test_61_play_music_carries_name() -> void:
	var got_name := [""]
	AGSRuntime.play_music_requested.connect(func(name: String) -> void:
		got_name[0] = name
	, CONNECT_ONE_SHOT)
	AGSRuntime.play_music("boss_theme")
	assert_eq(got_name[0], "boss_theme", "play_music_requested should carry 'boss_theme'")


# UT-M10-62: stop_music emits stop_music_requested.
func test_62_stop_music_emits_signal() -> void:
	var fired := [false]
	AGSRuntime.stop_music_requested.connect(func() -> void:
		fired[0] = true
	, CONNECT_ONE_SHOT)
	AGSRuntime.stop_music()
	assert_true(fired[0], "stop_music_requested should fire")


# UT-M10-63: play_sound emits play_sound_requested with the sound name.
func test_63_play_sound_emits_signal() -> void:
	var fired := [false]
	var got_name := [""]
	AGSRuntime.play_sound_requested.connect(func(name: String) -> void:
		fired[0] = true
		got_name[0] = name
	, CONNECT_ONE_SHOT)
	AGSRuntime.play_sound("door_creak")
	assert_true(fired[0], "play_sound_requested should fire")
	assert_eq(got_name[0], "door_creak", "play_sound_requested should carry the name")


# UT-M10-64: multiple play_music calls each emit a signal.
func test_64_multiple_play_music_calls() -> void:
	var count := [0]
	var names: Array[String] = []
	var handler := func(name: String) -> void:
		count[0] += 1
		names.append(name)
	AGSRuntime.play_music_requested.connect(handler)
	AGSRuntime.play_music("track_a")
	AGSRuntime.play_music("track_b")
	AGSRuntime.play_music_requested.disconnect(handler)
	assert_eq(count[0], 2, "play_music_requested should fire twice")
	assert_eq(names[0], "track_a", "first call should be track_a")
	assert_eq(names[1], "track_b", "second call should be track_b")


# UT-M10-65: play_sound and play_music are independent signals.
func test_65_sound_and_music_independent() -> void:
	var music_fired := [false]
	var sound_fired := [false]
	AGSRuntime.play_music_requested.connect(func(_n: String) -> void:
		music_fired[0] = true
	, CONNECT_ONE_SHOT)
	AGSRuntime.play_sound_requested.connect(func(_n: String) -> void:
		sound_fired[0] = true
	, CONNECT_ONE_SHOT)
	AGSRuntime.play_sound("click")
	assert_false(music_fired[0], "play_music_requested should not fire on play_sound")
	assert_true(sound_fired[0], "play_sound_requested should fire on play_sound")
