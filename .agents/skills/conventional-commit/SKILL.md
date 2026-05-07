---
name: conventional-commit
description: >
  Workflow for generating conventional commit messages following the Conventional Commits
  specification. MUST be invoked every time a commit is created. Guides construction of
  standardized commit messages with correct type, scope, description, body, and footer.
triggers:
  - on_commit
---

# Conventional Commit

## Commit Message Structure

```text
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

### Types

| Type       | When to use                                            |
| ---------- | ------------------------------------------------------ |
| `feat`     | A new feature                                          |
| `fix`      | A bug fix                                              |
| `docs`     | Documentation changes, no code                         |
| `style`    | Formatting, missing semicolons, etc. (no logic change) |
| `refactor` | Code change that is neither a fix nor a feature        |
| `perf`     | Performance improvement                                |
| `test`     | Adding or correcting tests                             |
| `ci`       | CI configuration changes                               |
| `chore`    | Maintenance tasks (updating deps, tooling, etc.)       |
| `revert`   | Reverts a previous commit                              |

Append `!` after the type/scope to signal a **breaking change**: `feat!:` or `feat(api)!:`
When a change breaks existing behavior, **both markers are mandatory**: the `!` suffix on the type **and** the `BREAKING CHANGE:` footer. They always appear together — never one without the other.

### Scope

Optional. Use the name of the area affected, e.g., `segment`, `cache`, `config`, `ui`.
Omit when the change is truly cross-cutting.

### Description

- Required. One short imperative sentence, no period at the end. The full header line (type + scope + description) must be **72 characters or fewer**. Aim for **50 characters or fewer for the description itself** — this almost always keeps the full header within budget regardless of type and scope length.
- Use the imperative mood: "add", not "added" or "adds". Never past tense or present-third-person: ✗ `added`, `fixed`, `bumped`, `implemented` → ✓ `add`, `fix`, `bump`, `implement`.
- **Never mirror the input's phrasing.** If the request uses past-tense words (`updated`, `added`, `bumped`, `was removed`, `got regenerated`), convert them to imperative before writing the description: `update`, `add`, `bump`, `remove`, `regenerate`.

### Body

Optional. Add context about _why_ the change was made, not _what_. The diff shows that.
Wrap at 72 characters.

### Footer

Use for:

- `BREAKING CHANGE: <description>` (required when `!` is used; explains the break).
- Issue references: `Closes #123`, `Fixes #456`.
- Co-authors: `Co-Authored-By: Name <email>`.

## Workflow

1. Run `git status` to review changed files.
2. Run `git diff` and `git diff --cached` to inspect staged and unstaged changes.
3. Identify the **type** from the table above. Ask yourself: does this change **remove, rename, or alter existing behavior** that callers depend on? If yes → it is a breaking change: use `!` after the type/scope **and** add a `BREAKING CHANGE:` footer. Both markers are always required together.
4. Identify the **scope** from the files/area changed.
5. Write a short **description** in the imperative mood.
6. Add a **body** if the _why_ needs explanation.
7. Add a **footer** for breaking changes or issue references.
8. Stage the relevant files explicitly (avoid `git add -A`).
9. Commit with a message that preserves multi-line formatting when body/footer are present.

## Examples

```text
feat(segment): add Ramadan segment with Aladhan API
fix(cache): always store mod time
docs(readme): update installation instructions
refactor(config): simplify option parsing logic
chore(deps): bump github.com/shirou/gopsutil/v4
feat(segment)!: rename template property StartTime to Start

BREAKING CHANGE: template strings using .StartTime must be updated to .Start
```

## Validation Checklist

- [ ] Type is one of the allowed values in .commitlintrc.yml
- [ ] The commit message respects the rules defined in .commitlintrc.yml
- [ ] Scope (if present) reflects the actual area changed
- [ ] Description is imperative mood, no trailing period
- [ ] Full header line (type + scope + description) is 72 characters or fewer
- [ ] Both `!` after type/scope **and** `BREAKING CHANGE:` footer are present whenever the change breaks existing behavior
- [ ] No sensitive files staged (.env, credentials, etc.)
