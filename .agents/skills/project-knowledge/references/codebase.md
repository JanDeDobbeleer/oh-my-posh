# Codebase

## Docs linting

- Two markdown gates cover skill docs (updated 2026-07-30): the Vale CI workflow
  (`.github/workflows/vale.yml`) explicitly lints `AGENTS.md`, `.github/copilot-instructions.md`,
  and `.agents/skills`, and `markdownlint-cli2`'s `**/*.md` glob in `.markdownlint-cli2.yaml` also
  reaches `.agents/skills` - only the two explicit `ignores` entries there are excluded. Lint skill
  doc changes with both `vale <path>` and `npx markdownlint-cli2 --config .markdownlint-cli2.yaml
  <path>` before pushing.
- Vale fails CI on error-level findings only; warnings pass. Justified terms (Go interface
  wording, zsh feature names) get file-scoped rule overrides in `.vale.ini`, each with a comment.

## Windows git rebase with core.autocrlf=true

- An interactive `git rebase --autosquash` can stop mid-sequence with "Your local changes to
  `<file>` would be overwritten by merge" on a plain `pick` that has no real content conflict
  (verified 2026-07-30). Root cause: `core.autocrlf=true` renormalizes line endings on checkout,
  which git treats as a working-tree modification that blocks the next pick. Toggle
  `git config core.autocrlf false` for the duration of the rebase (restore it after), rather than
  trying to resolve a conflict that doesn't reflect the actual diff.

## SVG renderer (src/svg)

- The window chrome's "shadow" must be a filled silhouette rect offset +4/+4 behind the window
  (`writeWindowChrome`, `shadowOffset`), never an `feDropShadow` filter (verified 2026-09-04).
  A filter casts the silhouette of what is actually painted; the window's border rect is
  `fill="none"`, so a filter shadows only the 1px stroke and no solid block appears on the
  right/bottom. The CSS it mirrors is `box-shadow: 4px 4px 0 var(--omp-frame-shadow)` on
  `.theme-code-block` in `website/src/css/custom.css` - a solid offset block, black-ish in
  light mode, white in dark mode.
- The canvas grows by exactly `shadowOffset` on the right/bottom so the silhouette is not
  clipped by the viewBox; tests that count `<rect` elements must include the shadow rect.
- After changing `src/svg/`, regenerate the website previews or the homepage/theme gallery stay
  stale: build from `src/` (`go build -o ..\oh-my-posh.exe .`), then run
  `node website/export_themes.mjs` with `OMP_BIN` pointing at that binary.

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

## Terminal output sanitization

- `prompt.Engine.write` (src/prompt/engine.go) appends straight to the prompt builder and
  bypasses the per-rune control filter in `terminal.write` (src/terminal/writer.go). Any
  attacker-influenceable string passed to `e.write` in one shot (console title, OSC payloads)
  must sanitize itself with `terminal.stripControlRunes`; `trimAnsi` alone is insufficient -
  it ignores strings without ESC (a bare BEL survives) and its regex misses CSI `! p`, APC,
  SOS, and ST (verified 2026-09-05, GHSA-fwjx-9p69-h25h follow-up in `FormatTitle`).
- `terminal.AnsiRegex`'s final-byte class intentionally omits `!`, `_`, `X` and `\`; do not
  extend it as a sanitization fix - strip control runes instead.

## Template markup trust boundary (2026-09-05, markup-injection fix)

- The renderer escapes chevrons in every print action's output (`__esc` appended to each
  pipeline post-parse, `template/render.go` `escapePrintActions`). Literal template text keeps
  its `<...>` anchors; action output only keeps them when typed `template.Markup`. A new
  segment that stores attacker-controlled strings (VCS refs, manifest fields, API responses,
  folder names) in plain string fields is safe by default - do NOT call `template.RawMarkup`
  on data.
- `template.Markup` (src/template/markup.go) is the bypass type: a named string, never a
  struct (text/template treats every struct as true, which broke the `{{ if .BranchStatus }}`
  guards in 60 shipped themes, and `eq` cannot compare a struct to a string). Constructors:
  `RawMarkup` (user config only), `EscapeMarkup` (data),
  `JoinMarkup` (composition). Segment fields composed from option strings with anchors
  (icons, `branch_icon`, `folder_separator_icon`, `status_formats` - all evidenced in shipped
  themes) MUST be `template.Markup` or their anchors render as literal text.
- Known limitations: template funcs typed `string` error on Markup args except the adapted set
  in `markupStringFuncs` (lower/upper/title/trunc/...). `mapped_branches` values keep anchors
  only when no `branch_template` is set (the sub-template receives `.Branch` as a plain string
  so `trunc` keeps working).
- `terminal.write`'s `isHyperlink` branch (OSC 8 URI region) applies shell escaping
  (`formats.EscapeSequences`) since 2026-09-05 - bash `@P` would otherwise re-interpret a URI
  backslash as a prompt escape. Never bypass it there.
- Regression net: `src/prompt/golden_test.go` renders all 125 shipped themes byte-exact; if a
  markup change alters golden output, the theme relied on the old (insecure) behavior -
  regenerate with `go test ./prompt/... -run TestGoldenThemes -update` only after inspecting
  the diff.

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
- Streaming pending-segment bookkeeping (`pendingSegments` map, `segment.Pending`,
  `streamingResults` channel) has no lock; correctness relies on ordering (verified 2026-09-05).
  A segment that finishes right around the streaming timeout used to leave the prompt (and the
  transient rendered from it) stuck on "..." for good - the producer's "nothing pending" check
  read 0 while the completion sat unread in the channel buffer, or the cleanup goroutine raced
  the tracker's deregistration and the notification was never sent. Invariants now: deregister
  an in-time segment BEFORE handing over its result, a timed-out one is deregistered only by
  the producer when it consumes the notification (so "neither registered nor queued" cannot
  happen while a render is owed), the producer exits only when nothing is registered AND the
  buffer is empty, the buffer holds one slot per segment so a send never drops, and
  `streamingResults` is never closed (a late tracker send on a closed channel panics in a
  goroutine with no recover and kills the serve daemon).
- Reproducing timing races in the daemon: `oh-my-posh debug` timings are cold-process numbers;
  the warm in-daemon segment duration is what has to straddle `streaming`. Sweep the timeout in
  a config copy against a scripted serve session (JSON line + env blob on stdin, count cycles
  whose last record still holds the placeholder) instead of trusting a single value.
