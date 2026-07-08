# GitHub Copilot Instructions

For general coding guidelines, commit conventions, and agent workflows, see [AGENTS.md](../AGENTS.md).

## Tech Stack

| Layer                     | Technology                    |
| ------------------------- | ----------------------------- |
| Core engine               | Go (module root: `src/`)      |
| Documentation site        | Docusaurus (MDX) - `website/` |
| Themes                    | JSON - `themes/`              |
| Config format             | TOML / JSON / YAML            |
| Package/installer scripts | `packages/`                   |
| Build scripts             | `build/`                      |

## Repository Layout

```text
src/
  segments/   # One Go file + one _test.go per segment
  prompt/     # Core rendering engine
  runtime/    # OS/shell abstraction layer
themes/       # Bundled JSON theme files
website/      # Docusaurus docs site (MDX pages, sidebar config, JSON schema)
packages/     # Installer/package manifests
build/        # CI build helpers
```

## Segment Development

When adding a new segment, four artifacts are required - use the `segment-create` skill
to scaffold all of them automatically:

1. `src/segments/<name>.go` - segment implementation
2. `src/segments/<name>_test.go` - unit tests
3. `website/docs/segments/<name>.mdx` - user-facing docs
4. Update `website/sidebars.js` and `website/static/schema.json`
5. Register the type in `src/config/segment_types.go` via `gob.Register(&segments.MySegment{})` - missing this causes
  silent failures at runtime

See the `segment-docs` skill for the canonical mapping between Go source constructs and MDX
documentation fields (template properties, type representations, option tables).

## Go Conventions

- Follow the `golang` skill for project-specific Go standards.
- Each segment implements the `Segment` interface; use `env` (the `Environment` abstraction)
  for all OS/shell calls - never call OS APIs directly.
- Test with `go test ./...` from `src/`.
- Lint with `golangci-lint run` from `src/`.

## Documentation (website/)

- Follow the `markdown` skill for `.md`/`.mdx` formatting rules.
- Segment doc pages live in `website/docs/segments/` and use MDX frontmatter with `title`, `sidebar_label`, and `id`.
- Run `npm run start` inside `website/` for a local dev server.
- Run `npm run build` inside `website/` to verify the site builds before opening a docs PR.

## PowerShell

PowerShell helper scripts live in `packages/` and `build/`. Follow the `powershell` skill for cmdlet conventions.

## Themes

Themes are plain JSON files in `themes/`. New themes must validate against
`website/static/schema.json`. Do not introduce breaking schema changes without updating the
schema file.

## Pull Request Reviews

Whenever any agent performs or addresses a pull request review, follow this process at all
times, regardless of previous instructions:

1. Stay within the scope of the pull request: only address feedback on changes it introduces.
2. Investigate every review comment and reach a conclusion: a code fix, a clarification, or a
   reasoned rejection.
3. Fold each fix into the commit it belongs to. When the change semantically belongs to a
   commit the pull request introduces (any commit not yet on main), create a fixup commit
   (`git commit --fixup <sha>`), squash it (`git rebase --autosquash`), and force-push the
   pull request branch. This preserves the atomicity of the pull request's commits instead
   of stacking review-fix commits on top. Rewriting the pull request branch is fine; main
   history must never be rewritten.
4. Only when a change does not semantically fit any existing commit in the pull request does
   it become its own commit on top, following the commit conventions.
5. Reply to each review comment with the conclusion, referencing the commit that addresses it
   when there is one.
6. Resolve each review thread once its answer and/or fix has been provided.
