#!/bin/zsh
# Zsh shell integration tests
# Tests the shell script functions that will be used in omp.zsh
# Run: zsh test_zsh_integration.zsh

emulate -L zsh

# Configuration - adjust path if needed
_omp_executable=${OMP_EXECUTABLE:-/tmp/omp-test}

if [[ ! -x $_omp_executable ]]; then
    echo "Error: oh-my-posh executable not found at $_omp_executable"
    echo "Build it first: cd src && go build -o /tmp/omp-test ."
    exit 1
fi

echo "=== Zsh Shell Integration Tests ==="
echo "Using executable: $_omp_executable"
echo ""

# Clean up and start daemon
pkill -f "daemon start --foreground" 2>/dev/null || true
sleep 0.3
$_omp_executable daemon start 2>/dev/null
sleep 0.5

echo "=== Test 1: _omp_daemon_parse_line function ==="

# Define the parse function from omp.zsh
function _omp_daemon_parse_line() {
    local line=$1
    local type=${line%%:*}
    local text=${line#*:}

    case $type in
        primary)
            PS1=$text
            echo "  Parsed primary: $text"
            ;;
        right)
            RPROMPT=$text
            echo "  Parsed right: $text"
            ;;
        secondary)
            PS2=$text
            echo "  Parsed secondary: $text"
            ;;
        status)
            echo "  Parsed status: $text"
            ;;
        *)
            echo "  Unknown type: $type"
            ;;
    esac
}

# Test parsing render output
echo "Parsing render output..."
PS1=""
RPROMPT=""
$_omp_executable render --shell=zsh --pwd=/tmp 2>&1 | while read -r line; do
    _omp_daemon_parse_line "$line"
done

echo ""
echo "Final PS1: '$PS1'"
echo "Final RPROMPT: '$RPROMPT'"

if [[ -z $PS1 ]]; then
    echo "FAIL: PS1 was not set"
    exit 1
fi
echo "PASS: Parse function works"
echo ""

echo "=== Test 2: Precmd parameters simulation ==="

# Simulate what _omp_daemon_precmd does (without zle)
_omp_status=0
_omp_pipestatus=(0)
_omp_job_count=0
_omp_stack_count=0
_omp_execution_time=100
_omp_no_status=false

echo "Calling render with precmd parameters..."
output=$($_omp_executable render \
    --shell=zsh \
    --shell-version=$ZSH_VERSION \
    --status=$_omp_status \
    --pipestatus="${_omp_pipestatus[*]}" \
    --no-status=$_omp_no_status \
    --execution-time=$_omp_execution_time \
    --job-count=$_omp_job_count \
    --stack-count=$_omp_stack_count \
    --terminal-width=120 \
    --pwd=$PWD 2>&1)

echo "$output"
if echo "$output" | grep -q "primary:"; then
    echo "PASS: Precmd parameters accepted"
else
    echo "FAIL: No primary prompt in output"
    exit 1
fi
echo ""

echo "=== Test 3: Different exit codes ==="
for code in 0 1 127 130; do
    echo -n "Exit code $code: "
    if $_omp_executable render --shell=zsh --pwd=/tmp --status=$code 2>&1 | grep -q "primary:"; then
        echo "PASS"
    else
        echo "FAIL"
        exit 1
    fi
done
echo ""

echo "=== Test 4: Environment variables ==="
export POSH_SESSION_ID="env-test-$$"
output=$($_omp_executable render --shell=zsh --pwd=/tmp 2>&1)
if echo "$output" | grep -q "status:complete"; then
    echo "PASS: Session ID from environment works"
else
    echo "FAIL: Render did not complete"
    exit 1
fi
echo ""

echo "=== Test 5: Terminal width parameter ==="
for width in 40 80 120 200; do
    echo -n "Width $width: "
    if $_omp_executable render --shell=zsh --pwd=/tmp --terminal-width=$width 2>&1 | grep -q "status:complete"; then
        echo "PASS"
    else
        echo "FAIL"
        exit 1
    fi
done
echo ""

echo "=== Cleanup ==="
pkill -f "daemon start --foreground" 2>/dev/null || true
echo "Done"
echo ""

echo "=== All zsh integration tests passed ==="
