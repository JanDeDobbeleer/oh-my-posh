---
name: conventional-commit
description: >
  Rules for writing Conventional Commits messages in this project. MUST be invoked every time a
  commit is created.
triggers:
  - on_commit
---

# Conventional commit

`.commitlintrc.yml` is the source of truth: use only the types it allows. Beyond the
specification:

- Keep the header (`type(scope): description`) at 72 characters or fewer.
- Write the description in the imperative, without a trailing period: `add`, `fix`, `bump`. Never
  copy past tense from the request (`added`, `fixed`, `bumped`).
- Use the affected area as the scope (`segment`, `cache`, `config`); omit it for cross-cutting
  changes.
- A change that removes, renames, or alters behavior callers depend on carries both `!` after the
  type or scope and a `BREAKING CHANGE:` footer. Never one without the other.
- Use the body for why, not what, wrapped at 72 characters.
- Review `git diff --cached` and stage files explicitly; never `git add -A`.

```text
feat(segment)!: rename template property StartTime to Start

BREAKING CHANGE: template strings using .StartTime must be updated to .Start
```
