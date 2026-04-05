## UT-M10-30..34 — AGSItem node + AGSRuntime.get_item() (T-GS02)
##
## All tests are synchronous.
##
## AGSRuntime is an engine singleton — never call AGSRuntime.new().
## Tests use unique per-test item name prefixes (t31_, t33_, …) and
## manually trigger EXIT_TREE before free() so the item is unregistered.
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


# UT-M10-31: AGSItem registers with AGSRuntime when NOTIFICATION_READY fires.
func test_31_item_registers_on_ready() -> void:
	var item := AGSItem.new()
	item.item_name = "t31_test_key"
	item.notification(Node.NOTIFICATION_READY)
	assert_not_null(AGSRuntime.get_item("t31_test_key"),
			"item should be registered after NOTIFICATION_READY")
	item.notification(Node.NOTIFICATION_EXIT_TREE)
	item.free()


# UT-M10-32: AGSRuntime.get_item returns null for unknown name.
func test_32_get_item_unknown_returns_null() -> void:
	var result = AGSRuntime.get_item("t32_no_such_item")
	assert_true(result == null, "get_item for unknown name should return null")


# UT-M10-33: AGSItem unregisters on EXIT_TREE.
func test_33_item_unregisters_on_exit() -> void:
	var item := AGSItem.new()
	item.item_name = "t33_test_key"
	item.notification(Node.NOTIFICATION_READY)
	assert_not_null(AGSRuntime.get_item("t33_test_key"), "item should be registered")
	item.notification(Node.NOTIFICATION_EXIT_TREE)
	assert_true(AGSRuntime.get_item("t33_test_key") == null,
			"item should be unregistered after EXIT_TREE")
	item.free()


# UT-M10-34: Multiple items can be registered simultaneously.
func test_34_multiple_items() -> void:
	var a := AGSItem.new()
	a.item_name = "t34_key"
	var b := AGSItem.new()
	b.item_name = "t34_coin"
	a.notification(Node.NOTIFICATION_READY)
	b.notification(Node.NOTIFICATION_READY)
	assert_not_null(AGSRuntime.get_item("t34_key"), "key should be registered")
	assert_not_null(AGSRuntime.get_item("t34_coin"), "coin should be registered")
	a.notification(Node.NOTIFICATION_EXIT_TREE)
	b.notification(Node.NOTIFICATION_EXIT_TREE)
	a.free()
	b.free()
