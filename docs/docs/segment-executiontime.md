---
id: executiontime
title: Execution Time
sidebar_label: Execution Time
---

## What

Displays the execution time of the previously executed command.

To use this, use the PowerShell module, or confirm that you are passing an `execution-time` argument containing the
elapsed milliseconds to the oh-my-posh executable.
The installation guide shows how to include this argument for PowerShell and Zsh.

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
    "style": "austin",
    "prefix": " <#fefefe>\ufbab</> "
  }
}
```

## Properties

- always_enabled: `boolean` - always show the duration - defaults to `false`
- threshold: `number` - minimum duration (milliseconds) required to enable this segment - defaults to `500`
- style: `enum` - one of the available format options - defaults to `austin`

## Style

Style specifies the format in which the time will be displayed. The table below shows some example times in each option.

| format    | 0.001s         | 2.1s         | 3m2.1s        | 4h3m2.1s         |
| --------- | -------------- | ------------ | ------------- | ---------------- |
| austin    | `1ms`          | `2.1s`       | `3m 2.1s`     | `4h 3m 2.1s`     |
| roundrock | `1ms`          | `2s 100ms`   | `3m 2s 100ms` | `4h 3m 2s 100ms` |
| dallas    | `0.001`        | `2.1`        | `3:2.1`       | `4:3:2.1`        |
| galveston | `00:00:00`     | `00:00:02`   | `00:03:02`    | `04:03:02`       |
| houston   | `00:00:00.001` | `00:00:02.1` | `00:03:02.1`  | `04:03:02.1`     |
| amarillo  | `0.001s`       | `2.1s`       | `182.1s`      | `14,582.1s`      |
| round     | `1ms`          | `2s`         | `3m 2s`       | `4h 3m`          |

## Template Properties

- `.Ms`: `number` - the execution time in milliseconds
- `.FormattedMs`: `string` - the formatted value based on the `style` above.
