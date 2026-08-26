# Special case — Issue triage

Entry point when the task is bare triage — "look at issue #n", "triage #n", "is #n still valid" —
with no fix requested. This is Phase 1 with a sharper deliverable: the analysis itself is the
product; implementation only happens on an explicit go. If the request already asks for a fix
("fix issue #n"), use [analyze.md](analyze.md) directly instead — that phrasing already carries
the go past the stop gate, which this deliverable-only entry point does not.

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

This maps onto the same `root_cause` / `proposed_change` / `out_of_scope` / `repro_status` /
`open_questions` artifact Phase 1 produces — see [artifacts.md](artifacts.md) — so Plan can pick it
up unchanged.

## Gate

This is the Phase 1 stop gate: implement only on go, then continue from Phase 2.
