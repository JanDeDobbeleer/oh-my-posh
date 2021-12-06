---
id: juliandate
title: Julian Date
sidebar_label: Julian Date
---

## What

The Julian date format is used in some industries. This is usually in the format *YYDDD* where:

- *YY* : is the year
- *DDD* : is the day of the year

### Example

For the gregorian calendar date 6th December 2021. We would take the year to be 21 and the day of the year to be 340
(31 days in January then keep counting until 6th December)
which results in 21340 as the julian date

## Sample Configuration

```json
        {
          "type": "juliandate",
          "style": "diamond",
          "foreground": "#ffffff",
          "background": "#0000ff",
          "leading_diamond": "\uE0B2",
          "trailing_diamond": "\uE0B0",
          "properties": {
            "template": "\uf5f5 {{.Year}}{{.DayOfYear}}"
          }
        },
```

## Properties

- template: `string` - You can add your own icon into the template if you wish as in the example, otherwise the
segment will just display the YYDDD as default if no template property is provided

## Template Properties

- .Year: `string` - The two digit year value (for 2021 will be 21)
- .DayOfYear: `string` - The three digit day of the year value (this is padded by 0's for numbers less than 100,
for example, 1st January will be 001)

[go-text-template]: https://golang.org/pkg/text/template/
[sprig]: https://masterminds.github.io/sprig/
[nightscout]: http://www.nightscout.info/
