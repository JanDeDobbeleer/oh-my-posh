# cmd / Clink

- Clink's `io.popenrw` reads block with no peek or timeout. A protocol read from Lua must emit a fixed number of
  records per request (serve wait-mode: exactly 2, see [codebase](codebase.md)).
- `io.popenrw` runs through `%COMSPEC% /c`; keep `2>nul` in the command, or the child's stderr corrupts the display.
- Windows has no SIGPIPE: stdin EOF is the daemon's only exit signal. Clink's pipes are `_O_NOINHERIT`, so cmd's
  death closes the daemon's stdin.
- Test with `luac -p` plus a Lua harness that stubs the Clink API (`winget install DEVCOM.Lua`,
  `chrisant996.Clink`). Clink cannot run headless, so live smoke tests stay manual.
