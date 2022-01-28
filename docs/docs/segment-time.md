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

## [Template][templates] Properties

- `.CurrentDate`: `time` - The time to display(testing purpose)

### Standard time and date formats

- January 2, 2006 **Date**
- 01/02/06
- Jan-02-06
- 15:04:05 **Time**
- 3:04:05 PM
- Jan _2 15:04:05 **Timestamp**
- Jan _2 15:04:05.000000 **with microseconds**
- 2006-01-02T15:04:05-0700 **ISO 8601 (RFC 3339)**
- 2006-01-02
- 15:04:05
- 02 Jan 06 15:04 MST **RFC 822**
- 02 Jan 06 15:04 -0700 **with numeric zone**
- Mon, 02 Jan 2006 15:04:05 MST 27e95cb
- Mon, 02 Jan 2006 15:04:05 -0700 **with numeric zone**

#### The following predefined date and timestamp format constants are also available

- ANSIC       = "Mon Jan _2 15:04:05 2006"
- UnixDate    = "Mon Jan _2 15:04:05 MST 2006"
- RubyDate    = "Mon Jan 02 15:04:05 -0700 2006"
- RFC822      = "02 Jan 06 15:04 MST"
- RFC822Z     = "02 Jan 06 15:04 -0700"
- RFC850      = "Monday, 02-Jan-06 15:04:05 MST"
- RFC1123     = "Mon, 02 Jan 2006 15:04:05 MST"
- RFC1123Z    = "Mon, 02 Jan 2006 15:04:05 -0700"
- RFC3339     = "2006-01-02T15:04:05Z07:00"
- RFC3339Nano = "2006-01-02T15:04:05.999999999Z07:00"
- Kitchen     = "3:04PM"
// Useful time stamps.
- Stamp      = "Jan _2 15:04:05"
- StampMilli = "Jan _2 15:04:05.000"
- StampMicro = "Jan _2 15:04:05.000000"
- StampNano  = "Jan _2 15:04:05.000000000"

[templates]: /docs/config-templates
[format]: https://yourbasic.org/golang/format-parse-string-time-date-example/
