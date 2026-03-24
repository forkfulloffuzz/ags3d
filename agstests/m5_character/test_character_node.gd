## UT-M5-01..03 — AGSCharacter instantiation and AGSRuntime registration tests.
extends "res://utils/test_base.gd"

func suite_name() -> String:
	return "M5: CharacterNode"

# UT-M5-01: AGSCharacter instantiates without crash.
func test_01_ags_character_instantiates() -> void:
	var ch: AGSCharacter = AGSCharacter.new()
	assert_not_null(ch, "AGSCharacter.new() returned null")
	ch.free()

# UT-M5-02: Character registers with AGSRuntime by character_name on ready.
func test_02_character_registers_with_runtime() -> void:
	var ch: AGSCharacter = AGSCharacter.new()
	ch.character_name = "player"
	ch.notification(Node.NOTIFICATION_READY)

	var found = AGSRuntime.get_character("player")
	assert_not_null(found, "AGSRuntime.get_character('player') returned null after registration")
	assert_eq(found, ch, "AGSRuntime returned wrong character instance")

	# Clean up — unregister by triggering EXIT_TREE notification.
	ch.notification(Node.NOTIFICATION_EXIT_TREE)
	ch.free()

# UT-M5-03: Two characters with different names are both retrievable from AGSRuntime.
func test_03_two_characters_independently_retrievable() -> void:
	var ch_a: AGSCharacter = AGSCharacter.new()
	var ch_b: AGSCharacter = AGSCharacter.new()
	ch_a.character_name = "hero"
	ch_b.character_name = "villain"
	ch_a.notification(Node.NOTIFICATION_READY)
	ch_b.notification(Node.NOTIFICATION_READY)

	assert_eq(AGSRuntime.get_character("hero"), ch_a, "get_character('hero') returned wrong node")
	assert_eq(AGSRuntime.get_character("villain"), ch_b, "get_character('villain') returned wrong node")
	assert_ne(AGSRuntime.get_character("hero"), ch_b, "hero and villain resolve to the same node")

	ch_a.notification(Node.NOTIFICATION_EXIT_TREE)
	ch_b.notification(Node.NOTIFICATION_EXIT_TREE)
	ch_a.free()
	ch_b.free()
