# Escalation triggers

The coordinator owns every phase by default. Escalation is a bounded subagent call to the
strongest available model for one specific judgment call — not a handoff of the phase. The
coordinator frames the question, gets the answer, and stays the one accountable for the result.

Escalate when one of these is actually true, not on a general feeling that the task is hard:

- Root cause can't be pinned with confidence after actually reading the code — not after
  re-reading the report again.
- The change is architectural: it crosses a module boundary, touches a public API/interface, or
  introduces a cross-cutting abstraction.
- The code is security-, auth-, crypto-, payments-, or data-migration-sensitive.
- The operation is irreversible or high-blast-radius (schema migration, deletion, force-push,
  production config).
- An implementer has reported a spec gap or contradiction more than once on the same task. Track
  this the same way as the Verify retry count — state it explicitly in the report each time,
  since there is no memory across turns besides what was written down.
- Verify has sent the same task back a second consecutive time (to Supervise or to Analyze) —
  see the retry cap in [verify.md](verify.md). Two failed cycles means the diagnosis isn't
  converging, not that the third attempt will. This trigger fires once per task, for that second
  failure only — if the post-escalation cycle fails again, verify.md's terminal step applies
  instead: stop looping and report to the user rather than escalating a second time.
- Reviewing a diff leaves the coordinator unsure whether the fix is correct or merely plausible.
- The user explicitly asked for a second opinion or an adversarial review.

None of these fire on routine work. Most tasks should complete without ever calling Escalation
tier — the coordinator is capable enough to do the analysis, review, and verification itself, and
the strongest model gets paid for only on the calls that actually need it.

## How to escalate

Frame the specific question, not the whole task. Hand over the pinned context needed to answer
it — the relevant code, the hypothesis so far, why it's uncertain — get the answer, and resume
the phase. Escalation never becomes the new owner of the task, and it never decides scope; it
answers the question it was asked and returns control to the coordinator.
