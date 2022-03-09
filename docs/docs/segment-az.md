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
  "properties": {
    "template": " \uFD03 {{ .EnvironmentName }}"
  }
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

- `.EnvironmentName`: `string` - the account environment name
- `.HomeTenantID`: `string` - the home tenant id
- `.ID`: `string` - the account/subscription id
- `.IsDefault`: `boolean` - is the default account or not
- `.Name`: `string` - the account name
- `.State`: `string` - the account state
- `.TenantID`: `string` - the tenant id
- `.User.Name`: `string` - the user name
- `.Origin`: `string` - where we received the information from, can be `CLI` or `PWSH`

[templates]: /docs/config-templates
