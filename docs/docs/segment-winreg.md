---
id: winreg    
title: Windows Registry Key Query
sidebar_label: Windows Registry Key Query
---

## What

Display the content of the requested Windows registry key.

Supported registry key types:

* String
* DWORD (displayed in upper-case 0x hex)

## Sample Configuration

```json
  {
    "type": "winreg",
    "style": "powerline",
    "powerline_symbol": "\uE0B0",
    "foreground": "#ffffff",
    "background": "#444444",
    "properties": {
      "registry_path": "HKLM\\software\\microsoft\\windows nt\\currentversion",
      "registry_key":"buildlab",
      "template":"{{if .GotKeyValue}}{{ .KeyValue }}{{else}}unknown{{ end }}",
      "prefix": " \uE62A "
    }
  }, 
```

## Properties

* registry_path: `string` - registry path to the desired key using backslashes and with a valid root HKEY name.
* registry_key: `string` - the key to read from the registry_path location.  If "", will read the "(Default)" value.
