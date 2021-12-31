---
id: config-fonts
title: Fonts
sidebar_label: Fonts
---

### Nerd Fonts

Oh My Posh was designed to use [Nerd Fonts][nerdfonts]. Nerd Fonts are popular fonts that are patched to include icons.
We recommend [Meslo LGM NF][meslo], but any Nerd Font should be compatible with the standard [themes][themes].

To see the icons displayed in Oh My Posh, **install** a [Nerd Font][nerdfonts], and **configure** your terminal to use it.

#### Windows

Download your chosen Nerd Font, and install the font system-wide. See this [thread][font-thread] for more context.

#### Windows Terminal

Once you have installed a Nerd Font, you will need to configure the Windows Terminal to use it. This can be easily done
by modifying the Windows Terminal settings (default shortcut: `CTRL + SHIFT + ,`). In your `settings.json` file, add the
`fontFace` attribute under the `defaults` attribute in `profiles`:

```json
{
    "profiles":
    {
        "defaults":
        {
            "font":
            {
                "face": "MesloLGM NF"
            }
        }
    }
}
```

### Other Fonts

If you are not interested in using a Nerd Font, you will want to use a theme which doesn't include any Nerd Font icons.
The `minimal` themes do not make use of Nerd Font icons.

[Creating your own theme][configuration] is always an option too 😊

[nerdfonts]: https://www.nerdfonts.com/
[meslo]: https://github.com/ryanoasis/nerd-fonts/releases/download/v2.1.0/Meslo.zip
[themes]: https://github.com/JanDeDobbeleer/oh-my-posh/tree/main/themes
[font-thread]: https://github.com/JanDeDobbeleer/oh-my-posh/issues/145#issuecomment-730162622
[configuration]: /docs/config-overview
