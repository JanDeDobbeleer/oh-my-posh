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
  "template": "\u276F"
}
```

## Template ([info][templates])

### Properties

- `.Root`: `boolean` - is the current user root/admin or not
- `.Path`: `string` - the current working directory
- `.Folder`: `string` - the current working folder
- `.Shell`: `string` - the current shell name
- `.UserName`: `string` - the current user name
- `.HostName`: `string` - the host name
- `.Env.VarName`: `string` - Any environment variable where `VarName` is the environment variable name

[coloring]: /docs/configuration/colors
[templates]: /docs/configuration/templates
