---
id: battery
title: Battery
sidebar_label: Battery
---

## What

Battery displays the remaining power percentage for your battery.

## Sample Configuration

```json
{
  "type": "battery",
  "style": "powerline",
  "powerline_symbol": "\uE0B0",
  "foreground": "#193549",
  "background": "#ffeb3b",
  "properties": {
    "battery_icon": "",
    "discharging_icon": "\uE231 ",
    "charging_icon": "\uE234 ",
    "charged_icon": "\uE22F ",
    "color_background": true,
    "charged_color": "#4caf50",
    "charging_color": "#40c4ff",
    "discharging_color": "#ff5722",
    "postfix": "\uF295 ",
    "display_charging": true,
    "display_charged": true
  }
}
```

## Properties

- template: `string` - A go [text/template][go-text-template] template extended with [sprig][sprig] utilizing the
properties below. Defaults to `{{.Icon}}{{ if not .Error }}{{.Percentage}}{{ end }}{{.Error}}`
- display_error: `boolean` - show the error context when failing to retrieve the battery information - defaults to `false`
- charging_icon: `string` - icon to display on the left when charging - defaults to empty
- discharging_icon: `string` - icon to display on the left when discharging - defaults to empty
- charged_icon: `string` - icon to display on the left when fully charged - defaults to empty
- color_background: `boolean` - color the background or foreground for properties below - defaults to `false`
- charged_color: `string` [color][colors] - color to use when fully charged - defaults to segment color
- charging_color: `string` [color][colors] - color to use when charging - defaults to segment color
- discharging_color: `string` [color][colors] - color to use when discharging - defaults to segment color
- display_charging: `bool` - displays the battery status while charging (Charging)
- display_charged: `bool` - displays the battery status when charged (Full)

## Template Properties

- `.Battery`: `struct` - the [battery][battery] object, you can use any property it has e.g. `.Battery.State`
- `.Percentage`: `float64` - the current battery percentage
- `.Error`: `string` - the error in case fetching the battery information failed
- `.Icon`: `string` - the icon based on the battery state

[colors]: /docs/config-colors
[battery]: https://github.com/distatus/battery/blob/master/battery.go#L78
[go-text-template]: https://golang.org/pkg/text/template/
[sprig]: https://masterminds.github.io/sprig/
