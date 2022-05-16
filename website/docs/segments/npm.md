---
id: npm
title: NPM
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

## Properties

- home_enabled: `boolean` - display the segment in the HOME folder or not - defaults to `false`
- missing_command_text: `string` - text to display when the command is missing - defaults to empty
- fetch_version: `boolean` - fetch the NPM version - defaults to `true`
- display_mode: `string` - determines when the segment is displayed
  - `always`: the segment is always displayed
  - `files`: the segment is only displayed when `package.json` or `package-lock.json` file are present
- version_url_template: `string` - a go [text/template][go-text-template] [template][templates] that creates
the URL of the version info / release notes

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
- `.URL`: `string` - URL of the version info / release notes
- `.Error`: `string` - error encountered when fetching the version string

[go-text-template]: https://golang.org/pkg/text/template/
[templates]: /docs/configuration/templates
[npm-docs]: https://docs.npmjs.com/about-npm
