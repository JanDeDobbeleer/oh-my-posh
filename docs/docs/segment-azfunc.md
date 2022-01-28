---
id: azfunc
title: Azure functions
sidebar_label: Azure functions
---

## What

Display the currently active Azure functions CLI version.

## Sample Configuration

```json
{
    "type": "azfunc",
    "style": "powerline",
    "powerline_symbol": "\uE0B0",
    "foreground": "#ffffff",
    "background": "#FEAC19",
    "properties": {
        "prefix": " \uf0e7 ",
        "fetch_version": true,
        "display_mode": "files"
    }
}
```

## Properties

- fetch_version: `boolean` - fetch the Azure functions CLI version - defaults to `true`
- display_error: `boolean` - show the error context when failing to retrieve the version information - defaults to `true`
- missing_command_text: `string` - text to display when the command is missing - defaults to empty
- display_mode: `string` - determines when the segment is displayed
  - `always`: the segment is always displayed
  - `files`: the segment is only displayed when a `host.json` or `local.settings.json` files is present (default)

## [Template][templates] Properties

- `.Full`: `string` - the full version
- `.Major`: `string` - major number
- `.Minor`: `string` - minor number
- `.Patch`: `string` - patch number
- `.Prerelease`: `string` - prerelease info text
- `.BuildMetadata`: `string` - build metadata
- `.Error`: `string` - when fetching the version string errors

[templates]: /docs/config-templates
