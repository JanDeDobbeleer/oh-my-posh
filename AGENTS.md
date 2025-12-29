# Agent Instructions

## General File Creation Guidelines

When creating new files:

- **Always use LF (Unix-style) line endings**, not CRLF (Windows-style)
- This repository uses `.gitattributes` to enforce LF line endings
- Ensures consistency across all platforms and avoids Git warnings

## Golang

When editing Go files (`*.go`):

- Read `.github/instructions/golang.md` and announce once per task that you are following it.
- Before committing, ensure code is formatted and linted:
  - Run `gofmt` (or `go fmt`) and organize imports.
  - Run `golangci-lint run` at the repository root and address findings.

## Markdown

When editing Markdown (`*.md`, `*.mdx`):

- Read `.github/instructions/markdown.md` and announce once per task that you are following it.
- Use proper headings (`##`, `###`), fenced code blocks with language, and keep lines within the configured limit.

## PowerShell

When editing PowerShell files (`*.ps1`, `*.psm1`, `*.psd1`):

- Read `.github/instructions/powershell.md` and announce once per task that you are following it.
- Follow PowerShell best practices for naming, formatting, and error handling.
- Include comment-based help for public functions and ensure proper parameter validation.

## Commit and Pull Requests Guidelines

- Use [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/#summary) for PR titles and commit messages.
- The repository specific rules are in `.commitlintrc.json`.
- Always run `gofmt` and `golangci-lint run` before submitting changes.
- Limit commit message lines to a maximum of 200 characters.
- **Do not commit initial plans or progress updates as separate commits.**
  Include planning information in the PR description instead.

Examples:

- `feat(config): cache remote configs via HEAD check`
- `fix(markdown): correct reference link syntax in docs`
- `chore(ci): run golangci-lint in build step`
