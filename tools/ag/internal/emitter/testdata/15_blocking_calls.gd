func test_character_blocking():
	await player.walk_to(point.door_left)
	await player.walk_straight(point.window)
	await player.say("Hello, world.")
	await player.think("I wonder what that is.")
	await player.play_animation("wave")
	await player.face_direction(e_north)
	await player.face_character(character.guard)
	await player.face_point(point.npc_guard)
	await player.run_interaction(1)

func test_global_blocking():
	await wait(60)
	await wait_key(300)
	await wait_mouse(300)
	await wait_input(300)
	await fade_out(30)
	await fade_in(30)
	await display_message("Chapter 1: The Beginning")

func test_blocking_chain():
	await player.walk_to(point.stage_centre)
	await player.face_point(point.audience)
	await fade_out(15)
	await player.say("And so it ends.")
	await fade_in(15)

func walk_and_greet():
	await player.walk_to(point.npc_guard)
	await player.say("Good day!")

func room_after_fade_in():
	await walk_and_greet()
