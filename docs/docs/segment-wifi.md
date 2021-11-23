---
id: wifi
title: WiFi
sidebar_label: WiFi
---

## What

Show details about the connected WiFi network.

## Sample Configuration

```json
{
  "type": "wifi",
  "style": "powerline",
  "background": "#8822ee",
  "foreground": "#eeeeee",
  "powerline_symbol": "\uE0B0",
  "properties": {
    "template": "{{if .Connectected}}\uFAA8{{else}}\uFAA9{{end}}{{if .Connected}}{{.SSID}} {{.Signal}}% {{.ReceiveRate}}Mbps{{else}}{{.State}}{{end}}"
  }
}
```

## Properties

- connected_icon: `string` - the icon to use when WiFi is connected - defaults to `\uFAA8`
- connected_icon: `string` - the icon to use when WiFi is disconnected - defaults to `\uFAA9`
- template: `string` - A go [text/template][go-text-template] template extended with [sprig][sprig]
utilizing the properties below -
defaults to `{{if .Connectected}}\uFAA8{{else}}\uFAA9{{end}}{{if .Connected}}{{.SSID}} {{.Signal}}% {{.ReceiveRate}}Mbps{{else}}{{.State}}{{end}}`

## Template Properties

- `.Connected`: `bool` - if WiFi is currently connected
- `.State`: `string` - WiFi connection status - e.g. connected or disconnected
- `.SSID`: `string` - the SSID of the current wifi network
- `.RadioType`: `string` - the radio type - e.g. 802.11ac, 802.11ax, 802.11n, etc.
- `.Authentication`: `string` - the authentication type - e.g. WPA2-Personal, WPA2-Enterprise, etc.
- `.Channel`: `int` - the current channel number
- `.ReceiveRate`: `int` - the receive rate (Mbps)
- `.TransmitRate`: `int` - the transmit rate (Mbps)
- `.Signal`: `int` - the signal strength (%)

[go-text-template]: https://golang.org/pkg/text/template/
[sprig]: https://masterminds.github.io/sprig/
