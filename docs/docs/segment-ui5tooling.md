---
id: ui5tooling
title: UI5 Tooling
sidebar_label: UI5 Tooling
---

## What

Display the active [UI5 tooling][ui5-homepage] version (global or local if present -
see [the documentation][ui5-version-help]).

## Sample Configuration

```json
{
  "background": "#f5a834",
  "foreground": "#100e23",
  "powerline_symbol": "\ue0b0",
  "template": " \ufab6ui5 {{ .Full }} ",
  "style": "powerline",
  "type": "ui5tooling"
}
```

## Properties

- home_enabled: `boolean` - display the segment in the HOME folder or not - defaults to `false`
- fetch_version: `boolean` - display the UI5 tooling version - defaults to `true`
- display_error: `boolean` - show the error context when failing to retrieve the version information - defaults to `true`
- missing_command_text: `string` - text to display when the java command is missing - defaults to empty
- display_mode: `string` - determines when the segment is displayed
  - `always`: the segment is always displayed
  - `files`: the segment is only displayed when `*ui5*.y(a)ml` file is present in the current folder
  - `context`: (default) the segment is only displayed when `*ui5*.y(a)ml` file is present in the current folder
    or it has been found in the parent folders (check up to 4 levels)

## Template ([info][templates])

:::note default template

```template
{{ if .Error }}{{ .Error }}{{ else }}{{ .Full }}{{ end }}
```

:::

## Template Properties

- `.Full`: `string` - the full version
- `.Major`: `string` - major number
- `.Minor`: `string` - minor number
- `.Patch`: `string` - patch number
- `.Prerelease`: `string` - prerelease info text
- `.BuildMetadata`: `string` - build metadata
- `.Error`: `string` - when fetching the version string errors

[templates]: /docs/config-templates
[ui5-homepage]: https://sap.github.io/ui5-tooling
[ui5-version-help]: https://sap.github.io/ui5-tooling/pages/CLI/#ui5-versions
