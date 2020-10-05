---
id: spotify
title: Spotify
sidebar_label: Spotify
---

## What

Show the currently playing song in the Spotify MacOS client. Only available on MacOS for obvious reasons.
Be aware this can make the prompt a tad bit slower as it needs to get a response from the Spotify player using Applescript.

## Sample Configuration

```json
{
  "type": "spotify",
  "style": "powerline",
  "powerline_symbol": "",
  "foreground": "#ffffff",
  "background": "#1BD760",
  "properties": {
    "prefix": "  ",
    "paused_icon": " ",
    "playing_icon": " "
  }
}
```

## Properties

- playing_icon: `string` - text/icon to show when playing.
- paused_icon: `string` - text/icon to show when paused.
- track_separator: `string` - text/icon to put between the artist and song name.
