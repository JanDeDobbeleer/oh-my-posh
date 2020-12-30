$global:PoshSettings = New-Object -TypeName PSObject -Property @{
    Theme = "";
}

$config = "::CONFIG::"
if (Test-Path $config) {
    $global:PoshSettings.Theme = (Resolve-Path -Path $config).Path
}

function global:Set-PoshContext {}

function global:Set-PoshGitStatus {
    if (Get-Module -Name "posh-git") {
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
    # Save current encoding and swap for UTF8
    $originalOutputEncoding = [Console]::OutputEncoding
    [Console]::OutputEncoding = [System.Text.Encoding]::UTF8
    $omp = "::OMP::"
    $config = $global:PoshSettings.Theme
    $cleanPWD = $PWD.ProviderPath.TrimEnd("\")
    $cleanPSWD = $PWD.ToString().TrimEnd("\")
    $standardOut = @(&$omp --error="$errorCode" --pwd="$cleanPWD" --pswd="$cleanPSWD" --execution-time="$executionTime" --config="$config" 2>&1)
    # Restore initial encoding
    [Console]::OutputEncoding = $originalOutputEncoding
    # the ouput can be multiline, joining these ensures proper rendering by adding line breaks with `n
    $standardOut -join "`n"
    Set-PoshGitStatus
    $global:LASTEXITCODE = $realLASTEXITCODE
    #remove temp variables
    Remove-Variable realLASTEXITCODE -Confirm:$false
    Remove-Variable lastCommandSuccess -Confirm:$false
}
Set-Item -Path Function:prompt -Value $Prompt -Force
