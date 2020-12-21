---
id: python
title: Python
sidebar_label: Python
---

## What

Display the currently active python version and virtualenv.
Supports conda, virtualenv and pyenv.

## Sample Configuration

```json
{
  "type": "python",
  "style": "powerline",
  "powerline_symbol": "\uE0B0",
  "foreground": "#100e23",
  "background": "#906cff",
  "properties": {
    "prefix": " \uE235 "
  }
}
```

## Properties

- display_virtual_env: `boolean` - show the name of the virtualenv or not - defaults to `true`
- display_mode: `string` - determines when the segment is displayed
  - `always`: The segment is always displayed
  - `context`: The segment is only displayed when *.py or *.ipynb files are present (default)
  - `never`: The segement is hidden
