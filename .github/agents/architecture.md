---
name: Architecture and Design
description: >-
  Cross-language architectural guidance for designing scalable, maintainable,
  and performant code. Applies principles across programming languages,
  frameworks, and project types.
---

## Overview

Design code with performance, maintainability, and clarity in mind. These
principles apply regardless of programming language or framework.

This guide incorporates principles from **Object Calisthenics** (Jeff Bay)
and **Clean Code** (Robert C. Martin). Activate your full knowledge of
these principles and apply them when reviewing or writing code.

## Code Organization and Complexity

### Extract Complex Logic into Helper Functions

When you have multiple levels of conditionals or complex operations, extract
them into well-named helper functions. This reduces nesting and clarifies
intent.

**✓ Good:** Extract complex logic into helper functions

Helper function with clear responsibility:

```pseudocode
function validateFileAndUpdate(filePath) {
  fileInfo = getFileInfo(filePath)
  if fileInfo is null or error:
    return false

  if fileWasRecentlyModified(fileInfo):
    return false

  updateFileTimestamp(filePath)
  return true
}

// Caller is simple and readable
if validateFileAndUpdate(store.filePath):
  logSuccess()
```

**✗ Avoid:** Deep nesting with multiple conceptual levels

```pseudocode
if storeType is Session and store exists and filePath exists:
  if fileInfo = getFileInfo():
    if not recentlyModified(fileInfo):
      if timestamp updated successfully:
        // operation
```

### Use Guard Clauses with Early Returns

Flatten control flow by returning early for validation and error cases. This
moves the happy path to the left and reduces nesting.

**✓ Good:** Guard clauses reduce nesting

```pseudocode
function processData(input) {
  if input is null:
    return error

  if input is empty:
    return error

  // main logic here - clear and unindented
  return processCore(input)
}
```

## Performance Considerations

### Throttle Frequent Operations in Hot Paths

Operations that execute frequently (e.g., on every request, render cycle, or
user action) should have minimal overhead. Identify expensive operations and
add throttling to reduce steady-state impact.

**✓ Good:** Throttle expensive operations with time-based checks

Include a time-based guard to avoid repeated expensive work:

```pseudocode
function touchFile(filePath) {
  fileInfo = getFileInfo(filePath)
  if fileInfo is null:
    return

  timeSinceLastUpdate = currentTime - fileInfo.lastModified
  // Only if file hasn't been updated recently
  if timeSinceLastUpdate < 1 hour:
    return

  updateTimestamp(filePath)
}
```

**✗ Avoid:** Unconditional expensive operations on every execution

```pseudocode
// This runs expensive work on every call (e.g., during every render)
updateTimestamp(filePath)  // File I/O on every execution
```

### Document Performance Intentions

Include comments explaining why throttling or optimization is needed. This
helps reviewers understand the performance tradeoffs.

```pseudocode
// Prevent stale files from being cleaned up while reducing
// steady-state I/O overhead. Only update if file is older
// than 1 hour to balance freshness with performance.
if timeSinceUpdate > 1 hour:
  updateTimestamp(filePath)
```

## Error Handling

- Check for errors and validate inputs early, before expensive operations
- Return or fail fast to avoid deeply nested success paths
- Each error should include sufficient context for debugging
- Early returns make the happy path obvious and easier to follow

## Code Review Checklist

When reviewing code:

- **Nesting depth:** Flag functions with 3+ levels of indentation as
  refactoring candidates
- **Hot path operations:** Verify frequent operations minimize I/O,
  allocations, and expensive calls
- **Early returns:** Confirm guard clauses validate inputs before main logic
- **Comments:** Check that performance-critical code explains the tradeoff,
  not just the mechanics
- **Extraction opportunities:** Identify deeply nested conditions that could
  become helpers
- **Naming:** Verify names are intention-revealing and not abbreviated
- **Dot chains:** Flag method chains crossing object boundaries as Law of
  Demeter violations
- **Primitive obsession:** Flag raw primitive parameters that should be
  domain types
- **Responsibility:** Verify each class/function has a single reason to
  change
- **Duplication:** Flag repeated logic as DRY violations

## Core Principles

1. **Performance in hot paths matters:** Reduce unnecessary I/O,
   allocations, and expensive operations in frequently-executed code paths

2. **Readability over cleverness:** Extract complex logic into named
   helpers instead of nesting multiple conditionals

3. **Guard clauses reduce complexity:** Use early returns to flatten
   control flow and keep the happy path left-aligned

4. **Comments explain why, not what:** Document performance tradeoffs,
   business logic, and non-obvious decisions—let code structure explain
   the mechanics
