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
- display_default_env: `boolean` - show the name of the virtualenv when it's default (`system`, `base`)
or not - defaults to `true`
- display_version: `boolean` - display the python version - defaults to `true`
- missing_command_text: `string` - text to display when the command is missing - defaults to empty
- display_mode: `string` - determines when the segment is displayed
  - `always`: The segment is always displayed
  - `files`: The segment is only displayed when `*.py`, `*.ipynb`, `pyproject.toml`, `venv.bak`, `venv`, or `.venv`
    files are present (default)
  - `environment`: The segment is only displayed when a virtual env is present
  - `context`: The segment is only displayed when either `environment` or `files` is active
