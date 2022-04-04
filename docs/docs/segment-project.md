---
id: project
title: Project
sidebar_label: Project
---

## What

Display the current version of your project defined in the package file.

Supports:

- Node.js project (`package.json`)
- Cargo project (`Cargo.toml`)
- Poetry project (`pyproject.toml`)
- PHP project (`composer.json`)

## Sample Configuration

```json
{
  "type": "project",
  "style": "powerline",
  "powerline_symbol": "\uE0B0",
  "foreground": "#193549",
  "background": "#ffeb3b",
  "template": " {{ if .Error }}{{ .Error }}{{ else }}{{ if .Version }}\uf487 {{.Version}}{{ end }} {{ if .Name }}{{ .Name }}{{ end }}{{ end }} "
}
```

## Template ([info][templates])

:::note default template

``` template
 {{ if .Error }}{{ .Error }}{{ else }}{{ if .Version }}\uf487 {{.Version}}{{ end }} {{ if .Name }}{{ .Name }}{{ end }}{{ end }}
```

:::

### Properties

- `.Version`: `string` - The version of your project
- `.Name`: `string` - The name of your project

[templates]: /docs/config-templates
