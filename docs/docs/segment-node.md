---
id: node
title: Node
sidebar_label: Node
---

## What

Display the currently active node version when a folder contains node files.

## Sample Configuration

```json
{
  "type": "node",
  "style": "powerline",
  "powerline_symbol": "\uE0B0",
  "foreground": "#ffffff",
  "background": "#6CA35E",
  "properties": {
    "prefix": " \uE718 "
  }
}
```

## Properties

- display_version: `boolean` - display the node version - defaults to `true`
- missing_command_text: `string` - text to display when the command is missing - defaults to empty
- display_mode: `string` - determines when the segment is displayed
  - `always`: The segment is always displayed
  - `files`: The segment is only displayed when `*.js`, `*.ts`, or `package.json` files are present (default)
