# Phase 1 — Analyze

Start every task here, no matter how it arrived: issue link, PR number, verbal idea, bug report.
No code gets written or edited during this phase.

## Gather the full context

- Issues and PRs: `gh issue view <n> --comments` / `gh pr view <n> --comments`, plus linked
  issues, referenced discussions, and any code the report points at.
- Ideas and verbal requests: restate the goal and constraints in your own words. If the request
  is ambiguous, resolve the ambiguity now — not halfway through implementation.
- Check for prior art: existing helpers, similar segments/modules, and past commits that touched
  the same area (`git log -- <path>`).

## Reproduce before theorizing

Reproduce the problem when possible. A reproduction turns the analysis from a hypothesis into a
fact and gives Phase 5 its verification case for free. When reproduction is impossible (platform,
hardware, credentials), say so explicitly in the report and mark the fix as unverified-by-repro.

## Find the root cause in the code

- Read the actual implementation. Never reason from the issue text, a review comment, or a stack
  trace alone — reports and bot reviewers are frequently wrong.
- Distinguish the root cause from the symptom. Fixing where it crashes is not the same as fixing
  why it crashes.
- State what the change should be, which files it touches, and what it deliberately leaves alone.

## When to escalate

If root cause can't be pinned with confidence, or the fix looks architectural, security-sensitive,
or irreversible, hand the specific question to the strongest available model instead of guessing —
see [references/escalate.md](references/escalate.md). Resume ownership of the phase once the
question is answered.

## Output of this phase

A short analysis report to the user containing:

1. What is actually happening and why (root cause, with file references).
2. The proposed change and its scope.
3. What is intentionally out of scope.
4. Open questions, if any remain.

## Stop gate

Report the analysis and wait for a go before implementing. Skip the gate only when the user
already gave the go in the request itself ("do it", "fix it and commit", "implement with
Sonnet"). A go given for analysis is not a go for implementation.
