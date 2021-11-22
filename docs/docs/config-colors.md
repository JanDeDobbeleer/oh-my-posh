---
id: config-colors
title: Colors
sidebar_label: Colors
---

## Standard colors

Oh My Posh supports multiple different color references, being:

- Typical [hex colors][hexcolors] (for example `#CB4B16`).
- 16 [ANSI color names][ansicolors].
- The `transparent` keyword which can be used to create either a transparent foreground override
  or transparent background color using the segment's foreground property.
- The `foreground` keyword which can be used to reference the current segment's foreground color.
- The `background` keyword which can be used to reference the current segment's background color.
- The `parentForeground` keyword which can be used to inherit the previous active segment's foreground color.
- The `parentBackground` keyword which can be used to inherit the previous active segment's background color.

  These include 8 basic ANSI colors and `default`:

  `black` `red` `green` `yellow` `blue` `magenta` `cyan` `white` `default`

  as well as 8 extended ANSI colors:

  `darkGray` `lightRed` `lightGreen` `lightYellow` `lightBlue` `lightMagenta` `lightCyan` `lightWhite`

## Color overrides

You have the ability to override the foreground and/or background color for text in any property that accepts it.
The syntax is custom but should be rather straight-forward: `<foreground,background>text</>`. For example,
`<#ffffff,#000000>this is white with black background</> <#FF479C>but this is pink</>`.
Anything between the color start `<#FF479C>` and end `</>` will be colored accordingly.

If you want `prefix` to print a colored bracket that isn't the same as the segment's `foreground`, you can
do so like this:

```json
"prefix": "<#CB4B16>┏[</>",
```

If you also wanted to change the background color in the previous command, you would do so like this:

```json
"prefix": "<#CB4B16,#FFFFFF>┏[</>",
```

To change *only* the background color, just omit the first color from the above string:

```json
"prefix": "<,#FFFFFF>┏[</>",
```

## Palette

If your theme defined the Palette, you can use the _Palette reference_ `p:<palette key>` in places where the
__Standard color__ is expected.

### Defining a Palette

Palette is a set of named __Standard colors__. To use a Palette, define a `"palette"` object
at the top level of your theme:

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
        "red":   "#FF0000",
        "green": "#00FF00",
        "blue":  "#0000FF",
        "white": "#FFFFFF",
        "black": "#111111"
    },
    "blocks": {
        ...
    }
}
```

Color names (palette keys) can have any string value, so be creative.
Color values, on the other hand, should adhere to the __Standard color__ format.

### Using a Palette

You can now _Palette references_ in any [Segment's][segment] `foreground`, `foreground_templates`,
`background`, `background_templates` properties, and other config properties that expect __Standard color__ value.
_Palette reference_ format is `p:<palette key>`. Take a look at the [Git][git] segment using _Palette references_:

```json
{
    "type": "git",
    "style": "powerline",
    "powerline_symbol": "\uE0B0",
    "foreground": "p:git-foreground",
    "background": "p:git",
    "background_templates": [
        "{{ if or (.Working.Changed) (.Staging.Changed) }}p:git-modified{{ end }}",
        "{{ if and (gt .Ahead 0) (gt .Behind 0) }}p:git-diverged{{ end }}",
        "{{ if gt .Ahead 0 }}p:git-ahead{{ end }}",
        "{{ if gt .Behind 0 }}p:git-behind{{ end }}"
    ],
    ...
},
```

Having all of the colors defined in one place allows you to import existing color themes (usually with slight
tweaking to adhere to the format), easily change colors of multiple segments at once, and have a more
organized theme overall. Be creative!

### _Palette references_ and __Standard colors__

Using Palette does not interfere with using __Standard colors__ in your theme. You can still use __Standard colors__
everywhere. This can be useful if you want to use a specific color for a single segment element, or in a
_Color override_ ([Battery segment][battery]):

```json
{
    "type": "battery",
    "style": "powerline",
    "invert_powerline": true,
    "powerline_symbol": "\uE0B2",
    "foreground": "p:white",
    "background": "p:black",
    "properties": {
        "battery_icon": "<#ffa500> [II ]- </>", // icon should always be orange
        "discharging_icon": "- ",
        "charging_icon": "+ ",
        "charged_icon": "* ",
        "color_background": true,
        "charged_color": "#4caf50",     //
        "charging_color": "#40c4ff",    // battery should use specific colors for status
        "discharging_color": "#ff5722", //
    }
},
```

### Handling of invalid references

Should you use an invalid _Palette reference_ as a color (for example typo `p:bleu` instead of `p:blue`),
the Pallete engine will use the Transparent keyword as a fallback value. So if you see your prompt segments
rendered with incorrect colors, and you are using a Palette, be sure to check the correctness of your references.

### Recursive resolution

Palette allows for recursive _Palette reference_ resolution. You can use a _Palette reference_ as a color
value in Palette. This allows you to define named colors, and use references to those colors as Palette values.
For example, `p:foreground` and `p:background`  will be correctly set to "#CAF0F80" and "#023E8A":

```json
    "$schema": "https://raw.githubusercontent.com/JanDeDobbeleer/oh-my-posh/main/themes/schema.json",
    "palette": {
        "light-blue": "#CAF0F8",
        "dark-blue": "#023E8A",
        "foreground": "p:light-blue",
        "background": "p:dark-blue"
    },
    "blocks": {
        ...
    }
```

[hexcolors]: https://htmlcolorcodes.com/color-chart/material-design-color-chart/
[ansicolors]: https://htmlcolorcodes.com/color-chart/material-design-color-chart/
[git]: /docs/segment-git
[battery]: /docs/segment-battery
