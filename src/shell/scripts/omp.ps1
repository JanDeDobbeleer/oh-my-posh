# remove any existing dynamic module of OMP
if ($null -ne (Get-Module -Name "oh-my-posh-core")) {
    Remove-Module -Name "oh-my-posh-core" -Force
}
New-Module -Name "oh-my-posh-core" -ScriptBlock {
    $script:ErrorCode = 0
    $script:OMPExecutable = '::OMP::'
    $script:PSVersion = $PSVersionTable.PSVersion.ToString()
    $script:TransientPrompt = $false
    $env:POWERLINE_COMMAND = "oh-my-posh"
    $env:CONDA_PROMPT_MODIFIER = $false
    if (Test-Path '::CONFIG::') {
        $env:POSH_THEME = (Resolve-Path -Path '::CONFIG::').ProviderPath
    }
    # specific module support (disabled by default)
    if ($null -eq $env:POSH_GIT_ENABLED) {
        $env:POSH_GIT_ENABLED = $false
    }

    function Start-Utf8Process {
        param(
            [string] $FileName,
            [string] $Arguments
        )

        $Process = New-Object System.Diagnostics.Process
        $StartInfo = $Process.StartInfo
        $StartInfo.StandardErrorEncoding = $StartInfo.StandardOutputEncoding = [System.Text.Encoding]::UTF8
        $StartInfo.RedirectStandardError = $StartInfo.RedirectStandardInput = $StartInfo.RedirectStandardOutput = $true
        $StartInfo.FileName = $Filename
        $StartInfo.Arguments = $Arguments
        $StartInfo.UseShellExecute = $false
        if ($PWD.Provider.Name -eq 'FileSystem') {
            $StartInfo.WorkingDirectory = $PWD.ProviderPath
        }
        $StartInfo.CreateNoWindow = $true
        [void]$Process.Start()
        $stdoutTask = $Process.StandardOutput.ReadToEndAsync()
        $stderrTask = $Process.StandardError.ReadToEndAsync()
        [void]$Process.WaitForExit()
        $stderr = $stderrTask.Result.Trim()
        if ($stderr -ne '') {
            $Host.UI.WriteErrorLine($stderr)
        }
        $stdoutTask.Result
    }

    function Set-PoshContext {}

    function Get-PoshContext {
        $cleanPWD = $PWD.ProviderPath
        $cleanPSWD = $PWD.ToString()
        $cleanPWD = $cleanPWD.TrimEnd('\')
        $cleanPSWD = $cleanPSWD.TrimEnd('\')
        return $cleanPWD, $cleanPSWD
    }

    function Initialize-ModuleSupport {
        if ($env:POSH_GIT_ENABLED -eq $true) {
            $env:POSH_GIT_STATUS = Write-GitStatus -Status (Get-GitStatus)
        }
    }

    function Enable-PoshTooltips {
        Set-PSReadlineKeyHandler -Key SpaceBar -ScriptBlock {
            [Microsoft.PowerShell.PSConsoleReadLine]::Insert(' ')
            $position = $host.UI.RawUI.CursorPosition
            $cleanPWD, $cleanPSWD = Get-PoshContext
            $command = $null
            $cursor = $null
            [Microsoft.PowerShell.PSConsoleReadLine]::GetBufferState([ref]$command, [ref]$cursor)
            $command = ($command -split " ")[0]
            $standardOut = @(Start-Utf8Process $script:OMPExecutable "print tooltip --pwd=""$cleanPWD"" --pswd=""$cleanPSWD"" --config=""$env:POSH_THEME"" --command=""$command"" --shell-version=""$script:PSVersion""")
            Write-Host $standardOut -NoNewline
            $host.UI.RawUI.CursorPosition = $position
        }
    }

    function Enable-PoshTransientPrompt {
        Set-PSReadlineKeyHandler -Key Enter -ScriptBlock {
            $script:TransientPrompt = $true
            $previousOutputEncoding = [Console]::OutputEncoding
            [Console]::OutputEncoding = [Text.Encoding]::UTF8
            try {
                [Microsoft.PowerShell.PSConsoleReadLine]::InvokePrompt()
            } finally {
                [Microsoft.PowerShell.PSConsoleReadLine]::AcceptLine()
                [Console]::OutputEncoding = $previousOutputEncoding
            }
        }
    }

    function Enable-PoshLineError {
        $validLine = @(Start-Utf8Process $script:OMPExecutable "print valid --config=""$env:POSH_THEME""") -join "`n"
        $errorLine = @(Start-Utf8Process $script:OMPExecutable "print error --config=""$env:POSH_THEME""") -join "`n"
        Set-PSReadLineOption -PromptText $validLine, $errorLine
    }

    <#
    .SYNOPSIS
        Exports the current oh-my-posh theme.
    .DESCRIPTION
        By default the config is exported in JSON to the clipboard.
    .EXAMPLE
        Export-PoshTheme

        Export the current theme in JSON to the clipboard.
    .EXAMPLE
        Export-PoshTheme -Format toml

        Export the current theme in TOML to the clipboard.
    .EXAMPLE
        Export-PoshTheme C:\temp\theme.yaml yaml

        Export the current theme in YAML to 'C:\temp\theme.yaml'.
    .EXAMPLE
        Export-PoshTheme ~\theme.toml toml

        Export the current theme in TOML to '$HOME\theme.toml'
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
            # The format of the theme
            $Format = 'json'
        )

        $configString = @(Start-Utf8Process $script:OMPExecutable "config export --config=""$env:POSH_THEME"" --format=$Format")
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
            while (-not (Test-Path -Path $temp))
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
        $themes = Get-ChildItem -Path "$Path\*" -Include '*.omp.json' | Sort-Object Name
        if ($List -eq $true) {
            $themes | Select-Object @{ Name = 'hyperlink'; Expression = { Get-FileHyperlink -uri $_.FullName } } | Format-Table -HideTableHeaders
        } else {
            $themes | ForEach-Object -Process {
                Write-Host "Theme: $(Get-FileHyperlink -uri $_.FullName -Name ($_.BaseName -replace '\.omp$', ''))`n"
                @(Start-Utf8Process $script:OMPExecutable "print primary --config=""$($_.FullName)"" --pwd=""$PWD"" --shell pwsh")
                Write-Host "`n"
            }
        }
        Write-Host @"

Themes location: $(Get-FileHyperlink -uri "$Path")

To change your theme, adjust the init script in $PROFILE.
Example:
  oh-my-posh init pwsh --config $Path/jandedobbeleer.omp.json | Invoke-Expression

"@
    }

    function prompt {
        # store if the last command was successful
        $lastCommandSuccess = $?
        # store the last exit code for restore
        $realLASTEXITCODE = $global:LASTEXITCODE
        $cleanPWD, $cleanPSWD = Get-PoshContext
        if ($script:TransientPrompt -eq $true) {
            @(Start-Utf8Process $script:OMPExecutable "print transient --error=$script:ErrorCode --pwd=""$cleanPWD"" --pswd=""$cleanPSWD"" --execution-time=$script:ExecutionTime --config=""$env:POSH_THEME"" --shell-version=""$script:PSVersion""") -join "`n"
            $script:TransientPrompt = $false
            return
        }
        if (Test-Path variable:/PSDebugContext) {
            @(Start-Utf8Process $script:OMPExecutable "print debug --pwd=""$cleanPWD"" --pswd=""$cleanPSWD"" --config=""$env:POSH_THEME""") -join "`n"
            return
        }
        Initialize-ModuleSupport

        $script:ExecutionTime = -1
        $lastHistory = Get-History -ErrorAction Ignore -Count 1
        if ($null -ne $lastHistory -and $script:LastHistoryId -ne $lastHistory.Id) {
            $script:LastHistoryId = $lastHistory.Id
            $script:ExecutionTime = ($lastHistory.EndExecutionTime - $lastHistory.StartExecutionTime).TotalMilliseconds
            # error code should be changed only when a non-empty command is called
            $script:ErrorCode = 0
            if (!$lastCommandSuccess) {
                $invocationInfo = try {
                    # retrieve info of the most recent error
                    $global:Error[0] | Where-Object { $_ -ne $null } | Select-Object -ExpandProperty InvocationInfo
                } catch { $null }
                # check if the last command caused the last error
                if ($null -ne $invocationInfo -and $lastHistory.CommandLine -eq $invocationInfo.Line) {
                    $script:ErrorCode = 1
                } elseif ($realLASTEXITCODE -is [int] -and $realLASTEXITCODE -ne 0) {
                    # native app exit code
                    $script:ErrorCode = $realLASTEXITCODE
                }
            }
        }

        $stackCount = (Get-Location -Stack).Count
        try {
            if ($null -ne $global:OMP_GLOBAL_SESSIONSTATE) {
                $stackCount = $global:OMP_GLOBAL_SESSIONSTATE.Path.LocationStack('').Count
            }
        } catch {}

        Set-PoshContext
        $terminalWidth = $Host.UI.RawUI.WindowSize.Width
        $standardOut = @(Start-Utf8Process $script:OMPExecutable "print primary --error=$script:ErrorCode --pwd=""$cleanPWD"" --pswd=""$cleanPSWD"" --execution-time=$script:ExecutionTime --stack-count=$stackCount --config=""$env:POSH_THEME"" --shell-version=""$script:PSVersion"" --terminal-width=$terminalWidth")
        # make sure PSReadLine knows we have a multiline prompt
        $extraLines = ($standardOut | Measure-Object -Line).Lines - 1
        if ($extraLines -gt 0) {
            Set-PSReadlineOption -ExtraPromptLineCount $extraLines
        }
        # the output can be multiline, joining these ensures proper rendering by adding line breaks with `n
        $standardOut -join "`n"
        $global:LASTEXITCODE = $realLASTEXITCODE
    }

    # set secondary prompt
    Set-PSReadLineOption -ContinuationPrompt (@(Start-Utf8Process $script:OMPExecutable "print secondary --config=""$env:POSH_THEME""") -join "`n")

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
