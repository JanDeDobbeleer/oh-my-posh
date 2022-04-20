---
id: spotify
title: Spotify
sidebar_label: Spotify
---

## What

Show the currently playing song in the Spotify MacOS/Windows client.
On Windows/WSL, only the playing state is supported (no information when paused/stopped).
On macOS, all states are supported (playing/paused/stopped).
**Be aware this can make the prompt a tad bit slower as it needs to get a response from the Spotify player.**

## Sample Configuration

```json
{
  "type": "spotify",
  "style": "powerline",
  "powerline_symbol": "\uE0B0",
  "foreground": "#ffffff",
  "background": "#1BD760",
  "properties": {
    "playing_icon": "\uE602 ",
    "paused_icon": "\uF8E3 ",
    "stopped_icon": "\uF04D "
  }
}
```

## Properties

- playing_icon: `string` - text/icon to show when playing - defaults to `\uE602 `
- paused_icon: `string` - text/icon to show when paused - defaults to `\uF8E3 `
- stopped_icon: `string` - text/icon to show when stopped - defaults to `\uF04D `

## Template ([info][templates])

:::note default template

``` template
{{ .Icon }}{{ if ne .Status \"stopped\" }}{{ .Artist }} - {{ .Track }}{{ end }}
```

:::

### Properties

- `.Status`: `string` - player status (`playing`, `paused`, `stopped`)
- `.Artist`: `string` - current artist
- `.Track`: `string` - current track
- `.Icon`: `string` - icon (based on `.Status`)

[templates]: /docs/configuration/templates
