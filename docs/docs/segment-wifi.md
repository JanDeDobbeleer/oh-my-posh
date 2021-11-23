---
id: wifi
title: WiFi
sidebar_label: WiFi
---

## What

Show details about the currently connected WiFi network.

::info
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
    "{{ if (not .Connected) }}#FF1111{{ end }}"
    "{{ if (lt .Signal 60) }}#DDDD11{{ else if (lt .Signal 90) }}#DD6611{{ else }}#11CC11{{ end }}"
  ],
  "powerline_symbol": "\uE0B0",
  "properties": {
    "template": "{{ if .Connected }}\uFAA8{{ else }}\uFAA9{{ end }}
    {{ if .Connected }}{{ .SSID }} {{ .Signal }}% {{ .ReceiveRate }}Mbps{{ else }}{{ .State }}{{ end }}"
  }
}
```

## Properties

- template: `string` - A go [text/template][go-text-template]  extended with [sprig][sprig] using the properties below

## Template Properties

- `.Connected`: `bool` - if WiFi is currently connected
- `.State`: `string` - WiFi connection status - _e.g. connected or disconnected_
- `.SSID`: `string` - the SSID of the current wifi network
- `.RadioType`: `string` - the radio type - _e.g. 802.11ac, 802.11ax, 802.11n, etc._
- `.Authentication`: `string` - the authentication type - _e.g. WPA2-Personal, WPA2-Enterprise, etc._
- `.Channel`: `int` - the current channel number
- `.ReceiveRate`: `int` - the receive rate (Mbps)
- `.TransmitRate`: `int` - the transmit rate (Mbps)
- `.Signal`: `int` - the signal strength (%)

[go-text-template]: https://golang.org/pkg/text/template/
[sprig]: https://masterminds.github.io/sprig/
