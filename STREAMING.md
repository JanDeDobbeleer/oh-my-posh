# Streaming Prompt Rendering (Experimental)

## Overview

The `stream` command provides a foundation for asynchronous prompt rendering in shells that support concurrent execution. This feature is currently **experimental** and can be integrated with NuShell's background job system.

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

NuShell has experimental support for thread-based background jobs (available since version 0.80+), which can be used for async streaming. The integration uses:

- `job spawn` - Start the stream command as a background job
- `job send` / `job recv` - Communicate between the main thread and stream job
- `job list` / `job kill` - Manage job lifecycle

### Current Implementation Status

The stream command infrastructure is complete and ready for integration. A full NuShell integration would:

1. Spawn the stream command as a background job on shell startup
2. Send JSON requests via `job send` when prompt is needed
3. Poll for responses with `job recv` 
4. Update prompts as responses arrive
5. Handle job failures with fallback to synchronous rendering

### Example Integration Pattern

```nu
# Initialize streaming job
let stream_job = job spawn { ^oh-my-posh stream }

# Send request
{"id": "uuid", "flags": {...}} | to json | job send $stream_job

# Receive response (non-blocking with timeout)
let response = job recv --timeout 100ms
```

### Limitations

1. **Experimental Feature**: NuShell's job system is marked as experimental
2. **Job Lifetime**: Jobs terminate when the shell exits (no `disown` equivalent)
3. **Version Requirement**: Requires NuShell 0.80+ for job support

### Future Work

- Implement complete NuShell integration script using job system
- Add timeout handling for partial prompt updates
- Test across different NuShell versions
- Add configuration option to enable/disable streaming mode

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
