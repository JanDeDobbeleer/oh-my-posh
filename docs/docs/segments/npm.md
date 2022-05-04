---
id: npm
title: npm
sidebar_label: NPM
---

## What

Display the currently active [npm][npm-docs] version.

## Sample Configuration

```json
{
  "type": "npm",
  "style": "powerline",
  "powerline_symbol": "\uE0B0",
  "foreground": "#193549",
  "background": "#ffeb3b",
  "template": "\ue71e {{ .Full }} "
}
```

## Template ([info][templates])

:::note default template

``` template
\ue71e {{.Full}}
```

:::

### Properties

- `.Full`: `string` - the full version
- `.Major`: `string` - major number
- `.Minor`: `string` - minor number
- `.Patch`: `string` - patch number
- `.Error`: `string` - when fetching the version string errors

[templates]: /docs/configuration/templates
[npm-docs]: https://docs.npmjs.com/about-npm
