---
id: aws
title: AWS Context
sidebar_label: AWS
---

## What

Display the currently active AWS profile and region.

## Sample Configuration

```json
{
  "type": "aws",
  "style": "powerline",
  "powerline_symbol": "\uE0B0",
  "foreground": "#ffffff",
  "background": "#FFA400",
  "properties": {
    "prefix": " \uE7AD ",
    "template": "{{.Profile}}{{if .Region}}@{{.Region}}{{end}}"
  }
}
```

## Properties

- template: `string` - A go [text/template][go-text-template] template extended with [sprig][sprig] utilizing the
properties below - defaults to `{{.Context}}{{if .Namespace}} :: {{.Namespace}}{{end}}`
- display_default: `boolean` - display the segment or not when the user profile matches `default` - defaults
to `true`

## Template Properties

- `.Profile`: `string` - the currently active profile
- `.Region`: `string` - the currently active region

[go-text-template]: https://golang.org/pkg/text/template/
[sprig]: https://masterminds.github.io/sprig/
