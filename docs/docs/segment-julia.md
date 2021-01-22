---
id: julia
title: Julia
sidebar_label: Julia
---

## What

Display the currently active julia version when a folder contains julia files.

## Sample Configuration

```json
{
  "type": "julia",
  "style": "powerline",
  "powerline_symbol": "\uE0B0",
  "foreground": "#ffffff",
  "background": "#4063D8",
  "properties": {
    "prefix": " \uE624 "
  }
}
```

## Properties

- display_version: `boolean` - display the julia version - defaults to `true`
- missing_command_text: `string` - text to display when the command is missing - defaults to empty
- display_mode: `string` - determines when the segment is displayed
  - `always`: The segment is always displayed
  - `files`: The segment is only displayed when `*.jl` files are present (default)
