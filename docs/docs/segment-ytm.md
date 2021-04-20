---
id: ytm
title: YouTube Music
sidebar_label: YouTube Music
---

## What

Shows the currently playing song in the [YouTube Music Desktop App](https://github.com/ytmdesktop/ytmdesktop).

**NOTE**: You **must** enable Remote Control in YTMDA for this segment to work: `Settings > Integrations > Remote Control`

It is fine if `Protect remote control with password` is automatically enabled. This segment does not require the
Remote Control password.

## Sample Configuration

```json
{
  "type": "ytm",
  "style": "powerline",
  "powerline_symbol": "\uE0B0",
  "foreground": "#ffffff",
  "background": "#FF0000",
  "properties": {
    "prefix": "\uF16A ",
    "playing_icon": "\uE602 ",
    "paused_icon": "\uF8E3 ",
    "stopped_icon": "\uF04D ",
    "track_separator" : " - "
  }
}
```

## Properties

- playing_icon: `string` - text/icon to show when playing - defaults to `\uE602 `
- paused_icon: `string` - text/icon to show when paused - defaults to `\uF8E3 `
- stopped_icon: `string` - text/icon to show when paused - defaults to `\uF04D `
- track_separator: `string` - text/icon to put between the artist and song name - defaults to ` - `
- api_url: `string` - the YTMDA Remote Control API URL- defaults to `http://127.0.0.1:9863`
