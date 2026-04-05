func test_character_blocking():
	await AGSRuntime.get_character("player").walk_to("door_left")
	await AGSRuntime.get_global("player").walk_straight(point.window)
	await AGSRuntime.get_character("player").say("Hello, world.")
	await AGSRuntime.get_character("player").think("I wonder what that is.")
	await AGSRuntime.get_global("player").play_animation("wave")
	await AGSRuntime.get_global("player").face_direction(e_north)
	await AGSRuntime.get_global("player").face_character(character.guard)
	await AGSRuntime.get_global("player").face_point(point.npc_guard)
	await AGSRuntime.get_global("player").run_interaction(1)

func test_global_blocking():
	await AGSCutscene.wait(60)
	await wait_key(300)
	await wait_mouse(300)
	await wait_input(300)
	await AGSCutscene.fade_out(30)
	await AGSCutscene.fade_in(30)
	await display_message("Chapter 1: The Beginning")

func test_blocking_chain():
	await AGSRuntime.get_character("player").walk_to("stage_centre")
	await AGSRuntime.get_global("player").face_point(point.audience)
	await AGSCutscene.fade_out(15)
	await AGSRuntime.get_character("player").say("And so it ends.")
	await AGSCutscene.fade_in(15)

func walk_and_greet():
	await AGSRuntime.get_character("player").walk_to("npc_guard")
	await AGSRuntime.get_character("player").say("Good day!")

func room_after_fade_in():
	await walk_and_greet()
