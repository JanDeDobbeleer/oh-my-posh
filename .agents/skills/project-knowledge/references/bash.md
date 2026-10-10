# bash

Verified on bash 5.2 (2026-07).

- Never run the `exec` builtin from PROMPT_COMMAND, not even `exec {fd}</dev/null`: it silently disables readline
  for the session (prompt never paints, commands still run).
- Keep PS1 as `'$(_omp_get_primary)'` (single-quoted). With promptvars, literal prompt content executes `$(...)`,
  so a directory named `$(cmd)` becomes command injection.
- Transient prompt and rprompt exist only under ble.sh (`bashBLEsession`, gated on `BLE_SESSION_ID` in
  `src/shell/bash.go`); plain bash gets no code for either.
- Correct `${PS1@P}` bytes do not prove the prompt displays; verify in a real pty ([testing](testing.md)).
- Don't retry a bash serve daemon (built and reverted 2026-07): native Linux spawns cost 11-16ms, and sync
  wait-mode plus the display-time subshell cannot beat that.
