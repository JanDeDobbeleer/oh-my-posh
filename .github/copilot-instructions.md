# GitHub Copilot Instructions

> For full coding guidelines, commit conventions, and agent workflows, see [AGENTS.md](../AGENTS.md).

## Project Overview

Oh My Posh is a cross-shell prompt theme engine written in Go. It renders prompt segments by
querying an `Environment` abstraction that wraps all OS/shell interactions.

## Tech Stack

| Layer | Technology |
| --- | --- |
| Core engine | Go - `src/` |
| Docs site | Docusaurus (MDX) - `website/` |
| Themes | JSON - `themes/` |
| Config formats | TOML · JSON · YAML |
| Installer scripts | `packages/` |
| CI/build helpers | `build/` |

## Key Commands

```bash
# Go — run from src/
go test ./...
go test ./segments/... -run TestFoo  # single test
golangci-lint run

# Docs — run from website/
npm run start    # local dev server
npm run build    # validate before opening a docs PR
```

## Codebase Exploration

**Always explore the actual codebase before planning or implementing.** Do not rely on memory
or assumptions. Use the file system tools to read relevant files first - the codebase evolves
and the feature you're asked to add may already exist.

## Source Layout

| Path | Purpose |
| --- | --- |
| `src/segments/` | One `.go` + one `_test.go` per segment |
| `src/config/segment_types.go` | Segment type registry (gob + string constants) |
| `src/cli/` | CLI commands (Cobra); `root.go` is the entry point |
| `src/prompt/engine.go` | Segment rendering loop |
| `src/cache/` | Existing TTL/file/command-path cache infrastructure |
| `src/runtime/` | `Environment` abstraction + mock |

## Segment Architecture

Every segment lives in `src/segments/` and implements the `SegmentWriter` interface. Use the
`Environment` abstraction (`env`) for **all** OS/shell calls - never call OS APIs directly.

Adding a segment requires **five** artifacts (use the `segment-create` skill to scaffold):

1. `src/segments/<name>.go`
2. `src/segments/<name>_test.go`
3. `website/docs/segments/<name>.mdx`
4. Updates to `website/sidebars.js` and `website/static/schema.json`
5. `gob.Register(&segments.MySegment{})` in `src/config/segment_types.go`

Missing step 5 will cause the segment to fail silently at runtime.

## Shell Integration

`oh-my-posh init <shell>` is how users wire oh-my-posh into their shell. It:

1. Writes a shell-specific init script to the cache (source: `src/shell/scripts/omp.<ext>`)
2. Returns a one-liner for the shell to `eval` - this sources the cached script, which hooks
   into prompt rendering

The `src/shell/` package contains per-shell logic (`pwsh.go`, `bash.go`, `zsh.go`, etc.) that
generates the hook commands. The scripts in `src/shell/scripts/` are embedded and templated at
init time. When modifying shell behaviour, changes typically span both the `.go` file and the
corresponding `.ext` script.

Supported shells: `bash`, `zsh`, `fish`, `powershell`/`pwsh`, `cmd`, `nu`, `elvish`, `xonsh`.

## CLI Commands

CLI commands use [Cobra](https://github.com/spf13/cobra) and live in `src/cli/`. To add a new command:

1. Create `src/cli/<name>.go` with a `var <name>Cmd = &cobra.Command{...}`
2. Register it in `src/cli/root.go` via `RootCmd.AddCommand(<name>Cmd)`

## Caching

`src/cache/` provides the existing caching infrastructure - use it instead of building new
cache logic. It supports TTL-based key/value storage, file-based persistence, and command-path
caching. Do not introduce new cache packages unless `src/cache/` genuinely cannot meet the
requirement.

## Themes

Themes are plain JSON in `themes/`. All themes must validate against `website/static/schema.json`.
Do not introduce breaking schema changes without updating the schema file.

## Documentation

Segment doc pages use MDX frontmatter with `title`, `sidebar_label`, and `id`. See the
`segment-docs` skill for the canonical Go→MDX mapping.
