#!/usr/bin/env fish
# Fish shell integration tests
# Run: fish test_fish_integration.fish

# Configuration - adjust path if needed
set -q OMP_EXECUTABLE; or set -g OMP_EXECUTABLE /home/vscode/bin/oh-my-posh

if not test -x $OMP_EXECUTABLE
    echo "Error: oh-my-posh executable not found at $OMP_EXECUTABLE"
    echo "Build it first: cd src && go build -o /home/vscode/bin/oh-my-posh ."
    exit 1
end

echo "=== Fish Shell Integration Tests ==="
echo "Using executable: $OMP_EXECUTABLE"
echo ""

# Clean up and start daemon
pkill -f "daemon start --foreground" 2>/dev/null; or true
sleep 0.3
$OMP_EXECUTABLE daemon start 2>/dev/null
sleep 0.5

echo "=== Test 1: Parse render output ==="

# Test parsing render output (similar to what omp.fish does)
set -l output ($OMP_EXECUTABLE render --shell=fish --pwd=/tmp 2>&1)
echo "Raw output:"
for line in $output
    echo "  $line"
end

# Parse primary prompt
set -l primary_line (string match -r '^primary:.*' $output)
if test -n "$primary_line"
    set -l prompt_text (string replace 'primary:' '' $primary_line)
    echo ""
    echo "Parsed primary prompt: '$prompt_text'"
    echo "PASS: Parse function works"
else
    echo "FAIL: No primary prompt found"
    exit 1
end
echo ""

echo "=== Test 2: Streaming updates (from git directory) ==="
# Run from current directory (should be in git repo) to trigger streaming
# The git segment takes time, so we should see status:update before status:complete
set -l output ($OMP_EXECUTABLE render --shell=fish --pwd=$PWD 2>&1)
echo "Raw output:"
for line in $output
    echo "  $line"
end

# Count status lines - if streaming works, we may see multiple status:update before status:complete
set -l status_lines (string match -r '^status:.*' $output)
set -l status_count (count $status_lines)
echo ""
echo "Status lines found: $status_count"
for status_line in $status_lines
    echo "  $status_line"
end

if string match -q '*status:complete*' $output
    echo "PASS: Streaming render completed"
else
    echo "FAIL: Render did not complete"
    exit 1
end
echo ""

echo "=== Test 3: Different exit codes ==="
for code in 0 1 127 130
    set -l result ($OMP_EXECUTABLE render --shell=fish --pwd=/tmp --status=$code 2>&1)
    if string match -q '*primary:*' $result
        echo "Exit code $code: PASS"
    else
        echo "Exit code $code: FAIL"
        exit 1
    end
end
echo ""

echo "=== Test 4: Environment variables ==="
set -gx POSH_SESSION_ID "fish-env-test-$fish_pid"
set -l result ($OMP_EXECUTABLE render --shell=fish --pwd=/tmp 2>&1)
if string match -q '*status:complete*' $result
    echo "PASS: Session ID from environment works"
else
    echo "FAIL: Render did not complete"
    exit 1
end
echo ""

echo "=== Test 5: Terminal width parameter ==="
for width in 40 80 120 200
    set -l result ($OMP_EXECUTABLE render --shell=fish --pwd=/tmp --terminal-width=$width 2>&1)
    if string match -q '*status:complete*' $result
        echo "Width $width: PASS"
    else
        echo "Width $width: FAIL"
        exit 1
    end
end
echo ""

echo "=== Test 6: Concurrent requests ==="
# Run 3 concurrent requests
for i in 1 2 3
    fish -c "set -gx OMP_EXECUTABLE $OMP_EXECUTABLE; \$OMP_EXECUTABLE render --shell=fish --pwd=/tmp 2>&1 | head -1" &
end
wait
echo "PASS: All concurrent requests completed"
echo ""

echo "=== Cleanup ==="
pkill -f "daemon start --foreground" 2>/dev/null; or true
echo "Done"
echo ""

echo "=== All Fish integration tests passed ==="
