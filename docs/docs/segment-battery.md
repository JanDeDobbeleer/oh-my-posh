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
    "postfix": "\uF295 "
  }
}
```

## Properties

- battery_icon: `string` - the icon to use as a prefix for the battery percentage
- display_error: `boolean` - show the error context when failing to retrieve the battery information
- charging_icon: `string` - icon to display on the left when charging
- discharging_icon: `string` - icon to display on the left when discharging
- charged_icon: `string` - icon to display on the left when fully charged
- color_background: `boolean` - color the background or foreground for properties below
- charged_color: `string` [hex color code][colors] - color to use when fully charged
- charging_color: `string` [hex color code][colors] - color to use when charging
- discharging_color: `string` [hex color code][colors] - color to use when discharging

[colors]: https://htmlcolorcodes.com/color-chart/material-design-color-chart/
