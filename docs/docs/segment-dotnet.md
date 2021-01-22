---
id: dotnet
title: Dotnet
sidebar_label: Dotnet
---

## What

Display the currently active .NET SDK version when a folder contains .NET files.

## Sample Configuration

```json
{
  "type": "dotnet",
  "style": "powerline",
  "powerline_symbol": "",
  "foreground": "#000000",
  "background": "#00ffff",
  "properties": {
    "prefix": " \uE77F "
  }
}
```

## Properties

- display_version: `boolean` - display the active version or not; useful if all you need is an icon indicating `dotnet`
  is present - defaults to `true`
- missing_command_text: `string` - text to display when the command is missing - defaults to empty
- display_mode: `string` - determines when the segment is displayed
  - `always`: The segment is always displayed
  - `files`: The segment is only displayed when `*.cs`, `*.vb`, `*.sln`, `*.csproj`, or `*.vbproj` files are present (default)
- unsupported_version_icon: `string` - text/icon that is displayed when the active .NET SDK version (e.g., one specified
  by `global.json`) is not installed/supported - defaults to `\uf071` (X in a rectangle box)
