# zsh

## zle facts

- Every new editor invocation starts in keymap `main`; a `vicmd` selection does not survive into
  the next line.
- `$?` after the transient prompt's `zle .send-break` is 1; native Ctrl+C yields 130. Known gap,
  documented as a TODO in `_omp_zle-line-init` - do not "fix" one without solving the other path.
- Ubuntu's `/etc/zsh/zshrc` defines a `zle-line-init` (terminfo smkx), so omp's widget takes the
  decorate path even in minimal setups - never assume the widget slot is empty.

## zsh-vi-mode (ZVM) interaction (verified 2026-07-14, issue #5992)

- ZVM initializes lazily at first `precmd`, so it ALWAYS wraps omp's `zle-line-init` regardless of
  source order: its wrapper runs our widget first, `zvm_zle-line-init` second. Because the
  transient prompt's `zle .recursive-edit` consumes the whole editing session, ZVM's line-init
  effectively ran at line END - any line accepted or interrupted from normal mode desynced
  `ZVM_MODE` from the active keymap, and `zvm_select_vi_mode`'s same-mode early return made the
  break permanent.
- Fix in `_omp_zle-line-init`: call `zvm_zle-line-init` up front (guarded on
  `$+functions[zvm_zle-line-init]` and `ZVM_INIT_DONE == true`). Keep this if the function is ever
  restructured.
- Landmine: `zvm_reset_prompt` resolves `$rawfunc` via **dynamic scoping**. Any ZVM code running
  `zle reset-prompt` while inside our line-init picks up the line-init wrapper's `rawfunc`
  (`_omp_decorated_zle-line-init`) and re-enters the widget recursively. The `local rawfunc=` at
  the top of `_omp_zle-line-init` shadows it - load-bearing, keep it.

## coproc and signals

- An interactive zsh with MONITOR prints "[n] pid" at coproc spawn; `disown` is too late and a
  `{ coproc ... } 2>/dev/null` block does NOT suppress it. `setopt localoptions no_monitor` does -
  side effect: the child inherits SIGINT/SIGQUIT ignored (POSIX no-job-control), which Go
  preserves.
- Duplicate coproc fds to session fds (`exec {out}<&p {in}>&p`) - duplicates survive a later
  `coproc` replacing the slot. `disown %+` keeps the daemon out of `jobs` and the job-count
  segment.
- Writing to a dead coproc pipe raises SIGPIPE, which **kills a non-interactive zsh outright**
  (`2>/dev/null` cannot stop a signal). Guard daemon writes with a `kill -0 $pid` pre-check plus
  `setopt localoptions localtraps; trap '' PIPE` (function-local, user pipelines unaffected).
- Never pass a possibly-zero pid to `kill -0` - `kill -0 0` signals the caller's own process group
  and always succeeds.

## Footguns

- A redirection-only `exec` applies EVERY listed redirection to the shell permanently:
  `exec {fd}<&p {fd}>&p 2>/dev/null` silences the session's stderr for good (caused issue #7653).
  Scope the stderr suppression with a brace block: `{ exec ... } 2>/dev/null`.
- zsh 5.9: `read -r -u $fd -d $'\0' -t N` ignores `-t` entirely and blocks forever on a silent fd.
  Without `-d`, the timeout works.
- Teardown belongs in `zshexit_functions`; the daemon lifecycle is fd-governed (closing the fds -
  or the shell dying, even by SIGKILL - EOFs the daemon's stdin).
- To debug widget re-entry, log `${funcstack[*]}` inside the widget - `zle` calls from shell
  functions appear on the stack and expose who invoked what.
