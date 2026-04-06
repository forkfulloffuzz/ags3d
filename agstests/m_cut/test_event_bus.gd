## T-CUT10 — AGSEventBus synchronous pub/sub tests.
extends "res://utils/test_base.gd"

func suite_name() -> String:
	return "M-CUT: EventBus"

# UT-CUT10-01: AGSEventBus singleton is accessible.
func test_01_singleton_accessible() -> void:
	assert_not_null(AGSEventBus, "AGSEventBus singleton is null")

# UT-CUT10-02: Subscribing and emitting calls the callable synchronously.
func test_02_emit_calls_subscriber() -> void:
	AGSEventBus.clear()
	var called := false
	var handler := func(_p: Dictionary) -> void:
		called = true
	AGSEventBus.subscribe("event:test:fire", handler)
	AGSEventBus.emit_event("event:test:fire", {})
	assert_true(called, "Subscriber was not called after emit_event")
	AGSEventBus.clear()

# UT-CUT10-03: Payload dictionary is passed to the subscriber.
func test_03_payload_passed() -> void:
	AGSEventBus.clear()
	var received: Dictionary = {}
	var handler := func(p: Dictionary) -> void:
		received = p
	AGSEventBus.subscribe("event:test:payload", handler)
	AGSEventBus.emit_event("event:test:payload", {"x": 42})
	assert_eq(received.get("x", -1), 42, "Payload value not received correctly")
	AGSEventBus.clear()

# UT-CUT10-04: Multiple subscribers all fire in subscription order.
func test_04_multiple_subscribers_fire_in_order() -> void:
	AGSEventBus.clear()
	var order: Array[int] = []
	AGSEventBus.subscribe("event:test:order", func(_p: Dictionary) -> void: order.append(1))
	AGSEventBus.subscribe("event:test:order", func(_p: Dictionary) -> void: order.append(2))
	AGSEventBus.subscribe("event:test:order", func(_p: Dictionary) -> void: order.append(3))
	AGSEventBus.emit_event("event:test:order", {})
	assert_eq(order.size(), 3, "Expected 3 calls, got %d" % order.size())
	assert_eq(order[0], 1, "First subscriber should be called first")
	assert_eq(order[2], 3, "Third subscriber should be called third")
	AGSEventBus.clear()

# UT-CUT10-05: Emitting with no subscribers does not crash.
func test_05_emit_no_subscribers_no_crash() -> void:
	AGSEventBus.clear()
	AGSEventBus.emit_event("event:test:empty", {})
	assert_true(true, "emit_event with no subscribers should not crash")

# UT-CUT10-06: unsubscribe removes the callable; it is not called on next emit.
func test_06_unsubscribe_removes_callable() -> void:
	AGSEventBus.clear()
	var called := false
	var handler := func(_p: Dictionary) -> void:
		called = true
	AGSEventBus.subscribe("event:test:unsub", handler)
	AGSEventBus.unsubscribe("event:test:unsub", handler)
	AGSEventBus.emit_event("event:test:unsub", {})
	assert_false(called, "Unsubscribed callable should not be invoked")
	AGSEventBus.clear()

# UT-CUT10-07: subscriber_count returns correct count before and after subscribe/unsubscribe.
func test_07_subscriber_count() -> void:
	AGSEventBus.clear()
	assert_eq(AGSEventBus.subscriber_count("event:test:count"), 0, "Count should be 0 before any subscribe")
	var h1 := func(_p: Dictionary) -> void: pass
	var h2 := func(_p: Dictionary) -> void: pass
	AGSEventBus.subscribe("event:test:count", h1)
	AGSEventBus.subscribe("event:test:count", h2)
	assert_eq(AGSEventBus.subscriber_count("event:test:count"), 2, "Count should be 2 after 2 subscribes")
	AGSEventBus.unsubscribe("event:test:count", h1)
	assert_eq(AGSEventBus.subscriber_count("event:test:count"), 1, "Count should be 1 after 1 unsubscribe")
	AGSEventBus.clear()

# UT-CUT10-08: clear removes all subscribers across all events.
func test_08_clear_removes_all() -> void:
	AGSEventBus.clear()
	AGSEventBus.subscribe("event:a", func(_p: Dictionary) -> void: pass)
	AGSEventBus.subscribe("event:b", func(_p: Dictionary) -> void: pass)
	AGSEventBus.clear()
	assert_eq(AGSEventBus.subscriber_count("event:a"), 0, "event:a should have 0 subscribers after clear")
	assert_eq(AGSEventBus.subscriber_count("event:b"), 0, "event:b should have 0 subscribers after clear")

# UT-CUT10-09: Subscribers added during emit are NOT called in the same dispatch.
func test_09_snapshot_semantics() -> void:
	AGSEventBus.clear()
	var second_called := false
	var second_handler := func(_p: Dictionary) -> void:
		second_called = true
	var first_handler := func(_p: Dictionary) -> void:
		AGSEventBus.subscribe("event:test:snap", second_handler)
	AGSEventBus.subscribe("event:test:snap", first_handler)
	AGSEventBus.emit_event("event:test:snap", {})
	assert_false(second_called, "Subscriber added during emit should not be called in same dispatch")
	AGSEventBus.clear()
