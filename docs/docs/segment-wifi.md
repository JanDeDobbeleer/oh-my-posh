---
id: wifi
title: WiFi
sidebar_label: WiFi
---

## What

Show details about the currently connected WiFi network.

:::info
Currently only supports Windows and WSL. Pull requests for Darwin and Linux support are welcome :)
:::

## Sample Configuration

```json
{
  "type": "wifi",
  "style": "powerline",
  "background": "#8822ee",
  "foreground": "#222222",
  "background_templates": [
    "{{ if (lt .Signal 60) }}#DDDD11{{ else if (lt .Signal 90) }}#DD6611{{ else }}#11CC11{{ end }}"
  ],
  "powerline_symbol": "\uE0B0",
  "properties": {
    "template": "\uFAA8 {{ .SSID }} {{ .Signal }}% {{ .ReceiveRate }}Mbps"
  }
}
```

## [Template][templates] Properties

- `.SSID`: `string` - the SSID of the current wifi network
- `.RadioType`: `string` - the radio type - _e.g. 802.11ac, 802.11ax, 802.11n, etc._
- `.Authentication`: `string` - the authentication type - _e.g. WPA2-Personal, WPA2-Enterprise, etc._
- `.Channel`: `int` - the current channel number
- `.ReceiveRate`: `int` - the receive rate (Mbps)
- `.TransmitRate`: `int` - the transmit rate (Mbps)
- `.Signal`: `int` - the signal strength (%)

[templates]: /docs/config-text#templates
