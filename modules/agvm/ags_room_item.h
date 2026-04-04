#pragma once

#include "scene/3d/camera_3d.h"
#include "scene/3d/physics/area_3d.h"

// AGSRoomItem is an interactable item placed in a room.
// It links to an AGSItem definition (via item_name) and emits item_clicked
// when the player clicks it. AGSRoom wires item_clicked → item_interact()
// on the room script.
class AGSRoomItem : public Area3D {
	GDCLASS(AGSRoomItem, Area3D);

	String item_name; // must match an AGSItem.item_name in the scene

protected:
	static void _bind_methods();
	void _notification(int p_what);
	void _input_event_call(Camera3D *p_camera, const Ref<InputEvent> &p_event,
			const Vector3 &p_pos, const Vector3 &p_normal, int p_shape) override;

public:
	void set_item_name(const String &p_name);
	String get_item_name() const;

	// Programmatically fire item_clicked — used in headless tests.
	void simulate_click();
};
