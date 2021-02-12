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
properties below. Defaults to `{{.Context}}{{if .Namespace}} :: {{.Namespace}}{{end}}`

## Template Properties

- `.Profile`: `string` - the currently active profile
- `.Region`: `string` - the currently active region

[go-text-template]: https://golang.org/pkg/text/template/
[sprig]: https://masterminds.github.io/sprig/
