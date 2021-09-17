---
id: shell
title: Shell
sidebar_label: Shell
---

## What

Show the current shell name (ZSH, powershell, bash, ...).

## Sample Configuration

```json
{
  "type": "shell",
  "style": "powerline",
  "powerline_symbol": "\uE0B0",
  "foreground": "#ffffff",
  "background": "#0077c2",
  "properties": {
    "mapped_shell_names": {
      "pwsh": "PS"
    }
  }
}
```

## Properties

- mapped_shell_names: `object` - custom glyph/text to use in place of specified shell names (case-insensitive)
