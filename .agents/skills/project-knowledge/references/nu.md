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
  | each {...} }` reading the daemon's own stdout redirected to a second FIFO. This is
  plausible in principle (FIFO reads/writes are ordinary file syscalls, not the broken
  external-stdin-streaming path above) but **unverified** - no Unix nu build was available to
  test against in this environment - and it would be Unix-only regardless (Windows named pipes
  need `CreateNamedPipe`-based server plumbing on the Go side, a much larger change, and were not
  attempted). Windows nu process-start overhead (measured ~50ms for a minimal config, before any
  segment work, via `oh-my-posh print primary`) therefore has **no known nu-scriptable fix**
  today; it is an accepted cost of the current per-prompt `oh-my-posh stream` design until nu
  gains an official coprocess/persistent-external primitive (tracked upstream in
  nushell/nushell#18486, open as of 2026-08-04, not yet implemented).
