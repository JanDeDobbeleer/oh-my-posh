---
id: rust
title: Rust
sidebar_label: Rust
---

## What

Display the currently active rust version.

## Sample Configuration

```json
{
  "type": "rust",
  "style": "powerline",
  "powerline_symbol": "\uE0B0",
  "foreground": "#193549",
  "background": "#99908a",
  "template": " \uE7a8 {{ .Full }} "
}
```

## Properties

- home_enabled: `boolean` - display the segment in the HOME folder or not - defaults to `false`
- fetch_version: `boolean` - display the rust version (`rustc --version`) - defaults to `true`
- missing_command_text: `string` - text to display when the command is missing - defaults to empty
- display_mode: `string` - determines when the segment is displayed
  - `always`: the segment is always displayed
  - `files`: the segment is only displayed when `*.rs`, `Cargo.toml` or `Cargo.lock` files are present (default)

## Template ([info][templates])

:::note default template

``` template
{{ if .Error }}{{ .Error }}{{ else }}{{ .Full }}{{ end }}
```

:::

### Properties

- `.Full`: `string` - the full version
- `.Major`: `string` - major number
- `.Minor`: `string` - minor number
- `.Patch`: `string` - patch number
- `.Error`: `string` - error encountered when fetching the version string

[templates]: /docs/configuration/templates
