# Testing shells and terminals

## Script edits

- `src/shell/scripts/` is embedded at build time: rebuild the binary after every script edit, then inspect what
  users source with `oh-my-posh init <shell> --print`.

## e2e harness (`e2e/`, see `e2e/README.md`)

- PSReadLine on a raw Unix pty floods `CSI 6n` and wedges without a reply (ConPTY answers it on Windows); the
  harness answers from the vt10x cursor (`harness/session.go`).
- nu autoloads `$nu.vendor-autoload-dirs` after `--config`, so a real omp install clobbers the test prompt. Point
  `XDG_DATA_HOME` at an empty dir.
- go-pty's Windows `Cmd` resolves bare names relative to `Cmd.Dir`; pass absolute binary paths.
- On Windows, `bash` on PATH is System32's WSL launcher; `harness.LookupShellBinary` derives Git Bash from
  `git.exe`.
- Nushell evaluates `TRANSIENT_PROMPT_COMMAND` before Reedline reads input. Test its render timestamp against
  prompt readiness, not submission; `pre_render` only controls Oh My Posh's own streaming pre-render.
- Keep Nushell daemon handles and cycle counters in `hooks.pre_prompt`, not `PROMPT_COMMAND`: prompt closures
  discard environment changes after evaluation. Use the prompt closure only to return the prepared prompt.
- Reedline swaps `PROMPT_INDICATOR_VI_INSERT` and `PROMPT_INDICATOR_VI_NORMAL` itself on a mode change,
  multiline strings and typed input included, so vi and helix mode render every variant up front.
  `commandline set-prompt` tags and job ids must be positive.
- Decode external Nushell protocol records with `lines | each { from json }`; `split row` does not incrementally
  frame an external byte stream. Parenthesize multiline external invocations so trailing flags are not expressions.

## WSL

- Run e2e from a Windows worktree: cross-compile (`GOOS=linux GOARCH=arm64 go build`), then in WSL
  `cd /mnt/c/.../e2e && OMP_E2E_BINARY=/mnt/c/.../oh-my-posh go test .`. If the distro's Go is too old for e2e's
  `go.mod`, unpack a go.dev tarball into `~/go-sdk`.
- WSL `/tmp` is wiped between `wsl.exe` invocations: keep a test in one `wsl -e` call or stage under `$HOME`.
- From Git Bash, prefix `MSYS_NO_PATHCONV=1` so `/mnt/c/...` arguments reach wsl.exe unmangled.

## Getting a pty

- `zsh -i` under plain `wsl -e` hangs (no foreground pty). Use `wsl -e script -qec 'zsh -i' /dev/null` for a real pty
  with job control, run backgrounded with output to a file, and poll a done-marker.
- For keystrokes use `zmodload zsh/zpty`: `zpty omp zsh -i`, `zpty -w -n omp $'\x03'` (real SIGINT), assert via
  state files, drain with `zpty -r`. `script(1)` does not work inside zpty.
- `script(1)` quirks: bash prompt bytes are not relayed (probe `${PS1@P}` in-session); fish discards piped
  typeahead (assert via files written by event handlers); zsh works.
- vi-mode probes must be mode-aware: after a normal-mode accept the next line starts in normal mode, so prefix `i`.
  A stray self-inserted `i` shows up as `last_status=127`.
- To hold a render open, point an `http` segment with a large `http_timeout` at a silent local TCP listener; kill
  the listener to trigger the async update. There is no shell-command segment.

## Protocol streams

- One `bufio.Scanner` per pipe, ever; a second scanner loses buffered data.
- Never treat wall-clock silence as "records done": cold-cache CI runs exceed any fixed gap. Wait for a record that
  proves the assertion and keep a timeout only as a hang bound. Reproduce
  with `go clean -cache && go test -count=1 ./...`.

## Windows Terminal and ConPTY

- "[process exited with code 0]" at shell exit is Windows Terminal's own teardown message, never omp output.
- Panes opened via `wt.exe <commandline>` never auto-close under `closeOnExit=automatic`; profile tabs do.
- A headless conhost (Windows 11 build 26300) reports cursor 0,0, a blank buffer, and drops pty input: assert state,
  not row placement.
- pywinpty's ConPTY EOF lags process death by ~5s (harness artifact, not a hang).
- SendKeys into Windows Terminal drops keystrokes while the user works; a "hung" test may never have got `exit`.
- `Set-Content -NoNewline` joins piped lines into one and corrupts scripts; write byte-exact files with Git Bash
  redirection.
