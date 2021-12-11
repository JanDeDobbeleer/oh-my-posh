---
id: wakatime
title: Wakatime
sidebar_label: Wakatime
---

## What

Shows the tracked time on [wakatime](wakatime.com) of the current day

:::caution

You **must** request an API key at the [wakatime](https://wakatime.com) website.
The free tier for *Current weather and forecasts collection* is sufficient. You'll
find the API key in your profile settings page.

:::

## Sample Configuration

```json
{
    "type": "wakatime",
    "style": "powerline",
    "powerline_symbol": "\uE0B0",
    "foreground": "#ffffff",
    "background": "#007acc",
    "properties": {
        "prefix": " \uf7d9  ",
        "apikey": "<<HERE GOES THE API KEY>>",
        "cache_timeout": 10,
        "http_timeout": 500
    }
},
```

## Properties

- apikey: `string` - Your apikey from [wakatime](https://wakatime.com)
- http_timeout: `int` - The default timeout for http request is 20ms. If no segment is shown, try increasing this timeout.
- cache_timeout: `int` - The default timeout for request caching is 10m. A value of 0 disables the cache.
- template: `string` - A go [text/template][go-text-template] template extended with [sprig][sprig] utilizing the
properties below - defaults to `{{ if gt .Hours 0 }}{{.Hours}}h {{ end }}{{.Minutes}}m`

## Template Properties

- `.Hours`: `int` - total hours tracked today (0 - 23)
- `.Minutes`: `int` - additional minutes tracked (0 - 59)
- `.MinutesTotal`: `int` - total minutes tracked today (0 - 1440)

[go-text-template]: https://golang.org/pkg/text/template/
[sprig]: https://masterminds.github.io/sprig/
