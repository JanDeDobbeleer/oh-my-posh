---
id: environment
title: Environment Variable
sidebar_label: Environment Variable
---

## What

Show the content of an environment variable.
Can be used to visualize a local settings/context unavailable to Go my Posh otherwise.

For example, in PowerShell, adding the below configuration to a block and extending the prompt
function to set an environment variable before the prompt, you can work a bit of magic.

```powershell
[ScriptBlock]$Prompt = {
  $realLASTEXITCODE = $global:LASTEXITCODE
  $env:POSH = "hello from Powershell"
  & "C:\tools\oh-my-posh.exe" -config "~/downloadedtheme.json" -error $realLASTEXITCODE -pwd $PWD
  $global:LASTEXITCODE = $realLASTEXITCODE
  Remove-Variable realLASTEXITCODE -Confirm:$false
}
```

If you're using the PowerShell module, you can override a function to achieve the same effect.
make sure to do this after importing `go-my-posh` and you're good to go.

```powershell
function Set-EnvVar {
    $env:POSH=$(Get-Date)
}
New-Alias -Name 'Set-PoshContext' -Value 'Set-EnvVar' -Scope Global -Force
```

The segment will show when the value of the environment variable isn't empty.

## Sample Configuration

```json
{
  "type": "envvar",
  "style": "powerline",
  "powerline_symbol": "\uE0B0",
  "foreground": "#ffffff",
  "background": "#0077c2",
  "properties": {
    "var_name": "POSH"
  }
}
```

- var_name: `string` - the name of the environment variable
- template: `string` - A go [text/template][go-text-template] template extended with [sprig][sprig] utilizing the
properties below. Defaults to the value of the environment variable.

## Template Properties

- `.Value`: `string` - the value of the environment variable

[go-text-template]: https://golang.org/pkg/text/template/
[sprig]: https://masterminds.github.io/sprig/
