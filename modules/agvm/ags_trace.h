#pragma once

// AGS_TRACE(p_type, p_func, p_msg) — emit one trace line when runtime tracing is on.
// p_type and p_func must be C string literals (e.g. "AGSCharacter", "walk_to").
// p_msg must be a Godot String expression; use vformat() for formatted messages.
//
// Example:
//   AGS_TRACE("AGSCharacter", "walk_to", vformat("point=%s", p_point_name))
//   // prints: [AGS/AGSCharacter::walk_to] point=door_left

#include "ags_runtime.h"
#include "core/string/print_string.h"

#define AGS_TRACE(p_type, p_func, p_msg)                                              \
	if (AGSRuntime::is_trace_enabled()) {                                              \
		print_line(String("[AGS/") + (p_type) + "::" + (p_func) + "] " + (p_msg));   \
	}
