<#
        .SYNOPSIS
        Generates the prompt before each line in the console
#>

# Powershell doesn't default to UTF8 just yet, so we're forcing it as there are too many problems
# that pop up when we don't
if ($ExecutionContext.SessionState.LanguageMode -ne "ConstrainedLanguage") {
    [console]::InputEncoding = [console]::OutputEncoding = New-Object System.Text.UTF8Encoding
}
elseif ($env:POSH_CONSTRAINED_LANGUAGE -ne 1) {
    Write-Host "[WARNING] ConstrainedLanguage mode detected, unable to set console to UTF-8.
When using PowerShell in ConstrainedLanguage mode, please set the
console mode manually to UTF-8. See here for more information:
https://ohmyposh.dev/docs/faq#powershell-running-in-constrainedlanguage-mode
"
    $env:POSH_CONSTRAINED_LANGUAGE = 1
}

function Get-PoshDownloadUrl {
    param(
        [Parameter(Mandatory = $true)]
        [string]
        $Version
    )

    $executable = ""
    if ($IsMacOS) {
        $executable = "posh-darwin-amd64"
    }
    elseif ($IsLinux) {
        # this is rather hacky but there's no other way for the time being
        $arch = uname -m
        if ($arch -eq 'aarch64') {
            $executable = "posh-linux-arm64"
        }
        elseif ($arch -eq 'armv7l') {
            $executable = "posh-linux-arm"
        }
        else {
            $executable = "posh-linux-amd64"
        }
    }
    else {
        $arch = (Get-CimInstance -Class Win32_Processor -Property Architecture).Architecture
        switch ($arch) {
            0 { $executable = "posh-windows-386.exe" } # x86
            5 { $executable = "posh-windows-arm64.exe" } # ARM
            9 { $executable = "posh-windows-amd64.exe" } # x64
            12 { $executable = "posh-windows-amd64.exe" } # x64 emulated on Surface Pro X
        }
    }
    if ($executable -eq "") {
        throw "oh-my-posh: Unsupported architecture: $arch"
    }
    return "https://github.com/jandedobbeleer/oh-my-posh/releases/download/v$Version/$executable"
}

function Get-PoshExecutable {
    param(
        [Parameter(Mandatory = $true)]
        [string]
        $Url,
        [Parameter(Mandatory = $true)]
        [string]
        $Destination
    )

    Invoke-WebRequest $Url -Out $Destination
    if (-Not (Test-Path $executable)) {
        # This should only happen with a corrupt installation
        throw "Executable at $executable was not found, please try importing oh-my-posh again."
    }
    # Set the right binary to executable before doing anything else
    # Permissions don't need to be set on Windows
    if ($PSVersionTable.PSEdition -ne "Core" -or $IsWindows) {
        return
    }
    chmod a+x $executable 2>&1
}

function Get-PoshCommand {
    $extension = ""
    if ($PSVersionTable.PSEdition -ne "Core" -or $IsWindows) {
        $extension = ".exe"
    }
    return "$((Get-Item $MyInvocation.MyCommand.ScriptBlock.Module.ModuleBase).Parent)/oh-my-posh$extension"
}

function Sync-PoshExecutable {
    $executable = Get-PoshCommand
    $moduleVersion = Split-Path -Leaf $MyInvocation.MyCommand.ScriptBlock.Module.ModuleBase
    if (-not (Test-Path $executable)) {
        Write-Host "Downloading oh-my-posh executable"
        $url = Get-PoshDownloadUrl -Version $moduleVersion
        Get-PoshExecutable -Url $url -Destination $executable
        return
    }
    $poshVersion = & $executable --version
    if ($poshVersion -eq $moduleVersion) {
        return
    }
    Write-Host "Updating oh-my-posh executable to $moduleVersion"
    $url = Get-PoshDownloadUrl -Version $moduleVersion
    Get-PoshExecutable -Url $url -Destination $executable
}

Sync-PoshExecutable

function Set-PoshPrompt {
    param(
        [Parameter(Mandatory = $false)]
        [string]
        $Theme
    )

    $config = ""
    if (Test-Path "$PSScriptRoot/themes/$Theme.omp.json") {
        $path = "$PSScriptRoot/themes/$Theme.omp.json"
        $config = (Resolve-Path -Path $path).ProviderPath
    }
    elseif (Test-Path $Theme) {
        $config = (Resolve-Path -Path $Theme).ProviderPath
    }
    else {
        $config = "$PSScriptRoot/themes/jandedobbeleer.omp.json"
    }

    # Workaround for get-location/push-location/pop-location from within a module
    # https://github.com/PowerShell/PowerShell/issues/12868
    # https://github.com/JanDeDobbeleer/oh-my-posh2/issues/113
    $global:omp_global_sessionstate = $PSCmdlet.SessionState

    $poshCommand = Get-PoshCommand
    (& $poshCommand --init --shell=pwsh --config="$config") | Invoke-Expression
}

<#
.SYNOPSIS
    Returns an ansi formatted hyperlink
    if name not set, uri is used as the name of the hyperlink
.EXAMPLE
    Get-Hyperlink
#>
function Get-Hyperlink {
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

<#
.SYNOPSIS
    Display a preview or a list of installed themes.
.EXAMPLE
    Get-PoshThemes
.Example
    Gest-PoshThemes -list
#>
function Get-PoshThemes() {
    param(
        [switch]
        [Parameter(Mandatory = $false, HelpMessage = "List themes path")]
        $list
    )
    $consoleWidth = $Host.UI.RawUI.WindowSize.Width
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
    $themes = Get-ChildItem -Path "$PSScriptRoot\themes\*" -Include '*.omp.json' | Sort-Object Name
    Write-Host ("-" * $consoleWidth)
    if ($list -eq $true) {
        $themes | Select-Object @{ Name = 'hyperlink'; Expression = { Get-Hyperlink -uri $_.fullname } } | Format-Table -HideTableHeaders
    }
    else {
        $poshCommand = Get-PoshCommand
        $themes | ForEach-Object -Process {
            Write-Host "Theme: $(Get-Hyperlink -uri $_.fullname -name $_.BaseName.Replace('.omp', ''))"
            Write-Host ""
            & $poshCommand -config $($_.FullName) -pwd $PWD
            Write-Host ""
        }
    }
    Write-Host ("-" * $consoleWidth)
    Write-Host ""
    Write-Host "Themes location: $(Get-Hyperlink -uri "$PSScriptRoot/themes")"
    Write-Host ""
    Write-Host "To change your theme, use the Set-PoshPrompt command. Example:"
    Write-Host "  Set-PoshPrompt -Theme jandedobbeleer"
    Write-Host ""
}

# Helper function to create argument completion results
function New-CompletionResult {
    param(
        [Parameter(Mandatory)]
        [string]$CompletionText,
        [string]$ListItemText = $CompletionText,
        [System.Management.Automation.CompletionResultType]$CompletionResultType = [System.Management.Automation.CompletionResultType]::ParameterValue,
        [string]$ToolTip = $CompletionText
    )

    New-Object System.Management.Automation.CompletionResult $CompletionText, $ListItemText, $CompletionResultType, $ToolTip
}

function ThemeCompletion {
    param(
        $commandName,
        $parameterName,
        $wordToComplete,
        $commandAst,
        $fakeBoundParameter
    )
    $themes = Get-ChildItem -Path "$PSScriptRoot\themes\*" -Include '*.omp.json' | Sort-Object Name | Select-Object -Property @{
        label      = 'BaseName'
        expression = { $_.BaseName.Replace('.omp', '') }
    }
    $themes |
    Where-Object { $_.BaseName.ToLower().StartsWith($wordToComplete.ToLower()); } |
    Select-Object -Unique -ExpandProperty BaseName |
    ForEach-Object { New-CompletionResult -CompletionText $_ }
}

Register-ArgumentCompleter `
    -CommandName Set-PoshPrompt `
    -ParameterName Theme `
    -ScriptBlock $function:ThemeCompletion


# V2 compatibility functions
# These should be removed at a certain point in time
# but to facilitate ease of transition they are kept
# as long as issues/feature requests keep popping up

function Get-PoshInfoForV2Users {
    Write-Host @'

Hi there!

It seems you're using an oh-my-posh V2 cmdlet while running V3.
To migrate your current setup to V3, have a look the documentation.

https://ohmyposh.dev/docs/upgrading

'@
}

Set-Alias -Name Set-Prompt -Value Get-PoshInfoForV2Users -Force
Set-Alias -Name Get-ThemesLocation -Value Get-PoshInfoForV2Users -Force
Set-Alias -Name Show-ThemeSymbols -Value Get-PoshInfoForV2Users -Force
Set-Alias -Name Show-ThemeColors -Value Get-PoshInfoForV2Users -Force
Set-Alias -Name Show-Colors -Value Get-PoshInfoForV2Users -Force
Set-Alias -Name Write-ColorPreview -Value Get-PoshInfoForV2Users -Force
