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

Additional flags are supported to configure behaviour if the key cannot be retrieved from the registry.

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
    "query_fail_behaviour":"display_fallback_string",
    "query_fail_fallback_string":"unknown",
    "prefix": " \uE62A ",
  }
}, 
```

## Properties

* registry_path: `string` - registry path to the desired key using backslashes and with a valid root HKEY name.
* registry_key: `string` - the key to read from the registry_path location.  If "", will read the "(Default)" value.
* query_fail_behaviour: `string` - what to do if unable to get key value from the registry for any reason:
  * `hide_segment`: will not display this segment.
  * `display_fallback_string`: will display the string supplied in the 'query_fail_fallback_string' property.
* query_fail_fallback_string: `string` - display if key value could not be retrieved, per `display_fallback_string`.
