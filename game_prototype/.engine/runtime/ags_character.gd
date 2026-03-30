extends AGSCharacter

var _nav_agent: NavigationAgent3D
var _navigating: bool = false

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
	await get_tree().physics_frame  # Frame-0: nav map syncs on first physics frame
	_nav_agent.target_position = target

func walk_to(point_name: String) -> void:
	var room := _find_parent_room()
	if not room:
		push_error("AGSCharacter.walk_to: no parent AGSRoom found")
		return
	await navigate_to(room.get_point(point_name))
	await walk_completed

func face_to(point_name: String) -> void:
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

## Display [param text] as dialogue above the character for [param duration] seconds,
## then emit say_completed. Awaiting this method blocks until the line finishes.
func say(text: String, duration: float = 2.0) -> void:
	say_text = text
	await get_tree().create_timer(duration).timeout
	say_text = ""
	emit_signal("say_completed")

## Display [param text] as a thought (same mechanic as say, different visual intent).
func think(text: String, duration: float = 2.0) -> void:
	await say(text, duration)

func _on_navigation_finished() -> void:
	_navigating = false
	velocity = Vector3.ZERO
	emit_signal("walk_completed")

func _find_parent_room() -> AGSRoom:
	var parent := get_parent()
	while parent:
		if parent is AGSRoom:
			return parent as AGSRoom
		parent = parent.get_parent()
	return null
