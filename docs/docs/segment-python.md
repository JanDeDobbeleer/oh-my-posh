---
id: python
title: Python
sidebar_label: Python
---

## What

Display the currently active python version and virtualenv when a folder contains `.py` files or `.ipynb` files.
Supports conda, virtualenv and pyenv.

## Sample Configuration

```json
{
  "type": "python",
  "style": "powerline",
  "powerline_symbol": "",
  "foreground": "#100e23",
  "background": "#906cff",
  "properties": {
    "prefix": "  "
  }
}
```

## Properties

- display_virtual_env: `boolean` - show the name of the virtualenv or not, defaults to `true`.
