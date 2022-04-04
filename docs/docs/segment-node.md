---
id: node
title: Node
sidebar_label: Node
---

## What

Display the currently active node version.

## Sample Configuration

```json
{
  "type": "node",
  "style": "powerline",
  "powerline_symbol": "\uE0B0",
  "foreground": "#ffffff",
  "background": "#6CA35E",
  "template": " \uE718 {{ .Full }} "
}
```

## Properties

- home_enabled: `boolean` - display the segment in the HOME folder or not - defaults to `false`
- missing_command_text: `string` - text to display when the command is missing - defaults to empty
- display_mode: `string` - determines when the segment is displayed
  - `always`: The segment is always displayed
  - `files`: The segment is only displayed when `*.js`, `*.ts`, `package.json`, `.nvmrc`, `pnpm-workspace.yaml`,
`.pnpmfile.cjs`, `.npmrc` or `.vue` files are present (default)
- fetch_package_manager: `boolean` - define if the current project uses Yarn or NPM - defaults to `false`
- yarn_icon: `string` - the icon/text to display when using Yarn - defaults to ` \uF61A`
- npm_icon: `string` - the icon/text to display when using NPM - defaults to ` \uE71E`

## Template ([info][templates])

:::note default template

``` template
{{ if .PackageManagerIcon }}{{ .PackageManagerIcon }} {{ end }}{{ .Full }}
```

:::

### Properties

- `.Full`: `string` - the full version
- `.Major`: `string` - major number
- `.Minor`: `string` - minor number
- `.Patch`: `string` - patch number
- `.Prerelease`: `string` - prerelease info text
- `.BuildMetadata`: `string` - build metadata
- `.Error`: `string` - when fetching the version string errors
- `.PackageManagerIcon`: `string` - the Yarn on NPM icon when setting `fetch_package_manager` to `true`
- `.Mismatch`: `boolean` - if the version in `.nvmrc` matches with `.Full`

[templates]: /docs/config-templates
