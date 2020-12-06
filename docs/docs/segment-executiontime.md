---
id: executiontime
title: Execution Time
sidebar_label: Execution Time
---

## What

Displays the execution time of the previously executed command.

To use this, use the PowerShell module, or confirm that you are passing an `execution-time` argument contianing the elapsed milliseconds to the oh-my-posh executable. The [installation guide][install] shows how to include this argument for PowerShell and Zsh.

## Sample Configuration

```json
{
  "type": "executiontime",
  "style": "powerline",
  "powerline_symbol": "\uE0B0",
  "foreground": "#ffffff",
  "background": "#8800dd",
  "properties": {
    "threshold": 500,
    "prefix": " <#fefefe>\ufbab</> "
  }
}
```

## Properties

- threshold: `number` - minimum duration (milliseconds) required to enable this segment - defaults to `500`

[install]: /docs/installation