# Testing shells end-to-end

Patterns for functionally driving omp's shell integrations, mostly in WSL (verified on aarch64,
zsh 5.9, fish 4.1.2).

## WSL basics

- WSL `/tmp` is wiped between separate `wsl.exe` invocations (instance auto-shutdown). Either make
  a test fully self-contained in ONE `wsl -e` call, or stage everything under `$HOME` (e.g.
  `~/omp-test`).
- From Git Bash, prefix `MSYS_NO_PATHCONV=1` so `/mnt/c/...` arguments reach wsl.exe unmangled.
- Build for WSL either inside WSL (`/usr/local/go/bin/go`) or cross-compile from Windows
  (`GOOS=linux GOARCH=arm64`). Shell scripts are embedded at build time - rebuild after every
  script edit, then regenerate the sourced init with `oh-my-posh init <shell> --print`.
- Differential testing: build the control binary from `git show HEAD:<file>` or via
  `git stash push -- <file>`; the fixed binary from the working tree.
- `pkill -f <pattern>` matches its own caller's command line when inlined - pick patterns that
  cannot appear in the runner script.

## Getting a pty

- `zsh -i` under plain `wsl -e` hangs before the first prompt (no foreground pty). Options:
  - `wsl -e script -qec 'zsh -i' /dev/null` - real pty, MONITOR enabled, job-control behavior
    testable. Run backgrounded with output to a file and poll a done-marker; foreground runs can
    wedge under the wsl relay.
  - `zmodload zsh/zpty` in a driver zsh - best for keystroke-level scenarios: `zpty omp zsh -i`,
    write raw keys with `zpty -w -n omp $'\x03'` (Ctrl+C delivers a real SIGINT through the pty),
    assert via state files written by in-session commands, drain output with `zpty -r` after exit.
    Note `script(1)` does NOT work inside zpty (typescript never appears).
- Per-shell `script(1)` quirks: bash prompt bytes are not relayed to stdout (probe `${PS1@P}`
  in-session); fish discards piped typeahead (in-session probes never run); zsh relays fine.
- Do not test the zsh serve path from a non-interactive `zsh script.zsh` - it hangs in
  `read -u fd -d $'\0' -t N` (the `-t` is ignored with `-d`, see [zsh](zsh.md)).

## Driving vi-mode / keystroke scenarios (zpty pattern, verified 2026-07-14)

- Harness pattern (built for #5992): stage a binary, config, zdot dirs, and zpty driver scripts
  under `$HOME` in WSL; one driver runs mode-aware ESC/Enter/Ctrl+C scenarios into a state log,
  another asserts the transient prompt renders. Always compare a fixed run against a control zdot
  (same setup, feature under test disabled).
- Keep probes mode-aware: with zsh-vi-mode, a line after a normal-mode accept starts in normal
  mode - prefix typed commands with `i` where needed, and remember a stray self-inserted `i`
  turns the probe into a failing command (useful as a broken-state detector: `last_status=127`).

## Keeping a render alive

- There is no shell-command segment type. To hold `omp stream`/a render open for a controllable
  window, point an `http` segment at a silent local TCP listener
  (`python3 -c 'socket...accept...sleep'`) with a large `http_timeout`; kill the listener to
  trigger the async update record.

## Reading protocol streams

- One `bufio.Scanner` per pipe, ever - a second scanner on the same pipe loses buffered data
  (use a shared reader helper in Go tests).
