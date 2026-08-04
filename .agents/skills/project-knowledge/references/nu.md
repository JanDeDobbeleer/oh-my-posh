# Nushell (`omp.nu`)

- Nushell parses and arity-/type-checks an entire sourced script up front, including inside
  branches that never execute (`if false { ... }`) and inside files pulled in via `source`
  (also a parser-time construct, not deferred to runtime). A runtime capability check (e.g.
  `scope commands | any {|c| $c.name == "commandline set-prompt"}`) does **not** protect against
  this: if the literal call doesn't parse against the current binary's command signatures, the
  whole script fails to source, even when the call is unreachable at runtime. Verified
  2026-08-04: a stock nu build fails with `nu::parser::extra_positional` on `commandline
  set-prompt $text` sitting inside `if $unsupported { ... }`, and identically inside a file
  brought in via a guarded `source`.
- Corollary: any oh-my-posh nu feature that depends on unreleased/experimental nu syntax (e.g.
  `commandline set-prompt` from the unmerged nushell/nushell#18660) must be emitted as text only
  when the feature is actually enabled (from Go, via `Features.Nu()`), never embedded
  unconditionally in the always-sourced `scripts/omp.nu` behind a runtime flag - doing so breaks
  `oh-my-posh init nu` for every nu user on a released build, not just the ones opting in.
- Confirmed against Nushell's own docs (2026-08-04, via "Thinking in Nu" and "How Nushell Code
  Gets Run" on nushell.sh, plus `ParseError::ExtraPositional` in `nu-protocol`): this eager
  whole-file parse/arity-check is intentional, documented design ("think of Nushell as a
  compiled language" - Stage 1 parses the entire source, Stage 2 evaluates it; a bad call
  anywhere in the text fails Stage 1 regardless of runtime reachability), not a nu bug. There is
  no supported way to defer this check to runtime inside a single sourced script; the only fix
  is deciding at Go codegen time whether to emit the syntax at all (which is what
  `nuSupportsStreaming` in `src/shell/nu.go` does).
- The capability probe (`nuSupportsStreaming`) shells out to a bare `nu` command, resolved via
  the OS's normal `PATH` lookup, at **codegen time** (whenever `oh-my-posh init nu`/`--print`
  runs) - not at prompt-render time. If multiple nu installs are on `PATH` (e.g. a stock release
  ahead of a patched/experimental build), whichever one resolves first is the one probed, which
  can silently disable streaming even if the "real" nu you intend to run supports it. `nu.go`
  exposes `POSH_NU_STREAMING_EXECUTABLE` to override which binary gets probed, bypassing PATH
  order for exactly this scenario.
- `def <name> [...] { ... }` cannot be redefined in the same parsed scope - a second `def` with an
  identical name is `nu::parser::duplicate_command_def`, a parse error, even if the first
  definition is never called. This rules out "define a default, then redefine it later in the
  same script" as a way to layer optional behavior; reassigning an `$env.*` variable that holds a
  closure is fine (last write wins), but redefining a named `def` is not.
- `job spawn { ... }`, `job id`, `job send <id> --tag <n>`, `job recv --tag <n> --timeout <dur>`
  form a real mailbox: `job id` called from inside a spawned job's own block returns that job's
  id (usable to correlate a specific spawn with later sends/recvs across prompt cycles). `job
  send` never blocks. `job recv` blocks indefinitely without `--timeout`; with `--timeout` it
  **raises** `nu::shell::job::recv_timeout` on expiry rather than returning `null`, so callers
  need `try { job recv ... } catch { <fallback> }` to implement a bounded wait with fallback.
- `bytes split 0x[00]` on a piped external command's stdout streams incrementally in real time
  (verified via a delayed HTTP-backed test segment: first `oh-my-posh stream` record arrived
  bounded by the segment's own timeout, the resolved final record arrived at the segment's real
  completion time, not at process exit). Do not validate this from a wrapping PowerShell harness
  piping the same external command - `ForEach-Object` over an external process is line-buffered
  and silently waits for process exit when there's no `\n`, producing a false "arrives all at
  once" result. Test incremental behavior from inside nu itself.
- `continue` is only valid inside `for`/`while`/`loop`, not inside an `each {|x| ...}` closure
  (nu treats that closure body as a function boundary, not a loop). Use `for` (which also allows
  `mut` state carried across iterations) when the loop body needs `continue`.
- Multi-line external-command invocations (`^cmd --flag value` spread across lines) don't parse;
  either keep the whole call on one flat expression or build `let args = [...]` and spread with
  `...$args`.
- `bytes starts-with` needs **binary** input - check it on the raw record before `decode utf8`,
  not on the decoded string.
- `src/shell/init.go` has **two separate nu codegen entry points** that both need to apply the
  same feature-gating: `generateNuScript` (used by the real `initNu` autoload-write path that
  runs at actual shell startup) and `generateScript`'s `NU` case (used by `Script()`, which backs
  the `--print` flag and is what manual/CLI testing normally goes through). A capability check
  added to only one of them will look broken during `--print` testing while working correctly at
  real shell startup, or vice versa - always add gating logic to both, and prefer a small shared
  helper (e.g. `nuSupportsStreaming(env)`) called from each site rather than duplicating checks.
- **A persistent `oh-my-posh serve` daemon (the pwsh/zsh/fish/cmd pattern) is NOT achievable in
  nu via a `job spawn`-fed external-command-stdin "coprocess", despite an LLM research pass
  concluding otherwise.** Verified empirically 2026-08-04 against real nu 0.114.1/0.114.2 (both
  stock and the patched `async-prompt` build): a lazy `generate {|_| {out: (job recv), next:
  null}}` stream (or any `job recv`-driven loop) piped into an external command
  (`generate {...} | ^exe`) delivers **zero bytes** to the child's stdin - not even one write,
  no error, exit code 0 - regardless of item count. A finite, already-materialized `list` piped
  to an external gets rendered as a **formatted table string** (e.g. `│ 0 │ a │`), never as raw
  line-delimited bytes. Only a single, already-built `string`/`binary` *value* (not a stream)
  reaches an external's stdin as raw bytes. This directly contradicts the plausible-sounding
  claim (sourced from reading `run_external.rs`'s doc comments/thread names, e.g. "external
  stdin worker") that Nushell's external-stdin plumbing forwards a lazy `ListStream`
  value-by-value in real time as it's produced - in practice, for `generate`/list-sourced input,
  nothing streams through at all. **Do not trust source-reading-based claims about nu's
  ByteStream/ListStream stdin behavior without a hands-on repro; test the exact `generate | ^exe`
  or `job recv-loop | ^exe` shape directly before designing around it.**
- Corollary: the only known-working nu analogue to fish's persistent daemon is fish's own
  mechanism - a Unix FIFO (`--request-pipe` in `src/cli/serve.go`, already implemented) that the
  daemon holds open for both ends, with the client doing an **open-write-close per request**
  (never holding a live stdin handle open across cycles) using ordinary file I/O (nu's `save
  --append <fifo-path>`), and a separate `job spawn { open --raw <out-fifo> | bytes split 0x[00]
  | each {...} }` reading the daemon's own stdout redirected to a second FIFO. **This recipe is
  now verified working end-to-end** (2026-08-04, WSL/aarch64, real nu 0.107.0 + a cross-compiled
  real `oh-my-posh` binary) and implemented in `src/shell/nu.go`'s `Streaming` case
  (`_omp_serve_start`/`_omp_serve_stop`/`_omp_serve_render`). The FIFO transport is Unix-only
  (gated by `which mkfifo | is-empty`); on Windows the same daemon runs over an HTTP loopback
  transport instead (see below).
- **`job spawn { ^exe args... out> $fifo }` (a direct external-command file redirect written
  straight inside a job-spawn closure body) silently never forks the child process at all** - no
  error, `job spawn` still returns a job id, but `ps`/`pgrep` show nothing. This is NOT a general
  "externals in job spawn are broken" problem: `job spawn { ^sleep 30 }`, `job spawn { ^exe
  --version }`, and even `job spawn { ^exe serve ... }` (no redirect) all fork correctly. It is
  specifically `out>` (and `| save --raw/--force <path>`) used *directly* inside a job-spawn
  closure that fails to fork - and the exact same redirect works fine as a **top-level** `nu -c
  '...'` call. **Workaround (verified, now in production code):** wrap the daemon invocation in a
  **nested nu subprocess** - `job spawn { ^$nu.current-exe --no-config-file -c '^$exe serve ...
  out> $fifo' }` - so the `out>` redirect is parsed at that nested nu's own top level rather than
  inside the outer job-spawn closure. Root cause not fully isolated (likely something about how
  job-spawn's thread context handles parser-level redirect setup for externals vs. top-level
  command execution) but the workaround is solid and reproducible.
- **Top-level `let` bindings from the sourced script are NOT visible inside a `job spawn { ... }`
  closure body**, even though the same top-level `let` IS visible inside an ordinary `def`'s body
  in the same file (e.g. `$_omp_executable` works fine inside `_omp_get_prompt`/
  `_omp_stream_primary`, both plain `def`s, but referencing it directly inside a `job spawn {}`
  closure raises `nu::shell::variable_not_found`). Fix: capture it into a **local** `let` right
  before the `job spawn` call (e.g. `let exe = $_omp_executable` then use `$exe` inside the
  closure) - a closure captures locals from its immediately enclosing function scope, not
  arbitrary outer module-level state. This also applies across a `source` boundary when testing
  standalone (a `let` in the sourcing script is not visible inside a `def` from a separately
  `source`d file) - always test embedded-script behavior as a single combined file, matching how
  `oh-my-posh init nu` actually emits one script, not as two files joined by `source`.
- `def` functions do **not** propagate `$env.*` mutations back to the caller unless declared with
  `def --env`. Without `--env`, any `$env.foo = ...` inside the function body is invisible once
  the function returns - this silently breaks any pattern that stashes daemon/fifo/job-id state in
  env vars across calls (every call looks like "first call", e.g. a fresh daemon gets spawned on
  every single prompt instead of being reused). All three serve functions
  (`_omp_serve_start`/`_omp_serve_stop`/`_omp_serve_render`) must be `def --env`. This also bites
  test harnesses: invoking a stored closure via `do $env.PROMPT_COMMAND` does NOT propagate its
  env mutations back either - use `do --env $env.PROMPT_COMMAND` (real prompt-command invocation
  inside nu itself already behaves like `--env`; only a manual `do` call in a test script needs
  the flag spelled out).
- `job send <value> --tag <id>` / `job recv --tag <id>` requires the **same** tag on both sides -
  omitting `--tag` on the sending side (e.g. forgetting it in `job send 0 --tag $self_id`) means a
  tag-filtered `job recv` on the other side never matches and always times out, with no error
  indicating a mismatch. Always pass `--tag` symmetrically.
- `$env.PATH` is a **list** in Nu, not a string. Putting it straight into a record destined for
  `to json --raw` (e.g. `{PATH: $env.PATH, ...}`) serializes it as a JSON array; if the Go side
  decodes that JSON into a `map[string]string` (as `serveRequest.Env` in `src/cli/serve.go` does),
  `encoding/json` fails to unmarshal an array into a string field, and the **whole request line**
  is silently dropped (`serve.go`'s `runServeLoop` ignores malformed JSON lines for forward/
  backward compatibility) - the daemon keeps running, but no response ever arrives, with zero
  errors on either side. Always join it first: `$env.PATH | str join ":"`.
- `job kill <jid>` on a job spawned via the nested-nu-wrapper pattern above only reaches the
  **direct child** (the nested nu wrapper process), not the real `oh-my-posh serve` process it
  launched as its own child - `ps` shows the real serve process orphaned and still running after
  `job kill`. Use the existing graceful protocol instead: write `{"command":"quit"}\n` to the
  daemon's request fifo (matches `serveCommandQuit` in `serve.go`), which makes the real daemon
  exit cleanly, which in turn lets the nested nu wrapper's script line finish naturally.
- **Windows named pipes are a dead end for nu-driven daemon I/O** (verified 2026-08-04, nu
  0.114.2, Windows 11): nu *can* read from `\\.\pipe\<name>` via `open --raw` (the server must
  be multi-instance and tolerate nu's extra probe connections - nu opens the path once for
  metadata before the real read), but **every nu write primitive fails against a pipe path**:
  `save --append` → "permission denied" (Rust's `OpenOptions::append` is invalid on pipes),
  `save --force`/`o>` → "truncate not supported", plain `save` → "file exists", `o>>` →
  silently writes nothing. With no way to send requests, a two-pipe FIFO-analogue design is
  impossible no matter how the Go server is written. Don't re-litigate this with go-winio
  server tweaks - the blocker is nu's client-side write path, not the server.
- **The working Windows transport is HTTP over loopback** (implemented 2026-08-04 in
  `src/cli/serve_http.go` + the `_omp_serve_start_http` branch of `src/shell/nu.go`): the
  daemon binds `127.0.0.1:0`, writes `"<port> <token>\n"` atomically (temp file + rename) to a
  `--port-file` the shell polls, and serves two endpoints - `POST /` (one JSON request per
  prompt, same wire format as the fifo/stdin transports) and `GET /stream` (one long-lived
  chunked response carrying every NUL-delimited record for the daemon's lifetime, held open by
  the shell's reader job). Nu's built-in `http get ... | bytes split 0x[00]` consumes chunked
  responses **incrementally** (verified: first record ~3ms after POST, later records arrive
  live mid-response), and `http post --max-time 3sec` is a fire-and-forget request writer, so
  the shell script shape is identical to the fifo recipe with `http` calls swapped in for
  `save --append`/`open --raw`. Measured end-to-end on the same machine as the earlier
  baselines: warm prompt cycles ~9-12ms, 6-cycle average incl. cold daemon start ~21ms, vs
  ~57ms for the per-prompt subprocess fallback and ~43-48ms for the plain streamed spawn.
  A 16-byte hex token (`X-Omp-Token` header, constant-time compare) gates both endpoints
  since any local process can reach loopback. Caveat: `HTTP_PROXY`-style env vars could in
  principle divert loopback requests if nu's client honors them without a `NO_PROXY`
  exemption - if a machine shows instant permanent fallback (3 failures), check proxy env.
- **Piping a stream through a `def` that consumes `$in` destroys incrementality** (verified:
  first record arrives at response *completion* ~1.5s instead of ~3ms when the
  `bytes split`/`each` loop is moved into a helper `def`). The reader loop must live inline in
  the `job spawn` closure body - the duplication between the fifo and HTTP reader jobs in
  `nu.go` is deliberate, do not refactor it into a shared `def`.
- `$nu.temp-path` no longer exists in nu 0.114.x - it was renamed to **`$nu.temp-dir`**. Since
  any nu build new enough for the Streaming feature (needs `commandline set-prompt`) is well
  past the rename, generated scripts must use `temp-dir`; `temp-path` raises
  `nu::shell::name_not_found` at runtime (it passes parsing, so it only explodes when the
  serve path actually executes).
- `o+e> NUL` works at the top level of a nu script on Windows (`NUL` resolves to the Win32
  null device; no literal `NUL` file is created) - used by the nested-nu daemon wrapper to
  silence the daemon's stderr, mirroring the `out> $fifo` role the wrapper plays on Unix.
- The serve daemon `os.Chdir`s into each render request's `pwd` (`startRenderCycle` in
  `serve.go`), and Windows locks a process's working directory against deletion - so a
  directory you just `cd`'d out of can still be undeletable ("being used by another process")
  until the *next* prompt render moves the daemon elsewhere. The pwsh daemon has the same
  behavior; it only surfaces in scripted tests that render in a temp dir and immediately
  `rm` it.
- `$env.CMD_DURATION_MS` only exists in an interactive REPL session - a script that `source`s
  the init and calls `_omp_stream_primary` directly (e.g. an E2E harness) must set it
  manually first or the `match $env.CMD_DURATION_MS` line raises `column_not_found`.
- `which mkfifo | is-empty` is a reliable, portable way to gate Unix-FIFO-only functionality
  without any Go-side platform detection/build tags - it's simply empty (falsy) on Windows, no
  error.
- `to json --raw` on a nu record (including nested records, e.g. an `env: {...}` sub-record)
  produces compact JSON that round-trips cleanly through Go's `encoding/json` - safer than
  hand-building JSON strings with manual escaping (as `omp.zsh`'s `_omp_serve_escape` does for
  lack of a native serializer in zsh).
- `$nu.current-exe` is the correct, PATH-independent way to get the path to the currently-running
  nu binary (needed to launch the nested-nu daemon wrapper above) - prefer it over `which nu | get
  0.path`, which can raise `nu::shell::access_beyond_end` if the `which` result table happens to
  be empty in some invocation contexts.
- **Nu evaluates `$env.PROMPT_COMMAND` and then `$env.PROMPT_COMMAND_RIGHT` sequentially on the
  same thread before painting anything** (verified in nushell's `prompt_update.rs`,
  `update_prompt`: `get_prompt_string(PROMPT_COMMAND, ...)` runs to completion before
  `get_prompt_string(PROMPT_COMMAND_RIGHT, ...)`). With an `rprompt` block configured, the base
  `omp.nu` right closure (`_omp_get_prompt right`) added a second full blocking process spawn to
  every prompt cycle, on top of the streamed primary - measured 2026-08-04 on Windows: left
  (streamed, per-prompt fallback) ~46ms + right ~34ms = ~80ms per cycle. The Streaming feature in
  `nu.go` now overlaps them: `PROMPT_COMMAND` spawns the right render as a `job spawn` before
  blocking on the primary's first record, collects it via `job recv` afterward, and
  `PROMPT_COMMAND_RIGHT` just reads `$env._omp_right_prompt` (an env write from `PROMPT_COMMAND`
  propagates to the right closure; the right closure itself never needs to write env, avoiding any
  dependency on whether its own mutations would propagate). Same-binary before/after: ~47ms →
  ~34.5ms per full cycle, right prompt byte-identical.
- Head-to-head prompt-cycle numbers on the same Windows machine/config/binary (2026-08-04,
  measured, not estimated): pwsh serve daemon warm ~12-15ms; pwsh legacy per-prompt stream spawn
  (daemon disabled) ~61-63ms; nu streaming per-prompt fallback ~43-48ms (left only). Nu's
  fallback is *faster* than pwsh's own no-daemon fallback - the entire "nu feels slower than
  pwsh" gap on Windows was (a) pwsh's working serve daemon (nu's was Unix-FIFO-only until the
  HTTP loopback transport above closed that gap: nu warm cycles now ~9-12ms, on par with pwsh's
  ~12-15ms) plus (b) the
  extra sequential right-prompt spawn above, now overlapped. Nothing measured suggests inherent
  nu-runtime overhead: raw external spawn cost is comparable from both shells (`^cmd /c exit`
  ~40ms from nu vs `& cmd /c exit` ~35ms from pwsh; Windows process creation dominates),
  `job spawn` + mailbox round-trip costs ~0.1ms, `term size`/`job list | length` are ~0.05ms.
- Per-prompt housekeeping costs worth knowing (Windows, nu 0.114.2): `history` with a ~14k-entry
  SQLite history takes ~9ms per call (and `| length`/`| last 1` don't make it cheaper - the
  full read happens regardless), so the history-growth bookkeeping in `PROMPT_COMMAND` has a
  real, history-size-scaling cost; `which mkfifo` costs ~7ms per call (moot on Windows now that
  `_omp_serve_start` dispatches on `$nu.os-info.name` before any mkfifo probe).
- A `oh-my-posh.exe` resolved through the WindowsApps app-execution alias
  (`...\Microsoft\WindowsApps\oh-my-posh.exe`, MSIX install) costs ~15-20ms *extra per process
  spawn* versus invoking the same-version binary at a plain filesystem path (measured: right
  render ~34ms via alias vs ~15ms via direct path, `--version` spawn ~87ms avg via alias vs
  ~29ms direct from pwsh). Benchmarks of per-prompt spawn cost must pin the executable path, or
  the alias tax drowns out the effect being measured.
- **Known unresolved limitation (Unix fifo transport only):** a job spawned via `job spawn`
  outlives the nu process that spawned it if that process exits normally (verified: `nu -c 'job
  spawn { ^sleep 30 }; print done'` prints "done" and exits while `sleep 30` keeps running,
  orphaned). Nu has no general "on shell exit" hook (only
  `pre_prompt`/`pre_execution`/`env_change`/`display_output`), unlike zsh's `zshexit_functions`
  or fish's `fish_exit`. This means the fifo-transport daemon (and its nested-nu wrapper and
  reader job) can be left running as orphaned background processes if a nu session ends abruptly
  (window closed, terminal killed) without ever calling `_omp_serve_stop` - there is currently
  no nu-side mechanism to guarantee cleanup in that case, unlike the fd-ownership-based natural
  cleanup zsh/fish get from their coprocess mechanisms. **The HTTP transport does not share this
  gap:** when the nu process dies for any reason its jobs die with it, the `/stream` TCP
  connection drops, and the Go server treats that disconnect as the shell-disappeared signal
  (the SIGPIPE analogue) and exits gracefully - plus a belt-and-braces guard shuts the daemon
  down if no stream client connects within 30s of startup.

