---
name: segment-docs
description: >
  Reference mapping between Oh My Posh Go segment source code and MDX documentation.
  Use when creating, updating, or auditing segment documentation, or when reading
  segment Go code and needing to understand how code constructs map to user-facing
  options, template properties, and type representations in docs.
---

## Go-to-Documentation Type Mapping

When reading a segment's Go source to document its options, determine the type from the
`Provider` method used to read the option value:

| Go method call | Documentation type |
| -------------- | :----------------: |
| `options.Bool(option, default)` | `boolean` |
| `options.String(option, default)` | `string` |
| `options.StringArray(option, default)` | `[]string` |
| `options.Int(option, default)` | `int` |
| `options.Float64(option, default)` | `float64` |
| `options.KeyValueMap(option, default)` | `map[string]string` |
| `options.Color(option, default)` | `string` |
| `options.Template(option, default, ctx)` | `string` |
| `options.Any(option, default)` | `any` |

## Extracting Options from Go Source

Options are declared as `options.Option` string constants in the segment's `const` block:

```go
const (
    BranchIcon    options.Option = "branch_icon"    // option name used in config
    FetchStatus   options.Option = "fetch_status"
)
```

The **option name** is the string literal value (e.g., `"branch_icon"`).

The **type** and **default value** come from the getter call in `Enabled()` or `init`-like methods:

```go
s.icon = s.options.String(BranchIcon, "\uE0A0")
//                  ^^^^^^ → string type, "\uE0A0" → default
```

### Shared options

The following options are defined in `src/segments/options/map.go` and available to all
segments without appearing in a segment-specific `const` block:

| Option name | Type | Default |
| ----------- | ---- | ------- |
| `fetch_version` | `boolean` | varies |
| `always_enabled` | `boolean` | `false` |
| `display_default` | `boolean` | `true` |
| `display_error` | `boolean` | `true` |
| `http_timeout` | `int` | `20` |
| `cache_duration` | `int` | varies |
| `access_token` | `string` | `""` |
| `refresh_token` | `string` | `""` |
| `version_url_template` | `string` | `""` |
| `files` | `[]string` | varies |

## Extracting Template Properties from Go Source

Template properties are values available inside the segment's `template` string.

**Rules for what becomes a template property:**

- **Exported struct fields** (capitalized) on the segment struct → `.FieldName`
- **Unexported fields** are not accessible in templates
- **Nested struct fields** use dot notation: a field `Working ScmStatus` exposes
  `.Working.Modified`, `.Working.Added`, etc.
- **Zero-argument methods** with a single return value → `.MethodName`
- **Document fields that are assigned** in `Enabled()` or its callees,
  not every inherited field from embedded structs is populated by every segment

The default template comes from the `Template()` method:

```go
func (s *MySegment) Template() string {
    return " {{ .FieldName }} "
}
```

### Nested struct subsection convention

When a property's type is itself a struct, add an `#### TypeName` subsection within
the Properties table section listing its exported fields. Example:

```mdx
#### Status

| Name | Type | Description |
| ---- | ---- | ----------- |
| `.Modified` | `int` | Number of modified files |
| `.Added` | `int` | Number of added files |
| `.Deleted` | `int` | Number of deleted files |
| `.Untracked` | `int` | Number of untracked files |
| `.Changed` | `boolean` | Whether any changes exist |
```

## Key Source Files for Inherited Fields

When a segment embeds a base struct, its inherited fields may also be template properties.
Check these files to discover what fields are available:

| File | What it defines |
| ---- | --------------- |
| `src/segments/options/map.go` | `Provider` interface method signatures; shared `Option` constants |
| `src/segments/base.go` | `Base` struct embedded by all segments; provides `options`, `env` |
| `src/segments/scm.go` | `ScmStatus` type with working/staging status fields (git, mercurial, etc.) |
| `src/segments/language.go` | `Version` type and language base fields (Python, Node, Go, etc.) |
