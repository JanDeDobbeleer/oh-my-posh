# PowerShell (pwsh)

## Engine events and runspaces (pwsh 7, 2026-07)

- `Register-EngineEvent -Action` (e.g. `PowerShell.OnIdle`): `-MessageData` arrives as `$null`, and
  `.GetNewClosure()` captures are lost when the scriptblock is created in a module function. Pass state through a
  `$global:` variable (`$global:_ompStreamingState` in `omp.ps1`).
- PSReadLine raises `PowerShell.OnIdle` every ~300-450ms, only while the input buffer is empty. Anything the user
  can trigger sooner (transient prompt on a fast Enter) must also drain synchronously at its call site.
- PSEvents raised from a non-engine thread (e.g. `Register-ObjectEvent DataAdded` fed by a background runspace)
  can crash the host (`InvalidPipelineStateException`). Consume records only on the engine thread.
- `Register-ObjectEvent` actions may or may not fire during a busy-wait. Never share a consume cursor between an
  event action and a synchronous waiter: give the waiter a private cursor and make the action idempotent.

## PSReadLine prompt redraws (PSReadLine 2.4.5 / 2.3.6, 2026-09)

- `InvokePrompt()` blanks `ExtraPromptLineCount + 1` rows before it calls the prompt function; a count set during
  that call only affects the next redraw. Padding the transient prompt cannot fix under-erasure.
- `ExtraPromptLineCount` must count screen rows, including wrapping. The engine computes `terminal.CursorRow` and
  prefixes pwsh prompts with `ESC]7777;<row>BEL`, which `Set-PoshExtraPromptLineCount` strips. Don't measure the
  prompt in PowerShell (`Measure-Object -Line` ignores wrapping and skips empty elements).
- Overcounting is as bad: it erases the previous output line, or at the screen top `InvokePrompt` returns before
  calling the prompt (`newY < 0`), leaving `$script:TransientPrompt` set so the next primary renders as transient.
  Zero-width sequences such as the rprompt's `ESC 7`/`ESC 8` count 0 cells.
- Match escape-prefixed strings with `-match '^\u001b...'`, not `StartsWith`: culture-aware `StartsWith` ignores
  ESC, so `"abc".StartsWith("$([char]27)a")` is `True`.
- Redraw bugs only show when a line wraps; reproduce with the e2e `harness.MultiLine` overlay.

## Exit lifecycle

- pwsh cannot exit while a `[powershell]::Create()` pipeline thread runs (foreground threads). A reader blocked in
  `ReadByte()` on a child that exits only on stdin EOF deadlocks `exit` (#7643).
- Module `OnRemove` never runs on a normal `exit`. Tear down a daemon in a `Register-EngineEvent PowerShell.Exiting`
  handler: write `quit`, close the child's stdin (the guaranteed EOF), kill after a short `WaitForExit`.

## Performance (ARM64 Windows 11, 2026-07)

- Process creation floor is ~70ms; any omp spawn costs ~100-130ms. For `omp.ps1` perf, count spawns per Enter
  first; cmdlet and module overhead is negligible.
- `&` saves only ~10-17ms over the Process API and, under a CP437/1252 console, mangles UTF-8 output unless
  `[Console]::OutputEncoding = UTF8` is set once at init.
- `Start-Sleep -Milliseconds 1` sleeps a ~15.6ms timer tick. Don't sleep-poll; signal a `ManualResetEventSlim`.

## Testing

- `New-Event -SourceIdentifier PowerShell.OnIdle` runs the real OnIdle action deterministically; its side effects can
  short-circuit later prompt calls.
- Set module-scope flags from a test: `& (Get-Module oh-my-posh-core) { $script:X = $true }`.
- Serve is gated to pwsh 6+ outside ConstrainedLanguage; pwsh 5.1 and ConstrainedLanguage keep the legacy path.
- Exit-deadlock repro: `Start-Process pwsh -File test.ps1` + `WaitForExit(timeout)`; `findstr x` with redirected
  stdio stands in for a daemon.
