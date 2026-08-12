# Phase 5 — Verify

Verification is never delegated downward and runs on the final, merged state of the change. It
has two halves: the project's quality gates, and functional proof. Both are the coordinator's own
work by default.

## Quality gates

- Build, full test suite, formatters, and linters — all must pass with zero errors.
- Re-read the language and framework skills pinned in Phase 2 and review the final diff against
  them by hand, then run whatever pre-commit gate each one defines. Linters do not cover control
  flow, test structure, logging, or comment conventions — a green linter is not evidence the skill
  was followed.
- Cross-compile or re-lint for every target platform when platform-specific files changed — the
  local toolchain skips the other platform's rules. The language skill names how those files are
  marked.

## Functional proof

Tests passing is necessary, not sufficient. Run the real flow and confirm concrete outputs:
render the prompt, execute the command, hit the endpoint. Record the actual values observed —
the final report quotes them as evidence, not adjectives.

Exercise the flow the way the user will reach it, not only the way the harness reaches it: the
built artifact rather than the dev server, a cold start rather than the running process, the entry
point they actually open. When a dev server or file watcher is in play, restart it before judging
behavior — a stale bundle produces confident, wrong verification.

When the user said they will do the manual validation, state exactly what they should check and
what the expected result is.

## Escalate on high-stakes results

Running the gates and the functional proof stays with the coordinator. If the result is
ambiguous, or the change is high-blast-radius (migrations, security, irreversible operations), get
the strongest available model to judge the evidence before declaring done — see
[references/escalate.md](references/escalate.md).

## Documentation

Documentation changes ship in the same change as the code they describe. Check for every
user-visible behavior change:

- Project docs / website pages for the touched feature.
- README or setup instructions when flags, commands, or defaults changed.

## On failure

A failed gate or a wrong functional result sends the task back to Phase 4 (fix via the
implementer) or Phase 1 (the analysis was wrong). Never weaken a gate, skip a linter, or delete
a test to get to green.
