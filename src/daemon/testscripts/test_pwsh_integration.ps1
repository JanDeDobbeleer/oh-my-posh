#!/usr/bin/env pwsh
# PowerShell shell integration tests
# Run: pwsh test_pwsh_integration.ps1

$ErrorActionPreference = "Stop"

# Configuration - adjust path if needed
$OmpExecutable = if ($env:OMP_EXECUTABLE) { $env:OMP_EXECUTABLE } else { "/home/vscode/bin/oh-my-posh" }

if (-not (Test-Path $OmpExecutable)) {
    Write-Host "Error: oh-my-posh executable not found at $OmpExecutable"
    Write-Host "Build it first: cd src && go build -o /home/vscode/bin/oh-my-posh ."
    exit 1
}

Write-Host "=== PowerShell Shell Integration Tests ==="
Write-Host "Using executable: $OmpExecutable"
Write-Host ""

# Clean up and start daemon
try { pkill -f "daemon start --foreground" 2>$null } catch {}
Start-Sleep -Milliseconds 300
& $OmpExecutable daemon start 2>$null
Start-Sleep -Milliseconds 500

Write-Host "=== Test 1: Parse render output ==="

# Test parsing render output (similar to what omp.ps1 does)
$output = & $OmpExecutable render --shell=pwsh --pwd=/tmp 2>&1
Write-Host "Raw output:"
foreach ($line in $output) {
    Write-Host "  $line"
}

# Parse primary prompt
$primaryLine = $output | Where-Object { $_ -match '^primary:' } | Select-Object -First 1
if ($primaryLine) {
    $promptText = $primaryLine -replace '^primary:', ''
    Write-Host ""
    Write-Host "Parsed primary prompt: '$promptText'"
    Write-Host "PASS: Parse function works"
} else {
    Write-Host "FAIL: No primary prompt found"
    exit 1
}
Write-Host ""

Write-Host "=== Test 2: Streaming updates (from git directory) ==="
# Run from current directory (should be in git repo) to trigger streaming
# The git segment takes time, so we should see status:update before status:complete
$currentDir = Get-Location
$output = & $OmpExecutable render --shell=pwsh --pwd=$currentDir 2>&1
Write-Host "Raw output:"
foreach ($line in $output) {
    Write-Host "  $line"
}

# Count status lines - if streaming works, we may see multiple status:update before status:complete
$statusLines = $output | Where-Object { $_ -match '^status:' }
$statusCount = ($statusLines | Measure-Object).Count
Write-Host ""
Write-Host "Status lines found: $statusCount"
foreach ($statusLine in $statusLines) {
    Write-Host "  $statusLine"
}

if ($output -match 'status:complete') {
    Write-Host "PASS: Streaming render completed"
} else {
    Write-Host "FAIL: Render did not complete"
    exit 1
}
Write-Host ""

Write-Host "=== Test 3: Different exit codes ==="
foreach ($code in @(0, 1, 127, 130)) {
    $result = & $OmpExecutable render --shell=pwsh --pwd=/tmp --status=$code 2>&1
    if ($result -match 'primary:') {
        Write-Host "Exit code ${code}: PASS"
    } else {
        Write-Host "Exit code ${code}: FAIL"
        exit 1
    }
}
Write-Host ""

Write-Host "=== Test 4: Environment variables ==="
$env:POSH_SESSION_ID = "pwsh-env-test-$PID"
$result = & $OmpExecutable render --shell=pwsh --pwd=/tmp 2>&1
if ($result -match 'status:complete') {
    Write-Host "PASS: Session ID from environment works"
} else {
    Write-Host "FAIL: Render did not complete"
    exit 1
}
Write-Host ""

Write-Host "=== Test 5: Terminal width parameter ==="
foreach ($width in @(40, 80, 120, 200)) {
    $result = & $OmpExecutable render --shell=pwsh --pwd=/tmp --terminal-width=$width 2>&1
    if ($result -match 'status:complete') {
        Write-Host "Width ${width}: PASS"
    } else {
        Write-Host "Width ${width}: FAIL"
        exit 1
    }
}
Write-Host ""

Write-Host "=== Test 6: Rapid sequential requests ==="
$successCount = 0
foreach ($i in 1..10) {
    $result = & $OmpExecutable render --shell=pwsh --pwd=/tmp 2>&1
    if ($result -match 'status:complete') {
        $successCount++
    }
}
Write-Host "Completed: $successCount/10"
if ($successCount -eq 10) {
    Write-Host "PASS: All rapid requests completed"
} else {
    Write-Host "FAIL: Some requests did not complete"
    exit 1
}
Write-Host ""

Write-Host "=== Cleanup ==="
try { pkill -f "daemon start --foreground" 2>$null } catch {}
Write-Host "Done"
Write-Host ""

Write-Host "=== All PowerShell integration tests passed ==="
