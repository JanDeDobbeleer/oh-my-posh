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

function Set-PoshContext {}

function Set-GitStatus {
    if (Get-Command -Name "Get-GitStatus" -ErrorAction SilentlyContinue) {
        $Global:GitStatus = Get-GitStatus
    }
}

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
        $global:PoshSettings.Theme = Resolve-Path -Path $Theme
    }
    else {
        $global:PoshSettings.Theme = "$PSScriptRoot/themes/jandedobbeleer.json"
    }

    [ScriptBlock]$Prompt = {
        $realLASTEXITCODE = $global:LASTEXITCODE
        Set-PoshContext
        if ($realLASTEXITCODE -isnot [int]) {
            $realLASTEXITCODE = 0
        }
        $startInfo = New-Object System.Diagnostics.ProcessStartInfo
        $startInfo.FileName = Get-PoshCommand
        $config = $global:PoshSettings.Theme
        $cleanPWD = $PWD.ProviderPath.TrimEnd("\")
        $startInfo.Arguments = "-config=""$config"" -error=$realLASTEXITCODE -pwd=""$cleanPWD"""
        $startInfo.Environment["TERM"] = "xterm-256color"
        $startInfo.CreateNoWindow = $true
        $startInfo.StandardOutputEncoding = [System.Text.Encoding]::UTF8
        $startInfo.RedirectStandardOutput = $true
        $startInfo.UseShellExecute = $false
        if ($PWD.Provider.Name -eq 'FileSystem') {
            $startInfo.WorkingDirectory = $PWD.ProviderPath
        }
        $process = New-Object System.Diagnostics.Process
        $process.StartInfo = $startInfo
        $process.Start() | Out-Null
        $standardOut = $process.StandardOutput.ReadToEnd()
        $process.WaitForExit()
        $standardOut
        $global:LASTEXITCODE = $realLASTEXITCODE
        Remove-Variable realLASTEXITCODE -Confirm:$false
        Set-GitStatus
    }
    Set-Item -Path Function:prompt -Value $Prompt -Force
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
