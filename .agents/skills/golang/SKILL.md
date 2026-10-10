---
name: golang
description: >
  Go conventions for this project that differ from common Go practice. Apply when writing,
  reviewing, or refactoring any Go source or test file.
triggers:
  - on_commit
---

# Go conventions

Standard idiomatic Go applies. These are the rules that differ or that agents keep breaking.

## Code

- Never use `else`, in tests too. Return early, `continue`, `break`, or use a `switch`.
- Comments explain a why the code cannot express: a hidden constraint, a workaround, a surprising
  invariant. No doc comments that restate a name or signature, exported symbols included.
- Use named fields in every struct literal, never positional ones: fixing fieldalignment findings
  reorders struct fields.
- Keep lines under 180 characters (`lll`).
- When a standard library package collides with an existing import, alias it `lib<name>`, for
  example `libtime`.
- Start error strings lowercase, without trailing punctuation.
- Ask before adding a dependency.

## Tests

- Assert with `testify` (`assert`, `require`).
- Write table-driven tests: one behavior is one test function with a table. Per-case fixtures,
  mock returns, and expected errors are table fields, not reasons for a new function. Extend an
  existing table instead of adding a function. Split when the API under test or the setup and
  assertion flow differs, and for no other reason.
- Test behavior, not the patch. Name the behavior a test proves, then ask:
  1. Could the original bug still occur while this test passes?
  2. Could a correct refactor make it fail?

  If either answer is yes, redesign or drop the test. Never assert on source text, embedded
  scripts, or statement order; assert on text when the emitted text is itself the contract
  (generated shell code, escaping, serialization). When the behavior cannot be exercised, say so
  instead of adding a proxy test.
- A test that mutates package-level state saves the original value and restores it with
  `t.Cleanup`, never a hardcoded value.

## Before committing

Run from the module root and fix every finding:

```bash
modernize -fix ./...
fieldalignment ./...
golangci-lint run
```

- Never run `fieldalignment -fix`: it drops the comments of the fields it reorders. Reorder by
  hand.
- Commit the `modernize` rewrites with the change that triggered them.
- After touching a `_windows.go`, `_unix.go`, `_darwin.go`, `_linux.go`, or `//go:build` file,
  also run `go build ./...` and `golangci-lint run` with `GOOS` set to another platform. The host
  toolchain skips other platforms' files, so their errors otherwise surface in CI.
- `dupl` on near-identical tests: merge them into one table. Reserve `//nolint:dupl`, with a reason,
  for tests that must stay separate. `goconst`: extract a string literal used three or more
  times into a constant.
