# Special case — Pull request review comments

Entry point when the task is "handle the review comments on PR #n". This replaces Phase 1's
issue analysis; Phases 2–6 apply unchanged, including the stop gate, the merged-state Verify run,
and Deliver's push policy — this special case does not opt out of any of them.

For each valid comment, the classification below (valid/invalid, with the code evidence) stands in
for the full analysis-report artifact from [artifacts.md](artifacts.md): `root_cause` /
`proposed_change` is the valid-comment list with its code evidence, `out_of_scope` is the invalid
comments (named explicitly, not silently dropped), `repro_status` is the code-path confirmation
used to classify each comment, and `open_questions` is empty once every thread is classified. Plan
still pins a spec from this and the stop gate still applies before any fixup is written, even when
the fix itself is small.

## Validate before touching anything

- Fetch every unresolved comment (`gh pr view <n> --comments`, or the review threads via
  `gh api`).
- Validate each comment against the actual code before changing anything. Classify it as valid
  or invalid. Automated reviewers (Copilot and friends) regularly flag non-issues — treat their
  comments as leads, not verdicts.

## Valid comments

- Plan each fix as a task targeting the commit that owns the code (Phase 2), then apply it as a
  `git commit --fixup <sha>` on top of the branch. The fixup commit is the working state; it is
  not yet the delivered change.
- `git rebase --autosquash` to fold each fixup into the commit it targets.
- Confirm the rewritten tree is byte-identical to the pre-rebase tree plus the intended fix
  (`git diff` between old and new tip). A fixup that changes anything else went to the wrong
  commit.
- Run the full quality gates and functional proof (Phase 5) on the rebased tree — the same "on
  the merged/final state, never per-branch" rule from [verify.md](verify.md) applies here, with
  the rebased tree standing in for the merged state. Only after this passes is the task done.
- Pushing (including force-push) follows Deliver's push policy unchanged: only when the user
  asked for it, and with `--force-with-lease` when pushing a rewritten branch. Handling PR review
  comments does not by itself imply a push — confirm the user wants the branch updated remotely.

## Invalid comments

- Do not change code to appease a wrong comment.
- Reply with the evidence that refutes it: the code path, the existing test, the verified
  behavior. Specific beats polite-but-vague.
- When a comment is wrong but exposes something genuinely confusing, harden the code or comment
  against the misreading instead — and say so in the reply.

## Reply to every thread

Each thread gets a reply describing what was done and how it was verified. Leave resolving the
threads to the humans.
