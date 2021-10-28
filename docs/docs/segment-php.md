---
id: php
title: php
sidebar_label: PHP
---

## What

Display the currently active php version.

## Sample Configuration

```json
{
  "type": "php",
  "style": "powerline",
  "powerline_symbol": "\uE0B0",
  "foreground": "#ffffff",
  "background": "#4063D8",
  "properties": {
    "prefix": " \ue73d ",
    "enable_hyperlink": false
  }
}
```

## Properties

- home_enabled: `boolean` - display the segment in the HOME folder or not - defaults to `false`
- display_version: `boolean` - display the php version - defaults to `true`
- display_error: `boolean` - show the error context when failing to retrieve the version information - defaults to `true`
- missing_command_text: `string` - text to display when the command is missing - defaults to empty
- display_mode: `string` - determines when the segment is displayed
  - `always`: the segment is always displayed
  - `files`: the segment is only displayed when `*.php, composer.json, composer.lock, .php-version` files are present (default)
- enable_hyperlink: `bool` - display an hyperlink to the php release notes - defaults to `false`
