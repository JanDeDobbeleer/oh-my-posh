---
id: julia
title: Julia
sidebar_label: Julia
---

## What

Display the currently active julia version.

## Sample Configuration

```json
{
  "type": "julia",
  "style": "powerline",
  "powerline_symbol": "\uE0B0",
  "foreground": "#ffffff",
  "background": "#4063D8",
  "template": " \uE624 {{ .Full }} "
}
```

## Properties

- home_enabled: `boolean` - display the segment in the HOME folder or not - defaults to `false`
- fetch_version: `boolean` - display the julia version - defaults to `true`
- missing_command_text: `string` - text to display when the command is missing - defaults to empty
- display_mode: `string` - determines when the segment is displayed
  - `always`: the segment is always displayed
  - `files`: the segment is only displayed when `*.jl` files are present (default)
- version_url_template: `string` - a go [text/template][go-text-template] [template][templates] that creates
the URL of the version info / release notes

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
- `.URL`: `string` - URL of the version info / release notes
- `.Error`: `string` - error encountered when fetching the version string

[go-text-template]: https://golang.org/pkg/text/template/
[templates]: /docs/configuration/templates
