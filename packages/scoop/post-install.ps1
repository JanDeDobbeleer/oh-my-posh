if (!(Test-Path $PROFILE)) {
    $profileDir = Split-Path $PROFILE
    if (!(Test-Path $profileDir)) {
        New-Item -Path $profileDir -ItemType Directory | Out-Null
    }
    '' > $PROFILE
}

function Write-ExitIfNeeded {
    param (
        [parameter(Mandatory = $true)]
        [int]
        $Decision
    )
    if ($Decision -ne 0) {
        Write-Host 'Thanks for installing Oh my Posh.'
        Write-Host 'Have a look at https://ohmyposh.dev/docs/installation for instructions.'
        exit 0
    }
}

function Set-Prompt {
    param (
        [parameter(Mandatory = $true)]
        [string]
        $ProfilePath
    )

    $promptOverride = @'
function Get-PoshCommand {
    $poshCommand = "posh-windows-amd64.exe"
    if ($IsLinux) {
        $poshCommand = "posh-linux-amd64"
    }
    return $poshCommand
}

[ScriptBlock]$Prompt = {
    $realLASTEXITCODE = $global:LASTEXITCODE
    if ($realLASTEXITCODE -isnot [int]) {
        $realLASTEXITCODE = 0
    }
    $startInfo = New-Object System.Diagnostics.ProcessStartInfo
    $startInfo.FileName = Get-PoshCommand
    $startInfo.Arguments = "-pwd ""$PWD"" -error $realLASTEXITCODE"
    $startInfo.Environment["TERM"] = "xterm-256color"
    $startInfo.CreateNoWindow = $true
    $startInfo.StandardOutputEncoding = [System.Text.Encoding]::UTF8
    $startInfo.RedirectStandardOutput = $true
    $startInfo.UseShellExecute = $false
    if ($PWD.Provider.Name -eq "FileSystem") {
        $startInfo.WorkingDirectory = "$PWD"
    }
    $process = New-Object System.Diagnostics.Process
    $process.StartInfo = $startInfo
    Set-PoshContext
    $process.Start() | Out-Null
    $standardOut = $process.StandardOutput.ReadToEnd()
    $process.WaitForExit()
    $standardOut
    $global:LASTEXITCODE = $realLASTEXITCODE
    Remove-Variable realLASTEXITCODE -Confirm:$false
}
Set-Item -Path Function:prompt -Value $Prompt -Force
'@
    Add-Content -Path $ProfilePath -Value $promptOverride
    Write-Host 'Thanks for installing Oh my Posh.'
    Write-Host 'Have a look at the configuration posibilities at https://ohmyposh.dev'
}

if (-not (Test-Path $PROFILE)) {
    Write-Host "The Powershell profile can't be found, have a look at https://ohmyposh.dev/docs/installation for instructions"
    exit 0
}

$title = @'
   __  _____ _      ___  ___       ______         _      __
  / / |  _  | |     |  \/  |       | ___ \       | |     \ \
 / /  | | | | |__   | .  . |_   _  | |_/ /__  ___| |__    \ \
< <   | | | | '_ \  | |\/| | | | | |  __/ _ \/ __| '_ \    > >
 \ \  \ \_/ / | | | | |  | | |_| | | | | (_) \__ \ | | |  / /
  \_\  \___/|_| |_| \_|  |_/\__, | \_|  \___/|___/_| |_| /_/
                             __/ |
                            |___/
'@
$choices = '&Yes', '&No'
$question = "Do you want to add Oh my Posh to $PROFILE ?"
$decision = $Host.UI.PromptForChoice($title, $question, $choices, 1)
Write-ExitIfNeeded -Decision $decision
if (!(Get-Content $PROFILE)) {
    Set-Prompt -ProfilePath $PROFILE
    exit 0
}
$profileContent = (Get-Content $PROFILE).ToLower()
if ($profileContent -match 'function:prompt' -or $profileContent -match 'function prompt') {
    $title = "$ProfilePath already contains a prompt function override."
    $question = "Do you want to override it with Oh my Posh?"
    $decision = $Host.UI.PromptForChoice($title, $question, $choices, 1)
    Write-ExitIfNeeded -Decision $decision
}
elseif ($profileContent -match 'oh-my-posh') {
    $title = "$ProfilePath already contains an Oh my Posh import statement."
    $question = "Do you want to override it?"
    $decision = $Host.UI.PromptForChoice($title, $question, $choices, 1)
    Write-ExitIfNeeded -Decision $decision
}
Set-Prompt -ProfilePath $PROFILE
