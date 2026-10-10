# Agent Instructions

Oh My Posh is a cross-shell prompt theme engine written in Go, with the module root in `src/`. Rendering runs on every
shell prompt, so it is a hot path: no extra process spawns, network calls, or file reads per render without
`src/cache`.

## Non-negotiables

Corrections in past sessions keep landing on these. Check each one before you report a change as done.

1. **Reuse before building.** Find how the codebase already solves the problem before designing anything: a similar
   segment, an `Environment` method, `src/cache`, or a CLI call behind an option like other segments use. Never add a
   dependency without asking. Remove deprecated logic instead of maintaining it.
2. **Go through `runtime.Environment`** (`s.env` in segments) for every OS, shell, file, and command call, so tests can
   mock it. Check its methods before reaching for `os`, `os/exec`, or `path/filepath`.
3. **Tests prove behavior, not the patch.** A test must fail when the behavior breaks and survive a correct refactor.
   Write them table-driven; the `golang` skill has the rules.
4. **No `else`.** Use guard clauses, early returns, or a `switch`. The Stop hook blocks new `else` branches.
5. **Comments explain a why the code cannot.** Never restate a name or signature, exported symbols included, in any
   language. Default to no comment.
6. **Log with `src/log`:** `log.Error(err)` where the error occurs, and `defer log.Trace(time.Now(), args...)` in
   functions worth tracing. Do not format errors yourself.
7. **Commit only when asked, push only when asked.** Fold every fix to a commit that is not on `main` into that commit
   with `git commit --fixup <sha>` and `git rebase --autosquash`, lint and CI fallout included. Never rewrite `main`.

## Repository map

| Path                          | Purpose                                                                      |
| ----------------------------- | ---------------------------------------------------------------------------- |
| `src/segments/`               | One `.go` and one `_test.go` per segment                                     |
| `src/config/`                 | Segment registry; an unregistered segment fails silently at runtime          |
| `src/prompt/`                 | Rendering engine                                                             |
| `src/runtime/`                | `Environment` abstraction and its mock                                       |
| `src/cache/`                  | TTL, file, and command-path caching; do not build another cache              |
| `src/cli/`                    | CLI commands on `src/cmdtree`, registered in `root.go`                       |
| `src/shell/`                  | `oh-my-posh init <shell>`; `scripts/` are embedded, so rebuild after editing |
| `themes/`                     | Bundled themes, validated against `themes/schema.json`                       |
| `website/`                    | Docusaurus docs; segment pages live in `website/docs/segments/<category>/`   |

A shell integration change usually spans `src/shell/<shell>.go` and its script in `src/shell/scripts/`.

## Commands

```bash
# from src/
go test ./...
golangci-lint run

# from website/
npm run build    # before a docs pull request
```

The Stop hook (`.agents/hooks/main.go`) formats the changed files, runs `modernize`, `fieldalignment`, `golangci-lint`,
`go test`, `markdownlint`, and the `else` check, and builds platform-specific code for the other operating systems. Fix
what it reports; never run `fieldalignment -fix`. `.claude/settings.json` and `.github/hooks/quality.json` both wire
it, so change them together.

## Docs

Docs pages have no H1, because the front-matter `title` renders as the heading; use only H2 and H3. Wrap lines at 120
characters. Use the `segment-create` skill for a new segment and `segment-docs` for its documentation.

## Project knowledge

Before touching shell scripts, terminal or pty behavior, WSL-based testing, engine internals, or release and CI work,
read the matching file in the `project-knowledge` skill. When a session uncovers a verified, actionable gotcha, append
it to that file and commit it with the related change.

## Pull request reviews

Validate every comment against the code. Fold each fix into the commit that introduced the code with a fixup and
`git rebase --autosquash`, then force-push the pull request branch with `--force-with-lease`. Reply to every thread
with the conclusion and the commit, then resolve it.
