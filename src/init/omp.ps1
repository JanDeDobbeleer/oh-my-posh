$global:PoshSettings = New-Object -TypeName PSObject -Property @{
    Theme = "";
}

if (Test-Path "::CONFIG::") {
    $global:PoshSettings.Theme = Resolve-Path -Path "::CONFIG::"
}
function Set-PoshContext {}

function Set-GitStatus {
    if (Get-Command -Name "Get-GitStatus" -ErrorAction SilentlyContinue) {
        [Diagnostics.CodeAnalysis.SuppressMessageAttribute('PSProvideCommentHelp', '', Justification='Variable used later(not in this scope)')]
        $Global:GitStatus = Get-GitStatus
    }
}

[ScriptBlock]$Prompt = {
    #store if the last command was successfull
    $lastCommandSuccess = $?
    #store the last exit code for restore
    $realLASTEXITCODE = $global:LASTEXITCODE
    $errorCode = 0
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

    $executionTime = -1
    $history = Get-History -ErrorAction Ignore -Count 1
    if ($null -ne $history -and $null -ne $history.EndExecutionTime -and $null -ne $history.StartExecutionTime) {
        $executionTime = ($history.EndExecutionTime - $history.StartExecutionTime).TotalMilliseconds
    }

    $startInfo = New-Object System.Diagnostics.ProcessStartInfo
    $startInfo.FileName = "::OMP::"
    $config = $global:PoshSettings.Theme
    $cleanPWD = $PWD.ProviderPath.TrimEnd("\")
    $startInfo.Arguments = "--config=""$config"" --error=$errorCode --pwd=""$cleanPWD"" --execution-time=$executionTime"
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
    Set-GitStatus
    $global:LASTEXITCODE = $realLASTEXITCODE
    #remove temp variables
    Remove-Variable realLASTEXITCODE -Confirm:$false
    Remove-Variable lastCommandSuccess -Confirm:$false
}
Set-Item -Path Function:prompt -Value $Prompt -Force
