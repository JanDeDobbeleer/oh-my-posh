---
id: dotnet
title: Dotnet
sidebar_label: Dotnet
---

## What

Display the currently active .NET SDK version.

## Sample Configuration

```json
{
  "type": "dotnet",
  "style": "powerline",
  "powerline_symbol": "\uE0B0",
  "foreground": "#000000",
  "background": "#00ffff",
  "properties": {
    "prefix": " \uE77F "
  }
}
```

## Properties

- home_enabled: `boolean` - display the segment in the HOME folder or not - defaults to `false`
- fetch_version: `boolean` - fetch the active version or not; useful if all you need is an icon indicating `dotnet`
  is present - defaults to `true`
- display_error: `boolean` - show the error context when failing to retrieve the version information - defaults to `true`
- missing_command_text: `string` - text to display when the command is missing - defaults to empty
- display_mode: `string` - determines when the segment is displayed
  - `always`: the segment is always displayed
  - `files`: the segment is only displayed when `*.cs`, `*.vb`, `*.fs`, `*.fsx`, `*.sln`, `*.csproj`, `*.vbproj`,
  or `*.fsproj` files are present (default)
- version_url_template: `string` - A go text/template [template][templates] that creates the changelog URL

## Template ([info][templates])

:::note default template

``` template
{{ if .Unsupported }}\uf071{{ else }}{{ .Full }}{{ end }}
```

:::

### Properties

- `.Full`: `string` - the full version
- `.Major`: `string` - major number
- `.Minor`: `string` - minor number
- `.Patch`: `string` - patch number
- `.Prerelease`: `string` - prerelease info text
- `.BuildMetadata`: `string` - build metadata
- `.Error`: `string` - when fetching the version string errors

[templates]: /docs/config-templates
