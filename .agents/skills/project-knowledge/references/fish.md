# fish

Verified on fish 4.1.2 in WSL (2026-07).

## Process and job model

- fish does NOT fork a backgrounded pipeline stage that is a *function* - it blocks the main
  shell. Run readers as an external process: `fish --no-config -c $script args &`.
- fish 4.x `jobs --last --pid` prints a "Process" header plus one pid per pipeline stage - filter
  with `string match --regex '^\d+$'` and treat the result as a list.
- `kill -0 0` always succeeds (signals the caller's own process group) - guard zero pids
  everywhere before liveness checks.

## Fifos and lifecycle

- A fifo write with no reader blocks forever - liveness-check the reader before every fifo write.
- The serve daemon opens its request fifo O_RDWR so fish's open-write-close pattern never EOFs it;
  consequence: SIGKILL of fish orphans daemon and reader permanently (no EOF, no SIGPIPE ever).
  Normal teardown runs via the `fish_exit` event handler (quit + kill the pipeline).
- `--on-signal SIGUSR1` handlers fire between commands in non-interactive scripts (verified with a
  200ms poll loop).

## Testing quirks

- fish under `script(1)` discards/paste-buffers piped typeahead, so in-session probes never
  execute - a harness limitation, not a bug (zsh under the same setup is fine). Assert via files
  written by event handlers instead.
- The streaming transient prompt is cached in a tempfile (`$_omp_streaming_tempfile.transient`),
  not a variable, because `_omp_cleanup_stream` runs before the transient repaint.
