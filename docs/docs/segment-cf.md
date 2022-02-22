---
id: cf
title: Cloud Foundry
sidebar_label: Cloud Foundry
---

## What

Display the active [Cloud Foundry CLI][cloud-foundry] version.

## Sample Configuration

```json
{
  "background": "#a7cae1",
  "foreground": "#100e23",
  "powerline_symbol": "\ue0b0",
  "properties": {
    "template": " \uf40a cf {{ .Full }} "
  },
  "style": "powerline",
  "type": "cf"
}
```

## Properties

- home_enabled: `boolean` - display the segment in the HOME folder or not - defaults to `false`
- fetch_version: `boolean` - display the Cloud Foundry CLI version - defaults to `true`
- display_error: `boolean` - show the error context when failing to retrieve the version information - defaults to `true`
- missing_command_text: `string` - text to display when the java command is missing - defaults to empty
- display_mode: `string` - determines when the segment is displayed
  - `always`: the segment is always displayed
  - `files`: the segment is displayed when `manifest.yml` or `mta.yaml` file is present (default)

## Template ([info][templates])

:::note default template

```template
{{ if .Error }}{{ .Error }}{{ else }}{{ .Full }}{{ end }}
```

:::

## Template Properties

- `.Full`: `string` - the full version
- `.Major`: `string` - major number
- `.Minor`: `string` - minor number
- `.Patch`: `string` - patch number
- `.Prerelease`: `string` - prerelease info text
- `.BuildMetadata`: `string` - build metadata
- `.Error`: `string` - when fetching the version string errors

[templates]: /docs/config-templates
[cloud-foundry]: https://github.com/cloudfoundry/cli
