# PowerShell (pwsh)

## Engine events and runspaces

- `Register-EngineEvent` (e.g. `PowerShell.OnIdle`) `-Action` handlers, verified pwsh 7 (2026-07):
  - `-MessageData` is accepted but arrives as `$null` in the action (unlike `Register-ObjectEvent`,
    where it works).
  - `.GetNewClosure()` bindings are lost when the scriptblock is created inside a module function -
    captured variables resolve to `$null` when the action fires.
  - The only reliable way to pass state into an engine-event action is a `$global:` variable
    (see `$global:_ompStreamingState` in `src/shell/scripts/omp.ps1`).
- PSReadLine generates `PowerShell.OnIdle` itself every ~300-450ms while waiting for input (only
  when the input buffer is empty) and pumps ALL queued subscriber actions via a nested pipeline.
  `InvokePrompt()` is designed for OnIdle subscribers.
- PSEvents raised from a NON-engine thread (e.g. `Register-ObjectEvent DataAdded` on a collection
  appended by a background runspace) can re-enter `RunspaceBase.Pulse()` and **crash the host**
  with InvalidPipelineStateException under rapid prompt cycles. Consume records only on the engine
  thread: sync waiter plus OnIdle drain.
- Whether `Register-ObjectEvent` actions fire during a busy-wait depends on where the wait runs
  (top-level script vs busy pipeline thread) - it is not a given either way. Never share
  a consume-cursor between an event action and a synchronous waiter - give the waiter a private
  cursor (records stay in the PSDataCollection) and make the action idempotent.
- OnIdle needs ~300ms of idle time. Anything the user can trigger sooner (transient prompt on a
  fast Enter) must ALSO drain synchronously at its call site or it pays a CLI-spawn fallback.

## PSReadLine prompt redraws

- Verified 2026-09-27, PSReadLine 2.4.5 (pwsh 7.6.5 Windows) and 2.3.6 (pwsh 7.5.2 Linux):
  `InvokePrompt()` blanks `ExtraPromptLineCount + 1` rows above its input line **before** it
  calls the prompt function. The value the prompt function sets during that call only affects the
  *next* redraw. Padding the transient prompt to the primary prompt's height cannot fix
  under-erasure (the fix suggested in starship#7737 does not carry over).
- `ExtraPromptLineCount` must count screen rows, not logical lines. `Measure-Object -Line`
  ignores wrapping, so a long path left the extra rows on screen after the transient prompt
  (#7881). The engine now computes the row the cursor ends on (`terminal.CursorRow`) and prefixes
  pwsh primary, transient and debug prompts with `ESC]7777;<row>BEL`. `Set-PoshExtraPromptLineCount`
  in `omp.ps1` strips it. Don't measure the prompt in PowerShell: a regex plus
  `LengthInBufferCells` has to guess which bytes are zero width.
- Too high a count is as bad as too low. A first attempt counted the rprompt's `ESC 7` / `ESC 8`
  (cursor save/restore) as 2 cells, so a single-line prompt with an rprompt counted 2 rows. The
  transient redraw then erased the previous output line, or at the top of the screen
  `InvokePrompt` silently returned (`newY < 0`) before calling the prompt function. That left
  `$script:TransientPrompt` set, so the next primary prompt rendered as the transient one.
- Match escape-prefixed strings with `-match '^\u001b...'`, not `StartsWith`: culture-aware
  `StartsWith` ignores ESC, so `"abc".StartsWith("$([char]27)a")` is `True` (pwsh 7.6.5 Windows
  and 7.5.2 Linux).
- `Measure-Object -Line` also skips empty array elements: `@("","a","b")` counts 2.
- A typed parameter keeps its type on reassignment: after `$p = $p -join "`n"`, a
  `[string[]]$p` is still an array. When a prompt function throws, the host falls back to `PS>`.
- A prompt with several logical lines and no wrapping clears correctly. To reproduce redraw bugs,
  make a line wrap (the e2e `harness.MultiLine` overlay does this).

## Exit lifecycle

- pwsh cannot exit while a `[powershell]::Create()` pipeline thread runs - pipeline threads are
  foreground threads. A reader runspace blocked in `ReadByte()` on a child's stdout deadlocks
  `exit` when that child exits only on stdin EOF (issue #7643).
- Module `OnRemove` only runs on `Remove-Module`, never on normal `exit` - it cannot be a
  daemon's teardown path. Use a `Register-EngineEvent PowerShell.Exiting` handler that writes
  `quit`, **closes the child's stdin** (the guaranteed EOF signal), and kills after a short
  `WaitForExit`. State reaches the action via `$global:` (see above).

## Performance (measured 2026-07-06, ARM64 Windows 11)

- Process creation floor is ~70ms (`cmd /c exit`); any omp spawn from pwsh costs ~100-130ms wall
  regardless of exe speed. When touching `omp.ps1` perf, count process spawns per Enter first -
  the spawn-per-prompt architecture dominates, script/module cmdlet overhead is negligible
  (<0.1ms; module vs plain script is a non-issue).
- `&` call operator vs the Process API is only ~10-17ms faster - CreateProcess dominates. Under a
  stock CP437/1252 console, `&` mangles UTF-8 output (U+E0B6 becomes U+03B5); dropping the Process
  API requires `[Console]::OutputEncoding = UTF8` once at init.
- Serve daemon warm render ~10ms; warm prompt cost 18-23ms vs 160-220ms for the per-prompt
  streaming cycle it replaced (runspace creation ~35ms + event churn + 15.6ms-quantized sleeps).
- Windows timer resolution: each `Start-Sleep -Milliseconds 1` is a ~15.6ms tick - sleep-polling
  loops are ~16x slower than intended. Prefer a `ManualResetEventSlim` signaled by the reader.

## Testing

- `New-Event -SourceIdentifier PowerShell.OnIdle` runs the real OnIdle action deterministically
  (engines never idle mid-script). Beware the action's side effects short-circuiting later prompt
  calls in the harness.
- Set module-scope flags from a test: `& (Get-Module oh-my-posh-core) { $script:X = $true }`.
- pwsh 5.1 / ConstrainedLanguage keep the legacy stream path - serve is gated to pwsh 6+.
- Exit-deadlock repro pattern: `Start-Process pwsh -File test.ps1` + `WaitForExit(timeout)`;
  `findstr x` with redirected stdio is a good stand-in daemon.
