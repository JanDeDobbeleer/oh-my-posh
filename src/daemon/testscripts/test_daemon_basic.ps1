#!/usr/bin/env pwsh
# Basic daemon functionality tests for PowerShell
# Run: pwsh test_daemon_basic.ps1

$ErrorActionPreference = "Stop"

# Configuration - adjust path if needed
$OmpExecutable = if ($env:OMP_EXECUTABLE) { $env:OMP_EXECUTABLE } else { "/home/vscode/bin/oh-my-posh" }

if (-not (Test-Path $OmpExecutable)) {
    Write-Host "Error: oh-my-posh executable not found at $OmpExecutable"
    Write-Host "Build it first: cd src && go build -o /home/vscode/bin/oh-my-posh ."
    exit 1
}

# Set up test environment
$env:POSH_SESSION_ID = "test-pwsh-basic-$PID"

Write-Host "=== Daemon Basic Tests (PowerShell) ==="
Write-Host "Using executable: $OmpExecutable"
Write-Host ""

# Clean up any existing daemon
try { pkill -f "daemon start --foreground" 2>$null } catch {}
Start-Sleep -Milliseconds 300

Write-Host "=== Test 1: Verify daemon not running ==="
& $OmpExecutable daemon status
Write-Host ""

Write-Host "=== Test 2: Start daemon ==="
& $OmpExecutable daemon start 2>$null
Write-Host "Exit code: $LASTEXITCODE"
Start-Sleep -Milliseconds 500
Write-Host ""

Write-Host "=== Test 3: Verify daemon running ==="
& $OmpExecutable daemon status
Write-Host ""

Write-Host "=== Test 4: Test render command ==="
& $OmpExecutable render `
    --shell=pwsh `
    --shell-version=$($PSVersionTable.PSVersion.ToString()) `
    --status=0 `
    --pipestatus="0" `
    --no-status=false `
    --execution-time=100 `
    --job-count=0 `
    --stack-count=0 `
    --terminal-width=80 `
    --pwd=/tmp
Write-Host ""

Write-Host "=== Test 5: Multiple sequential renders ==="
foreach ($i in 1..3) {
    Write-Host "--- Render $i ---"
    & $OmpExecutable render --shell=pwsh --pwd=/tmp --status=$i 2>&1 | Select-Object -First 5
}
Write-Host ""

Write-Host "=== Test 6: Cleanup ==="
try { pkill -f "daemon start --foreground" 2>$null } catch {}
Start-Sleep -Milliseconds 300
& $OmpExecutable daemon status
Write-Host ""

Write-Host "=== All basic tests complete (PowerShell) ==="
