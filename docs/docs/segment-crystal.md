---
id: crystal
title: Crystal
sidebar_label: Crystal
---

## What

Display the currently active crystal version.

## Sample Configuration

```json
{
  "type": "crystal",
  "style": "powerline",
  "powerline_symbol": "\uE0B0",
  "foreground": "#ffffff",
  "background": "#4063D8",
  "properties": {
    "prefix": " \uE370 "
  }
}
```

## Properties

- home_enabled: `boolean` - display the segment in the HOME folder or not - defaults to `false`
- display_version: `boolean` - display the julia version - defaults to `true`
- display_error: `boolean` - show the error context when failing to retrieve the version information - defaults to `true`
- missing_command_text: `string` - text to display when the command is missing - defaults to empty
- display_mode: `string` - determines when the segment is displayed
  - `always`: the segment is always displayed
  - `files`: the segment is only displayed when `*.cr` or `shard.yml` files are present (default)
- template: `string` - A go [text/template][go-text-template] template extended with [sprig][sprig] utilizing the
  properties below. Defaults to `{{ .Full }}`

## Template Properties

- `.Full`: `string` - the full version
- `.Major`: `string` - is the major version
- `.Minor`: `string` - is the minor version
- `.Patch`: `string` - is the patch version
- `.Prerelease`: `string` - is the prerelease version
- `.BuildMetadata`: `string` - is the build metadata

[go-text-template]: https://golang.org/pkg/text/template/
[sprig]: https://masterminds.github.io/sprig/
