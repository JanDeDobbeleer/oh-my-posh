---
id: cds
title: CDS (SAP CAP)
sidebar_label: CDS
---

## What

Display the active [CDS CLI][sap-cap-cds] version.

## Sample Configuration

```json
{
  "background": "#a7cae1",
  "foreground": "#100e23",
  "powerline_symbol": "\ue0b0",
  "template": " \ue311 cds {{ .Full }} ",
  "style": "powerline",
  "type": "cds"
}
```

## Properties

- home_enabled: `boolean` - display the segment in the HOME folder or not - defaults to `false`
- display_error: `boolean` - show the error context when failing to retrieve the version information - defaults to `true`
- missing_command_text: `string` - text to display when the cds command is missing - defaults to empty
- display_mode: `string` - determines when the segment is displayed
  - `always`: the segment is always displayed
  - `files`: the segment is displayed when `.cdsrc.json`, `.cdsrc-private` or `*.cds` file is present
  - `context`: (default) the segment is displayed when conditions from `files` mode are fulfilled or
    `package.json` file is present and `@sap/cds` is in `dependencies` section

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
- `.HasDependency`: `bool` - a flag if `@sap/cds` was found in `package.json`

[templates]: /docs/configuration/templates
[sap-cap-cds]: https://cap.cloud.sap/docs/tools/#command-line-interface-cli
