# AGS-Spirit Language Grammar

Formal specification for the AGS-spirit scripting language. This document is the authoritative source for the lexer (T07), parser (T08–T09), and emitter (T13–T17).

Derived from the official AGS scripting language — see [AGS Manual](https://adventuregamestudio.github.io/ags-manual/) for the source language reference, specifically:
- [ScriptKeywords](https://adventuregamestudio.github.io/ags-manual/ScriptKeywords.html)
- [ScriptingLanguage](https://adventuregamestudio.github.io/ags-manual/ScriptingLanguage.html)
- [BlockingScripts](https://adventuregamestudio.github.io/ags-manual/BlockingScripts.html)
- [Character API](https://adventuregamestudio.github.io/ags-manual/Character.html)

## Design Decisions (Divergences from AGS)

| Area | Real AGS | AGS-Spirit | Reason |
|------|----------|------------|--------|
| String type | `String` (capital, managed) | `string` (lowercase keyword) | Simpler single string type |
| Player ref | `player` magic global | `global.player` | Explicit namespaced global state |
| Game globals | scattered magic vars | `global.*` namespace | All engine-owned state in one place |
| Coordinates | `Walk(x, y)` raw coords | `WalkTo(point.NAME)` named points | No 3D coords exposed to authors |
| Hotspot prefix | `hName_Event` | `hotspot_NAME_Event` | Clearer, no single-letter prefix |
| Item prefix | `iName_Event` | `item_NAME_Event` | Same |
| Type system | class inheritance | structural typing | See §Type System |
| Cross-file visibility | `import`/`export` headers | `namespace` + `export` | See §Visibility and Name Resolution |
| Preprocessor | `#define`, `#ifdef`, etc. | Not supported | Out of scope |
| `::` scope / extender functions | Yes | Not supported | No author-defined structs in prototype |

---

## Notation

Extended BNF (EBNF):

```
x y        — sequence
x | y      — alternation
[ x ]      — optional (zero or one)
{ x }      — repetition (zero or more)
( x )      — grouping
"x"        — literal terminal
NAME       — non-terminal
name       — lexical terminal (defined in §Lexical Structure)
```

---

## Lexical Structure

### Source Encoding

Source files are UTF-8. Filenames carry the `.agscript` extension.

### Whitespace and Comments

Whitespace (spaces, tabs, carriage returns, newlines) is ignored between tokens.

```
LineComment  = "//" { any_char_except_newline } newline
BlockComment = "/*" { any_char } "*/"
```

Block comments do not nest.

### Literals

```
digit        = "0" … "9"
letter       = "a" … "z" | "A" … "Z" | "_"

IntLit       = digit { digit }
FloatLit     = digit { digit } "." { digit }
StringLit    = '"' { StringChar } '"'
StringChar   = any_char_except_'"'_and_newline | EscapeSeq
EscapeSeq    = "\" ( '"' | "\" | "n" | "t" | "r" )
BoolLit      = "true" | "false"
NullLit      = "null"
```

### Identifiers and Keywords

```
Ident        = letter { letter | digit }
```

Reserved keywords — cannot be used as identifiers:

```
bool        break       case        char
continue    default     do          else
enum        export      false       float
for         function    global      if
int         namespace   null        return
short       string      String      switch
true        void        while
```

Notes:
- `char` and `short` are accepted as type names and mapped to `int` semantically — they exist for compatibility with AGS scripts.
- `String` (capital S) is accepted as a synonym for `string`.
- `global` is a reserved namespace identifier, not a general-purpose keyword (see §Global Namespace).
- `namespace` and `export` control cross-file visibility (see §Visibility and Name Resolution).

### Operators and Punctuation

```
+    -    *    /    %    !    ~
++   --
==   !=   <    <=   >    >=
&&   ||
&    |    ^    <<   >>
=    +=   -=   *=   /=   %=   &=   |=   ^=   <<=  >>=
(    )    {    }    [    ]
,    ;    .
```

---

## Grammar

### File

```
File        = { TopDecl }

TopDecl     = FunctionDecl
            | EnumDecl
            | VarDecl ";"
```

### Declarations

```
FunctionDecl  = [ Type ] "function" Ident "(" [ ParamList ] ")" Block

ParamList     = Param { "," Param }
Param         = Type Ident

Type          = "int" | "short" | "char"    — all map to int
              | "float"
              | "bool"
              | "string" | "String"          — both accepted, same type
              | "void"
              | Ident                        — user-defined / engine type

EnumDecl      = "enum" Ident "{" EnumBody "}" ";"
EnumBody      = EnumMember { "," EnumMember } [ "," ]
EnumMember    = Ident [ "=" Expr ]
```

### Statements

```
Block         = "{" { Stmt } "}"

Stmt          = VarDecl ";"
              | IfStmt
              | WhileStmt
              | DoWhileStmt
              | ForStmt
              | SwitchStmt
              | ReturnStmt ";"
              | BreakStmt ";"
              | ContinueStmt ";"
              | ExprStmt ";"

VarDecl       = Type Ident [ "=" Expr ]

IfStmt        = "if" "(" Expr ")" Block [ "else" ( IfStmt | Block ) ]

WhileStmt     = "while" "(" Expr ")" Block

DoWhileStmt   = "do" Block "while" "(" Expr ")" ";"

ForStmt       = "for" "(" ForInit ";" [ Expr ] ";" [ Expr ] ")" Block
ForInit       = VarDecl | Expr | ε

SwitchStmt    = "switch" "(" Expr ")" "{" { CaseClause } "}"
CaseClause    = ( "case" Expr | "default" ) ":" { Stmt }

ReturnStmt    = "return" [ Expr ]
BreakStmt     = "break"
ContinueStmt  = "continue"

ExprStmt      = Expr
```

### Expressions

Precedence from lowest (top) to highest (bottom):

```
Expr              = AssignExpr

AssignExpr        = LogicOrExpr [ AssignOp AssignExpr ]   — right-associative
AssignOp          = "=" | "+=" | "-=" | "*=" | "/=" | "%="
                  | "&=" | "|=" | "^=" | "<<=" | ">>="

LogicOrExpr       = LogicAndExpr { "||" LogicAndExpr }

LogicAndExpr      = BitOrExpr { "&&" BitOrExpr }

BitOrExpr         = BitXorExpr { "|" BitXorExpr }

BitXorExpr        = BitAndExpr { "^" BitAndExpr }

BitAndExpr        = EqualExpr { "&" EqualExpr }

EqualExpr         = RelExpr { ( "==" | "!=" ) RelExpr }

RelExpr           = ShiftExpr { ( "<" | "<=" | ">" | ">=" ) ShiftExpr }

ShiftExpr         = AddExpr { ( "<<" | ">>" ) AddExpr }

AddExpr           = MulExpr { ( "+" | "-" ) MulExpr }

MulExpr           = UnaryExpr { ( "*" | "/" | "%" ) UnaryExpr }

UnaryExpr         = ( "!" | "-" | "~" | "++" | "--" ) UnaryExpr
                  | PostfixExpr

PostfixExpr       = PrimaryExpr { PostfixSuffix }
PostfixSuffix     = "++"                         — post-increment
                  | "--"                          — post-decrement
                  | "." Ident                     — member access
                  | "(" [ ArgList ] ")"           — call
                  | "[" Expr "]"                  — index

PrimaryExpr       = Ident
                  | Literal
                  | "(" Expr ")"

ArgList           = Expr { "," Expr }

Literal           = IntLit | FloatLit | StringLit | BoolLit | NullLit
```

---

## Event Handler Naming Convention

Event handlers are ordinary functions called by the engine by name at runtime.

### Room Events

| Handler | When called |
|---------|-------------|
| `room_Load` | Room loads (every visit) |
| `room_FirstLoad` | Room loads for the first time only |
| `room_AfterFadeIn` | After fade-in completes |
| `room_Leave` | Player is leaving the room |
| `room_RepExec` | Every game loop tick while in this room |

### Hotspot Events

| Handler | When called |
|---------|-------------|
| `hotspot_NAME_Look` | Player examined hotspot |
| `hotspot_NAME_Interact` | Player used / clicked hotspot |
| `hotspot_NAME_WalkOn` | Player walked onto hotspot |
| `hotspot_NAME_WalkOff` | Player walked off hotspot |
| `hotspot_NAME_AnyClick` | Any mouse button on hotspot |
| `hotspot_NAME_MouseOver` | Mouse cursor over hotspot |

### Character Events

| Handler | When called |
|---------|-------------|
| `character_NAME_Talk` | Character begins talking |
| `character_NAME_Idle` | Character becomes idle |

### Inventory Item Events

| Handler | When called |
|---------|-------------|
| `item_NAME_Look` | Player examined item |
| `item_NAME_Use` | Player used item |
| `item_NAME_UsedWith` | Player used item with another |

### Global Script Events

| Handler | Signature | When called |
|---------|-----------|-------------|
| `game_start` | `function game_start()` | Once at game startup |
| `repeatedly_execute` | `function repeatedly_execute()` | Every game loop (blockable) |
| `repeatedly_execute_always` | `function repeatedly_execute_always()` | Every game loop (never blocked) |
| `on_event` | `function on_event(int event, int data)` | General game events |
| `on_key_press` | `function on_key_press(int key, int mod)` | Keyboard input |
| `on_mouse_click` | `function on_mouse_click(int button)` | Mouse input |

---

## Blocking Calls — Semantic Annotation

Blocking calls are syntactically ordinary function calls. The transpiler annotates them during symbol resolution (T11) and emits `await` in the generated GDScript (T16).

A call is blocking if:
1. It is a call to one of the **built-in blocking functions** listed below, **or**
2. It is a call to a user-defined function whose body contains (directly or transitively) a blocking call.

A function that contains any blocking call is marked `IsBlocking = true` on its `FunctionDecl` node. The emitter prepends `await` to every call site of a blocking function.

### Character Methods (Blocking by Default)

All of the following block execution until the action completes. In real AGS most accept an optional `eNoBlock` parameter — AGS-spirit always blocks.

| AGS-spirit call | AGS equivalent | Semantics |
|-----------------|---------------|-----------|
| `character.WalkTo(point)` | `Character.Walk(x,y)` | Walk to named point |
| `character.WalkStraight(point)` | `Character.WalkStraight(x,y)` | Walk in straight line (may go off walkable area) |
| `character.Say(text)` | `Character.Say(msg)` | Speech bubble; always blocking |
| `character.Think(text)` | `Character.Think(msg)` | Thought bubble; always blocking |
| `character.PlayAnimation(name)` | `Character.Animate(loop, delay)` | Play animation; wait until complete |
| `character.FaceDirection(dir)` | `Character.FaceDirection(dir)` | Rotate to face direction |
| `character.FaceCharacter(other)` | `Character.FaceCharacter(char)` | Rotate to face another character |
| `character.FacePoint(point)` | `Character.FaceLocation(x,y)` | Rotate to face a named point |
| `character.RunInteraction(mode)` | `Character.RunInteraction(mode)` | Run interaction event |

### Global Blocking Functions

| AGS-spirit call | AGS equivalent | Semantics |
|-----------------|---------------|-----------|
| `Wait(frames)` | `Wait(time)` | Pause N game loops |
| `WaitKey(timeout)` | `WaitKey(timeout)` | Wait for keypress |
| `WaitMouse(timeout)` | `WaitMouse(timeout)` | Wait for mouse click |
| `WaitInput(timeout)` | `WaitMouseKey(timeout)` | Wait for any input |
| `FadeIn(speed)` | `FadeIn(speed)` | Screen fade in |
| `FadeOut(speed)` | `FadeOut(speed)` | Screen fade out |
| `DisplayMessage(text)` | `Display(msg)` | Modal message; wait for dismiss |

---

## Built-in Types

| AGS-Spirit type | AGS equivalent | Description | Godot mapping |
|----------------|---------------|-------------|---------------|
| `int` | `int` | 32-bit signed integer | `int` |
| `short` | `short` | 16-bit signed (maps to int) | `int` |
| `char` | `char` | 8-bit value (maps to int) | `int` |
| `float` | `float` | 32-bit float | `float` |
| `bool` | `bool` | Boolean | `bool` |
| `string` / `String` | `String` | Text string | `String` |
| `void` | `void` | No return value | `void` |
| `Character` | `Character*` | Character node ref | `AGSCharacter` |
| `Room` | none direct | Room node ref | `AGSRoom` |
| `Point` | `(x,y)` pair | Named 3D point + orientation | `Vector3` pair |

---

## Global Namespace

`global` is a reserved namespace that exposes engine-owned game state. It is not a variable — it cannot be assigned to, passed as a value, or redefined. All properties are read-only from script unless documented otherwise.

```agscript
global.player       // current player character (Character) — reflects SetAsPlayer()
global.room         // current room (Room)
global.score        // current score (int) — write via GiveScore()
global.camera       // active camera
```

`global.player` replaces AGS's magic `player` variable. The name makes it obvious this is engine state, not a local or imported value. It always reflects whatever the engine considers the current player character — if a cutscene temporarily switches control, `global.player` updates accordingly.

---

## Type System — Structural Typing

AGS-spirit does **not** use class inheritance. Types are matched by **structure** — a function that accepts a `Point` accepts anything that has the same shape, regardless of its declared type. This is Go's interface model applied to the scripting layer.

**Practical meaning for authors:**

A function that expects a position argument will accept a `Point`, a `Character` (uses its current position), or any other type the engine exposes that has spatial properties. The author does not need to cast or wrap values.

```agscript
// This accepts anything that can be used as a destination
function walkPlayerTo(Point dest) {
    global.player.WalkTo(dest);
}

// These all work — all have the required spatial shape:
walkPlayerTo(point.door_left)         // explicit named point
walkPlayerTo(character.guard)         // character's current position
```

**For the transpiler:** structural compatibility is checked at build time (T11). The emitter does not emit casts — the GDScript output relies on Godot's duck-typed runtime for the same flexibility.

**No plans for full OOP.** Author-defined classes with inheritance are not in scope. If authors need to group data, `enum` and the built-in engine types cover the prototype. More complex grouping will be addressed in post-prototype language design.

---

## Visibility and Name Resolution

### Rules at a glance

| Where defined | Callable from |
|---------------|---------------|
| File-level `function` | That file only |
| `namespace X { function }` | Within namespace X (any file contributing to X) |
| `namespace X { export function }` | Anywhere, as `X.FuncName()` |
| Engine event handlers | Always implicitly accessible to the engine |

### File-scoped by default

Functions defined at the top level of a file, outside any namespace, are **private to that file**. There is no way to call them from another file. This prevents accidental collisions without any boilerplate.

```agscript
// Only callable within this file
function helperCalc(int x) { return x * 2; }
```

### Namespaces

A `namespace` block groups related functions under a named scope. Multiple files can contribute to the same namespace — the transpiler merges them.

```agscript
// characters/utils.agscript
namespace CharUtils {

    // Private to this namespace — callable by other CharUtils functions
    // across any file, but NOT callable as CharUtils.validate() from outside
    function validateItem(string name) { ... }

    // Exported — callable from anywhere as CharUtils.AddSpecialItem(...)
    export function AddSpecialItem(Character target, string itemName) {
        if (validateItem(itemName)) {
            // ...
        }
    }

    export function RemoveItem(Character target, string itemName) { ... }
}
```

Calling exported namespace members:

```agscript
// From any other .agscript file — no import needed
CharUtils.AddSpecialItem(global.player, "magic_sword");
```

### `export` is only valid inside a namespace

Using `export` at the file level is a transpiler error. Cross-file sharing always requires a named namespace so it is clear where the function comes from.

### Conflict detection

The transpiler errors at build time if:
- Two files define `export function` with the same name inside the same namespace
- A call to a namespace-private function is made from outside the namespace
- A file-level function is called from a different file

### Engine event handlers

Event handlers (`room_Load`, `hotspot_NAME_Interact`, etc.) are **always implicitly accessible to the engine** — they do not need `export` and do not need to be in a namespace. The engine calls them by name directly.

### Grammar

```ebnf
TopDecl       = NamespaceDecl | FunctionDecl | EnumDecl | VarDecl ";"

NamespaceDecl = "namespace" Ident "{" { NamespaceMember } "}"
NamespaceMember = [ "export" ] FunctionDecl
                | EnumDecl
                | VarDecl ";"

FunctionDecl  = [ Type ] "function" Ident "(" [ ParamList ] ")" Block
```

---

## Named Points

Authors never write 3D coordinates. All spatial references use the named-point system:

```
point.NAME          — named point defined in the .agroom file
character.NAME      — named character defined in a .agchar file
```

Examples:
```agscript
global.player.WalkTo(point.door_left)
global.player.FacePoint(point.npc_guard)
```

---

## Grammar — Complete Reference (single block)

```ebnf
File          = { TopDecl }
TopDecl       = NamespaceDecl | FunctionDecl | EnumDecl | VarDecl ";"

NamespaceDecl   = "namespace" Ident "{" { NamespaceMember } "}"
NamespaceMember = [ "export" ] FunctionDecl | EnumDecl | VarDecl ";"

FunctionDecl  = [ Type ] "function" Ident "(" [ ParamList ] ")" Block
ParamList     = Param { "," Param }
Param         = Type Ident
Type          = "int" | "short" | "char" | "float" | "bool"
              | "string" | "String" | "void" | Ident

EnumDecl      = "enum" Ident "{" EnumMember { "," EnumMember } [ "," ] "}" ";"
EnumMember    = Ident [ "=" Expr ]

Block         = "{" { Stmt } "}"
Stmt          = VarDecl ";" | IfStmt | WhileStmt | DoWhileStmt | ForStmt
              | SwitchStmt | ReturnStmt ";" | BreakStmt ";" | ContinueStmt ";"
              | ExprStmt ";"
VarDecl       = Type Ident [ "=" Expr ]
IfStmt        = "if" "(" Expr ")" Block [ "else" ( IfStmt | Block ) ]
WhileStmt     = "while" "(" Expr ")" Block
DoWhileStmt   = "do" Block "while" "(" Expr ")" ";"
ForStmt       = "for" "(" [ ForInit ] ";" [ Expr ] ";" [ Expr ] ")" Block
ForInit       = VarDecl | Expr
SwitchStmt    = "switch" "(" Expr ")" "{" { CaseClause } "}"
CaseClause    = ( "case" Expr | "default" ) ":" { Stmt }
ReturnStmt    = "return" [ Expr ]
BreakStmt     = "break"
ContinueStmt  = "continue"
ExprStmt      = Expr

Expr          = AssignExpr
AssignExpr    = LogicOrExpr [ AssignOp AssignExpr ]
AssignOp      = "=" | "+=" | "-=" | "*=" | "/=" | "%=" | "&=" | "|=" | "^=" | "<<=" | ">>="
LogicOrExpr   = LogicAndExpr { "||" LogicAndExpr }
LogicAndExpr  = BitOrExpr { "&&" BitOrExpr }
BitOrExpr     = BitXorExpr { "|" BitXorExpr }
BitXorExpr    = BitAndExpr { "^" BitAndExpr }
BitAndExpr    = EqualExpr { "&" EqualExpr }
EqualExpr     = RelExpr { ( "==" | "!=" ) RelExpr }
RelExpr       = ShiftExpr { ( "<" | "<=" | ">" | ">=" ) ShiftExpr }
ShiftExpr     = AddExpr { ( "<<" | ">>" ) AddExpr }
AddExpr       = MulExpr { ( "+" | "-" ) MulExpr }
MulExpr       = UnaryExpr { ( "*" | "/" | "%" ) UnaryExpr }
UnaryExpr     = ( "!" | "-" | "~" | "++" | "--" ) UnaryExpr | PostfixExpr
PostfixExpr   = PrimaryExpr { "++" | "--" | "." Ident | "(" [ ArgList ] ")" | "[" Expr "]" }
PrimaryExpr   = Ident | Literal | "(" Expr ")"
ArgList       = Expr { "," Expr }
Literal       = IntLit | FloatLit | StringLit | BoolLit | NullLit
```

---

## Worked Example

```agscript
// rooms/start/start.agscript

enum Direction {
    eNorth = 0,
    eSouth,
    eEast,
    eWest
}

function room_Load() {
    int greeting = 1;
    if (greeting == 1) {
        global.player.Say("Hello, world.");
    }
}

function hotspot_door_Interact() {
    global.player.WalkTo(point.door_left);
    global.player.FacePoint(point.door_outside);
    FadeOut(30);
}

function room_RepExec() {
    for (int i = 0; i < 3; i++) {
        Wait(10);
    }
}
```

Transpiled GDScript (`.engine/generated/rooms/start/start.agscript.gd`):

```gdscript
enum Direction { eNorth = 0, eSouth = 1, eEast = 2, eWest = 3 }

func room_Load():
    var greeting: int = 1
    if greeting == 1:
        await character_ego.say("Hello, world.")

func hotspot_door_Interact():
    await character_ego.walk_to(point_door_left)
    await character_ego.face_point(point_door_outside)
    await AGSRuntime.fade_out(30)

func room_RepExec():
    for i in range(3):
        await AGSRuntime.wait(10)
```
