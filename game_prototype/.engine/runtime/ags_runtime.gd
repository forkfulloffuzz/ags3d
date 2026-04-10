## ags_runtime.gd — Central AGS3D runtime API (T-GS10)
##
## Add as an AutoLoad named AGSRuntime. This is the central hub for all
## game-level API calls emitted from .agscript.
##
## Emitted signals (consumed by other AutoLoads):
##   room_change_requested(room_name, spawn_name) — consumed by AGSRoomManager
##   play_music_requested(track_name)           — consumed by AGSAudio
##   stop_music_requested()                     — consumed by AGSAudio
##   play_sound_requested(sfx_name)             — consumed by AGSAudio
##
## Player control signal:
##   player_control_changed(enabled: bool)       — consumed by input handlers
##
## Methods consumed by generated GDScript (emitter output):
##   load_room(room_name: String, spawn_name: String = "") → emits room_change_requested
##   set_player_control(enabled: bool)           → emits player_control_changed
##   play_music(track_name: String)             → emits play_music_requested
##   stop_music()                               → emits stop_music_requested
##   play_sound(sfx_name: String)               → emits play_sound_requested
##   set_status_text(text: String)
##   set_active_verb(verb: String)
##   get_active_verb() -> String
##   hide_room_item(name: String)
##   show_room_item(name: String)
##   save_game(slot: int)
##   load_game(slot: int)
##   game_saved(slot: int) -> bool
##   get_character(name: String) -> Node
##   add_inventory(character_name: String, item_name: String)
##   lose_inventory(character_name: String, item_name: String)
##   has_inventory(character_name: String, item_name: String) -> bool
extends Node

## ---------------------------------------------------------------------------
## Signals
## ---------------------------------------------------------------------------

signal room_change_requested(room_name: String, spawn_name: String)
signal play_music_requested(track_name: String)
signal stop_music_requested()
signal play_sound_requested(sfx_name: String)
signal player_control_changed(enabled: bool)

## ---------------------------------------------------------------------------
## State
## ---------------------------------------------------------------------------

var _active_verb: String = ""
var _player_control_enabled: bool = true

## Cached character nodes by name (populated on first get_character call).
var _character_cache: Dictionary = {}

## ---------------------------------------------------------------------------
## Room transitions
## ---------------------------------------------------------------------------

func load_room(room_name: String, spawn_name: String = "") -> void:
	room_change_requested.emit(room_name, spawn_name)


## ---------------------------------------------------------------------------
## Player control
## ---------------------------------------------------------------------------

func set_player_control(enabled: bool) -> void:
	_player_control_enabled = enabled
	player_control_changed.emit(enabled)


## ---------------------------------------------------------------------------
## Audio
## ---------------------------------------------------------------------------

func play_music(track_name: String) -> void:
	play_music_requested.emit(track_name)


func stop_music() -> void:
	stop_music_requested.emit()


func play_sound(sfx_name: String) -> void:
	play_sound_requested.emit(sfx_name)


## ---------------------------------------------------------------------------
## HUD / GUI
## ---------------------------------------------------------------------------

var _status_text_node: Label = null

func set_status_text(text: String) -> void:
	if _status_text_node == null:
		var cl := get_tree().get_first_node_in_group("AGS_GUI")
		if cl == null:
			return
		for ch in cl.get_children():
			if ch.name.begins_with("StatusLine"):
				_status_text_node = ch
				break
	if _status_text_node != null:
		_status_text_node.text = text


func set_active_verb(verb: String) -> void:
	_active_verb = verb


func get_active_verb() -> String:
	return _active_verb


## ---------------------------------------------------------------------------
## Room items
## ---------------------------------------------------------------------------

func hide_room_item(name: String) -> void:
	var room := _get_current_room()
	if room == null:
		return
	var item := room.find_child(name, true, false)
	if item != null:
		item.visible = false
	else:
		print("[AGSRuntime] hide_room_item: '%s' not found in room" % name)


func show_room_item(name: String) -> void:
	var room := _get_current_room()
	if room == null:
		return
	var item := room.find_child(name, true, false)
	if item != null:
		item.visible = true
	else:
		print("[AGSRuntime] show_room_item: '%s' not found in room" % name)


func _get_current_room() -> Node:
	for ch in get_tree().get_children():
		if ch.has_method("is_room_root"):
			return ch
	return null


## ---------------------------------------------------------------------------
## Characters
## ---------------------------------------------------------------------------

func get_character(name: String) -> Node:
	if _character_cache.has(name):
		return _character_cache[name]
	var node := get_tree().get_first_node_in_group("AGSCharacter")
	while node != null:
		if node.character_name == name:
			_character_cache[name] = node
			return node
		node = node.get_parent().find_child("*", true, false)
	return null


## ---------------------------------------------------------------------------
## Inventory
## ---------------------------------------------------------------------------

var _inventories: Dictionary = {}

func _ensure_inventory(character_name: String) -> Array:
	if not _inventories.has(character_name):
		_inventories[character_name] = []
	return _inventories[character_name]


func add_inventory(character_name: String, item_name: String) -> void:
	var inv := _ensure_inventory(character_name)
	if not inv.has(item_name):
		inv.append(item_name)


func lose_inventory(character_name: String, item_name: String) -> void:
	var inv := _ensure_inventory(character_name)
	inv.erase(item_name)


func has_inventory(character_name: String, item_name: String) -> bool:
	var inv := _ensure_inventory(character_name)
	return inv.has(item_name)


## ---------------------------------------------------------------------------
## Save / load
## ---------------------------------------------------------------------------

const SAVE_DIR := "user://saves/"
const SAVE_EXT := ".json"

func save_game(slot: int) -> bool:
	var dir := DirAccess.open(SAVE_DIR)
	if dir == null:
		DirAccess.make_dir_recursive_absolute(SAVE_DIR)
	var path := SAVE_DIR + "save_%d%s" % [slot, SAVE_EXT]
	var data := {
		"slot": slot,
		"inventories": _inventories,
	}
	var f := FileAccess.open(path, FileAccess.WRITE)
	if f == null:
		push_error("[AGSRuntime] save_game: failed to open %s" % path)
		return false
	f.store_string(JSON.stringify(data, "\t"))
	f.close()
	return true


func load_game(slot: int) -> bool:
	var path := SAVE_DIR + "save_%d%s" % [slot, SAVE_EXT]
	if not FileAccess.file_exists(path):
		push_error("[AGSRuntime] load_game: no save file at slot %d" % slot)
		return false
	var f := FileAccess.open(path, FileAccess.READ)
	if f == null:
		push_error("[AGSRuntime] load_game: failed to open %s" % path)
		return false
	var json_str := f.get_as_text()
	f.close()
	var json := JSON.new()
	if json.parse(json_str) != OK:
		push_error("[AGSRuntime] load_game: JSON parse error")
		return false
	var data: Dictionary = json.get_data()
	if data.has("inventories"):
		_inventories = data["inventories"]
	return true


func game_saved(slot: int) -> bool:
	var path := SAVE_DIR + "save_%d%s" % [slot, SAVE_EXT]
	return FileAccess.file_exists(path)
