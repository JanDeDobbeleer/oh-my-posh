---
id: time
title: Time
sidebar_label: Time
---

## What

Show the current timestamp.

## Sample Configuration

```json
{
  "type": "time",
  "style": "plain",
  "foreground": "#007ACC",
  "properties": {
    "time_format": "15:04:05"
  }
}
```

## Properties

- time_format: `string` - format to use, follows the [golang standard][format] - defaults to `15:04:05`

[format]: https://gobyexample.com/time-formatting-parsing
