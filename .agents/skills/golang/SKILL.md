---
name: golang
description: >
  Go coding standards and conventions for this project. Apply when writing,
  reviewing, or refactoring any Go source file.
triggers:
  - on_commit
---

# Go Development Instructions

Follow idiomatic Go practices and community standards when writing Go code.
These instructions are based on [Effective Go](https://go.dev/doc/effective_go),
[Go Code Review Comments](https://go.dev/wiki/CodeReviewComments),
and [Google's Go Style Guide](https://google.github.io/styleguide/go/).

## General Instructions

- Write simple, clear, and idiomatic Go code
- Favor clarity and simplicity over cleverness
- Follow the principle of least surprise
- Keep the happy path left-aligned (reduce indentation)
- Return early to reduce nesting
- Make the zero value useful
- Document exported types, functions, methods, and packages
- Use Go modules for dependency management
- **AVOID `else` statements - use early returns, continue, or break instead**
- Avoid wrapping primitives without a clear semantic benefit; define new types when they add meaning.
- Use typed slices/maps and document element semantics when not obvious.
- Start error strings with a lowercase letter.

## Naming Conventions

### Packages

- Use lowercase, single-word package names
- Avoid `_` characters, hyphens, or mixedCaps
- Choose names that describe what the package provides, not what it contains
- Avoid generic names like `util`, `common`, or `base`
- Package names should be singular, not plural

### Variables and Functions

- Use mixedCaps or MixedCaps (camelCase) rather than `_` characters
- Keep names short but descriptive
- Use single-letter variables for very short scopes (like loop indices)
- Exported names start with a capital letter
- Unexported names start with a lowercase letter
- Avoid stuttering (e.g., avoid `http.HTTPServer`, prefer `http.Server`)

### Interfaces

- Name interfaces with -er suffix when possible (e.g., `Reader`, `Writer`, `Formatter`)
- Single-method interfaces should be named after the method (e.g., `Read` → `Reader`)
- Keep interfaces small and focused

### Constants

- Use MixedCaps for exported constants
- Use mixedCaps for unexported constants
- Group related constants using `const` blocks
- Consider using typed constants for better type safety

## Code Style and Formatting

### Formatting

- Always use `gofmt` to format code
- Use `goimports` to manage imports automatically
- Keep line to 180 max at all times
- Add blank lines to separate logical groups of code

### Comments

- Write comments in complete sentences
- Start sentences with the name of the thing being described
- Package comments should start with "Package [name]"
- Use line comments (`//`) for most comments
- Use block comments (`/* */`) sparingly, mainly for package documentation
- Document why, not what, unless the what is complex

### Error Handling

- Check errors immediately after the function call
- Don't ignore errors using `_` unless you have a good reason (document why)
- Wrap errors with context using `fmt.Errorf` with `%w` verb
- Create custom error types when you need to check for specific errors
- Place error returns as the last return value
- Name error variables `err`
- Keep error messages lowercase and don't end with punctuation

### Logging

- Always use the codebase `log` package for logging
- Log errors at the point they occur using `log.Error(err)`
- Do not format the errors, let the `log` package handle it
- For complex function calls, use `defer log.Trace(time.Now(), args)`
    where args are the function arguments at the start of the function.

### Control Flow

- **NEVER use `else` statements** - they create unnecessary nesting and reduce readability
- Use early returns to handle error cases and edge conditions first
- Use `continue` in loops to skip to the next iteration instead of nesting
- Use `break` to exit loops early instead of complex conditional logic
- Keep the main logic (happy path) left-aligned with minimal indentation

**❌ BAD - Don't do this:**

```go
func processEntry(entry *Entry) string {
    if entry.Expired() {
        return "expired"
    } else {
        if entry.TTL < 0 {
            return "never expires"
        } else {
            return fmt.Sprintf("expires at %s", time.Unix(entry.Timestamp, 0))
        }
    }
}
```

**✅ GOOD - Do this instead:**

```go
func processEntry(entry *Entry) string {
    if entry.Expired() {
        return "expired"
    }

    if entry.TTL < 0 {
        return "never expires"
    }

    return fmt.Sprintf("expires at %s", time.Unix(entry.Timestamp, 0))
}
```

**❌ BAD - Nested loop logic:**

```go
for _, item := range items {
    if item.IsValid() {
        if item.ShouldProcess() {
            // complex processing logic
        }
    }
}
```

**✅ GOOD - Early continue:**

```go
for _, item := range items {
    if !item.IsValid() {
        continue
    }
    if !item.ShouldProcess() {
        continue
    }

    // complex processing logic (happy path)
}
```

## Architecture and Project Structure

### Package Organization

- Follow standard Go project layout conventions
- Group related functionality into packages
- Avoid circular dependencies

### Dependency Management

- Use Go modules (`go.mod` and `go.sum`)
- Keep dependencies minimal
- Regularly update dependencies for security patches
- Use `go mod tidy` to clean up unused dependencies
- Vendor dependencies when necessary

## Type Safety and Language Features

### Type Definitions

- Define types to add meaning and type safety
- Use struct tags for JSON, YAML and TOML on exported fields
- Prefer explicit type conversions
- Use type assertions carefully and check the second return value

### Pointers vs Values

- Use pointers for large structs or when you need to modify the receiver
- Use values for small structs and when immutability is desired
- Be consistent within a type's method set
- Consider the zero value when choosing pointer vs value receivers

### Interfaces and Composition

- Accept interfaces, return concrete types
- Keep interfaces small (1-3 methods is ideal)
- Use embedding for composition
- Define interfaces close to where they're used, not where they're implemented
- Don't export interfaces unless necessary

## Concurrency

### Goroutines

- Don't create goroutines in libraries; let the caller control concurrency
- Always know how a goroutine will exit
- Use `sync.WaitGroup` or channels to wait for goroutines
- Avoid goroutine leaks by ensuring cleanup

### Channels

- Use channels to communicate between goroutines
- Don't communicate by sharing memory; share memory by communicating
- Close channels from the sender side, not the receiver
- Use buffered channels when you know the capacity
- Use `select` for non-blocking operations

### Synchronization

- Use `sync.Mutex` for protecting shared state
- Keep critical sections small
- Use `sync.RWMutex` when you have many readers
- Prefer channels over mutexes when possible
- Use `sync.Once` for one-time initialization

## Error Handling Patterns

### Creating Errors

- Use `errors.New` for simple static errors
- Use `fmt.Errorf` for errors with runtime values
- Create custom error types for domain-specific errors
- Export error variables for sentinel errors
- Use `errors.Is` and `errors.As` for error checking

### Error Propagation

- Add context when propagating errors up the stack
- Don't log and return errors (choose one)
- Handle errors at the appropriate level
- Consider using structured errors for better debugging

## Performance Optimization

### Memory Management

- Minimize allocations in hot paths
- Reuse objects when possible (consider `sync.Pool`)
- Use value receivers for small structs
- Preallocate slices when size is known
- Avoid unnecessary string conversions

### Profiling

- Use built-in profiling tools (`pprof`)
- Benchmark critical code paths
- Profile before making performance changes
- Focus on algorithmic improvements first
- Consider using `testing.B` for benchmarks

## Testing

### Test Organization

- Keep tests in the same package (white-box testing)
- Use `_test` package suffix for black-box testing
- Name test files with `_test.go` suffix
- Place test files next to the code they test

### Writing Tests

- Name tests descriptively using `TestFunctionNameScenario`
- Use subtests with `t.Run` for better organization
- Test both success and error cases
- Use `testify/assert` and `testify/require` for assertions
- Include both positive and negative test cases
- Test edge cases and error conditions
- When including a standard library that conflicts with an existing import,
  use the lib(library name) pattern to avoid conflicts.
  Such as: `libtime` for the `time` package.

#### Table-driven tests are the default

One behavior under test = one test function with a table of cases. Never write several
near-identical test functions that differ only in input data, fixtures, or expected outcome —
those differences are table fields. A per-case fixture (a different map, config, or mock return)
is not a reason to split; put the fixture in the table. Shared setup (mocks, caches, `Init`
calls) runs once before the loop.

When adding cases to an existing test file, extend the existing table instead of adding a new
test function.

Only split into separate test functions when the flow genuinely differs: a different API under
test, or a setup/assertion sequence that cannot be expressed as table fields.

```go
// ✅ CORRECT: fixture and error expectation are table fields
cases := []struct {
    Fixture       Palette
    Case          string
    Input         Ansi
    Expected      Ansi
    ExpectedError bool
}{
    {Case: "literal", Fixture: Palette{"a": "#123456"}, Input: "p:a", Expected: "#123456"},
    {Case: "invalid", Fixture: Palette{"a": "{{ broken"}, Input: "p:a", ExpectedError: true},
}

// ❌ WRONG: TestResolveLiteral, TestResolveReference, TestResolveInvalid —
// three functions repeating the same setup with different data
```

### Test Helpers

- Mark helper functions with `t.Helper()`
- Create test fixtures for complex setup
- Use `testing.TB` interface for functions used in tests and benchmarks
- Clean up resources using `t.Cleanup()`

#### Global state: always save the original value and restore it

When a test mutates package-level variables (resolvers, loggers, clocks, `time.Local`, etc.),
save the original value into a local variable and restore it via `t.Cleanup`. Never restore to a
hardcoded value; you would overwrite whatever state preceded your test.

```go
// ✅ CORRECT: save original, restore original
origResolver := myPackageResolver
t.Cleanup(func() { myPackageResolver = origResolver })
myPackageResolver = fakeResolver

origLocal := time.Local
t.Cleanup(func() { time.Local = origLocal })
time.Local = time.UTC

// ❌ WRONG: restores to a hardcoded value instead of the pre-test value
defer func() { time.Local = time.FixedZone("UTC", 0) }()
```

## Security Best Practices

### Input Validation

- Validate all external input
- Use strong typing to prevent invalid states
- Sanitize data before using in SQL queries
- Be careful with file paths from user input
- Validate and escape data for different contexts (HTML, SQL, shell)

### Cryptography

- Use standard library crypto packages
- Never write your own cryptography
- Use crypto/rand for random number generation
- Store passwords using bcrypt or similar
- Use TLS for network communication

## Documentation

### Code Documentation

- Document all exported symbols
- Start documentation with the symbol name
- Use examples in documentation when helpful
- Keep documentation close to code
- Update documentation when code changes

### README and Documentation Files

- Include clear setup instructions
- Document dependencies and requirements
- Provide usage examples
- Document configuration options
- Include troubleshooting section

## Tools and Development Workflow

### Essential Tools

- `go fmt`: Format code
- `go vet`: Find suspicious constructs
- `golint` or `golangci-lint`: Additional linting
- `go test`: Run tests
- `go mod`: Manage dependencies
- `go generate`: Code generation

### Development Practices

- Run tests before committing
- Use pre-commit hooks for formatting and linting
- Keep commits focused and atomic
- Write clear, descriptive commit messages
- Review diffs before committing

### Pre-Commit Quality Gate

**REQUIRED BEFORE EVERY COMMIT.** Run the following commands in sequence after any Go code
change. Commit after all pass with zero errors. Never skip this step; these
linters catch real bugs and style violations that will be flagged in CI or code review.

1. **Code Modernization**: Apply modern Go best practices; this rewrites files in place

   ```bash
   modernize --fix "./..."
   ```

   > `modernize` modifies source files (e.g. replacing `strings.Split`
   > with `strings.SplitSeq` for Go 1.24+ range loops). Always stage its changes and
   > include them in the same commit as your feature code.

2. **Field Alignment**: Optimize struct field ordering for memory efficiency; this rewrites files in place

   ```bash
   fieldalignment --fix "./..."
   ```

   > **Warning:** `fieldalignment` rewrites struct field order. Any inline struct
   > literals that use **positional** (unnamed) field initialization (common in
   > table-driven test files) will break after the reorder.
   > **Always use named fields** in struct literals (e.g. `{Case: "foo", Now: t}`)
   > so that the order of fields in the struct definition does not matter.

3. **Dependency Management**: Clean up and organize module dependencies

   ```bash
   go mod tidy
   ```

4. **Formatting and Linting**: Ensure code follows standards (**must report zero errors**)

   ```bash
   gofmt -w .
   golangci-lint run
   ```

After steps 1 through 3, always run `git diff` to review auto-applied changes before staging them.
All four steps must complete with zero errors before the commit is created.

#### Platform-specific files (`_unix.go`, `_windows.go`, `_darwin.go`)

If you add or modify a file with a platform-specific suffix, also cross-compile to catch
issues the local OS linter skips. On Windows, run:

```powershell
$env:GOOS = "linux"; go build ./...; $env:GOOS = ""
```

On Linux/macOS, run:

```bash
GOOS=windows go build ./...
```

This catches import mismatches, missing symbols, and linter rules (like `modernize`
`strings.SplitSeq`) that apply on the non-host platform.

#### Common golangci-lint violations to fix before committing

These rules frequently fire on new code and are quick to resolve before linting:

| Linter | Trigger | Fix |
| ------ | ------- | --- |
| `goconst` | Same string literal occurs 3+ times | Extract to a named `const` |
| `gofmt` | Incorrect indentation or comment spacing | Run `gofmt -w .`; it fixes automatically |
| `dupl` | Two functions/test cases with near-identical structure | Add `//nolint:dupl` with a brief reason comment |
| `modernize` | `strings.Split` used in a `for range` (Go 1.24+) | Run `modernize --fix "./..."` (auto-fixes) |

## Common Pitfalls to Avoid

- Not checking errors
- Ignoring race conditions
- Creating goroutine leaks
- Not using defer for cleanup
- Modifying maps concurrently
- Confusing nil interfaces with nil pointers
- Forgetting to close resources (files, connections)
- Using global variables unnecessarily
- Over-using empty interfaces (`interface{}` or `any`)
- Not considering the zero value of types
