## Base class for all AGS3D test suites.
##
## Each test class extends this. Methods named test_*() are auto-discovered
## and run by run_tests.gd. setUp/tearDown run around each test.
class_name TestBase
extends RefCounted

var _pass_count: int = 0
var _fail_count: int = 0
var _failures: Array[String] = []
var _current_test: String = ""

## Override to return a human-readable name for this suite.
func suite_name() -> String:
	return get_script().resource_path.get_file().get_basename()

## Lifecycle hooks — override as needed.
func setUpSuite() -> void: pass
func tearDownSuite() -> void: pass
func setUp() -> void: pass
func tearDown() -> void: pass

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
	var entry := "  FAIL [%s::%s] %s" % [suite_name(), _current_test, msg]
	_failures.append(entry)
	print(entry)

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
		setUp()
		call(method)
		tearDown()

	tearDownSuite()

	return {
		"suite": suite_name(),
		"pass": _pass_count,
		"fail": _fail_count,
		"failures": _failures,
	}
