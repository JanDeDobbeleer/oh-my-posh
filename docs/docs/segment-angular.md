---
id: angular
title: Angular
sidebar_label: Angular
---

## What

Display the currently active Angular CLI version.

## Sample Configuration

```json
{
  "type": "angular",
  "style": "powerline",
  "powerline_symbol": "",
  "foreground": "#000000",
  "background": "#1976d2",
  "properties": {
    "prefix": "\uE753"
  }
}
```

## Properties

- display_version: `boolean` - display the active version or not; useful if all you need is an icon indicating `ng`
- display_mode: `string` - determines when the segment is displayed
  - `always`: the segment is always displayed
  - `files`: the segment is only displayed when `angular.json` file is present (default)
