---
id: config-templates
title: Templates
sidebar_label: Templates
---

Every segment has a `template` property to tweak the text that is displayed.
Under the hood, this uses go's [text/template][go-text-template] feature extended with [sprig][sprig] and
offers a few standard properties to work with.

## Global properties

- `.Root`: `boolean` - is the current user root/admin or not
- `.PWD`: `string` - the current working directory
- `.Folder`: `string` - the current working folder
- `.Shell`: `string` - the current shell name
- `.UserName`: `string` - the current user name
- `.HostName`: `string` - the host name
- `.Code`: `int` - the last exit code

## Environment variables

- `.Env.VarName`: `string` - Any environment variable where `VarName` is the environment variable name

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

## Helper functions

- url: create a hyperlink to a website to open your default browser `{{ url .UpstreamIcon .UpstreamURL }}`
(needs terminal [support][terminal-list-hyperlinks])
- path: create a link to a folder to open your file explorer `{{ path .Path .Location }}`
(needs terminal [support][terminal-list-hyperlinks])
- secondsRound: round seconds to a time indication `{{ secondsRound 3600 }}` -> 1h
- glob: exposes [filepath.Glob][glob] as a boolean template function `{{ if glob "*.go" }}OK{{ else }}NOK{{ end }}`

## Text decoration

You can make use of the following syntax to decorate text:

- `<b>bold</b>`: renders `bold` as bold text
- `<u>underline</u>`: renders `underline` as underlined text
- `<i>italic</i>`: renders `italic` as italic text
- `<s>strikethrough</s>`: renders `strikethrough` as strikethrough text
- `<d>dimmed</d>`: renders `dimmed` as dimmed text
- `<f>blink</f>`: renders `blink` as blinking (flashing) text
- `<r>reversed</r>`: renders `reversed` as reversed text

This can be used in templates and icons/text inside your config.

## Hiding segments

To hide a whole segment including the leading and trailing symbol based on a template, the template must render into
an empty string. This can be achieved with conditional statements (`if`). The example below will render a diamond
segment, only if the environment variable `POSH_ENV` is not empty.

Only spaces are excluded, meaning you can still add line breaks and tabs if that is something you're after.

```json
{
  "type": "text",
  "style": "diamond",
  "leading_diamond": " \ue0b6",
  "trailing_diamond": "\ue0b4",
  "foreground": "#ffffff",
  "background": "#d53c14",
  "template": "{{ if .Env.POSH_ENV }} \uf8c5 {{ .Env.POSH_ENV }} {{ end }}"
}
```

[terminal-list-hyperlinks]: https://gist.github.com/egmontkob/eb114294efbcd5adb1944c9f3cb5feda
[path-segment]: /docs/path
[git-segment]: /docs/git
[go-text-template]: https://golang.org/pkg/text/template/
[sprig]: https://masterminds.github.io/sprig/
[glob]: https://pkg.go.dev/path/filepath#Glob
