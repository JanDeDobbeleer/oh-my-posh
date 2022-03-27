---
id: config-secondary-prompt
title: Secondary prompt
sidebar_label: Secondary prompt
---

:::info
This feature only works in `powershell`, `zsh` and `bash` for the time being.
:::

The secondary prompt is displayed when a command text spans multiple lines. The default is `> `.

You can use go [text/template][go-text-template] templates extended with [sprig][sprig] to enrich the text.
Environment variables are available, just like the [`console_title_template`][console-title] functionality.

## Configuration

You need to extend or create a custom theme with your secondary prompt override. For example:

```json
{
    "$schema": "https://raw.githubusercontent.com/JanDeDobbeleer/oh-my-posh/main/themes/schema.json",
    "blocks": {
        ...
    },
    "secondary_prompt": {
        "background": "transparent",
        "foreground": "#ffffff",
        "template": "-> "
    }
}
```

The configuration has the following properties:

- background: `string` [color][colors]
- foreground: `string` [color][colors]
- template: `string` - A go [text/template][go-text-template] template extended with [sprig][sprig] utilizing the
properties below - defaults to `> `

## Template ([info][templates])

- `.Root`: `boolean` - is the current user root/admin or not
- `.Shell`: `string` - the current shell name
- `.UserName`: `string` - the current user name
- `.HostName`: `string` - the host name

[go-text-template]: https://golang.org/pkg/text/template/
[sprig]: https://masterminds.github.io/sprig/
[console-title]: /docs/config-title#console-title-template
[templates]: /docs/config-templates
[colors]: /docs/config-colors
