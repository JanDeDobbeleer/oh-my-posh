---
id: osicon
title: osicon
sidebar_label: OS Icon
---

## What

Display OS specific icon.

## Sample Configuration

```json
{
  "type": "osicon",
  "style": "plain",
  "foreground": "#26C6DA",
  "background": "#546E7A",
  "properties": {
    "postfix": " ",
    "macos_icon": ""
  }
}
```

## Properties

Optional:
- macos_icon: `string` - the icon to use for macOS
- linux_icon: `string` - the icon to use for Linux
- windows_icon: `string` - the icon to use for Windows

