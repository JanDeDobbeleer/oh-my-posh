---
id: text
title: Text
sidebar_label: Text
---

## What

Display text.

## Sample Configuration

```json
{
  "type": "text",
  "style": "plain",
  "foreground": "#E06C75",
  "properties": {
    "prefix": "",
    "text": " \u276F"
  }
}
```

## Properties

- text: `string` - text/icon to display. Accepts [coloring foreground][coloring] just like `prefix` and `postfix`.
Powered by go [text/template][go-text-template] templates extended with [sprig][sprig] utilizing the
properties below.

## Template Properties

- `.Root`: `boolean` - is the current user root/admin or not
- `.Path`: `string` - the current working directory
- `.Folder`: `string` - the current working folder
- `.Shell`: `string` - the current shell name
- `.User`: `string` - the current user name
- `.Host`: `string` - the host name
- `.Env.VarName`: `string` - Any environment variable where `VarName` is the environment variable name

[coloring]: /docs/config-colors
[go-text-template]: https://golang.org/pkg/text/template/
[sprig]: https://masterminds.github.io/sprig/
