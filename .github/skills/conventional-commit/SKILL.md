---
name: conventional-commit
description: >
  Workflow for generating conventional commit messages following the Conventional Commits
  specification. MUST be invoked every time a commit is created. Guides construction of
  standardized commit messages with correct type, scope, description, body, and footer.
triggers:
  - on_commit
---

## Commit Message Structure

```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

### Types

| Type | When to use |
| ---- | ----------- |
| `feat` | A new feature |
| `fix` | A bug fix |
| `docs` | Documentation changes, no code |
| `style` | Formatting, missing semicolons, etc. (no logic change) |
| `refactor` | Code change that is neither a fix nor a feature |
| `perf` | Performance improvement |
| `test` | Adding or correcting tests |
| `build` | Changes to build system or external dependencies |
| `ci` | CI configuration changes |
| `chore` | Maintenance tasks (updating deps, tooling, etc.) |
| `revert` | Reverts a previous commit |

Append `!` after the type/scope to signal a **breaking change**: `feat!:` or `feat(api)!:`

### Scope

Optional. Use the name of the area affected, e.g., `segment`, `cache`, `config`, `ui`.
Omit when the change is truly cross-cutting.

### Description

- Required. One short imperative sentence, no period at the end.
- Use the imperative mood: "add", not "added" or "adds".

### Body

Optional. Add context about *why* the change was made, not *what*. The diff shows that.
Wrap at 72 characters.

### Footer

Use for:
- `BREAKING CHANGE: <description>` (required when `!` is used; explains the break).
- Issue references: `Closes #123`, `Fixes #456`.
- Co-authors: `Co-Authored-By: Name <email>`.

## Workflow

1. Run `git status` to review changed files.
2. Run `git diff` and `git diff --cached` to inspect staged and unstaged changes.
3. Identify the **type** from the table above.
4. Identify the **scope** from the files/area changed.
5. Write a short **description** in the imperative mood.
6. Add a **body** if the *why* needs explanation.
7. Add a **footer** for breaking changes or issue references.
8. Stage the relevant files explicitly (avoid `git add -A`).
9. Commit with a message that preserves multi-line formatting when body/footer are present.

## Examples

```
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
- [ ] `BREAKING CHANGE:` footer present when `!` is used
- [ ] No sensitive files staged (.env, credentials, etc.)
