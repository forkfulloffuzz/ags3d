## UT-M10-40..44 — AGSRoomItem node + item_clicked signal (T-GS03)
##
## All tests are synchronous.
extends "res://utils/test_base.gd"

func suite_name() -> String:
	return "M10: RoomItem"


# UT-M10-40: AGSRoomItem has item_name property.
func test_40_room_item_property() -> void:
	var ri := AGSRoomItem.new()
	ri.item_name = "rusty_key"
	assert_eq(ri.item_name, "rusty_key", "item_name not stored")
	ri.free()


# UT-M10-41: AGSRoomItem has item_clicked signal.
func test_41_item_clicked_signal_exists() -> void:
	var ri := AGSRoomItem.new()
	assert_true(ri.has_signal("item_clicked"), "AGSRoomItem should have item_clicked signal")
	ri.free()


# UT-M10-42: simulate_click emits item_clicked with the correct item name.
func test_42_simulate_click_emits_signal() -> void:
	var ri := AGSRoomItem.new()
	ri.item_name = "coin"
	var fired := [false]
	var got_name := [""]
	ri.item_clicked.connect(func(name: String) -> void:
		fired[0] = true
		got_name[0] = name
	, CONNECT_ONE_SHOT)
	ri.simulate_click()
	assert_true(fired[0], "item_clicked should have fired")
	assert_eq(got_name[0], "coin", "item_clicked should carry the item_name")
	ri.free()


# UT-M10-43: AGSRoom has item_clicked signal.
func test_43_room_has_item_clicked_signal() -> void:
	var room := AGSRoom.new()
	assert_true(room.has_signal("item_clicked"), "AGSRoom should have item_clicked signal")
	room.free()


# UT-M10-44: simulate_click on AGSRoomItem propagates item_clicked to AGSRoom.
func test_44_click_propagates_to_room() -> void:
	var room := AGSRoom.new()
	var ri := AGSRoomItem.new()
	ri.item_name = "key"
	room.add_child(ri)

	var room_fired := [false]
	var room_name := [""]
	room.item_clicked.connect(func(name: String) -> void:
		room_fired[0] = true
		room_name[0] = name
	, CONNECT_ONE_SHOT)

	ri.simulate_click()
	assert_true(room_fired[0], "AGSRoom item_clicked should fire when room item is clicked")
	assert_eq(room_name[0], "key", "AGSRoom item_clicked should carry the item name")
	room.free()
