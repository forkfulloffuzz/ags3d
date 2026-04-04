## UT-M10-30..34 — AGSItem node + AGSRuntime.get_item() (T-GS02)
##
## All tests are synchronous.
extends "res://utils/test_base.gd"

func suite_name() -> String:
	return "M10: Item"


# UT-M10-30: AGSItem has item_name, display_name, description properties.
func test_30_item_properties() -> void:
	var item := AGSItem.new()
	item.item_name = "rusty_key"
	item.display_name = "Rusty Key"
	item.description = "An old iron key."
	assert_eq(item.item_name, "rusty_key", "item_name not stored")
	assert_eq(item.display_name, "Rusty Key", "display_name not stored")
	assert_eq(item.description, "An old iron key.", "description not stored")
	item.free()


# UT-M10-31: AGSItem registers with AGSRuntime when added to tree.
func test_31_item_registers_on_ready() -> void:
	var rt := AGSRuntime.new()
	var item := AGSItem.new()
	item.item_name = "test_key_31"
	# Manually call notification to simulate NOTIFICATION_READY without SceneTree.
	# AGSItem._notification uses AGSRuntime::get_singleton(); we must set rt as singleton.
	# Since the singleton is set in constructor, create rt first, then notify item.
	item.notification(Node.NOTIFICATION_READY)
	assert_not_null(rt.get_item("test_key_31"), "item should be registered after NOTIFICATION_READY")
	item.free()
	rt.free()


# UT-M10-32: AGSRuntime.get_item returns null for unknown name.
func test_32_get_item_unknown_returns_null() -> void:
	var rt := AGSRuntime.new()
	var result = rt.get_item("no_such_item")
	assert_true(result == null, "get_item for unknown name should return null")
	rt.free()


# UT-M10-33: AGSItem unregisters on EXIT_TREE.
func test_33_item_unregisters_on_exit() -> void:
	var rt := AGSRuntime.new()
	var item := AGSItem.new()
	item.item_name = "test_key_33"
	item.notification(Node.NOTIFICATION_READY)
	assert_not_null(rt.get_item("test_key_33"), "item should be registered")
	item.notification(Node.NOTIFICATION_EXIT_TREE)
	assert_true(rt.get_item("test_key_33") == null, "item should be unregistered after EXIT_TREE")
	item.free()
	rt.free()


# UT-M10-34: Multiple items can be registered simultaneously.
func test_34_multiple_items() -> void:
	var rt := AGSRuntime.new()
	var a := AGSItem.new()
	a.item_name = "key"
	var b := AGSItem.new()
	b.item_name = "coin"
	a.notification(Node.NOTIFICATION_READY)
	b.notification(Node.NOTIFICATION_READY)
	assert_not_null(rt.get_item("key"), "key should be registered")
	assert_not_null(rt.get_item("coin"), "coin should be registered")
	a.free()
	b.free()
	rt.free()
