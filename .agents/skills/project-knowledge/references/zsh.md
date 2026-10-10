# zsh

## zle

- Every new editor invocation starts in keymap `main`; a `vicmd` selection does not carry into the next line.
- `$?` after the transient prompt's `zle .send-break` is 1, native Ctrl+C gives 130 (TODO in `_omp_zle-line-init`).
  Don't fix one path without the other.
- Ubuntu's `/etc/zsh/zshrc` defines a `zle-line-init`; never assume the widget slot is empty.
- zsh-vi-mode always wraps omp's `zle-line-init` (it initializes lazily at first `precmd`), and our
  `.recursive-edit` swallows the session, so `_omp_zle-line-init` must call `zvm_zle-line-init` up front (guarded on
  `$+functions[zvm_zle-line-init]` and `ZVM_INIT_DONE`). Keep that call (#5992).
- Keep `local rawfunc=` at the top of `_omp_zle-line-init`: `zvm_reset_prompt` resolves `$rawfunc` by dynamic
  scoping and would otherwise re-enter our widget.
- Debug widget re-entry by logging `${funcstack[*]}` inside the widget.

## Quoting

- Quote every parameter expansion used as a command argument if its value can contain `[`, `?`, `*` or `#`. Under a
  user's `setopt GLOB_SUBST` it becomes a glob, and `BAD_PATTERN` aborts the widget (`$zle_bracketed_paste[1]` is
  `\e[?2004h`, #7816).

## coproc, fds, and signals

- An interactive zsh with MONITOR prints `[n] pid` at coproc spawn; neither `disown` nor `2>/dev/null` stops it.
  `setopt localoptions no_monitor` does, and the child then inherits SIGINT/SIGQUIT ignored.
- Duplicate coproc fds to session fds (`exec {out}<&p {in}>&p`) so they survive a later `coproc`; `disown %+` keeps
  the daemon out of `jobs` and the job-count segment.
- A redirection-only `exec` applies every listed redirection permanently: `exec {fd}<&p 2>/dev/null` silences the
  session's stderr (#7653). Scope it: `{ exec ... } 2>/dev/null`.
- Writing to a dead coproc raises SIGPIPE, which kills a non-interactive zsh. Guard writes with a `kill -0 $pid`
  check plus `setopt localoptions localtraps; trap '' PIPE`.
- Never `kill -0` a possibly-zero pid: `kill -0 0` signals your own process group and always succeeds.
- zsh 5.9 `read -r -u $fd -d $'\0' -t N` ignores `-t` and blocks forever on a silent fd, so the serve path hangs
  in a non-interactive `zsh script.zsh`; test it in an interactive session.
- Tear down in `zshexit_functions`; closing the fds (or the shell dying, even by SIGKILL) EOFs the daemon's stdin.
