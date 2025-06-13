---
id: share
title: Share theme
sidebar_label: 📸 Share theme
---

You can export your prompt to an image which you can share online. You have the ability to align
it correctly and add your name for credits too.

:::caution
Some glyphs aren't rendered correctly, that's not you but the limitations of the renderer.
Depending on your config, you might have to tweak the output a little bit.
:::

The oh-my-posh executable has the `config export image` command to export your current theme configuration
to a PNG image file (if no other options are specified this will be the name of the config file, or `prompt.png`).

```powershell
oh-my-posh config export image
```

There are a couple of additional flags you can use to tweak the image rendering:

- `--author`: the name of the creator, added after `ohmyposh.dev`
- `--background-color`: the hex background color to use (e.g. `#222222`)
- `--output`: the file to export to (e.g. `mytheme.png`)

For all options, and additional examples, use `oh-my-posh config export image --help`

### Using an alternative NerdFont:

Hack NerdFont is used by default for the image rendering.
You can specify an alternative font by setting the following
environment vars:

    POSH_FONT_REGULAR
    POSH_FONT_BOLD
    POSH_FONT_ITALIC

All three must be set use the same font path/filename for all of them
if you only want to use one font. OpenType will handle them as well as it can.

Example (assuming you have `3270NerdFontMono-Regular.ttf` in the current directory):

    $env:POSH_FONT_REGULAR = "3270NerdFontMono-Regular.ttf"
    $env:POSH_FONT_BOLD = "3270NerdFontMono-Regular.ttf"
    $env:POSH_FONT_ITALIC = "3270NerdFontMono-Regular.ttf"
    oh-my-posh config export image --config my-theme.omp.yaml