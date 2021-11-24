---
id: winreg
title: Windows Registry Key Query
sidebar_label: Windows Registry Key Query
---

## What

Display the content of the requested Windows registry key.

Supported registry key types:

- String
- DWORD (displayed in upper-case 0x hex)

## Sample Configuration

```json
  {
    "type": "winreg",
    "style": "powerline",
    "powerline_symbol": "\uE0B0",
    "foreground": "#ffffff",
    "background": "#444444",
    "properties": {
      "path": "HKLM\\software\\microsoft\\windows nt\\currentversion",
      "key":"buildlab",
      "template":"{{ if .Value }}{{ .Value }}{{ else }}unknown{{ end }}",
      "prefix": " \uE62A "
    }
  },
```

## Properties

- path: `string` - registry path to the desired key using backslashes and with a valid root HKEY name.
- key: `string` - the key to read from the `path location.  If `""`, will read the default value.
- template: `string` - a go [text/template][go-text-template] template extended
  with [sprig][sprig] utilizing the properties below.

## Template Properties

- .Value: `string` - The result of your query

[go-text-template]: https://golang.org/pkg/text/template/
[sprig]: https://masterminds.github.io/sprig/
