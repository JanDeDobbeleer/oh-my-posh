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

New-Module -Name "oh-my-posh-core" -ScriptBlock {
    $script:ErrorCode = 0
    $script:ExecutionTime = 0
    $script:OMPExecutable = ::OMP::
    $script:ShellName = "::SHELL::"
    $script:PSVersion = $PSVersionTable.PSVersion.ToString()
    $script:TransientPrompt = $false
    $env:POWERLINE_COMMAND = "oh-my-posh"
    $env:CONDA_PROMPT_MODIFIER = $false
    if ((::CONFIG:: -ne '') -and (Test-Path -LiteralPath ::CONFIG::)) {
        $env:POSH_THEME = (Resolve-Path -Path ::CONFIG::).ProviderPath
    }

    # specific module support (disabled by default)
    if (($null -eq $env:POSH_GIT_ENABLED) -or ($null -eq (Get-Module 'posh-git'))) {
        $env:POSH_GIT_ENABLED = $false
    }
    if (($null -eq $env:POSH_AZURE_ENABLED) -or ($null -eq (Get-Module 'az'))) {
        $env:POSH_AZURE_ENABLED = $false
    }

    function Initialize-ModuleSupport {
        if ($env:POSH_GIT_ENABLED -eq $true) {
            # We need to set the status so posh-git can facilitate autocomplete
            $global:GitStatus = Get-GitStatus
            $env:POSH_GIT_STATUS = $global:GitStatus | ConvertTo-Json
        }
        if ($env:POSH_AZURE_ENABLED -eq $true) {
            $env:POSH_AZURE_SUBSCRIPTION = Get-AzContext | ConvertTo-Json
        }
    }

    function Start-Utf8Process {
        param(
            [string]$FileName,
            [string[]]$Arguments = @()
        )

        if ($ExecutionContext.SessionState.LanguageMode -eq "ConstrainedLanguage") {
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
        } else {
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
        # we do this to remove a deadlock potential on Windows
        $stdoutTask = $Process.StandardOutput.ReadToEndAsync()
        $stderrTask = $Process.StandardError.ReadToEndAsync()
        $Process.WaitForExit()
        $stderr = $stderrTask.Result.Trim()
        if ($stderr -ne '') {
            $Host.UI.WriteErrorLine($stderr)
        }
        $stdoutTask.Result
    }

    function Set-PoshContext {}

    function Get-CleanPSWD {
        $pswd = $PWD.ToString()
        if ($pswd -ne '/') {
            return $pswd.TrimEnd('\') -replace '^Microsoft\.PowerShell\.Core\\FileSystem::', ''
        }
        return $pswd
    }

    if (("::TOOLTIPS::" -eq "true") -and ($ExecutionContext.SessionState.LanguageMode -ne "ConstrainedLanguage")) {
        Set-PSReadLineKeyHandler -Key Spacebar -BriefDescription 'OhMyPoshSpaceKeyHandler' -ScriptBlock {
            [Microsoft.PowerShell.PSConsoleReadLine]::Insert(' ')
            $position = $host.UI.RawUI.CursorPosition
            $cleanPSWD = Get-CleanPSWD
            $command = $null
            [Microsoft.PowerShell.PSConsoleReadLine]::GetBufferState([ref]$command, [ref]$null)
            $standardOut = @(Start-Utf8Process $script:OMPExecutable @("print", "tooltip", "--error=$script:ErrorCode", "--shell=$script:ShellName", "--pswd=$cleanPSWD", "--config=$env:POSH_THEME", "--command=$command", "--shell-version=$script:PSVersion"))
            Write-Host $standardOut -NoNewline
            $host.UI.RawUI.CursorPosition = $position
            # we need this workaround to prevent the text after cursor from disappearing when the tooltip is rendered
            [Microsoft.PowerShell.PSConsoleReadLine]::Insert(' ')
            [Microsoft.PowerShell.PSConsoleReadLine]::Undo()
        }
    }

    if (("::TRANSIENT::" -eq "true") -and ($ExecutionContext.SessionState.LanguageMode -ne "ConstrainedLanguage")) {
        Set-PSReadLineKeyHandler -Key Enter -BriefDescription 'OhMyPoshEnterKeyHandler' -ScriptBlock {
            $previousOutputEncoding = [Console]::OutputEncoding
            try {
                $parseErrors = $null
                [Microsoft.PowerShell.PSConsoleReadLine]::GetBufferState([ref]$null, [ref]$null, [ref]$parseErrors, [ref]$null)
                if ($parseErrors.Count -eq 0) {
                    $script:TransientPrompt = $true
                    [Console]::OutputEncoding = [Text.Encoding]::UTF8
                    [Microsoft.PowerShell.PSConsoleReadLine]::InvokePrompt()
                }
            } finally {
                # If PSReadline is set to display suggestion list, this workaround is needed to clear the buffer below
                # before accepting the current commandline. The max amount of items in the list is 10, so 12 lines
                # are cleared (10 + 1 more for the prompt + 1 more for current commandline).
                if ((Get-PSReadLineOption).PredictionViewStyle -eq 'ListView') {
                    $terminalHeight = $Host.UI.RawUI.WindowSize.Height
                    # only do this on an valid value
                    if ([int]$terminalHeight -gt 0) {
                        [Microsoft.PowerShell.PSConsoleReadLine]::Insert("`n" * [System.Math]::Min($terminalHeight - $Host.UI.RawUI.CursorPosition.Y - 1, 12))
                        [Microsoft.PowerShell.PSConsoleReadLine]::Undo()
                    }
                }
                [Microsoft.PowerShell.PSConsoleReadLine]::AcceptLine()
                [Console]::OutputEncoding = $previousOutputEncoding
            }
        }
    }

    if ("::ERROR_LINE::" -eq "true") {
        $validLine = @(Start-Utf8Process $script:OMPExecutable @("print", "valid", "--config=$env:POSH_THEME", "--shell=$script:ShellName")) -join "`n"
        $errorLine = @(Start-Utf8Process $script:OMPExecutable @("print", "error", "--config=$env:POSH_THEME", "--shell=$script:ShellName")) -join "`n"
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

        $configString = @(Start-Utf8Process $script:OMPExecutable @("config", "export", "--config=$env:POSH_THEME", "--format=$Format"))
        # if no path, copy to clipboard by default
        if ('' -ne $FilePath) {
            # https://stackoverflow.com/questions/3038337/powershell-resolve-path-that-might-not-exist
            $FilePath = $ExecutionContext.SessionState.Path.GetUnresolvedProviderPathFromPSPath($FilePath)
            [IO.File]::WriteAllLines($FilePath, $configString)
        } else {
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
        if ("" -eq $name) {
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

        if ($Path -eq "") {
            do {
                $temp = Read-Host 'Please enter the themes path'
            }
            while (-not (Test-Path -LiteralPath $temp))
            $Path = (Resolve-Path -Path $temp).ProviderPath
        }

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
        } else {
            $cleanPSWD = Get-CleanPSWD
            $themes | ForEach-Object -Process {
                Write-Host "Theme: $(Get-FileHyperlink -uri $_.FullName -Name ($_.BaseName -replace '\.omp$', ''))`n"
                @(Start-Utf8Process $script:OMPExecutable @("print", "primary", "--config=$($_.FullName)", "--pswd=$cleanPSWD", "--shell=$script:ShellName"))
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
        if (-not ((Get-PSCallStack).Location -join "").StartsWith("<")) {
            $script:PromptType = "debug"
            return
        }
        $script:PromptType = "primary"
        Initialize-ModuleSupport
    }

    function Update-PoshErrorCode {
        $lastHistory = Get-History -ErrorAction Ignore -Count 1
        # error code should be updated only when a non-empty command is run
        if (($null -eq $lastHistory) -or ($script:LastHistoryId -eq $lastHistory.Id)) {
            $script:ExecutionTime = 0
            return
        }
        $script:LastHistoryId = $lastHistory.Id
        $script:ExecutionTime = ($lastHistory.EndExecutionTime - $lastHistory.StartExecutionTime).TotalMilliseconds
        if ($script:OriginalLastExecutionStatus) {
            $script:ErrorCode = 0
            return
        }
        $invocationInfo = try {
            # retrieve info of the most recent error
            $global:Error[0] | Where-Object { $_ -ne $null } | Select-Object -ExpandProperty InvocationInfo
        } catch { $null }
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

    function prompt {
        # store the orignal last command execution status and last exit code
        $script:OriginalLastExecutionStatus = $?
        $script:OriginalLastExitCode = $global:LASTEXITCODE

        Set-PoshPromptType
        if ($script:PromptType -ne 'transient') {
            Update-PoshErrorCode
        }
        $cleanPSWD = Get-CleanPSWD
        $stackCount = global:Get-PoshStackCount
        Set-PoshContext
        $terminalWidth = $Host.UI.RawUI.WindowSize.Width
        # set a sane default when the value can't be retrieved
        if (-not $terminalWidth) {
            $terminalWidth = 0
        }
        $standardOut = @(Start-Utf8Process $script:OMPExecutable @("print", $script:PromptType, "--error=$script:ErrorCode", "--pswd=$cleanPSWD", "--execution-time=$script:ExecutionTime", "--stack-count=$stackCount", "--config=$env:POSH_THEME", "--shell-version=$script:PSVersion", "--terminal-width=$terminalWidth", "--shell=$script:ShellName"))
        # make sure PSReadLine knows if we have a multiline prompt
        Set-PSReadLineOption -ExtraPromptLineCount (($standardOut | Measure-Object -Line).Lines - 1)
        # the output can be multiline, joining these ensures proper rendering by adding line breaks with `n
        $standardOut -join "`n"

        # remove any posh-git status
        $env:POSH_GIT_STATUS = $null

        # restore the orignal last exit code
        $global:LASTEXITCODE = $script:OriginalLastExitCode
    }

    # set secondary prompt
    Set-PSReadLineOption -ContinuationPrompt (@(Start-Utf8Process $script:OMPExecutable @("print", "secondary", "--config=$env:POSH_THEME", "--shell=$script:ShellName")) -join "`n")

    # legacy functions
    function Enable-PoshTooltips {}
    function Enable-PoshTransientPrompt {}
    function Enable-PoshLineError {}

    # perform cleanup on removal so a new initialization in current session works
    if ($ExecutionContext.SessionState.LanguageMode -ne "ConstrainedLanguage") {
        $ExecutionContext.SessionState.Module.OnRemove += {
            if ((Get-PSReadLineKeyHandler -Key Spacebar).Function -eq 'OhMyPoshSpaceKeyHandler') {
                Remove-PSReadLineKeyHandler -Key Spacebar
            }
            if ((Get-PSReadLineKeyHandler -Key Enter).Function -eq 'OhMyPoshEnterKeyHandler') {
                Set-PSReadLineKeyHandler -Key Enter -Function AcceptLine
            }
        }
    }

    Export-ModuleMember -Function @(
        "Set-PoshContext"
        "Enable-PoshTooltips"
        "Enable-PoshTransientPrompt"
        "Enable-PoshLineError"
        "Export-PoshTheme"
        "Get-PoshThemes"
        "prompt"
    )
} | Import-Module -Global
