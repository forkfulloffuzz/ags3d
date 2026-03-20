## Collects results from all test suites and prints a summary.
class_name Reporter
extends RefCounted

const C_GREEN  := "\u001b[32m"
const C_RED    := "\u001b[31m"
const C_BOLD   := "\u001b[1m"
const C_RESET  := "\u001b[0m"

var _total_pass: int = 0
var _total_fail: int = 0
var _suite_results: Array[Dictionary] = []

func record(result: Dictionary) -> void:
	_suite_results.append(result)
	_total_pass += result["pass"]
	_total_fail += result["fail"]

	var ok: bool = result["fail"] == 0
	var color: String = C_GREEN if ok else C_RED
	var status: String = "PASS" if ok else "FAIL"
	print("%s[%s]%s %s — %d passed, %d failed" % [
		color, status, C_RESET, result["suite"], result["pass"], result["fail"]
	])

func print_summary() -> void:
	print("")
	print("=" .repeat(50))
	print("%sAGS3D Test Results%s" % [C_BOLD, C_RESET])
	print("  Suites : %d" % _suite_results.size())
	print("  %sPassed : %d%s" % [C_GREEN, _total_pass, C_RESET])
	if _total_fail > 0:
		print("  %sFailed : %d%s" % [C_RED, _total_fail, C_RESET])
	else:
		print("  Failed : %d" % _total_fail)
	print("=" .repeat(50))
	if _total_fail > 0:
		print("%sFAILURES:%s" % [C_RED, C_RESET])
		for r in _suite_results:
			for f in r["failures"]:
				print(f)

func all_passed() -> bool:
	return _total_fail == 0
