# AGS3D — Project Skills

Reusable skill prompts for common development tasks. Invoke with `/skill-name` in Claude Code.

---

## /test

Run the full AGS3D test suite (GDScript headless + Go unit tests).

```
Run the full AGS3D test suite by executing `.dev/test-all.sh` from the project root.
Report the number of passing and failing tests. If any tests fail, show the full
failure output and identify which test file and method failed.
```

---

## /test-go

Run the Go CLI unit tests only.

```
Run Go tests for the ag CLI by executing `cd tools/ag && go test ./... -v` from
the project root. Report pass/fail counts. On failure show the full test output.
```

---

## /test-gd

Run the GDScript headless tests only.

```
Run the GDScript headless test suite by executing `.dev/test.sh` from the project
root. Report pass/fail counts per suite. On failure show the failing test names
and assertion messages.
```

---

## /build

Build the game prototype with the ag CLI.

```
Run `ag build game_prototype` to process all changed source files (.agroom,
.agchar, .agcut, .agscript) and regenerate .tscn scenes and GDScript under
game_prototype/.engine/generated/. Report which files were rebuilt and any
errors.
```

---

## /validate

Validate the game prototype project.

```
Run `ag validate game_prototype` to perform full cross-system validation:
character references, room references, goto targets, loc_key presence, item
flags, and cutscene localisation. Print all [ERROR] and [WARN] lines. Exit
status non-zero means blocking errors are present.
```

---

## /next-task

Show the next uncompleted task from TODO.md.

```
Read TODO.md and find the first task not marked [x]. Print the task ID, title,
and full description. Then state whether it has any unmet dependencies listed
in the Dependencies line at the top of the batch.
```

---

## /commit

Commit the current task following the project commit protocol.

```
Follow the standard commit protocol:
1. Run `git status` (never -uall) and `git diff` to review changes.
2. Run `git log --oneline -5` to match the commit style.
3. Stage only the files changed by the current task (never `git add -A`).
4. Write a commit message in the format: `type(scope): T-ID — short description`
   followed by a blank line and a one-paragraph body explaining what changed.
   End with: Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>
5. Commit, then run `git status` to confirm the working tree is clean.
```

---

## /pick-next-batch

Select the next 10 tasks from the active milestones.

```
The current TODO.md batch is complete. Analyse the active milestones (M9, M10,
M11) by reading their docs in docs/. Apply the ordering rules: critical-path
tasks first (unblock other tasks), then unblocked parallel work. Write the next
10 tasks to TODO.md in the same format as the previous batch, including a
Dependencies line and a Notes section. Then report the task list to the user.
```
