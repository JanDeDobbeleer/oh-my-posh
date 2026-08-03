# Phase 6 — Deliver

The change is verified; now package it.

## Commits

- Use the conventional-commit skill for every commit message.
- One logical unit per commit. A feature and its lint fallout can be separate commits when they
  answer different "why"s.
- Stage files explicitly — never `git add -A`.
- Review the staged diff before committing, especially after auto-fixing tools rewrote files.

## Push and PR policy

Do not push, force-push, or open a PR unless the user asked for it. "Commit" means commit;
nothing more. When a push is asked for on a rewritten branch, use `--force-with-lease`.

## The final report

Outcome first, then evidence. It contains:

1. What changed, with clickable file references, and why — one paragraph before any detail.
2. Where implementer output was overridden, and the reason.
3. Verification evidence: the gates that ran and the concrete functional results observed.
4. Loose ends explicitly left to the user (secrets to delete, manual validation steps, decisions
   deferred), each with the exact command or check when applicable.

Never report an unverified step as done, and never bury a failure in the middle of a success
story.
