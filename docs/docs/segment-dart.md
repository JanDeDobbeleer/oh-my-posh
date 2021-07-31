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
  "properties": {
    "prefix": " \uE798 "
  }
}
```

## Properties

- home_enabled: `boolean` - display the segment in the HOME folder or not - defaults to `false`
- display_version: `boolean` - display the dart version - defaults to `true`
- display_error: `boolean` - show the error context when failing to retrieve the version information - defaults to `true`
- missing_command_text: `string` - text to display when the command is missing - defaults to empty
- display_mode: `string` - determines when the segment is displayed
  - `always`: the segment is always displayed
  - `files`: the segment is only displayed when `*.dart`, `pubspec.yaml`, `pubspec.yml`, `pubspec.lock` files or the `.dart_tool`
folder are present (default)
