# GitHub Copilot Instructions for oh-my-posh

## Commit and Pull Requests Guidelines

- When naming a PR or commits, always follow [conventional commits](https://www.conventionalcommits.org/en/v1.0.0/#summary)
- All commits must follow the conventional commit format. You can find the repository speciffic rules in .commitlintrc.json.
- Use golangci-lint and gofmt before submitting
- The max length of a commit message line is 200 characters

## Go-Specific Guidelines

### Function Structure

- Keep functions focused and single-purpose
- Prefer early returns to reduce nesting
- Validate inputs at the beginning of functions
- Use guard clauses to handle edge cases early
- Do not use if/else but prefer switch statements for multiple conditions, or early returns for simple conditions
- Never create exported functions which aren't used in the codebase.

### Testing

- Follow table-driven test patterns established in the codebase
- Use `testify/assert` and `testify/require` for assertions
- Name test cases descriptively to explain what is being tested
- Include both positive and negative test cases
- Test edge cases and error conditions
- When including a standard library that conflicts with an existing import, use the lib(library name) pattern to avoid conflicts. e.g. `libtime` for `time` package.

### Coding Style

- One level of indentation per function.
- Don't use the ELSE keyword.
- Wrap all primitives and Strings in structs or types.
- First class collections.
- One dot per line.
- Start error strings with a lowercase letter.


