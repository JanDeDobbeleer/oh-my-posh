---
id: os
title: os
sidebar_label: OS
---

## What

Display OS specific info. Defaults to Icon.

## Sample Configuration

```json
{
  "type": "os",
  "style": "plain",
  "foreground": "#26C6DA",
  "background": "#546E7A",
  "properties": {
    "postfix": " ",
    "macos": "mac"
  }
}
```

## Properties

- macos: `string` - the string to use for macOS - defaults to macOS icon 
- linux: `string` - the icon to use for Linux - defaults to Linux icon
- windows: `string` - the icon to use for Windows - defaults to Windows icon

