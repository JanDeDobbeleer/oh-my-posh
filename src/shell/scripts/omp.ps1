# remove any existing dynamic module of OMP
if ($null -ne (Get-Module -Name "oh-my-posh-core")) {
    Remove-Module -Name "oh-my-posh-core" -Force
}

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
    $script:ShellName = "::SHELL::"
    $script:PSVersion = $PSVersionTable.PSVersion.ToString()
    $script:TransientPrompt = $false
    $script:TooltipCommand = ''
    $script:JobCount = 0

    $env:POWERLINE_COMMAND = "oh-my-posh"
    $env:POSH_SHELL_VERSION = $script:PSVersion
    $env:POSH_PID = $PID
    $env:CONDA_PROMPT_MODIFIER = $false

    # set the default theme
    if (::CONFIG:: -and (Test-Path -LiteralPath ::CONFIG::)) {
        $env:POSH_THEME = (Resolve-Path -Path ::CONFIG::).ProviderPath
    }

    function Start-Utf8Process {
        param(
            [string]$FileName,
            [string[]]$Arguments = @()
        )

        if ($script:ConstrainedLanguageMode) {
            $standardOut = Invoke-Expression "& `$FileName `$Arguments 2>&1"
            $standardOut -join "`n"
            return
        }

        $Process = New-Object System.Diagnostics.Process
        $StartInfo = $Process.StartInfo
        $StartInfo.FileName = $FileName
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

    function Set-PoshContext([bool]$originalStatus) {}

    function Get-CleanPSWD {
        $pswd = $PWD.ToString()
        if ($pswd -ne '/') {
            return $pswd.TrimEnd('\') -replace '^Microsoft\.PowerShell\.Core\\FileSystem::', ''
        }
        return $pswd
    }

    function Get-TerminalWidth {
        $terminalWidth = $Host.UI.RawUI.WindowSize.Width
        # Set a sane default when the value can't be retrieved.
        if (-not $terminalWidth) {
            return 0
        }
        $terminalWidth
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
                $column = $Host.UI.RawUI.CursorPosition.X
                $terminalWidth = Get-TerminalWidth
                $cleanPSWD = Get-CleanPSWD
                $stackCount = global:Get-PoshStackCount

                $standardOut = (Start-Utf8Process $global:_ompExecutable @("print", "tooltip", "--status=$script:ErrorCode", "--shell=$script:ShellName", "--pswd=$cleanPSWD", "--execution-time=$script:ExecutionTime", "--stack-count=$stackCount", "--command=$command", "--shell-version=$script:PSVersion", "--column=$column", "--terminal-width=$terminalWidth", "--no-status=$script:NoExitCode", "--job-count=$script:JobCount")) -join ''
                if (!$standardOut) {
                    return
                }

                Write-Host $standardOut -NoNewline

                # Workaround to prevent the text after cursor from disappearing when the tooltip is printed.
                [Microsoft.PowerShell.PSConsoleReadLine]::Insert(' ')
                [Microsoft.PowerShell.PSConsoleReadLine]::Undo()
            }
            finally {}
        }
    }

    function Set-TransientPrompt {
        $previousOutputEncoding = [Console]::OutputEncoding
        try {
            $script:TransientPrompt = $true
            [Console]::OutputEncoding = [Text.Encoding]::UTF8
            [Microsoft.PowerShell.PSConsoleReadLine]::InvokePrompt()
        }
        finally {
            [Console]::OutputEncoding = $previousOutputEncoding
        }
    }

    function Enable-PoshTransientPrompt {
        if ($script:ConstrainedLanguageMode) {
            return
        }

        Set-PSReadLineKeyHandler -Key Enter -BriefDescription 'OhMyPoshEnterKeyHandler' -ScriptBlock {
            try {
                $parseErrors = $null
                [Microsoft.PowerShell.PSConsoleReadLine]::GetBufferState([ref]$null, [ref]$null, [ref]$parseErrors, [ref]$null)
                $executingCommand = $parseErrors.Count -eq 0
                if ($executingCommand) {
                    $script:TooltipCommand = ''
                    Set-TransientPrompt
                }
            }
            finally {
                [Microsoft.PowerShell.PSConsoleReadLine]::AcceptLine()
                if ($global:_ompFTCSMarks -and $executingCommand) {
                    # Write FTCS_COMMAND_EXECUTED after accepting the input - it should still happen before execution
                    Write-Host "$([char]0x1b)]133;C`a" -NoNewline
                }
            }
        }

        Set-PSReadLineKeyHandler -Key Ctrl+c -BriefDescription 'OhMyPoshCtrlCKeyHandler' -ScriptBlock {
            try {
                $start = $null
                [Microsoft.PowerShell.PSConsoleReadLine]::GetSelectionState([ref]$start, [ref]$null)
                # only render a transient prompt when no text is selected
                if ($start -eq -1) {
                    $script:TooltipCommand = ''
                    Set-TransientPrompt
                }
            }
            finally {
                [Microsoft.PowerShell.PSConsoleReadLine]::CopyOrCancelLine()
            }
        }
    }

    function Enable-PoshLineError {
        $validLine = (Start-Utf8Process $global:_ompExecutable @("print", "valid", "--shell=$script:ShellName")) -join "`n"
        $errorLine = (Start-Utf8Process $global:_ompExecutable @("print", "error", "--shell=$script:ShellName")) -join "`n"
        Set-PSReadLineOption -PromptText $validLine, $errorLine
    }

    <#
    .SYNOPSIS
        Exports the current oh-my-posh theme.
    .DESCRIPTION
        By default the config is exported in JSON to the clipboard.
    .EXAMPLE
        Export-PoshTheme

        Exports the current theme in JSON to the clipboard.
    .EXAMPLE
        Export-PoshTheme -Format toml

        Exports the current theme in TOML to the clipboard.
    .EXAMPLE
        Export-PoshTheme C:\temp\theme.yaml yaml

        Exports the current theme in YAML to 'C:\temp\theme.yaml'.
    .EXAMPLE
        Export-PoshTheme ~\theme.toml toml

        Exports the current theme in TOML to '$HOME\theme.toml'
    #>
    function Export-PoshTheme {
        param(
            [Parameter(Mandatory = $false)]
            [string]
            # The file path where the theme will be exported. If not provided, the config is copied to the clipboard by default.
            $FilePath,
            [Parameter(Mandatory = $false)]
            [ValidateSet('json', 'yaml', 'toml')]
            [string]
            # The format of the theme.
            $Format = 'json'
        )

        $configString = Start-Utf8Process $global:_ompExecutable @("config", "export", "--format=$Format")
        # if no path, copy to clipboard by default
        if ($FilePath) {
            # https://stackoverflow.com/questions/3038337/powershell-resolve-path-that-might-not-exist
            $FilePath = $ExecutionContext.SessionState.Path.GetUnresolvedProviderPathFromPSPath($FilePath)
            [IO.File]::WriteAllLines($FilePath, $configString)
        }
        else {
            Set-Clipboard $configString
            Write-Output "Theme copied to clipboard"
        }
    }

    function Get-FileHyperlink {
        param(
            [Parameter(Mandatory, ValuefromPipeline = $True)]
            [string]$uri,
            [Parameter(ValuefromPipeline = $True)]
            [string]$name
        )

        $esc = [char]27
        if (!$name) {
            # if name not set, uri is used as the name of the hyperlink
            $name = $uri
        }

        if ($null -ne $env:WSL_DISTRO_NAME) {
            # wsl conversion if needed
            $uri = &wslpath -m $uri
        }

        # return an ANSI formatted hyperlink
        return "$esc]8;;file://$uri$esc\$name$esc]8;;$esc\"
    }

    function Get-PoshThemes {
        param(
            [Parameter(Mandatory = $false, HelpMessage = "The themes folder")]
            [string]
            $Path = $env:POSH_THEMES_PATH,
            [switch]
            [Parameter(Mandatory = $false, HelpMessage = "List themes path")]
            $List
        )

        while (-not (Test-Path -LiteralPath $Path)) {
            $Path = Read-Host 'Please enter the themes path'
        }

        $Path = (Resolve-Path -Path $Path).ProviderPath

        $logo = @'
   __  _____ _      ___  ___       ______         _      __
  / / |  _  | |     |  \/  |       | ___ \       | |     \ \
 / /  | | | | |__   | .  . |_   _  | |_/ /__  ___| |__    \ \
< <   | | | | '_ \  | |\/| | | | | |  __/ _ \/ __| '_ \    > >
 \ \  \ \_/ / | | | | |  | | |_| | | | | (_) \__ \ | | |  / /
  \_\  \___/|_| |_| \_|  |_/\__, | \_|  \___/|___/_| |_| /_/
                             __/ |
                            |___/
'@
        Write-Host $logo
        $themes = Get-ChildItem -Path "$Path/*" -Include '*.omp.json' | Sort-Object Name
        if ($List -eq $true) {
            $themes | Select-Object @{ Name = 'hyperlink'; Expression = { Get-FileHyperlink -uri $_.FullName } } | Format-Table -HideTableHeaders
        }
        else {
            $cleanPSWD = Get-CleanPSWD
            $themes | ForEach-Object -Process {
                Write-Host "Theme: $(Get-FileHyperlink -uri $_.FullName -Name ($_.BaseName -replace '\.omp$', ''))`n"
                Start-Utf8Process $global:_ompExecutable @("print", "primary", "--config=$($_.FullName)", "--pswd=$cleanPSWD", "--shell=$script:ShellName")
                Write-Host "`n"
            }
        }
        Write-Host @"

Themes location: $(Get-FileHyperlink -uri "$Path")

To change your theme, adjust the init script in $PROFILE.
Example:
  oh-my-posh init pwsh --config '$((Join-Path $Path "jandedobbeleer.omp.json") -replace "'", "''")' | Invoke-Expression

"@
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
            catch {}
        }

        if ($global:_ompPoshGit) {
            try {
                $global:GitStatus = Get-GitStatus
                $env:POSH_GIT_STATUS = $global:GitStatus | ConvertTo-Json
            }
            catch {}
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
            $global:Error[0] | Where-Object { $_ -ne $null } | Select-Object -ExpandProperty InvocationInfo
        }
        catch { $null }

        # check if the last command caused the last error
        if ($null -ne $invocationInfo -and $lastHistory.CommandLine -eq $invocationInfo.Line) {
            $script:ErrorCode = 1
            return
        }

        if ($script:OriginalLastExitCode -is [int] -and $script:OriginalLastExitCode -ne 0) {
            # native app exit code
            $script:ErrorCode = $script:OriginalLastExitCode
            return
        }
    }

    $promptFunction = {
        # store the orignal last command execution status
        if ($global:NVS_ORIGINAL_LASTEXECUTIONSTATUS -is [bool]) {
            # make it compatible with NVS auto-switching, if enabled
            $script:OriginalLastExecutionStatus = $global:NVS_ORIGINAL_LASTEXECUTIONSTATUS
        }
        else {
            $script:OriginalLastExecutionStatus = $?
        }
        # store the orignal last exit code
        $script:OriginalLastExitCode = $global:LASTEXITCODE

        Set-PoshPromptType

        if ($script:PromptType -ne 'transient') {
            Update-PoshErrorCode
        }

        Set-PoshContext $script:ErrorCode

        $cleanPSWD = Get-CleanPSWD
        $stackCount = global:Get-PoshStackCount
        $terminalWidth = Get-TerminalWidth

        # set the cursor positions, they are zero based so align with other platforms
        $env:POSH_CURSOR_LINE = $Host.UI.RawUI.CursorPosition.Y + 1
        $env:POSH_CURSOR_COLUMN = $Host.UI.RawUI.CursorPosition.X + 1

        $standardOut = Start-Utf8Process $global:_ompExecutable @("print", $script:PromptType, "--status=$script:ErrorCode", "--pswd=$cleanPSWD", "--execution-time=$script:ExecutionTime", "--stack-count=$stackCount", "--shell-version=$script:PSVersion", "--terminal-width=$terminalWidth", "--shell=$script:ShellName", "--no-status=$script:NoExitCode", "--job-count=$script:JobCount")
        # make sure PSReadLine knows if we have a multiline prompt
        Set-PSReadLineOption -ExtraPromptLineCount (($standardOut | Measure-Object -Line).Lines - 1)

        # The output can be multi-line, joining them ensures proper rendering.
        $standardOut -join "`n"

        # remove any posh-git status
        $env:POSH_GIT_STATUS = $null

        # restore the orignal last exit code
        $global:LASTEXITCODE = $script:OriginalLastExitCode
    }

    $Function:prompt = $promptFunction

    # set secondary prompt
    Set-PSReadLineOption -ContinuationPrompt ((Start-Utf8Process $global:_ompExecutable @("print", "secondary", "--shell=$script:ShellName")) -join "`n")

    # perform cleanup on removal so a new initialization in current session works
    if (!$script:ConstrainedLanguageMode) {
        $ExecutionContext.SessionState.Module.OnRemove += {
            Remove-Item Function:Get-PoshStackCount
            $Function:prompt = $script:OriginalPromptFunction

            (Get-PSReadLineOption).ContinuationPrompt = $script:OriginalContinuationPrompt
            (Get-PSReadLineOption).PromptText = $script:OriginalPromptText

            if ((Get-PSReadLineKeyHandler Spacebar).Function -eq 'OhMyPoshSpaceKeyHandler') {
                Remove-PSReadLineKeyHandler Spacebar
            }

            if ((Get-PSReadLineKeyHandler Enter).Function -eq 'OhMyPoshEnterKeyHandler') {
                Set-PSReadLineKeyHandler Enter -Function AcceptLine
            }

            if ((Get-PSReadLineKeyHandler Ctrl+c).Function -eq 'OhMyPoshCtrlCKeyHandler') {
                Set-PSReadLineKeyHandler Ctrl+c -Function CopyOrCancelLine
            }
        }
    }

    $notice = Start-Utf8Process $global:_ompExecutable @("notice")
    if ($notice) {
        Write-Host $notice -NoNewline
    }

    Export-ModuleMember -Function @(
        "Set-PoshContext"
        "Enable-PoshTooltips"
        "Enable-PoshTransientPrompt"
        "Enable-PoshLineError"
        "Export-PoshTheme"
        "Get-PoshThemes"
        "Start-Utf8Process"
        "prompt"
    )
} | Import-Module -Global
