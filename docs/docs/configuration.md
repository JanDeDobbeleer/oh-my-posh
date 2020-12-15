---
id: configure
title: Configuration
sidebar_label: Configuration
---

Oh my Posh renders your prompt based on the definition of _blocks_ (like Lego) which contain or more _segments_.
A really simple configuration could look like this.

```json
{
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
            "prefix": " \uE5FF ",
            "style": "folder"
          }
        }
      ]
    }
  ]
}
```

With this configuration, a single powerline segment is rendered that shows the name of the folder you're currently in.
To set this configuration in combination with a Oh my Posh [executable][releases], use the `-config` flag to
set a path to a json file containing the above code. The `-shell universal` flag is used to print the prompt without
escape characters to see the prompt as it would be shown inside a prompt function for your shell.

:::info
The command below will not persist the configuration for your shell but print the prompt in your terminal.
If you want to use your own configuration permanently, adjust the [prompt configuration][promptconfig] to use your custom
theme.
:::

```bash
oh-my-posh -config sample.json -shell universal
```

If all goes according to plan, you should see the prompt being printed out on the line below. In case you see a lot of
boxes with question marks, [set up your terminal][setupterm] to use a supported font before continuing.

## General Settings

- final_space: `boolean` - when true adds a space at the end of the prompt
- console_title: `boolean` - when true sets the current location as the console title
- console_title_style: `string` - the title to set in the console - defaults to `folder`

> "I Like The Way You Speak Words" - Gary Goodspeed

### Console Title Style

- `folder`: show the current folder name
- `path`: show the current path

## Block

Let's take a closer look at what defines a block.

- type: `prompt` | `rprompt` | `newline`
- alignment: `left` | `right`
- vertical_offset: `int`
- horizontal_offset: `int`
- segments: `array` of one or more `segments`

### Type

Tells the engine what to do with the block. There are three options:

- `prompt` renders one or more segments
- `rprompt` renders one or more segments aligned to the right of the cursor. Only one `rprompt` block is permitted.
Supported on [ZSH][rprompt] and Powershell.
- `newline` inserts a new line to start the next block on a new line. `newline` blocks require no additional
configuration other than the `type`.

### Alignment

Tell the engine if the block should be left or right aligned.

### Vertical offset

Move the block up or down x lines. For example `vertical_offset: 1` moves the prompt down one line, `vertical_offset: -1`
moves it up one line.

### Horizontal offset

Moves the segment to the left or the right to have it exactly where you want it to be. Works like `vertical_offset`
but on a horizontal level where a negative number moves the block left and a positive number right.

### Segments

Array of one or more segments.

## Segment

A segments is a part of the prompt with a certain context. There are different types available out of the box, if you're
looking for what's included, feel free to skip this part and browse through the [segments][segments]. Keep reading to
understand how to configure a segment.

- type: `string` any of the included [segments][segments]
- style: `powerline` | `plain` | `diamond`
- powerline_symbol: `string`
- invert_powerline: `boolean`
- leading_diamond: `string`
- trailing_diamond: `string`
- foreground: `string` [color][colors]
- background: `string` [color][colors]
- properties: `array` of `Property`: `string`

### Type

Takes the `string` value referencing which segment logic it needs to run (see [segments][segments] for possible values).

### Style

Oh Hi! You made it to a really interesting part, great! Style defines how a prompt is rendered. Looking at most prompt
themes out there, we identified 3 types. All of these require a different configuration and depending on the look
you want to achieve you might need to understand/use them all.

#### Powerline

What started it all for us. Makes use of a single symbol (`powerline_symbol`) to separate the segments. It takes the
background color of the previous segment (or transparent if none) and the foreground of the current one (or transparent
if we're at the last segment). Expects segments to have a colored background, else there little use for this one.

#### Plain

Simple. Colored text on a transparent background. Make sure to set `foreground` for maximum enjoyment.

#### Diamond

While Powerline works great with as single symbol, sometimes you want a segment to have a different start and end symbol.
Just like a diamond: `< my segment text >`. The difference between this and plain is that the diamond symbols take the
segment background as their foreground color.

### Powerline symbol

Text character to use when `"style": "powerline"`.

### Invert Powerline

If `true` this swaps the foreground and background colors. Can be useful when the character you want does not exist
in the perfectly mirrored variant for example.

### Leading diamond

Text character to use at the start of the segment. Will take the background color of the segment as
its foreground color.

### Trailing diamond

Text character to use at the end of the segment. Will take the background color of the segment as its foreground color.

### Foreground

Hex [color][colors] to use as the segment text foreground color. Also supports transparency using the `transparent` keyword.

### Background

Hex [color][colors] to use as the segment text background color. Also supports transparency using the `transparent` keyword.

### Properties

An array of **Properties** with a value. This is used inside of the segment logic to tweak what the output of the segment
will be. Segments have the ability to define their own Properties, but there are some general ones being used by the
engine which allow you to customize the output even more.

#### General purpose properties

You can use these on any segment, the engine is responsible for adding them correctly.

- prefix: `string`
- postfix: `string`
- ignore_folders: `[]string`

##### Prefix

The string content will be put in front of the segment's output text. Useful for symbols, text or other customizations.

##### Postfix

The string content will be put after the segment's output text. Useful for symbols, text or other customizations.

##### Ignore Folders

Sometimes you want might want to not have a segment rendered at a certain location. If so, adding the path to the
segment's configuration will not render it when in that location. The engine will simply skip it.

```json
"ignore_folders": [
  "/super/secret/project"
]
```

You can also specify a [regular expression][regex] to create wildcards to exclude certain folders.
In the sample below, folders inside the `/Users/posh/Projects` path will not show the segment.

```json
"ignore_folders": [
  "/Users/posh/Projects/.*"
]
```

Want to only show the segment inside certain folders? Use the [negative lookahead][regex-nl] to only match folders
in a certain path. Everything else will be ignored. In the sample below, only folders inside the
`/Users/posh/Projects/` path will show the segment.

```json
"ignore_folders": [
  "(?!/Users/posh/Projects/).*"
]
```

#### Colors

You have the ability to override the foreground and/or background color for text in any property that accepts it.
The syntax is custom but should be rather straighforward:
`<#ffffff,#000000>this is white with black background</> <#FF479C>but this is pink</>`. Anything between the color start
`<#FF479C>` and end `</>` will be colored accordingly.

For example, if you want `prefix` to print a colored bracket which isn't the same as the segment's `foreground`, you can
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

Oh my Posh mainly supports three different color types being

- Typical [hex colors][hexcolors] (for example `#CB4B16`).
- The `transparent` keyword which can be used to create either a transparent foreground override
  or transparent background color using the segement's foreground property.
- 16 [ANSI color names][ansicolors].

  These include 8 basic ANSI colors and `default`:

  `black` `red` `green` `yellow` `blue` `magenta` `cyan` `white` `default`

  as well as 8 extended ANSI colors:

  `darkGray` `lightRed` `lightGreen` `lightYellow` `lightBlue` `lightMagenta` `lightCyan` `lightWhite`

## Full Sample

```json
{
  "final_space": true,
  "blocks": [
    {
      "type": "prompt",
      "alignment": "right",
      "vertical_offset": -1,
      "segments": [
        {
          "type": "time",
          "style": "plain",
          "foreground": "#007ACC",
          "properties": {
            "time_format": "15:04:05"
          }
        }
      ]
    },
    {
      "type": "newline"
    },
    {
      "type": "prompt",
      "alignment": "left",
      "segments": [
        {
          "type": "session",
          "style": "diamond",
          "foreground": "#ffffff",
          "background": "#ffb300",
          "leading_diamond": "\uE0B6",
          "trailing_diamond": "\uE0B0",
          "properties": {
            "postfix": " "
          }
        },
        {
          "type": "path",
          "style": "powerline",
          "powerline_symbol": "\uE0B0",
          "foreground": "#ffffff",
          "background": "#61AFEF",
          "properties": {
            "prefix": " \uE5FF ",
            "style": "folder",
            "ignore_folders": [
              "/super/secret/project"
            ]
          }
        },
        {
          "type": "git",
          "style": "powerline",
          "powerline_symbol": "\uE0B0",
          "foreground": "#193549",
          "background": "#ffeb3b"
        },
        {
          "type": "exit",
          "style": "diamond",
          "foreground": "#ffffff",
          "background": "#00897b",
          "leading_diamond": "",
          "trailing_diamond": "\uE0B4",
          "properties": {
            "display_exit_code": false,
            "always_enabled": true,
            "error_color": "#e91e63",
            "color_background": true,
            "prefix": "<#193549>\uE0B0 \uE23A</>"
          }
        }
      ]
    }
  ]
}
```

[promptconfig]: /docs/installation#4-replace-your-existing-prompt
[setupterm]: /docs/installation#1-setup-your-terminal
[releases]: https://github.com/JanDeDobbeleer/oh-my-posh3/releases/latest
[nf]: https://www.nerdfonts.com/
[segments]: /docs/battery
[colors]: #colors
[hexcolors]: https://htmlcolorcodes.com/color-chart/material-design-color-chart/
[ansicolors]: https://htmlcolorcodes.com/color-chart/material-design-color-chart/
[fg]: /docs/configure#foreground
[regex]: https://www.regular-expressions.info/tutorial.html
[regex-nl]: https://www.regular-expressions.info/lookaround.html
[rprompt]: https://scriptingosx.com/2019/07/moving-to-zsh-06-customizing-the-zsh-prompt/
