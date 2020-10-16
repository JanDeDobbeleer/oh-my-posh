---
id: path
title: Path
sidebar_label: Path
---

## What

Display the current path.

## Sample Configuration

```json
{
  "type": "path",
  "style": "powerline",
  "powerline_symbol": "\uE0B0",
  "foreground": "#ffffff",
  "background": "#61AFEF",
  "properties": {
    "style": "folder"
  }
}
```

## Properties

- folder_separator_icon: `string` - the symbol to use as a separator between folders
- home_icon: `string` - the icon to display when at `$HOME`
- folder_icon: `string` - the icon to use as a folder indication
- windows_registry_icon: `string` - the icon to display when in the Windows registry
- style: `enum` - how to display the current path

## Style

Style sets the way the path is displayed. Based on previous experience and popular themes, there are 4 flavors.

- agnoster
- short
- full
- folder

### Agnoster

Renders each folder as the `folder_icon` separated by the `folder_separator_icon`.
Only the current folder name is displayed at the end, `$HOME` is replaced by the `home_icon` if you're
inside the `$HOME` location or one of its children.

### Short

Display `$PWD` as a string, replace `$HOME` with the `home_icon` if you're inside the `$HOME` location or
one of its children.

### Full

Display `$PWD` as a string

### Folder

Display the name of the current folder.
