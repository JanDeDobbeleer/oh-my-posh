# Codebase

## Docs linting

- Two markdown gates with different coverage (verified 2026-07-14): the Vale CI workflow
  (`.github/workflows/vale.yml`) explicitly lints `AGENTS.md`, `.github/copilot-instructions.md`,
  and `.agents/skills`, while `markdownlint-cli2` skips dot-directories entirely - its globs never
  match `.agents/` or `.github/`, even when passed explicit paths. Lint skill docs with Vale
  before pushing; for markdownlint, copy them to a non-dot directory alongside
  `.markdownlint-cli2.yaml`.
- Vale fails CI on error-level findings only; warnings pass. Justified terms (Go interface
  wording, zsh feature names) get file-scoped rule overrides in `.vale.ini`, each with a comment.

## Dev environment

- The Go module root is `src/`, not the repo root - run all `go` commands from there.
- On windows/arm64 dev machines `go test -race` is NOT supported. Concurrency-sensitive changes
  must rely on CI (amd64) for race detection.
- Rendering hot-path benchmarks live in `src/template/bench_test.go`,
  `src/terminal/bench_test.go`, and `src/prompt/bench_test.go`; compare runs with `benchstat`.
- `template.Init` resets the parsed-template cache - a macro benchmark that calls it per iteration
  measures the cold-parse path, not steady state.

## Shell integration scripts

- Everything under `src/shell/scripts/` is **embedded at build time** (`go:embed`). After editing a
  script, rebuild the binary before testing; `oh-my-posh init <shell> --print` shows the generated
  output and is the fastest way to inspect what a user actually sources.
- Features (transient prompt, tooltips, vi mode, streaming) are emitted per shell from
  `src/shell/<shell>.go` - a script function is dead code unless the feature switch emits its
  activation line.

## Segments and panics

- Segment `Execute` runs in bare goroutines with **no recover** (`src/prompt/segments.go`), and
  template rendering re-panics runtime errors. Any panic there kills the whole process - the user
  sees a completely blank prompt. So when a user reports a blank prompt: find the panic.
- If the panic trigger persists (e.g. a poisoned cache entry with a TTL), every prompt crashes
  until the entry expires.
- Segment writers gob-encode only exported fields. `segments.Base.env/options` are unexported and
  MUST survive a cache restore: overlay the restored data onto the writer initialized by
  `MapSegmentWithWriter`, never replace the writer.
- `runtime/cmd.RunWithEnv` applies `strings.TrimSpace` to the complete command output before
  returning it (verified 2026-07-26). If a segment encodes an empty final field as a trailing blank
  line, that field is lost. Use a record format that retains a final non-whitespace delimiter or
  sentinel, and validate the record before assigning parsed state.

## Cache

- Cache persistence only happens with the hidden `--save-cache` flag (print/stream commands);
  without it, stores never write on close. Redirect the location with `OMP_CACHE_DIR`.
- Debug logs are buffered and only printed by the `oh-my-posh debug` command (grep for
  `restored segment from cache` / `setting entry`). `POSH_TRACE=1` and stderr show nothing for
  print commands.
- On Windows the cache file is a memory-mapped 50KB+5 "persistent shared string" with a 4-byte
  length header; a fresh file is all zeros and logs a harmless `store.go:init EOF` error on first
  read.

## Streaming and serve daemon

- Streaming is enabled by the top-level `"streaming": <ms>` config key. That value is ALSO each
  segment's pending-timeout and overwrites segment-level `timeout`.
- `stream` always emits the transient prompt as a `\x1e`-prefixed NUL record (initial + refreshed
  once all segments resolve); serve records are `<id>\x1f<payload>\0`.
- Serve pitfall class: process-lifetime initializers (`if X != nil return`) pin first-render state
  in a daemon. `template.Cache` did exactly that (pinned PWD/Folder/Code/Jobs) - fixed with
  `template.ResetCache()` per render in `startRenderCycle` (flag-based rebuild; never nil the
  global, abandoned segment goroutines may still read it). Audit for this pattern when extending
  serve.
- Do NOT memoize the config in serve: the per-render gob decode is load-bearing - a fresh segment
  graph per cycle isolates the active render from abandoned-cycle goroutines holding pointers into
  their own graph.
- Daemon tests must vary per-request context (cwd, status) across cycles; single-context tests
  cannot catch one-shot-assumption state.
- `config.Get` prefers the session gob cache over `POSH_THEME`.
- Go guarantees exactly 2 records per wait-mode serve request even on segment panic
  (`renderComplete`) - blocking clients (Clink) rely on this.

## js/wasm runtime footprint (verified 2026-07-30)

- gopsutil (v4.26.6) ships `*_fallback.go` build-tag stubs for `GOOS=js` across process/host/mem/
  load/disk/net/cpu/sensors, so it links fine into the `src/wasm` build without any split - it is
  NOT a multi-MB cost by itself. The real, measurable cost was `runtime.Terminal.Shell()`'s parent-
  process fallback (`gopsutil/process.NewProcess`) and `SystemInfo()` (`gopsutil/load` +
  `gopsutil/disk` + `Memory()`/`gopsutil/mem`) plus `TerminalWidth()`'s
  `wayneashleyberry/terminal-dimensions` (which shells out to `stty` via `os/exec`). `go build
  -ldflags=-dumpdep` (GOOS=js) showed `gopsutil/process.Process` retaining dozens of exported
  methods' `.namedata` (CPUPercent, Connections, Children, ...) even though wasm code only ever
  calls `.Name()` - all of it is reachable, not DCE'd. Splitting those three methods into
  `!js`/`js` file pairs on the existing `*Terminal` type (mirroring `terminal_root_js.go` /
  `terminal_writable_js.go`) dropped the stripped `omp.wasm` from 21,866,822 to 21,833,786 bytes
  (~33 KB) and fully eliminated `gopsutil/process`, `gopsutil/cpu`, `gopsutil/load`, and
  `terminal-dimensions` from the dumpdep graph - only the type-only `gopsutil/disk.IOCountersStat`
  (referenced by `SystemInfo.Disks`, unavoidable without touching `environment.go`) remains.
- Why per-method `_js.go` splits are the only viable shape, not a new Environment type:
  `text/template` calls `reflect.Value.MethodByName` with a non-constant name, so the linker keeps
  every exported method of any type that gets converted to an interface in live code. `render.Config`
  does exactly that (`env := &runtime.Terminal{}` assigned to the `Environment` interface), so ALL
  of `*Terminal`'s methods stay linked regardless of which ones the wasm path actually calls - the
  only way to drop a method's cost is to replace its body for the `js` tag,
  not to introduce a slimmer type.
