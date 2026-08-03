# Phase 4 — Supervise

Delegation is not fire-and-forget. The coordinator tracks delivery and owns the outcome.

## Monitor and unblock

- Track each subagent's progress against its spec.
- When a subagent stalls or loops on a problem: stop it, diagnose the problem yourself, hand it
  the answer, and let it proceed. Do not let it burn turns rediscovering what you already know.
- When a subagent reports a spec gap, decide — update the spec or cut the scope — and send it
  back with the decision. Never let it decide scope on its own.

## Review the output critically

Review every subagent diff as if it were an external PR:

- Check the diff against the spec: everything asked for, nothing beyond it.
- Override solutions that are wrong or overbuilt. Prefer the change that removes code over the
  one that adds it. It is normal to keep a subagent's diagnosis but replace its fix with a
  simpler one — document the override and its reason for the final report.
- Watch for spec-compliant-but-ugly: a change can satisfy the letter of the spec and still not
  belong in the codebase. Consistency with surrounding code wins.

## Escalate on low confidence

If a diff leaves you unsure whether the fix is correct or merely plausible, or it touches
security, data-migration, or otherwise irreversible territory, get a second read from the
strongest available model before signing off — see
[references/escalate.md](references/escalate.md). Don't rubber-stamp a diff you can't fully
verify yourself.

## Trust nothing unverified

"Tests pass" from a subagent is a claim, not a result. Phase 5 re-verifies everything
independently, on the merged state — not per-task.
