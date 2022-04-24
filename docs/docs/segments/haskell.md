---
id: haskell
title: Haskell
sidebar_label: Haskell
---

## What

Display the currently active Glasgow Haskell Compiler (GHC) version.

## Sample Configuration

```json
{
  "type": "haskell",
  "style": "powerline",
  "powerline_symbol": "\uE0B0",
  "foreground": "#906cff",
  "background": "#100e23",
  "template": " \ue61f {{ .Full }}"
}
```

## Properties

- home_enabled: `boolean` - display the segment in the HOME folder or not - defaults to `false`
- fetch_version: `boolean` - display the GHC version - defaults to `true`
- display_error: `boolean` - show the error context when failing to retrieve the version information - defaults to `true`
- missing_command_text: `string` - text to display when the command is missing - defaults to empty
- display_mode: `string` - determines when the segment is displayed
  - `always`: the segment is always displayed
  - `files`: the segment is only displayed when `*.hs`, `*.lhs`, `stack.yaml`, `package.yaml`, `*.cabal`,
or `cabal.project` files are present (default)
- stack_ghc_mode: `string` - determines when to use stack ghc to retrieve the version information.
Using stack ghc will decrease performance.
  - `never`: never use stack ghc (default)
  - `package`: only use stack ghc when `stack.yaml` is in the root of the package
  - `always`: always use stack ghc

## Template ([info][templates])

:::note default template

```template
{{ if .Error }}{{ .Error }}{{ else }}{{ .Full }}{{ end }}
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
- `.StackGhc`: `boolean` - `true` if stack ghc was used, otherwise `false`

[templates]: /docs/configuration/templates
