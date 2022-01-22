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
    "template": "{{ .EnvironmentName }}",
    "prefix": " \uFD03 "
  }
}
```

## [Template][templates] Properties

- `.EnvironmentName`: `string` - the account environment name
- `.HomeTenantID`: `string` - the home tenant id
- `.ID`: `string` - the account/subscription id
- `.IsDefault`: `boolean` - is the default account or not
- `.Name`: `string` - the account name
- `.State`: `string` - the account state
- `.TenantID`: `string` - the tenant id
- `.UserName`: `string` - the user name
- `.Origin`: `string` - where we received the information from, can be `CLI` or `PWSH`

[templates]: /docs/config-text#templates
