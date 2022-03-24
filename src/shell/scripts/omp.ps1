# Powershell doesn't default to UTF8 just yet, so we're forcing it as there are too many problems
# that pop up when we don't
if ($ExecutionContext.SessionState.LanguageMode -ne "ConstrainedLanguage") {
    [console]::InputEncoding = [console]::OutputEncoding = New-Object System.Text.UTF8Encoding
}
elseif ($env:POSH_CONSTRAINED_LANGUAGE_MODE -ne $true) {
    Write-Host "[WARNING] ConstrainedLanguage mode detected, unable to set console to UTF-8.
When using PowerShell in ConstrainedLanguage mode, please set the
console mode manually to UTF-8. See here for more information:
https://ohmyposh.dev/docs/faq#powershell-running-in-constrainedlanguage-mode
"
}

$env:POWERLINE_COMMAND = "oh-my-posh"
$env:CONDA_PROMPT_MODIFIER = $false

# specific module support (disabled by default)
$omp_value = $env:POSH_GIT_ENABLED
if ($null -eq $omp_value) {
    $env:POSH_GIT_ENABLED = $false
}

# used to detect empty hit
$global:omp_lastHistoryId = -1

$omp_config = "::CONFIG::"
if (Test-Path $omp_config) {
    $env:POSH_THEME = (Resolve-Path -Path $omp_config).ProviderPath
}

Remove-Variable omp_value -Confirm:$false
Remove-Variable omp_config -Confirm:$false

# set secondary prompt
$secondaryPrompt = @(&"::OMP::" prompt print secondary --config="$Env:POSH_THEME" 2>&1) -join "`n"
Set-PSReadLineOption -ContinuationPrompt $secondaryPrompt

function global:Set-PoshContext {}

function global:Get-PoshContext {
    $config = $env:POSH_THEME
    $cleanPWD = $PWD.ProviderPath
    $cleanPSWD = $PWD.ToString()
    return $config, $cleanPWD, $cleanPSWD
}

function global:Initialize-ModuleSupport {
    if ($env:POSH_GIT_ENABLED -eq $true -and (Get-Module -Name "posh-git")) {
        [Diagnostics.CodeAnalysis.SuppressMessageAttribute('PSProvideCommentHelp', '', Justification = 'Variable used later (not in this scope)')]
        $global:GitStatus = Get-GitStatus
        [Diagnostics.CodeAnalysis.SuppressMessageAttribute('PSProvideCommentHelp', '', Justification = 'Variable used by posh-git (not in this script)')]
        $GitPromptSettings = $null
        $env:POSH_GIT_STATUS = Write-GitStatus -Status $global:GitStatus
    }
}

[ScriptBlock]$Prompt = {
    #store if the last command was successful
    $lastCommandSuccess = $?
    #store the last exit code for restore
    $realLASTEXITCODE = $global:LASTEXITCODE
    $omp = "::OMP::"
    $config, $cleanPWD, $cleanPSWD = Get-PoshContext
    if ($env:POSH_TRANSIENT -eq $true) {
        $standardOut = @(&$omp prompt print transient --pwd="$cleanPWD" --pswd="$cleanPSWD" --config="$config" 2>&1)
        $standardOut -join "`n"
        $env:POSH_TRANSIENT = $false
        return
    }
    if (Test-Path variable:/PSDebugContext) {
        $standardOut = @(&$omp prompt print debug --pwd="$cleanPWD" --pswd="$cleanPSWD" --config="$config" 2>&1)
        $standardOut -join "`n"
        return
    }
    $errorCode = 0
    Initialize-ModuleSupport
    Set-PoshContext
    if ($lastCommandSuccess -eq $false) {
        #native app exit code
        if ($realLASTEXITCODE -is [int] -and $realLASTEXITCODE -gt 0) {
            $errorCode = $realLASTEXITCODE
        }
        else {
            $errorCode = 1
        }
    }

    # read stack count from current stack(if invoked from profile=right value,otherwise use the global variable set in Set-PoshPrompt(stack scoped to module))
    $stackCount = (Get-Location -stack).Count
    try {
        if ($global:omp_global_sessionstate -ne $null) {
            $stackCount = ($global:omp_global_sessionstate).path.locationstack('').count
        }
    }
    catch {}

    $executionTime = -1
    $history = Get-History -ErrorAction Ignore -Count 1
    if ($null -ne $history -and $null -ne $history.EndExecutionTime -and $null -ne $history.StartExecutionTime -and $global:omp_lastHistoryId -ne $history.Id) {
        $executionTime = ($history.EndExecutionTime - $history.StartExecutionTime).TotalMilliseconds
        $global:omp_lastHistoryId = $history.Id
    }
    $terminalWidth = $Host.UI.RawUI.WindowSize.Width
    $standardOut = @(&$omp prompt print primary --error="$errorCode" --pwd="$cleanPWD" --pswd="$cleanPSWD" --execution-time="$executionTime" --stack-count="$stackCount" --config="$config" --terminal-width=$terminalWidth 2>&1)
    # make sure PSReadLine knows we have a multiline prompt
    $extraLines = $standardOut.Count - 1
    if ($extraLines -gt 0) {
        Set-PSReadlineOption -ExtraPromptLineCount $extraLines
    }
    # the output can be multiline, joining these ensures proper rendering by adding line breaks with `n
    $standardOut -join "`n"
    $global:LASTEXITCODE = $realLASTEXITCODE
    #remove temp variables
    Remove-Variable realLASTEXITCODE -Confirm:$false
    Remove-Variable lastCommandSuccess -Confirm:$false
}
Set-Item -Path Function:prompt -Value $Prompt -Force

function global:Write-PoshDebug {
    $omp = "::OMP::"
    $config, $cleanPWD, $cleanPSWD = Get-PoshContext
    $standardOut = @(&$omp prompt debug --config="$config" 2>&1)
    $standardOut -join "`n"
}

<#
.SYNOPSIS
    Exports the current oh-my-posh theme
.DESCRIPTION
    By default the config is exported in json to the clipboard
.EXAMPLE
    Export-PoshTheme
    Current theme exported in json to clipboard
.EXAMPLE
    Export-PoshTheme -Format toml
    Current theme exported in toml to clipboard
.EXAMPLE
    Export-PoshTheme c:\temp\theme.toml toml
    Current theme exported in toml to c:\temp\theme.toml
.EXAMPLE
    Export-PoshTheme ~\theme.toml toml
    Current theme exported in toml to your home\theme.toml
#>
function global:Export-PoshTheme {
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

    $config = $env:POSH_THEME
    $omp = "::OMP::"
    $configString = @(&$omp config export --config="$config" --format="$Format" 2>&1)
    # if no path, copy to clipboard by default
    if ($FilePath -ne "") {
        #https://stackoverflow.com/questions/3038337/powershell-resolve-path-that-might-not-exist
        $FilePath = $ExecutionContext.SessionState.Path.GetUnresolvedProviderPathFromPSPath($FilePath)
        [IO.File]::WriteAllLines($FilePath, $configString)
    }
    else {
        Set-Clipboard $configString
        Write-Output "Theme copied to clipboard"
    }
}

function global:Enable-PoshTooltips {
    Set-PSReadlineKeyHandler -Key SpaceBar -ScriptBlock {
        [Microsoft.PowerShell.PSConsoleReadLine]::Insert(' ')
        $position = $host.UI.RawUI.CursorPosition
        $omp = "::OMP::"
        $config, $cleanPWD, $cleanPSWD = Get-PoshContext
        $command = $null
        $cursor = $null
        [Microsoft.PowerShell.PSConsoleReadLine]::GetBufferState([ref]$command, [ref]$cursor)
        $command = $command -replace "`"", "'"
        $standardOut = @(&$omp prompt print tooltip --pwd="$cleanPWD" --pswd="$cleanPSWD" --config="$config" --command="$command" 2>&1)
        Write-Host $standardOut -NoNewline
        $host.UI.RawUI.CursorPosition = $position
    }
}

function global:Enable-PoshTransientPrompt {
    Set-PSReadlineKeyHandler -Key Enter -ScriptBlock {
        $env:POSH_TRANSIENT = $true
        [Microsoft.PowerShell.PSConsoleReadLine]::InvokePrompt()
        [Microsoft.PowerShell.PSConsoleReadLine]::AcceptLine()
    }
}

function global:Enable-PoshLineError {
    $omp = "::OMP::"
    $validLine = @(&$omp prompt print valid --config="$Env:POSH_THEME" 2>&1) -join "`n"
    $errorLine = @(&$omp prompt print error --config="$Env:POSH_THEME" 2>&1) -join "`n"
    Set-PSReadLineOption -PromptText $validLine, $errorLine
}

<#
 .SYNOPSIS
     Returns an ansi formatted hyperlink
     if name not set, uri is used as the name of the hyperlink
 .EXAMPLE
     Get-Hyperlink
 #>
function global:Get-Hyperlink {
    param(
        [Parameter(Mandatory, ValuefromPipeline = $True)]
        [string]$uri,
        [Parameter(ValuefromPipeline = $True)]
        [string]$name
    )
    $esc = [char]27
    if ("" -eq $name) {
        $name = $uri
    }
    if ($null -ne $env:WSL_DISTRO_NAME) {
        # wsl conversion if needed
        $uri = &wslpath -m $uri
    }
    return "$esc]8;;file://$uri$esc\$name$esc]8;;$esc\"
}

function global:Get-PoshThemes() {
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
        $themes | Select-Object @{ Name = 'hyperlink'; Expression = { Get-Hyperlink -uri $_.fullname } } | Format-Table -HideTableHeaders
    }
    else {
        $omp = "::OMP::"
        $themes | ForEach-Object -Process {
            Write-Host "Theme: $(Get-Hyperlink -uri $_.fullname -name $_.BaseName.Replace('.omp', ''))"
            Write-Host ""
            & $omp prompt print primary --config $($_.FullName) --pwd $PWD --shell pwsh
            Write-Host ""
            Write-Host ""
        }
    }
    Write-Host ""
    Write-Host "Themes location: $(Get-Hyperlink -uri "$Path")"
    Write-Host ""
    Write-Host "To change your theme, adjust the init script in $PROFILE."
    Write-Host "Example:"
    Write-Host "  oh-my-posh prompt init pwsh --config $Path/jandedobbeleer.omp.json | Invoke-Expression"
    Write-Host ""
}
