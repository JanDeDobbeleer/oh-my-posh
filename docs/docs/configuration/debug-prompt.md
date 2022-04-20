---
id: debug-prompt
title: Debug prompt
sidebar_label: Debug prompt
---

:::info
This feature only works in `powershell` for the time being.
:::

The debug prompt is displayed when you debug a script from the command line or Visual Studio Code.
The default is `[DBG]: `.

You can use go [text/template][go-text-template] templates extended with [sprig][sprig] to enrich the text.
Environment variables are available, just like the [`console_title_template`][console-title] functionality.

## Configuration

You need to extend or create a custom theme with your debug prompt override. For example:

```json
{
    "$schema": "https://raw.githubusercontent.com/JanDeDobbeleer/oh-my-posh/main/themes/schema.json",
    "blocks": {
        ...
    },
    "debug_prompt": {
        "background": "transparent",
        "foreground": "#ffffff",
        "template": "Debugging "
    }
}
```

The configuration has the following properties:

- foreground: `string` [color][colors]
- foreground_templates: foreground [color templates][color-templates]
- background: `string` [color][colors]
- background_templates: background [color templates][color-templates]
- template: `string` - A go [text/template][go-text-template] template extended with [sprig][sprig] utilizing the
properties below - defaults to `[DBG]: `

## Template ([info][templates])

- `.Root`: `boolean` - is the current user root/admin or not
- `.PWD`: `string` - the current working directory
- `.Folder`: `string` - the current working folder
- `.Shell`: `string` - the current shell name
- `.UserName`: `string` - the current user name
- `.HostName`: `string` - the host name
- `.Env.VarName`: `string` - Any environment variable where `VarName` is the environment variable name

[go-text-template]: https://golang.org/pkg/text/template/
[sprig]: https://masterminds.github.io/sprig/
[console-title]: /docs/configuration/title#console-title-template
[templates]: /docs/configuration/templates
[colors]: /docs/configuration/colors
[color-templates]: /docs/configuration/colors#color-templates
