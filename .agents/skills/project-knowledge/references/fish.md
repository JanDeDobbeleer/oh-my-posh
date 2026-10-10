# fish

Verified on fish 4.1.2 (2026-07 to 2026-09).

## Jobs, fifos, and lifecycle

- fish does not fork a backgrounded function; it blocks the shell. Run readers as an external process:
  `fish --no-config -c $script args &`.
- `jobs --last --pid` prints a "Process" header plus one pid per pipeline stage: filter with
  `string match --regex '^\d+$'` and treat the result as a list.
- Guard zero pids before `kill -0`: `kill -0 0` signals your own process group and always succeeds.
- A fifo write with no reader blocks forever; liveness-check the reader before every write.
- The serve daemon opens its request fifo O_RDWR, so it never sees EOF and a SIGKILLed fish orphans it. Teardown
  must go through the `fish_exit` handler.
- `--on-signal SIGUSR1` handlers fire between commands in non-interactive scripts.
- The streaming transient prompt lives in a tempfile (`$_omp_streaming_tempfile.transient`), not a variable,
  because `_omp_cleanup_stream` runs before the transient repaint.

## Key bindings

- `commandline --is-valid`: 0 valid, 1 erroneous, 2 incomplete. An empty buffer also returns 1, so check the buffer
  is non-empty before treating 1 as an error. A trailing backslash returns 0 (not a continuation).
- A parse error from `commandline --function execute` inside a binding is attributed to the binding's file and
  function (#7862). Rethrow through a pristine parser instead:
  `commandline --current-buffer | string collect | $fish_bin --no-execute` (keeps the buffer and `$status`).
- fish 4.x rejects command substitutions in command position (`(status fish-path) --no-execute`); assign to a
  variable (`set fish_bin (status fish-path)`) first.
