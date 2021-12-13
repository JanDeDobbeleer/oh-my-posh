---
id: wakatime
title: Wakatime
sidebar_label: Wakatime
---

## What

Shows the tracked time on [wakatime][wt] of the current day

:::caution

You **must** request an API key at the [wakatime][wt] website.
The free tier for is sufficient. You'll find the API key in your profile settings page.

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
        "url": "https://wakatime.com/api/v1/users/current/summaries?start=today&end=today&api_key=API_KEY",
        "cache_timeout": 10,
        "http_timeout": 500
    }
},
```

## Properties

- url: `string` - Your Wakatime [summaries][wk-summaries] URL, including the API key. Example above. You'll know this
works if you can curl it yourself and a result. - defaults to ``
- http_timeout: `int` - The default timeout for http request is 20ms. If no segment is shown, try increasing this timeout.
- cache_timeout: `int` - The default timeout for request caching is 10m. A value of 0 disables the cache.
- template: `string` - A go [text/template][go-text-template] template extended with [sprig][sprig] utilizing the
properties below - defaults to `{{ secondsRound .CummulativeTotal.Seconds }}`

## Template Properties

- `.CummulativeTotal`: `wtTotals` - object holding total tracked time values

### wtTotal Properties

- `.Seconds`: `int` - a number reprecenting the total tracked time in seconds
- `.Text`: `string` - a string with human readable tracked time (eg: "2 hrs 30 mins")

[wt]: https://wakatime.com
[wk-summaries]: https://wakatime.com/developers#summaries
[go-text-template]: https://golang.org/pkg/text/template/
[sprig]: https://masterminds.github.io/sprig/
