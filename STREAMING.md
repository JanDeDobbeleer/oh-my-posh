# Streaming Prompt Rendering (Experimental)

## Overview

The `stream` command provides a foundation for asynchronous prompt rendering in shells that support concurrent execution. This feature is currently **experimental** and provides infrastructure for future async implementations.

## Command Usage

```bash
oh-my-posh stream
```

The stream command runs oh-my-posh in server mode, reading JSON requests from stdin and writing JSON responses to stdout.

## Protocol

### Request Format

```json
{
  "id": "unique-request-id",
  "flags": {
    "config": "/path/to/config.json",
    "shell": "nu",
    "shell_version": "0.80.0",
    "pwd": "/current/directory",
    "status": 0,
    "no_status": false,
    "execution_time": 1.5,
    "terminal_width": 120,
    "job_count": 0,
    "cleared": false
  }
}
```

### Response Format

```json
{
  "id": "unique-request-id",
  "type": "complete",
  "prompts": {
    "primary": "rendered primary prompt",
    "right": "rendered right prompt"
  }
}
```

## Current Status

### Implemented
- ✅ Stream command with JSON protocol
- ✅ Request ID isolation
- ✅ Basic synchronous rendering
- ✅ Error handling and fallback

### Not Yet Implemented
- ❌ Partial prompt updates (100ms timeout)
- ❌ Progressive segment rendering
- ❌ Shell-side async integration for NuShell
- ❌ Background job management

## NuShell Integration

Currently, NuShell does not support the background job system required for async streaming as described in the original specification. The stream command exists as infrastructure for when such capabilities become available.

### Limitations

1. **No Background Jobs**: NuShell doesn't have traditional job control like Bash/Zsh
2. **No Async Closure Support**: Closure execution is synchronous
3. **No Process Communication**: Limited IPC between NuShell and background processes

### Future Work

When NuShell adds support for:
- Background task execution
- Async closures or promises
- Better process communication

The streaming feature can be fully integrated.

## Testing

Run tests with:

```bash
cd src
go test ./cli -run TestStream -v
```

## For Developers

The stream command is marked as hidden since it's for internal use only. It serves as:

1. **Foundation**: Infrastructure for future async implementations
2. **Protocol**: JSON-based request/response format
3. **Testing**: Unit tests validate the marshaling and structure

To extend this feature:

1. Implement partial rendering in `processRequest()`
2. Add progressive updates with channels
3. Integrate with shell-side async when available
4. Add more comprehensive tests

## See Also

- [NuShell Documentation](https://www.nushell.sh/book/)
