---
id: regquery    
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
    "type": "regquery",
    "style": "powerline",
    "powerline_symbol": "\uE0B0",
    "foreground": "#ffffff",
    "background": "#444444",
    "properties": {
        "registry_root": "HKLM",
        "registry_path": "software\\microsoft\\xboxlive",
        "registry_key":"sandbox",
        "query_fail_behaviour":"display_fallback_string",
        "query_fail_fallback_string":"RETAIL",
        "prefix": " \uFAB8 (",
        "postfix": ") "
    }
},
```
## Properties

- registry_root: `string` - the abbreviation of the root HKEY of the registry path:
  - 'HKCR': HKEY_CLASSES_ROOT
  - 'HKCC': HKEY_CURRENT_CONFIG
  - 'HKCU': HKEY_CURRENT_USER
  - 'HKLM': HKEY_LOCAL_MACHINE
  - 'HKU': HKEY_USERS
- registry_path: `string` - the path from root to get to the desired key.
- registry_key: `string` - the key at the destiation root\path location to read.
- query_fail_behaviour: `string` - what to do if unable to get key value from the registry for any reason:
  - `hide_segment`: will not display this segment.
  - `display_fallback_string`: will display the string supplied in the 'query_fail_fallback_string' property.
  - `show_debug_info`: will show details about why the key value could not be retrieved.
- query_fail_fallback_string: `string` - string to display when the requested key value could not be retrieved, and when `display_fallback_string` is supplied as the value for 'query_fail_behaviour'.
