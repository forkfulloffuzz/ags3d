# AGS-Spirit Test Fixtures

Fixture files used by scanner, parser, emitter, dialogue, and cutscene tests.

## Structure

```
testdata/
  scripts/     — .agscript fixtures (scanner / parser / emitter)
    valid/
    invalid/
  rooms/       — .agroom fixtures
    valid/
    invalid/
  items/       — .agitem fixtures
    valid/
    invalid/
  dialogues/   — .agdlg fixtures (dialogue lexer / parser / linker / emitter)
    valid/
    invalid/
  cutscenes/   — .agcut fixtures (cutscene parser / validator)
    valid/
    invalid/
  locale/      — .agstrings fixtures (T-LOC02: parser / Diff / Apply)
    valid/
    invalid/
```

---

## scripts/

### Valid

| File | What it exercises |
|------|-------------------|
| `01_minimal.agscript` | Empty function — simplest possible file |
| `02_literals.agscript` | All literal kinds: int, float, string, bool, null |
| `03_all_types.agscript` | Every type keyword as var and param types; char/short compat |
| `04_arithmetic_expressions.agscript` | `+ - * / %`, unary `-`, `++`/`--`, compound assign |
| `05_comparison_logical.agscript` | `== != < <= > >=`, `&& \|\| !`, precedence ladder |
| `06_bitwise_operators.agscript` | `& \| ^ ~ << >>`, compound bitwise assign, flag patterns |
| `07_if_else.agscript` | if / else / else-if chains, deeply nested, complex conditions |
| `08_loops.agscript` | while, do..while, for (all optional-part combinations), break, continue |
| `09_switch.agscript` | switch/case/default, fall-through, string switch, complex bodies, enum discriminant |
| `10_functions.agscript` | All return types, params, early return, mutual recursion, recursive |
| `11_enums.agscript` | Explicit/implicit values, trailing comma, use in switch and var decl |
| `12_namespaces.agscript` | Private members, export function, cross-namespace calls, enum inside namespace |
| `13_global_namespace.agscript` | global.player, global.room, global.score, chained access, as argument |
| `14_event_handlers.agscript` | Every documented event handler name — room, hotspot, character, item, global |
| `15_blocking_calls.agscript` | Every documented blocking call; transitive blocking via user function |
| `16_member_access_and_calls.agscript` | Postfix chains: `.`, `()`, `[]`, combinations, call-as-argument |
| `17_operator_precedence.agscript` | Full precedence ladder, left-associativity, right-associativity, unary chains |
| `18_comments.agscript` | Comments in every position: inline, between tokens, spanning lines |
| `19_var_declarations.agscript` | Top-level, local, in blocks, expr init, call init, shared across functions |
| `20_realistic_room.agscript` | Full room script: enum, namespace, all event handlers, blocking calls, switch |
| `21_realistic_global_script.agscript` | game_start, repeatedly_execute, on_key_press, on_event, namespaces |
| `22_edge_cases.agscript` | Empty bodies, deeply nested exprs, assignment chains, postfix edge cases |

### Invalid

Each invalid file has a header comment:
```
// EXPECT_ERROR: <description>
```

| File | Error | Layer |
|------|-------|-------|
| `err_01_unterminated_string.agscript` | Unterminated string literal | Scanner / Parser |
| `err_02_unclosed_brace.agscript` | Unexpected EOF, expected `}` | Parser |
| `err_03_unclosed_paren.agscript` | Unexpected token, expected `)` | Parser |
| `err_04_missing_semicolon.agscript` | Unexpected token, expected `;` | Parser |
| `err_05_export_outside_namespace.agscript` | `export` outside namespace | Semantic (T10) |
| `err_06_invalid_token.agscript` | Unexpected character `@` | Scanner / Parser |
| `err_07_missing_function_body.agscript` | Unexpected token, expected `{` | Parser |
| `err_08_missing_if_condition.agscript` | Unexpected token, expected `(` | Parser |
| `err_09_bad_for_loop.agscript` | Missing `;` in for header | Parser |
| `err_10_switch_no_brace.agscript` | Unexpected token, expected `{` | Parser |
| `err_11_enum_missing_brace.agscript` | Unexpected token, expected `{` | Parser |
| `err_12_namespace_missing_name.agscript` | Expected identifier after `namespace` | Parser |
| `err_13_unterminated_block_comment.agscript` | Unexpected EOF in block comment | Scanner / Parser |
| `err_14_break_outside_loop.agscript` | `break` outside loop or switch | Semantic (T10) |
| `err_15_double_operator.agscript` | Missing operand between operators | Parser |
| `err_16_missing_return_value.agscript` | Bare `return` in non-void function | Semantic (T10) |
| `err_17_duplicate_export.agscript` | Duplicate export name in namespace | Semantic (T10) |
| `err_18_case_outside_switch.agscript` | `case` outside of switch | Parser |
| `err_19_function_inside_function.agscript` | Nested function declaration | Parser |
| `err_20_global_assigned.agscript` | Assignment to read-only namespace | Semantic (T10) |

---

## rooms/

### Valid

| File | What it exercises |
|------|-------------------|
| `01_minimal.agroom` | Bare room definition — required fields only |
| `02_full_room.agroom` | Hotspots, walkable areas, walk-behinds, scale zones, edges |

### Invalid

| File | Error |
|------|-------|
| `err_01_missing_room_name.agroom` | Room block with no `name:` field |
| `err_02_unclosed_brace.agroom` | Block opened but never closed |
| `err_03_bad_vector.agroom` | Vector field with wrong component count |

---

## items/

### Valid

| File | What it exercises |
|------|-------------------|
| `01_minimal.agitem` | Bare item definition — required fields only |
| `02_full_item.agitem` | All optional fields: description, sprite, combine target |

### Invalid

| File | Error |
|------|-------|
| `err_01_missing_name.agitem` | Item block with no `name:` field |
| `err_02_missing_keyword.agitem` | Content outside any block keyword |

---

## dialogues/

### Valid

| File | What it exercises |
|------|-------------------|
| `01_minimal.agdlg` | Single node, one speaker line, `<<end>>` |
| `02_options.agdlg` | `->` choice branches, nested option body |
| `03_loc_keys.agdlg` | Explicit `#loc:key` tags and auto-loc inference |

### Invalid

Each invalid file has a header comment:
```
// EXPECT_ERROR: <description>
```

| File | Error | Layer |
|------|-------|-------|
| `err_01_missing_title.agdlg` | Header has no `title:` field | Parser |
| `err_02_unclosed_command.agdlg` | `<<command` without closing `>>` | Lexer |
| `err_03_missing_separator.agdlg` | No `---` between header and body | Parser |

#### Identifier naming errors (DLG-E011)

| File | Error |
|------|-------|
| `err_04_invalid_title_uppercase.agdlg` | Node `title:` contains uppercase letters |
| `err_05_invalid_title_hyphen.agdlg` | Node `title:` contains a hyphen |
| `err_06_invalid_jump_target.agdlg` | `<<jump>>` target contains uppercase and dot |
| `err_07_title_starts_with_digit.agdlg` | Node `title:` starts with a digit |

---

## cutscenes/

Sequence body lines are indented with tabs for readability; the parser ignores all leading/trailing whitespace.

### Valid

| File | What it exercises |
|------|-------------------|
| `01_minimal.agcut` | Title + sequence with `<<fade_in>>` and `<<end>>` |
| `02_full_cutscene.agcut` | Camera, music, title card, character, sync, action, skip policy |
| `03_parallel_and_if.agcut` | `<<parallel>>` block and `<<if flag>>` conditional |
| `04_all_header_fields.agcut` | Every optional header field: `skip`, `tags`, `fallback`, `loc_group`, `voice_session` |
| `05_skip_policies.agcut` | `author_controlled` skip with multiple `<<label>>` and `<<skip_to>>` |
| `06_background_steps_and_sync.agcut` | Named `bg:id` steps, `<<sync id1 id2>>`, and bare `<<sync>>` |
| `07_if_else_chain.agcut` | `<<if>>` / `<<else_if>>` / `<<else>>` / `<<end_if>>` nested chain |
| `08_on_event_block.agcut` | `<<on event:char:tag>>` reactive handler block with `<<end_on>>` |
| `09_nested_cutscene.agcut` | `<<cutscene file:name>>` reference to an existing title |
| `10_inline_cutscene.agcut` | `<<cutscene skip:policy>>` ... `<<end_cutscene>>` inline block |
| `11_all_camera_commands.agcut` | All camera sub-commands: `set`, `move_to`, `look_at`, `follow`, `shake`, `fov`, `return` |
| `12_all_character_commands.agcut` | All character sub-commands: `spawn_at`, `walk_to`, `run_to`, `animation`, `face_to`, `expression`, `move_speed`, `hide`, `show` |
| `13_all_audio_commands.agcut` | All audio commands: `music`, `sound`, `ambient`, `voice`, stops |
| `14_all_visual_commands.agcut` | All visual commands: `fade_in/out`, `letterbox`, `vignette`, `flash`, `overlay`, `video` |
| `15_state_and_flow_commands.agcut` | `title_card`, `subtitle`, `line`, `wait`, `action`, `set` |
| `16_save_block_false_ambient.agcut` | `save_block:false` with no state-change commands (valid) |
| `17_audio_scope_pause.agcut` | `audio_scope:pause` — room audio paused at start, auto-resumed on end |
| `18_audio_scope_keep_with_room_channels.agcut` | `audio_scope:keep` with manual `channel:room_music` crossfade |
| `19_duck_header_defaults.agcut` | `auto_duck:true` with header duck config; per-call `duck_level:` override |
| `20_duck_per_call_overrides.agcut` | `auto_duck:false`; explicit `duck:channels` per line, `duck:none` suppression, `<<dialogue>>` duck |
| `21_duck_room_ambient_volume.agcut` | Manual `<<ambient volume channel:room_ambient>>` duck/restore without auto_duck |

### Invalid

Each invalid file has a header comment:
```
// EXPECT_ERROR: <error code> — description
```

#### Parser errors

| File | Error | Code |
|------|-------|------|
| `err_01_missing_title.agcut` | Header has no `title:` field | — |
| `err_13_empty_command.agcut` | `<<>>` empty command body | Parser |
| `err_14_bare_body_line.agcut` | Sequence body line is not a `<<command>>` | Parser |
| `err_15_malformed_header.agcut` | Header line missing `:` separator | Parser |

#### Cutscene format errors (CUT-E)

| File | Error | Code |
|------|-------|------|
| `err_02_duplicate_title.agcut` | Title already declared in another file | CUT-E001 |
| `err_03_unknown_named_point.agcut` | `point.nowhere` does not exist in any room | CUT-E002 |
| `err_04_unknown_character.agcut` | Character `ghost_npc` not defined | CUT-E003 |
| `err_05_unknown_audio.agcut` | Audio file `nonexistent_track` not found | CUT-E004 |
| `err_06_unknown_video.agcut` | Video file `missing_video` not found | CUT-E005 |
| `err_07_skip_to_missing_label.agcut` | `<<skip_to act2>>` — no `<<label act2>>` in sequence | CUT-E006 |
| `err_08_choice_in_parallel.agcut` | `<<choice>>` inside `<<parallel>>` block | CUT-E007 |
| `err_09_nested_cutscene_not_found.agcut` | `<<cutscene file:does_not_exist>>` — title unknown | CUT-E008 |
| `err_10_circular_nested_cutscene.agcut` | Cutscene references itself | CUT-E009 |
| `err_11_save_block_false_with_action.agcut` | `save_block:false` with `<<action>>` state change | CUT-E012 |
| `err_12_save_block_false_with_set.agcut` | `save_block:false` with `<<set>>` state change | CUT-E012 |

#### Sequencing errors (SEQ-E)

| File | Error | Code |
|------|-------|------|
| `err_16_seq_sync_undeclared_id.agcut` | `<<sync ghost_step>>` — id never declared as `bg:` | SEQ-E001 |
| `err_17_seq_sync_foreground_id.agcut` | `<<sync main_step>>` — id is a foreground step | SEQ-E002 |
| `err_18_seq_leaked_bg_step.agcut` | Background step `music_bg` has no eventual `<<sync>>` | SEQ-E003 |
| `err_19_seq_on_fail_jump_missing_label.agcut` | `on_fail:jump_to:fallback_scene` — label not in sequence | SEQ-E004 |
| `err_20_seq_duplicate_step_id.agcut` | Two steps both declare `bg:step_a` | SEQ-E007 |

#### Cutscene format warnings (CUT-W)

| File | Warning | Code |
|------|---------|------|
| `err_21_leaked_audio_music.agcut` | `<<music>>` started with no reachable stop | CUT-W009 |
| `err_22_leaked_audio_ambient.agcut` | `<<ambient>>` started with no reachable stop | CUT-W009 |
| `err_23_duck_all.agcut` | `duck:all` — channel set unverifiable at build time | CUT-W010 |
| `err_24_auto_duck_no_channels.agcut` | `auto_duck:true` with no `duck_channels` declared | CUT-W011 |

#### Identifier naming errors (CUT-E013)

| File | Error |
|------|-------|
| `err_25_invalid_title_uppercase.agcut` | `title:` contains uppercase letters |
| `err_26_invalid_title_spaces.agcut` | `title:` contains spaces |
| `err_27_invalid_label_name.agcut` | `<<label>>` name contains a hyphen |
| `err_28_invalid_bg_id.agcut` | `bg:` step id contains uppercase and dot |
| `err_29_title_starts_with_digit.agcut` | `title:` starts with a digit |

#### T-CUT30 loc_key fixtures

| File | What it exercises |
|------|------------------|
| `22_all_loc_commands.agcut` | All localizable commands: `<<line #loc:>>`, `<<title_card #loc:>>`, `<<subtitle #loc:>>`, `<<choice #loc:>>`; also `voice_session:` and `loc_group:` headers |
| `23_voice_session_lines.agcut` | `voice_session:` header with multiple `<<line>>` commands for `CollectVoiceLines` / voice coverage |

---

## locale/ — `.agstrings` format fixtures (T-LOC02)

### Valid

| File | What it exercises |
|------|------------------|
| `01_minimal.agstrings` | Required `[meta]` fields only (`base_locale`, `locale`), minimal `[strings]` block |
| `02_full_meta.agstrings` | All meta fields: `locale_name`, `rtl: true`, `fallback_chain` |
| `03_with_metadata.agstrings` | Per-entry metadata comments: `// type:`, `// char:`, `// scene:`, `// ctx:` (T-LOC11 format) |
| `04_with_stale.agstrings` | `// [stale]` orphan marker on a stale entry |
| `05_with_orphan.agstrings` | `// [orphan]` orphan marker on a removed key |

### Invalid

Each invalid file has a header comment:
```
// EXPECT_ERROR: <description>
```

| File | Error |
|------|-------|
| `err_01_missing_locale.agstrings` | `[meta]` block missing required `locale` field |
| `err_02_missing_base_locale.agstrings` | `[meta]` block missing required `base_locale` field |
| `err_03_duplicate_key.agstrings` | Duplicate key in `[strings]` block |
| `err_04_unknown_meta_field.agstrings` | Unknown field in `[meta]` block |
| `err_05_malformed_entry.agstrings` | Malformed key=value line in `[strings]` block |

---

## Using fixtures in tests

```go
// Load all valid .agscript fixtures and assert no parse errors
func TestParser_ValidFixtures(t *testing.T) {
    paths, _ := filepath.Glob("../../testdata/scripts/valid/*.agscript")
    for _, path := range paths {
        t.Run(filepath.Base(path), func(t *testing.T) {
            src, _ := os.ReadFile(path)
            _, errs := parser.Parse(path, string(src))
            if len(errs) != 0 {
                t.Errorf("unexpected errors: %v", errs)
            }
        })
    }
}

// Load invalid .agdlg fixtures and assert at least one error
func TestDialogue_InvalidFixtures(t *testing.T) {
    paths, _ := filepath.Glob("../../testdata/dialogues/invalid/*.agdlg")
    for _, path := range paths {
        t.Run(filepath.Base(path), func(t *testing.T) {
            src, _ := os.ReadFile(path)
            _, err := dlg.Parse(path, string(src))
            if err == nil {
                t.Errorf("expected error, got none")
            }
        })
    }
}
```
