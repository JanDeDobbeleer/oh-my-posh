---
id: root
title: Root
sidebar_label: Root
---

## What

Show when the current user is root or when in an elevated shell (Windows).

## Sample Configuration

```json
{
  "type": "root",
  "style": "powerline",
  "powerline_symbol": "\uE0B0",
  "foreground": "#111111",
  "background": "#ffff66",
  "properties": {
    "root_icon": ""
  }
}
```

## Properties

- root_icon: `string` - icon to display in case of root/elevated.
