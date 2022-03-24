---
id: project
title: Project
sidebar_label: Project
---

## What

Display the current version of your project defined in the package file.

Supports:

- Node.js project (`package.json`)

## Sample Configuration

```json
{
  "type": "project",
  "style": "powerline",
  "powerline_symbol": "\uE0B0",
  "foreground": "#193549",
  "background": "#ffeb3b",
  "properties": {
    "template": " {{ if .Error }}{{ .Error }}{{ else }}{{ if .Version }}\uf487 {{.Version}}{{ end }}{{ end }} "
  }
}
```

## Template ([info][templates])

:::note default template

``` template
{{ if .Error }}{{ .Error }}{{ else }}{{ if .Version }}\uf487 {{.Version}}{{ end }}{{ end }}
```

:::

### Properties

- `.Version`: `string` - The version of your project

[templates]: /docs/config-templates
