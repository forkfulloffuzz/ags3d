---
name: github-issues
description: "Use when updating, closing, or commenting on GitHub issues for completed, started, or blocked AGS3D prototype tasks (T01–T36). Use for marking tasks done, recording blockers, linking PRs to issues, or querying open work across milestones M1–M6."
argument-hint: "Task ID or description, e.g. 'T02 done' or 'T09 blocked — need grammar spec first'"
---

# AGS3D GitHub Issue Management

## Task → Issue Mapping

Each prototype task maps 1:1 to an issue. Issue numbers are sequential starting at #2:

| Task | Issue |
| ---- | ----- |
| T01  | #2    |
| T02  | #3    |
| T03  | #4    |
| T04  | #5    |
| T05  | #6    |
| T06  | #7    |
| T07  | #8    |
| T08  | #9    |
| T09  | #10   |
| T10  | #11   |
| T11  | #12   |
| T12  | #13   |
| T13  | #14   |
| T14  | #15   |
| T15  | #16   |
| T16  | #17   |
| T17  | #18   |
| T18  | #19   |
| T19  | #20   |
| T20  | #21   |
| T21  | #22   |
| T22  | #23   |
| T23  | #24   |
| T24  | #25   |
| T25  | #26   |
| T26  | #27   |
| T27  | #28   |
| T28  | #29   |
| T29  | #30   |
| T30  | #31   |
| T31  | #32   |
| T32  | #33   |
| T33  | #34   |
| T34  | #35   |
| T35  | #36   |
| T36  | #37   |

If the issue number is uncertain, confirm it with: `gh issue list --search "T0X" --state all`

## Workflows

### Task completed

1. Close the issue with a summary comment explaining what was done and how the acceptance criteria were met.

```
gh issue close <N> --comment "<What was built. How acceptance criteria were met. Any notable decisions.>"
```

### Task started / in progress

Add a comment to signal work has begun (do not close):

```
gh issue comment <N> --body "Work started. <Brief note on approach or current state.>"
```

### Task blocked

Add a comment with the blocker, then add a `blocked` label if it exists:

```
gh issue comment <N> --body "Blocked: <reason>. Depends on <task or external factor>."
gh issue edit <N> --add-label "blocked"
```

### Link a PR to an issue

When creating or editing a PR, use closing keywords in the PR body so GitHub auto-closes the issue on merge:

```
gh pr create --body "Closes #<N>\n\n<PR description>"
```

### Query open work

```
# All open issues in a milestone
gh issue list --label "M2" --state open

# All high-risk open issues
gh issue list --label "risk:high" --state open

# A specific task
gh issue view <N>
```

## Comment Quality Guidelines

- Always state **what** was done (files created/changed, approach taken)
- Always confirm **acceptance criteria** from the task doc were met
- For blockers, name the specific dependency and which task unblocks it
- Keep comments factual and brief — they are the project's audit trail

## Labels Used

| Label                                 | Meaning                                |
| ------------------------------------- | -------------------------------------- |
| `M1`–`M6`                             | Milestone                              |
| `risk:low` / `risk:med` / `risk:high` | Risk level from task doc               |
| `prototype`                           | All prototype tasks                    |
| `blocked`                             | Cannot proceed — waiting on dependency |
