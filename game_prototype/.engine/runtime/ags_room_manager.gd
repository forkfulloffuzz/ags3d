## ags_room_manager.gd — AGSRuntime room transition handler (T-GS10)
##
## Add this as an AutoLoad (or child of the game root) to handle room changes
## triggered by AGSRuntime.load_room("room_name").
##
## When room_change_requested fires:
##   1. Derives the .tscn path from the room name
##   2. Loads and instances the new room scene
##   3. Swaps it into the scene tree, freeing the old room
##   4. Characters are NOT moved — they are global AutoLoads and persist

extends Node

## Map room names to res:// .tscn paths. Populated at start by scanning for
## .tscn files whose root node has room_name matching the key.
## Falls back to the default convention: rooms/<name>/<name>.tscn
var _room_paths: Dictionary = {}

var _current_room: Node = null


func _ready() -> void:
	var runtime := Engine.get_singleton("AGSRuntime")
	print("[AGS/RoomManager] _ready: runtime found=", runtime != null)
	if runtime:
		runtime.connect("room_change_requested", _on_room_change_requested)
		print("[AGS/RoomManager] connected to room_change_requested")
	_scan_rooms()
	print("[AGS/RoomManager] known rooms: ", _room_paths.keys())


## Register a known room manually (useful for tests or non-standard paths).
func register_room_path(room_name: String, res_path: String) -> void:
	_room_paths[room_name] = res_path


## Change to [param room_name], freeing the current room scene.
func _on_room_change_requested(room_name: String) -> void:
	print("[AGS/RoomManager] room_change_requested: room_name=", room_name)
	var path: String = _room_paths.get(room_name, "")
	if path.is_empty():
		path = "res://rooms/%s/%s.tscn" % [room_name, room_name]
	print("[AGS/RoomManager] resolved path=", path, " exists=", ResourceLoader.exists(path))

	if not ResourceLoader.exists(path):
		push_error("[AGS] RoomManager: scene not found for room '%s' at '%s'" % [room_name, path])
		return

	var packed: PackedScene = load(path)
	if not packed:
		push_error("[AGS] RoomManager: failed to load '%s'" % path)
		return

	if _current_room and is_instance_valid(_current_room):
		print("[AGS/RoomManager] freeing current room: ", _current_room.name)
		_current_room.get_parent().remove_child(_current_room)
		_current_room.queue_free()
		_current_room = null

	var new_room: Node = packed.instantiate()
	print("[AGS/RoomManager] adding new room: ", new_room.name)
	get_parent().add_child(new_room)
	_current_room = new_room


func _scan_rooms() -> void:
	_scan_dir("res://rooms")


func _scan_dir(dir_path: String) -> void:
	var da := DirAccess.open(dir_path)
	if not da:
		return
	da.list_dir_begin()
	var entry: String = da.get_next()
	while entry != "":
		var full: String = dir_path.path_join(entry)
		if da.current_is_dir():
			_scan_dir(full)
		elif entry.ends_with(".tscn"):
			_try_register(full)
		entry = da.get_next()
	da.list_dir_end()


func _try_register(res_path: String) -> void:
	# Quick heuristic: if the filename stem matches a parent directory name,
	# treat it as a room scene and register by stem.
	var stem: String = res_path.get_file().get_basename()
	var parent_dir: String = res_path.get_base_dir().get_file()
	if stem == parent_dir:
		_room_paths[stem] = res_path
