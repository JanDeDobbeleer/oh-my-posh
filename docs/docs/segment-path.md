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

- enable_hyperlink: `boolean` - displays an hyperlink for the path - defaults to `false`
- mixed_threshold: `number` - the maximum length of a path segment that will be displayed when using `Mixed` -
  defaults to `4`
- stack_count_enabled: `boolean` - displays the stack count when using pushd/popd - defaults to `false`

## Mapped Locations

Allows you to override a location with an icon. It validates if the current path **starts with** the value and replaces
it with the icon if there's a match. To avoid issues with nested overrides, Oh my posh will sort the list of mapped
locations before doing a replacement.

- mapped_locations_enabled: `boolean` - replace known locations in the path with the replacements before applying the
style. defaults to `true`
- mapped_locations: `map[string]string` - custom glyph/text for specific paths. Works regardless of the `mapped_locations_enabled`
setting.

For example, to swap out `C:\Users\Leet\GitHub` with a GitHub icon, you can do the following:

```json
"mapped_locations": {
  "C:\\Users\\Leet\\GitHub": "\uF09B"
}
```

## Style

Style sets the way the path is displayed. Based on previous experience and popular themes, there are 5 flavors.

- agnoster
- agnoster_full
- agnoster_short
- full
- folder
- mixed
- letter

### Agnoster

Renders each folder as the `folder_icon` separated by the `folder_separator_icon`.
Only the current folder name is displayed at the end, `$HOME` is replaced by the `home_icon` if you're
inside the `$HOME` location or one of its children.

### Agnoster Full

Renders each folder name separated by the `folder_separator_icon`.

### Agnoster Short

When more than 1 level deep, it renders one `folder_icon` followed by the name of the current folder separated by the `folder_separator_icon`.

### Full

Display `$PWD` as a string.

### Folder

Display the name of the current folder.

### Mixed

Works like `Agnoster Full`, but for any middle folder short enough it will display its name instead. The maximum length
for the folders to display is governed by the `mixed_threshold` property.

### Letter

Works like `Full`, but will write every subfolder name using the first letter only.
