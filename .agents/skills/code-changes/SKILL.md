---
name: code-changes
description: >
  Orchestration workflow for any task that ends in code changes: issue analysis, pull request
  review, feature implementation, bug fixes, refactors, or fleshing out an idea. MUST be invoked
  at the start of such a task, before reading or writing any code. Defines how to analyze first,
  gate on user approval, plan, pick the right executor model, delegate and supervise subagents,
  verify, and deliver.
---

# Code Changes

The workflow for going from an issue, pull request, idea, or feature request to shipped code.
Follow the phases in order. Analysis always comes first; code comes last.

Each phase has a reference file with the full instructions. **Read the reference file when you
enter the phase** — not before, and never skip it because the phase "looks obvious".

## Roles

Assign work based on model capability. The coordinator is not the strongest model — it's the one
that stays resident, owns every phase by default, and knows when it's out of its depth.

| Role        | Capability tier                                               | Owns                                                                          |
| ----------- | -------------------------------------------------------------- | ------------------------------------------------------------------------------ |
| Coordinator | Capable mid-tier, resident for the whole task                | Analysis, plan, delegation, supervision, verification, delivery — by default |
| Escalation  | Strongest reasoning model available; invoked only on trigger | The specific judgment call the coordinator flagged, then control returns     |
| Implementer | Same tier as coordinator, or smaller for trivial edits        | Executing one pinned, self-contained task                                    |

Concrete model names per vendor (Anthropic, OpenAI, Google) and how to run the split in Claude
Code, GitHub Copilot, Codex, or IDE agents: [references/model-tiers.md](references/model-tiers.md).

When the current agent already runs at implementer tier, there's no separate delegation step for
standard work: plan and execute directly, still following every phase. Escalate to the strongest
model only when a trigger fires — see [references/escalate.md](references/escalate.md) — never by
default and never for a routine judgment call the coordinator is equipped to make itself.

## The flow

1. **Analyze** ([references/analyze.md](references/analyze.md)) — root cause and scope,
   validated against the code, never against the report alone.
2. **Plan** ([references/plan.md](references/plan.md)) — pinned spec, task split, parallel vs
   sequential, workspace per task.
3. **Delegate** ([references/delegate.md](references/delegate.md)) — match each task to the
   right executor, coordinator-tier by default.
4. **Supervise** ([references/supervise.md](references/supervise.md)) — monitor, unblock, and
   critically review implementer output.
5. **Verify** ([references/verify.md](references/verify.md)) — quality gates plus functional
   proof, never delegated downward.
6. **Deliver** ([references/deliver.md](references/deliver.md)) — conventional commits and an
   outcome-first report.

Analyze, Supervise, and Verify each carry an escalation checkpoint — see
[references/escalate.md](references/escalate.md) — for handing one specific judgment call to the
strongest available model without giving up ownership of the phase.

**Stop gate:** Phase 1 ends with reporting the analysis and proposed approach to the user and
waiting for a go. Skip the gate only when the user already gave the go in the request itself
("do it", "fix it and commit", "implement with Sonnet").

## Special cases

These entry points replace or extend Phase 1; the rest of the flow applies unchanged.

- Pull request review comments →
  [references/pr-review-comments.md](references/pr-review-comments.md)
- Issue triage → [references/issue-triage.md](references/issue-triage.md)
