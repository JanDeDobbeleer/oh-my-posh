---
id: config-overview
title: General
sidebar_label: General
---

Oh My Posh renders your prompt based on the definition of _blocks_ (like Lego) which contain one or more _segments_.
A really simple configuration could look like this. The default format is `json`, but we also support `toml` and `yaml`.
There's a [schema][schema] available which is kept up-to-date and helps with autocomplete and validation of the configuration.

:::info
There are a few [themes][themes] available which are basically predefined configs. You can use these as they are, or as a
starting point to create your own configuration.
:::

```json
{
  "$schema": "https://raw.githubusercontent.com/JanDeDobbeleer/oh-my-posh/main/themes/schema.json",
  "final_space": true,
  "blocks": [
    {
      "type": "prompt",
      "alignment": "left",
      "segments": [
        {
          "type": "path",
          "style": "powerline",
          "powerline_symbol": "\uE0B0",
          "foreground": "#ffffff",
          "background": "#61AFEF",
          "properties": {
            "style": "folder"
          }
        }
      ]
    }
  ]
}
```

With this configuration, a single powerline segment is rendered that shows the name of the folder you're currently in.
To set this configuration in combination with a Oh My Posh [executable][releases], use the `--config` flag to
set a path to a JSON file containing the above code. The `--shell universal` flag is used to print the prompt without
escape characters to see the prompt as it would be shown inside a prompt function for your shell.

:::caution
The command below will not persist the configuration for your shell but print the prompt in your terminal.
If you want to use your own configuration permanently, adjust the prompt configuration to use your custom
theme.
:::

```bash
oh-my-posh --config sample.json --shell uni
```

If all goes according to plan, you should see the prompt being printed out on the line below. In case you see a lot of
boxes with question marks, set up your terminal to use a [supported font][font] before continuing.

## General Settings

- final_space: `boolean` - when true adds a space at the end of the prompt
- osc99: `boolean` - when true adds support for OSC9;9; (notify terminal of current working directory)
- terminal_background: `string` [color][colors] - terminal background color, set to your terminal's background color when
you notice black elements in Windows Terminal or the Visual Studio Code integrated terminal

[releases]: https://github.com/JanDeDobbeleer/oh-my-posh/releases/latest
[font]: /docs/config-fonts
[schema]: https://github.com/JanDeDobbeleer/oh-my-posh/blob/main/themes/schema.json
[themes]: https://github.com/JanDeDobbeleer/oh-my-posh/tree/main/themes
