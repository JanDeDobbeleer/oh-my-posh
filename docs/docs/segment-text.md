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
    "template": "\u276F"
  }
}
```

:::tip
If you're using PowerShell, you can override a function to populate environment variables before the
prompt is rendered.

```powershell
function Set-EnvVar {
    $env:POSH=$(Get-Date)
}
New-Alias -Name 'Set-PoshContext' -Value 'Set-EnvVar' -Scope Global -Force
```

:::

## Template ([info][templates])

:::note default template

``` template
{{ .Text }}
```

:::

### Properties

- `.Root`: `boolean` - is the current user root/admin or not
- `.Path`: `string` - the current working directory
- `.Folder`: `string` - the current working folder
- `.Shell`: `string` - the current shell name
- `.UserName`: `string` - the current user name
- `.HostName`: `string` - the host name
- `.Env.VarName`: `string` - Any environment variable where `VarName` is the environment variable name

[coloring]: /docs/config-colors
[templates]: /docs/config-templates
