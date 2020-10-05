---
id: powershell
title: PowerShell
sidebar_label: PowerShell
---

A package that includes useful helper functions to install and configure Oh my Posh.

## Installation

```powershell
Install-Module oh-my-posh -Scope CurrentUser -AllowPrerelease
```

## Usage

### Show all themes

To display every available theme in the current directory, use the following
command.

```powershell
Get-PoshThemes
```

### Set the prompt

Autocompletion is available so it will loop through all available themes.

```powershell
Set-PoshPrompt -Theme ~/downloadedtheme.json
```

### Print the theme

Useful when you find a theme you like, but you want to tweak the settings a bit. The output is a formatted `json` string.

```powershell
Write-PoshTheme
```
