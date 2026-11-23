# Codebase

## Template markup trust boundary

- Print-action output is chevron-escaped (`escapePrintActions`); only `template.Markup` values keep `<...>`
  anchors. Keep attacker-controlled data (VCS refs, API responses, folder names) in plain strings. `RawMarkup` is for
  user config only, `EscapeMarkup` for data, `JoinMarkup` for composition.
- Segment fields built from option strings that carry anchors (icons, `branch_icon`, `folder_separator_icon`,
  `status_formats`) must be `template.Markup`, or the anchors render as literal text.
- Keep `template.Markup` a named string, never a struct: text/template treats every struct as true (breaks
  `{{ if .BranchStatus }}` guards in shipped themes) and `eq` cannot compare it to a string.
- Untrusted templates (`RenderUntrusted`, path segment) escape Markup results too. The path segment escapes folder
  names with `template.EscapeSource` on every branch of `replaceMappedLocations`; a new early return needs its own
  escape.
- New func-map entries: output not derived from input (`readFile`, `cmd`, `b64dec`, `env`) goes in `noPromote`;
  attacker-chosen counts (`repeat`, `seq`, `until`, `rand*`) go in `dangerousFuncs` (trusted templates only). For a
  hot function, add a typed case to `markupAwareTyped` first.
- Never bypass `formats.EscapeSequences` in `terminal.write`'s `isHyperlink` branch: bash `@P` re-interprets a
  backslash in the OSC 8 URI.
- `src/prompt/golden_test.go` renders every shipped theme byte-exact. A golden diff after a markup change means a
  theme relied on the old, insecure behavior: inspect it, then `go test ./prompt/... -run TestGoldenThemes -update`.

## Terminal output sanitization

- `prompt.Engine.write` bypasses `terminal.write`'s per-rune control filter. Sanitize attacker-influenceable strings
  written in one shot (console title, OSC payloads) with `terminal.stripControlRunes`; `trimAnsi` misses ESC-less
  input (bare BEL) and CSI `! p`, APC, SOS, ST.
- Don't extend `terminal.AnsiRegex`'s final-byte class (`!`, `_`, `X`, `\` omitted on purpose) as a sanitization fix;
  strip control runes instead.

## Segments and cache

- After adding config fields, regenerate the embedded DSC schema with
  `UPDATE_SCHEMAS=1 go test ./cli/dsc ./cli/font`; shared segment fields affect it even without DSC code changes.
- Treat `mem.SwapMemory()` errors as returning a nil stat: gopsutil can fail reading `/proc/vmstat` on Android/Termux,
  so fall back to `VirtualMemoryStat.SwapTotal` and `SwapFree` instead of dereferencing the result.
- Segment writers gob-encode only exported fields; `segments.Base.env/options` are unexported. On cache restore,
  overlay restored data onto the writer from `MapSegmentWithWriter`; never replace the writer.
- `runtime/cmd.RunWithEnv` trims whitespace from the whole output, so a trailing empty field is lost. End records
  with a non-whitespace delimiter or sentinel and validate before assigning parsed state.
- A poisoned cache entry replays its failure every prompt until its TTL ends.
- Caches persist only with the hidden `--save-cache` flag (print/stream); relocate with `OMP_CACHE_DIR`.
- Don't infer segment visibility from template data: segments populate it before a whole right-aligned block can be
  rejected for overflow. Tooltip suppression uses the accepted-output visibility snapshot instead.
- `OMP_CACHE_DIR` must name an existing parent directory: `cache.Path` only creates its `oh-my-posh` child and
  falls back to the normal cache otherwise. Create the parent first when isolating real prompt tests.
- Logs are buffered and print only via `oh-my-posh debug` (grep `restored segment from cache`, `setting entry`);
  `POSH_TRACE=1` and stderr show nothing for print commands.
- Windows: a fresh memory-mapped cache file logs a harmless `store.go:init EOF` on first read.

## Streaming and serve daemon

- Default-on config booleans must preserve an explicit `false` through the gob cache: use `config.Bool` rather
  than `*bool`, because gob omits a plain pointer-to-false struct field and restores it as nil.
- Disabling transient pre-rendering must still emit a marker-only transient record: shell integrations fall
  back to on-demand rendering for an empty payload, and Clink's wait-mode client requires both records.
- With transient pre-rendering disabled, publish the session snapshot before the marker-only record: the
  on-demand CLI fallback cannot access the serve daemon's in-memory template context.
- `"streaming": <ms>` also becomes every segment's pending timeout, overriding segment `timeout`.
- `stream` emits the transient prompt as a `\x1e`-prefixed NUL-terminated record initially and after the final refresh,
  except zsh with a transient `right_template` and default/true `pre_render`. Non-JSON serve records are
  `<id>\x1f<payload>\0`; a wait-mode render that reaches `renderComplete` yields exactly 2, even on segment panic,
  because Clink blocks on that count.
- A timed-out segment keeps running after its cycle. On any path that can run then, publish through the
  `renderCache` captured at the top of `Execute`, never the global `template.Cache` (serve resets it per cycle,
  tests reassign it). `restoreCache`/`restoreData` still read the global (open race, 2026-09).
- Don't gate `Execute`'s cache-hit early return on streaming: placeholders never show cached data, so skipping the
  cache only adds live probes that can time out.
- Never read a killed, pending, or abandoned segment's mutable state (`writer`, `Enabled`): its `Execute` goroutine can
  still write after the cycle stops waiting. Check these states first; pending segments render through
  `Segment.RenderPlaceholder` (config only). Don't add a `Pending` flag to `config.Segment`.
- The event channel in `prompt/streaming.go` is sized `2*segments+1` so plain sends never block or drop; don't add
  `select`/`default`. The producer's `recover` must log a stack: a silent recover once hid a stuck-segment bug.
- `prompt/streaming_writer_test.go`'s fake writer writes fields late on purpose, as race bait. `-race` only runs in
  CI ([release](release.md)).
- Serve: process-lifetime initializers (`if X != nil return`) pin first-render state. Reset per render
  (`template.ResetCache()` in `startRenderCycle`); never nil a global that abandoned goroutines may still read.
- Don't memoize the config in serve: the per-render gob decode gives each cycle its own segment graph, isolated from
  abandoned-cycle goroutines.
- Daemon tests must vary per-request context (cwd, status) across cycles to catch one-shot state.
- Timing races: `oh-my-posh debug` timings are cold-process. Sweep the `streaming` value against a scripted serve
  session (JSON line + env blob on stdin) and count cycles whose last record still holds the placeholder.

## Rendering and statusline

- `template.Init` resets the parsed-template cache: calling it per benchmark iteration measures cold parse.
- Stale website previews: a set `OMP_BIN` overrides the auto-rebuilt `src/bin` in `ensure-artifacts.mjs`.
- Unix `terminal-dimensions` runs `stty size` on stdin, which is not a TTY under Claude Code's statusline. A valid
  `COLUMNS` fallback must clear the `stty` error, or `prompt.Engine.canWriteRightBlock` rejects right-aligned blocks.
