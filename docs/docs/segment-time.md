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

- template: `string` - A go [text/template][go-text-template] template extended with [sprig][sprig] utilizing the
properties below. Only used when a value is set, making the above properties obsolete.

  example: `{{ now | date \"January 02, 2006 15:04:05 PM\" | lower }}`

## Template Properties

- `.CurrentDate`: `time` - The time to display(testing purpose)
