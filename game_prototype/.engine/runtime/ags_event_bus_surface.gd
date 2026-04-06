## ags_event_bus_surface.gd — AGS3D Event Bus AGS-Spirit Surface (T-CUT11)
##
## Add as an AutoLoad named AGSEventBusSurface.
##
## Thin GDScript layer over the AGSEventBus C++ singleton that adds:
##
##   • Room hook — the active room's on_event() receives every dispatched event.
##   • One-shot subscriptions — handlers that auto-remove after first call.
##   • Coroutine waiting — `await wait_for("event:player:land")`.
##
## Priority order (per AGS-spirit spec):
##   character handlers → room on_event() → cutscene OnEvent → dialogue
##
##   Character handlers and cutscene OnEvent handlers subscribe directly via
##   AGSEventBus. The room hook is managed here and fires as part of dispatch.
##
## AGS-spirit API (compiled from .agscript):
##   AGSEventBusSurface.emit_event("event:player:land")
##   AGSEventBusSurface.on_event("event:player:land", handler)         # persistent
##   AGSEventBusSurface.on_event("event:player:land", handler, true)   # one-shot
##   var payload = await AGSEventBusSurface.wait_for("event:player:land")
##
## Room script:
##   func _ready(): AGSEventBusSurface.set_room_handler(_on_event)
##   func _on_room_exit(): AGSEventBusSurface.clear_room_handler()
##   func _on_event(name, payload): ...

extends Node

# ---------------------------------------------------------------------------
# Signals
# ---------------------------------------------------------------------------

## Broadcast for every event dispatched through this surface.
## Used internally by wait_for(); also available for monitoring.
signal event_dispatched(event_name: String, payload: Dictionary)

# ---------------------------------------------------------------------------
# Internal state
# ---------------------------------------------------------------------------

## Active room-level handler. Receives every event while set.
var _room_handler: Callable = Callable()

## One-shot handlers. Key: event_name, Value: Array[Callable].
## Handlers are consumed (removed) on first matching dispatch.
var _one_shot: Dictionary = {}

## Reference to the AGSEventBus C++ singleton.
var _bus: Object = null

# ---------------------------------------------------------------------------
# Lifecycle
# ---------------------------------------------------------------------------

func _ready() -> void:
	_bus = _resolve_bus()

func _resolve_bus() -> Object:
	if Engine.has_singleton("AGSEventBus"):
		return Engine.get_singleton("AGSEventBus")
	# Fallback for headless tests: look for it in the scene tree.
	if get_tree() != null:
		return get_tree().root.get_node_or_null("/root/AGSEventBus")
	return null

# ---------------------------------------------------------------------------
# Public API — AGS-spirit surface
# ---------------------------------------------------------------------------

## Emit an event. All C++ subscribers fire synchronously first, then the
## GDScript surface routes room hooks, one-shot handlers, and event_dispatched.
func emit_event(event_name: String, payload: Dictionary = {}) -> void:
	# Persistent subscribers on the C++ bus fire here.
	if _bus != null:
		_bus.emit_event(event_name, payload)
	# GDScript routing (room hook, one-shots, signal).
	_dispatch_surface(event_name, payload)

## Register a handler for event_name.
## once=false (default) — persistent; remove with remove_handler().
## once=true — fires once then auto-removes.
func on_event(event_name: String, handler: Callable, once: bool = false) -> void:
	if once:
		if not _one_shot.has(event_name):
			_one_shot[event_name] = [] as Array
		(_one_shot[event_name] as Array).append(handler)
	else:
		if _bus != null:
			_bus.subscribe(event_name, handler)

## Remove a persistent handler previously registered via on_event(once=false).
func remove_handler(event_name: String, handler: Callable) -> void:
	if _bus != null:
		_bus.unsubscribe(event_name, handler)

## Suspend the calling coroutine until event_name fires.
## Returns the payload Dictionary passed with the event.
##
## Usage:
##   var payload = await AGSEventBusSurface.wait_for("event:player:land")
func wait_for(event_name: String) -> Dictionary:
	var received: bool = false
	var result: Dictionary = {}

	var handler := func(p: Dictionary) -> void:
		result = p
		received = true

	on_event(event_name, handler, true)

	while not received:
		await get_tree().process_frame

	return result

# ---------------------------------------------------------------------------
# Room hook
# ---------------------------------------------------------------------------

## Set the active room's event handler.
## Call from the room script's _ready() (or room-enter hook).
## Signature: func handler(event_name: String, payload: Dictionary) -> void
func set_room_handler(handler: Callable) -> void:
	_room_handler = handler

## Clear the room handler (call on room exit).
func clear_room_handler() -> void:
	_room_handler = Callable()

# ---------------------------------------------------------------------------
# Internal dispatch
# ---------------------------------------------------------------------------

func _dispatch_surface(event_name: String, payload: Dictionary) -> void:
	# 1. Room hook.
	if _room_handler.is_valid():
		_room_handler.call(event_name, payload)

	# 2. One-shot handlers for this event.
	if _one_shot.has(event_name):
		var handlers: Array = (_one_shot[event_name] as Array).duplicate()
		_one_shot.erase(event_name)
		for h: Callable in handlers:
			h.call(payload)

	# 3. Broadcast signal (used by wait_for and monitoring).
	event_dispatched.emit(event_name, payload)
