---
id: session
title: Session
sidebar_label: Session
---

## What

Show the current user and host name.

## Sample Configuration

```json
{
  "type": "session",
  "style": "diamond",
  "foreground": "#ffffff",
  "background": "#c386f1",
  "leading_diamond": "\uE0B6",
  "trailing_diamond": "\uE0B0"
}
```

## Properties

- user_info_separator: `string` - text/icon to put in between the user and host name - defaults to `@`
- ssh_icon: `string` - text/icon to display first when in an active SSH session - defaults
to `\uF817 `
- user_color: `string` [hex color code][colors] - override the foreground color of the user name
- host_color: `string` [hex color code][colors] - override the foreground color of the host name
- display_user: `boolean` - display the user name or not - defaults to `true`
- display_host: `boolean` - display the host name or not - defaults to `true`

[colors]: https://htmlcolorcodes.com/color-chart/material-design-color-chart/
