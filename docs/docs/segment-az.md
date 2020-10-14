---
id: az
title: Azure Subscription
sidebar_label: Azure
---

## What

Display the currently active Azure subscription information.

## Sample Configuration

```json
{
  "type": "az",
  "style": "powerline",
  "powerline_symbol": "",
  "foreground": "#000000",
  "background": "#9ec3f0",
  "properties": {
    "display_id": true,
    "display_name": true,
    "info_separator": " @ ",
    "prefix": " ﴃ "
  }
}
```

## Properties

- info_separator: `string` - text/icon to put in between the subscription name and ID - defaults to ` | `
- display_id: `boolean` - display the subscription ID or not - defaults to `false`
- display_name: `boolean` - display the subscription name or not - defaults to `true`
