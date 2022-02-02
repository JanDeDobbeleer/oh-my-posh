---
id: angular
title: Angular
sidebar_label: Angular
---

## What

Display the currently active Angular CLI version.

## Sample Configuration

```json
{
  "type": "angular",
  "style": "powerline",
  "powerline_symbol": "\uE0B0",
  "foreground": "#000000",
  "background": "#1976d2",
  "properties": {
    "template": " \uE753 {{ .Full }} "
  }
}
```

## Properties

- fetch_version: `boolean` - fetch the active version or not; useful if all you need is an icon indicating `ng`
- display_mode: `string` - determines when the segment is displayed
  - `always`: the segment is always displayed
  - `files`: the segment is only displayed when `angular.json` file is present (default)

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
- `.Prerelease`: `string` - prerelease info text
- `.BuildMetadata`: `string` - build metadata
- `.Error`: `string` - when fetching the version string errors

[templates]: /docs/config-templates
