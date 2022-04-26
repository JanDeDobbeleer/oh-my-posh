---
id: migrating
title: PowerShell module
sidebar_label: 🪦 PowerShell Module
---

## Problem statement

Traditionally, the module was the only way to install oh-my-posh using `Install-Module oh-my-posh`.
Today, with the shift to the executable version over a year ago, it only acts as a wrapper around the
executable, offering no additional functionality. Throughout the year, the following changes have been made:

- don't ship all binaries in the Module but download on `Import-Module`
- move all functionality from the Module to the [init][init] script

There's a problem with the Module due to the following:

- downloading the binary is a problem on company managed computers
- the module syncs cross device thanks to OneDrive sync, causing versions to be out of sync and [configs to break][idiots]
- it's impactful having to explain the difference time and time again (for me)

## Migration steps

### Remove the module's cached files

```powershell
Remove-Item $env:POSH_PATH -Force -Recurse
```

:::warning
If you added custom elements to this location, they will be deleted with the command above.
Make sure to move these before running the command.
:::

### Install oh-my-posh

See your platform's installation guide. The preferred ways are **winget** and **homebrew**.

- [Windows][windows]
- [macOS][macos]
- [Linux][linux]

### Uninstall the PowerShell module

```powershell
Uninstall-Module oh-my-posh --AllVersions
```

Delete the import of the PowerShell module in your `$PROFILE`

```powershell
Import-Module oh-my-posh
```

### Adjust setting the prompt

If you're still using `Set-PoshPrompt`, replace that statement with the following:

#### I have a custom theme

```powershell
oh-my-posh init pwsh --config ~/.custom.omp.json | Invoke-Expression
```

And replace `~/.custom.omp.json` with the location of your theme.

#### I have an out of the box theme

```powershell
oh-my-posh init pwsh --config "$env:POSH_THEMES_PATH\jandedobbeleer.omp.json" | Invoke-Expression
```

Replace `jandedobbeleer.omp.json` with the theme you use.

:::caution
Only winget can add the `$env:POSH_THEMES_PATH` variable. For homebrew, use
`$(brew --prefix oh-my-posh)/themes/jandedobbeleer.omp.json`
:::

[init]: https://github.com/JanDeDobbeleer/oh-my-posh/blob/main/src/shell/scripts/omp.ps1
[idiots]: https://ohmyposh.dev/blog/idiots-everywhere
[windows]: /docs/installation/windows
[macos]: /docs/installation/macos
[linux]: /docs/installation/linux
[set-prompt]: /docs/installation/prompt
