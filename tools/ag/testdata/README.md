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

---

## cutscenes/

### Valid

| File | What it exercises |
|------|-------------------|
| `01_minimal.agcut` | Title + sequence with `<<fade_in>>` and `<<end>>` |
| `02_full_cutscene.agcut` | Camera, music, title card, character, sync, action, skip policy |
| `03_parallel_and_if.agcut` | `<<parallel>>` block and `<<if flag>>` conditional |

### Invalid

Each invalid file has a header comment:
```
// EXPECT_ERROR: <description>
```

| File | Error | Layer |
|------|-------|-------|
| `err_01_missing_title.agcut` | Header has no `title:` field | Parser |

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
