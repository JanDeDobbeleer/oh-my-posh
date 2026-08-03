# Phase 2 — Plan

Turn the approved analysis into an executable plan. The quality bar: an implementer-tier model
must be able to execute each task without asking questions.

## Pin the spec

Write the spec down before delegating anything. It contains:

- The decided approach — including decisions already made, so the implementer does not relitigate
  them.
- Files to touch, and the entry points to start from.
- Constraints: style rules, patterns to follow (point at existing code), performance or
  compatibility requirements.
- Verification commands the implementer must run locally (build, tests, lint).
- Explicit non-goals: what the task must NOT change. This is what keeps subagents from wandering.

## Split into tasks

- One task = one self-contained unit an implementer can finish and verify on its own.
- Mark which tasks are independent and which consume another task's output. Parallelize the
  independent ones; sequence the rest. When in doubt, sequence — a merge conflict between two
  parallel subagents costs more than the parallelism saves.
- Documentation updates belong to the task that changes the behavior, not to a separate task.

## Decide the workspace per task

- Main working tree: when the task depends on uncommitted local changes, or when you will review
  and commit the result in the current session.
- Isolated worktree: everything else, especially parallel tasks — they must never share a
  working tree.

## Output of this phase

A task list where each entry names its executor tier (see Phase 3), its workspace, its
dependencies, and carries its pinned spec.
