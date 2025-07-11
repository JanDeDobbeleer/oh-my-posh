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
- `--font`: path to a regular NerdFont .ttf or .otf file (filename must include "NerdFont"). This font will also be used for bold and italic styles if `--font-bold` and/or `--font-italic` are not specified.
- `--font-regular`: alias for `--font`.
- `--font-bold`: (optional) path to a bold NerdFont .ttf or .otf file (filename must include "NerdFont").
- `--font-italic`: (optional) path to an italic NerdFont .ttf or .otf file (filename must include "NerdFont").

:::info
**Font Handling Details:**

*   **Font Collections:** If `--font` (or `--font-regular`) points to a `.ttc` or `.otc` font collection, the first font in the collection will be used as the regular font.
*   **Style Fallbacks:** If `--font-bold` is not provided, the font specified by `--font` will be used for bold text. Similarly, if `--font-italic` is not provided, the font from `--font` will be used for italic text.
*   **Bundled Fallback:** Oh My Posh will fall back to a bundled Nerd Font (Hack) if no custom fonts are specified or if the specified custom fonts cannot be loaded.
*   **NerdFont Requirement:** Custom fonts provided via `--font`, `--font-bold`, or `--font-italic` **must** be NerdFonts to ensure proper icon display. The font filename must also contain the case-sensitive substring "NerdFont" (e.g., `MyAwesomeNerdFont-Regular.ttf`). If a non-NerdFont or a font with a non-compliant filename is specified, Oh My Posh will attempt to fall back to the bundled Nerd Font.
:::

For all options, and additional examples, use `oh-my-posh config export image --help`
