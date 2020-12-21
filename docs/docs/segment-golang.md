---
id: golang
title: Golang
sidebar_label: Golang
---

## What

Display the currently active golang version when a folder contains `.go` files.

## Sample Configuration

```json
{
  "type": "go",
  "style": "powerline",
  "powerline_symbol": "\uE0B0",
  "foreground": "#ffffff",
  "background": "#7FD5EA",
  "properties": {
    "prefix": " \uFCD1 "
  }
}
```

## Properties

- display_version: `boolean` - display the golang version - defaults to `true`
- display_mode: `string` - determines when the segment is displayed
  - `always`: The segment is always displayed
  - `context`: The segment is only displayed when *.go or go.mod files are present (default)
  - `never`: The segement is hidden
