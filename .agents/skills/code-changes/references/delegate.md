# Phase 3 — Delegate

Match the executor to the task, not the other way around. The point is cost and speed without
losing quality — the coordinator stays accountable for the result.

## Executor matrix

Tier definitions and per-vendor model examples: [model-tiers.md](model-tiers.md).

| Task profile                                             | Executor                                                       |
| ---------------------------------------------------------- | ------------------------------------------------------------ |
| Trivial: mechanical edit, config tweak, typo, doc update  | Do it directly, or batch on a trivial-tier model              |
| Standard, well-specified implementation                   | Coordinator itself, or a same-tier subagent for parallelism   |
| Hits an escalation trigger (see [escalate.md](escalate.md)) | Escalation-tier subagent, for that one specific question    |

- The coordinator typically runs at implementer tier itself, so standard tasks are usually
  executed directly rather than delegated. Delegate anyway when there's real parallelism —
  independent tasks, each in its own worktree — or to shed trivial mechanical work onto a
  cheaper model.
- Escalation is never the default path for "ambiguous" or "architectural" — it fires only on the
  concrete triggers in [escalate.md](escalate.md), and only for the specific question, not the
  whole task.

## What a delegation carries

Hand the subagent its full pinned spec from Phase 2, plus:

- The verification commands it must run and pass before reporting done.
- The instruction to report what it changed and what it verified — not just "done".
- The instruction to stop and report when it hits something the spec does not cover, instead of
  improvising scope.

## What is never delegated downward

Analysis, final verification, and delivery never go to an implementer — they stay with the
coordinator or go up to Escalation tier for a specific question:

- Analysis and root-cause work (Phase 1).
- Final verification (Phase 5) — an implementer's green run is a claim, not a result.
- Commits, history rewrites, pushes, and user-facing reporting (Phase 6) — these stay with the
  coordinator regardless of tier; they're about accountability, not capability, so they never go
  to Escalation tier either.
