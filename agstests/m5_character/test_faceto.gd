## UT-M5-07..09 — AGSCharacterBase signal contract (headless).
##
## walk_to() and face_to() are GDScript coroutines on the runtime script
## (ags_character.gd), not C++ methods. Headless tests can only verify
## the C++ signal declarations; behavioural tests live in M6: EndToEnd.
extends "res://utils/test_base.gd"

func suite_name() -> String:
	return "M5: FaceTo"

# UT-M5-07: walk_completed signal is declared and addressable on AGSCharacterBase.
func test_07_walk_completed_signal_declared_on_base() -> void:
	var ch := AGSCharacter3D.new()
	assert_true(ch.has_signal("walk_completed"),
		"AGSCharacterBase must declare walk_completed signal")
	ch.free()

# UT-M5-08: face_completed signal is declared and addressable on AGSCharacterBase.
func test_08_face_completed_signal_declared_on_base() -> void:
	var ch := AGSCharacter3D.new()
	assert_true(ch.has_signal("face_completed"),
		"AGSCharacterBase must declare face_completed signal")
	ch.free()

# UT-M5-09: say_completed signal is declared and addressable on AGSCharacterBase.
func test_09_say_completed_signal_declared_on_base() -> void:
	var ch := AGSCharacter3D.new()
	assert_true(ch.has_signal("say_completed"),
		"AGSCharacterBase must declare say_completed signal")
	ch.free()
