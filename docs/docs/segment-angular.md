---
id: angular
title: Angular
sidebar_label: Angular
---

## What

Display the currently active Angular CLI version.

## Sample Configuration

```json
{
  "type": "angular",
  "style": "powerline",
  "powerline_symbol": "\uE0B0",
  "foreground": "#000000",
  "background": "#1976d2",
  "properties": {
    "prefix": "\uE753"
  }
}
```

## Properties

- display_version: `boolean` - display the active version or not; useful if all you need is an icon indicating `ng`
- display_mode: `string` - determines when the segment is displayed
  - `always`: the segment is always displayed
  - `files`: the segment is only displayed when `angular.json` file is present (default)
- template: `string` - A go [text/template][go-text-template] template extended with [sprig][sprig] utilizing the
properties below. Defaults to `{{ .Full }}`

## Template Properties

- `.Full`: `string` - the full version
- `.Major`: `string` - is the major version
- `.Minor`: `string` - is the minor version
- `.Patch`: `string` - is the patch version
- `.Prerelease`: `string` - is the prerelease version
- `.BuildMetadata`: `string` - is the build metadata

[go-text-template]: https://golang.org/pkg/text/template/
[sprig]: https://masterminds.github.io/sprig/
