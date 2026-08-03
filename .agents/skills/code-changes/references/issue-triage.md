# Special case — Issue triage

Entry point when the task is "look at issue #n". This is Phase 1 with a sharper deliverable:
the analysis itself is the product; implementation only happens on an explicit go.

## Steps

1. `gh issue view <n> --comments` — read the full report, every comment, and linked issues.
2. Reproduce the reported behavior. When reproduction needs an environment you lack (OS, shell,
   font, hardware), state that and reason from the code instead — flagged as such.
3. Locate the root cause in the code, not in the issue text. Issue reports describe symptoms and
   often guess wrong about causes.
4. Assess blast radius: who else is affected, since when (which release or commit introduced
   it), and whether workarounds exist.

## Deliverable

An analysis report to the user:

1. Confirmed or could-not-reproduce, with evidence.
2. Root cause, with file references.
3. Proposed fix and its scope, or the reason no fix is warranted (works-as-intended, duplicate,
   environment problem).
4. Suggested reply to the issue when the finding should be communicated upstream.

## Gate

This is the Phase 1 stop gate: implement only on go, then continue from Phase 2.
