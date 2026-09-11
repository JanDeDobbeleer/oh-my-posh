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

## Key bindings and error attribution

- `commandline --is-valid` exit codes (fish >= 3.4, verified on 4.1.2, 2026-09-11): 0 = valid,
  1 = erroneous, 2 = incomplete. These map 1:1 onto the reader's own branches in
  `handle_execute` (reader.rs): ERROR prints the parse error and keeps the buffer, INCOMPLETE
  inserts a continuation newline. A buffer ending in a backslash (e.g. `echo a; and \`) reports
  0 - the reader executes it and errors at runtime, it is NOT a continuation case for
  `--is-valid`.
- A parse error raised by `commandline --function execute` inside a key binding is attributed to
  the binding's source file and function (`<init>.fish (line 1): ... in function '_omp_enter_key_handler'`)
  because the parser backtrace includes the handler's block (issue #7862). Rethrow through a
  pristine parser instead: `commandline --current-buffer | string collect | (status fish-path) --no-execute`
  reproduces fish's exact interactive message (`fish: <msg>` + caret), preserves $status, and
  keeps the buffer - matching default-reader behavior.
- Command substitutions are not allowed in command position in fish 4.x:
  `(status fish-path) --no-execute` is a syntax error; assign to a variable first.

## Testing quirks

- fish under `script(1)` discards/paste-buffers piped typeahead, so in-session probes never
  execute - a harness limitation, not a bug (zsh under the same setup is fine). Assert via files
  written by event handlers instead.
- The streaming transient prompt is cached in a tempfile (`$_omp_streaming_tempfile.transient`),
  not a variable, because `_omp_cleanup_stream` runs before the transient repaint.
