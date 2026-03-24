## UT-M4-11..12 — AGSHotspot signal and multi-hotspot disambiguation tests.
extends "res://utils/test_base.gd"

func suite_name() -> String:
	return "M4: Hotspots"

# Instance variable so the bound callback can update state visible after the call.
var _clicked_name := ""

func _on_hotspot_clicked(n: String) -> void:
	_clicked_name = n

# UT-M4-11: HotspotSurface fires hotspot_clicked with the correct name.
func test_11_hotspot_clicked_fires_with_correct_name() -> void:
	var hotspot: AGSHotspot = AGSHotspot.new()
	hotspot.hotspot_name = "chest"
	hotspot.notification(Node.NOTIFICATION_READY)

	_clicked_name = ""
	hotspot.connect("hotspot_clicked", _on_hotspot_clicked)
	hotspot.simulate_click()

	assert_eq(_clicked_name, "chest", "hotspot_clicked fired with wrong name")

	hotspot.free()

# UT-M4-12: Two hotspots in the same room — simulating hotspot_b fires only
# hotspot_b's signal, not hotspot_a's.
func test_12_two_hotspots_do_not_cross_fire() -> void:
	var room: AGSRoom = AGSRoom.new()
	var hotspot_a: AGSHotspot = AGSHotspot.new()
	var hotspot_b: AGSHotspot = AGSHotspot.new()
	hotspot_a.hotspot_name = "chest"
	hotspot_b.hotspot_name = "door"
	room.add_child(hotspot_a)
	room.add_child(hotspot_b)
	hotspot_a.notification(Node.NOTIFICATION_READY)
	hotspot_b.notification(Node.NOTIFICATION_READY)

	# Array is captured by reference in GDScript closures — safe to use here.
	var clicked_names: Array[String] = []
	hotspot_a.connect("hotspot_clicked", func(n: String) -> void: clicked_names.append(n))
	hotspot_b.connect("hotspot_clicked", func(n: String) -> void: clicked_names.append(n))

	hotspot_b.simulate_click()

	assert_eq(clicked_names.size(), 1, "Expected exactly one hotspot_clicked signal, got %d" % clicked_names.size())
	assert_eq(clicked_names[0], "door", "hotspot_b fired wrong name — expected 'door'")

	room.free()
