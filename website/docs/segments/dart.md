---
id: dart
title: Dart
sidebar_label: Dart
---

## What

Display the currently active dart version.

## Sample Configuration

```json
{
  "type": "dart",
  "style": "powerline",
  "powerline_symbol": "\uE0B0",
  "foreground": "#ffffff",
  "background": "#06A4CE",
  "template": " \uE798 {{ .Full }} "
}
```

## Properties

- home_enabled: `boolean` - display the segment in the HOME folder or not - defaults to `false`
- fetch_version: `boolean` - fetch the dart version - defaults to `true`
- missing_command_text: `string` - text to display when the command is missing - defaults to empty
- display_mode: `string` - determines when the segment is displayed
  - `always`: the segment is always displayed
  - `files`: the segment is only displayed when `*.dart`, `pubspec.yaml`, `pubspec.yml`, `pubspec.lock` files or the `.dart_tool`
folder are present (default)
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
