**Godot 4 Navigation System**

Developer Reference

*Based on Godot Engine latest docs (≈4.4--4.7 branch)*

|                                                                                                                                                                                                                                                                                                                                                            |
|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **⚠ Experimental** All Navigation nodes (NavigationRegion3D, NavigationAgent3D, NavigationObstacle3D, NavigationLink3D, NavigationServer3D) are marked Experimental. The API is functional and used in shipped games, but method/property names may change in future minor releases. Wrap navigation calls in a thin abstraction layer in larger projects. |

**1 Architecture overview**

Godot\'s navigation system is built around a server/client model. NavigationServer3D is a singleton that owns all internal state. Scene-tree nodes are convenience wrappers that translate to server API calls via RID handles.

**1.1 Core objects**

|                           |                                                                                                         |
|---------------------------|---------------------------------------------------------------------------------------------------------|
| **Property / method**     | **Description**                                                                                         |
| NavigationServer3D        | Singleton. Owns maps, regions, agents, links, obstacles. All pathfinding and avoidance runs here.       |
| NavigationMesh (resource) | Holds baked polygon data. Attached to a NavigationRegion3D.                                             |
| NavigationRegion3D        | Node. Registers a NavigationMesh with the server. Can be enabled/disabled at runtime.                   |
| NavigationAgent3D         | Node. Child of your actor. Wraps pathfinding + avoidance calls. Movement is entirely up to your script. |
| NavigationObstacle3D      | Node. Affects avoidance (dynamic/static) and optionally carves the navmesh during bake.                 |
| NavigationLink3D          | Node. Connects two navmesh positions across a gap (ladders, teleporters, jumps).                        |
| AStar3D                   | Alternative graph-based pathfinder for cell/grid layouts. Not mesh-based.                               |

|                                                                                                                                                                                                           |
|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **ℹ Key distinction** NavigationObstacle3D does NOT affect pathfinding. It only affects avoidance velocity. To block a path, you must modify the navigation mesh (re-bake or use affect_navigation_mesh). |

**1.2 Navigation maps**

Each World3D has a default navigation map. Regions and agents automatically join it. Multiple maps can exist (e.g. separate indoor/outdoor maps) and be queried via NavigationServer3D.

The map synchronises once per physics frame. Any path query made **before** the first physics frame returns an empty path.

**2 Minimal scene setup**

**2.1 Scene tree**

Build this structure in the editor:

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<tbody>
<tr class="odd">
<td><p>Node3D (scene root)</p>
<p>└─ NavigationRegion3D</p>
<p>└─ MeshInstance3D (your floor / level geometry)</p>
<p>└─ CharacterBody3D (your actor)</p>
<p>├─ CollisionShape3D</p>
<p>├─ MeshInstance3D (visuals)</p>
<p>└─ NavigationAgent3D</p>
<p>└─ NavigationObstacle3D (optional blocker, own node)</p></td>
</tr>
</tbody>
</table>

**2.2 Baking the navmesh (editor)**

**1.** Select NavigationRegion3D → Inspector → create a new NavigationMesh resource.

**2.** Set Geometry → Parsed Geometry Type to Static Colliders or Mesh Instances.

**3.** Under Agent → Radius, set the half-width of your character (e.g. 0.5). This erodes the navmesh away from walls.

**4.** Press **Bake NavigationMesh** in the toolbar. A blue overlay appears showing the walkable area.

|                                                                                                                                                                                                                                                                          |
|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **⚠ Cell size tip** Smaller Cell Size (Cells → Size) = more accurate navmesh but slower bake. If navmesh is missing in tight spaces, lower Cell Size and raise Agent Radius slightly. Match the map cell size in Project Settings → navigation → 3d → default_cell_size. |

**2.3 Minimal actor script (GDScript)**

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<tbody>
<tr class="odd">
<td><p>extends CharacterBody3D</p>
<p>const SPEED := 3.5</p>
<p>@onready var agent: NavigationAgent3D = $NavigationAgent3D</p>
<p>func _ready() -&gt; void:</p>
<p>agent.path_desired_distance = 0.5</p>
<p>agent.target_desired_distance = 0.5</p>
<p># Must defer: nav map is empty on frame 0</p>
<p>actor_setup.call_deferred()</p>
<p>func actor_setup() -&gt; void:</p>
<p>await get_tree().physics_frame</p>
<p>move_to(Vector3(-3.0, 0.0, 2.0))</p>
<p>func move_to(target: Vector3) -&gt; void:</p>
<p>agent.target_position = target</p>
<p>func _physics_process(_delta: float) -&gt; void:</p>
<p>if agent.is_navigation_finished():</p>
<p>return</p>
<p>var next := agent.get_next_path_position()</p>
<p>velocity = global_position.direction_to(next) * SPEED</p>
<p>move_and_slide()</p></td>
</tr>
</tbody>
</table>

|                                                                                                                                                                                                                           |
|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **ℹ Frame-0 rule** Always defer the first target_position assignment until after at least one physics frame. Calling it in \_ready() returns an empty path because the NavigationServer has not yet synchronised the map. |

**3 NavigationAgent3D reference**

NavigationAgent3D is a helper node that wraps NavigationServer3D calls. It must be a child of a Node3D-inheriting node (the actor). The navigation system never moves the parent --- movement is entirely up to your script.

**3.1 Pathfinding properties**

|                       |                                                                                                          |
|-----------------------|----------------------------------------------------------------------------------------------------------|
| **Property / method** | **Description**                                                                                          |
| target_position       | Set this (global coords) to trigger a new path query.                                                    |
| navigation_layers     | Bitmask. Limits which navmesh regions the agent can use.                                                 |
| pathfinding_algorithm | How polygons are traversed (AStar by default).                                                           |
| path_postprocessing   | Post-process the raw corridor (e.g. CORRIDORFUNNEL for smoother paths).                                  |
| simplify_path         | Remove less critical points from the path.                                                               |
| simplify_epsilon      | Tolerance for path simplification.                                                                       |
| path_metadata_flags   | Enable extra data (owner RID, type, etc.) on each path point. Disabling breaks related signal emissions. |

**3.2 Path-following properties**

|                         |                                                                                     |
|-------------------------|-------------------------------------------------------------------------------------|
| **Property / method**   | **Description**                                                                     |
| path_desired_distance   | Distance from the next waypoint at which the agent advances its path index.         |
| target_desired_distance | Distance from the final target at which navigation is considered finished.          |
| path_max_distance       | If the actor drifts further than this from the ideal path, a new path is requested. |

**3.3 Key methods**

|                          |                                                                                                              |
|--------------------------|--------------------------------------------------------------------------------------------------------------|
| **Property / method**    | **Description**                                                                                              |
| get_next_path_position() | Returns the next world-space waypoint. Call once per \_physics_process(). Also advances internal path state. |
| is_navigation_finished() | Returns true when the actor has reached target_desired_distance from the target. Always check this first.    |
| set_velocity(v: Vector3) | Feed current velocity for avoidance calculation. Triggers velocity_computed signal.                          |
| get_rid() -\> RID        | Returns the underlying server RID for direct NavigationServer3D calls.                                       |

**3.4 Signals**

|                                  |                                                                                                  |
|----------------------------------|--------------------------------------------------------------------------------------------------|
| **Property / method**            | **Description**                                                                                  |
| velocity_computed(safe_velocity) | Emitted when avoidance has computed a safe velocity. Use this instead of the raw input velocity. |
| path_changed()                   | Emitted when the internal path is recalculated.                                                  |
| target_reached()                 | Emitted when the actor enters target_desired_distance of the target.                             |
| waypoint_reached(details)        | Emitted when the agent advances to the next path waypoint.                                       |
| navigation_finished()            | Emitted when is_navigation_finished() becomes true.                                              |
| link_reached(details)            | Emitted when the path contains a NavigationLink and the agent reaches it.                        |

**3.5 Common pitfalls**

- **Empty path on frame 0:** defer target_position until after the first physics_frame.

- **Agent dancing in place:** caused by path updates every frame. Check path_max_distance isn\'t set too short.

- **Agent backtracking:** agent is moving faster than path_desired_distance. Increase that value.

- **Brief backwards step:** precision issue when agent spawns directly over a navmesh edge. Usually cosmetic.

- **get_next_path_position() after finish:** calling it when is_navigation_finished() is true causes jitter. Guard with an early return.

**4 NavigationObstacle3D reference**

Obstacles serve two independent purposes, controlled by separate flags:

- **affect_navigation_mesh** --- carves the navmesh during bake (static geometry blocking).

- **avoidance_enabled** --- pushes agents away at runtime via the avoidance simulation.

|                                                                                                                                                                                                                                                         |
|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **⚠ Critical distinction** Avoidance does NOT reroute the agent\'s path. If an obstacle completely blocks the only route to the target, the agent will walk into it rather than going around. Use affect_navigation_mesh + rebake for hard path blocks. |

**4.1 Navmesh baking mode**

|                        |                                                                                                                               |
|------------------------|-------------------------------------------------------------------------------------------------------------------------------|
| **Property / method**  | **Description**                                                                                                               |
| affect_navigation_mesh | When true, the obstacle is included in the next navmesh bake and removes geometry inside its shape.                           |
| carve_navigation_mesh  | When true, the obstacle acts as a stencil --- it cuts into the already-offset navmesh surface, ignoring agent radius offsets. |
| vertices               | Array of Vector3 positions defining the obstacle polygon. Y-axis of each vertex is ignored; the obstacle is projected flat.   |
| height                 | Vertical extent of the obstacle shape (3D only).                                                                              |

**4.2 Avoidance: static vs dynamic**

**Static obstacle (polygon-based)**

Active when vertices is populated. Acts as a hard do-not-cross boundary. Only works with 2D avoidance mode (xz plane).

- Cannot move smoothly --- warp only. Rebuilds from scratch on every position change (expensive per frame).

- Define vertices winding order to push agents out (CCW) or suck them in (CW).

- If warped on top of agents, agents can get stuck inside.

**Dynamic obstacle (radius-based)**

Active when radius \> 0. Acts as a soft push-away zone, similar to other agents.

- Can change position every frame cheaply.

- Works with both 2D and 3D avoidance modes.

- Not reliable for constraining agents in narrow spaces.

- Set velocity on the obstacle so agents can predict its movement.

**4.3 Key properties (avoidance)**

|                       |                                                                               |
|-----------------------|-------------------------------------------------------------------------------|
| **Property / method** | **Description**                                                               |
| avoidance_enabled     | Enables the avoidance push effect. Disable if only using for navmesh carving. |
| radius                | Radius of the dynamic avoidance circle/sphere. Set \> 0 for dynamic mode.     |
| avoidance_layers      | Bitmask. Agents avoid this obstacle only if their avoidance_mask matches.     |

**4.4 Recommended pattern for moving obstacles**

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<tbody>
<tr class="odd">
<td><p># When obstacle starts moving:</p>
<p>obstacle.vertices = [] # remove static polygon</p>
<p>obstacle.radius = 1.2 # activate dynamic push</p>
<p># When obstacle reaches its destination:</p>
<p># 1. Gradually increase radius to clear space</p>
<p>obstacle.radius = 2.5</p>
<p># 2. Wait a frame for agents to move away</p>
<p>await get_tree().physics_frame</p>
<p># 3. Rebuild static polygon and remove radius</p>
<p>obstacle.vertices = my_polygon</p>
<p>obstacle.radius = 0.0</p></td>
</tr>
</tbody>
</table>

**5 NavigationLink3D**

Links connect two navmesh positions that are not physically adjacent --- ladders, jump pads, teleporters, stairways across gaps. The pathfinder treats a link as a traversable edge with a configurable cost. Actual movement across the link is entirely your script\'s responsibility.

|                               |                                                                                |
|-------------------------------|--------------------------------------------------------------------------------|
| **Property / method**         | **Description**                                                                |
| start_position / end_position | World-space endpoints of the link (set in global coords or local to the node). |
| bidirectional                 | If false, the link is one-way from start to end.                               |
| navigation_layers             | Bitmask matching the regions this link should connect.                         |
| enter_cost                    | Cost added when entering the link (affects path preference).                   |
| travel_cost                   | Cost per unit of link length (multiplier).                                     |

|                                                                                                                                                                                                                                                                                      |
|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **ℹ Link traversal** The agent emits the link_reached signal when it arrives at the link\'s start (or end, for bidirectional). Your script must take over movement --- animate the climb, teleport, jump --- and then call agent.target_position again to resume normal pathfinding. |

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<tbody>
<tr class="odd">
<td><p>func _ready():</p>
<p>agent.link_reached.connect(_on_link_reached)</p>
<p>func _on_link_reached(details: Dictionary) -&gt; void:</p>
<p>var link: NavigationLink3D = details["link"]</p>
<p>var exit: Vector3 = details["exit_location"]</p>
<p># Custom movement: teleport, animate, etc.</p>
<p>global_position = exit</p>
<p># Resume pathfinding</p>
<p>agent.target_position = movement_target</p></td>
</tr>
</tbody>
</table>

**6 Navigation layers**

Navigation layers work like physics layers: a bitmask on regions, links, and agents. An agent can only path through a region if at least one bit matches.

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<tbody>
<tr class="odd">
<td><p># In Project Settings → Layer Names → 3D Navigation, name your layers.</p>
<p># Then assign via Inspector or script:</p>
<p># Region: only ground layer (bit 0)</p>
<p>navigation_region.navigation_layers = 1</p>
<p># Agent: can use ground (bit 0) + water (bit 1)</p>
<p>navigation_agent.navigation_layers = 0b11 # = 3</p>
<p># Toggle a layer at runtime:</p>
<p>NavigationServer3D.region_set_navigation_layers(</p>
<p>region.get_region_rid(),</p>
<p>NavigationServer3D.region_get_navigation_layers(region.get_region_rid()) | (1 &lt;&lt; 2)</p>
<p>)</p></td>
</tr>
</tbody>
</table>

|                                                                                                                                                                                                                                                      |
|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **ℹ Use case** Layers are the standard way to support multiple actor types. A flying enemy uses a different region (air navmesh) than a ground enemy. Set non-overlapping layer bits and assign matching navigation_layers to each actor and region. |

**7 Agent avoidance (RVO)**

Avoidance uses a Reciprocal Velocity Obstacles (RVO) algorithm. Agents compute safe velocities that avoid other avoidance-enabled agents and obstacles within their neighbor_distance.

|                                                                                                                                                                                                            |
|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **⚠ What avoidance is NOT** Avoidance knows nothing about navmeshes, physics collision, or geometry. It is purely a velocity-adjustment layer. Do not use it as a substitute for collision or pathfinding. |

**7.1 Enabling avoidance on an agent**

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<tbody>
<tr class="odd">
<td><p># Inspector: check avoidance_enabled on NavigationAgent3D</p>
<p># Connect the signal:</p>
<p>agent.velocity_computed.connect(_on_safe_velocity)</p>
<p>func _physics_process(_delta):</p>
<p>if agent.is_navigation_finished(): return</p>
<p>var next := agent.get_next_path_position()</p>
<p>var desired := global_position.direction_to(next) * SPEED</p>
<p>agent.set_velocity(desired) # hands velocity to avoidance</p>
<p>func _on_safe_velocity(safe_vel: Vector3) -&gt; void:</p>
<p>velocity = safe_vel</p>
<p>move_and_slide()</p></td>
</tr>
</tbody>
</table>

**7.2 Avoidance properties**

|                        |                                                                                                         |
|------------------------|---------------------------------------------------------------------------------------------------------|
| **Property / method**  | **Description**                                                                                         |
| avoidance_enabled      | Master toggle. Disable to save CPU when not needed.                                                     |
| radius                 | Avoidance body size (not related to NavMesh agent radius).                                              |
| neighbor_distance      | Search radius for other agents. Lower = cheaper.                                                        |
| max_neighbors          | Max agents considered. Lower = cheaper but may miss some.                                               |
| time_horizon_agents    | Seconds ahead to predict agent collisions. Lower = faster reaction, higher = smoother.                  |
| time_horizon_obstacles | Seconds ahead to predict obstacle movement.                                                             |
| max_speed              | If the parent moves faster, safe_velocity accuracy degrades.                                            |
| use_3d_avoidance       | Switch between 2D (xz plane) and full 3D (xyz) avoidance. 2D and 3D agents are in separate simulations. |
| avoidance_layers       | Bitmask of which avoidance layer this agent is on.                                                      |
| avoidance_mask         | Bitmask of which layers this agent avoids.                                                              |
| avoidance_priority     | Higher priority agents are ignored by lower priority agents.                                            |

|                                                                                                                                                                                                                                                                                 |
|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **ℹ 2D vs 3D avoidance** The default 2D avoidance (xz plane) is cheaper and sufficient for most ground-based games. Agents and obstacles automatically ignore each other when vertically separated. Only enable use_3d_avoidance for flying agents that need true 3D push-away. |

**8 Runtime navmesh baking**

Navmeshes can be baked at runtime (not just in the editor). This is necessary for procedurally generated levels or dynamically placed obstacles that affect pathfinding.

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<tbody>
<tr class="odd">
<td><p>extends NavigationRegion3D</p>
<p>func bake_async() -&gt; void:</p>
<p># Optional: connect to completion signal</p>
<p>bake_navigation_mesh.connect(_on_bake_finished, CONNECT_ONE_SHOT)</p>
<p># Trigger async bake (non-blocking)</p>
<p>bake_navigation_mesh(true)</p>
<p>func _on_bake_finished() -&gt; void:</p>
<p>print("NavMesh baked — agent paths updated.")</p></td>
</tr>
</tbody>
</table>

For procedural obstacles, add them to source geometry data before baking:

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<tbody>
<tr class="odd">
<td><p>var nav_mesh := NavigationMesh.new()</p>
<p>var src_data := NavigationMeshSourceGeometryData3D.new()</p>
<p>NavigationServer3D.parse_source_geometry_data(</p>
<p>nav_mesh, src_data, $LevelRoot</p>
<p>)</p>
<p># Add a custom obstacle outline (projected flat):</p>
<p>var verts := PackedVector3Array([</p>
<p>Vector3(-2, 0, -2), Vector3(2, 0, -2),</p>
<p>Vector3(2, 0, 2), Vector3(-2, 0, 2)</p>
<p>])</p>
<p>src_data.add_projected_obstruction(verts, 2.0, 0.0, false)</p>
<p>NavigationServer3D.bake_from_source_geometry_data(</p>
<p>nav_mesh, src_data</p>
<p>)</p>
<p>navigation_mesh = nav_mesh</p></td>
</tr>
</tbody>
</table>

**9 Debug tools**

Navigation debug visualisation is enabled by default in the editor (Project → Project Settings → Debug → Navigation).

|                                |                                                                                    |
|--------------------------------|------------------------------------------------------------------------------------|
| **Property / method**          | **Description**                                                                    |
| Visible Navigation Mesh        | Show the baked navmesh polygon overlay (blue).                                     |
| Visible Navigation Links       | Show NavigationLink3D connections.                                                 |
| Visible Navigation Agent Paths | Show the current computed path for each agent.                                     |
| Visible Navigation Avoidance   | Show avoidance radius circles and safe velocity vectors.                           |
| Enable at runtime              | NavigationServer3D.set_debug_enabled(true) --- also needs the Visible\* flags set. |

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<tbody>
<tr class="odd">
<td><p># Enable navigation debug at runtime (e.g. for testing builds):</p>
<p>NavigationServer3D.set_debug_enabled(true)</p>
<p>NavigationServer3D.set_debug_navigation_edge_connection_disabled_color(Color.RED)</p></td>
</tr>
</tbody>
</table>

**10 Connecting multiple regions**

The NavigationServer automatically joins navmesh edges of different regions that are within the map\'s edge_connection_margin. Use NavigationLink3D for connections that cannot be joined by proximity (different floors, large gaps).

|                                                                                                                                                                                                                                                                                  |
|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **⚠ Merge margin gotcha** The default edge_connection_margin is 5.0 --- often too high, causing invalid floating path points near region boundaries. Lower it to 0.5--1.0 for most scenes: NavigationServer3D.map_set_edge_connection_margin(get_world_3d().navigation_map, 1.0) |

<table>
<colgroup>
<col style="width: 100%" />
</colgroup>
<tbody>
<tr class="odd">
<td><p>func _ready():</p>
<p>var map := get_world_3d().navigation_map</p>
<p>NavigationServer3D.map_set_edge_connection_margin(map, 1.0)</p></td>
</tr>
</tbody>
</table>

**11 Performance notes**

- **Path queries are async per server frame.** Multiple agents querying simultaneously is fine --- they are batched.

- **Avoidance scales with neighbor_distance and max_neighbors.** Reduce both for large crowds. Most agents don\'t need to see far.

- **Disable avoidance on stationary agents.** Toggle avoidance_enabled off when the agent is idle.

- **Separate maps for separate regions.** Agents only pay synchronisation cost for regions on their map.

- **Avoid per-frame rebakes.** Baking is expensive. Cache bakes and trigger only on meaningful geometry changes.

- **Static obstacles are expensive to move.** Use dynamic radius mode while an obstacle is moving; rebuild static polygon only at rest.

- **path_max_distance prevents jitter.** Tune it relative to agent speed --- too short causes constant repath spam.

**Quick reference card**

**Minimum viable navigation checklist**

1.  NavigationRegion3D with a baked NavigationMesh in the scene.

2.  NavigationAgent3D as a child of the actor.

3.  Set path_desired_distance and target_desired_distance in \_ready().

4.  Defer first target_position until after one physics_frame.

5.  In \_physics_process: guard with is_navigation_finished(), call get_next_path_position(), set velocity, call move_and_slide().

**Obstacle decision tree**

- **Static wall / permanent block:** Enable affect_navigation_mesh on NavigationObstacle3D, rebake. No avoidance needed.

- **Moving actor avoidance (other NPCs):** Enable avoidance_enabled on NavigationAgent3D. Tune radius and neighbor_distance.

- **Moving prop (crate, door):** NavigationObstacle3D with dynamic radius while moving, static vertices when stopped.

- **Cross-gap connection (ladder, jump):** NavigationLink3D + custom movement script on link_reached signal.

**Signal summary**

|                               |                                                        |
|-------------------------------|--------------------------------------------------------|
| **Property / method**         | **Description**                                        |
| velocity_computed(safe_vel)   | Agent. Use safe_vel as velocity when avoidance is on.  |
| target_reached()              | Agent. Final destination reached.                      |
| navigation_finished()         | Agent. Same as is_navigation_finished() becoming true. |
| waypoint_reached(details)     | Agent. Advanced to next path point.                    |
| link_reached(details)         | Agent. At a NavigationLink --- take over movement.     |
| bake_navigation_mesh (signal) | NavigationRegion3D. Emitted when async bake finishes.  |
