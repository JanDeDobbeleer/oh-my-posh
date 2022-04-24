---
id: ipify
title: Ipify
sidebar_label: Ipify
---

## What

[Ipify][ipify] is a simple Public IP Address API, it returns your public IP Adress in plain text.

## Sample Configuration

```json
{
  "type": "ipify",
  "style": "diamond",
  "foreground": "#ffffff",
  "background": "#c386f1",
  "leading_diamond": "",
  "trailing_diamond": "\uE0B0",
  "template": "{{ .IP }}",
  "properties": {
    "cache_timeout": 5,
    "http_timeout": 1000
  }
}
```

## Properties

- url: `string` - The Ipify URL, by default IPv4 is used, use `https://api64.ipify.org` for IPv6 - defaults to `https://api.ipify.org`
- http_timeout: `int` - How long may the segment wait for a response of the ipify API? -
  defaults to 20ms
- cache_timeout: `int` in minutes - How long do you want your IP address cached? -
  defaults to 10 min

## Template ([info][templates])

:::note default template

``` template
{{ .IP }}
```

:::

### Properties

- .IP: `string` - Your external IP address

[templates]: /docs/configuration/templates
[ipify]: https://www.ipify.org/
