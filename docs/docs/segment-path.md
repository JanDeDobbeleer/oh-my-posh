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
    "style": "folder",
    "mapped_locations": {
      "C:\\temp": "\ue799"
    }
  }
}
```

## Properties

- folder_separator_icon: `string` - the symbol to use as a separator between folders - defaults to platform path separator
- home_icon: `string` - the icon to display when at `$HOME` - defaults to `~`
- folder_icon: `string` - the icon to use as a folder indication - defaults to `..`
- windows_registry_icon: `string` - the icon to display when in the Windows registry - defaults to `\uE0B1`
- style: `enum` - how to display the current path
- mixed_threshold: `number` - the maximum length of a path segment that will be displayed when using `Mixed` -
  defaults to `4`
- max_depth: `number` - maximum path depth to display before shortening when using `Agnoster Short` - defaults to `1`

## Mapped Locations

Allows you to override a location with an icon. It validates if the current path **starts with** the value and replaces
it with the icon if there's a match. To avoid issues with nested overrides, Oh My Posh will sort the list of mapped
locations before doing a replacement.

- mapped_locations_enabled: `boolean` - replace known locations in the path with the replacements before applying the
style - defaults to `true`
- mapped_locations: `object` - custom glyph/text for specific paths. Works regardless of the `mapped_locations_enabled`
setting.

For example, to swap out `C:\Users\Leet\GitHub` with a GitHub icon, you can do the following:

```json
"mapped_locations": {
  "C:\\Users\\Leet\\GitHub": "\uF09B"
}
```

### Notes

- Oh My Posh will accept both `/` and `\` as path separators for a mapped location and will match regardless of which
is used by the current operating system.
- The character `~` at the start of a mapped location will match the user's home directory.
- The match is case-insensitive on Windows and macOS, but case-sensitive on other operating systems.

This means that for user Bill, who has a user account `Bill` on Windows and `bill` on Linux,  `~/Foo` might match
`C:\Users\Bill\Foo` or `C:\Users\Bill\foo` on Windows but only `/home/bill/Foo` on Linux.

## Style

Style sets the way the path is displayed. Based on previous experience and popular themes, there are 5 flavors.

- agnoster
- agnoster_full
- agnoster_short
- agnoster_left
- full
- folder
- mixed
- letter

### Agnoster

Renders each folder as the `folder_icon` separated by the `folder_separator_icon`.
Only the current folder name is displayed at the end.

### Agnoster Full

Renders each folder name separated by the `folder_separator_icon`.

### Agnoster Short

When more than `max_depth` levels deep, it renders one `folder_icon` followed by the names of the last `max_depth` folders,
separated by the `folder_separator_icon`.

### Agnoster Left

Renders each folder as the `folder_icon` separated by the `folder_separator_icon`.
Only the root folder name and it's child are displayed in full.

### Full

Display `$PWD` as a string.

### Folder

Display the name of the current folder.

### Mixed

Works like `Agnoster Full`, but for any middle folder short enough it will display its name instead. The maximum length
for the folders to display is governed by the `mixed_threshold` property.

### Letter

Works like `Full`, but will write every subfolder name using the first letter only, except when the folder name
starts with a symbol or icon.

- `folder` will be shortened to `f`
- `.config` will be shortened to `.c`
- `__pycache__` will be shortened to `__p`
- `➼ folder` will be shortened to `➼ f`

## [Template][templates] Properties

- `.Path`: `string` - the current directory (styled)
- `.StackCount`: `int` - the stack count

[templates]: /docs/config-templates
