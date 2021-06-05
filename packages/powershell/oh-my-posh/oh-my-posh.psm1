<#
        .SYNOPSIS
        Generates the prompt before each line in the console
#>

# Powershell doesn't default to UTF8 just yet, so we're forcing it as there are too many problems
# that pop up when we don't
[console]::InputEncoding = [console]::OutputEncoding = New-Object System.Text.UTF8Encoding

function Get-PoshCommand {
    if ($IsMacOS) {
        return "$PSScriptRoot/bin/posh-darwin-amd64"
    }
    if ($IsLinux) {
        # this is rather hacky but there's no other way for the time being
        $arch = uname -m
        if (($arch -eq 'aarch64') -or ($arch -eq 'armv7l')) {
            return "$PSScriptRoot/bin/posh-linux-arm"
        }
        return "$PSScriptRoot/bin/posh-linux-amd64"
    }
    if ([Environment]::Is64BitOperatingSystem) {
        return "$PSScriptRoot/bin/posh-windows-amd64.exe"
    }
    return "$PSScriptRoot/bin/posh-windows-386.exe"
}

function Set-ExecutablePermissions {
    # Set the right binary to executable before doing anything else
    # Permissions don't need to be set on Windows
    if ($PSVersionTable.PSEdition -ne "Core" -or $IsWindows) {
        return
    }

    $executable = Get-PoshCommand
    if (-Not (Test-Path $executable)) {
        # This should only happen with a corrupt installation
        Write-Warning "Executable at $executable was not found"
        return
    }

    chmod a+x $executable 2>&1
}

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
    $esc = [char]27
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
        $themes | Select-Object fullname | Format-Table -HideTableHeaders
    }
    else {
        $poshCommand = Get-PoshCommand
        $themes | ForEach-Object -Process {
            Write-Host "Theme: $esc[1m$($_.BaseName.Replace('.omp', ''))$esc[0m"
            Write-Host ""
            & $poshCommand -config $($_.FullName) -pwd $PWD
            Write-Host ""
        }
    }
    Write-Host ("-" * $consoleWidth)
    Write-Host ""
    Write-Host "Themes location: $PSScriptRoot\themes"
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

Set-ExecutablePermissions

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
