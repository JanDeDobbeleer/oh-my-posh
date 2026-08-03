# Special case — Pull request review comments

Entry point when the task is "handle the review comments on PR #n". This replaces Phase 1's
issue analysis; Phases 2–6 apply unchanged.

## Validate before touching anything

- Fetch every unresolved comment (`gh pr view <n> --comments`, or the review threads via
  `gh api`).
- Validate each comment against the actual code before changing anything. Classify it as valid
  or invalid. Automated reviewers (Copilot and friends) regularly flag non-issues — treat their
  comments as leads, not verdicts.

## Valid comments

- Fix the issue folded into the commit that owns the code: `git commit --fixup <sha>`, then
  `git rebase --autosquash`.
- Confirm the rewritten tree is byte-identical to the pre-rebase tree plus the intended fix
  (`git diff` between old and new tip). A fixup that changes anything else went to the wrong
  commit.
- Run the quality gates (Phase 5) before rewriting history, not after.
- Force-push with `--force-with-lease`.

## Invalid comments

- Do not change code to appease a wrong comment.
- Reply with the evidence that refutes it: the code path, the existing test, the verified
  behavior. Specific beats polite-but-vague.
- When a comment is wrong but exposes something genuinely confusing, harden the code or comment
  against the misreading instead — and say so in the reply.

## Reply to every thread

Each thread gets a reply describing what was done and how it was verified. Leave resolving the
threads to the humans.
