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
  "template": " \uE7AD {{.Profile}}{{if .Region}}@{{.Region}}{{end}}"
}
```

## Properties

- display_default: `boolean` - display the segment or not when the user profile matches `default` - defaults
to `true`

## Template ([info][templates])

:::note default template

``` template
{{ .Profile }}{{ if .Region }}@{{ .Region }}{{ end }}
```

:::

### Properties

- `.Profile`: `string` - the currently active profile
- `.Region`: `string` - the currently active region

[templates]: /docs/configuration/templates
