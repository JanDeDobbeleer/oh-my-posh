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
[escalate.md](escalate.md).

## Documentation

Documentation changes ship in the same change as the code they describe. Check for every
user-visible behavior change:

- Project docs / website pages for the touched feature.
- README or setup instructions when flags, commands, or defaults changed.

## On failure

Pick the destination by what actually broke, not by default:

- Gate failure (build, test, lint) or a diff that doesn't match the spec → back to Phase 4
  (fix via the implementer). The spec was right; the execution wasn't.
- Functional proof contradicts the stated root cause, or the fix didn't change the observed
  behavior at all → back to Phase 1. The spec was built on a wrong diagnosis. Re-entering Phase 1
  re-arms its stop gate: report the revised analysis and wait for a go again before replanning,
  same as first entry — a Verify bounce is not a standing authorization to keep implementing
  unattended.

Never weaken a gate, skip a linter, or delete a test to get to green.

## Retry cap

There is no memory across turns other than what is written down, so the count must be carried as
text, not tracked mentally. Every time Verify fails, state explicitly in the report handed to the
next phase: the attempt number for this task, what broke, and the destination chosen. "The same
task" means the same originating user request — a revised diagnosis on a Phase 1 bounce is still
the same task; it does not reset the count.

On the second consecutive failure for that task — regardless of whether cycle 1 and cycle 2 went
to the same destination — stop cycling before sending it back a third time. Escalate the specific
question (why the fix isn't landing, or why the root cause keeps missing) to Escalation tier. See
[escalate.md](escalate.md).

If the cycle that follows the escalation answer *also* fails, do not escalate again and do not
send it back a fourth time. Stop the loop entirely and report to the user: the two prior attempts,
the escalation question and answer, and the latest failure evidence. Continuing to cycle past that
point means the workflow itself isn't converging on this task, and that decision belongs to the
user, not to another escalation call.
