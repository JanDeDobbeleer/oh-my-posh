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

## config/shell package decoupled from segments and dsc for wasm (verified 2026-07-30)

- `config/config.go` and `config/default.go` referenced bare `segments.CONST` string option-keys
  (e.g. `segments.Source`, `segments.BranchTemplate`) directly and unconditionally, which pulled the
  entire `segments` package - and its transitive HCL/go-cty/`golang.org/x/mod/modfile`/`gopkg.in/ini.v1`/
  `cli/auth` dependency tree - into every binary linking `config`, including `src/wasm`. Fix: mirror
  the option-key strings as local `options.Option` consts in the `config` package itself (`options` is
  the small standalone package the keys' type lives in - importing it alone costs nothing heavy) and
  drop the `segments` import entirely. Verified with `go list -deps ./wasm/` that `segments` is fully
  gone from the wasm dependency graph after this change.
- `src/dsc` (the CLI-only DSC/Desired State Configuration resource wrapper) pulls in
  `github.com/invopop/jsonschema` (→ `go/ast`, `go/parser`, `go/doc`) and `github.com/spf13/cobra`.
  Both `config/dsc.go` (the `Configuration`/`Resource` wrapper around parsed config files) and
  `shell/dsc.go` (the `Shell`/`Resource` wrapper around shell-rc-rewriting) imported it unconditionally,
  even though DSC tracking is purely a CLI bookkeeping feature never exercised by the wasm render path
  (`render.Config`/`config.ParseBytes`/`config.ParseData` never touch it). Fix: moved both DSC resource
  types into a new `src/cli/dsc` package (so they live where they're actually used) and decoupled the
  `config` package from `dsc` via a nil-by-default hook: `config.NewDSCTracker func() config.DSCTracker`,
  set by `cli/dsc`'s `init()`. `config.Parse()` (the file-loading path, `config/load.go`) checks the hook
  and no-ops when nil (i.e. in any binary that never imports `cli`, such as wasm) instead of directly
  constructing a `dsc.Resource`. `shell/dsc.go`'s shell-config-rewriting logic moved wholesale into
  `cli/dsc/shell.go` (renamed `Shell`→ still `Shell`, just in the new package) since nothing in `shell`
  itself needs it - only `cli/init.go`'s `init` command does.
- Combined effect measured on the post-terminal-split wasm baseline (21,833,910 bytes stripped):
  dropped to **17,825,752 bytes (~4.01 MB / ~18.4% additional reduction)**. Confirmed via
  `go list -deps ./wasm/` that `segments`, `dsc`, `invopop/jsonschema`, `spf13/cobra`, `go/ast`,
  `go/parser`, and HCL/go-cty are all absent from the wasm link graph after this change.
- Pattern for any future "CLI-only logic embedded in a shared package" cleanup: introduce a
  package-level func-var hook (nil by default) in the shared package, and have the CLI-only package
  register the real version via `init()`. This avoids the shared package importing the heavy
  dependency at all, while keeping the CLI behavior identical when `cli` (or whichever package sets
  the hook) is actually linked.
