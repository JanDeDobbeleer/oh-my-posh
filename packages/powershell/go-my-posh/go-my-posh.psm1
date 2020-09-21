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
if ($IsWindows) {
    # When this is not set, outputted fonts aren't rendered correctly in some terminals for the prompt function
    [console]::OutputEncoding = New-Object System.Text.UTF8Encoding
    # Not running it beforehand in the terminal will fail the prompt somehow
    $poshCommand = Get-PoshCommand
    & $poshCommand | Out-Null
}

function Set-PoshPrompt {
    param(
        [Parameter(Mandatory = $false)]
        [string]
        $Theme
    )

    if (Test-Path "$PSScriptRoot/Themes/$Theme.json") {
        $global:PoshSettings.Theme = "$PSScriptRoot/Themes/$Theme.json"
    }
    elseif (Test-Path $Theme) {
        $global:PoshSettings.Theme = $Theme
    }
    else {
        $global:PoshSettings.Theme = "$PSScriptRoot/Themes/jandedobbeleer.json"
    }

    [ScriptBlock]$Prompt = {
        $realLASTEXITCODE = $global:LASTEXITCODE
        $poshCommand = Get-PoshCommand
        $config = $global:PoshSettings.Theme
        & $poshCommand -config $config -error $realLASTEXITCODE
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
        & $poshCommand -config $($_.FullName)
        Write-Host ""
    }
    Write-Host ("=" * $consoleWidth)
}

function Write-PoshTheme {
    $config = $global:PoshSettings.Theme
    $poshCommand = Get-PoshCommand
    & $poshCommand -config $config -print-config
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
