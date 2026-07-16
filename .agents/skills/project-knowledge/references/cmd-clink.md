# cmd / Clink

## Lua and pipes

- Clink's `io.popenrw` reads are BLOCKING with no peek/timeout in Clink Lua (verified in
  `io_api.cpp`). Any protocol consumed from Lua must guarantee a fixed number of records per
  request - the serve wait-mode contract (exactly 2 records, upheld by Go's `renderComplete` even
  on segment panic) exists for this.
- `io.popenrw` runs the command via `%COMSPEC% /c`, so `2>nul` works in the command string and is
  REQUIRED - the child's stderr otherwise inherits the console and corrupts the display.
- Clink creates its pipes `_O_NOINHERIT` and only dups the child ends inheritable
  (`pipe_pair::init`), so cmd's death guarantees the daemon's stdin write handle closes.

## Windows lifecycle

- There is no SIGPIPE on Windows - stdin EOF is a daemon's ONLY exit signal. Design teardown
  around fd closure, not signals.
- The cmd feature line for the daemon is `serve_enabled = true` (Streaming feature in
  `src/shell/cmd.go`).

## Testing

- `luac -p` for syntax; a Lua harness with a stubbed Clink API covers logic
  (lua.exe via `winget install DEVCOM.Lua`, Clink via `winget install chrisant996.Clink`).
- Clink itself cannot run headless - live smoke tests stay manual.
