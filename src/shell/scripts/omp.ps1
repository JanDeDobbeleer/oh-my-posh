# remove any existing dynamic module of OMP
if ($null -ne (Get-Module -Name "oh-my-posh-core")) {
    Remove-Module -Name "oh-my-posh-core" -Force
}

# disable all known python virtual environment prompts
$env:VIRTUAL_ENV_DISABLE_PROMPT = 1
$env:PYENV_VIRTUALENV_DISABLE_PROMPT = 1

# Helper functions which need to be defined before the module is loaded
# See https://github.com/JanDeDobbeleer/oh-my-posh/discussions/2300
function global:Get-PoshStackCount {
    $locations = Get-Location -Stack
    if ($locations) {
        return $locations.Count
    }
    return 0
}

# global enablers
$global:_ompJobCount = $false
$global:_ompFTCSMarks = $false
$global:_ompPoshGit = $false
$global:_ompAzure = $false
$global:_ompExecutable = ::OMP::
$global:_ompTransientPrompt = $false
$global:_ompStreaming = $false

New-Module -Name "oh-my-posh-core" -ScriptBlock {
    # Check `ConstrainedLanguage` mode.
    $script:ConstrainedLanguageMode = $ExecutionContext.SessionState.LanguageMode -eq "ConstrainedLanguage"

    # The persistent `oh-my-posh serve` daemon needs ProcessStartInfo.ArgumentList,
    # System.Diagnostics.Process and [powershell]::Create() runspaces, none of which
    # are usable/available under ConstrainedLanguage mode or on Windows PowerShell
    # 5.1 (.NET Framework, no ArgumentList support). Both cases keep using the
    # legacy per-prompt stream spawn.
    $script:ServeSupported = -not $script:ConstrainedLanguageMode -and $PSVersionTable.PSVersion.Major -ge 6

    # Prompt related backup.
    $script:OriginalPromptFunction = $Function:prompt
    $originalPSReadLineOptions = Get-PSReadLineOption
    $script:OriginalContinuationPrompt = $originalPSReadLineOptions.ContinuationPrompt
    $script:OriginalPromptText = $originalPSReadLineOptions.PromptText
    $script:OriginalViModeIndicator = $originalPSReadLineOptions.ViModeIndicator
    $script:OriginalViModeChangeHandler = $originalPSReadLineOptions.ViModeChangeHandler

    $script:NoExitCode = $true
    $script:ErrorCode = 0
    $script:ExecutionTime = 0
    $script:ShellName = "pwsh"
    $script:PSVersion = $PSVersionTable.PSVersion.ToString()
    $script:TransientPrompt = $false
    $script:TooltipCommand = ''
    $script:JobCount = 0
    $script:Streaming = [hashtable]::Synchronized(@{
            Process      = $null
            Prompt       = ''
            Transient    = ''
            State        = 'NEW'
            Dirty        = $false
            # Session-scoped `oh-my-posh serve` process state (PowerShell 6+ only).
            # ServeProcess/StdIn live for the whole session; CycleId increments once
            # per render request so stale records from an aborted cycle can be
            # discarded by comparing against it.
            CycleId      = 0
            ServeProcess = $null
            StdIn        = $null
            # The serve reader runspace's PSDataCollection. It lives for the
            # daemon's lifetime and only grows - records are never removed.
            Output       = $null
            # Cursor into Output, shared by the synchronous waiter in
            # Get-PoshStreamingPrompt and the async drain in the OnIdle action.
            # Both run on the engine thread and never overlap (OnIdle is only
            # raised while the runspace is idle), so sharing is race-free.
            # Records are never removed from Output.
            RecordIndex  = 0
            # Set by the reader runspace after each record lands in Output (and on
            # EOF), so the waiter can block on it instead of sleep-polling -
            # Start-Sleep quantizes to ~15.6ms Windows timer ticks, the wait
            # handle wakes sub-millisecond.
            Signal       = $null
            # Set the first time either the serve or legacy path kicks off a
            # cycle; lets the shared PowerShell.OnIdle handler below know
            # whether a streaming prompt cycle is active at all, regardless
            # of which of the two mechanisms is driving it.
            CycleStarted = $false
            # Counts daemon failures (start failure, dead pipe, response
            # timeout). Deliberately never reset on success: a flapping daemon
            # should eventually stop taxing prompts with restarts.
            FailureCount = 0
            # Drains records the reader appended to Output: async segment
            # updates and the transient refresh. Serve records carry an
            # "<id>\x1f" prefix (stale cycles are discarded); legacy stream
            # records are the bare payload. Engine-thread only - shared by the
            # OnIdle action (which can't call module functions, hence a
            # scriptblock on the state it already holds) and the prompt
            # function's transient branch.
            Drain        = {
                param($s)

                $output = $s.Output
                if ($null -eq $output) {
                    return
                }

                while ($s.RecordIndex -lt $output.Count) {
                    $record = $output[$s.RecordIndex]
                    $s.RecordIndex++

                    if (-not $record) {
                        continue
                    }

                    $sep = $record.IndexOf([char]0x1F)
                    if ($sep -ge 0) {
                        if ($record.Substring(0, $sep) -ne [string]$s.CycleId) {
                            # Stale record from an aborted/previous cycle - discard.
                            continue
                        }
                        $payload = $record.Substring($sep + 1)
                    }
                    else {
                        $payload = $record
                    }

                    # A payload prefixed with U+001E carries the transient prompt:
                    # cache it for the Enter/Ctrl+C key handlers, never repaint.
                    if ($payload -and $payload[0] -eq [char]0x1E) {
                        $s.Transient = $payload.Substring(1)
                        continue
                    }

                    # An unchanged prompt needs no repaint.
                    if ($payload -ceq $s.Prompt) {
                        continue
                    }

                    $s.Prompt = $payload
                    $s.Dirty = $true
                }
            }
        })
    # Engine-event actions can't receive state via -MessageData (arrives as $null) and lose
    # closure bindings when created inside a module function, so expose the streaming state
    # globally for the OnIdle action to pick up.
    $global:_ompStreamingState = $script:Streaming
    $script:StreamingOnIdleJob = $null
    $script:StreamingExitingJob = $null

    $env:POWERLINE_COMMAND = "oh-my-posh"
    $env:POSH_SHELL = "pwsh"
    $env:POSH_SHELL_VERSION = $script:PSVersion
    $env:CONDA_PROMPT_MODIFIER = $false

    function Invoke-Utf8Posh {
        param([string[]]$Arguments = @())

        if ($script:ConstrainedLanguageMode) {
            $output = Invoke-Expression "& `$global:_ompExecutable `$Arguments 2>&1"
            $output -join "`n"
            return
        }

        $Process = New-Object System.Diagnostics.Process
        $StartInfo = $Process.StartInfo
        $StartInfo.FileName = $global:_ompExecutable
        if ($StartInfo.ArgumentList.Add) {
            # ArgumentList is supported in PowerShell 6.1 and later (built on .NET Core 2.1+)
            # ref-1: https://docs.microsoft.com/en-us/dotnet/api/system.diagnostics.processstartinfo.argumentlist?view=net-6.0
            # ref-2: https://docs.microsoft.com/en-us/powershell/scripting/whats-new/differences-from-windows-powershell?view=powershell-7.2#net-framework-vs-net-core
            $Arguments | ForEach-Object -Process { $StartInfo.ArgumentList.Add($_) }
        }
        else {
            # escape arguments manually in lower versions, refer to https://docs.microsoft.com/en-us/previous-versions/17w5ykft(v=vs.85)
            $escapedArgs = $Arguments | ForEach-Object {
                # escape N consecutive backslash(es), which are followed by a double quote, to 2N consecutive ones
                $s = $_ -replace '(\\+)"', '$1$1"'
                # escape N consecutive backslash(es), which are at the end of the string, to 2N consecutive ones
                $s = $s -replace '(\\+)$', '$1$1'
                # escape double quotes
                $s = $s -replace '"', '\"'
                # quote the argument
                "`"$s`""
            }
            $StartInfo.Arguments = $escapedArgs -join ' '
        }

        $StartInfo.StandardErrorEncoding = $StartInfo.StandardOutputEncoding = [System.Text.Encoding]::UTF8
        $StartInfo.RedirectStandardError = $StartInfo.RedirectStandardInput = $StartInfo.RedirectStandardOutput = $true
        $StartInfo.UseShellExecute = $false
        if ($PWD.Provider.Name -eq 'FileSystem') {
            # make sure we're in a valid directory
            # if not, go back HOME
            if (-not (Test-Path -LiteralPath $PWD)) {
                Write-Host "Unable to find the current directory, falling back to $HOME" -ForegroundColor Red
                Set-Location $HOME
            }
            $StartInfo.WorkingDirectory = $PWD.ProviderPath
        }

        $StartInfo.CreateNoWindow = $true
        [void]$Process.Start()

        # Remove deadlock potential on Windows.
        $stdoutTask = $Process.StandardOutput.ReadToEndAsync()
        $stderrTask = $Process.StandardError.ReadToEndAsync()

        $Process.WaitForExit()
        $stderr = $stderrTask.Result.Trim()
        if ($stderr) {
            $Host.UI.WriteErrorLine($stderr)
        }

        $stdoutTask.Result
    }

    function Get-NonFSWD {
        # We only need to return a non-filesystem working directory.
        if ($PWD.Provider.Name -ne 'FileSystem') {
            return $PWD.ToString()
        }
    }

    function Get-TerminalWidth {
        $terminalWidth = $Host.UI.RawUI.WindowSize.Width
        # Set a sane default when the value can't be retrieved.
        if (-not $terminalWidth) {
            return 0
        }
        $terminalWidth
    }

    function Set-TransientPrompt {
        $previousOutputEncoding = [Console]::OutputEncoding
        try {
            $script:TransientPrompt = $true
            [Console]::OutputEncoding = [Text.Encoding]::UTF8
            [Microsoft.PowerShell.PSConsoleReadLine]::InvokePrompt()
        }
        catch [System.ArgumentOutOfRangeException] {
        }
        finally {
            [Console]::OutputEncoding = $previousOutputEncoding
        }
    }

    function Set-PoshPromptType {
        if ($script:TransientPrompt -eq $true) {
            $script:PromptType = "transient"
            $script:TransientPrompt = $false
            return
        }

        # for details about the trick to detect a debugging context, see these comments:
        # 1) https://github.com/JanDeDobbeleer/oh-my-posh/issues/2483#issuecomment-1175761456
        # 2) https://github.com/JanDeDobbeleer/oh-my-posh/issues/2502#issuecomment-1179968052
        # 3) https://github.com/JanDeDobbeleer/oh-my-posh/issues/5153
        if ($Host.Runspace.Debugger.InBreakpoint) {
            $script:PromptType = "debug"
            return
        }

        $script:PromptType = "primary"

        if ($global:_ompJobCount) {
            $script:JobCount = (Get-Job -State Running).Count
        }

        if ($global:_ompAzure) {
            try {
                $env:POSH_AZURE_SUBSCRIPTION = Get-AzContext | ConvertTo-Json
            }
            catch {
            }
        }

        if ($global:_ompPoshGit) {
            try {
                $global:GitStatus = Get-GitStatus
                $env:POSH_GIT_STATUS = $global:GitStatus | ConvertTo-Json
            }
            catch {
            }
        }
    }

    function Update-PoshErrorCode {
        $lastHistory = Get-History -ErrorAction Ignore -Count 1

        # error code should be updated only when a non-empty command is run
        if (($null -eq $lastHistory) -or ($script:LastHistoryId -eq $lastHistory.Id)) {
            $script:ExecutionTime = 0
            $script:NoExitCode = $true
            return
        }

        $script:NoExitCode = $false
        $script:LastHistoryId = $lastHistory.Id
        $script:ExecutionTime = ($lastHistory.EndExecutionTime - $lastHistory.StartExecutionTime).TotalMilliseconds
        if ($script:OriginalLastExecutionStatus) {
            $script:ErrorCode = 0
            return
        }

        $invocationInfo = try {
            # retrieve info of the most recent error
            $global:Error | Where-Object { $_.GetType().Name -eq 'ErrorRecord' } | Select-Object -First 1 -ExpandProperty InvocationInfo
        }
        catch {
            $null
        }

        # Check if the error occurred in the current command scope
        if ($null -ne $invocationInfo -and
            $invocationInfo.HistoryId -eq $lastHistory.Id) {
            $script:ErrorCode = 1
            return
        }

        if ($script:OriginalLastExitCode -is [int] -and $script:OriginalLastExitCode -ne 0) {
            # native app exit code
            $script:ErrorCode = $script:OriginalLastExitCode
            return
        }
    }

    function Get-PoshPrompt {
        param(
            [string]$Type,
            [string[]]$Arguments
        )
        $nonFSWD = Get-NonFSWD
        $stackCount = Get-PoshStackCount
        $terminalWidth = Get-TerminalWidth
        Invoke-Utf8Posh @(
            "print", $Type
            "--save-cache"
            "--shell=$script:ShellName"
            "--shell-version=$script:PSVersion"
            "--status=$script:ErrorCode"
            "--no-status=$script:NoExitCode"
            "--execution-time=$script:ExecutionTime"
            "--pswd=$nonFSWD"
            "--stack-count=$stackCount"
            "--terminal-width=$terminalWidth"
            "--job-count=$script:JobCount"
            if ($Arguments) {
                $Arguments
            }
        )
    }

    function Register-PoshStreamingOnIdle {
        if ($null -ne $script:StreamingOnIdleJob) {
            return
        }

        # PSReadLine anchors the prompt position when ReadLine starts, and PowerShell.OnIdle
        # can only fire while ReadLine is waiting for input, i.e. after that anchor exists.
        # That makes OnIdle the earliest safe point to allow redraws (State = 'RUNNING') and
        # to flush updates that arrived before the anchor existed. Calling InvokePrompt()
        # any earlier redraws at the previous prompt's coordinates.
        #
        # OnIdle is also the ONLY async consumer of streamed records. It is an
        # engine-generated event on the pipeline thread: consuming records here
        # instead of in a DataAdded subscription means no PSEvent is ever raised
        # from a background thread. Cross-thread event delivery can re-enter the
        # engine's pulse pipeline and crash the host with
        # InvalidPipelineStateException ("Cannot invoke pipeline because it has
        # already been invoked") under rapid prompt cycles.
        $script:StreamingOnIdleJob = Register-EngineEvent -SourceIdentifier PowerShell.OnIdle -Action {
            $s = $global:_ompStreamingState

            # No streaming prompt cycle has ever been started (neither serve nor legacy).
            if (-not $s.CycleStarted) {
                return
            }

            if ($s.State -eq 'NEW') {
                $s.State = 'RUNNING'
            }

            # Drain records the reader appended while idle: async segment
            # updates and the transient refresh. This runs on the same thread
            # as the synchronous waiter, so sharing the cursor is race-free.
            & $s.Drain $s

            if (-not $s.Dirty) {
                return
            }

            $s.Dirty = $false

            $previousOutputEncoding = [Console]::OutputEncoding

            try {
                [Console]::OutputEncoding = [Text.Encoding]::UTF8
                [Microsoft.PowerShell.PSConsoleReadLine]::InvokePrompt()
            }
            catch {}
            finally {
                [Console]::OutputEncoding = $previousOutputEncoding
            }
        }
    }

    function Stop-StreamingProcess {
        if (-not $global:_ompStreaming) {
            return
        }

        # Kill the streaming process
        if ($null -ne $script:Streaming.Process -and -not $script:Streaming.Process.HasExited) {
            try {
                $script:Streaming.Process.Kill()
            }
            catch {
            }
        }

        $script:Streaming.Process = $null
        $script:Streaming.State = 'NEW'
        $script:Streaming.Dirty = $false
    }

    function Stop-ActiveRenderCycle {
        # Serve mode: the daemon persists across cycles, only the in-flight
        # render needs to be interrupted - write abort instead of killing
        # anything. Guard on the process still being alive.
        if ($null -ne $script:Streaming.ServeProcess -and -not $script:Streaming.ServeProcess.HasExited) {
            try {
                $script:Streaming.StdIn.WriteLine('{"command":"abort"}')
                $script:Streaming.StdIn.Flush()
            }
            catch {
            }

            $script:Streaming.State = 'NEW'
            $script:Streaming.Dirty = $false
            return
        }

        # Legacy per-prompt process: nothing to abort, kill it outright.
        Stop-StreamingProcess
    }

    # Byte-level reader for NUL-delimited prompt records, shared by the serve
    # daemon (which passes a wake signal) and the legacy per-prompt stream
    # (which passes $null). Runs in its own runspace for the lifetime of the
    # stream it reads.
    $script:StreamingReaderScript = {
        param($stream, $signal)
        while ($true) {
            $bytes = [System.Collections.Generic.List[byte]]::new()
            while (($b = $stream.ReadByte()) -notin -1, 0) {
                $bytes.Add($b)
            }

            if ($bytes.Count -gt 0) {
                Write-Output ([Text.Encoding]::UTF8.GetString($bytes.ToArray()))
            }

            # Wake the waiter - also on EOF, so a dying daemon triggers the
            # fallback path immediately instead of after the timeout.
            if ($signal) {
                $signal.Set()
            }

            if ($b -eq -1) {
                return
            }
        }
    }

    function Start-PoshServe {
        $Process = New-Object System.Diagnostics.Process
        $StartInfo = $Process.StartInfo
        $StartInfo.FileName = $global:_ompExecutable

        # ArgumentList is supported in PowerShell 6.1+; $script:ServeSupported already
        # requires major version 6, but guard defensively like Invoke-Utf8Posh does.
        if ($StartInfo.ArgumentList.Add) {
            $StartInfo.ArgumentList.Add("serve")
            $StartInfo.ArgumentList.Add("--shell=$script:ShellName")
        }
        else {
            $StartInfo.Arguments = "serve --shell=$script:ShellName"
        }

        # IMPORTANT: BOM-less UTF-8 for stdin. [System.Text.Encoding]::UTF8
        # emits a BOM preamble on the writer's first write, which would
        # corrupt the first JSON request line and make the daemon silently
        # drop the first render of every fresh process.
        $StartInfo.StandardInputEncoding = [System.Text.UTF8Encoding]::new($false)
        $StartInfo.StandardOutputEncoding = [System.Text.UTF8Encoding]::new($false)
        $StartInfo.RedirectStandardInput = $true
        $StartInfo.RedirectStandardOutput = $true
        # stdout carries ONLY protocol records; redirect stderr too so a Go
        # panic in the daemon can never spew into the user's terminal - it's
        # simply discarded (not read) since we never attach a consumer to it.
        $StartInfo.RedirectStandardError = $true
        $StartInfo.UseShellExecute = $false
        $StartInfo.CreateNoWindow = $true
        if ($PWD.Provider.Name -eq 'FileSystem') {
            $StartInfo.WorkingDirectory = $PWD.ProviderPath
        }

        try {
            [void]$Process.Start()
        }
        catch {
            return $false
        }

        # Drain stderr fire-and-forget: an undrained redirected pipe can fill
        # up and block the daemon mid-write (e.g. an unrecovered panic's stack
        # trace). The content is deliberately discarded.
        $null = $Process.StandardError.ReadToEndAsync()

        # Read the persistent stdout stream asynchronously for the lifetime of the session.
        $output = New-Object 'System.Management.Automation.PSDataCollection[PSObject]'
        $inputData = New-Object 'System.Management.Automation.PSDataCollection[PSObject]'
        $inputData.Complete()
        $signal = [System.Threading.ManualResetEventSlim]::new($false)
        $ps = [powershell]::Create().AddScript($script:StreamingReaderScript).AddArgument($Process.StandardOutput.BaseStream).AddArgument($signal)

        # There is deliberately NO DataAdded subscription on the collection:
        # the reader appends from a background thread, and a PSEvent raised
        # from a non-engine thread can re-enter the engine's pulse pipeline
        # and crash the host (InvalidPipelineStateException) under rapid
        # prompt cycles. Records are consumed exclusively on the engine
        # thread: synchronously by Get-PoshStreamingPrompt's waiter, and
        # asynchronously by the OnIdle action's drain.
        $ps.BeginInvoke($inputData, $output) | Out-Null

        $script:Streaming.ServeProcess = $Process
        $script:Streaming.StdIn = $Process.StandardInput
        # Fresh daemon, fresh collection: nothing consumed yet. The previous
        # daemon's signal (if any) is intentionally not disposed - its reader
        # may still hold a reference; the GC reclaims it.
        $script:Streaming.Output = $output
        $script:Streaming.RecordIndex = 0
        $script:Streaming.Signal = $signal

        return $true
    }

    function ConvertTo-PoshServeJsonString($Value) {
        if ($null -eq $Value) {
            return '""'
        }

        # Minimal, fast escaping for the flat string values we send: backslash
        # and double-quote first (order matters), then control characters.
        $escaped = $Value.Replace('\', '\\').Replace('"', '\"')
        $escaped = $escaped -replace "`r", '\r' -replace "`n", '\n' -replace "`t", '\t'
        return '"' + $escaped + '"'
    }

    function Get-PoshFSWD {
        # Serve needs an actual filesystem path for the daemon to os.Chdir into;
        # a non-filesystem provider (e.g. a registry drive) has no such path -
        # let the daemon keep its previous/last-good working directory.
        if ($PWD.Provider.Name -eq 'FileSystem') {
            return $PWD.ProviderPath
        }
        return ''
    }

    function Get-PoshServeEnvOverlay {
        # v1 env overlay: PATH, every POSH_* variable, VIRTUAL_ENV and
        # CONDA_PROMPT_MODIFIER. Deliberately not derived from config
        # templates yet (see implementation plan TODO).
        $overlay = [ordered]@{}
        $overlay['PATH'] = $env:PATH

        Get-ChildItem env:POSH_* -ErrorAction Ignore | ForEach-Object {
            $overlay[$_.Name] = $_.Value
        }

        if (Test-Path env:VIRTUAL_ENV) {
            $overlay['VIRTUAL_ENV'] = $env:VIRTUAL_ENV
        }

        if (Test-Path env:CONDA_PROMPT_MODIFIER) {
            $overlay['CONDA_PROMPT_MODIFIER'] = $env:CONDA_PROMPT_MODIFIER
        }

        return $overlay
    }

    function Suspend-PoshServeOnFailure {
        $script:Streaming.FailureCount++
        if ($script:Streaming.FailureCount -ge 3) {
            # Degrade to the per-prompt stream path for the rest of the
            # session - a repeatedly failing daemon must not add a restart
            # plus a response timeout to every single prompt.
            $script:ServeSupported = $false
        }
    }

    function Get-PoshStreamingPrompt {
        if (-not $script:ServeSupported) {
            return Get-PoshStreamingPromptLegacy
        }

        Register-PoshStreamingOnIdle

        # The reader's record collection only grows - both consumers key off
        # add-time indices, so in-place trimming would corrupt their cursors.
        # Recycle the daemon once the collection gets large: one slower prompt
        # every few thousand beats unbounded growth in long-lived sessions.
        if ($null -ne $script:Streaming.Output -and $script:Streaming.Output.Count -ge 4096) {
            try {
                $script:Streaming.StdIn.WriteLine('{"command":"quit"}')
                $script:Streaming.StdIn.Flush()
            }
            catch {
            }
            $script:Streaming.ServeProcess = $null
        }

        if ($null -eq $script:Streaming.ServeProcess -or $script:Streaming.ServeProcess.HasExited) {
            if (-not (Start-PoshServe)) {
                # Could not start the daemon at all - fall back for this cycle.
                Suspend-PoshServeOnFailure
                return Get-PoshStreamingPromptLegacy
            }
        }

        $script:Streaming.CycleId++
        $script:Streaming.Transient = ''
        $script:Streaming.CycleStarted = $true

        $envOverlay = Get-PoshServeEnvOverlay
        $envJson = ($envOverlay.Keys | ForEach-Object {
                '"' + $_ + '":' + (ConvertTo-PoshServeJsonString $envOverlay[$_])
            }) -join ','

        $json = '{' +
        '"command":"render"' +
        ',"id":' + $script:Streaming.CycleId +
        ',"shell":' + (ConvertTo-PoshServeJsonString $script:ShellName) +
        ',"shell-version":' + (ConvertTo-PoshServeJsonString $script:PSVersion) +
        ',"status":' + [int]$script:ErrorCode +
        ',"no-status":' + $(if ($script:NoExitCode) { 'true' } else { 'false' }) +
        ',"execution-time":' + $script:ExecutionTime +
        ',"pwd":' + (ConvertTo-PoshServeJsonString (Get-PoshFSWD)) +
        ',"pswd":' + (ConvertTo-PoshServeJsonString (Get-NonFSWD)) +
        ',"stack-count":' + (Get-PoshStackCount) +
        ',"terminal-width":' + (Get-TerminalWidth) +
        ',"job-count":' + $script:JobCount +
        ',"cleared":false' +
        ',"env":{' + $envJson + '}' +
        '}'

        try {
            $script:Streaming.StdIn.WriteLine($json)
            $script:Streaming.StdIn.Flush()
        }
        catch {
            # The daemon died between the health check above and this write - restart once.
            Suspend-PoshServeOnFailure
            # Mirror the timeout path: kill before dropping the reference, so a
            # process with a broken stdin but a live body can never be leaked.
            try {
                $script:Streaming.ServeProcess.Kill()
            }
            catch {
            }
            $script:Streaming.ServeProcess = $null
            if (-not (Start-PoshServe)) {
                return Get-PoshStreamingPromptLegacy
            }

            try {
                $script:Streaming.StdIn.WriteLine($json)
                $script:Streaming.StdIn.Flush()
            }
            catch {
                # The legacy path repoints $script:Streaming.Output at its own
                # collection - a live daemon must not linger with an orphaned one.
                try {
                    $script:Streaming.ServeProcess.Kill()
                }
                catch {
                }
                $script:Streaming.ServeProcess = $null
                return Get-PoshStreamingPromptLegacy
            }
        }

        # Wait for the first primary record of THIS cycle by scanning the
        # reader's PSDataCollection with the waiter's PRIVATE cursor. The
        # DataAdded action cannot be relied on here (whether it fires during
        # this loop depends on the calling context) and must not be raced
        # against either - records stay in the collection, so this scan works
        # regardless of whether the action also processed them, and the action
        # dedupes on unchanged content. Between scans, block on the reader's
        # signal (sub-millisecond wake) rather than Start-Sleep (~15.6ms timer
        # tick). The bounded Wait keeps the Stopwatch timeout authoritative,
        # and re-scanning after every wake makes lost wakeups impossible:
        # a record landing after a scan leaves the signal set, so the next
        # Wait returns immediately.
        $s = $script:Streaming
        $output = $s.Output
        $signal = $s.Signal
        $firstPrompt = $null
        $stopwatch = [System.Diagnostics.Stopwatch]::StartNew()

        while ($null -eq $firstPrompt -and $stopwatch.ElapsedMilliseconds -lt 2000) {
            while ($s.RecordIndex -lt $output.Count) {
                $record = $output[$s.RecordIndex]
                $s.RecordIndex++

                if (-not $record) {
                    continue
                }

                $sep = $record.IndexOf([char]0x1F)
                if ($sep -lt 0) {
                    continue
                }

                $id = $record.Substring(0, $sep)
                if ($id -ne [string]$s.CycleId) {
                    # Stale record from an aborted/previous cycle - discard.
                    continue
                }

                $payload = $record.Substring($sep + 1)

                if ($payload -and $payload[0] -eq [char]0x1E) {
                    $s.Transient = $payload.Substring(1)
                    continue
                }

                # Keep scanning instead of stopping at the primary: records
                # that already arrived in the same burst (typically the
                # transient) are consumed for free, without waiting. Later
                # records are drained by the OnIdle action.
                $s.Prompt = $payload
                $firstPrompt = $payload
            }

            if ($null -eq $firstPrompt) {
                [void]$signal.Wait(100)
                $signal.Reset()
            }
        }

        if ($null -eq $firstPrompt) {
            # Daemon stopped responding - kill it and fall back for this cycle.
            Suspend-PoshServeOnFailure
            try {
                $s.ServeProcess.Kill()
            }
            catch {
            }
            $s.ServeProcess = $null
            return Get-PoshStreamingPromptLegacy
        }

        return $firstPrompt
    }

    function Get-PoshStreamingPromptLegacy {
        Register-PoshStreamingOnIdle

        # Start streaming process (State stays 'NEW' until the first OnIdle event confirms
        # PSReadLine has rendered the initial prompt)
        $script:Streaming.Process = New-Object System.Diagnostics.Process
        $StartInfo = $script:Streaming.Process.StartInfo
        $StartInfo.FileName = $global:_ompExecutable

        # The transient prompt for this cycle streams in alongside the primary
        # prompt updates, invalidate the previous cycle's version.
        $script:Streaming.Transient = ''
        $script:Streaming.CycleStarted = $true

        # Build arguments array
        $Arguments = @(
            "stream"
            "--save-cache"
            "--shell=$script:ShellName"
            "--shell-version=$script:PSVersion"
            "--status=$script:ErrorCode"
            "--no-status=$script:NoExitCode"
            "--execution-time=$script:ExecutionTime"
            "--pswd=$(Get-NonFSWD)"
            "--stack-count=$(Get-PoshStackCount)"
            "--terminal-width=$(Get-TerminalWidth)"
            "--job-count=$script:JobCount"
        )

        # Use ArgumentList if available (PowerShell 6.1+), otherwise escape manually
        if ($StartInfo.ArgumentList.Add) {
            $Arguments | ForEach-Object -Process { $StartInfo.ArgumentList.Add($_) }
        }
        else {
            # escape arguments manually in lower versions, refer to https://docs.microsoft.com/en-us/previous-versions/17w5ykft(v=vs.85)
            $escapedArgs = $Arguments | ForEach-Object {
                # escape N consecutive backslash(es), which are followed by a double quote, to 2N consecutive ones
                $s = $_ -replace '(\\+)"', '$1$1"'
                # escape N consecutive backslash(es), which are at the end of the string, to 2N consecutive ones
                $s = $s -replace '(\\+)$', '$1$1'
                # escape double quotes
                $s = $s -replace '"', '\"'
                # quote the argument
                "`"$s`""
            }
            $StartInfo.Arguments = $escapedArgs -join ' '
        }

        $StartInfo.StandardOutputEncoding = [System.Text.Encoding]::UTF8
        $StartInfo.RedirectStandardOutput = $true
        $StartInfo.UseShellExecute = $false
        $StartInfo.CreateNoWindow = $true
        if ($PWD.Provider.Name -eq 'FileSystem') {
            $StartInfo.WorkingDirectory = $PWD.ProviderPath
        }

        [void]$script:Streaming.Process.Start()

        # Read output asynchronously
        $output = New-Object 'System.Management.Automation.PSDataCollection[PSObject]'
        $inputData = New-Object 'System.Management.Automation.PSDataCollection[PSObject]'
        $inputData.Complete()
        $ps = [powershell]::Create().AddScript($script:StreamingReaderScript).AddArgument($script:Streaming.Process.StandardOutput.BaseStream).AddArgument($null)

        # No DataAdded subscription here either - see Start-PoshServe. Async
        # updates (unprefixed records) are drained by the OnIdle action; the
        # initial prompt is consumed synchronously below.
        $ps.BeginInvoke($inputData, $output) | Out-Null

        while ($output.Count -eq 0) {
            Start-Sleep -Milliseconds 1
        }

        $script:Streaming.Prompt = $output[0]

        # Hand the collection to the OnIdle drain, index 0 already consumed.
        $script:Streaming.Output = $output
        $script:Streaming.RecordIndex = 1

        return $script:Streaming.Prompt
    }

    $promptFunction = {
        # store the original last command execution status
        if ($global:NVS_ORIGINAL_LASTEXECUTIONSTATUS -is [bool]) {
            # make it compatible with NVS auto-switching, if enabled
            $script:OriginalLastExecutionStatus = $global:NVS_ORIGINAL_LASTEXECUTIONSTATUS
        }
        else {
            $script:OriginalLastExecutionStatus = $?
        }

        # store the original last exit code
        $script:OriginalLastExitCode = $global:LASTEXITCODE

        # Only return the cached prompt when this is a streaming redraw, that is an
        # InvokePrompt() call during an active streaming cycle (RUNNING state) which
        # isn't rendering a transient prompt.
        if ($script:PromptType -ne 'transient' -and $script:Streaming.State -ne 'NEW') {
            # Update ExtraPromptLineCount for PSReadLine to properly clear previous prompt
            Set-PSReadLineOption -ExtraPromptLineCount (($script:Streaming.Prompt | Measure-Object -Line).Lines - 1)
            return $script:Streaming.Prompt
        }

        # Stop any previous render cycle (abort the serve daemon's in-flight
        # cycle, or kill the legacy per-prompt process) and reset state.
        Stop-ActiveRenderCycle

        # Reset tooltip command.
        $script:TooltipCommand = ''

        Set-PoshPromptType

        if ($script:PromptType -ne 'transient') {
            Update-PoshErrorCode
        }

        Set-PoshContext $script:ErrorCode

        # set the cursor positions, they are zero based so align with other platforms
        $env:POSH_CURSOR_LINE = $Host.UI.RawUI.CursorPosition.Y + 1
        $env:POSH_CURSOR_COLUMN = $Host.UI.RawUI.CursorPosition.X + 1

        # Use streaming prompt if enabled, otherwise use regular prompt
        if ($global:_ompStreaming -and $script:PromptType -eq 'primary') {
            $output = Get-PoshStreamingPrompt
        }
        elseif ($script:PromptType -eq 'transient') {
            if (-not $script:Streaming.Transient) {
                # The engine only raises PowerShell.OnIdle after ~300ms of
                # idle - an Enter that lands sooner would miss a transient
                # that is already sitting in the record collection and pay a
                # full CLI call instead. Drain here, on the same engine
                # thread, then discard any repaint the drain flagged: the
                # primary prompt is being replaced by the transient anyway.
                & $script:Streaming.Drain $script:Streaming
                $script:Streaming.Dirty = $false
            }

            if ($script:Streaming.Transient) {
                # rendered ahead of time by the streaming process, saves a CLI call on Enter
                $output = $script:Streaming.Transient
            }
            else {
                $output = Get-PoshPrompt $script:PromptType
            }
        }
        else {
            $output = Get-PoshPrompt $script:PromptType
        }

        # make sure PSReadLine knows if we have a multiline prompt
        Set-PSReadLineOption -ExtraPromptLineCount (($output | Measure-Object -Line).Lines - 1)

        # The output can be multi-line, joining them ensures proper rendering.
        $output = $output -join "`n"

        if ($script:PromptType -eq 'transient') {
            # Workaround to prevent a command from eating the tail of a transient prompt, when we're at the end of the line.
            $command = ''
            [Microsoft.PowerShell.PSConsoleReadLine]::GetBufferState([ref]$command, [ref]$null)
            if ($command) {
                $output += "  `b`b"
            }
        }

        $output

        # remove any posh-git status
        $env:POSH_GIT_STATUS = $null

        # restore the original last exit code
        $global:LASTEXITCODE = $script:OriginalLastExitCode
    }

    $Function:prompt = $promptFunction

    # set secondary prompt
    Set-PSReadLineOption -ContinuationPrompt ((Invoke-Utf8Posh @("print", "secondary", "--shell=$script:ShellName")) -join "`n")

    ### Exported Functions ###

    function Set-PoshContext([bool]$originalStatus) {
    }

    function Enable-PoshStreaming {
        $global:_ompStreaming = $true

        if (-not $script:ServeSupported) {
            return
        }

        # A normal `exit` never runs the module's OnRemove handler, so nothing
        # would tell the serve daemon to quit - and it only exits on stdin EOF,
        # which requires this process to be gone. But pwsh's shutdown in turn
        # waits for the reader runspace's pipeline thread, which is blocked on
        # the daemon's stdout: a circular wait that hangs the terminal on exit.
        # Break the cycle on PowerShell.Exiting: ask the daemon to quit (so it
        # flushes its caches) and close its stdin - the EOF signal that works
        # even if the quit line is lost - then kill it if it lingers. Its
        # stdout then EOFs, the reader returns, and shutdown proceeds.
        #
        # Engine-event actions receive $null MessageData and lose module-scope
        # closures, so state comes from $global:_ompStreamingState, like the
        # OnIdle action.
        if ($null -eq $script:StreamingExitingJob) {
            $script:StreamingExitingJob = Register-EngineEvent -SourceIdentifier PowerShell.Exiting -Action {
                $s = $global:_ompStreamingState

                if ($null -eq $s) {
                    return
                }

                if ($null -ne $s.ServeProcess -and -not $s.ServeProcess.HasExited) {
                    try {
                        $s.StdIn.WriteLine('{"command":"quit"}')
                        $s.StdIn.Flush()
                        $s.StdIn.Close()
                    }
                    catch {
                    }

                    if (-not $s.ServeProcess.WaitForExit(500)) {
                        try {
                            $s.ServeProcess.Kill()
                        }
                        catch {
                        }
                    }
                }

                # A lingering legacy per-prompt stream process exits by itself
                # after its render, but don't let it outlive the session either.
                if ($null -ne $s.Process -and -not $s.Process.HasExited) {
                    try {
                        $s.Process.Kill()
                    }
                    catch {
                    }
                }
            }
        }

        # Start the daemon during shell init rather than at the first prompt:
        # Process.Start() returns quickly and the spawn + engine warmup then
        # overlaps with the rest of the profile instead of delaying the first
        # prompt. Failure is fine - the first prompt retries and can still
        # fall back to the legacy per-prompt stream.
        [void](Start-PoshServe)
    }

    function Enable-PoshTooltips {
        if ($script:ConstrainedLanguageMode) {
            return
        }

        Set-PSReadLineKeyHandler -Key Spacebar -BriefDescription 'OhMyPoshSpaceKeyHandler' -ScriptBlock {
            param([ConsoleKeyInfo]$key)
            [Microsoft.PowerShell.PSConsoleReadLine]::SelfInsert($key)
            try {
                $command = ''
                [Microsoft.PowerShell.PSConsoleReadLine]::GetBufferState([ref]$command, [ref]$null)
                # Get the first word of command line as tip.
                $command = $command.TrimStart().Split(' ', 2) | Select-Object -First 1

                # Ignore an empty/repeated tooltip command.
                if (!$command -or ($command -eq $script:TooltipCommand)) {
                    return
                }

                $script:TooltipCommand = $command

                $output = (Get-PoshPrompt "tooltip" @(
                        "--column=$($Host.UI.RawUI.CursorPosition.X)"
                        "--command=$command"
                    )) -join ''
                if (!$output) {
                    return
                }

                Write-Host $output -NoNewline

                # Workaround to prevent the text after cursor from disappearing when the tooltip is printed.
                [Microsoft.PowerShell.PSConsoleReadLine]::Insert(' ')
                [Microsoft.PowerShell.PSConsoleReadLine]::Undo()
            }
            finally {
            }
        }

        Set-PSReadLineKeyHandler -Key Backspace -BriefDescription 'OhMyPoshBackspaceKeyHandler' -ScriptBlock {
            [Microsoft.PowerShell.PSConsoleReadLine]::BackwardDeleteChar()
            if (!$script:TooltipCommand) { return }

            $command = ''
            [Microsoft.PowerShell.PSConsoleReadLine]::GetBufferState([ref]$command, [ref]$null)
            $command = $command.TrimStart().Split(' ', 2) | Select-Object -First 1

            if ($command -eq $script:TooltipCommand) { return }

            $script:TooltipCommand = $command

            $output = (Get-PoshPrompt "tooltip" @(
                    "--column=$($Host.UI.RawUI.CursorPosition.X)"
                    "--command=$command"
                )) -join ''
            if (!$output) {
                $previousOutputEncoding = [Console]::OutputEncoding
                try {
                    [Console]::OutputEncoding = [Text.Encoding]::UTF8
                }
                catch [System.ArgumentOutOfRangeException] {
                }
                finally {
                    [Console]::OutputEncoding = $previousOutputEncoding
                }
                [Microsoft.PowerShell.PSConsoleReadLine]::InvokePrompt()
                return
            }

            Write-Host $output -NoNewline

            # Workaround to prevent the text after cursor from disappearing when the tooltip is printed.
            [Microsoft.PowerShell.PSConsoleReadLine]::Insert(' ')
            [Microsoft.PowerShell.PSConsoleReadLine]::Undo()
        }
    }

    function Enable-KeyHandlers {
        if ($script:ConstrainedLanguageMode) {
            return
        }

        # Helper function to create Enter key handler script block
        function New-EnterKeyHandler {
            param(
                [scriptblock]$AcceptLineFunction,
                [hashtable]$Streaming
            )
            return {
                try {
                    $Streaming.State = 'NEW'
                    $parseErrors = $null
                    [Microsoft.PowerShell.PSConsoleReadLine]::GetBufferState([ref]$null, [ref]$null, [ref]$parseErrors, [ref]$null)
                    $executingCommand = $parseErrors.Count -eq 0
                    if ($global:_ompTransientPrompt -and $executingCommand) {
                        Set-TransientPrompt
                    }
                }
                finally {
                    & $AcceptLineFunction
                    if ($global:_ompFTCSMarks -and $executingCommand) {
                        # Write FTCS_COMMAND_EXECUTED after accepting the input - it should still happen before execution
                        Write-Host "$([char]27)]133;C$([char]7)" -NoNewline
                    }
                }
            }.GetNewClosure()
        }

        # Helper function to create Ctrl+C key handler script block
        function New-CtrlCKeyHandler {
            param(
                [scriptblock]$CancelFunction,
                [hashtable]$Streaming
            )
            return {
                try {
                    $Streaming.State = 'NEW'
                    $start = $null
                    [Microsoft.PowerShell.PSConsoleReadLine]::GetSelectionState([ref]$start, [ref]$null)
                    # only render a transient prompt when no text is selected
                    if ($global:_ompTransientPrompt -and $start -eq -1) {
                        Set-TransientPrompt
                    }
                }
                finally {
                    & $CancelFunction
                }
            }.GetNewClosure()
        }

        # Register Enter key handlers
        Set-PSReadLineKeyHandler -Key Enter -BriefDescription 'OhMyPoshEnterKeyHandler' -ScriptBlock (New-EnterKeyHandler { [Microsoft.PowerShell.PSConsoleReadLine]::AcceptLine() } $script:Streaming)

        if ((Get-PSReadLineOption).EditMode -eq "Vi") {
            Set-PSReadLineKeyHandler -ViMode Command -Key Enter -BriefDescription 'OhMyPoshViEnterKeyHandler' -ScriptBlock (New-EnterKeyHandler { [Microsoft.PowerShell.PSConsoleReadLine]::ViAcceptLine() } $script:Streaming)
        }

        # Register Ctrl+C key handlers
        Set-PSReadLineKeyHandler -Key Ctrl+c -BriefDescription 'OhMyPoshCtrlCKeyHandler' -ScriptBlock (New-CtrlCKeyHandler { [Microsoft.PowerShell.PSConsoleReadLine]::CopyOrCancelLine() } $script:Streaming)

        if ((Get-PSReadLineOption).EditMode -eq "Vi") {
            Set-PSReadLineKeyHandler -ViMode Command -Key Ctrl+c -BriefDescription 'OhMyPoshViCtrlCKeyHandler' -ScriptBlock (New-CtrlCKeyHandler { [Microsoft.PowerShell.PSConsoleReadLine]::CancelLine() } $script:Streaming)
        }
    }

    function Enable-PoshLineError {
        $validLine = (Invoke-Utf8Posh @("print", "valid", "--shell=$script:ShellName")) -join "`n"
        $errorLine = (Invoke-Utf8Posh @("print", "error", "--shell=$script:ShellName")) -join "`n"
        Set-PSReadLineOption -PromptText $validLine, $errorLine
    }

    function Enable-PoshVIMode {
        if ($script:ConstrainedLanguageMode) {
            return
        }

        if ((Get-PSReadLineOption).EditMode -ne "Vi") {
            return
        }

        if (-not (Get-Command Set-PSReadLineOption).Parameters.ContainsKey('ViModeChangeHandler')) {
            return
        }

        $env:POSH_VI_MODE = "viins"
        Set-PSReadLineOption -ViModeIndicator Script -ViModeChangeHandler {
            param($mode)

            if ($mode -eq "Command") {
                $env:POSH_VI_MODE = "vicmd"
            }
            else {
                $env:POSH_VI_MODE = "viins"
            }

            $previousOutputEncoding = [Console]::OutputEncoding
            try {
                $script:Streaming.State = 'NEW'
                [Console]::OutputEncoding = [Text.Encoding]::UTF8
                [Microsoft.PowerShell.PSConsoleReadLine]::InvokePrompt()
            }
            catch {
            }
            finally {
                [Console]::OutputEncoding = $previousOutputEncoding
            }
        }
    }

    # perform cleanup on removal so a new initialization in current session works
    if (!$script:ConstrainedLanguageMode) {
        $ExecutionContext.SessionState.Module.OnRemove += {
            # Clean up the serve daemon: ask it to quit and flush its caches,
            # give it a moment to exit on its own, then kill it if it hasn't.
            if ($null -ne $script:Streaming.ServeProcess -and -not $script:Streaming.ServeProcess.HasExited) {
                try {
                    $script:Streaming.StdIn.WriteLine('{"command":"quit"}')
                    $script:Streaming.StdIn.Flush()
                    $script:Streaming.StdIn.Close()
                }
                catch {
                }

                if (-not $script:Streaming.ServeProcess.WaitForExit(500)) {
                    try {
                        $script:Streaming.ServeProcess.Kill()
                    }
                    catch {
                    }
                }
            }

            $script:Streaming.ServeProcess = $null
            $script:Streaming.StdIn = $null

            # Clean up the legacy per-prompt streaming process, if any.
            Stop-StreamingProcess

            if ($null -ne $script:StreamingOnIdleJob) {
                # only remove our own PowerShell.OnIdle subscriber, other modules may have theirs
                Get-EventSubscriber -SourceIdentifier PowerShell.OnIdle -ErrorAction Ignore |
                    Where-Object { $null -ne $_.Action -and $_.Action.InstanceId -eq $script:StreamingOnIdleJob.InstanceId } |
                    Unregister-Event -ErrorAction Ignore
                Remove-Job $script:StreamingOnIdleJob -Force -ErrorAction Ignore
                $script:StreamingOnIdleJob = $null
            }

            if ($null -ne $script:StreamingExitingJob) {
                # only remove our own PowerShell.Exiting subscriber, other modules may have theirs
                Get-EventSubscriber -SourceIdentifier PowerShell.Exiting -ErrorAction Ignore |
                    Where-Object { $null -ne $_.Action -and $_.Action.InstanceId -eq $script:StreamingExitingJob.InstanceId } |
                    Unregister-Event -ErrorAction Ignore
                Remove-Job $script:StreamingExitingJob -Force -ErrorAction Ignore
                $script:StreamingExitingJob = $null
            }

            Remove-Variable -Name _ompStreamingState -Scope Global -ErrorAction Ignore

            Remove-Item Function:Get-PoshStackCount -ErrorAction SilentlyContinue

            $Function:prompt = $script:OriginalPromptFunction

            (Get-PSReadLineOption).ContinuationPrompt = $script:OriginalContinuationPrompt
            (Get-PSReadLineOption).PromptText = $script:OriginalPromptText

            if ((Get-Command Set-PSReadLineOption).Parameters.ContainsKey('ViModeChangeHandler')) {
                Set-PSReadLineOption -ViModeIndicator $script:OriginalViModeIndicator -ViModeChangeHandler $script:OriginalViModeChangeHandler
            }

            Remove-Item Env:POSH_VI_MODE -ErrorAction Ignore

            if ((Get-PSReadLineKeyHandler Spacebar).Function -eq 'OhMyPoshSpaceKeyHandler') {
                Remove-PSReadLineKeyHandler Spacebar
            }

            if ((Get-PSReadLineKeyHandler Enter).Function -eq 'OhMyPoshEnterKeyHandler') {
                Set-PSReadLineKeyHandler Enter -Function AcceptLine
                if ((Get-PSReadLineOption).EditMode -eq "Vi") {
                    Set-PSReadLineKeyHandler -ViMode Command -Key Enter -Function ViAcceptLine
                }
            }

            if ((Get-PSReadLineKeyHandler Ctrl+c).Function -eq 'OhMyPoshCtrlCKeyHandler') {
                Set-PSReadLineKeyHandler Ctrl+c -Function CopyOrCancelLine
                if ((Get-PSReadLineOption).EditMode -eq "Vi") {
                    Set-PSReadLineKeyHandler -ViMode Command -Key Ctrl+c -Function CancelLine
                }
            }
        }
    }

    Export-ModuleMember -Function @(
        "Set-PoshContext"
        "Enable-PoshTooltips"
        "Enable-KeyHandlers"
        "Enable-PoshLineError"
        "Enable-PoshVIMode"
        "Enable-PoshStreaming"
        "Set-TransientPrompt"
        "prompt"
    )
} | Import-Module -Global
