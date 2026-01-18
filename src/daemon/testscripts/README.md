# Daemon Manual Test Scripts

Manual test scripts for verifying daemon mode functionality across shells.

## Quick Start

### Zsh (run on host)

```bash
cd src
go build -o /tmp/omp-test .
zsh daemon/testscripts/test_daemon_basic.zsh
zsh daemon/testscripts/test_daemon_advanced.zsh
zsh daemon/testscripts/test_zsh_integration.zsh
```

### Fish & PowerShell (run in devcontainer)

Build the devcontainer image first:

```bash
# From the oh-my-posh root directory:
docker build -t omp-dev .devcontainer/
```

Then run tests:

```bash
# Fish tests
docker run --rm -v "$(pwd):/workspace" -w /workspace/src omp-dev bash -c '
    go build -o /tmp/omp-test . 2>/dev/null
    OMP_EXECUTABLE=/tmp/omp-test fish daemon/testscripts/test_daemon_basic.fish
    OMP_EXECUTABLE=/tmp/omp-test fish daemon/testscripts/test_fish_integration.fish
'

# PowerShell tests
docker run --rm -v "$(pwd):/workspace" -w /workspace/src omp-dev bash -c '
    go build -o /tmp/omp-test . 2>/dev/null
    OMP_EXECUTABLE=/tmp/omp-test pwsh daemon/testscripts/test_daemon_basic.ps1
    OMP_EXECUTABLE=/tmp/omp-test pwsh daemon/testscripts/test_pwsh_integration.ps1
'
```

## Prerequisites

1. **Go installed** - to build the binary
2. **Shell installed** - Zsh (host), Fish/PowerShell (devcontainer has both)
3. **No daemon running** - tests manage their own daemon lifecycle
4. **Docker** - for Fish/PowerShell tests

To check/stop an existing daemon:
```bash
/tmp/omp-test daemon status
pkill -f "daemon start --foreground"  # stop if running
```

## Test Scripts

### `test_daemon_basic.zsh`

Basic daemon functionality:
- Daemon start/stop lifecycle
- Status reporting
- Render command output format
- Multiple sequential renders

```bash
cd src
zsh daemon/testscripts/test_daemon_basic.zsh
```

### `test_daemon_advanced.zsh`

Advanced daemon tests:
- Graceful failure when daemon not running
- Concurrent request handling (5 parallel requests)
- Rapid sequential requests (10 in quick succession)
- Daemon stability under load

```bash
cd src
zsh daemon/testscripts/test_daemon_advanced.zsh
```

### `test_zsh_integration.zsh`

Zsh shell integration:
- `_omp_daemon_parse_line()` output parsing
- PS1/RPROMPT variable assignment
- Precmd parameter passing
- Exit code handling
- Terminal width parameter

```bash
cd src
zsh daemon/testscripts/test_zsh_integration.zsh
```

## Using a Custom Binary Path

By default, scripts look for `/tmp/omp-test`. To use a different path:

```bash
OMP_EXECUTABLE=/path/to/your/omp zsh daemon/testscripts/test_daemon_basic.zsh
```

## Expected Output

Successful tests show:
- `PASS:` messages for each test case
- `=== All ... tests passed ===` at the end
- Exit code 0

Example output:

```text
=== Daemon Basic Tests ===
Using executable: /tmp/omp-test

=== Test 1: Verify daemon not running ===
daemon is not running

=== Test 2: Start daemon ===
daemon started
Exit code: 0
...
=== All basic tests complete ===
```

**Note:** Render output shows `primary:daemon mode not yet implemented` because the actual
prompt rendering isn't wired up yet. The IPC layer and shell integration are functional.

## Troubleshooting

### "executable not found"
```bash
# Build the binary first
cd src
go build -o /tmp/omp-test .
```

### Daemon won't start
```bash
# Check for stale lock file
ls -la ~/.local/state/oh-my-posh/daemon.lock

# Check PID in lock file
cat ~/.local/state/oh-my-posh/daemon.lock

# Remove stale lock (only if PID is dead)
rm ~/.local/state/oh-my-posh/daemon.lock
```

### Socket connection errors
```bash
# Check socket location
ls -la ${XDG_RUNTIME_DIR:-/tmp}/oh-my-posh-$(id -u).sock

# Or check fallback location
ls -la ~/.cache/oh-my-posh/
```

### Tests fail intermittently
Increase sleep times in the scripts between daemon start and first request.

## Fish Test Scripts

### `test_daemon_basic.fish`

Basic daemon functionality (Fish version):
- Daemon start/stop lifecycle
- Status reporting
- Render command output format
- Multiple sequential renders

### `test_fish_integration.fish`

Fish shell integration:
- Output parsing functions
- Environment variable handling
- Terminal width parameter
- Rapid sequential requests

## PowerShell Test Scripts

### `test_daemon_basic.ps1`

Basic daemon functionality (PowerShell version):
- Daemon start/stop lifecycle
- Status reporting
- Render command output format
- Multiple sequential renders

### `test_pwsh_integration.ps1`

PowerShell shell integration:
- Output parsing functions
- Exit code handling
- Environment variable handling
- Terminal width parameter
- Rapid sequential requests
