<#
        .SYNOPSIS
        Generates the prompt before each line in the console
#>

$global:PoshSettings = New-Object -TypeName PSObject -Property @{
    Theme = "$PSScriptRoot\themes\jandedobbeleer.json"
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
    # It can also fail when we're not running in a console (and then, well, you're on your own)
    try {
        [console]::OutputEncoding = New-Object System.Text.UTF8Encoding
    }
    catch {
        Write-Host "oh-my-posh: unable to set output encoding to UTF8, fonts might be rendered incorrectly."
    }
    # Not running it beforehand in the terminal will fail the prompt somehow
    $poshCommand = Get-PoshCommand
    & $poshCommand | Out-Null
}

function Set-PoshContext {}

function Set-PoshPrompt {
    param(
        [Parameter(Mandatory = $false)]
        [string]
        $Theme
    )

    if (Test-Path "$PSScriptRoot/themes/$Theme.json") {
        $global:PoshSettings.Theme = "$PSScriptRoot/themes/$Theme.json"
    }
    elseif (Test-Path $Theme) {
        $global:PoshSettings.Theme = $Theme
    }
    else {
        $global:PoshSettings.Theme = "$PSScriptRoot/themes/jandedobbeleer.json"
    }

    [ScriptBlock]$Prompt = {
        $realLASTEXITCODE = $global:LASTEXITCODE
        $poshCommand = Get-PoshCommand
        $config = $global:PoshSettings.Theme
        Set-PoshContext
        & $poshCommand -config $config -error $realLASTEXITCODE  -pwd $PWD
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
    Get-ChildItem -Path "$PSScriptRoot\themes\*" -Include '*.json' | Sort-Object Name | ForEach-Object -Process {
        Write-Host ("=" * $consoleWidth)
        Write-Host "$esc[1m$($_.BaseName)$esc[0m"
        Write-Host ""
        & $poshCommand -config $($_.FullName) -pwd $PWD
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
    $themes = Get-ChildItem -Path "$PSScriptRoot\themes\*" -Include '*.json' | Sort-Object Name | Select-Object -Property BaseName
    $themes |
    Where-Object { $_.BaseName.ToLower().StartsWith($wordToComplete.ToLower()); } |
    Select-Object -Unique -ExpandProperty BaseName |
    ForEach-Object { New-CompletionResult -CompletionText $_ }
}

Register-ArgumentCompleter `
    -CommandName Set-PoshPrompt `
    -ParameterName Theme `
    -ScriptBlock $function:ThemeCompletion
