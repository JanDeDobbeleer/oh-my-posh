---
id: media
title: Media
sidebar_label: Media
---

## What

Show the currently playing media information from Windows NowPlayingSession API or media-related process window title.

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
    "track_separator": " - ",
    "time_separator": "/",
    "display_time": true
  }
}
```

## Properties

* playing_icon: `string` - text/icon to show when playing - defaults to `\uE602 `
* paused_icon: `string` - text/icon to show when paused - defaults to `\uF8E3 `
* stopped_icon: `string` - text/icon to show when stopped - defaults to `\uF04D `
* track_separator: `string` - text/icon to put between the artist and song name - defaults to ` - `
* time_separator: `string` - text/icon to put between the media position and total time - defaults to `/`
* display_time:`bolean` - show or hidden media position and total time - defaults to `true`

## Example

**Groove:**
![Groove](https://user-images.githubusercontent.com/20531439/129418053-5cf80721-8b27-402e-b8e9-f1df61fefc5b.gif)

**Spotify：**
![Spotify](https://user-images.githubusercontent.com/20531439/129418076-aa65b032-3faa-4e6a-87c9-4ad44cb134e4.gif)

**YouTube (Edge Chromium):**
![Edge Chromium](https://user-images.githubusercontent.com/20531439/129418840-42c790c2-8252-4cd7-ad5a-b6732f3656f3.gif)
