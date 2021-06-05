---
id: configure
title: Configuration
sidebar_label: ⚙️ Configuration
---

Oh My Posh renders your prompt based on the definition of _blocks_ (like Lego) which contain one or more _segments_.
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
To set this configuration in combination with a Oh My Posh [executable][releases], use the `--config` flag to
set a path to a json file containing the above code. The `--shell universal` flag is used to print the prompt without
escape characters to see the prompt as it would be shown inside a prompt function for your shell.

:::info
The command below will not persist the configuration for your shell but print the prompt in your terminal.
If you want to use your own configuration permanently, adjust the prompt configuration to use your custom
theme.
:::

```bash
oh-my-posh --config sample.json --shell universal
```

If all goes according to plan, you should see the prompt being printed out on the line below. In case you see a lot of
boxes with question marks, set up your terminal to use a supported font before continuing.

## General Settings

- final_space: `boolean` - when true adds a space at the end of the prompt
- osc99: `boolean` - when true adds support for OSC9;9; (notify terminal of current working directory)
- console_title: `boolean` - when true sets the current location as the console title
- console_title_style: `string` - the title to set in the console - defaults to `folder`
- console_title_template: `string` - the template to use when `"console_title_style" = "template"`
- terminal_background: `string` [color][colors] - terminal background color, set to your terminal's background color when
you notice black elements in Windows Terminal or the Visual Studio Code integrated terminal

> "I Like The Way You Speak Words" - Gary Goodspeed

### Console Title Style

- `folder`: show the current folder name
- `path`: show the current path
- `template`: show a custom template

### Console Title Template

You can create a more custom console title with the use of `"console_title_style" = "template"`.
When this is set, a `console_title_template` is also expected, otherwise the title will remain empty.
Under the hood this uses go's [text/template][go-text-template] feature extended with [sprig][sprig] and
offers a few standard properties to work with.

- `.Root`: `boolean` - is the current user root/admin or not
- `.Path`: `string` - the current working directory
- `.Folder`: `string` - the current working folder
- `.Shell`: `string` - the current shell name
- `.User`: `string` - the current user name
- `.Host`: `string` - the host name
- `.Env.VarName`: `string` - Any environment variable where `VarName` is the environment variable name

A `boolean` can be used for conditional display purposes, a `string` can be displayed.

The following examples illustrate possible contents for `console_title_template`, provided
the current working directory is `/usr/home/omp` and the shell is `zsh`.

```json
{
    "console_title_template": "{{.Folder}}{{if .Root}} :: root{{end}} :: {{.Shell}}",
    // outputs:
    // when root == false: omp :: zsh
    // when root == true: omp :: root :: zsh
    "console_title_template": "{{.Folder}}", // outputs: omp
    "console_title_template": "{{.Shell}} in {{.Path}}", // outputs: zsh in /usr/home/omp
    "console_title_template": "{{.User}}@{{.Host}} {{.Shell}} in {{.Path}}", // outputs: MyUser@MyMachine zsh in /usr/home/omp
    "console_title_template": "{{.Env.USERDOMAIN}} {{.Shell}} in {{.Path}}", // outputs: MyCompany zsh in /usr/home/omp
}
```

## Block

Let's take a closer look at what defines a block.

- type: `prompt` | `rprompt`
- newline: `boolean`
- alignment: `left` | `right`
- vertical_offset: `int`
- horizontal_offset: `int`
- segments: `array` of one or more `segments`

### Type

Tells the engine what to do with the block. There are three options:

- `prompt` renders one or more segments
- `rprompt` renders one or more segments aligned to the right of the cursor. Only one `rprompt` block is permitted.
Supported on [ZSH][rprompt], Bash and Powershell.

### Newline

Start the block on a new line. Defaults to `false`.

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
- foreground_templates: `array` of `string` values
- background: `string` [color][colors]
- background_templates: `array` of `string` values
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

[Color][colors] to use as the segment text foreground color. Also supports transparency using the `transparent` keyword.

### Foreground Templates

Array if string templates to define the foreground color for the given Segment based on the Segment's Template Properties.
Under the hood this uses go's [text/template][go-text-template] feature extended with [sprig][sprig] and
offers a few standard properties to work with. For supported Segments, look for the **Template Properties** section in
the documentation.

The following sample is based on the [AWS Segment][aws].

```json
{
  "type": "aws",
  "style": "powerline",
  "powerline_symbol": "",
  "foreground": "#ffffff",
  "background": "#111111",
  "foreground_templates": [
    "{{if contains \"default\" .Profile}}#FFA400{{end}}",
    "{{if contains \"jan\" .Profile}}#f1184c{{end}}"
  ],
  "properties": {
    "prefix": " \uE7AD "
  }
}
```

The logic is as follows, when `background_templates` contains an array, we will check every template line until there's
one that returns a non-empty string. So, when the contents of `.Profile` contain the word `default`, the first template
returns `#FFA400` and that's the color that will be used. If it contains `jan`, it returns `#f1184c`. When none of the
templates return a value, the foreground value `#ffffff` is used.

### Background

[Color][colors] to use as the segment text background color. Also supports transparency using the `transparent` keyword.

### Background Templates

Same as [Foreground Templates][fg-templ] but for the background color.

### Properties

An array of **Properties** with a value. This is used inside of the segment logic to tweak what the output of the segment
will be. Segments have the ability to define their own Properties, but there are some general ones being used by the
engine which allow you to customize the output even more.

#### General purpose properties

You can use these on any segment, the engine is responsible for adding them correctly.

- prefix: `string`
- postfix: `string`
- include_folders: `[]string`
- exclude_folders: `[]string`

##### Prefix

The string content will be put in front of the segment's output text. Useful for symbols, text or other customizations.

##### Postfix

The string content will be put after the segment's output text. Useful for symbols, text or other customizations.

##### Include / Exclude Folders

Sometimes you might want to have a segment only rendered in certain folders. If `include_folders` is specified,
the segment will only be rendered when in one of those locations. If `exclude_folders` is specified, the segment
will not be rendered when in one of the excluded locations.

```json
"include_folders": [
  "/Users/posh/Projects"
]
```

```json
"exclude_folders": [
  "/Users/posh/Projects"
]
```

You can also specify a [regular expression][regex] to create folder wildcards.
In the sample below, any folders inside the `/Users/posh/Projects` path will be matched.

```json
"include_folders": [
  "/Users/posh/Projects.*"
]
```

You can also combine these properties:

```json
"include_folders": [
  "/Users/posh/Projects.*"
],
"exclude_folders": [
  "/Users/posh/Projects/secret-project.*"
]
```

Note for Windows users: Windows directory separators should be specified as 4 backslashes.

```json
"include_folders": [
  "C:\\\\Projects.*"
],
"exclude_folders": [
  "C:\\\\Projects\\\\secret-project.*"
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

Oh My Posh mainly supports three different color types being

- Typical [hex colors][hexcolors] (for example `#CB4B16`).
- The `transparent` keyword which can be used to create either a transparent foreground override
  or transparent background color using the segment's foreground property.
- 16 [ANSI color names][ansicolors].

  These include 8 basic ANSI colors and `default`:

  `black` `red` `green` `yellow` `blue` `magenta` `cyan` `white` `default`

  as well as 8 extended ANSI colors:

  `darkGray` `lightRed` `lightGreen` `lightYellow` `lightBlue` `lightMagenta` `lightCyan` `lightWhite`

### Text decorations

You can make use of the following syntax to decorate text:

- `<b>bold</b>`: renders `bold` as bold text
- `<u>underline</u>`: renders `underline` as underlined text
- `<i>italic</i>`: renders `italic` as italic text
- `<s>strikethrough</s>`: renders `strikethrough` as strikethrough text

This can be used in templates and icons/text inside your config.

### Hyperlinks

The engine has the ability to render hyperlinks. Your terminal has to support it and the option
has to be enabled at the segment level. Hyperlink generation is disabled by default.

#### Supported segments

- [Path][path-segment]

#### Supported terminals

- [Terminal list][terminal-list-hyperlinks]

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
      "type": "prompt",
      "alignment": "left",
      "newline": true,
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
            "exclude_folders": [
              "/super/secret/project"
            ],
            "enable_hyperlink": false
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

[releases]: https://github.com/JanDeDobbeleer/oh-my-posh/releases/latest
[nf]: https://www.nerdfonts.com/
[segments]: /docs/battery
[colors]: #colors
[hexcolors]: https://htmlcolorcodes.com/color-chart/material-design-color-chart/
[ansicolors]: https://htmlcolorcodes.com/color-chart/material-design-color-chart/
[fg]: /docs/configure#foreground
[regex]: https://www.regular-expressions.info/tutorial.html
[rprompt]: https://scriptingosx.com/2019/07/moving-to-zsh-06-customizing-the-zsh-prompt/
[path-segment]: /docs/path
[terminal-list-hyperlinks]: https://gist.github.com/egmontkob/eb114294efbcd5adb1944c9f3cb5feda
[go-text-template]: https://golang.org/pkg/text/template/
[sprig]: https://masterminds.github.io/sprig/
[aws]: /docs/aws
[fg-templ]: /docs/configure#foreground-templates
