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

[hexcolors]: https://htmlcolorcodes.com/color-chart/material-design-color-chart/
[ansicolors]: https://htmlcolorcodes.com/color-chart/material-design-color-chart/
