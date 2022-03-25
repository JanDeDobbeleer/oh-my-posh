---
id: config-line-error
title: Line error
sidebar_label: Line error
---

:::info
This feature only works in `powershell` for the time being.
:::

Line error, when enabled, replaces the last part of the prompt when the text entered is invalid. It leverages
[PSReadLine's][psreadline] `-PromptText` setting by adding two distinct prompts. One for a valid line,
and one for when there's an error. As PSReadLine will rewrite the last part of
your prompt with the value of either based on the line's context, you will need to make sure everything
is compatible with your config as **these values are only set once** on shell start.

There are two config settings you need to tweak:

- `valid_line`:  displays when the line is valid (again)
- `error_line`:  displays when the line is faulty

You can use go [text/template][go-text-template] templates extended with [sprig][sprig] to enrich the text.
Environment variables are available, just like the [`console_title_template`][console-title] functionality.

## Configuration

You need to extend or create a custom theme with your prompt overrides. For example:

```json
{
    "$schema": "https://raw.githubusercontent.com/JanDeDobbeleer/oh-my-posh/main/themes/schema.json",
    "blocks": {
        ...
    },
    "valid_line": {
        "background": "transparent",
        "foreground": "#ffffff",
        "template": "<#e0def4,#286983>\uf42e </><#286983,transparent>\ue0b4</> "
    },
    "error_line": {
        "background": "transparent",
        "foreground": "#ffffff",
        "template": "<#eb6f92,#286983>\ue009 </><#286983,transparent>\ue0b4</> "
    }
}
```

The configuration has the following properties:

- background: `string` [color][colors]
- foreground: `string` [color][colors]
- template: `string` - A go [text/template][go-text-template] template extended with [sprig][sprig] utilizing the
properties below - defaults to ` `

## Template ([info][templates])

- `.Root`: `boolean` - is the current user root/admin or not
- `.PWD`: `string` - the current working directory
- `.Folder`: `string` - the current working folder
- `.Shell`: `string` - the current shell name
- `.UserName`: `string` - the current user name
- `.HostName`: `string` - the host name
- `.Env.VarName`: `string` - Any environment variable where `VarName` is the environment variable name

## Enable the feature

Invoke Oh My Posh in your `$PROFILE` and add the following line below.

```powershell
oh-my-posh init pwsh --config $env:POSH_THEMES_PATH/jandedobbeleer.omp.json | Invoke-Expression
// highlight-start
Enable-PoshLineError
// highlight-end
```

:::caution
If you import **PSReadLine** separately, make sure to import it before the `Enable-PoshLineError` command.
:::

Restart your shell or reload your `$PROFILE` using `. $PROFILE` for the changes to take effect.

[go-text-template]: https://golang.org/pkg/text/template/
[sprig]: https://masterminds.github.io/sprig/
[console-title]: /docs/config-title#console-title-template
[psreadline]: https://github.com/PowerShell/PSReadLine
[templates]: /docs/config-templates
