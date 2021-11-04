---
id: owm
title: Open Weather Map
sidebar_label: Open Weather Map 
---

## What

Shows the current weather of a given location.

:::caution

You **must** request an API key at the [Open Weather Map](https://openweathermap.org/price) website.
The free tier for *Current weather and forecasts collection* is sufficient.

:::

## Sample Configuration

```json
{
  "type": "owm",
  "style": "powerline",
  "powerline_symbol": "\uE0B0",
  "foreground": "#ffffff",
  "background": "#FF0000",
  "properties": {
    "apikey": "<YOUR_API_KEY>",
    "location": "AMSTERDAM,NL",
    "units": "metric",
    "enable_hyperlink" : false,
    "http_timeout": 20,
    "cache_timeout": 10,
    "template": "{{.Weather}} ({{.Temperature}}{{.UnitIcon}})"
  }
}
```

## Properties

- apikey: `string` - Your apikey from [Open Weather Map](https://openweathermap.org)
- location: `string` - The requested location.
                        Formatted as <City,STATE,COUNTRY_CODE>. City name, state code and country code divided by comma.
                        Please, refer to ISO 3166 for the state codes or country codes - defaults to `DE BILT,NL`
- units: `string` - Units of measurement.
                    Available values are standard (kelvin), metric (celsius), and imperial (fahrenheit) - defaults to `standard`
- enable_hyperlink: `bool` - Displays an hyperlink to get openweathermap data
- http_timeout: `int` - The default timeout for http request is 20ms.
- cache_timeout: `int` - The default timeout for request caching is 10m. A value of 0 disables the cache.
- template: `string` - A go [text/template][go-text-template] template extended with [sprig][sprig] utilizing the
properties below. Defaults to `{{.Weather}} ({{.Temperature}}{{.UnitIcon}})`

## Template Properties

- `.Weather`: `string` - the current weather icon
- `.Temperature`: `string` - the current temperature
- `.UnitIcon`: `string` - the current unit icon(based on units property)
- `.URL`: `string` - the url of the current api call
