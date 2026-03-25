## Base class for all AGS3D test suites.
##
## Each test class extends this. Methods named test_*() are auto-discovered
## and run by run_tests.gd. setUp/tearDown run around each test.
class_name TestBase
extends RefCounted

const C_GREEN  := "\u001b[32m"
const C_RED    := "\u001b[31m"
const C_DIM    := "\u001b[2m"
const C_RESET  := "\u001b[0m"

var _pass_count: int = 0
var _fail_count: int = 0
var _failures: Array[String] = []
var _current_test: String = ""
var _current_test_failures: Array[String] = []
var _tree: SceneTree = null

## Override to return a human-readable name for this suite.
func suite_name() -> String:
	return get_script().resource_path.get_file().get_basename()

## Lifecycle hooks — override as needed.
func setUpSuite() -> void: pass
func tearDownSuite() -> void: pass
func setUp() -> void: pass
func tearDown() -> void: pass

## Attach p_node to the real scene tree so viewport-dependent nodes
## (e.g. NavigationAgent3D, CharacterBody3D) initialise without errors.
## Works correctly only when called from _run_tests() (deferred from _init()),
## at which point SceneTree.initialize() has already run and root.is_inside_tree()
## is true.  NOTIFICATION_ENTER_TREE propagates automatically to p_node and all
## children added afterwards.  Free the returned node in teardown — Godot
## propagates EXIT_TREE to all descendants automatically.
## Requires _tree to be set by run_tests.gd before run_suite() is called.
func add_to_tree(p_node: Node) -> Node:
	_tree.root.add_child(p_node)
	return p_node

# ─── Assert helpers ───────────────────────────────────────────────────────────

func assert_eq(actual, expected, msg: String = "") -> void:
	if actual != expected:
		_fail("assert_eq failed: expected %s, got %s. %s" % [expected, actual, msg])
	else:
		_pass()

func assert_ne(actual, expected, msg: String = "") -> void:
	if actual == expected:
		_fail("assert_ne failed: expected not %s. %s" % [expected, msg])
	else:
		_pass()

func assert_true(condition: bool, msg: String = "") -> void:
	if not condition:
		_fail("assert_true failed. %s" % msg)
	else:
		_pass()

func assert_false(condition: bool, msg: String = "") -> void:
	if condition:
		_fail("assert_false failed. %s" % msg)
	else:
		_pass()

func assert_not_null(value, msg: String = "") -> void:
	if value == null:
		_fail("assert_not_null failed — got null. %s" % msg)
	else:
		_pass()

func assert_null(value, msg: String = "") -> void:
	if value != null:
		_fail("assert_null failed — expected null, got %s. %s" % [value, msg])
	else:
		_pass()

func assert_no_crash(callable: Callable, msg: String = "") -> void:
	callable.call()
	_pass()  # reaching here means no crash

## Explicitly mark a test as failed with a message.
func fail(msg: String) -> void:
	_fail(msg)

## Explicitly mark a test as passed (for coroutine-based tests).
func pass_test() -> void:
	_pass()

# ─── Internal ─────────────────────────────────────────────────────────────────

func _pass() -> void:
	_pass_count += 1

func _fail(msg: String) -> void:
	_fail_count += 1
	var entry := "  %sFAIL%s [%s::%s] %s" % [C_RED, C_RESET, suite_name(), _current_test, msg]
	_failures.append(entry)
	_current_test_failures.append(msg)

## Run all test_* methods. Called by run_tests.gd.
func run_suite() -> Dictionary:
	setUpSuite()

	var methods := []
	for m in get_method_list():
		if m["name"].begins_with("test_"):
			methods.append(m["name"])
	methods.sort()

	for method in methods:
		_current_test = method
		_current_test_failures = []
		var fail_before := _fail_count
		setUp()
		call(method)
		tearDown()

		if _fail_count == fail_before:
			print("  %s✓%s %s" % [C_GREEN, C_RESET, method])
		else:
			print("  %s✗%s %s" % [C_RED, C_RESET, method])
			for msg in _current_test_failures:
				print("    %s%s%s" % [C_DIM, msg, C_RESET])

	tearDownSuite()

	return {
		"suite": suite_name(),
		"pass": _pass_count,
		"fail": _fail_count,
		"failures": _failures,
	}
