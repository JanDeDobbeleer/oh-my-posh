---
id: media
title: Media
sidebar_label: Media
---

## What

Show the currently playing media info from Windows NowPlayingSession API or media-related process window title.

## Prerequisites

* Windows 10 1511 (10586) or newer
* .NET 5 CLI or newer
* Installed **[sys-media-info](https://github.com/zuozishi/sys-media-info)** (A .NET global tool)

## Sample Configuration

```json
{
  "type": "media",
  "style": "powerline",
  "powerline_symbol": "\uE0B0",
  "foreground": "#ffffff",
  "background": "#1BD760",
  "properties": {
    "prefix": "\uF9C6 ",
    "playing_icon": "\uE602 ",
    "paused_icon": "\uF8E3 ",
    "stopped_icon": "\uF04D ",
    "track_separator" : " - "
  }
}
```

## Properties

* playing_icon: `string` - text/icon to show when playing - defaults to `\uE602 `
* paused_icon: `string` - text/icon to show when paused - defaults to `\uF8E3 `
* stopped_icon: `string` - text/icon to show when stopped - defaults to `\uF04D `
* track_separator: `string` - text/icon to put between the artist and song name - defaults to ` - `
