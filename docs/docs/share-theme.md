---
id: share
title: Share theme
sidebar_label: 📸 Share theme
---

You can export your prompt to an image which you can share online. You have the ability to align
it correctly and add your name for credits too.

:::warning
Some glyphs aren't rendered correctly, that's not you but the limitations of the renderer.
Depending on your config, you might have to tweak the output a little bit.
:::

The oh-my-posh executable has the `config export image` command to export your current theme configuration
to the current directory.

```powershell
oh-my-posh config export image --cursor-padding 50
```

There are a couple of additional switches you can use to tweak the image rendering:

- `--cursor-padding`: spaces to add after the cursor indication (`_`)
- `--rprompt-offset`: spaces to add **before** a block that's right aligned
- `--author`: the name of the creator, added after `ohmyposh.dev`
- `--background-color`: the hex background color to use (e.g. `#222222`)

For all options, and additional examples, use `oh-my-posh config export image --help`
