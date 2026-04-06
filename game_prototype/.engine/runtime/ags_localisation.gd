## ags_localisation.gd — AGS3D Localisation Runtime (T-DLG17)
##
## Add as an AutoLoad named AGSLocalisation.
##
## Loads translation tables from .engine/generated/locale/<code>.json at
## startup (or on-demand via set_locale). Provides string lookup with
## fallback chain and RTL layout flag.
##
## AGS-spirit API:
##   Game.SetLocale("fr")          → switches active locale
##   AGSLocalisation.get(loc_key)  → translated string (or fallback)
##
## Generated locale file format (.engine/generated/locale/<code>.json):
##   {
##     "locale":   "fr",
##     "rtl":      false,
##     "strings":  { "guard_greeting:0:abc123": "Halte-là !" }
##   }
##
## Fallback chain (set in game.agp [localisation]):
##   active locale → fallback locales in order → base locale
##   If none found, the raw source text passed to get() is returned.
##
## Dialogue integration:
##   AGSDialogue calls AGSLocalisation.get(loc_key, fallback_text) to
##   resolve each line. On locale switch mid-conversation, AGSDialogue
##   restarts the current node via _on_locale_changed().

extends Node

# ---------------------------------------------------------------------------
# Signals
# ---------------------------------------------------------------------------

## Emitted after the active locale changes.
signal locale_changed(new_locale: String)

# ---------------------------------------------------------------------------
# Configuration
# ---------------------------------------------------------------------------

## Directory where generated locale JSON files live (relative to res://).
@export var locale_dir: String = ".engine/generated/locale/"

## Base locale — always loaded; used as ultimate fallback.
@export var base_locale: String = "en"

## Fallback chain: if a key is not in the active locale, each locale in
## this list is tried in order before falling back to base_locale.
@export var fallback_chain: Array[String] = []

# ---------------------------------------------------------------------------
# State
# ---------------------------------------------------------------------------

var _active_locale: String = ""
var _rtl: bool = false

## Loaded string tables. Key: locale code, Value: Dictionary(loc_key→str).
var _tables: Dictionary = {}  # String → Dictionary

# ---------------------------------------------------------------------------
# Lifecycle
# ---------------------------------------------------------------------------

func _ready() -> void:
	_active_locale = base_locale
	_load_locale(base_locale)
	# Connect to AGSDialogue so we can restart current node on locale change.
	var dlg: Node = _get_autoload("AGSDialogue")
	if dlg != null:
		locale_changed.connect(dlg._on_locale_changed)

func _get_autoload(name: String) -> Node:
	if get_tree() == null:
		return null
	return get_tree().root.get_node_or_null("/root/" + name)

# ---------------------------------------------------------------------------
# Public API — AGS-spirit surface
# ---------------------------------------------------------------------------

## Switch the active locale. Loads the locale file if not already cached.
## Emits locale_changed(new_locale) after the switch.
## AGSDialogue will restart the current node when it receives this signal.
func set_locale(code: String) -> void:
	if code == _active_locale:
		return
	_load_locale(code)
	_active_locale = code
	_rtl = _get_rtl(code)
	locale_changed.emit(code)

## Returns the currently active locale code.
func active_locale() -> String:
	return _active_locale

## Returns true if the active locale uses right-to-left layout.
func is_rtl() -> bool:
	return _rtl

## Look up loc_key in the active locale (then fallback chain, then base).
## Returns fallback_text if the key is not found in any loaded locale.
func get_string(loc_key: String, fallback_text: String = "") -> String:
	# Try active locale first.
	var result := _lookup(_active_locale, loc_key)
	if result != "":
		return result

	# Try fallback chain.
	for code: String in fallback_chain:
		_load_locale(code)
		result = _lookup(code, loc_key)
		if result != "":
			return result

	# Try base locale.
	if _active_locale != base_locale:
		result = _lookup(base_locale, loc_key)
		if result != "":
			return result

	# Return source text as last resort.
	return fallback_text if fallback_text != "" else loc_key

# Alias — used by AGSDialogue: AGSLocalisation.get(loc_key, fallback)
func get(loc_key: String, fallback_text: String = "") -> String:
	return get_string(loc_key, fallback_text)

# ---------------------------------------------------------------------------
# Internal helpers
# ---------------------------------------------------------------------------

func _lookup(code: String, key: String) -> String:
	if not _tables.has(code):
		return ""
	var table: Dictionary = _tables[code] as Dictionary
	return table.get(key, "") as String

func _get_rtl(code: String) -> bool:
	# RTL flag is stored alongside the strings table under key "__rtl__".
	if not _tables.has(code):
		return false
	return (_tables[code] as Dictionary).get("__rtl__", false) as bool

func _load_locale(code: String) -> void:
	if _tables.has(code):
		return  # Already loaded.
	var path := "res://" + locale_dir + code + ".json"
	if not FileAccess.file_exists(path):
		# Not found — register empty table so we don't retry every call.
		_tables[code] = {}
		return

	var file := FileAccess.open(path, FileAccess.READ)
	if file == null:
		_tables[code] = {}
		return

	var json := JSON.new()
	var parse_err := json.parse(file.get_as_text())
	file.close()
	if parse_err != OK:
		push_warning("AGSLocalisation: failed to parse locale file %s" % path)
		_tables[code] = {}
		return

	var data = json.get_data()
	if data is Dictionary:
		var strings: Dictionary = (data as Dictionary).get("strings", {}) as Dictionary
		# Store RTL flag alongside strings using a sentinel key.
		strings["__rtl__"] = (data as Dictionary).get("rtl", false)
		_tables[code] = strings
	else:
		_tables[code] = {}
