---
id: ruby
title: Ruby
sidebar_label: Ruby
---

## What

Display the currently active ruby version.

## Sample Configuration

```json
{
  "type": "ruby",
  "style": "powerline",
  "powerline_symbol": "\uE0B0",
  "foreground": "#ffffff",
  "background": "#4063D8",
  "properties": {
    "prefix": " \uE791 "
  }
}
```

## Properties

- home_enabled: `boolean` - display the segment in the HOME folder or not - defaults to `false`
- display_version: `boolean` - display the ruby version - defaults to `true`
- display_error: `boolean` - show the error context when failing to retrieve the version information - defaults to `true`
- missing_command_text: `string` - text to display when the command is missing - defaults to empty
- display_mode: `string` - determines when the segment is displayed
  - `always`: the segment is always displayed
  - `files`: the segment is only displayed when `*.rb`, `Gemfile` or `Rakefile` files are present (default)
