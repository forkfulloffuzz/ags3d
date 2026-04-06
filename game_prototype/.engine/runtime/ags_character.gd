extends AGSCharacter3D

## Animation clip name for the idle state (set by ag build from .agchar animations block).
@export var anim_idle: String = ""
## Animation clip name for the walk state.
@export var anim_walk: String = ""
## Animation clip name for the talk state.
@export var anim_talk: String = ""

var _nav_agent: NavigationAgent3D
var _navigating: bool = false
var _anim_player: AnimationPlayer = null
var _anim_state: String = ""

func _ready() -> void:
	_nav_agent = NavigationAgent3D.new()
	_nav_agent.path_desired_distance = 0.5
	_nav_agent.target_desired_distance = 0.5
	add_child(_nav_agent)
	_nav_agent.navigation_finished.connect(_on_navigation_finished)

func _physics_process(_delta: float) -> void:
	if not _navigating or _nav_agent.is_navigation_finished():
		return
	var next_pos: Vector3 = _nav_agent.get_next_path_position()
	velocity = (next_pos - global_position).normalized() * move_speed
	move_and_slide()

func navigate_to(target: Vector3) -> void:
	_navigating = true
	_play_anim_state("walk")
	await get_tree().physics_frame  # Frame-0: nav map syncs on first physics frame
	# Snap target to the closest point on the active navmesh so the character
	# always moves as close as possible, even when the target is unreachable
	# (e.g. on a disconnected navmesh island). Never silently fails.
	var map_rid := get_world_3d().navigation_map
	var snapped := NavigationServer3D.map_get_closest_point(map_rid, target)
	_nav_agent.target_position = snapped

func walk_to(point_name: String) -> void:
	print("[AGS/AGSCharacter::walk_to] char=", character_name, " point=", point_name, " start")
	var room := _find_parent_room()
	if not room:
		push_error("AGSCharacter.walk_to: no parent AGSRoom found")
		return
	await navigate_to(room.get_point(point_name))
	await walk_completed
	print("[AGS/AGSCharacter::walk_to] char=", character_name, " point=", point_name, " done")

func face_to(point_name: String) -> void:
	print("[AGS/AGSCharacter::face_to] char=", character_name, " point=", point_name, " start")
	var room := _find_parent_room()
	if not room:
		push_error("AGSCharacter.face_to: no parent AGSRoom found")
		return
	var target: Vector3 = room.get_point(point_name)
	var dir: Vector3 = target - global_position
	dir.y = 0.0
	if dir.length_squared() > 0.001:
		var look_basis := Basis.looking_at(dir.normalized())
		var target_y: float = look_basis.get_euler().y
		var tween := create_tween()
		tween.tween_property(self, "rotation:y", target_y, 0.3)
		await tween.finished
	emit_signal("face_completed")
	print("[AGS/AGSCharacter::face_to] char=", character_name, " point=", point_name, " done")

## Display [param text] as dialogue above the character for [param duration] seconds,
## then emit say_completed. Awaiting this method blocks until the line finishes.
func say(text: String, duration: float = 2.0) -> void:
	print("[AGS/AGSCharacter::say] char=", character_name, " text=", text, " start")
	say_text = text
	_play_anim_state("talk")
	await get_tree().create_timer(duration).timeout
	say_text = ""
	_play_anim_state("idle")
	emit_signal("say_completed")
	print("[AGS/AGSCharacter::say] char=", character_name, " done")

## Display [param text] as a thought (same mechanic as say, different visual intent).
func think(text: String, duration: float = 2.0) -> void:
	print("[AGS/AGSCharacter::think] char=", character_name, " text=", text, " start")
	await say(text, duration)
	print("[AGS/AGSCharacter::think] char=", character_name, " done")

func _on_navigation_finished() -> void:
	_navigating = false
	velocity = Vector3.ZERO
	_play_anim_state("idle")
	emit_signal("walk_completed")

## Drive the AnimationPlayer to the clip mapped to [param state].
## No-ops when no AnimationPlayer is present or the state is already active.
func _play_anim_state(state: String) -> void:
	if _anim_player == null:
		_anim_player = _find_anim_player(self)
	if _anim_player == null or state == _anim_state:
		return
	_anim_state = state
	var clip := ""
	match state:
		"idle": clip = anim_idle
		"walk": clip = anim_walk
		"talk": clip = anim_talk
	if clip.is_empty():
		return
	if _anim_player.has_animation(clip):
		_anim_player.play(clip)
	else:
		push_warning("ags_character: clip '%s' not found in AnimationPlayer" % clip)

## Recursively search the subtree for an AnimationPlayer.
## Unlike find_child(), this works regardless of node ownership.
func _find_anim_player(node: Node) -> AnimationPlayer:
	for child in node.get_children():
		if child is AnimationPlayer:
			return child as AnimationPlayer
		var found := _find_anim_player(child)
		if found != null:
			return found
	return null

## Inventory management — backing store for AddInventory/LoseInventory/HasInventory.
var _inventory: Array[StringName] = []

func add_inventory(item_name: String) -> void:
	var n := StringName(item_name)
	if not _inventory.has(n):
		_inventory.append(n)

func lose_inventory(item_name: String) -> void:
	_inventory.erase(StringName(item_name))

func has_inventory(item_name: String) -> bool:
	return _inventory.has(StringName(item_name))

## Returns a copy of the inventory as an Array of Strings (for save/load).
func get_inventory() -> Array:
	var out: Array = []
	for n in _inventory:
		out.append(String(n))
	return out

## Replaces the inventory (used by load_game). Accepts Array[String].
func set_inventory(items: Array) -> void:
	_inventory.clear()
	for item in items:
		_inventory.append(StringName(String(item)))

func _find_parent_room() -> AGSRoom:
	# Walk parent chain — works when character is a child of an AGSRoom.
	var parent := get_parent()
	while parent:
		if parent is AGSRoom:
			return parent as AGSRoom
		parent = parent.get_parent()
	# Fallback for AutoLoad characters: rooms are direct children of root.
	for child in get_tree().get_root().get_children():
		if child is AGSRoom:
			return child as AGSRoom
	return null
