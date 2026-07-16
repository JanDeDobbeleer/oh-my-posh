# E2E Test Suite

End-to-end tests for the oh-my-posh shell integrations themselves (not the Go internals under
`src/`). They generate real init scripts with the `oh-my-posh` binary, feed them to the actual
shells, and drive interactive sessions in a pseudo-terminal.

This is a separate Go module (`github.com/jandedobbeleer/oh-my-posh/e2e`, own `go.mod`/`go.sum`)
so its dependencies never leak into `src/`.

## Layers

1. **Syntax** (`syntax_test.go`) — generate the init script per shell x config overlay and
   validate it with the shell's own parser (`bash -n`, `zsh -n`, `fish --no-execute`, a
   `System.Management.Automation.Language.Parser` call for pwsh, `nu-check` for nu).
2. **Smoke** (`smoke_test.go`) — boot each shell interactively in a real pty with the generated
   init script, assert the prompt renders cleanly, a typed command runs, and the shell exits.
3. **Behavior** (`features_test.go`) — per-feature scenarios: exit-code propagation, transient
   prompt, right prompt, styling colors, and FTCS marks.

Both layers 2 and 3 drive the shell through `harness.Session`, a pty wrapper with a vt10x screen
emulator for rendered-screen assertions and a raw-byte buffer for escape-sequence assertions.

The suite targets the "big five" shells: bash, zsh, fish, pwsh, nu.

## Running locally

```shell
cd e2e
go test -count=1 ./...
```

Add `-v` for per-shell/per-overlay output, or `-run <Test>/<shell>` to scope to one case.

Every test skips cleanly (`t.Skip`) when a shell's binary is not on `PATH`, or when the shell is
not supported on the current platform. Set `OMP_E2E_REQUIRE` to a comma-separated list of shell
names (e.g. `pwsh,nu`) to turn those skips into hard failures for the listed shells instead —
useful to catch a shell that's supposed to be installed but silently isn't. It defaults to unset
(everything skips); CI sets it to the shells each job installs so an expected shell hard-fails
instead of skipping quietly.

| Shell | Linux/macOS                  | Windows                                                    |
|-------|------------------------------|------------------------------------------------------------|
| bash  | yes (skip if binary missing) | skipped (msys/WSL bash under ConPTY is not representative) |
| zsh   | yes (skip if binary missing) | skipped                                                    |
| fish  | yes (skip if binary missing) | skipped                                                    |
| pwsh  | yes (skip if binary missing) | yes (skip if binary missing)                               |
| nu    | yes (skip if binary missing) | yes (skip if binary missing)                               |

Layer 1 (syntax) still runs its non-empty/no-placeholder assertions unconditionally for every
shell; only the parser check itself is skipped when the binary is missing.

### Building the omp binary

The suite builds `../src` once per test run (guarded by `sync.Once`) into a temp directory and
uses that binary for every `init` invocation. Set `OMP_E2E_BINARY` to an absolute path to skip
the build and reuse a prebuilt binary instead — useful when iterating on tests without rebuilding
oh-my-posh every run.

### Isolation

Every omp invocation and shell session gets `OMP_CACHE_DIR` pointed at a fresh `t.TempDir()`, so
runs never share or pollute a developer's real cache. Sessions also fix `TERM=xterm-256color`
and a 120x30 pty size.

## Adding a new shell

Add an entry to the `Shells` table in `harness/shells.go`. A `ShellDef` needs:

- `Name` — the omp shell name passed to `oh-my-posh init <Name>`.
- `Binary` — the executable looked up via `LookupShellBinary`; missing means skip.
- `SyntaxCheck(scriptPath) *exec.Cmd` — a command that parses (not executes) the script and
  exits non-zero on a syntax error.
- `Launch(t, scriptPath, workDir) (bin string, args []string, env []string)` — how to boot the
  shell interactively with `scriptPath` sourced (e.g. a temp rc file plus `-i`).
- `Fail` — an `ExitCommand{Command, Code}`: a command line that reliably exits non-zero when
  typed interactively, and the exit code oh-my-posh should report in the next prompt.

If the shell can't be driven faithfully on every platform, add a case to
`ShellDef.SupportedOnHost` (see the bash/zsh/fish Windows exclusion for the reasoning).

Layers 1-3 all iterate `harness.Shells`, so a correctly filled-in entry is picked up everywhere
automatically — no test file needs editing for a new shell on its own.

## Adding a new feature scenario

`features_test.go` drives every scenario through one table-driven test, `TestFeatures`, whose
subtests are named `TestFeatures/<scenario>/<shell>`. To add a scenario:

1. If the feature needs a config change, add an `Overlay` function to `harness/config.go` that
   mutates the base config map (see `Transient`, `RPrompt`, `Colored`, `ShellIntegration` for
   examples). If the overlay changes the generated init script, register it in `overlaySets` in
   `syntax_test.go` so layer 1 covers it too (`Colored` doesn't change the script, so it's left
   out of that matrix).
2. Append a `scenario` entry to `featureScenarios` in `features_test.go`:
   - `overlays` — the `Overlay`s to apply, if any.
   - `skips` — a `map[string]string` of shell name to skip reason, for shells that don't support
     the feature (see the bash entries on `transient`/`rprompt`, or the nu entry on `ftcs`, for the
     pattern). Leave it `nil` when every shell is expected to pass.
   - `run` — a `func(t *testing.T, sh harness.ShellDef, s *harness.Session)` with the scenario's
     assertions. The `TestFeatures` runner has already applied the overlays, skipped
     unsupported shells, started the session and waited for the first prompt by the time `run`
     is called; drive the rest of the session (`SendLine`, `WaitFor`) and assert against
     `Screen()`/`ScreenLines()` for rendered output, `Raw()` for escape-sequence assertions, or
     `MarkerColor()` for a rendered marker's cell colors.
3. Never make a shell silently succeed or fail a feature it doesn't support — add it to `skips`
   with a comment explaining why instead.

## Harness internals

Things the harness does that are easy to break by accident:

- **DSR replies** — PSReadLine on a Unix pty repeatedly queries the cursor position (`CSI 6n`)
  and blocks rendering until it gets a reply. ConPTY answers this internally on Windows; on
  Linux/macOS the harness itself replies with the vt10x cursor position from its reader
  goroutine (`harness/session.go`). Remove that and pwsh-on-Linux wedges until the 30s timeout.
- **Single pty reader** — exactly one goroutine reads the pty and feeds both the vt10x screen
  and the raw buffer under one mutex. A second reader silently loses buffered data.
- **nu vendor autoload** — nu loads every `.nu` under `$nu.vendor-autoload-dirs` after
  `--config`, so a machine with the real oh-my-posh nu integration installed would clobber the
  test prompt. Sessions point `XDG_DATA_HOME` at an empty temp directory to prevent this.
- **Absolute binary paths** — go-pty's Windows `Cmd` resolves bare executable names relative to
  `Cmd.Dir` when `Dir` is set, so `Start` always resolves the shell binary to an absolute path
  first.
- **bash lookup on Windows** — plain `PATH` lookup finds System32's WSL launcher `bash.exe`,
  which cannot run Windows-style script paths; `LookupShellBinary` derives Git Bash's location
  from `git.exe` instead.

## Known limitations

- bash only renders a transient prompt or right prompt inside a `ble.sh` session (gated on
  `BLE_SESSION_ID`, see `src/shell/bash.go`); this harness's plain
  `bash --noprofile --rcfile ... -i` session doesn't provide one, so `TestFeatures/transient` and
  `TestFeatures/rprompt` skip bash explicitly.
- nu never emits any FTCS mark: `Features().Nu()`'s switch does list a case for `FTCSMarks` (see
  `src/shell/nu.go`), but that case deliberately returns an empty `Code`, so the generated script
  never gets the hook that prints them. `TestFeatures/ftcs` skips nu explicitly.
- cmd, elvish, xonsh and yash are not covered by any layer.

## CI

`.github/workflows/e2e.yml` runs this suite on `ubuntu-latest` (bash, pwsh preinstalled; zsh and
fish installed via `apt`; nu installed from a pinned GitHub release) and on `windows-latest`
(pwsh preinstalled; nu installed from a pinned release; bash/zsh/fish skip by design).
