<#
        .SYNOPSIS
        Generates the prompt before each line in the console
#>

function Get-PoshCommand {
    $poshCommand = "$PSScriptRoot/bin/posh-windows-amd64.exe"
    if ($IsMacOS) {
        $poshCommand = "$PSScriptRoot/bin/posh-darwin-amd64"
    }
    if ($IsLinux) {
        $poshCommand = "$PSScriptRoot/bin/posh-linux-amd64"
    }
    return $poshCommand
}

function Set-ExecutablePermissions {
    # Set the right binary to executable before doing anything else
    # Permissions don't need to be set on Windows
    if ($PSVersionTable.PSEdition -ne "Core" -or $IsWindows) {
        return
    }

    $executable = Get-PoshCommand
    if (-Not (Test-Path $executable)) {
        # This should only happend with a corrupt installation
        Write-Warning "Executable at $executable was not found"
        return
    }

    # Check the permissions on the file
    $permissions = ((ls -l $executable) -split ' ')[0]  # $permissions will be something like '-rw-r--r--'
    if ((id -u) -eq 0) {
        # Running as root, give global executable permissions if needed
        $hasExecutable = $permissions[3] -eq 'x'
        if (-not $hasExecutable) {
            Invoke-Expression -Command "chmod g+x $executable"
        }
        return
    }
    # Running as user, give user executable permissions if needed
    $hasExecutable = $permissions[9] -eq 'x'
    if (-not $hasExecutable) {
        Invoke-Expression -Command "chmod +x $executable"
    }
}

function Set-PoshPrompt {
    param(
        [Parameter(Mandatory = $false)]
        [string]
        $Theme
    )

    $config = ""
    if (Test-Path "$PSScriptRoot/themes/$Theme.omp.json") {
        $config = "$PSScriptRoot/themes/$Theme.omp.json"
    }
    elseif (Test-Path $Theme) {
        $config = (Resolve-Path -Path $Theme).Path
    }
    else {
        $config = "$PSScriptRoot/themes/jandedobbeleer.omp.json"
    }

    $poshCommand = Get-PoshCommand
    Invoke-Expression (& $poshCommand --init --shell=pwsh --config="$config")
}

function Get-PoshThemes {
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
    $poshCommand = Get-PoshCommand
    Get-ChildItem -Path "$PSScriptRoot\themes\*" -Include '*.omp.json' | Sort-Object Name | ForEach-Object -Process {
        Write-Host ("=" * $consoleWidth)
        Write-Host "$esc[1m$($_.BaseName)$esc[0m"
        Write-Host ""
        & $poshCommand -config $($_.FullName) -pwd $PWD
        Write-Host ""
    }
    Write-Host ("=" * $consoleWidth)
}

function Export-PoshTheme {
    param(
        [Parameter(Mandatory = $true)]
        [string]
        $FilePath
    )

    $config = $global:PoshSettings.Theme
    $poshCommand = Get-PoshCommand
    # Save current encoding and swap for UTF8
    $originalOutputEncoding = [Console]::OutputEncoding
    [Console]::OutputEncoding = [System.Text.Encoding]::UTF8
    & $poshCommand -config $config -print-config | Out-File -FilePath $FilePath
    # Restore initial encoding
    [Console]::OutputEncoding = $originalOutputEncoding
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
        label='BaseName'
        expression={$_.BaseName.Replace('.omp', '')}
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
Set-Alias -Name Get-Theme -Value Get-PoshInfoForV2Users -Force
Set-Alias -Name Show-ThemeSymbols -Value Get-PoshInfoForV2Users -Force
Set-Alias -Name Show-ThemeColors -Value Get-PoshInfoForV2Users -Force
Set-Alias -Name Show-Colors -Value Get-PoshInfoForV2Users -Force
Set-Alias -Name Write-ColorPreview -Value Get-PoshInfoForV2Users -Force
