<#
        .SYNOPSIS
        Generates the prompt before each line in the console
#>

$global:PoshSettings = New-Object -TypeName PSObject -Property @{
    Theme = "$PSScriptRoot\Themes\jandedobbeleer.json"
}

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

# Set the right binary to executable before doing anything else
if (!$IsWindows) {
    $executable = Get-PoshCommand
    Invoke-Expression -Command "chmod +x $executable"
}

function Set-PoshPrompt {
    param(
        [Parameter(Mandatory = $true)]
        [string]
        $Theme
    )

    $themePath = ""
    if (Test-Path "$PSScriptRoot/Themes/$Theme.json") {
        $themePath = "$PSScriptRoot/Themes/$Theme.json"
    }
    elseif (Test-Path $Theme) {
        $themePath = $Theme
    }
    else {
        Write-Error -Message "Unable to locate theme, please verify the name and/or location"
        return
    }
    $global:PoshSettings.Theme = $themePath

    [ScriptBlock]$Prompt = {
        $realLASTEXITCODE = $global:LASTEXITCODE
        $poshCommand = Get-PoshCommand
        $config = $global:PoshSettings.Theme
        Invoke-Expression -Command "$poshCommand -config $config -error $realLASTEXITCODE"
        $global:LASTEXITCODE = $realLASTEXITCODE
        Remove-Variable realLASTEXITCODE -Confirm:$false
    }
    Set-Item -Path Function:prompt -Value $Prompt -Force
}

function Get-PoshThemes {
    $esc = [char]27
    $consoleWidth = $Host.UI.RawUI.WindowSize.Width
    $logo = @'
    __                                                  _      __
    / /                                                 | |     \ \
   / /    __ _  ___    _ __ ___  _   _   _ __   ___  ___| |__    \ \
  < <    / _` |/ _ \  | '_ ` _ \| | | | | '_ \ / _ \/ __| '_ \    > >
   \ \  | (_| | (_) | | | | | | | |_| | | |_) | (_) \__ \ | | |  / /
    \_\  \__, |\___/  |_| |_| |_|\__, | | .__/ \___/|___/_| |_| /_/
          __/ |                   __/ | | |
         |___/                   |___/  |_|

'@
    Write-Host $logo
    $poshCommand = Get-PoshCommand
    Get-ChildItem -Path "$PSScriptRoot\Themes\*" -Include '*.json' | Sort-Object Name | ForEach-Object -Process {
        Write-Host ("=" * $consoleWidth)
        Write-Host "$esc[1m$($_.BaseName)$esc[0m"
        Write-Host ""
        Invoke-Expression -Command "$poshCommand -config $($_.FullName)"
        Write-Host ""
    }
    Write-Host ("=" * $consoleWidth)
}

function Write-PoshTheme {
    $poshCommand = Get-PoshCommand
    Invoke-Expression -Command "$poshCommand -print-config"
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
    $themes = Get-ChildItem -Path "$PSScriptRoot\Themes\*" -Include '*.json' | Sort-Object Name | Select-Object -Property BaseName
    $themes |
    Where-Object { $_.BaseName.ToLower().StartsWith($wordToComplete.ToLower()); } |
    Select-Object -Unique -ExpandProperty BaseName |
    ForEach-Object { New-CompletionResult -CompletionText $_ }
}

Register-ArgumentCompleter `
    -CommandName Set-PoshPrompt `
    -ParameterName Theme `
    -ScriptBlock $function:ThemeCompletion
