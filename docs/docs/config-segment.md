---
id: config-segment
title: Segment
sidebar_label: Segment
---

A segment is a part of the prompt with a certain context. There are different types available out-of-the-box, if you're
looking for what's included, feel free to skip this part and browse through the [segments][segments]. Keep reading to
understand how to configure a segment.

```json
{
  "$schema": "https://raw.githubusercontent.com/JanDeDobbeleer/oh-my-posh/main/themes/schema.json",
  ...
  "blocks": [
    {
      ...
      "segments": [
        {
          "type": "path",
          "style": "powerline",
          "powerline_symbol": "\uE0B0",
          "foreground": "#ffffff",
          "background": "#61AFEF",
          "properties": {
            ...
          }
        }
      ]
    }
  ]
}
```

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

## Type

Takes the `string` value referencing which segment logic it needs to run (see [segments][segments] for possible values).

## Style

Oh Hi! You made it to a really interesting part, great! Style defines how a prompt is rendered. Looking at the most prompt
themes out there, we identified 3 types. All of these require a different configuration and depending on the look
you want to achieve you might need to understand/use them all.

### Powerline

What started it all for us. Makes use of a single symbol (`powerline_symbol`) to separate the segments. It takes the
background color of the previous segment (or transparent if none) and the foreground of the current one (or transparent
if we're at the last segment). Expects segments to have a colored background, else there little use for this one.

### Plain

Simple. Colored text on a transparent background. Make sure to set `foreground` for maximum enjoyment.
Segments will be separated by empty spaces unless you specify `''` for the `prefix` and `postfix` settings for the segment.

### Diamond

While Powerline works great with a single symbol, sometimes you want a segment to have a different start and end symbol.
Just like a diamond: `< my segment text >`. The difference between this and plain is that the diamond symbols take the
segment background as their foreground color.

## Powerline symbol

Text character to use when `"style": "powerline"`.

## Invert Powerline

If `true` this swaps the foreground and background colors. Can be useful when the character you want does not exist
in the perfectly mirrored variant for example.

## Leading diamond

Text character to use at the start of the segment. Will take the background color of the segment as
its foreground color.

## Trailing diamond

Text character to use at the end of the segment. Will take the background color of the segment as its foreground color.

## Foreground

[Color][colors] to use as the segment text foreground color. Also supports transparency using the `transparent` keyword.

## Foreground Templates

Array if string templates to define the foreground color for the given Segment based on the Segment's Template Properties.
Under the hood this uses go's [text/template][go-text-template] feature extended with [sprig][sprig] and
offers a few standard properties to work with. For supported Segments, look for the **Template Properties** section in
the documentation.

The following sample is based on the [AWS Segment][aws].

```json
{
  "type": "aws",
  "style": "powerline",
  "powerline_symbol": "\uE0B0",
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

The logic is as follows: when `background_templates` contains an array, we will check every template line until there's
one that returns a non-empty string. So, when the contents of `.Profile` contain the word `default`, the first template
returns `#FFA400` and that's the color that will be used. If it contains `jan`, it returns `#f1184c`. When none of the
templates returns a value, the foreground value `#ffffff` is used.

## Background

[Color][colors] to use as the segment text background color. Also supports transparency using the `transparent` keyword.

## Background Templates

Same as [Foreground Templates][fg-templ] but for the background color.

## Properties

An array of **Properties** with a value. This is used inside of the segment logic to tweak what the output of the segment
will be. Segments have the ability to define their own Properties, but there are some general ones being used by the
engine which allow you to customize the output even more.

### General-purpose properties

You can use these on any segment, the engine is responsible for adding them correctly.

- prefix: `string`
- postfix: `string`
- include_folders: `[]string`
- exclude_folders: `[]string`
- template: `string` - A go text/template [template][templates] to render the text

#### Prefix

The string content will be put in front of the segment's output text. Useful for symbols, text or other customizations.
If this is not set, it will be an empty space in `plain` mode. If you want to remove the space before the segment,
specify this as `''`.

#### Postfix

The string content will be put after the segment's output text. Useful for symbols, text or other customizations.
If this is not set, it will default to an empty space in `plain` mode. If you want to remove the space after the segment,
specify this as `''`.

#### Include / Exclude Folders

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

The strings specified in these properties are evaluated as [regular expressions][regex]. You
can use any valid regular expression construct, but the regular expression must match the entire directory
name. The following will match `/Users/posh/Projects/Foo` but not `/home/Users/posh/Projects/Foo`.

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

#### Notes

- Oh My Posh will accept both `/` and `\` as path separators for a folder and will match regardless of which
is used by the current operating system.
- Because the strings are evaluated as regular expressions, if you want to use a `\` in a Windows
directory name, you need to specify it as `\\\\`.
- The character `~` at the start of a specified folder will match the user's home directory.
- The comparison is case-insensitive on Windows and macOS, but case-sensitive on other operating systems.

This means that for user Bill, who has a user account `Bill` on Windows and `bill` on Linux,  `~/Foo` might match
`C:\Users\Bill\Foo` or `C:\Users\Bill\foo` on Windows but only `/home/bill/Foo` on Linux.

[segments]: /docs/battery
[colors]: /docs/config-colors
[go-text-template]: https://golang.org/pkg/text/template/
[sprig]: https://masterminds.github.io/sprig/
[fg-templ]: /docs/config-overview#foreground-templates
[regex]: https://www.regular-expressions.info/tutorial.html
[aws]: /docs/aws
[templates]: /docs/config-text#templates
