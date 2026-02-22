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

    # Prompt related backup.
    $script:OriginalPromptFunction = $Function:prompt
    $script:OriginalContinuationPrompt = (Get-PSReadLineOption).ContinuationPrompt
    $script:OriginalPromptText = (Get-PSReadLineOption).PromptText

    $script:NoExitCode = $true
    $script:ErrorCode = 0
    $script:ExecutionTime = 0
    $script:ShellName = "pwsh"
    $script:PSVersion = $PSVersionTable.PSVersion.ToString()
    $script:TransientPrompt = $false
    $script:TooltipCommand = ''
    $script:JobCount = 0
    $script:Streaming = [hashtable]::Synchronized(@{
            Process     = $null
            # Runspace    = $null
            AsyncResult = $null
            Prompt      = ''
            State       = 'NEW'
        })

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

    function Stop-StreamingProcess {
        if (-not $global:_ompStreaming) {
            return
        }

        Unregister-Event -SourceIdentifier "OhMyPoshStreaming" -ErrorAction Ignore
        Get-Job -Name "OhMyPoshStreaming" -ErrorAction Ignore | Remove-Job -Force

        # Then kill the streaming process
        if ($null -ne $script:Streaming.Process -and -not $script:Streaming.Process.HasExited) {
            try {
                $script:Streaming.Process.Kill()
            }
            catch {
            }
        }

        $script:Streaming.Process = $null
        $script:Streaming.State = 'NEW'
    }

    function Get-PoshStreamingPrompt {
        # Start streaming process (State stays 'NEW' until initial prompt fully rendered)
        $script:Streaming.Process = New-Object System.Diagnostics.Process
        $StartInfo = $script:Streaming.Process.StartInfo
        $StartInfo.FileName = $global:_ompExecutable

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
        $ps = [powershell]::Create().AddScript({
                param($stream)
                while ($true) {
                    $bytes = [System.Collections.Generic.List[byte]]::new()
                    while (($b = $stream.ReadByte()) -notin -1, 0) {
                        $bytes.Add($b)
                    }

                    if ($bytes.Count -gt 0) {
                        Write-Output ([Text.Encoding]::UTF8.GetString($bytes.ToArray()))
                    }

                    if ($b -eq -1) {
                        return
                    }
                }
            }).AddArgument($script:Streaming.Process.StandardOutput.BaseStream)

        $ps.BeginInvoke($inputData, $output) | Out-Null

        # Update prompt when output arrives
        Register-ObjectEvent -InputObject $output -EventName DataAdded -SourceIdentifier "OhMyPoshStreaming" -MessageData @{
            Streaming = $script:Streaming
        } -Action {
            $s = $event.MessageData.Streaming
            $index = $event.SourceEventArgs.Index

            if ($index -eq 0 -or $s.State -ne 'RUNNING') {
                return
            }

            $previousOutputEncoding = [Console]::OutputEncoding

            try {
                [Console]::OutputEncoding = [Text.Encoding]::UTF8
                $s.Prompt = $event.SourceArgs[0][$index]
                [Microsoft.PowerShell.PSConsoleReadLine]::InvokePrompt()
            }
            catch {}
            finally {
                [Console]::OutputEncoding = $previousOutputEncoding
            }
        } | Out-Null

        while ($output.Count -eq 0) {
            Start-Sleep -Milliseconds 1
        }

        $script:Streaming.Prompt = $output[0]
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

        # Only return cached prompt if we're in a streaming redraw (RUNNING state)
        # AND it's not a transient prompt. Don't use cached prompt for FINAL state
        # as that means the previous prompt is complete and we need a fresh one.
        if ($script:PromptType -ne 'transient' -and $script:Streaming.State -ne 'NEW') {
            # Update ExtraPromptLineCount for PSReadLine to properly clear previous prompt
            Set-PSReadLineOption -ExtraPromptLineCount (($script:Streaming.Prompt | Measure-Object -Line).Lines - 1)
            return $script:Streaming.Prompt
        }

        # Stop any previous streaming process and reset state
        Stop-StreamingProcess

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

        # Now that we're about to return, mark streaming as ready for updates
        if ($global:_ompStreaming -and $script:PromptType -eq 'primary') {
            $script:Streaming.State = 'RUNNING'
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

    # perform cleanup on removal so a new initialization in current session works
    if (!$script:ConstrainedLanguageMode) {
        $ExecutionContext.SessionState.Module.OnRemove += {
            # Clean up streaming process
            Stop-StreamingProcess

            Remove-Item Function:Get-PoshStackCount -ErrorAction SilentlyContinue

            $Function:prompt = $script:OriginalPromptFunction

            (Get-PSReadLineOption).ContinuationPrompt = $script:OriginalContinuationPrompt
            (Get-PSReadLineOption).PromptText = $script:OriginalPromptText

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
        "Set-TransientPrompt"
        "prompt"
    )
} | Import-Module -Global
