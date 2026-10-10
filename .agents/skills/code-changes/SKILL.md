---
name: code-changes
description: >
  Workflow for any task that ends in code changes: issue analysis or triage, pull request review comments, features,
  bug fixes, refactors. Invoke at the start of such a task, before reading or writing code.
---

# Code changes

## 1. Analyze, then stop

- Read the code, not only the issue, review comment, or stack trace: reports and bot reviewers are often wrong. Fetch
  context with `gh issue view <n> --comments` or `gh pr view <n> --comments`.
- Reproduce the problem when possible; say so when you cannot.
- Look for what already exists before designing anything: a helper, a similar module, a pattern another feature uses,
  a CLI to call instead of reimplementing it. Prefer the smallest change that reuses existing code, and the change that
  removes code over the one that adds it. Deprecated logic gets removed, not maintained.
- Report the root cause with file references, the proposed change and its scope, what stays out of scope, and open
  questions.
- **Stop and wait for a go.** Skip the stop only when the request already asks for the implementation ("fix it",
  "do it"). A go for analysis is not a go for implementation, and a new diagnosis after a failed verification needs a
  new go.

For triage only ("look at #n"), the report is the deliverable: add the blast radius and a suggested reply to the issue.

## 2. Plan

- Load the language and format skills for every file you will touch before writing any code, and follow them from the
  first line. A rule read at review time turns a violation into rework.
- When replacing working code, list what the old code does (defaults, key bindings, error messages) and keep all of it.
  Dropping any of it needs a go first.
- For a judgment call you cannot make with confidence (unclear root cause, security, irreversible change), ask the
  strongest available model that one question, then continue.

## 3. Implement

- Do standard work yourself. Delegate only independent tasks that can run in parallel (one worktree each) or
  mechanical batches for a cheaper model.
- A delegation carries the decided approach, the files, the binding skill rules, the verification commands, the
  non-goals, and the instruction to stop and report anything the spec does not cover.
- Never delegate analysis, final verification, commits, or pushes.

## 4. Review the diff

Review every diff as an external pull request, your own included:

- Everything asked for, nothing more. Replace overbuilt solutions with simpler ones.
- Consistency with the surrounding code beats a new pattern.
- Cut tests that assert implementation details, copy data declared elsewhere in the diff, duplicate another test, or
  would still pass with the bug present.
- Delete comments that restate the code.

## 5. Verify

- Build, tests, formatters, and linters pass with zero findings. Never weaken a gate or delete a test to get there.
- Prove it works: run the real flow the way the user reaches it (the built binary, a cold start) and record the actual
  output.
- Update documentation in the same change as the behavior it describes.
- After two failed verification rounds on the same task, stop and report instead of trying a third time.

## 6. Deliver

- Commit only when the user asks for it. Push or open a pull request only when asked; push a rewritten branch with
  `--force-with-lease`.
- Write messages with the `conventional-commit` skill. One logical change per commit; stage files explicitly.
- Fold a fix to a commit that is not on the default branch into that commit: `git commit --fixup <sha>`, then
  `git rebase --autosquash`. That includes lint, test, and CI fallout. Never stack "fix" commits on top.
- Report the outcome first: what changed and why, the verification evidence, what is left for the user, and anything
  removed.

## Pull request review comments

1. Fetch every unresolved thread and validate each against the code. Treat bot comments as leads, not verdicts.
2. Valid: fix it as a fixup into the commit that introduced the code, autosquash, verify, and force-push the pull
   request branch.
3. Invalid: do not change code to appease it. Reply with the evidence. When the misreading exposes confusing code,
   clarify the code and say so.
4. Reply to every thread with the conclusion and the commit that addresses it, then resolve the thread.
