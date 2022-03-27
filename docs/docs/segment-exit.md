---
id: exit
title: Exit code
sidebar_label: Exit code
---

## What

Displays the last exit code or that the last command failed based on the configuration.

## Sample Configuration

```json
{
  "type": "exit",
  "style": "diamond",
  "foreground": "#ffffff",
  "background": "#00897b",
  "background_templates": [
    "{{ if gt .Code 0 }}#e91e63{{ end }}",
  ],
  "leading_diamond": "",
  "trailing_diamond": "\uE0B4",
  "template": "<#193549>\uE0B0</> \uE23A ",
  "properties": {
    "always_enabled": true
  }
}
```

## Properties

- always_enabled: `boolean` - always show the status - defaults to `false`

[colors]: /docs/config-colors

## Template ([info][templates])

:::note default template

``` template
{{ if gt .Code 0 }}\uf00d {{ .Meaning }}{{ else }}\uf42e{{ end }}
```

:::

### Properties

- `.Code`: `number` - the last known exit code
- `.Meaning`: `string` - the textual meaning linked to exit code (if applicable, otherwise identical to `.Code`)

[templates]: /docs/config-templates
