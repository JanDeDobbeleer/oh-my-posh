#!/bin/zsh
# Advanced daemon tests: concurrency, load, error handling
# Run: zsh test_daemon_advanced.zsh

# Configuration - adjust path if needed
_omp_executable=${OMP_EXECUTABLE:-/tmp/omp-test}

if [[ ! -x $_omp_executable ]]; then
    echo "Error: oh-my-posh executable not found at $_omp_executable"
    echo "Build it first: cd src && go build -o /tmp/omp-test ."
    exit 1
fi

echo "=== Daemon Advanced Tests ==="
echo "Using executable: $_omp_executable"
echo ""

# Clean up first
pkill -f "daemon start --foreground" 2>/dev/null || true
sleep 0.3

echo "=== Test 1: Graceful failure when daemon not running ==="
$_omp_executable daemon status
result=$($_omp_executable render --shell=zsh --pwd=/tmp 2>&1)
exit_code=$?
echo "Output: $result"
echo "Exit code: $exit_code"
if [[ $exit_code -eq 0 ]]; then
    echo "FAIL: Expected non-zero exit code"
    exit 1
fi
echo "PASS: Graceful failure"
echo ""

echo "=== Test 2: Start daemon and verify concurrent requests ==="
$_omp_executable daemon start
sleep 0.5

# Run 5 concurrent render requests
echo "Running 5 concurrent requests..."
for i in 1 2 3 4 5; do
    (
        result=$($_omp_executable render --shell=zsh --pwd=/tmp 2>&1)
        echo "Request $i: $(echo "$result" | head -1)"
    ) &
done
wait
echo "PASS: All concurrent requests completed"
echo ""

echo "=== Test 3: Rapid sequential requests (same session) ==="
export POSH_SESSION_ID="rapid-test-$$"
echo "Running 10 rapid sequential requests..."
success_count=0
for i in {1..10}; do
    if $_omp_executable render --shell=zsh --pwd=/tmp 2>&1 | grep -q "status:complete"; then
        ((success_count++))
    fi
done
echo "Completed: $success_count/10"
if [[ $success_count -lt 10 ]]; then
    echo "FAIL: Some requests did not complete"
    exit 1
fi
echo "PASS: All rapid requests completed"
echo ""

echo "=== Test 4: Check daemon stability after load ==="
$_omp_executable daemon status
if ! $_omp_executable daemon status | grep -q "running"; then
    echo "FAIL: Daemon crashed under load"
    exit 1
fi
echo "PASS: Daemon still running"
echo ""

echo "=== Test 5: Stop daemon and verify ==="
pkill -f "daemon start --foreground" 2>/dev/null || true
sleep 0.3
if $_omp_executable daemon status | grep -q "not running"; then
    echo "PASS: Daemon stopped cleanly"
else
    echo "FAIL: Daemon did not stop"
    exit 1
fi
echo ""

echo "=== All advanced tests passed ==="
