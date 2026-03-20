#!/bin/bash

# AGS3D — GitHub Issues + Milestones bulk creator
# Run from inside your local ags3d repo directory
# Requires: gh auth login (already done)

REPO="forkfulloffuzz/ags3d"

# Milestone titles — used as strings for --milestone flag
M1="M1 — Godot Fork Setup"
M2="M2 — AGS-Spirit Parser"
M3="M3 — GDScript Emitter"
M4="M4 — Room Node"
M5="M5 — Character Node"
M6="M6 — End-to-End Wiring"

echo "========================================="
echo " AGS3D GitHub Setup"
echo " Repo: $REPO"
echo "========================================="
echo ""

# ─── LABELS ───────────────────────────────────────────────────────────────────

echo "Creating labels..."

gh label create "risk:low"  --color "1A6B3A" --description "Low risk task"               --repo $REPO 2>/dev/null || echo "  label risk:low already exists"
gh label create "risk:med"  --color "B45309" --description "Medium risk task"             --repo $REPO 2>/dev/null || echo "  label risk:med already exists"
gh label create "risk:high" --color "CC2200" --description "High risk task"               --repo $REPO 2>/dev/null || echo "  label risk:high already exists"
gh label create "M1"        --color "1E4D8C" --description "Milestone 1: Godot Fork"     --repo $REPO 2>/dev/null || echo "  label M1 already exists"
gh label create "M2"        --color "1E4D8C" --description "Milestone 2: Parser"         --repo $REPO 2>/dev/null || echo "  label M2 already exists"
gh label create "M3"        --color "1E4D8C" --description "Milestone 3: Emitter"        --repo $REPO 2>/dev/null || echo "  label M3 already exists"
gh label create "M4"        --color "1E4D8C" --description "Milestone 4: Room Node"      --repo $REPO 2>/dev/null || echo "  label M4 already exists"
gh label create "M5"        --color "1E4D8C" --description "Milestone 5: Character Node" --repo $REPO 2>/dev/null || echo "  label M5 already exists"
gh label create "M6"        --color "1E4D8C" --description "Milestone 6: End-to-End"     --repo $REPO 2>/dev/null || echo "  label M6 already exists"
gh label create "prototype" --color "6B21A8" --description "Prototype scope"              --repo $REPO 2>/dev/null || echo "  label prototype already exists"

echo ""

# ─── MILESTONES ───────────────────────────────────────────────────────────────

echo "Creating milestones (skips if already exists)..."

gh api repos/$REPO/milestones --method POST \
  --field title="$M1" \
  --field description="Establish the fork, module skeleton, and build pipeline before any engine code is written." \
  2>/dev/null && echo "  created: $M1" || echo "  exists: $M1"

gh api repos/$REPO/milestones --method POST \
  --field title="$M2" \
  --field description="Parse AGS-spirit source files into a well-formed AST. Highest-risk milestone." \
  2>/dev/null && echo "  created: $M2" || echo "  exists: $M2"

gh api repos/$REPO/milestones --method POST \
  --field title="$M3" \
  --field description="Walk the AST and emit valid, readable GDScript that Godot can execute." \
  2>/dev/null && echo "  created: $M3" || echo "  exists: $M3"

gh api repos/$REPO/milestones --method POST \
  --field title="$M4" \
  --field description="Implement the Room node type with logic geometry in the Godot editor." \
  2>/dev/null && echo "  created: $M4" || echo "  exists: $M4"

gh api repos/$REPO/milestones --method POST \
  --field title="$M5" \
  --field description="Implement the Character node with navigation and a blocking WalkTo call." \
  2>/dev/null && echo "  created: $M5" || echo "  exists: $M5"

gh api repos/$REPO/milestones --method POST \
  --field title="$M6" \
  --field description="Connect all systems so an AGS-spirit script drives a character in a room. Prototype success criterion." \
  2>/dev/null && echo "  created: $M6" || echo "  exists: $M6"

echo ""

# ─── HELPER ───────────────────────────────────────────────────────────────────

create_issue() {
  local title="$1"
  local body="$2"
  local milestone="$3"
  local labels="$4"
  echo "  Creating: $title"
  gh issue create \
    --repo "$REPO" \
    --title "$title" \
    --body "$body" \
    --milestone "$milestone" \
    --label "$labels" \
    --label "prototype"
  sleep 0.6
}

# ─── M1 ISSUES ────────────────────────────────────────────────────────────────

echo "Creating M1 issues..."

create_issue \
"T01 — Fork Godot and establish repository" \
"## Task
Fork official Godot and establish the AGS3D development repository.

## Subtasks
- Fork official Godot repo on GitHub
- Set upstream remote for future merges
- Confirm clean build on target dev platform
- Set up CI pipeline to build fork on push

## Acceptance Criteria
Clean build passes on dev machine. CI pipeline builds the fork on push.

## Depends on
None

## Notes
Take time to understand the Godot build system (SCons) before proceeding to T02." \
"$M1" "M1,risk:low"

create_issue \
"T02 — Create module skeleton: modules/agvm/" \
"## Task
Create the AGS3D module directory and register it with the Godot build system.

## Subtasks
- Create \`modules/agvm/\` directory with SCsub build file
- Register module in Godot build system
- Add stub class that prints a log line on engine init
- Confirm module appears in engine startup output

## Acceptance Criteria
Engine boots and stub log line appears confirming module is loaded.

## Depends on
T01" \
"$M1" "M1,risk:low"

create_issue \
"T03 — Register AGS-spirit ScriptLanguage interface" \
"## Task
Implement the Godot ScriptLanguage interface for .agscript files.

## Subtasks
- Implement ScriptLanguage subclass: AGSScriptLanguage
- Register with ScriptServer in module init
- Stub all required virtual methods so engine does not crash
- Verify .agscript appears as a recognised extension in the editor

## Acceptance Criteria
Godot editor recognises .agscript as a known file type. Stubs only — no functionality yet.

## Depends on
T02" \
"$M1" "M1,risk:low"

create_issue \
"T04 — Define project file format and directory scanner" \
"## Task
Define game.agp TOML schema and implement the project file scanner.

## Subtasks
- Define game.agp TOML schema (name, start_room, start_character, rendering_mode, etc.)
- Implement directory scanner finding all .agscript, .agroom, .agchar files recursively
- Expose scanned project structure as a C++ data model
- Handle missing or malformed game.agp gracefully with clear error message

## Acceptance Criteria
Scanner correctly enumerates all adventure game source files from a minimal test project directory.

## Depends on
T02" \
"$M1" "M1,risk:low"

create_issue \
"T05 — Build pipeline: ag CLI stub" \
"## Task
Create the ag command-line tool with stub build and run commands.

## Subtasks
- Create ag command-line tool (bash script or compiled binary)
- Implement \`ag build\` stub — prints plan, does not yet compile anything
- Implement \`ag run\` — launches Godot editor with the project path

## Acceptance Criteria
\`ag run\` launches the Godot editor with the test project loaded.

## Depends on
T01" \
"$M1" "M1,risk:low"

# ─── M2 ISSUES ────────────────────────────────────────────────────────────────

echo "Creating M2 issues..."

create_issue \
"T06 — Define the AGS-spirit grammar (formal spec)" \
"## Task
Write a formal grammar specification document before writing any parser code.

## Subtasks
- Document supported subset of AGS Script syntax for prototype scope
- Specify: functions, variables, if/else, while, return
- Specify: event handlers (room_Enter, hotspot_Interact, etc.)
- Specify: built-in type names (Character, Room, Point, Region)
- Explicitly list what is NOT supported in prototype scope

## Acceptance Criteria
Written grammar spec committed to repo before any parser code is written.

## Depends on
None

## Notes
Do not skip this task. Writing the grammar first prevents scope creep in the parser and gives a clear definition of done for T09." \
"$M2" "M2,risk:med"

create_issue \
"T07 — Implement lexer / tokenizer" \
"## Task
Implement a lexer that converts AGS-spirit source text into a token stream.

## Subtasks
- Define all token types: keywords, identifiers, literals, operators, punctuation
- Implement line and column tracking for accurate error messages
- Skip whitespace and comments correctly
- Unit test all token types

## Acceptance Criteria
Lexer correctly tokenizes all constructs from the grammar spec. Unit tests cover all token types.

## Depends on
T06" \
"$M2" "M2,risk:med"

create_issue \
"T08 — Define AST node types" \
"## Task
Define the full set of AST node types needed to represent all grammar spec constructs.

## Subtasks
- Statement nodes: FunctionDecl, EventHandler, Block, IfStmt, WhileStmt, AssignStmt, ReturnStmt, ExprStmt
- Expression nodes: CallExpr, MemberExpr, BinaryExpr, UnaryExpr
- Leaf nodes: Literal (int, float, string, bool), Identifier

## Acceptance Criteria
AST node hierarchy defined in code with clear ownership semantics. No parsing yet.

## Depends on
T06" \
"$M2" "M2,risk:low"

create_issue \
"T09 — Implement recursive descent parser" \
"## Task
Implement the parser that consumes the token stream and produces an AST.

## Subtasks
- Parse function declarations and event handler blocks
- Parse all statement types from grammar spec
- Parse expressions with correct operator precedence
- Build AST nodes from all parsed constructs

## Acceptance Criteria
Parser correctly produces a valid AST for a representative set of AGS-spirit scripts covering all grammar spec constructs.

## Depends on
T07, T08

## Notes
HIGH RISK. Allocate extra time. Correctness matters more than speed. Write tests as you go." \
"$M2" "M2,risk:high"

create_issue \
"T10 — Implement symbol table and type resolver" \
"## Task
Build a symbol table from the AST and resolve all identifier references.

## Subtasks
- First pass: collect all function and variable declarations
- Second pass: resolve all identifier references
- Validate referenced names exist — report clear errors for undefined references
- Record type of each expression node

## Acceptance Criteria
Symbol table correctly resolves all names in test scripts. Undefined references produce clear error messages with file and line number.

## Depends on
T09" \
"$M2" "M2,risk:med"

create_issue \
"T11 — Identify and annotate blocking calls" \
"## Task
During symbol resolution, identify and annotate calls to blocking built-in functions in the AST.

## Subtasks
- Maintain list of known blocking built-in functions (WalkTo, PlayAnimation, Wait, etc.)
- Annotate all blocking call sites in the AST
- Mark any function as async if it contains any blocking call (direct or transitive)

## Acceptance Criteria
All blocking call sites correctly annotated. Needed by emitter to insert await correctly.

## Depends on
T10" \
"$M2" "M2,risk:med"

create_issue \
"T12 — Parser error handling and recovery" \
"## Task
Ensure the parser produces actionable errors and never crashes on malformed input.

## Subtasks
- Error messages include: file, line, column, what was expected vs found
- Basic panic-mode recovery to continue parsing after errors
- Never crash on any malformed input — always return an error list
- Test with a suite of deliberately broken scripts

## Acceptance Criteria
Parser handles all malformed inputs without crashing. Error messages are clear without reading parser source code.

## Depends on
T09" \
"$M2" "M2,risk:med"

# ─── M3 ISSUES ────────────────────────────────────────────────────────────────

echo "Creating M3 issues..."

create_issue \
"T13 — Implement emitter scaffolding" \
"## Task
Set up the AST visitor and output infrastructure for the GDScript emitter.

## Subtasks
- AST visitor pattern or recursive walk function
- Output buffer with indentation tracking (GDScript is indentation-sensitive)
- File writer placing output in .engine/generated/ mirroring source directory structure

## Acceptance Criteria
Emitter can be invoked on any AST and produces an empty output file in the correct location.

## Depends on
T09" \
"$M3" "M3,risk:low"

create_issue \
"T14 — Emit function and event handler declarations" \
"## Task
Emit GDScript function and event handler skeletons from AST function declaration nodes.

## Subtasks
- Map AGS-spirit function declarations to GDScript func keyword
- Map event handler names to Godot signal handler naming convention
- Emit correct parameter lists
- Emit empty function bodies as placeholder

## Acceptance Criteria
Function and event handler skeletons emit correctly. GDScript parser accepts the output.

## Depends on
T13" \
"$M3" "M3,risk:low"

create_issue \
"T15 — Emit statements and expressions" \
"## Task
Emit all statement and expression types from the grammar spec as correct GDScript.

## Subtasks
- All statement types: if/else, while, assign, return, expression statement
- All expression types: call, member access, binary op, unary op, literal, identifier
- Correct GDScript operator mapping

## Acceptance Criteria
All grammar spec constructs emit valid GDScript. Round-trip test: parse then emit and confirm GDScript is syntactically valid.

## Depends on
T13, T14" \
"$M3" "M3,risk:med"

create_issue \
"T16 — Emit blocking calls with await" \
"## Task
For all annotated blocking call sites, emit the GDScript await keyword correctly.

## Subtasks
- For each blocking-annotated call site, emit await before the call expression
- For async-marked functions, emit correct async signature in GDScript
- Verify await chains correctly through multiple call levels
- Test that script execution actually blocks until the awaited call completes

## Acceptance Criteria
character.WalkTo(point.door) in AGS-spirit emits await character.walk_to(\"door\") in GDScript. Walk completes before the next line executes.

## Depends on
T11, T15

## Notes
HIGH RISK. Getting await wrong produces subtly incorrect behaviour. Test thoroughly with nested blocking calls." \
"$M3" "M3,risk:high"

create_issue \
"T17 — Emit source map file" \
"## Task
Write a sidecar source map file mapping each generated GDScript line back to its AGS-spirit source location.

## Subtasks
- Track AGS-spirit source location through all emitter output operations
- Write .agmap sidecar file alongside each generated GDScript file
- Format: JSON array of [gdscript_line, agscript_file, agscript_line]
- Manual verification: a known AGS-spirit line maps to the correct GDScript line

## Acceptance Criteria
Source map exists for every generated GDScript file. Line mapping is correct.

## Depends on
T15" \
"$M3" "M3,risk:low"

create_issue \
"T18 — Wire emitter into ag build" \
"## Task
Connect parser and emitter to the ag build command.

## Subtasks
- ag build scans for changed .agscript files
- Invokes parser then emitter for each changed file
- Writes GDScript to .engine/generated/
- Reports errors with AGS-spirit source locations, never GDScript paths

## Acceptance Criteria
ag build on the test project produces GDScript files. Error messages reference .agscript files and line numbers.

## Depends on
T05, T16" \
"$M3" "M3,risk:low"

# ─── M4 ISSUES ────────────────────────────────────────────────────────────────

echo "Creating M4 issues..."

create_issue \
"T19 — Implement Room node type" \
"## Task
Implement AGSRoom as a first-class Godot node type registered by the agvm module.

## Subtasks
- C++ class AGSRoom extending Node3D
- Register with Godot ClassDB via agvm module
- Expose room_name property in editor inspector
- Room appears in Add Node dialog under AGS3D category
- Distinct editor icon

## Acceptance Criteria
Author can add an AGSRoom node from the editor. Visible in scene tree with correct name and icon.

## Depends on
T02" \
"$M4" "M4,risk:low"

create_issue \
"T20 — Implement WalkableSurface node type" \
"## Task
Implement AGSWalkableSurface whose geometry automatically becomes the room navmesh.

## Subtasks
- C++ class AGSWalkableSurface extending StaticBody3D
- Mesh registered as navmesh source when room initialises
- Editor: semi-transparent green overlay
- Runtime: invisible
- Navmesh baked from surface on scene load

## Acceptance Criteria
A character with NavigationAgent3D navigates correctly on a WalkableSurface mesh.

## Depends on
T19" \
"$M4" "M4,risk:med"

create_issue \
"T21 — Implement BlockerVolume node type" \
"## Task
Implement AGSBlockerVolume — a collision volume characters cannot enter.

## Subtasks
- C++ class AGSBlockerVolume extending StaticBody3D with CollisionShape3D
- Characters cannot walk through the volume
- Editor: semi-transparent red
- Runtime: invisible

## Acceptance Criteria
Character pathfinding routes around a BlockerVolume correctly.

## Depends on
T19" \
"$M4" "M4,risk:low"

create_issue \
"T22 — Implement named point placement" \
"## Task
Implement AGSPoint — a named spatial point referenceable by name from scripts.

## Subtasks
- C++ class AGSPoint extending Node3D
- point_name property in editor inspector
- Visible as labelled gizmos in editor viewport
- Self-registers with parent AGSRoom on ready
- AGSRoom.get_point(name) returns correct world Vector3

## Acceptance Criteria
AGSRoom.get_point(\"door_left\") returns the correct world position. Points added in editor are immediately available in script.

## Depends on
T19" \
"$M4" "M4,risk:low"

create_issue \
"T23 — Implement TriggerRegion node type" \
"## Task
Implement AGSTriggerRegion — a named volume firing events on character enter/exit.

## Subtasks
- C++ class AGSTriggerRegion extending Area3D
- region_name property in inspector
- Emits region_entered(character) and region_exited(character) signals
- Self-registers with parent AGSRoom by name on ready

## Acceptance Criteria
Character entering the volume fires region_entered with the correct character reference. Connectable from room script.

## Depends on
T19" \
"$M4" "M4,risk:low"

create_issue \
"T24 — Implement HotspotSurface node type" \
"## Task
Implement AGSHotspot — a named interactive surface detecting mouse/touch clicks via raycast.

## Subtasks
- C++ class AGSHotspot extending Area3D with CollisionShape3D
- hotspot_name property in inspector
- Raycast from camera through mouse position hits collision shape
- Emits hotspot_clicked(hotspot_name) signal to parent AGSRoom

## Acceptance Criteria
Clicking a HotspotSurface fires hotspot_clicked with the correct name. Room script can handle it.

## Depends on
T19" \
"$M4" "M4,risk:med"

# ─── M5 ISSUES ────────────────────────────────────────────────────────────────

echo "Creating M5 issues..."

create_issue \
"T25 — Implement Character node type" \
"## Task
Implement AGSCharacter as a first-class Godot node type with runtime registration.

## Subtasks
- C++ class AGSCharacter extending CharacterBody3D
- character_name property in inspector
- Self-registers with AGSRuntime singleton by name on ready
- Placeholder capsule mesh visible in editor and at runtime

## Acceptance Criteria
AGSRuntime.get_character(\"player\") returns the correct node at runtime.

## Depends on
T19" \
"$M5" "M5,risk:low"

create_issue \
"T26 — Integrate Godot NavigationAgent3D" \
"## Task
Wire Godot NavigationAgent3D into AGSCharacter for pathfinding.

## Subtasks
- AGSCharacter owns a NavigationAgent3D child node
- Navigation target set from AGSRoom named point world positions
- Character moves along path using move_and_slide each physics frame
- Character stops correctly when navigation_finished fires

## Acceptance Criteria
Character navigates to a target point along the navmesh, routing correctly around BlockerVolumes.

## Depends on
T25, T20, T21" \
"$M5" "M5,risk:med"

create_issue \
"T27 — Implement WalkTo as a blocking function" \
"## Task
Implement walk_to(point_name) on AGSCharacter as a properly blocking async GDScript method.

## Subtasks
- GDScript method walk_to(point_name: String) on AGSCharacter
- Looks up point world position from parent room via AGSRuntime
- Sets NavigationAgent3D target
- Awaits navigation_finished signal before returning
- Script execution resumes only after character arrives

## Acceptance Criteria
await character.walk_to(\"door_left\") blocks until the character arrives. The next line runs only after arrival.

## Depends on
T26

## Notes
HIGH RISK. The blocking behaviour is the core prototype mechanism. Test with multiple sequential WalkTo calls." \
"$M5" "M5,risk:high"

create_issue \
"T28 — Implement FaceTo" \
"## Task
Implement face_to(point_name) on AGSCharacter as a blocking rotation method.

## Subtasks
- GDScript method face_to(point_name: String) on AGSCharacter
- Calculates facing direction from character position to named point
- Rotates character using a tween
- Awaits tween completion before returning

## Acceptance Criteria
await character.face_to(\"window\") rotates the character to face the named point. Script resumes only after rotation completes.

## Depends on
T27" \
"$M5" "M5,risk:low"

create_issue \
"T29 — Implement CharacterSpawnPoint node type" \
"## Task
Implement AGSSpawnPoint — places a named character at a room position on load.

## Subtasks
- C++ class AGSSpawnPoint extending Node3D
- spawn_character property: name of character to place on room load
- On room ready: find named character via AGSRuntime, move to spawn point world position

## Acceptance Criteria
Named character appears at spawn point when room loads. No manual positioning in script needed.

## Depends on
T25" \
"$M5" "M5,risk:low"

# ─── M6 ISSUES ────────────────────────────────────────────────────────────────

echo "Creating M6 issues..."

create_issue \
"T30 — Wire AGSScriptLanguage to generated GDScript" \
"## Task
Connect the Godot ScriptLanguage interface to the transpiler so attaching a .agscript triggers transpilation automatically.

## Subtasks
- AGSScriptLanguage::load() triggers parse + emit for the .agscript file if output is stale
- Returns a Script resource backed by the generated GDScript
- Godot attaches script to AGSRoom or AGSCharacter node normally
- Transpilation errors surface in the Godot editor output panel

## Acceptance Criteria
Attaching a .agscript to an AGSRoom in the editor triggers transpilation automatically. No manual ag build step needed from within the editor.

## Depends on
T03, T18

## Notes
HIGH RISK. This is the integration point between the scripting system and Godot runtime. Read Godot GDScript ScriptLanguage implementation as a reference." \
"$M6" "M6,risk:high"

create_issue \
"T31 — Implement AGSRuntime singleton" \
"## Task
Implement the AGSRuntime autoload singleton providing the global API surface for scripts.

## Subtasks
- Autoload singleton registered by agvm module on engine init
- Tracks all AGSRoom and AGSCharacter nodes in the scene tree via signals
- Provides: get_room(name), get_character(name), get_point(room_name, point_name)
- Ready before any room script runs

## Acceptance Criteria
From any script: AGSRuntime.get_character(\"player\") and AGSRuntime.get_point(\"market\", \"door_left\") return correct references.

## Depends on
T19, T25" \
"$M6" "M6,risk:med"

create_issue \
"T32 — Map AGS-spirit built-in names to runtime calls" \
"## Task
Implement the mapping table in the emitter that translates AGS-spirit standard library calls to AGSRuntime GDScript calls.

## Subtasks
- Define built-in name mapping table in the emitter
- character.WalkTo(point.x) maps to: await AGSRuntime.get_character(name).walk_to(point_name)
- character.FaceTo(point.x) maps to: await AGSRuntime.get_character(name).face_to(point_name)
- room.Enter handled via signal binding (T33)

## Acceptance Criteria
All AGS-spirit standard library calls in the prototype test script emit correct GDScript invoking AGSRuntime methods.

## Depends on
T15, T31

## Notes
HIGH RISK. Incorrect mappings produce GDScript that runs but does the wrong thing — subtle and hard to debug." \
"$M6" "M6,risk:high"

create_issue \
"T33 — Implement room script event binding" \
"## Task
Connect Godot node signals to AGS-spirit event handler functions at room load time.

## Subtasks
- On AGSRoom ready: find attached .agscript-derived GDScript
- Connect signals to event handler functions:
  - room_enter signal to func room_Enter()
  - hotspot_clicked(name) to func hotspot_Interact(name)
  - region_entered(char) to func region_WalkedInto(char)
  - region_exited(char) to func region_WalkedOff(char)

## Acceptance Criteria
room_Enter() in .agscript fires on room load. hotspot_Interact(\"painting\") fires when painting hotspot is clicked.

## Depends on
T30, T32" \
"$M6" "M6,risk:med"

create_issue \
"T34 — Error message routing through source maps" \
"## Task
Translate GDScript runtime error locations back to AGS-spirit source locations using source maps.

## Subtasks
- GDScript runtime errors intercepted by AGSScriptLanguage error handler
- Source map loaded for the offending generated GDScript file
- GDScript line number translated to AGS-spirit file and line
- Error displayed referencing .agscript source, not .engine/generated/ path

## Acceptance Criteria
A deliberate runtime error shows the author the correct .agscript filename and line number.

## Depends on
T17, T30" \
"$M6" "M6,risk:med"

create_issue \
"T35 — Author minimal test project" \
"## Task
Author the minimal AGS3D test project that exercises the full prototype stack.

## Subtasks
- game.agp manifest
- One room: WalkableSurface, two AGSPoints (door_left, window), one BlockerVolume
- One AGSCharacter (player) with AGSSpawnPoint
- One .agscript: room_Enter walks player to door_left then faces window

## Acceptance Criteria
Test project committed to repo. It is the canonical input for T36.

## Depends on
T19 — T29" \
"$M6" "M6,risk:low"

create_issue \
"T36 — Build and run end-to-end prototype" \
"## Task
Build and run the test project end to end. This task IS the prototype success criterion.

## Subtasks
- ag build compiles the test project with no errors
- ag run launches the project in Godot
- Character appears at spawn point
- Character walks to door_left on room load
- Character faces window after arriving
- No manual GDScript editing was required at any point

## Acceptance Criteria
Character walks to the named point and faces the second point as scripted, driven entirely by the .agscript file. The prototype is complete.

## Depends on
T30 — T35

## Notes
HIGH RISK. This is integration of all previous work. Allocate debugging time. When this passes the prototype is done." \
"$M6" "M6,risk:high"

# ─── DONE ─────────────────────────────────────────────────────────────────────

echo ""
echo "========================================="
echo " Done!"
echo " 6 milestones and 36 issues created."
echo " https://github.com/$REPO/issues"
echo "========================================="
