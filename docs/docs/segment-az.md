---
id: az
title: Azure Subscription
sidebar_label: Azure
---

## What

Display the currently active Azure subscription information.

## Sample Configuration

```json
{
  "type": "az",
  "style": "powerline",
  "powerline_symbol": "\uE0B0",
  "foreground": "#000000",
  "background": "#9ec3f0",
  "template": " \uFD03 {{ .EnvironmentName }}"
}
```

## Properties

- source: `string` - where to fetch the information from - defaults to `first_match`
  - `first_match`: try the CLI config first, then the PowerShell module. The first to resolve is displayed
  - `cli`: fetch the information from the CLI config
  - `pwsh`: fetch the information from the PowerShell Module config

## Template ([info][templates])

:::note default template

``` template
{{ .Name }}
```

:::

### Properties

- `.EnvironmentName`: `string` - account environment name
- `.HomeTenantID`: `string` - home tenant id
- `.ID`: `string` - account/subscription id
- `.IsDefault`: `boolean` - is the default account or not
- `.Name`: `string` - account name
- `.State`: `string` - account state
- `.TenantID`: `string` - tenant id
- `.User.Name`: `string` - user name
- `.User.Type`: `string` - user type
- `.Origin`: `string` - where we received the information from, can be `CLI` or `PWSH`

[templates]: /docs/config-templates
