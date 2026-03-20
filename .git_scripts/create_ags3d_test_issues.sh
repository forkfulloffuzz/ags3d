#!/bin/bash

# AGS3D — Testing Tasks GitHub Issues creator
# Run after create_ags3d_issues.sh (milestones must already exist)
# Requires: gh auth login

REPO="forkfulloffuzz/ags3d"

echo "========================================="
echo " AGS3D Testing Tasks Setup"
echo " Repo: $REPO"
echo "========================================="
echo ""

# Milestone titles — must match exactly what was created by create_ags3d_issues.sh
M1="M1 — Godot Fork Setup"
M2="M2 — AGS-Spirit Parser"
M3="M3 — GDScript Emitter"
M4="M4 — Room Node"
M5="M5 — Character Node"
M6="M6 — End-to-End Wiring"

# ─── LABELS ───────────────────────────────────────────────────────────────────

echo "Creating testing labels..."

gh label create "testing"     --color "0D7377" --description "Test coverage task"         --repo $REPO 2>/dev/null || echo "  label testing already exists"
gh label create "ci"          --color "0D7377" --description "CI / automation"             --repo $REPO 2>/dev/null || echo "  label ci already exists"
gh label create "unit-test"   --color "14A085" --description "Unit test"                   --repo $REPO 2>/dev/null || echo "  label unit-test already exists"
gh label create "integration" --color "148A72" --description "Integration test"            --repo $REPO 2>/dev/null || echo "  label integration already exists"

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
    --label "testing" \
    --label "prototype"
  sleep 0.6
}

# ─── INFRASTRUCTURE (no milestone — repo-wide) ────────────────────────────────

echo "Creating test infrastructure issues..."

gh issue create \
  --repo "$REPO" \
  --title "TEST-INFRA-01 — Set up test runner and base class" \
  --body "## Task
Create the test infrastructure that all milestone test suites build on.

## Subtasks
- Create \`tests/\` directory at repo root
- Create \`tests/run_tests.gd\` — master runner that instantiates all suites, collects results, exits with code 0 or 1
- Create \`tests/utils/test_base.gd\` — base class with assert_eq, assert_true, assert_not_null, assert_no_crash helpers
- Create \`tests/utils/reporter.gd\` — collects pass/fail counts, prints summary, returns exit code
- Verify headless execution works: \`./bin/godot.linuxbsd.editor.x86_64 --headless --script tests/run_tests.gd\` exits 0 with empty suite list

## Acceptance Criteria
Running the test runner headlessly with no test suites registered exits with code 0 and prints a summary line.

## Notes
This must be done before any milestone test tasks. All other test tasks depend on this." \
  --label "testing,ci,unit-test,prototype" \
  --label "prototype"
sleep 0.6

gh issue create \
  --repo "$REPO" \
  --title "TEST-INFRA-02 — Set up CI pipeline with GitHub Actions" \
  --body "## Task
Add a GitHub Actions workflow that builds AGS3D and runs the full test suite on every push and pull request.

## Subtasks
- Create \`.github/workflows/test.yml\`
- Install all Godot build dependencies in CI
- Build with \`scons platform=linuxbsd\`
- Run tests with \`--headless --script tests/run_tests.gd\`
- Fail the workflow if exit code is non-zero
- Cache SCons build objects to speed up subsequent runs

## Acceptance Criteria
A push to the repo triggers the workflow. Tests run headlessly. A deliberate failing test causes the workflow to report failure.

## Acceptance Criteria — workflow file
\`\`\`yaml
name: AGS3D Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Install dependencies
        run: |
          sudo apt install -y build-essential scons pkg-config \\
            libx11-dev libxcursor-dev libxinerama-dev \\
            libgl1-mesa-dev libglu1-mesa-dev libasound2-dev \\
            libpulse-dev libfreetype6-dev libudev-dev \\
            libxi-dev libxrandr-dev libwayland-dev
      - name: Build
        run: scons platform=linuxbsd
      - name: Test
        run: ./bin/godot.linuxbsd.editor.x86_64 --headless --script tests/run_tests.gd
\`\`\`

## Depends on
TEST-INFRA-01" \
  --label "testing,ci,prototype" \
  --label "prototype"
sleep 0.6

echo ""

# ─── M1 TESTS ─────────────────────────────────────────────────────────────────

echo "Creating M1 test issues..."

create_issue \
"TEST-M1-01 — Write module and ScriptLanguage tests (5 tests)" \
"## Task
Write the M1 test suite covering module initialisation and ScriptLanguage registration.

## Tests to implement
- UT-M1-01: Engine boots headlessly without crash
- UT-M1-02: AGSScriptLanguage registered with ScriptServer (check by iterating languages, find name AGSScript)
- UT-M1-03: .agscript recognised as valid extension (ScriptServer.is_language_for_extension)
- UT-M1-04: create_script() returns non-null AGSScript instance
- UT-M1-05: AGSScript.get_language() returns AGSScriptLanguage singleton

## Files to create
- \`tests/m1_module/test_script_language.gd\`
- Register suite in \`tests/run_tests.gd\`

## Acceptance Criteria
All 5 tests pass when run headlessly. Suite is registered in run_tests.gd.

## Depends on
TEST-INFRA-01, T03 (AGSScriptLanguage implemented)" \
"$M1" "M1,unit-test"

# ─── M2 TESTS ─────────────────────────────────────────────────────────────────

echo "Creating M2 test issues..."

create_issue \
"TEST-M2-01 — Write lexer tests and fixtures (4 tests)" \
"## Task
Write tests covering the AGS-spirit lexer/tokenizer.

## Tests to implement
- UT-M2-01: Keywords tokenised correctly (function, if, while, return)
- UT-M2-02: Line and column tracked correctly across multi-line input
- UT-M2-03: String literals tokenised with correct value
- UT-M2-04: Comments skipped, identifier after comment tokenised

## Files to create
- \`tests/m2_parser/test_lexer.gd\`
- Fixture inputs as inline strings (no fixture files needed for lexer)

## Acceptance Criteria
All 4 tests pass. AGSScriptLanguage exposes a tokenise(source: String) method returning Array of token objects with type, value, line, col properties.

## Depends on
TEST-INFRA-01, T07 (lexer implemented)" \
"$M2" "M2,unit-test"

create_issue \
"TEST-M2-02 — Write parser AST tests and fixtures (5 tests)" \
"## Task
Write tests verifying the parser produces correct AST structure.

## Tests to implement
- UT-M2-05: Empty function produces FunctionDecl node
- UT-M2-06: Event handler produces EventHandler node
- UT-M2-07: If statement produces IfStmt node with correct condition
- UT-M2-08: While statement produces WhileStmt node
- UT-M2-09: Member expression (character.WalkTo) produces correct nested AST

## Files to create
- \`tests/m2_parser/test_parser.gd\`
- \`tests/m2_parser/fixtures/if_stmt.agscript\`
- \`tests/m2_parser/fixtures/while_stmt.agscript\`
- \`tests/m2_parser/fixtures/member_call.agscript\`

## Acceptance Criteria
All 5 tests pass. AGSScriptLanguage exposes a parse(source: String) method returning an AST result object with a node_type property on each node.

## Depends on
TEST-INFRA-01, T09 (parser implemented)" \
"$M2" "M2,unit-test"

create_issue \
"TEST-M2-03 — Write symbol table and blocking annotation tests (4 tests)" \
"## Task
Write tests for symbol resolution and blocking call annotation.

## Tests to implement
- UT-M2-10: Undefined reference produces error in error list
- UT-M2-11: Valid function call resolves with zero errors
- UT-M2-12: WalkTo call annotated as blocking=true
- UT-M2-13: Non-blocking call annotated as blocking=false

## Files to create
- \`tests/m2_parser/test_symbols.gd\`
- \`tests/m2_parser/test_blocking.gd\`
- \`tests/m2_parser/fixtures/undefined_ref.agscript\`
- \`tests/m2_parser/fixtures/valid_call.agscript\`

## Acceptance Criteria
All 4 tests pass.

## Depends on
TEST-INFRA-01, T10, T11" \
"$M2" "M2,unit-test"

create_issue \
"TEST-M2-04 — Write error handling tests (2 tests)" \
"## Task
Write tests verifying the parser handles errors gracefully.

## Tests to implement
- UT-M2-14: Malformed input produces non-empty error list, no crash
- UT-M2-15: Error objects contain line number > 0

## Files to create
- \`tests/m2_parser/test_errors.gd\`
- \`tests/m2_parser/fixtures/syntax_error.agscript\`

## Acceptance Criteria
Both tests pass. Parser never crashes on any input in the fixture set.

## Depends on
TEST-INFRA-01, T12" \
"$M2" "M2,unit-test"

# ─── M3 TESTS ─────────────────────────────────────────────────────────────────

echo "Creating M3 test issues..."

create_issue \
"TEST-M3-01 — Write emitter golden file tests (7 tests)" \
"## Task
Write emitter tests using golden file comparison for all emit types.

## Tests to implement
- UT-M3-01: Empty function emits correct GDScript func
- UT-M3-02: Event handler maps to correct GDScript name
- UT-M3-03: If statement emits correctly (golden file)
- UT-M3-04: While statement emits correctly (golden file)
- UT-M3-05: Assignment emits correctly
- UT-M3-06: Member call emits correctly
- UT-M3-07: Blocking call emits await

## Files to create
- \`tests/m3_emitter/test_emit_statements.gd\`
- \`tests/m3_emitter/fixtures/if_stmt.agscript\` + \`if_stmt.gd\` (golden)
- \`tests/m3_emitter/fixtures/while_stmt.agscript\` + \`while_stmt.gd\` (golden)
- Golden file update script: \`tests/utils/update_goldens.gd\`

## Acceptance Criteria
All 7 tests pass. Golden files committed to repo. Update script regenerates them correctly.

## Depends on
TEST-INFRA-01, T15" \
"$M3" "M3,unit-test"

create_issue \
"TEST-M3-02 — Write await emission and source map tests (6 tests)" \
"## Task
Write tests for blocking call await emission and source map correctness.

## Tests to implement
- UT-M3-08: Async function emits correctly
- UT-M3-09: Nested blocking calls emit correct await chain (golden file)
- UT-M3-10: Source map line count matches GDScript line count
- UT-M3-11: Source map maps known AGS-spirit line to correct GDScript line
- UT-M3-12: Output file written to .engine/generated/ at correct path
- UT-M3-13: Error message references .agscript path not .gd path

## Files to create
- \`tests/m3_emitter/test_emit_await.gd\`
- \`tests/m3_emitter/test_sourcemaps.gd\`
- \`tests/m3_emitter/fixtures/nested_blocking.agscript\` + golden

## Acceptance Criteria
All 6 tests pass. Source map verification is automated, not manual.

## Depends on
TEST-INFRA-01, T16, T17, T18" \
"$M3" "M3,unit-test"

# ─── M4 TESTS ─────────────────────────────────────────────────────────────────

echo "Creating M4 test issues..."

create_issue \
"TEST-M4-01 — Write Room node and geometry tests (7 tests)" \
"## Task
Write tests for AGSRoom node and all logic geometry node types.

## Tests to implement
- UT-M4-01: AGSRoom instantiates, is Node3D subclass
- UT-M4-02: room_name property reads and writes correctly
- UT-M4-03: WalkableSurface bakes valid navmesh on scene load
- UT-M4-04: BlockerVolume causes pathfinding to route around it
- UT-M4-05: AGSPoint self-registers with parent AGSRoom
- UT-M4-06: get_point() returns correct Vector3 position
- UT-M4-07: get_point() with unknown name returns null without crash

## Files to create
- \`tests/m4_room/test_room_node.gd\`
- \`tests/m4_room/test_walkable.gd\`
- \`tests/m4_room/test_blocker.gd\`
- \`tests/m4_room/test_points.gd\`
- \`tests/m4_room/scenes/test_room_basic.tscn\`
- \`tests/m4_room/scenes/test_room_with_point.tscn\`
- \`tests/m4_room/scenes/test_room_with_blocker.tscn\`

## Acceptance Criteria
All 7 tests pass headlessly. Navmesh tests use await process_frame to allow navigation server to process.

## Depends on
TEST-INFRA-01, T19 — T22" \
"$M4" "M4,unit-test"

create_issue \
"TEST-M4-02 — Write region and hotspot tests (5 tests)" \
"## Task
Write tests for TriggerRegion signals and HotspotSurface raycast detection.

## Tests to implement
- UT-M4-08: TriggerRegion fires region_entered when body overlaps
- UT-M4-09: TriggerRegion fires region_exited when body leaves
- UT-M4-10: TriggerRegion registers name with parent AGSRoom
- UT-M4-11: HotspotSurface fires hotspot_clicked with correct name on raycast hit
- UT-M4-12: Two hotspots in same room — ray at hotspot_b fires hotspot_b not hotspot_a

## Files to create
- \`tests/m4_room/test_regions.gd\`
- \`tests/m4_room/test_hotspots.gd\`
- \`tests/m4_room/scenes/test_room_with_region.tscn\`
- \`tests/m4_room/scenes/test_room_two_hotspots.tscn\`

## Acceptance Criteria
All 5 tests pass. Region tests use a simple CharacterBody3D moved programmatically to trigger enter/exit.

## Depends on
TEST-INFRA-01, T23, T24" \
"$M4" "M4,unit-test"

# ─── M5 TESTS ─────────────────────────────────────────────────────────────────

echo "Creating M5 test issues..."

create_issue \
"TEST-M5-01 — Write character registration and navigation tests (6 tests)" \
"## Task
Write tests for AGSCharacter registration and navigation behaviour.

## Tests to implement
- UT-M5-01: AGSCharacter instantiates without crash
- UT-M5-02: AGSCharacter registers with AGSRuntime by character_name
- UT-M5-03: Two characters with different names both retrievable from AGSRuntime
- UT-M5-04: Character moves closer to target after walk_to called (2 second timeout)
- UT-M5-05: WalkTo blocks — arrived flag not set until character within 0.5m of target
- UT-M5-06: WalkTo routes around BlockerVolume

## Files to create
- \`tests/m5_character/test_character_node.gd\`
- \`tests/m5_character/test_navigation.gd\`
- \`tests/m5_character/test_walkto.gd\`
- \`tests/m5_character/scenes/test_nav_room.tscn\`

## Acceptance Criteria
All 6 tests pass. Blocking tests use frame-stepping loop with 5 second timeout. Timeout causes test failure with descriptive message.

## Depends on
TEST-INFRA-01, T25 — T27" \
"$M5" "M5,unit-test"

create_issue \
"TEST-M5-02 — Write FaceTo and SpawnPoint tests (5 tests)" \
"## Task
Write tests for FaceTo blocking rotation and CharacterSpawnPoint placement.

## Tests to implement
- UT-M5-07: Sequential WalkTo calls — character visits A before B
- UT-M5-08: FaceTo rotates character toward named point (check forward dot product > 0.95)
- UT-M5-09: FaceTo blocks until rotation complete (flag set after, not before)
- UT-M5-10: CharacterSpawnPoint places character at correct position on room load
- UT-M5-11: SpawnPoint with unknown character name — scene loads without crash

## Files to create
- \`tests/m5_character/test_faceto.gd\`
- \`tests/m5_character/test_spawnpoint.gd\`
- \`tests/m5_character/scenes/test_spawn_room.tscn\`

## Acceptance Criteria
All 5 tests pass. FaceTo test checks forward vector dot product against direction to target point.

## Depends on
TEST-INFRA-01, T28, T29" \
"$M5" "M5,unit-test"

# ─── M6 TESTS ─────────────────────────────────────────────────────────────────

echo "Creating M6 test issues..."

create_issue \
"TEST-M6-01 — Write script wiring and runtime API tests (5 tests)" \
"## Task
Write integration tests for the script language wiring and AGSRuntime API.

## Tests to implement
- UT-M6-01: Attaching .agscript to AGSRoom triggers transpilation — .gd appears in .engine/generated/
- UT-M6-02: AGSRuntime.get_room() returns correct room node
- UT-M6-03: AGSRuntime.get_character() returns correct character node
- UT-M6-04: AGSRuntime.get_point() returns correct Vector3
- UT-M6-05: room_Enter() event handler fires when scene loads

## Files to create
- \`tests/m6_integration/test_script_wiring.gd\`
- \`tests/m6_integration/test_runtime.gd\`
- \`tests/m6_integration/scenes/test_wired_room.tscn\` (AGSRoom with .agscript attached)
- \`tests/m6_integration/scripts/test_room_enter.agscript\`

## Acceptance Criteria
All 5 tests pass. Script wiring test verifies generated .gd file exists at expected path after scene load.

## Depends on
TEST-INFRA-01, T30, T31, T33" \
"$M6" "M6,integration"

create_issue \
"TEST-M6-02 — Write event binding and error routing tests (4 tests)" \
"## Task
Write integration tests for event handler binding and error message routing through source maps.

## Tests to implement
- UT-M6-06: hotspot_Interact() fires with correct hotspot name when hotspot clicked
- UT-M6-07: region_WalkedInto() fires when character enters region
- UT-M6-08: Runtime error message contains .agscript path not .gd path
- UT-M6-09: Runtime error message contains correct line number from .agscript

## Files to create
- \`tests/m6_integration/test_event_binding.gd\`
- \`tests/m6_integration/test_error_routing.gd\`
- \`tests/m6_integration/scripts/test_hotspot.agscript\`
- \`tests/m6_integration/scripts/deliberate_error.agscript\` (error on known line)

## Acceptance Criteria
All 4 tests pass. Error routing test verifies the error string contains the .agscript filename.

## Depends on
TEST-INFRA-01, T33, T34" \
"$M6" "M6,integration"

create_issue \
"TEST-M6-03 — Write end-to-end prototype tests (3 tests)" \
"## Task
Write the end-to-end integration tests that encode the prototype success criterion as automated assertions.

## Tests to implement
- UT-M6-10: character.WalkTo in .agscript drives character to destination
- UT-M6-11: character.FaceTo in .agscript drives character rotation
- UT-M6-12: Full prototype — script drives walk to door_left then face window (prototype success criterion)

## Files to create
- \`tests/m6_integration/test_end_to_end.gd\`
- Uses the canonical test project from T35 as scene input

## Acceptance Criteria
UT-M6-12 passing IS the prototype success criterion as an automated test. Player walks to door_left (within 0.5m), then faces window (dot product > 0.95). 10 second timeout with descriptive failure messages.

## Depends on
TEST-INFRA-01, T35, T36

## Notes
This is the most important test in the suite. When UT-M6-12 passes in CI, the prototype is done and verifiably reproducible on any machine." \
"$M6" "M6,integration"

# ─── DONE ─────────────────────────────────────────────────────────────────────

echo ""
echo "========================================="
echo " Done!"
echo " 2 infra issues + 13 test issues created."
echo " https://github.com/$REPO/issues"
echo "========================================="
