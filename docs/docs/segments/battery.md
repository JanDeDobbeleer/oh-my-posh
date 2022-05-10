---
id: battery
title: Battery
sidebar_label: Battery
---

## What

:::caution
The segment is not supported and automatically disabled on windows when WSL 1 is detected. Works fine with WSL 2.
:::

Battery displays the remaining power percentage for your battery.

## Sample Configuration

```json
{
  "type": "battery",
  "style": "powerline",
  "powerline_symbol": "\uE0B0",
  "foreground": "#193549",
  "background": "#ffeb3b",
  "background_templates": [
    "{{if eq \"Charging\" .State.String}}#40c4ff{{end}}",
    "{{if eq \"Discharging\" .State.String}}#ff5722{{end}}",
    "{{if eq \"Full\" .State.String}}#4caf50{{end}}"
  ],
  "template": " {{ if not .Error }}{{ .Icon }}{{ .Percentage }}{{ end }}\uF295 ",
  "properties": {
    "discharging_icon": "\uE231 ",
    "charging_icon": "\uE234 ",
    "charged_icon": "\uE22F "
  }
}
```

## Properties

- display_error: `boolean` - show the error context when failing to retrieve the battery information - defaults to `false`
- charging_icon: `string` - icon to display when charging - defaults to empty
- discharging_icon: `string` - icon to display when discharging - defaults to empty
- charged_icon: `string` - icon to display when fully charged - defaults to empty
- not_charging_icon: `string` - icon to display when fully charged - defaults to empty

## Template ([info][templates])

:::note default template

``` template
{{ if not .Error }}{{ .Icon }}{{ .Percentage }}{{ end }}{{ .Error }}
```

:::

### Properties

- `.State`: `struct` - the battery state, has a `.String` function
- `.Current`: `float64` - Current (momentary) charge rate (in mW).
- `.Full`: `float64` - Last known full capacity (in mWh)
- `.Design`: `float64` - Reported design capacity (in mWh)
- `.ChargeRate`: `float64` - Current (momentary) charge rate (in mW). It is always non-negative, consult .State
field to check whether it means charging or discharging (on some systems this might be always `0` if the battery
doesn't support it)
- `.Voltage`: `float64` - Current voltage (in V)
- `.DesignVoltage`: `float64` - Design voltage (in V). Some systems (e.g. macOS) do not provide a separate
value for this. In such cases, or if getting this fails, but getting `Voltage` succeeds, this field will have
the same value as `Voltage`, for convenience
- `.Percentage`: `float64` - the current battery percentage
- `.Error`: `string` - the error in case fetching the battery information failed
- `.Icon`: `string` - the icon based on the battery state

[colors]: /docs/configuration/colors
[battery]: https://github.com/distatus/battery/blob/master/battery.go#L78
[templates]: /docs/configuration/templates
