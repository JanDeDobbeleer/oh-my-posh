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
  "properties": {
    "always_enabled": true,
    "template": "\uE23A",
    "prefix": "<#193549>\uE0B0</> "
  }
}
```

## Properties

- always_enabled: `boolean` - always show the status - defaults to `false`

[colors]: /docs/config-colors

## [Template][templates] Properties

- `.Code`: `number` - the last known exit code
- `.Meaning`: `string` - the textual meaning linked to exit code (if applicable, otherwise identical to `.Code`)

[templates]: /docs/config-templates
