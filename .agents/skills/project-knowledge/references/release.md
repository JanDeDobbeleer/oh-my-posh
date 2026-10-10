# Release, CI, and git

## Release (`.github/workflows/release.yml`)

- Releases are immutable: the workflow creates a draft and publishes last. Every asset, including the signed
  `packslip.sigstore.json`, must be in the draft before the publish step.
- The `packslip` job signs the `build-artifacts` binaries (`posh-*` minus `.sha256`/`.sig`) with `upload: false`; the
  bundle reaches the release through the `release` job's merge-multiple download and `files: *`. Android and msix are
  excluded on purpose.
- Keep signing in `release.yml`: consumers pin that workflow file. `packslip` is `continue-on-error` so a signing
  outage never blocks a release.

## CI

- `go test -race` does not work on the windows/arm64 dev machine or its WSL (no C compiler, no sudo). CI's
  `Race Detector` step (`code.yml`, ubuntu) is the only race gate.

## Git on Windows

- With `core.autocrlf=true`, `git rebase --autosquash` can stop on a plain `pick` with "Your local changes would be
  overwritten" and no real conflict (line-ending renormalization). Set `core.autocrlf false` for the rebase and
  restore it after; don't try to resolve a phantom conflict.
