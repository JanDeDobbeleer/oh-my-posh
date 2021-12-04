---
id: python
title: Python
sidebar_label: Python
---

## What

Display the currently active python version and virtualenv.
Supports conda, virtualenv and pyenv.

## Sample Configuration

```json
{
  "type": "python",
  "style": "powerline",
  "powerline_symbol": "\uE0B0",
  "foreground": "#100e23",
  "background": "#906cff",
  "properties": {
    "prefix": " \uE235 "
  }
}
```

## Properties

- home_enabled: `boolean` - display the segment in the HOME folder or not - defaults to `false`
- display_virtual_env: `boolean` - show the name of the virtualenv or not - defaults to `true`
- display_default: `boolean` - show the name of the virtualenv when it's default (`system`, `base`)
or not - defaults to `true`
- display_version: `boolean` - display the python version - defaults to `true`
- display_error: `boolean` - show the error context when failing to retrieve the version information - defaults to `true`
- missing_command_text: `string` - text to display when the command is missing - defaults to empty
- display_mode: `string` - determines when the segment is displayed
  - `always`: the segment is always displayed
  - `files`: the segment is only displayed when `*.py`, `*.ipynb`, `pyproject.toml`, `venv.bak`, `venv`, or `.venv`
    files are present (default)
  - `environment`: the segment is only displayed when a virtual env is present
  - `context`: the segment is only displayed when either `environment` or `files` is active
- template: `string` - A go [text/template][go-text-template] template extended with [sprig][sprig] utilizing the
properties below. Defaults to `{{ .Full }}`

## Template Properties

- `.Venv`: `string` - the virtual environment name (if present)
- `.Full`: `string` - the full version
- `.Major`: `string` - major number
- `.Minor`: `string` - minor number
- `.Patch`: `string` - patch number
- `.Prerelease`: `string` - prerelease info text
- `.BuildMetadata`: `string` - build metadata
- `.Error`: `string` - when fetching the version string errors

[go-text-template]: https://golang.org/pkg/text/template/
[sprig]: https://masterminds.github.io/sprig/
