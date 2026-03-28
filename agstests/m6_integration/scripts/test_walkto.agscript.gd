extends AGSRoom

func room_enter():
	await AGSRuntime.get_character("e2e_player").walk_to("door_left")
	await AGSRuntime.get_character("e2e_player").face_to("window")
