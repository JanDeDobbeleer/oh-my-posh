# bash

Verified on bash 5.2 (2026-07).

## Readline and prompt

- Running the `exec` builtin from PROMPT_COMMAND - even a harmless `exec {fd}</dev/null` -
  **silently disables readline for the whole session**: the prompt never paints again, no error,
  commands still execute non-interactively. Use coproc fds directly; never exec-dup or exec-close
  from prompt context.
- PS1 must stay `'$(_omp_get_primary)'` (single-quoted). With promptvars on, literal prompt
  content executes `$(...)` - a directory named `$(cmd)` becomes command injection.
- "Expansion returns correct bytes" is NOT "prompt displays" - always capture a real typescript
  (`script -qe -c ... FILE`) when verifying prompt behavior. Note `script(1)`'s stdout relay does
  not carry bash prompt bytes (zsh's does); probe `${PS1@P}` in-session instead.

## coproc

- `coproc NAME { cmd; }` reports the pid of a wrapper subshell - `exec` in the body so the daemon
  replaces it.
- The coproc's ORIGINAL fds must be closed after duplication
  (`eval "exec ${NAME[0]}<&- ${NAME[1]}>&-"`) or the child's stdin never EOFs when the dups close.
- Subshells close coproc fds in non-interactive bash - a `(trap '' PIPE; ...)` subshell guard does
  not work there; save/ignore/restore `trap '' PIPE` in the parent instead (nothing forks while
  ignored).
- bash prints failed `>&fd` redirection errors before later redirections apply - put
  `2>/dev/null` BEFORE `>&"$fd"`.

## History

- A bash serve daemon was implemented and **reverted** (2026-07-07): no measurable speedup -
  native-Linux spawns are 11-16ms and sync-only wait-mode plus the display-time subshell cannot
  beat that. Do not re-propose without new evidence. The gotchas above came out of that work and
  remain valid.
