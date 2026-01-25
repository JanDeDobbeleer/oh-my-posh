#!/bin/zsh
# Basic daemon functionality tests
# Run: zsh test_daemon_basic.zsh

set -e

# Configuration - adjust path if needed
_omp_executable=${OMP_EXECUTABLE:-/tmp/omp-test}

if [[ ! -x $_omp_executable ]]; then
    echo "Error: oh-my-posh executable not found at $_omp_executable"
    echo "Build it first: cd src && go build -o /tmp/omp-test ."
    exit 1
fi

# Set up test environment
export POSH_SESSION_ID="test-basic-$$"

echo "=== Daemon Basic Tests ==="
echo "Using executable: $_omp_executable"
echo ""

# Clean up any existing daemon
pkill -f "daemon start --foreground" 2>/dev/null || true
sleep 0.3

echo "=== Test 1: Verify daemon not running ==="
$_omp_executable daemon status
echo ""

echo "=== Test 2: Start daemon ==="
$_omp_executable daemon start 2>/dev/null
echo "Exit code: $?"
sleep 0.5
echo ""

echo "=== Test 3: Verify daemon running ==="
$_omp_executable daemon status
echo ""

echo "=== Test 4: Test render command ==="
$_omp_executable render \
    --shell=zsh \
    --shell-version=$ZSH_VERSION \
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
for i in 1 2 3; do
    echo "--- Render $i ---"
    $_omp_executable render --shell=zsh --pwd=/tmp --status=$i 2>&1 | head -5
done
echo ""

echo "=== Test 6: Cleanup ==="
pkill -f "daemon start --foreground" 2>/dev/null || true
sleep 0.3
$_omp_executable daemon status
echo ""

echo "=== All basic tests complete ==="
