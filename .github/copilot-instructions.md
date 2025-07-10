# GitHub Copilot Instructions for oh-my-posh

---
description: 'Instructions for writing Go code following idiomatic Go practices and community standards'
applyTo: '/*.go,/go.mod,/go.sum'
---

# Go Development Instructions

Follow idiomatic Go practices and community standards when writing Go code. These instructions are based on [Effective Go](https://go.dev/doc/effective_go), [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments), and [Google's Go Style Guide](https://google.github.io/styleguide/go/).

## General Instructions

- Write simple, clear, and idiomatic Go code
- Favor clarity and simplicity over cleverness
- Follow the principle of least surprise
- Keep the happy path left-aligned (minimize indentation)
- Return early to reduce nesting
- Make the zero value useful
- Document exported types, functions, methods, and packages
- Use Go modules for dependency management
- One level of indentation per function.
- Don't use the ELSE keyword.
- Wrap all primitives and Strings in structs or types.
- First class collections.
- One dot per line.
- Start error strings with a lowercase letter.

## Naming Conventions

### Packages

- Use lowercase, single-word package names
- Avoid underscores, hyphens, or mixedCaps
- Choose names that describe what the package provides, not what it contains
- Avoid generic names like `util`, `common`, or `base`
- Package names should be singular, not plural

### Variables and Functions

- Use mixedCaps or MixedCaps (camelCase) rather than underscores
- Keep names short but descriptive
- Use single-letter variables only for very short scopes (like loop indices)
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
- For complex function calls, use `defer log.Trace(time.Now(), args)` where args are the function arguments at the start of the function.

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
- Vendor dependencies only when necessary

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
- Use `fmt.Errorf` for dynamic errors
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
- Profile before optimizing
- Focus on algorithmic improvements first
- Consider using `testing.B` for benchmarks

## Testing

### Test Organization

- Keep tests in the same package (white-box testing)
- Use `_test` package suffix for black-box testing
- Name test files with `_test.go` suffix
- Place test files next to the code they test

### Writing Tests

- Use table-driven tests for multiple test cases
- Name tests descriptively using `TestFunctionNameScenario`
- Use subtests with `t.Run` for better organization
- Test both success and error cases
- Use `testify/assert` and `testify/require` for assertions
- Include both positive and negative test cases
- Test edge cases and error conditions
- When including a standard library that conflicts with an existing import, use the lib(library name) pattern to avoid conflicts. e.g. `libtime` for `time` package.

### Test Helpers

- Mark helper functions with `t.Helper()`
- Create test fixtures for complex setup
- Use `testing.TB` interface for functions used in tests and benchmarks
- Clean up resources using `t.Cleanup()`

## Security Best Practices

### Input Validation

- Validate all external input
- Use strong typing to prevent invalid states
- Sanitize data before using in SQL queries
- Be careful with file paths from user input
- Validate and escape data for different contexts (HTML, SQL, shell)

### Cryptography

- Use standard library crypto packages
- Don't implement your own cryptography
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
- Write meaningful commit messages
- Review diffs before committing

## Common Pitfalls to Avoid

- Not checking errors
- Ignoring race conditions
- Creating goroutine leaks
- Not using defer for cleanup
- Modifying maps concurrently
- Not understanding nil interfaces vs nil pointers
- Forgetting to close resources (files, connections)
- Using global variables unnecessarily
- Over-using empty interfaces (`interface{}` or `any`)
- Not considering the zero value of types

---
description: 'Documentation and content creation standards'
applyTo: '/*.md', '/*.mdx'
---

# Markdown Content Rules

The following markdown content rules are enforced in the validators:

1. Headings: Use appropriate heading levels (H2, H3, etc.) to structure your content. Do not use an H1 heading, as this will be generated based on the title.
2. Lists: Use bullet points or numbered lists for lists. Ensure proper indentation and spacing.
3. Code Blocks: Use fenced code blocks for code snippets. Specify the language for syntax highlighting.
4. Links: Use proper markdown syntax for links. Ensure that links are valid and accessible.
5. Images: Use proper markdown syntax for images. Include alt text for accessibility.
6. Tables: Use markdown tables for tabular data. Ensure proper formatting and alignment.
7. Line Length: Limit line length to 400 characters for readability.
8. Whitespace: Use appropriate whitespace to separate sections and improve readability.
9. Front Matter: Include YAML front matter at the beginning of the file with required metadata fields.

## Formatting and Structure

Follow these guidelines for formatting and structuring your markdown content:

- Headings: Use `##` for H2 and `###` for H3. Ensure that headings are used in a hierarchical manner. Recommend restructuring if content includes H4, and more strongly recommend for H5.
- Lists: Use `-` for bullet points and `1.` for numbered lists. Indent nested lists with two spaces.
- Code Blocks: Use triple backticks (`) to create fenced code blocks. Specify the language after the opening backticks for syntax highlighting (e.g., ```powershell).
- Links: Use `[link text[link]` for links where `[link]: https://mylink.com` is added at the bottom of the page. Ensure that the link text is descriptive and the URL is valid.
- Images: Use `![alt text](image URL)` for images. Include a brief description of the image in the alt text.
- Tables: Use `|` to create tables. Ensure that columns are properly aligned and headers are included.
- Line Length: Break lines at 120 characters to improve readability. Use soft line breaks for long paragraphs.
- Whitespace: Use blank lines to separate sections and improve readability. Avoid excessive whitespace.

## Validation Requirements

Ensure compliance with the following validation requirements:

- Content Rules: Ensure that the content follows the markdown content rules specified above.
- Formatting: Ensure that the content is properly formatted and structured according to the guidelines.
- Validation: Run the validation tools to check for compliance with the rules and guidelines.

# Commit and Pull Requests Guidelines

- When naming a PR or commits, always follow [conventional commits](https://www.conventionalcommits.org/en/v1.0.0/#summary)
- All commits must follow the conventional commit format. You can find the repository speciffic rules in .commitlintrc.json.
- Use golangci-lint and gofmt before submitting
- The max length of a commit message line is 200 characters


