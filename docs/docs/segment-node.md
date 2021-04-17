---
id: node
title: Node
sidebar_label: Node
---

## What

Display the currently active node version.

## Sample Configuration

```json
{
  "type": "node",
  "style": "powerline",
  "powerline_symbol": "\uE0B0",
  "foreground": "#ffffff",
  "background": "#6CA35E",
  "properties": {
    "prefix": " \uE718 "
  }
}
```

## Properties

- display_version: `boolean` - display the node version - defaults to `true`
- display_error: `boolean` - show the error context when failing to retrieve the version information - defaults to `true`
- missing_command_text: `string` - text to display when the command is missing - defaults to empty
- display_mode: `string` - determines when the segment is displayed
  - `always`: The segment is always displayed
  - `files`: The segment is only displayed when `*.js`, `*.ts`, or `package.json` files are present (default)
- enable_version_mismatch: `boolean` - color the segment when the version in `.nvmrc` doesn't match the
returned node version
- color_background: `boolean` - color the background or foreground for `version_mismatch_color` - defaults to `false`
- version_mismatch_color: `string` [color][colors] - the color to use for `enable_version_mismatch` - defaults to
segment's background or foreground color
- display_package_manager: `boolean` - show whether the current project uses Yarn or NPM - defaults to `false`
- yarn_icon: `string` - the icon/text to display when using Yarn - defaults to ` \uF61A`
- npm_icon: `string` - the icon/text to display when using NPM - defaults to ` \uE71E`
