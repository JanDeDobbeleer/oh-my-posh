#!/usr/bin/env fish
# Basic daemon functionality tests for Fish
# Run: fish test_daemon_basic.fish

# Configuration - adjust path if needed
set -q OMP_EXECUTABLE; or set -g OMP_EXECUTABLE /home/vscode/bin/oh-my-posh

if not test -x $OMP_EXECUTABLE
    echo "Error: oh-my-posh executable not found at $OMP_EXECUTABLE"
    echo "Build it first: cd src && go build -o /home/vscode/bin/oh-my-posh ."
    exit 1
end

# Set up test environment
set -gx POSH_SESSION_ID "test-fish-basic-$fish_pid"

echo "=== Daemon Basic Tests (Fish) ==="
echo "Using executable: $OMP_EXECUTABLE"
echo ""

# Clean up any existing daemon
pkill -f "daemon start --foreground" 2>/dev/null; or true
sleep 0.3

echo "=== Test 1: Verify daemon not running ==="
$OMP_EXECUTABLE daemon status
echo ""

echo "=== Test 2: Start daemon ==="
$OMP_EXECUTABLE daemon start 2>/dev/null
echo "Exit code: $status"
sleep 0.5
echo ""

echo "=== Test 3: Verify daemon running ==="
$OMP_EXECUTABLE daemon status
echo ""

echo "=== Test 4: Test render command ==="
$OMP_EXECUTABLE render \
    --shell=fish \
    --shell-version=$FISH_VERSION \
    --status=0 \
    --pipestatus="0" \
    --no-status=false \
    --execution-time=100 \
    --job-count=0 \
    --stack-count=0 \
    --terminal-width=80 \
    --pwd=/tmp
echo ""

echo "=== Test 5: Multiple sequential renders ==="
for i in 1 2 3
    echo "--- Render $i ---"
    $OMP_EXECUTABLE render --shell=fish --pwd=/tmp --status=$i 2>&1 | head -5
end
echo ""

echo "=== Test 6: Cleanup ==="
pkill -f "daemon start --foreground" 2>/dev/null; or true
sleep 0.3
$OMP_EXECUTABLE daemon status
echo ""

echo "=== All basic tests complete (Fish) ==="
