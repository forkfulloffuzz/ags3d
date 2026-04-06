## T-CUT11 — AGSEventBusSurface GDScript surface tests.
extends "res://utils/test_base.gd"

const SurfaceScript = preload("res://../../game_prototype/.engine/runtime/ags_event_bus_surface.gd")

func suite_name() -> String:
	return "M-CUT: EventBusSurface"

func _make_surface() -> Node:
	var s: Node = SurfaceScript.new()
	_tree.root.add_child(s)
	return s

func _cleanup(s: Node) -> void:
	s.queue_free()
	await _tree.process_frame

# UT-CUT11-01: emit_event reaches one-shot handler registered via on_event.
func test_01_emit_reaches_oneshot() -> void:
	var s := _make_surface()
	var called: bool = false
	s.on_event("event:player:land", func(_p: Dictionary) -> void: called = true, true)
	s.emit_event("event:player:land")
	assert_true(called, "One-shot handler should have been called")
	await _cleanup(s)

# UT-CUT11-02: One-shot handler is removed after first call.
func test_02_oneshot_fires_once() -> void:
	var s := _make_surface()
	var count: int = 0
	s.on_event("event:player:land", func(_p: Dictionary) -> void: count += 1, true)
	s.emit_event("event:player:land")
	s.emit_event("event:player:land")
	assert_eq(count, 1, "One-shot handler should fire exactly once")
	await _cleanup(s)

# UT-CUT11-03: emit_event passes payload Dictionary to handler.
func test_03_payload_passed_to_handler() -> void:
	var s := _make_surface()
	var received: Dictionary = {}
	s.on_event("event:char:jump", func(p: Dictionary) -> void: received = p, true)
	s.emit_event("event:char:jump", {"height": 3.0})
	assert_eq(received.get("height", 0.0), 3.0, "Payload should be passed to handler")
	await _cleanup(s)

# UT-CUT11-04: room handler receives every event.
func test_04_room_handler_receives_all() -> void:
	var s := _make_surface()
	var events: Array[String] = []
	s.set_room_handler(func(name: String, _p: Dictionary) -> void: events.append(name))
	s.emit_event("event:a:x")
	s.emit_event("event:b:y")
	assert_eq(events.size(), 2, "Room handler should receive 2 events")
	assert_eq(events[0], "event:a:x", "First event mismatch")
	assert_eq(events[1], "event:b:y", "Second event mismatch")
	await _cleanup(s)

# UT-CUT11-05: clear_room_handler stops delivery to room.
func test_05_clear_room_handler() -> void:
	var s := _make_surface()
	var count: int = 0
	s.set_room_handler(func(_n: String, _p: Dictionary) -> void: count += 1)
	s.emit_event("event:a:x")
	s.clear_room_handler()
	s.emit_event("event:a:x")
	assert_eq(count, 1, "Room handler should not be called after clear")
	await _cleanup(s)

# UT-CUT11-06: Multiple one-shot handlers for same event all fire.
func test_06_multiple_oneshot_all_fire() -> void:
	var s := _make_surface()
	var a: bool = false
	var b: bool = false
	s.on_event("event:x:y", func(_p: Dictionary) -> void: a = true, true)
	s.on_event("event:x:y", func(_p: Dictionary) -> void: b = true, true)
	s.emit_event("event:x:y")
	assert_true(a, "First one-shot should fire")
	assert_true(b, "Second one-shot should fire")
	await _cleanup(s)

# UT-CUT11-07: event_dispatched signal is emitted on every dispatch.
func test_07_event_dispatched_signal() -> void:
	var s := _make_surface()
	var seen: Array[String] = []
	s.event_dispatched.connect(func(name: String, _p: Dictionary) -> void: seen.append(name))
	s.emit_event("event:player:land")
	s.emit_event("event:enemy:die")
	assert_eq(seen.size(), 2, "event_dispatched should emit twice")
	await _cleanup(s)

# UT-CUT11-08: wait_for returns when the event fires.
func test_08_wait_for_returns_on_event() -> void:
	var s := _make_surface()
	var result: Dictionary = {}

	# Fire the event after one frame so wait_for starts first.
	get_tree().process_frame.connect(func() -> void:
		s.emit_event("event:player:land", {"speed": 7.0})
	, CONNECT_ONE_SHOT)

	result = await s.wait_for("event:player:land")
	assert_eq(result.get("speed", 0.0), 7.0, "wait_for should return event payload")
	await _cleanup(s)

# UT-CUT11-09: Room handler fires before one-shot handlers (priority order).
func test_09_room_fires_before_oneshot() -> void:
	var s := _make_surface()
	var order: Array[String] = []
	s.set_room_handler(func(_n: String, _p: Dictionary) -> void: order.append("room"))
	s.on_event("event:z:z", func(_p: Dictionary) -> void: order.append("oneshot"), true)
	s.emit_event("event:z:z")
	assert_eq(order.size(), 2, "Both handlers should fire")
	assert_eq(order[0], "room", "Room handler should fire first")
	assert_eq(order[1], "oneshot", "One-shot should fire second")
	await _cleanup(s)

# UT-CUT11-10: emit_event with no handlers and no room handler is a no-op (no crash).
func test_10_emit_no_handlers_noop() -> void:
	var s := _make_surface()
	s.emit_event("event:unknown:event", {"x": 1})
	assert_true(true, "emit with no handlers should not crash")
	await _cleanup(s)
