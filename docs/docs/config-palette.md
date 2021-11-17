---
id: config-palette
title: Palette
sidebar_label: Palette
---

Palette allows you to define a set of named colors and use references to these colors in your theme.
Palette reference can be used everywhere where a [Standard color][colors] is expected.

Having all of the colors defined in one place allows you to import existing color themes (usually with slight
tweaking to adhere to the format), easily change colors of multiple segments at once, and have a more
organized theme overall.

## Defining a Palette

To use a Palette, define a `"palette"` object at the top level of your theme:

```json
{
    "$schema": "https://raw.githubusercontent.com/JanDeDobbeleer/oh-my-posh/main/themes/schema.json",
    "palette": {
        "git-foreground": "#193549",
        "git": "#FFFB38",
        "git-modified": "#FF9248",
        "git-diverged": "#FF4500",
        "git-ahead": "#B388FF",
        "git-behind": "#B388FF",
        "white": "#FFFFFF",
        "black": "#111111"
    },
    "blocks": {
        ...
    }
}
```

Color names can have any string value, so be creative.
Color values, on the other hand, should adhere to the [Standard color][colors] format.

## Using a Palette

You can now reference color from Palette in [Segment][segment] `foreground`, `foreground_templates`,
`background`, `background_templates` properties.
Reference can have 2 forms: either `palette:<color name>` or `p:<color name>`.
Take a look at the [Git][git] segment using Palette references:

```json
{
    "type": "git",
    "style": "powerline",
    "powerline_symbol": "\uE0B0",
    "foreground": "palette:git-foreground",
    "background": "palette:git",
    "background_templates": [
        "{{ if or (.Working.Changed) (.Staging.Changed) }}palette:git-modified{{ end }}",
        "{{ if and (gt .Ahead 0) (gt .Behind 0) }}palette:git-diverged{{ end }}",
        "{{ if gt .Ahead 0 }}palette:git-ahead{{ end }}",
        "{{ if gt .Behind 0 }}palette:git-behind{{ end }}"
    ],
    ...
},
```

## Handling of invalid references

Should you use an invalid Palette reference as a color (for example typo `p:bleu` instead of `p:blue`),
the Pallete engine will use the Transparent keyword as a fallback value. So if you see your prompt segments
rendered with incorrect colors, and you are using a Palette, be sure to check the correctness of your references.

## Limitations

### (Non) Recursive resolution

Palette does not allow for recursive reference resolution. You should not reference Palette colors in other
Palette colors. This configuration will not work, `p:background` and `p:foreground` will be set to Transparent:

```json
    "$schema": "https://raw.githubusercontent.com/JanDeDobbeleer/oh-my-posh/main/themes/schema.json",
    "palette": {
        "light-blue": "#CAF0F8",
        "dark-blue": "#023E8A",
        "background": "p:dark-blue",
        "foreground": "p:light-blue"
    },
    "blocks": {
        ...
    }
```

If you want to have different names for the same color, you can specify multiple keys.

[colors]: /docs/config-colors
[segment]: /docs/config-segment
[git]: /docs/segment-git
