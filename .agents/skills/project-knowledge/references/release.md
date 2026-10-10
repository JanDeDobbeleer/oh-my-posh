# Release, CI, and git

## Release (`.github/workflows/release.yml`)

- Releases are immutable: the workflow creates a draft and publishes last. Every asset must be in the draft before the
  publish step.

## CI

- `go test -race` does not work on the windows/arm64 dev machine or its WSL (no C compiler, no sudo). CI's
  `Race Detector` step (`code.yml`, ubuntu) is the only race gate.

## Git on Windows

- With `core.autocrlf=true`, `git rebase --autosquash` can stop on a plain `pick` with "Your local changes would be
  overwritten" and no real conflict (line-ending renormalization). Set `core.autocrlf false` for the rebase and
  restore it after; don't try to resolve a phantom conflict.
