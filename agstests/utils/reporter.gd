## Collects results from all test suites and prints a summary.
class_name Reporter
extends RefCounted

var _total_pass: int = 0
var _total_fail: int = 0
var _suite_results: Array[Dictionary] = []

func record(result: Dictionary) -> void:
	_suite_results.append(result)
	_total_pass += result["pass"]
	_total_fail += result["fail"]

	var status := "PASS" if result["fail"] == 0 else "FAIL"
	print("[%s] %s — %d passed, %d failed" % [
		status, result["suite"], result["pass"], result["fail"]
	])

func print_summary() -> void:
	print("")
	print("=" .repeat(50))
	print("AGS3D Test Results")
	print("  Suites : %d" % _suite_results.size())
	print("  Passed : %d" % _total_pass)
	print("  Failed : %d" % _total_fail)
	print("=" .repeat(50))
	if _total_fail > 0:
		print("FAILURES:")
		for r in _suite_results:
			for f in r["failures"]:
				print(f)

func all_passed() -> bool:
	return _total_fail == 0
